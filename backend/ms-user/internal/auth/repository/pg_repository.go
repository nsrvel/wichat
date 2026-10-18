package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/uptrace/bun"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
)

// Sentinel errors mapped by the auth service. No exception types here.
var (
	ErrNotFound    = errors.New("auth: not found")
	ErrConflict    = errors.New("auth: conflict")
	ErrAlreadyUsed = errors.New("auth: refresh already used")
)

// PgRepository is the Postgres auth store.
type PgRepository interface {
	CountUsers(ctx context.Context) (int, error)
	CreateFirstUser(ctx context.Context, user models.User) error
	FindUserByEmail(ctx context.Context, email string) (models.User, error)
	FindUserByID(ctx context.Context, id string) (models.User, error)
	InsertSession(ctx context.Context, session models.Session, refresh models.RefreshToken) error
	FindRefreshByHash(ctx context.Context, hash []byte) (models.RefreshToken, error)
	FindSession(ctx context.Context, id string) (models.Session, error)
	MarkRefreshUsed(ctx context.Context, currentHash, nextHash []byte, nextID, sessionID string, expiresAt, now time.Time) error
	RevokeSession(ctx context.Context, sessionID string, now time.Time) error
	DeleteExpired(ctx context.Context, now time.Time) error
	SaveGoogleUser(ctx context.Context, user models.User, identityID, provider, subject string) (models.User, error)
}

// Repository is the persistence boundary used by auth service (Postgres today).
type Repository = PgRepository

type pgRepository struct {
	db *bun.DB
}

// NewPgRepository stores accounts and sessions in Postgres.
func NewPgRepository(db *bun.DB) PgRepository {
	return pgRepository{db: db}
}

// CountUsers returns how many rows exist in auth.users.
func (r pgRepository) CountUsers(ctx context.Context) (int, error) {
	n, err := r.db.NewSelect().Model((*models.UserRow)(nil)).Count(ctx)

	return int(n), err
}

// CreateFirstUser inserts the sole bootstrap admin under an exclusive lock; ErrConflict if any user already exists.
func (r pgRepository) CreateFirstUser(ctx context.Context, user models.User) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewRaw(sqlLockUsers).Exec(ctx); err != nil {
			return err
		}

		n, err := tx.NewSelect().Model((*models.UserRow)(nil)).Count(ctx)
		if err != nil {
			return err
		}
		if n > 0 {
			return ErrConflict
		}

		row := models.UserRowFrom(user)
		_, err = tx.NewInsert().Model(&row).Exec(ctx)

		return err
	})
}

func (r pgRepository) FindUserByEmail(ctx context.Context, email string) (models.User, error) {
	return r.findUserBy(ctx, "email = ?", email)
}

func (r pgRepository) FindUserByID(ctx context.Context, id string) (models.User, error) {
	return r.findUserBy(ctx, "id = ?", id)
}

func (r pgRepository) findUserBy(ctx context.Context, where, arg string) (models.User, error) {
	var row models.UserRow
	err := r.db.NewSelect().Model(&row).Where(where, arg).Scan(ctx)
	if err != nil {
		return models.User{}, mapNotFound(err)
	}

	return row.ToModel(), nil
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
		return models.RefreshToken{}, mapNotFound(err)
	}

	return row.ToModel(), nil
}

func (r pgRepository) FindSession(ctx context.Context, id string) (models.Session, error) {
	var row models.SessionRow
	err := r.db.NewSelect().Model(&row).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return models.Session{}, mapNotFound(err)
	}

	return row.ToModel(), nil
}

// MarkRefreshUsed marks the current refresh hash used, inserts the successor, and extends session expiry; ErrAlreadyUsed if the row was already spent.
func (r pgRepository) MarkRefreshUsed(ctx context.Context, currentHash, nextHash []byte, nextID, sessionID string, expiresAt, now time.Time) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		var usedAt *time.Time
		err := tx.NewRaw(sqlSelectRefreshUsedAtForUpdate, currentHash).Scan(ctx, &usedAt)
		if err != nil {
			return mapNotFound(err)
		}
		if usedAt != nil {
			return ErrAlreadyUsed
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

// DeleteExpired removes sessions past expires_at and their refresh tokens.
func (r pgRepository) DeleteExpired(ctx context.Context, now time.Time) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewRaw(sqlDeleteExpiredRefreshTokens, now).Exec(ctx); err != nil {
			return err
		}

		_, err := tx.NewRaw(sqlDeleteExpiredSessions, now).Exec(ctx)

		return err
	})
}

// SaveGoogleUser returns an existing user by provider+subject, links identity to an email match, or inserts a new user.
func (r pgRepository) SaveGoogleUser(ctx context.Context, user models.User, identityID, provider, subject string) (models.User, error) {
	var out models.User
	err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewRaw(sqlLockUsers).Exec(ctx); err != nil {
			return err
		}

		linked, err := findUserByIdentity(ctx, tx, provider, subject)
		if err == nil {
			out = linked

			return nil
		}
		if !errors.Is(err, ErrNotFound) {
			return err
		}

		existing, err := findUserByEmailTx(ctx, tx, user.Email)
		if err == nil {
			if err := insertIdentity(ctx, tx, identityID, existing.ID, provider, subject, user.CreatedAt); err != nil {
				return err
			}
			out = existing

			return nil
		}
		if !errors.Is(err, ErrNotFound) {
			return err
		}

		row := models.UserRowFrom(user)
		if _, err := tx.NewInsert().Model(&row).Exec(ctx); err != nil {
			return err
		}
		if err := insertIdentity(ctx, tx, identityID, user.ID, provider, subject, user.CreatedAt); err != nil {
			return err
		}
		out = user

		return nil
	})
	if err != nil {
		return models.User{}, err
	}

	return out, nil
}

func findUserByIdentity(ctx context.Context, tx bun.Tx, provider, subject string) (models.User, error) {
	var row models.UserRow
	err := tx.NewRaw(sqlSelectUserByIdentity, provider, subject).Scan(ctx, &row)
	if err != nil {
		return models.User{}, mapNotFound(err)
	}

	return row.ToModel(), nil
}

func findUserByEmailTx(ctx context.Context, tx bun.Tx, email string) (models.User, error) {
	var row models.UserRow
	err := tx.NewSelect().Model(&row).Where("email = ?", email).Scan(ctx)
	if err != nil {
		return models.User{}, mapNotFound(err)
	}

	return row.ToModel(), nil
}

func insertIdentity(ctx context.Context, tx bun.Tx, id, userID, provider, subject string, createdAt time.Time) error {
	row := models.IdentityRow{ID: id, UserID: userID, Provider: provider, Subject: subject, CreatedAt: createdAt}
	_, err := tx.NewInsert().Model(&row).Exec(ctx)

	return err
}

func mapNotFound(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}

	return err
}
