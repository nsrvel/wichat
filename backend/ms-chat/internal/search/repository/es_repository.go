package repository

import (
	"context"

	"github.com/wichat/wichat/backend/wi-shared/infra/elasticsearch"
)

// EsRepository indexes chat messages for search.
type EsRepository interface {
	EnsureIndex(ctx context.Context) error
	IndexMessageCreated(ctx context.Context, doc elasticsearch.MessageDoc) error
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
