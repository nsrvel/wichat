package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"github.com/wichat/wichat/backend/ms-chat/internal/models"
	"github.com/wichat/wichat/backend/ms-chat/pkg/constants"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

// PgRepository persists event_outbox rows.
type PgRepository interface {
	Insert(ctx context.Context, in models.EventOutboxInsert) error
	InsertInTx(ctx context.Context, tx bun.Tx, in models.EventOutboxInsert) error
	ClaimPending(ctx context.Context, limit int) ([]models.EventOutboxRow, error)
	MarkPublished(ctx context.Context, id string, publishedAt time.Time) error
	RecordPublishFailure(ctx context.Context, id, errMsg string) error
}

type pgRepository struct {
	db  *bun.DB
	log logger.Logger
}

// NewPgRepository constructs an outbox Postgres store.
func NewPgRepository(db *bun.DB, log logger.Logger) PgRepository {
	return pgRepository{db: db, log: log}
}

func (r pgRepository) Insert(ctx context.Context, in models.EventOutboxInsert) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		return r.InsertInTx(ctx, tx, in)
	})
}

func (r pgRepository) InsertInTx(ctx context.Context, tx bun.Tx, in models.EventOutboxInsert) error {
	row, err := models.EventOutboxRowFromInsert(uuid.NewString(), in)
	if err != nil {
		return err
	}
	_, err = tx.NewInsert().Model(&row).Exec(ctx)
	if err != nil {
		r.logIfInfraErr("InsertInTx", err)
	}

	return err
}

func (r pgRepository) ClaimPending(ctx context.Context, limit int) ([]models.EventOutboxRow, error) {
	if limit <= 0 {
		limit = constants.OutboxPublishBatchSize
	}

	var tableRows []models.EventOutboxTable
	err := r.db.NewSelect().
		Model(&tableRows).
		Where("published_at IS NULL").
		OrderExpr("created_at ASC").
		Limit(int64(limit)).
		For("UPDATE SKIP LOCKED").
		Scan(ctx)
	if err != nil {
		r.logIfInfraErr("ClaimPending", err)

		return nil, err
	}

	out := make([]models.EventOutboxRow, len(tableRows))
	for i, row := range tableRows {
		out[i] = row.ToDomain()
	}

	return out, nil
}

func (r pgRepository) MarkPublished(ctx context.Context, id string, publishedAt time.Time) error {
	res, err := r.db.NewUpdate().
		Model((*models.EventOutboxTable)(nil)).
		Set("published_at = ?", publishedAt).
		Where("id = ? AND published_at IS NULL", id).
		Exec(ctx)
	if err != nil {
		r.logIfInfraErr("MarkPublished", err)

		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return constants.ErrNotFound
	}

	return nil
}

func (r pgRepository) RecordPublishFailure(ctx context.Context, id, errMsg string) error {
	_, err := r.db.NewUpdate().
		Model((*models.EventOutboxTable)(nil)).
		Set("publish_attempts = publish_attempts + 1").
		Set("last_error = ?", errMsg).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		r.logIfInfraErr("RecordPublishFailure", err)

		return err
	}

	return nil
}

func (r pgRepository) logIfInfraErr(op string, err error) {
	if err == nil {
		return
	}
	r.log.Error("outbox repository", "op", op, "err", err)
}
