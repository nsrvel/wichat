package repository

import (
	"context"
	"time"

	"github.com/uptrace/bun"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	pgErr "github.com/wichat/wichat/backend/ms-user/pkg/db"
)

// PgRepository is the Postgres auth session store.
type PgRepository interface {
	InsertSession(ctx context.Context, session models.Session, refresh models.RefreshToken) error
	FindRefreshByHash(ctx context.Context, hash []byte) (models.RefreshToken, error)
	FindSession(ctx context.Context, id string) (models.Session, error)
	MarkRefreshUsed(ctx context.Context, currentHash, nextHash []byte, nextID, sessionID string, expiresAt, now time.Time) error
	RevokeSession(ctx context.Context, sessionID string, now time.Time) error
	RevokeAllSessionsForUser(ctx context.Context, userID string, now time.Time) error
	DeleteExpired(ctx context.Context, now time.Time) error
}

type pgRepository struct {
	db *bun.DB
}

// NewPgRepository stores sessions and refresh tokens in Postgres.
func NewPgRepository(db *bun.DB) PgRepository {
	return pgRepository{db: db}
}

// InsertSession stores session and its initial refresh row in one transaction.
func (r pgRepository) InsertSession(ctx context.Context, session models.Session, refresh models.RefreshToken) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		sess := models.SessionRowFrom(session)
		if _, err := tx.NewInsert().Model(&sess).Exec(ctx); err != nil {
			return err
		}

		ref := models.RefreshRowFrom(refresh)
		_, err := tx.NewInsert().Model(&ref).Exec(ctx)

		return err
	})
}

func (r pgRepository) FindRefreshByHash(ctx context.Context, hash []byte) (models.RefreshToken, error) {
	var row models.RefreshRow
	err := r.db.NewSelect().Model(&row).Where("token_hash = ?", hash).Scan(ctx)
	if err != nil {
		return models.RefreshToken{}, pgErr.MapNotFound(err)
	}

	return row.ToModel(), nil
}

func (r pgRepository) FindSession(ctx context.Context, id string) (models.Session, error) {
	var row models.SessionRow
	err := r.db.NewSelect().Model(&row).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return models.Session{}, pgErr.MapNotFound(err)
	}

	return row.ToModel(), nil
}

// MarkRefreshUsed marks the current refresh hash used, inserts the successor, and extends session expiry; ErrAlreadyUsed if the row was already spent.
func (r pgRepository) MarkRefreshUsed(ctx context.Context, currentHash, nextHash []byte, nextID, sessionID string, expiresAt, now time.Time) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		var usedAt *time.Time
		err := tx.NewRaw(constants.SQLSelectRefreshUsedAtForUpdate, currentHash).Scan(ctx, &usedAt)
		if err != nil {
			return pgErr.MapNotFound(err)
		}
		if usedAt != nil {
			return constants.ErrAlreadyUsed
		}

		if _, err := tx.NewUpdate().Model((*models.RefreshRow)(nil)).
			Set("used_at = ?", now).
			Where("token_hash = ?", currentHash).
			Exec(ctx); err != nil {
			return err
		}

		rotated := models.RefreshRow{ID: nextID, SessionID: sessionID, TokenHash: nextHash, CreatedAt: now}
		if _, err := tx.NewInsert().Model(&rotated).Exec(ctx); err != nil {
			return err
		}

		_, err = tx.NewUpdate().Model((*models.SessionRow)(nil)).
			Set("expires_at = ?", expiresAt).
			Set("last_used_at = ?", now).
			Where("id = ?", sessionID).
			Exec(ctx)

		return err
	})
}

// RevokeSession sets revoked_at when the session is still active.
func (r pgRepository) RevokeSession(ctx context.Context, sessionID string, now time.Time) error {
	_, err := r.db.NewUpdate().Model((*models.SessionRow)(nil)).
		Set("revoked_at = ?", now).
		Where("id = ?", sessionID).
		Where("revoked_at IS NULL").
		Exec(ctx)

	return err
}

// RevokeAllSessionsForUser revokes every active session for the user.
func (r pgRepository) RevokeAllSessionsForUser(ctx context.Context, userID string, now time.Time) error {
	_, err := r.db.NewUpdate().Model((*models.SessionRow)(nil)).
		Set("revoked_at = ?", now).
		Where("user_id = ?", userID).
		Where("revoked_at IS NULL").
		Exec(ctx)

	return err
}

// DeleteExpired removes sessions past expires_at and their refresh tokens.
func (r pgRepository) DeleteExpired(ctx context.Context, now time.Time) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewRaw(constants.SQLDeleteExpiredRefreshTokens, now).Exec(ctx); err != nil {
			return err
		}

		_, err := tx.NewRaw(constants.SQLDeleteExpiredSessions, now).Exec(ctx)

		return err
	})
}
