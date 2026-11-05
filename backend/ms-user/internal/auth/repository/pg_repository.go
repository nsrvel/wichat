package repository

import (
	"context"
	"time"

	"github.com/uptrace/bun"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	pgErr "github.com/wichat/wichat/backend/ms-user/pkg/db"
)

// SessionRepository is the Postgres user session store (device rows only).
type SessionRepository interface {
	InsertSession(ctx context.Context, session models.Session) error
	FindSession(ctx context.Context, id string) (models.Session, error)
	FindSessionByRefreshHash(ctx context.Context, hash []byte, now time.Time) (models.Session, error)
	TouchSession(ctx context.Context, sessionID, refreshID string, refreshTokenHash []byte, expiresAt, lastUsedAt time.Time) error
	ListSessionIDsForUser(ctx context.Context, userID string) ([]string, error)
	RevokeSession(ctx context.Context, sessionID string, now time.Time) error
	RevokeAllSessionsForUser(ctx context.Context, userID string, now time.Time) error
	DeleteExpired(ctx context.Context, now time.Time) error
}

type pgSessionRepository struct {
	db *bun.DB
}

// NewSessionRepository stores user_sessions in Postgres.
func NewSessionRepository(db *bun.DB) SessionRepository {
	return pgSessionRepository{db: db}
}

func (r pgSessionRepository) InsertSession(ctx context.Context, session models.Session) error {
	sess := models.SessionRowFrom(session)
	_, err := r.db.NewInsert().Model(&sess).Exec(ctx)

	return err
}

func (r pgSessionRepository) FindSession(ctx context.Context, id string) (models.Session, error) {
	var row models.SessionRow
	err := r.db.NewSelect().Model(&row).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return models.Session{}, pgErr.MapNotFound(err)
	}

	return row.ToModel(), nil
}

func (r pgSessionRepository) FindSessionByRefreshHash(ctx context.Context, hash []byte, now time.Time) (models.Session, error) {
	var row models.SessionRow
	err := r.db.NewSelect().Model(&row).
		Where("refresh_token_hash = ?", hash).
		Where("revoked_at IS NULL").
		Where("expires_at > ?", now).
		Scan(ctx)
	if err != nil {
		return models.Session{}, pgErr.MapNotFound(err)
	}

	return row.ToModel(), nil
}

func (r pgSessionRepository) TouchSession(ctx context.Context, sessionID, refreshID string, refreshTokenHash []byte, expiresAt, lastUsedAt time.Time) error {
	_, err := r.db.NewUpdate().Model((*models.SessionRow)(nil)).
		Set("refresh_id = ?", refreshID).
		Set("refresh_token_hash = ?", refreshTokenHash).
		Set("expires_at = ?", expiresAt).
		Set("last_used_at = ?", lastUsedAt).
		Where("id = ?", sessionID).
		Exec(ctx)

	return err
}

func (r pgSessionRepository) ListSessionIDsForUser(ctx context.Context, userID string) ([]string, error) {
	var ids []string
	err := r.db.NewSelect().
		Model((*models.SessionRow)(nil)).
		Column("id").
		Where("user_id = ?", userID).
		Where("revoked_at IS NULL").
		Scan(ctx, &ids)

	return ids, err
}

func (r pgSessionRepository) RevokeSession(ctx context.Context, sessionID string, now time.Time) error {
	_, err := r.db.NewUpdate().Model((*models.SessionRow)(nil)).
		Set("revoked_at = ?", now).
		Where("id = ?", sessionID).
		Where("revoked_at IS NULL").
		Exec(ctx)

	return err
}

func (r pgSessionRepository) RevokeAllSessionsForUser(ctx context.Context, userID string, now time.Time) error {
	_, err := r.db.NewUpdate().Model((*models.SessionRow)(nil)).
		Set("revoked_at = ?", now).
		Where("user_id = ?", userID).
		Where("revoked_at IS NULL").
		Exec(ctx)

	return err
}

func (r pgSessionRepository) DeleteExpired(ctx context.Context, now time.Time) error {
	_, err := r.db.NewDelete().Model((*models.SessionRow)(nil)).
		Where("expires_at <= ?", now).
		Exec(ctx)

	return err
}

// PgRepository is kept as an alias bundle for wiring that still expects one auth repo type.
type PgRepository = SessionRepository

// NewPgRepository is deprecated naming; prefer NewSessionRepository.
func NewPgRepository(db *bun.DB) PgRepository {
	return NewSessionRepository(db)
}
