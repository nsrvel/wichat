package service

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/wichat/wichat/backend/ms-chat/internal/models"
	searchrepo "github.com/wichat/wichat/backend/ms-chat/internal/search/repository"
	channelacl "github.com/wichat/wichat/backend/ms-chat/pkg/channel-acl"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

// QueryService runs Elasticsearch search queries (messages and files).
type QueryService interface {
	SearchMessages(ctx context.Context, accessToken, groupID, query string, limit int) ([]models.MessageSearchHit, error)
	SearchFiles(ctx context.Context, accessToken, groupID, query string, limit int, imagesOnly bool) ([]models.FileSearchHit, error)
}

type queryService struct {
	repo    searchrepo.EsRepository
	channel channelacl.ChannelLister
	log     logger.Logger
}

// NewQueryService builds ES search query use cases.
func NewQueryService(
	repo searchrepo.EsRepository,
	channel channelacl.ChannelLister,
	log logger.Logger,
) QueryService {
	return queryService{repo: repo, channel: channel, log: log}
}

func (s queryService) SearchMessages(
	ctx context.Context,
	accessToken, groupID, query string,
	limit int,
) ([]models.MessageSearchHit, error) {
	// Validation
	accessToken = strings.TrimSpace(accessToken)
	groupID = strings.TrimSpace(groupID)
	query = strings.TrimSpace(query)
	if accessToken == "" {
		return nil, exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}
	if groupID == "" || query == "" {
		return nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "query"})
	}
	if utf8.RuneCountInString(query) < 2 {
		return nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "query"})
	}
	if limit <= 0 {
		limit = 8
	}
	if limit > 25 {
		limit = 25
	}
	if s.repo == nil {
		return []models.MessageSearchHit{}, nil
	}

	// Repository
	docs, err := s.repo.SearchMessages(ctx, groupID, query, limit*3)
	if err != nil {
		return nil, err
	}
	if len(docs) == 0 {
		return []models.MessageSearchHit{}, nil
	}

	allowed, err := channelacl.AllowedIDs(ctx, s.channel, accessToken, groupID)
	if err != nil {
		return nil, err
	}

	hits := make([]models.MessageSearchHit, 0, limit)
	for _, doc := range docs {
		if _, ok := allowed[doc.ChannelID]; !ok {
			continue
		}
		hits = append(hits, models.MessageSearchHit{
			MessageID:         doc.MessageID,
			GroupID:           doc.GroupID,
			ChannelID:         doc.ChannelID,
			BodyPreview:       previewBody(doc.Body, 140),
			AuthorDisplayName: doc.AuthorDisplayName,
			CreatedAt:         doc.CreatedAt,
		})
		if len(hits) >= limit {
			break
		}
	}
	return hits, nil
}

func (s queryService) SearchFiles(
	ctx context.Context,
	accessToken, groupID, query string,
	limit int,
	imagesOnly bool,
) ([]models.FileSearchHit, error) {
	// Validation
	accessToken = strings.TrimSpace(accessToken)
	groupID = strings.TrimSpace(groupID)
	query = strings.TrimSpace(query)
	if accessToken == "" {
		return nil, exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}
	if groupID == "" || query == "" {
		return nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "query"})
	}
	if utf8.RuneCountInString(query) < 2 {
		return nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "query"})
	}
	if limit <= 0 {
		limit = 8
	}
	if limit > 25 {
		limit = 25
	}
	if s.repo == nil {
		return []models.FileSearchHit{}, nil
	}

	docs, err := s.repo.SearchFiles(ctx, groupID, query, limit*3, imagesOnly)
	if err != nil {
		return nil, err
	}
	if len(docs) == 0 {
		return []models.FileSearchHit{}, nil
	}

	allowed, err := channelacl.AllowedIDs(ctx, s.channel, accessToken, groupID)
	if err != nil {
		return nil, err
	}

	hits := make([]models.FileSearchHit, 0, limit)
	for _, doc := range docs {
		if _, ok := allowed[doc.ChannelID]; !ok {
			continue
		}
		hits = append(hits, models.FileSearchHit{
			AttachmentID: doc.AttachmentID,
			MessageID:    doc.MessageID,
			GroupID:      doc.GroupID,
			ChannelID:    doc.ChannelID,
			Filename:     doc.Filename,
			ContentType:  doc.ContentType,
			CreatedAt:    doc.CreatedAt,
		})
		if len(hits) >= limit {
			break
		}
	}
	return hits, nil
}

func previewBody(body string, max int) string {
	body = strings.Join(strings.Fields(strings.TrimSpace(body)), " ")
	if max <= 0 || utf8.RuneCountInString(body) <= max {
		return body
	}
	runes := []rune(body)
	if len(runes) <= max {
		return body
	}
	return string(runes[:max-1]) + "…"
}
