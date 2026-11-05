package repository

import (
	"context"
	"errors"
	"time"

	"github.com/uptrace/bun"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	pgErr "github.com/wichat/wichat/backend/ms-user/pkg/db"
	"github.com/wichat/wichat/backend/ms-user/pkg/token"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

// PgRepository loads and persists channel metadata.
type PgRepository interface {
	ListByGroupID(ctx context.Context, groupID string) ([]models.Channel, error)
}

type pgRepository struct {
	db  *bun.DB
	log logger.Logger
}

// NewPgRepository constructs a Postgres channel store.
func NewPgRepository(db *bun.DB, log logger.Logger) PgRepository {
	return pgRepository{db: db, log: log}
}

func (r pgRepository) ListByGroupID(ctx context.Context, groupID string) ([]models.Channel, error) {
	var rows []models.ChannelRow
	err := r.db.NewSelect().Model(&rows).
		Where("group_id = ?", groupID).
		OrderExpr("name ASC").
		Scan(ctx)
	if err != nil {
		r.logIfInfraErr("ListByGroupID", err)

		return nil, err
	}

	out := make([]models.Channel, len(rows))
	for i, row := range rows {
		out[i] = row.ToModel()
	}

	return out, nil
}

func (r pgRepository) logIfInfraErr(op string, err error) {
	if err == nil {
		return
	}
	r.log.Error("channel repository", "op", op, "err", err)
}

// InsertGeneralChannelTx creates the default #general channel inside an existing transaction.
func InsertGeneralChannelTx(ctx context.Context, tx bun.Tx, groupID, createdBy string, now time.Time) error {
	for range 5 {
		channelID, err := token.NewID()
		if err != nil {
			return err
		}
		publicID, err := token.NewPublicID()
		if err != nil {
			return err
		}
		row := models.ChannelRowFrom(models.Channel{
			ID: channelID, PublicID: publicID, GroupID: groupID,
			Type: constants.ChannelTypePublic, Name: constants.DefaultChannelName,
			CreatedBy: createdBy, CreatedAt: now, UpdatedAt: now,
		})
		_, err = tx.NewInsert().Model(&row).Exec(ctx)
		if err == nil {
			return nil
		}
		if errors.Is(pgErr.MapUnique(err), constants.ErrConflict) {
			continue
		}

		return err
	}

	return constants.ErrConflict
}
