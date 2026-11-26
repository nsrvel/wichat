package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

const (
	DocKindMessage    = "message"
	DocKindAttachment = "attachment"
)

// MessageDoc is a message document for full-text search (v0 indexer).
type MessageDoc struct {
	DocKind           string `json:"doc_kind,omitempty"`
	MessageID         string `json:"message_id"`
	GroupID           string `json:"group_id"`
	ChannelID         string `json:"channel_id"`
	AuthorID          string `json:"author_id"`
	AuthorDisplayName string `json:"author_display_name"`
	Body              string `json:"body"`
	CreatedAt         string `json:"created_at"`
}

// AttachmentDoc indexes attachment filename metadata for group search.
type AttachmentDoc struct {
	DocKind      string `json:"doc_kind"`
	AttachmentID string `json:"attachment_id"`
	MessageID    string `json:"message_id,omitempty"`
	GroupID      string `json:"group_id"`
	ChannelID    string `json:"channel_id"`
	Filename     string `json:"filename"`
	ContentType  string `json:"content_type"`
	CreatedAt    string `json:"created_at"`
}

// Client indexes and queries chat messages.
type Client interface {
	EnsureIndex(ctx context.Context) error
	IndexMessage(ctx context.Context, doc MessageDoc) error
	IndexAttachment(ctx context.Context, doc AttachmentDoc) error
	DeleteMessage(ctx context.Context, messageID string) error
	SearchMessages(ctx context.Context, groupID, query string, limit int) ([]MessageDoc, error)
	SearchAttachments(ctx context.Context, groupID, query string, limit int, imagesOnly bool) ([]AttachmentDoc, error)
}

type esClient struct {
	es    *elasticsearch.Client
	index string
	log   logger.Logger
}

// New returns a Client. Empty URL yields nil (no error).
func New(cfg Config, log logger.Logger) (Client, error) {
	url := strings.TrimSpace(cfg.URL)
	if url == "" {
		return nil, nil
	}
	index := strings.TrimSpace(cfg.Index)
	if index == "" {
		return nil, fmt.Errorf("elasticsearch: index required when url is set")
	}

	es, err := elasticsearch.NewClient(elasticsearch.Config{Addresses: []string{url}})
	if err != nil {
		return nil, fmt.Errorf("elasticsearch client: %w", err)
	}

	return esClient{es: es, index: index, log: log}, nil
}

func (c esClient) EnsureIndex(ctx context.Context) error {
	res, err := c.es.Indices.Exists([]string{c.index}, c.es.Indices.Exists.WithContext(ctx))
	if err != nil {
		return err
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode == 200 {
		return nil
	}

	mapping := map[string]any{
		"mappings": map[string]any{
			"properties": map[string]any{
				"doc_kind":            map[string]string{"type": "keyword"},
				"message_id":          map[string]string{"type": "keyword"},
				"attachment_id":       map[string]string{"type": "keyword"},
				"group_id":            map[string]string{"type": "keyword"},
				"channel_id":          map[string]string{"type": "keyword"},
				"filename":            map[string]string{"type": "text"},
				"content_type":        map[string]string{"type": "keyword"},
				"author_id":           map[string]string{"type": "keyword"},
				"author_display_name": map[string]string{"type": "keyword"},
				"body":                map[string]string{"type": "text"},
				"created_at":          map[string]string{"type": "date"},
			},
		},
	}
	body, err := json.Marshal(mapping)
	if err != nil {
		return err
	}

	createRes, err := c.es.Indices.Create(
		c.index,
		c.es.Indices.Create.WithContext(ctx),
		c.es.Indices.Create.WithBody(bytes.NewReader(body)),
	)
	if err != nil {
		return err
	}
	defer func() { _ = createRes.Body.Close() }()

	if createRes.IsError() {
		return readAPIError("create index", createRes)
	}

	return nil
}

func (c esClient) IndexMessage(ctx context.Context, doc MessageDoc) error {
	if strings.TrimSpace(doc.DocKind) == "" {
		doc.DocKind = DocKindMessage
	}
	payload, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	res, err := c.es.Index(
		c.index,
		bytes.NewReader(payload),
		c.es.Index.WithContext(ctx),
		c.es.Index.WithDocumentID(doc.MessageID),
		c.es.Index.WithRefresh("false"),
	)
	if err != nil {
		return err
	}
	defer func() { _ = res.Body.Close() }()

	if res.IsError() {
		return readAPIError("index message", res)
	}

	return nil
}

func (c esClient) IndexAttachment(ctx context.Context, doc AttachmentDoc) error {
	doc.DocKind = DocKindAttachment
	payload, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	docID := strings.TrimSpace(doc.AttachmentID)
	if docID == "" {
		return fmt.Errorf("elasticsearch: empty attachment id")
	}

	res, err := c.es.Index(
		c.index,
		bytes.NewReader(payload),
		c.es.Index.WithContext(ctx),
		c.es.Index.WithDocumentID("att:"+docID),
		c.es.Index.WithRefresh("false"),
	)
	if err != nil {
		return err
	}
	defer func() { _ = res.Body.Close() }()
	if res.IsError() {
		return readAPIError("index attachment", res)
	}
	return nil
}

func (c esClient) DeleteMessage(ctx context.Context, messageID string) error {
	messageID = strings.TrimSpace(messageID)
	if messageID == "" {
		return fmt.Errorf("elasticsearch: empty message id")
	}

	res, err := c.es.Delete(
		c.index,
		messageID,
		c.es.Delete.WithContext(ctx),
		c.es.Delete.WithRefresh("false"),
	)
	if err != nil {
		return err
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode == 404 {
		return nil
	}
	if res.IsError() {
		return readAPIError("delete message", res)
	}

	return nil
}

func (c esClient) SearchMessages(ctx context.Context, groupID, query string, limit int) ([]MessageDoc, error) {
	groupID = strings.TrimSpace(groupID)
	query = strings.TrimSpace(query)
	if groupID == "" || query == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 8
	}
	if limit > 50 {
		limit = 50
	}

	payload := map[string]any{
		"size": limit,
		"sort": []any{map[string]any{"created_at": map[string]string{"order": "desc"}}},
		"query": map[string]any{
			"bool": map[string]any{
				"filter": []any{
					map[string]any{"term": map[string]any{"group_id": groupID}},
					map[string]any{
						"bool": map[string]any{
							"should": []any{
								map[string]any{"term": map[string]any{"doc_kind": DocKindMessage}},
								map[string]any{"bool": map[string]any{"must_not": map[string]any{"exists": map[string]string{"field": "doc_kind"}}}},
							},
							"minimum_should_match": 1,
						},
					},
				},
				"must": []any{
					map[string]any{
						"match": map[string]any{
							"body": map[string]any{
								"query":    query,
								"operator": "and",
							},
						},
					},
				},
			},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	res, err := c.es.Search(
		c.es.Search.WithContext(ctx),
		c.es.Search.WithIndex(c.index),
		c.es.Search.WithBody(bytes.NewReader(body)),
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()
	if res.IsError() {
		return nil, readAPIError("search messages", res)
	}

	var parsed struct {
		Hits struct {
			Hits []struct {
				Source MessageDoc `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return nil, err
	}

	out := make([]MessageDoc, 0, len(parsed.Hits.Hits))
	for _, hit := range parsed.Hits.Hits {
		if hit.Source.MessageID == "" {
			continue
		}
		out = append(out, hit.Source)
	}
	return out, nil
}

func (c esClient) SearchAttachments(ctx context.Context, groupID, query string, limit int, imagesOnly bool) ([]AttachmentDoc, error) {
	groupID = strings.TrimSpace(groupID)
	query = strings.TrimSpace(query)
	if groupID == "" || query == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 8
	}
	if limit > 50 {
		limit = 50
	}

	filters := []any{
		map[string]any{"term": map[string]any{"group_id": groupID}},
		map[string]any{"term": map[string]any{"doc_kind": DocKindAttachment}},
	}
	if imagesOnly {
		filters = append(filters, map[string]any{"prefix": map[string]any{"content_type": "image/"}})
	}

	payload := map[string]any{
		"size": limit,
		"sort": []any{map[string]any{"created_at": map[string]string{"order": "desc"}}},
		"query": map[string]any{
			"bool": map[string]any{
				"filter": filters,
				"must": []any{
					map[string]any{
						"match": map[string]any{
							"filename": map[string]any{
								"query":    query,
								"operator": "and",
							},
						},
					},
				},
			},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	res, err := c.es.Search(
		c.es.Search.WithContext(ctx),
		c.es.Search.WithIndex(c.index),
		c.es.Search.WithBody(bytes.NewReader(body)),
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()
	if res.IsError() {
		return nil, readAPIError("search attachments", res)
	}

	var parsed struct {
		Hits struct {
			Hits []struct {
				Source AttachmentDoc `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return nil, err
	}

	out := make([]AttachmentDoc, 0, len(parsed.Hits.Hits))
	for _, hit := range parsed.Hits.Hits {
		if hit.Source.AttachmentID == "" {
			continue
		}
		out = append(out, hit.Source)
	}
	return out, nil
}

func readAPIError(label string, res *esapi.Response) error {
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, res.Body)
	return fmt.Errorf("elasticsearch %s: %s %s", label, res.Status(), strings.TrimSpace(buf.String()))
}
