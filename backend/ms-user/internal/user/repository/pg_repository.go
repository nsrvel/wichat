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
	"github.com/wichat/wichat/backend/ms-user/pkg/token"
)

// PgRepository is the Postgres user and workspace store.
type PgRepository interface {
	CountUsers(ctx context.Context) (int, error)
	CompleteInstanceSetup(ctx context.Context, ownerUserID string, payload models.InstanceSetupPayload, now time.Time) error
	CreateFirstUser(ctx context.Context, user models.User) error
	CreateUser(ctx context.Context, user models.User) error
	FindUserByEmail(ctx context.Context, email string) (models.User, error)
	FindUserByUsername(ctx context.Context, username string) (models.User, error)
	FindUserByLogin(ctx context.Context, login string) (models.User, error)
	FindUserByID(ctx context.Context, id string) (models.User, error)
	UpdateUserPasswordHash(ctx context.Context, userID, hash string, now time.Time) error
	UpdateUserAvatarURL(ctx context.Context, userID, avatarURL string, now time.Time) error
	UpdateUserProfile(ctx context.Context, userID string, patch models.UserProfilePatch, now time.Time) error
	UpdateUserStatusNote(ctx context.Context, userID string, patch models.UserStatusNotePatch, now time.Time) error
	UpdateUserPreferences(ctx context.Context, userID string, patch models.UserPreferencesPatch, now time.Time) error
	UpdateUsername(ctx context.Context, userID, username string, now time.Time) error
	UpdateEmail(ctx context.Context, userID, email string, now time.Time) error
	DisableAccount(ctx context.Context, userID string, now time.Time) error
	SoftDeleteAccount(ctx context.Context, userID, anonymizedUsername string, now time.Time) error
	SaveGoogleUser(ctx context.Context, user models.User, identityID, provider, subject string) (models.User, error)
	SearchUsers(ctx context.Context, query string, limit int, excludeGroupID string) ([]models.UserSearchHit, error)
	GetMFATOTP(ctx context.Context, userID string) (models.MFATOTP, error)
	UpsertMFATOTP(ctx context.Context, totp models.MFATOTP) error
	UpdateMFATOTPLastUsedStep(ctx context.Context, userID string, step int64, now time.Time) error
	DeleteMFATOTP(ctx context.Context, userID string) error
	ListWebAuthnCredentials(ctx context.Context, userID string) ([]models.WebAuthnCredential, error)
	InsertWebAuthnCredential(ctx context.Context, cred models.WebAuthnCredential) error
	UpdateWebAuthnCredentialUsage(ctx context.Context, credentialID []byte, signCount uint32, cloneWarning, backupState bool, now time.Time) error
	RenameWebAuthnCredential(ctx context.Context, userID, id, name string) (models.WebAuthnCredential, error)
	DeleteWebAuthnCredential(ctx context.Context, userID, id string) error
	CountUnusedRecoveryCodes(ctx context.Context, userID string) (int32, error)
	ReplaceRecoveryCodes(ctx context.Context, userID string, codeHashes [][]byte, now time.Time) error
	ConsumeRecoveryCode(ctx context.Context, userID string, codeHash []byte, now time.Time) (bool, error)
	ListRelationships(ctx context.Context, userID string) ([]models.UserRelationship, error)
	UpsertRelationship(ctx context.Context, userID, targetUserID, kind string, now time.Time) error
	DeleteRelationship(ctx context.Context, userID, targetUserID string) error
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

// CompleteInstanceSetup creates the first group and owner membership when none exist yet.
func (r pgRepository) CompleteInstanceSetup(ctx context.Context, ownerUserID string, payload models.InstanceSetupPayload, now time.Time) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		n, err := tx.NewSelect().Model((*models.GroupRow)(nil)).Count(ctx)
		if err != nil {
			return err
		}
		if n > 0 {
			return constants.ErrConflict
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

// SearchUsers finds instance accounts by display name or username prefix.
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
		ColumnExpr("u.id, u.display_name, u.username, u.avatar_url").
		Where("u.is_guest = false").
		Where("u.deleted_at IS NULL").
		Where("u.disabled_at IS NULL").
		Where(
			"(lower(u.display_name) LIKE ? OR lower(u.username) LIKE ?)",
			pattern, pattern,
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
		ID          string `bun:"id"`
		DisplayName string `bun:"display_name"`
		Username    string `bun:"username"`
		AvatarURL   string `bun:"avatar_url"`
	}
	if err := sel.Scan(ctx, &rows); err != nil {
		return nil, err
	}

	out := make([]models.UserSearchHit, len(rows))
	for i, row := range rows {
		out[i] = models.UserSearchHit{
			UserID:      row.ID,
			DisplayName: row.DisplayName,
			Username:    row.Username,
			AvatarURL:   row.AvatarURL,
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
			if user.AvatarURL != "" && linked.AvatarURL != user.AvatarURL {
				if _, uerr := tx.NewUpdate().Model((*models.UserRow)(nil)).
					Set("avatar_url = ?", user.AvatarURL).
					Set("updated_at = ?", user.UpdatedAt).
					Where("id = ?", linked.ID).
					Exec(ctx); uerr != nil {
					return uerr
				}
				linked.AvatarURL = user.AvatarURL
			}
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
			if user.AvatarURL != "" && existing.AvatarURL != user.AvatarURL {
				if _, uerr := tx.NewUpdate().Model((*models.UserRow)(nil)).
					Set("avatar_url = ?", user.AvatarURL).
					Set("updated_at = ?", user.UpdatedAt).
					Where("id = ?", existing.ID).
					Exec(ctx); uerr != nil {
					return uerr
				}
				existing.AvatarURL = user.AvatarURL
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

// UpdateUserAvatarURL sets users.avatar_url for a user.
func (r pgRepository) UpdateUserAvatarURL(ctx context.Context, userID, avatarURL string, now time.Time) error {
	res, err := r.db.NewUpdate().Model((*models.UserRow)(nil)).
		Set("avatar_url = ?", avatarURL).
		Set("updated_at = ?", now).
		Where("id = ?", userID).
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

	return nil
}

// UpdateUserProfile replaces user-editable profile fields on users.
func (r pgRepository) UpdateUserProfile(ctx context.Context, userID string, patch models.UserProfilePatch, now time.Time) error {
	q := r.db.NewUpdate().Model((*models.UserRow)(nil)).
		Set("display_name = ?", patch.DisplayName).
		Set("pronouns = ?", patch.Pronouns).
		Set("bio = ?", patch.Bio).
		Set("banner_color = ?", patch.BannerColor).
		Set("updated_at = ?", now).
		Where("id = ?", userID)
	if patch.ClearAvatar {
		q = q.Set("avatar_url = ?", "")
	}
	res, err := q.Exec(ctx)
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

	return nil
}

// UpdateUserStatusNote replaces or clears the caller's status note on users.
func (r pgRepository) UpdateUserStatusNote(ctx context.Context, userID string, patch models.UserStatusNotePatch, now time.Time) error {
	res, err := r.db.NewUpdate().Model((*models.UserRow)(nil)).
		Set("status_note_text = ?", patch.Text).
		Set("status_note_emoji = ?", patch.Emoji).
		Set("status_note_expires_at = ?", patch.ExpiresAt).
		Set("updated_at = ?", now).
		Where("id = ?", userID).
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

	return nil
}

// UpdateUserPreferences updates the caller's display locale and time format on users.
func (r pgRepository) UpdateUserPreferences(ctx context.Context, userID string, patch models.UserPreferencesPatch, now time.Time) error {
	res, err := r.db.NewUpdate().Model((*models.UserRow)(nil)).
		Set("locale = ?", patch.Locale).
		Set("time_format = ?", patch.TimeFormat).
		Set("updated_at = ?", now).
		Where("id = ?", userID).
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

	return nil
}

func (r pgRepository) UpdateUsername(ctx context.Context, userID, username string, now time.Time) error {
	res, err := r.db.NewUpdate().Model((*models.UserRow)(nil)).
		Set("username = ?", username).
		Set("updated_at = ?", now).
		Where("id = ?", userID).
		Exec(ctx)
	if err != nil {
		return pgErr.MapUnique(err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return constants.ErrNotFound
	}

	return nil
}

// UpdateEmail updates the caller's unique email on users.
func (r pgRepository) UpdateEmail(ctx context.Context, userID, email string, now time.Time) error {
	res, err := r.db.NewUpdate().Model((*models.UserRow)(nil)).
		Set("email = ?", email).
		Set("updated_at = ?", now).
		Where("id = ?", userID).
		Exec(ctx)
	if err != nil {
		return pgErr.MapUnique(err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return constants.ErrNotFound
	}

	return nil
}

// DisableAccount marks an active user account as disabled.
func (r pgRepository) DisableAccount(ctx context.Context, userID string, now time.Time) error {
	res, err := r.db.NewUpdate().Model((*models.UserRow)(nil)).
		Set("disabled_at = ?", now).
		Set("updated_at = ?", now).
		Where("id = ?", userID).
		Where("deleted_at IS NULL").
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

	return nil
}

// SoftDeleteAccount anonymizes PII, clears credentials/MFA/memberships, and sets deleted_at atomically.
func (r pgRepository) SoftDeleteAccount(ctx context.Context, userID, anonymizedUsername string, now time.Time) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		res, err := tx.NewUpdate().Model((*models.UserRow)(nil)).
			Set("deleted_at = ?", now).
			Set("disabled_at = NULL").
			Set("email = NULL").
			Set("username = ?", anonymizedUsername).
			Set("display_name = ?", "Deleted User").
			Set("avatar_url = ?", "").
			Set("pronouns = ?", "").
			Set("bio = ?", "").
			Set("banner_color = ?", "").
			Set("updated_at = ?", now).
			Where("id = ?", userID).
			Where("deleted_at IS NULL").
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

		for _, tbl := range []any{
			(*models.UserCredentialRow)(nil),
			(*models.UserConnectionRow)(nil),
			(*models.MFATOTPRow)(nil),
			(*models.WebAuthnCredentialRow)(nil),
			(*models.MFARecoveryCodeRow)(nil),
			(*models.GroupMemberRoleRow)(nil),
			(*models.GroupMemberRow)(nil),
		} {
			if _, err := tx.NewDelete().Model(tbl).Where("user_id = ?", userID).Exec(ctx); err != nil {
				return err
			}
		}

		return nil
	})
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

func (r pgRepository) GetMFATOTP(ctx context.Context, userID string) (models.MFATOTP, error) {
	var row models.MFATOTPRow
	err := r.db.NewSelect().Model(&row).Where("user_id = ?", userID).Scan(ctx)
	if err != nil {
		return models.MFATOTP{}, pgErr.MapNotFound(err)
	}

	return row.ToModel(), nil
}

func (r pgRepository) UpsertMFATOTP(ctx context.Context, totp models.MFATOTP) error {
	row := models.MFATOTPRow{
		UserID:          totp.UserID,
		SecretEncrypted: totp.SecretEncrypted,
		LastUsedStep:    totp.LastUsedStep,
		EnabledAt:       totp.EnabledAt,
		CreatedAt:       totp.CreatedAt,
		UpdatedAt:       totp.UpdatedAt,
	}
	_, err := r.db.NewInsert().
		Model(&row).
		On("CONFLICT (user_id) DO UPDATE").
		Set("secret_encrypted = EXCLUDED.secret_encrypted").
		Set("last_used_step = EXCLUDED.last_used_step").
		Set("enabled_at = EXCLUDED.enabled_at").
		Set("updated_at = EXCLUDED.updated_at").
		Exec(ctx)

	return err
}

func (r pgRepository) UpdateMFATOTPLastUsedStep(ctx context.Context, userID string, step int64, now time.Time) error {
	_, err := r.db.NewUpdate().
		Model((*models.MFATOTPRow)(nil)).
		Set("last_used_step = ?", step).
		Set("updated_at = ?", now).
		Where("user_id = ?", userID).
		Exec(ctx)

	return err
}

func (r pgRepository) DeleteMFATOTP(ctx context.Context, userID string) error {
	res, err := r.db.NewDelete().
		Model((*models.MFATOTPRow)(nil)).
		Where("user_id = ?", userID).
		Exec(ctx)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return constants.ErrNotFound
	}

	return nil
}

func (r pgRepository) ListWebAuthnCredentials(ctx context.Context, userID string) ([]models.WebAuthnCredential, error) {
	var rows []models.WebAuthnCredentialRow
	err := r.db.NewSelect().
		Model(&rows).
		Where("user_id = ?", userID).
		OrderExpr("created_at ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]models.WebAuthnCredential, len(rows))
	for i, row := range rows {
		out[i] = row.ToModel()
	}

	return out, nil
}

func (r pgRepository) InsertWebAuthnCredential(ctx context.Context, cred models.WebAuthnCredential) error {
	transports := cred.Transports
	if transports == nil {
		transports = []string{}
	}
	row := models.WebAuthnCredentialRow{
		ID:              cred.ID,
		UserID:          cred.UserID,
		Name:            cred.Name,
		CredentialID:    cred.CredentialID,
		PublicKey:       cred.PublicKey,
		AttestationType: cred.AttestationType,
		AAGUID:          cred.AAGUID,
		SignCount:       int64(cred.SignCount),
		CloneWarning:    cred.CloneWarning,
		Transports:      transports,
		BackupEligible:  cred.BackupEligible,
		BackupState:     cred.BackupState,
		LastUsedAt:      cred.LastUsedAt,
		CreatedAt:       cred.CreatedAt,
	}
	_, err := r.db.NewInsert().Model(&row).Exec(ctx)

	return pgErr.MapUnique(err)
}

func (r pgRepository) UpdateWebAuthnCredentialUsage(ctx context.Context, credentialID []byte, signCount uint32, cloneWarning, backupState bool, now time.Time) error {
	_, err := r.db.NewUpdate().
		Model((*models.WebAuthnCredentialRow)(nil)).
		Set("sign_count = ?", int64(signCount)).
		Set("clone_warning = ?", cloneWarning).
		Set("backup_state = ?", backupState).
		Set("last_used_at = ?", now).
		Where("credential_id = ?", credentialID).
		Exec(ctx)

	return err
}

func (r pgRepository) RenameWebAuthnCredential(ctx context.Context, userID, id, name string) (models.WebAuthnCredential, error) {
	var row models.WebAuthnCredentialRow
	err := r.db.NewUpdate().
		Model(&row).
		Set("name = ?", name).
		Where("id = ? AND user_id = ?", id, userID).
		Returning("*").
		Scan(ctx)
	if err != nil {
		return models.WebAuthnCredential{}, pgErr.MapNotFound(err)
	}

	return row.ToModel(), nil
}

func (r pgRepository) DeleteWebAuthnCredential(ctx context.Context, userID, id string) error {
	res, err := r.db.NewDelete().
		Model((*models.WebAuthnCredentialRow)(nil)).
		Where("id = ? AND user_id = ?", id, userID).
		Exec(ctx)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return constants.ErrNotFound
	}

	return nil
}

func (r pgRepository) CountUnusedRecoveryCodes(ctx context.Context, userID string) (int32, error) {
	n, err := r.db.NewSelect().
		Model((*models.MFARecoveryCodeRow)(nil)).
		Where("user_id = ? AND used_at IS NULL", userID).
		Count(ctx)

	return int32(n), err //nolint:gosec // G115
}

func (r pgRepository) ReplaceRecoveryCodes(ctx context.Context, userID string, codeHashes [][]byte, now time.Time) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewDelete().
			Model((*models.MFARecoveryCodeRow)(nil)).
			Where("user_id = ?", userID).
			Exec(ctx); err != nil {
			return err
		}
		if len(codeHashes) == 0 {
			return nil
		}

		rows := make([]models.MFARecoveryCodeRow, len(codeHashes))
		for i, hash := range codeHashes {
			id, err := token.NewID()
			if err != nil {
				return err
			}
			rows[i] = models.MFARecoveryCodeRow{
				ID:        id,
				UserID:    userID,
				CodeHash:  hash,
				CreatedAt: now,
			}
		}
		_, err := tx.NewInsert().Model(&rows).Exec(ctx)
		return err
	})
}

func (r pgRepository) ConsumeRecoveryCode(ctx context.Context, userID string, codeHash []byte, now time.Time) (bool, error) {
	res, err := r.db.NewUpdate().
		Model((*models.MFARecoveryCodeRow)(nil)).
		Set("used_at = ?", now).
		Where("user_id = ? AND code_hash = ? AND used_at IS NULL", userID, codeHash).
		Exec(ctx)
	if err != nil {
		return false, err
	}
	rows, _ := res.RowsAffected()

	return rows > 0, nil
}

func (r pgRepository) ListRelationships(ctx context.Context, userID string) ([]models.UserRelationship, error) {
	var rows []models.UserRelationshipRow
	err := r.db.NewSelect().Model(&rows).
		Where("user_id = ?", userID).
		OrderExpr("created_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]models.UserRelationship, len(rows))
	for i, row := range rows {
		out[i] = models.UserRelationship{
			UserID:       row.UserID,
			TargetUserID: row.TargetUserID,
			Kind:         row.Kind,
			CreatedAt:    row.CreatedAt,
		}
	}

	return out, nil
}

func (r pgRepository) UpsertRelationship(ctx context.Context, userID, targetUserID, kind string, now time.Time) error {
	row := models.UserRelationshipRow{
		UserID:       userID,
		TargetUserID: targetUserID,
		Kind:         kind,
		CreatedAt:    now,
	}
	_, err := r.db.NewInsert().Model(&row).
		On("CONFLICT (user_id, target_user_id) DO UPDATE").
		Set("kind = EXCLUDED.kind").
		Set("created_at = EXCLUDED.created_at").
		Exec(ctx)

	return err
}

func (r pgRepository) DeleteRelationship(ctx context.Context, userID, targetUserID string) error {
	_, err := r.db.NewDelete().Model((*models.UserRelationshipRow)(nil)).
		Where("user_id = ? AND target_user_id = ?", userID, targetUserID).
		Exec(ctx)

	return err
}
