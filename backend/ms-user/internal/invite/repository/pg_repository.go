package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/uptrace/bun"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	pgErr "github.com/wichat/wichat/backend/ms-user/pkg/db"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

// PgRepository stores group_invites rows.
type PgRepository interface {
	InsertInvite(ctx context.Context, row models.GroupInviteRow) error
	FindOpenShareLink(ctx context.Context, groupID string, channelID *string, now time.Time) (models.GroupInviteRow, error)
	FindByToken(ctx context.Context, token string) (models.GroupInviteRow, error)
	MarkRedeemed(ctx context.Context, id, userID string, now time.Time) error
}

type pgRepository struct {
	db  *bun.DB
	log logger.Logger
}

// NewPgRepository stores invite rows in Postgres.
func NewPgRepository(db *bun.DB, log logger.Logger) PgRepository {
	return pgRepository{db: db, log: log}
}

// InsertInvite creates a group_invites row.
func (r pgRepository) InsertInvite(ctx context.Context, row models.GroupInviteRow) error {
	_, err := r.db.NewInsert().Model(&row).Exec(ctx)
	err = pgErr.MapUnique(err)
	r.logIfInfraErr("InsertInvite", err)

	return err
}

// FindOpenShareLink returns a non-expired reusable share link for a group (and optional channel).
func (r pgRepository) FindOpenShareLink(ctx context.Context, groupID string, channelID *string, now time.Time) (models.GroupInviteRow, error) {
	q := r.db.NewSelect().Model((*models.GroupInviteRow)(nil)).
		Where("group_id = ?", groupID).
		Where("kind = ?", constants.GroupInviteKindShareLink).
		Where("redeemed_at IS NULL").
		Where("expires_at > ?", now).
		Where("LENGTH(token) = ?", constants.GroupInviteTokenLen).
		OrderExpr("created_at DESC").
		Limit(1)

	if channelID == nil || strings.TrimSpace(*channelID) == "" {
		q = q.Where("channel_id IS NULL")
	} else {
		q = q.Where("channel_id = ?", *channelID)
	}

	var row models.GroupInviteRow
	if err := q.Scan(ctx, &row); err != nil {
		err = pgErr.MapNotFound(err)
		r.logIfInfraErr("FindOpenShareLink", err)

		return models.GroupInviteRow{}, err
	}

	return row, nil
}

// FindByToken loads an invite by opaque token.
func (r pgRepository) FindByToken(ctx context.Context, token string) (models.GroupInviteRow, error) {
	token = strings.TrimSpace(token)
	var row models.GroupInviteRow
	err := r.db.NewSelect().Model(&row).
		Where("token = ?", token).
		Scan(ctx)
	if err != nil {
		err = pgErr.MapNotFound(err)
		r.logIfInfraErr("FindByToken", err)

		return models.GroupInviteRow{}, err
	}

	return row, nil
}

// MarkRedeemed records invite consumption.
func (r pgRepository) MarkRedeemed(ctx context.Context, id, userID string, now time.Time) error {
	res, err := r.db.NewUpdate().Model((*models.GroupInviteRow)(nil)).
		Set("redeemed_at = ?", now).
		Set("redeemed_by_user_id = ?", userID).
		Where("id = ?", id).
		Where("redeemed_at IS NULL").
		Exec(ctx)
	if err != nil {
		r.logIfInfraErr("MarkRedeemed", err)

		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		r.logIfInfraErr("MarkRedeemed", err)

		return err
	}
	if n == 0 {
		return constants.ErrNotFound
	}

	return nil
}

func (r pgRepository) logIfInfraErr(op string, err error) {
	if err == nil || errors.Is(err, constants.ErrNotFound) || errors.Is(err, constants.ErrConflict) {
		return
	}
	r.log.Error("invite repository", "op", op, "err", err)
}
