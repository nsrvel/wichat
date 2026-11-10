package mock

import (
	"context"
	"time"

	"github.com/wichat/wichat/backend/ms-user/internal/channel/repository"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
)

type pgRepositoryMock struct{}

// NewPgRepositoryMock is an in-memory channel repo for fast tests (no default seed).
func NewPgRepositoryMock() repository.PgRepository {
	return pgRepositoryMock{}
}

func (pgRepositoryMock) ListByGroupID(_ context.Context, _ string) ([]models.Channel, error) {
	return nil, nil
}

func (pgRepositoryMock) FindByPublicIDInGroup(_ context.Context, _, _ string) (models.Channel, error) {
	return models.Channel{}, nil
}

func (pgRepositoryMock) SeedDefaultChannels(_ context.Context, _, _ string, _ time.Time) error {
	return nil
}
