package repository

import (
	"context"
	"errors"
	"time"

	"github.com/uptrace/bun"
	"github.com/wichat/wichat/backend/ms-chat/internal/models"
	"github.com/wichat/wichat/backend/ms-chat/pkg/constants"
	pgErr "github.com/wichat/wichat/backend/ms-chat/pkg/db"
	"github.com/wichat/wichat/backend/ms-chat/pkg/token"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

// PgRepository persists channel messages.
type PgRepository interface {
	ListByChannel(ctx context.Context, channelID string, limit int, before *time.Time) ([]models.MessageWithAuthor, error)
	Insert(ctx context.Context, channelID, authorID, body string, createdAt time.Time) (models.Message, error)
	DeleteByID(ctx context.Context, id string) error
	FindUserDisplayName(ctx context.Context, userID string) (string, error)
}

type pgRepository struct {
	db  *bun.DB
	log logger.Logger
}

// NewPgRepository constructs a Postgres message store.
func NewPgRepository(db *bun.DB, log logger.Logger) PgRepository {
	return pgRepository{db: db, log: log}
}

func (r pgRepository) ListByChannel(ctx context.Context, channelID string, limit int, before *time.Time) ([]models.MessageWithAuthor, error) {
	if limit <= 0 {
		limit = constants.DefaultMessageListLimit
	}

	var rows []models.MessageListRow
	q := r.db.NewSelect().Model(&rows).
		ColumnExpr("m.id, m.public_id, m.channel_id, m.author_id, m.body, m.created_at").
		ColumnExpr("u.display_name AS author_display_name").
		Join("JOIN users u ON u.id = m.author_id").
		Where("m.channel_id = ?", channelID).
		OrderExpr("m.created_at ASC").
		Limit(int64(limit))
	if before != nil {
		q = q.Where("m.created_at < ?", *before)
	}

	err := q.Scan(ctx)
	if err != nil {
		r.logIfInfraErr("ListByChannel", err)

		return nil, err
	}

	out := make([]models.MessageWithAuthor, len(rows))
	for i, row := range rows {
		out[i] = row.ToMessageWithAuthor()
	}

	return out, nil
}

func (r pgRepository) Insert(ctx context.Context, channelID, authorID, body string, createdAt time.Time) (models.Message, error) {
	for range 5 {
		id, err := token.NewID()
		if err != nil {
			return models.Message{}, err
		}
		publicID, err := token.NewPublicID()
		if err != nil {
			return models.Message{}, err
		}

		row := models.MessageRowFrom(models.Message{
			ID: id, PublicID: publicID, ChannelID: channelID, AuthorID: authorID,
			Body: body, CreatedAt: createdAt,
		})
		_, err = r.db.NewInsert().Model(&row).Exec(ctx)
		if err == nil {
			return row.ToModel(), nil
		}
		if errors.Is(pgErr.MapUnique(err), constants.ErrConflict) {
			continue
		}

		r.logIfInfraErr("Insert", err)

		return models.Message{}, err
	}

	return models.Message{}, constants.ErrConflict
}

func (r pgRepository) DeleteByID(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().
		Model((*models.MessageRow)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		r.logIfInfraErr("DeleteByID", err)
		return err
	}

	return nil
}

func (r pgRepository) FindUserDisplayName(ctx context.Context, userID string) (string, error) {
	var displayName string
	err := r.db.NewSelect().
		TableExpr("users").
		Column("display_name").
		Where("id = ?", userID).
		Scan(ctx, &displayName)
	if err != nil {
		r.logIfInfraErr("FindUserDisplayName", err)

		return "", pgErr.MapNotFound(err)
	}

	return displayName, nil
}

func (r pgRepository) logIfInfraErr(op string, err error) {
	if err == nil {
		return
	}
	r.log.Error("message repository", "op", op, "err", err)
}
