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
)

// PgRepository is the Postgres user and workspace store.
type PgRepository interface {
	CountUsers(ctx context.Context) (int, error)
	GetInstanceSettings(ctx context.Context) (models.InstanceSettings, error)
	GetFirstGroup(ctx context.Context) (models.Group, error)
	CompleteInstanceSetup(ctx context.Context, ownerUserID string, payload models.InstanceSetupPayload, now time.Time) error
	CreateFirstUser(ctx context.Context, user models.User) error
	CreateUser(ctx context.Context, user models.User) error
	FindUserByEmail(ctx context.Context, email string) (models.User, error)
	FindUserByUsername(ctx context.Context, username string) (models.User, error)
	FindUserByLogin(ctx context.Context, login string) (models.User, error)
	FindUserByID(ctx context.Context, id string) (models.User, error)
	UpdateUserPasswordHash(ctx context.Context, userID, hash string, now time.Time) error
	SaveGoogleUser(ctx context.Context, user models.User, identityID, provider, subject string) (models.User, error)
	ListGroupsForUser(ctx context.Context, userID string) ([]models.Group, error)
	ListJoinableGroups(ctx context.Context, userID string) ([]models.Group, error)
	FindGroupByID(ctx context.Context, id string) (models.Group, error)
	FindGroupMember(ctx context.Context, groupID, userID string) (models.GroupMember, error)
	CreateGroupWithOwner(ctx context.Context, group models.Group, owner models.GroupMember) error
	InsertGroupMember(ctx context.Context, member models.GroupMember) error
}

type pgRepository struct {
	db *bun.DB
}

// NewPgRepository stores users, groups, and instance policy in Postgres.
func NewPgRepository(db *bun.DB) PgRepository {
	return pgRepository{db: db}
}

// CountUsers returns how many rows exist in auth.users.
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

// GetFirstGroup returns the earliest group (bootstrap group for setup status).
func (r pgRepository) GetFirstGroup(ctx context.Context) (models.Group, error) {
	var row models.GroupRow
	err := r.db.NewSelect().Model(&row).Order("created_at ASC").Limit(1).Scan(ctx)
	if err != nil {
		return models.Group{}, pgErr.MapNotFound(err)
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
			Set("guest_enabled = ?", payload.GuestEnabled).
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

		group := models.GroupRowFrom(models.Group{
			ID:        payload.GroupID,
			Name:      payload.GroupName,
			LogoURL:   payload.GroupLogoURL,
			CreatedAt: now,
			UpdatedAt: now,
		})
		if _, err := tx.NewInsert().Model(&group).Exec(ctx); err != nil {
			return err
		}

		member := models.GroupMemberRowFrom(models.GroupMember{
			GroupID:   payload.GroupID,
			UserID:    ownerUserID,
			Role:      constants.GroupRoleOwner,
			CreatedAt: now,
		})
		_, err = tx.NewInsert().Model(&member).Exec(ctx)

		return err
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

		row := models.UserRowFrom(user)
		_, err = tx.NewInsert().Model(&row).Exec(ctx)

		return err
	})
}

// CreateUser inserts a new account row.
func (r pgRepository) CreateUser(ctx context.Context, user models.User) error {
	row := models.UserRowFrom(user)
	_, err := r.db.NewInsert().Model(&row).Exec(ctx)

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

	return row.ToModel(), nil
}

func (r pgRepository) FindUserByID(ctx context.Context, id string) (models.User, error) {
	return r.findUserBy(ctx, "id = ?", id)
}

// UpdateUserPasswordHash replaces the password hash for a member account.
func (r pgRepository) UpdateUserPasswordHash(ctx context.Context, userID, hash string, now time.Time) error {
	_, err := r.db.NewUpdate().Model((*models.UserRow)(nil)).
		Set("password_hash = ?", hash).
		Set("updated_at = ?", now).
		Where("id = ?", userID).
		Where("is_guest = ?", false).
		Exec(ctx)

	return err
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
			if err := insertIdentity(ctx, tx, identityID, existing.ID, provider, subject, user.CreatedAt); err != nil {
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

// ListGroupsForUser returns groups the user belongs to, ordered by name.
func (r pgRepository) ListGroupsForUser(ctx context.Context, userID string) ([]models.Group, error) {
	var rows []models.GroupRow
	err := r.db.NewSelect().
		Model(&rows).
		Where("id IN (SELECT group_id FROM auth.group_members WHERE user_id = ?)", userID).
		Order("name ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]models.Group, len(rows))
	for i, row := range rows {
		out[i] = row.ToModel()
	}

	return out, nil
}

// ListJoinableGroups returns instance groups the user is not in.
func (r pgRepository) ListJoinableGroups(ctx context.Context, userID string) ([]models.Group, error) {
	var rows []models.GroupRow
	err := r.db.NewSelect().
		Model(&rows).
		Where("id NOT IN (SELECT group_id FROM auth.group_members WHERE user_id = ?)", userID).
		Order("name ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]models.Group, len(rows))
	for i, row := range rows {
		out[i] = row.ToModel()
	}

	return out, nil
}

// FindGroupByID loads one group by id.
func (r pgRepository) FindGroupByID(ctx context.Context, id string) (models.Group, error) {
	var row models.GroupRow
	err := r.db.NewSelect().Model(&row).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return models.Group{}, pgErr.MapNotFound(err)
	}

	return row.ToModel(), nil
}

// FindGroupMember loads membership for a user in a group.
func (r pgRepository) FindGroupMember(ctx context.Context, groupID, userID string) (models.GroupMember, error) {
	var row models.GroupMemberRow
	err := r.db.NewSelect().Model(&row).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Scan(ctx)
	if err != nil {
		return models.GroupMember{}, pgErr.MapNotFound(err)
	}

	return models.GroupMember{
		GroupID: row.GroupID, UserID: row.UserID, Role: row.Role, CreatedAt: row.CreatedAt,
	}, nil
}

// CreateGroupWithOwner inserts a group and owner membership in one transaction.
func (r pgRepository) CreateGroupWithOwner(ctx context.Context, group models.Group, owner models.GroupMember) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		gRow := models.GroupRowFrom(group)
		if _, err := tx.NewInsert().Model(&gRow).Exec(ctx); err != nil {
			return err
		}
		mRow := models.GroupMemberRowFrom(owner)
		_, err := tx.NewInsert().Model(&mRow).Exec(ctx)

		return err
	})
}

// InsertGroupMember adds a user to a group.
func (r pgRepository) InsertGroupMember(ctx context.Context, member models.GroupMember) error {
	row := models.GroupMemberRowFrom(member)
	_, err := r.db.NewInsert().Model(&row).Exec(ctx)

	return pgErr.MapUnique(err)
}

func (r pgRepository) findUserBy(ctx context.Context, where, arg string) (models.User, error) {
	var row models.UserRow
	err := r.db.NewSelect().Model(&row).Where(where, arg).Scan(ctx)
	if err != nil {
		return models.User{}, pgErr.MapNotFound(err)
	}

	return row.ToModel(), nil
}

func findUserByIdentity(ctx context.Context, tx bun.Tx, provider, subject string) (models.User, error) {
	var row models.UserRow
	err := tx.NewRaw(constants.SQLSelectUserByIdentity, provider, subject).Scan(ctx, &row)
	if err != nil {
		return models.User{}, pgErr.MapNotFound(err)
	}

	return row.ToModel(), nil
}

func findUserByEmailTx(ctx context.Context, tx bun.Tx, email string) (models.User, error) {
	var row models.UserRow
	err := tx.NewSelect().Model(&row).Where("email = ?", email).Scan(ctx)
	if err != nil {
		return models.User{}, pgErr.MapNotFound(err)
	}

	return row.ToModel(), nil
}

func insertIdentity(ctx context.Context, tx bun.Tx, id, userID, provider, subject string, createdAt time.Time) error {
	row := models.IdentityRow{ID: id, UserID: userID, Provider: provider, Subject: subject, CreatedAt: createdAt}
	_, err := tx.NewInsert().Model(&row).Exec(ctx)

	return err
}
