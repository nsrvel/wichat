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

// MessageDoc is a message document for full-text search (v0 indexer).
type MessageDoc struct {
	MessageID         string `json:"message_id"`
	GroupID           string `json:"group_id"`
	ChannelID         string `json:"channel_id"`
	AuthorID          string `json:"author_id"`
	AuthorDisplayName string `json:"author_display_name"`
	Body              string `json:"body"`
	CreatedAt         string `json:"created_at"`
}

// Client indexes chat messages.
type Client interface {
	EnsureIndex(ctx context.Context) error
	IndexMessage(ctx context.Context, doc MessageDoc) error
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
				"message_id":          map[string]string{"type": "keyword"},
				"group_id":            map[string]string{"type": "keyword"},
				"channel_id":          map[string]string{"type": "keyword"},
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

func readAPIError(label string, res *esapi.Response) error {
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, res.Body)
	return fmt.Errorf("elasticsearch %s: %s %s", label, res.Status(), strings.TrimSpace(buf.String()))
}
