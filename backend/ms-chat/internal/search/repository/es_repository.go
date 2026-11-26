package repository

import (
	"context"

	"github.com/wichat/wichat/backend/wi-shared/infra/elasticsearch"
)

// EsRepository indexes chat messages for search.
type EsRepository interface {
	EnsureIndex(ctx context.Context) error
	IndexMessageCreated(ctx context.Context, doc elasticsearch.MessageDoc) error
	DeleteMessage(ctx context.Context, messageID string) error
	SearchMessages(ctx context.Context, groupID, query string, limit int) ([]elasticsearch.MessageDoc, error)
	IndexAttachmentCreated(ctx context.Context, doc elasticsearch.AttachmentDoc) error
	SearchFiles(ctx context.Context, groupID, query string, limit int, imagesOnly bool) ([]elasticsearch.AttachmentDoc, error)
}

type esRepository struct {
	es elasticsearch.Client
}

// NewEsRepository wraps the shared Elasticsearch client.
func NewEsRepository(es elasticsearch.Client) EsRepository {
	if es == nil {
		return nil
	}
	return esRepository{es: es}
}

func (r esRepository) EnsureIndex(ctx context.Context) error {
	return r.es.EnsureIndex(ctx)
}

func (r esRepository) IndexMessageCreated(ctx context.Context, doc elasticsearch.MessageDoc) error {
	return r.es.IndexMessage(ctx, doc)
}

func (r esRepository) DeleteMessage(ctx context.Context, messageID string) error {
	return r.es.DeleteMessage(ctx, messageID)
}

func (r esRepository) SearchMessages(ctx context.Context, groupID, query string, limit int) ([]elasticsearch.MessageDoc, error) {
	if r.es == nil {
		return nil, nil
	}
	return r.es.SearchMessages(ctx, groupID, query, limit)
}

func (r esRepository) IndexAttachmentCreated(ctx context.Context, doc elasticsearch.AttachmentDoc) error {
	return r.es.IndexAttachment(ctx, doc)
}

func (r esRepository) SearchFiles(ctx context.Context, groupID, query string, limit int, imagesOnly bool) ([]elasticsearch.AttachmentDoc, error) {
	if r.es == nil {
		return nil, nil
	}
	return r.es.SearchAttachments(ctx, groupID, query, limit, imagesOnly)
}
