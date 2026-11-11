package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/uptrace/bun"
	channelrepo "github.com/wichat/wichat/backend/ms-user/internal/channel/repository"
	grouprepo "github.com/wichat/wichat/backend/ms-user/internal/group/repository"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	pgErr "github.com/wichat/wichat/backend/ms-user/pkg/db"
)

// PgRepository is the Postgres user and workspace store.
type PgRepository interface {
	CountUsers(ctx context.Context) (int, error)
	GetInstanceSettings(ctx context.Context) (models.InstanceSettings, error)
	CompleteInstanceSetup(ctx context.Context, ownerUserID string, payload models.InstanceSetupPayload, now time.Time) error
	CreateFirstUser(ctx context.Context, user models.User) error
	CreateUser(ctx context.Context, user models.User) error
	FindUserByEmail(ctx context.Context, email string) (models.User, error)
	FindUserByUsername(ctx context.Context, username string) (models.User, error)
	FindUserByLogin(ctx context.Context, login string) (models.User, error)
	FindUserByID(ctx context.Context, id string) (models.User, error)
	UpdateUserPasswordHash(ctx context.Context, userID, hash string, now time.Time) error
	SaveGoogleUser(ctx context.Context, user models.User, identityID, provider, subject string) (models.User, error)
	SearchUsers(ctx context.Context, query string, limit int, excludeGroupID string) ([]models.UserSearchHit, error)
}

type pgRepository struct {
	db *bun.DB
}

// NewPgRepository stores users and instance policy in Postgres.
func NewPgRepository(db *bun.DB) PgRepository {
	return pgRepository{db: db}
}

// CountUsers returns how many rows exist in users.
func (r pgRepository) CountUsers(ctx context.Context) (int, error) {
	n, err := r.db.NewSelect().Model((*models.UserRow)(nil)).Count(ctx)

	return int(n), err
}

// GetInstanceSettings loads the singleton instance policy row.
func (r pgRepository) GetInstanceSettings(ctx context.Context) (models.InstanceSettings, error) {
	var row models.InstanceSettingsRow
	err := r.db.NewSelect().Model(&row).Where("id = ?", 1).Scan(ctx)
	if err != nil {
		return models.InstanceSettings{}, pgErr.MapNotFound(err)
	}

	return row.ToModel(), nil
}

// CompleteInstanceSetup creates the first group, owner membership, and instance policy.
func (r pgRepository) CompleteInstanceSetup(ctx context.Context, ownerUserID string, payload models.InstanceSetupPayload, now time.Time) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		n, err := tx.NewSelect().Model((*models.GroupRow)(nil)).Count(ctx)
		if err != nil {
			return err
		}
		if n > 0 {
			return constants.ErrConflict
		}

		res, err := tx.NewUpdate().Model(&models.InstanceSettingsRow{}).
			Set("username_enabled = ?", payload.UsernameEnabled).
			Set("registration_mode = ?", payload.RegistrationMode).
			Set("setup_completed = ?", true).
			Set("updated_at = ?", now).
			Where("id = ? AND setup_completed = ?", 1, false).
			Exec(ctx)
		if err != nil {
			return err
		}
		rows, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			return constants.ErrNotFound
		}

		group := models.Group{
			ID: payload.GroupID, PublicID: payload.GroupPublicID, Name: payload.GroupName, LogoURL: payload.GroupLogoURL,
			CreatedAt: now, UpdatedAt: now,
		}
		owner := models.GroupMember{
			GroupID: payload.GroupID, UserID: ownerUserID, Role: constants.GroupRoleOwner, CreatedAt: now,
		}

		if err := grouprepo.InsertGroupWithOwnerTx(ctx, tx, group, owner); err != nil {
			return err
		}

		return channelrepo.InsertDefaultChannelsTx(ctx, tx, group.ID, owner.UserID, now)
	})
}

// CreateFirstUser inserts the sole bootstrap admin under an exclusive lock; ErrConflict if any user already exists.
func (r pgRepository) CreateFirstUser(ctx context.Context, user models.User) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewRaw(constants.SQLLockUsers).Exec(ctx); err != nil {
			return err
		}

		n, err := tx.NewSelect().Model((*models.UserRow)(nil)).Count(ctx)
		if err != nil {
			return err
		}
		if n > 0 {
			return constants.ErrConflict
		}

		return insertUserWithCredential(ctx, tx, user)
	})
}

// CreateUser inserts a new account row.
func (r pgRepository) CreateUser(ctx context.Context, user models.User) error {
	err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		return insertUserWithCredential(ctx, tx, user)
	})

	return pgErr.MapUnique(err)
}

func (r pgRepository) FindUserByEmail(ctx context.Context, email string) (models.User, error) {
	return r.findUserBy(ctx, "email = ?", email)
}

func (r pgRepository) FindUserByUsername(ctx context.Context, username string) (models.User, error) {
	return r.findUserBy(ctx, "username = ?", username)
}

func (r pgRepository) FindUserByLogin(ctx context.Context, login string) (models.User, error) {
	login = strings.TrimSpace(login)
	if login == "" {
		return models.User{}, constants.ErrNotFound
	}

	var row models.UserRow
	err := r.db.NewSelect().Model(&row).
		Where("lower(email) = lower(?) OR lower(username) = lower(?)", login, login).
		Scan(ctx)
	if err != nil {
		return models.User{}, pgErr.MapNotFound(err)
	}

	return r.userRowToModel(ctx, row), nil
}

func (r pgRepository) FindUserByID(ctx context.Context, id string) (models.User, error) {
	return r.findUserBy(ctx, "id = ?", id)
}

// SearchUsers finds instance accounts by display name, username, or email prefix.
func (r pgRepository) SearchUsers(ctx context.Context, query string, limit int, excludeGroupID string) ([]models.UserSearchHit, error) {
	q := strings.TrimSpace(query)
	if q == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = constants.DefaultUserSearchLimit
	}
	if limit > constants.MaxUserSearchLimit {
		limit = constants.MaxUserSearchLimit
	}

	pattern := "%" + strings.ReplaceAll(strings.ToLower(q), "%", "") + "%"
	sel := r.db.NewSelect().
		TableExpr("users AS u").
		ColumnExpr("u.id, u.display_name, u.email, u.username").
		Where("u.is_guest = false").
		Where(
			"lower(u.display_name) LIKE ? OR lower(coalesce(u.email, '')) LIKE ? OR lower(coalesce(u.username, '')) LIKE ?",
			pattern, pattern, pattern,
		).
		OrderExpr("u.display_name ASC").
		Limit(int64(limit))

	if strings.TrimSpace(excludeGroupID) != "" {
		sel = sel.Where(
			"u.id NOT IN (SELECT user_id FROM group_members WHERE group_id = ?)",
			excludeGroupID,
		)
	}

	var rows []struct {
		ID          string  `bun:"id"`
		DisplayName string  `bun:"display_name"`
		Email       *string `bun:"email"`
		Username    *string `bun:"username"`
	}
	if err := sel.Scan(ctx, &rows); err != nil {
		return nil, err
	}

	out := make([]models.UserSearchHit, len(rows))
	for i, row := range rows {
		email := ""
		if row.Email != nil {
			email = *row.Email
		}
		username := ""
		if row.Username != nil {
			username = *row.Username
		}
		out[i] = models.UserSearchHit{
			UserID:      row.ID,
			DisplayName: row.DisplayName,
			Email:       email,
			Username:    username,
		}
	}

	return out, nil
}

// UpdateUserPasswordHash replaces the password hash for a member account.
func (r pgRepository) UpdateUserPasswordHash(ctx context.Context, userID, hash string, now time.Time) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewUpdate().Model((*models.UserRow)(nil)).
			Set("updated_at = ?", now).
			Where("id = ?", userID).
			Where("is_guest = ?", false).
			Exec(ctx); err != nil {
			return err
		}

		res, err := tx.NewUpdate().Model((*models.UserCredentialRow)(nil)).
			Set("password_hash = ?", hash).
			Set("updated_at = ?", now).
			Where("user_id = ?", userID).
			Exec(ctx)
		if err != nil {
			return err
		}
		n, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if n > 0 {
			return nil
		}

		_, err = tx.NewInsert().Model(&models.UserCredentialRow{
			UserID: userID, PasswordHash: hash, UpdatedAt: now,
		}).Exec(ctx)

		return err
	})
}

// SaveGoogleUser returns an existing user by provider+subject, links identity to an email match, or inserts a new user.
func (r pgRepository) SaveGoogleUser(ctx context.Context, user models.User, identityID, provider, subject string) (models.User, error) {
	var out models.User
	err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewRaw(constants.SQLLockUsers).Exec(ctx); err != nil {
			return err
		}

		linked, err := findUserByIdentity(ctx, tx, provider, subject)
		if err == nil {
			out = linked

			return nil
		}
		if !errors.Is(err, constants.ErrNotFound) {
			return err
		}

		existing, err := findUserByEmailTx(ctx, tx, user.Email)
		if err == nil {
			if err := insertLoginConnection(ctx, tx, identityID, existing.ID, provider, subject, user.CreatedAt); err != nil {
				return err
			}
			out = existing

			return nil
		}
		if !errors.Is(err, constants.ErrNotFound) {
			return err
		}

		row := models.UserRowFrom(user)
		if _, err := tx.NewInsert().Model(&row).Exec(ctx); err != nil {
			return err
		}
		if err := insertLoginConnection(ctx, tx, identityID, user.ID, provider, subject, user.CreatedAt); err != nil {
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

func (r pgRepository) findUserBy(ctx context.Context, where, arg string) (models.User, error) {
	var row models.UserRow
	err := r.db.NewSelect().Model(&row).Where(where, arg).Scan(ctx)
	if err != nil {
		return models.User{}, pgErr.MapNotFound(err)
	}

	return r.userRowToModel(ctx, row), nil
}

func (r pgRepository) userRowToModel(ctx context.Context, row models.UserRow) models.User {
	pass, err := r.loadPasswordHash(ctx, row.ID)
	if err != nil {
		pass = ""
	}

	return row.ToModel(pass)
}

func (r pgRepository) loadPasswordHash(ctx context.Context, userID string) (string, error) {
	var cred models.UserCredentialRow
	err := r.db.NewSelect().Model(&cred).Where("user_id = ?", userID).Scan(ctx)
	if err != nil {
		return "", pgErr.MapNotFound(err)
	}

	return cred.PasswordHash, nil
}

func findUserByIdentity(ctx context.Context, tx bun.Tx, provider, subject string) (models.User, error) {
	var conn models.UserConnectionRow
	err := tx.NewSelect().Model(&conn).
		Where("kind = ?", constants.ConnectionKindLogin).
		Where("provider = ?", provider).
		Where("subject = ?", subject).
		Scan(ctx)
	if err != nil {
		return models.User{}, pgErr.MapNotFound(err)
	}

	return findUserByIDTx(ctx, tx, conn.UserID)
}

func findUserByIDTx(ctx context.Context, tx bun.Tx, id string) (models.User, error) {
	var row models.UserRow
	err := tx.NewSelect().Model(&row).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return models.User{}, pgErr.MapNotFound(err)
	}

	var cred models.UserCredentialRow
	pass := ""
	if err := tx.NewSelect().Model(&cred).Where("user_id = ?", id).Scan(ctx); err == nil {
		pass = cred.PasswordHash
	}

	return row.ToModel(pass), nil
}

func findUserByEmailTx(ctx context.Context, tx bun.Tx, email string) (models.User, error) {
	var row models.UserRow
	err := tx.NewSelect().Model(&row).Where("email = ?", email).Scan(ctx)
	if err != nil {
		return models.User{}, pgErr.MapNotFound(err)
	}

	return findUserByIDTx(ctx, tx, row.ID)
}

func insertLoginConnection(ctx context.Context, tx bun.Tx, id, userID, provider, subject string, at time.Time) error {
	row := models.UserConnectionRow{
		ID: id, UserID: userID, Kind: constants.ConnectionKindLogin,
		Provider: provider, Subject: subject, Scopes: []string{},
		CreatedAt: at, UpdatedAt: at,
	}
	_, err := tx.NewInsert().Model(&row).Exec(ctx)

	return err
}

func insertUserWithCredential(ctx context.Context, tx bun.Tx, user models.User) error {
	row := models.UserRowFrom(user)
	if _, err := tx.NewInsert().Model(&row).Exec(ctx); err != nil {
		return err
	}
	if user.PasswordHash == "" {
		return nil
	}

	_, err := tx.NewInsert().Model(&models.UserCredentialRow{
		UserID: user.ID, PasswordHash: user.PasswordHash, UpdatedAt: user.UpdatedAt,
	}).Exec(ctx)

	return err
}
