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

// PgRepository is the Postgres store for group memberships and bans.
type PgRepository interface {
	FindGroupMember(ctx context.Context, groupID, userID string) (models.GroupMember, error)
	FindGroupMemberAccount(ctx context.Context, groupID, userID string) (displayName, username string, err error)
	InsertGroupMember(ctx context.Context, member models.GroupMember) error
	DeleteGroupMember(ctx context.Context, groupID, userID string) error
	CountGroupMembers(ctx context.Context, groupID string) (int, error)
	ListGroupMemberIDs(ctx context.Context, groupID string) ([]string, error)
	ListGroupMembersPage(ctx context.Context, groupID string, limit int, cursor *models.GroupMemberCursor) ([]models.GroupMemberProfile, *models.GroupMemberCursor, error)
	IsMemberBanned(ctx context.Context, groupID, userID string) (bool, error)
	BanMembers(ctx context.Context, groupID string, userIDs []string, bannedBy, reason string, deleteSeconds int, expiresAt *time.Time) (int, error)
	ListBansPage(ctx context.Context, groupID, search string, limit int, cursor *models.GroupMemberCursor) ([]models.GroupBanEntry, *models.GroupMemberCursor, int, error)
	UnbanMember(ctx context.Context, groupID, userID string) error
	UpdateGroupMemberTimeout(ctx context.Context, groupID, userID string, timeoutUntil *time.Time) error
}

type pgRepository struct {
	db  *bun.DB
	log logger.Logger
}

// NewPgRepository stores group memberships and bans in Postgres.
func NewPgRepository(db *bun.DB, log logger.Logger) PgRepository {
	return pgRepository{db: db, log: log}
}

// FindGroupMember loads membership for a user in a group.
func (r pgRepository) FindGroupMember(ctx context.Context, groupID, userID string) (models.GroupMember, error) {
	var row models.GroupMemberRow
	err := r.db.NewSelect().Model(&row).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Scan(ctx)
	if err != nil {
		err = pgErr.MapNotFound(err)
		r.logIfInfraErr("FindGroupMember", err)
		return models.GroupMember{}, err
	}

	return models.GroupMember{
		GroupID: row.GroupID, UserID: row.UserID, Role: row.Role, TimeoutUntil: row.TimeoutUntil, CreatedAt: row.CreatedAt,
	}, nil
}

func (r pgRepository) FindGroupMemberAccount(ctx context.Context, groupID, userID string) (string, string, error) {
	var displayName, username string
	err := r.db.NewSelect().
		TableExpr("group_members AS m").
		ColumnExpr("u.display_name, u.username").
		Join("JOIN users AS u ON u.id = m.user_id").
		Where("m.group_id = ? AND m.user_id = ?", groupID, userID).
		Scan(ctx, &displayName, &username)
	if err != nil {
		err = pgErr.MapNotFound(err)
		r.logIfInfraErr("FindGroupMemberAccount", err)
		return "", "", err
	}

	return displayName, username, nil
}

// InsertGroupMember adds a user to a group and auto-assigns the default Member role.
func (r pgRepository) InsertGroupMember(ctx context.Context, member models.GroupMember) error {
	err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		row := models.GroupMemberRowFrom(member)
		if _, err := tx.NewInsert().Model(&row).Exec(ctx); err != nil {
			return pgErr.MapUnique(err)
		}
		var roleID string
		err := tx.NewSelect().Model((*models.GroupRoleRow)(nil)).
			Column("id").
			Where("group_id = ? AND slug = ?", member.GroupID, constants.RoleSlugMember).
			Scan(ctx, &roleID)
		if err == nil && roleID != "" {
			mr := models.GroupMemberRoleRow{
				GroupID:   member.GroupID,
				UserID:    member.UserID,
				RoleID:    roleID,
				CreatedAt: member.CreatedAt,
			}
			_, _ = tx.NewInsert().Model(&mr).On("CONFLICT DO NOTHING").Exec(ctx)
		}
		return nil
	})
	r.logIfInfraErr("InsertGroupMember", err)
	return err
}

// DeleteGroupMember removes a user's membership in a group.
func (r pgRepository) DeleteGroupMember(ctx context.Context, groupID, userID string) error {
	res, err := r.db.NewDelete().Model((*models.GroupMemberRow)(nil)).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Exec(ctx)
	if err != nil {
		r.logIfInfraErr("DeleteGroupMember", err)
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		r.logIfInfraErr("DeleteGroupMember", err)
		return err
	}
	if n == 0 {
		return constants.ErrNotFound
	}
	return nil
}

// CountGroupMembers returns how many users belong to a group.
func (r pgRepository) CountGroupMembers(ctx context.Context, groupID string) (int, error) {
	n, err := r.db.NewSelect().Model((*models.GroupMemberRow)(nil)).
		Where("group_id = ?", groupID).
		Count(ctx)
	if err != nil {
		r.logIfInfraErr("CountGroupMembers", err)
	}
	return int(n), err
}

// ListGroupMemberIDs returns all user IDs belonging to a group.
func (r pgRepository) ListGroupMemberIDs(ctx context.Context, groupID string) ([]string, error) {
	var ids []string
	err := r.db.NewSelect().Model((*models.GroupMemberRow)(nil)).
		Column("user_id").
		Where("group_id = ?", groupID).
		Scan(ctx, &ids)
	if err != nil {
		r.logIfInfraErr("ListGroupMemberIDs", err)
		return nil, err
	}
	return ids, nil
}

type groupMemberJoinedRow struct {
	UserID               string     `bun:"user_id"`
	DisplayName          string     `bun:"display_name"`
	Username             string     `bun:"username"`
	AvatarURL            string     `bun:"avatar_url"`
	BannerColor          string     `bun:"banner_color"`
	Bio                  string     `bun:"bio"`
	Pronouns             string     `bun:"pronouns"`
	UserCreatedAt        time.Time  `bun:"user_created_at"`
	StatusNoteText       string     `bun:"status_note_text"`
	StatusNoteEmoji      string     `bun:"status_note_emoji"`
	StatusNoteExpiresAt  *time.Time `bun:"status_note_expires_at"`
	JoinMethod           string     `bun:"join_method"`
	InviteCode           string     `bun:"invite_code"`
	InvitedByUserID      string     `bun:"invited_by_user_id"`
	InvitedByUsername    string     `bun:"invited_by_username"`
	InvitedByDisplayName string     `bun:"invited_by_display_name"`
	TimeoutUntil         *time.Time `bun:"timeout_until"`
	Role                 string     `bun:"role"`
	CreatedAt            time.Time  `bun:"created_at"`
}

// ListGroupMembersPage returns members ordered by join time then user id.
func (r pgRepository) ListGroupMembersPage(
	ctx context.Context,
	groupID string,
	limit int,
	cursor *models.GroupMemberCursor,
) ([]models.GroupMemberProfile, *models.GroupMemberCursor, error) {
	if limit <= 0 {
		limit = constants.DefaultMemberListLimit
	}
	if limit > constants.MaxMemberListLimit {
		limit = constants.MaxMemberListLimit
	}

	q := r.db.NewSelect().
		TableExpr("group_members AS m").
		ColumnExpr("m.user_id, m.role, m.created_at, m.join_method, m.invite_code, COALESCE(m.invited_by_user_id::text, '') AS invited_by_user_id, m.timeout_until").
		ColumnExpr("u.display_name, u.username, u.avatar_url, u.banner_color, u.bio, u.pronouns, u.created_at AS user_created_at, u.status_note_text, u.status_note_emoji, u.status_note_expires_at").
		ColumnExpr("COALESCE(inv.username, '') AS invited_by_username, COALESCE(inv.display_name, '') AS invited_by_display_name").
		Join("JOIN users AS u ON u.id = m.user_id").
		Join("LEFT JOIN users AS inv ON inv.id = m.invited_by_user_id").
		Where("m.group_id = ?", groupID).
		OrderExpr("m.created_at ASC, m.user_id ASC").
		Limit(int64(limit + 1))

	if cursor != nil {
		q = q.Where(
			"(m.created_at, m.user_id) > (?, ?)",
			cursor.CreatedAt,
			cursor.UserID,
		)
	}

	var rows []groupMemberJoinedRow
	if err := q.Scan(ctx, &rows); err != nil {
		r.logIfInfraErr("ListGroupMembersPage", err)
		return nil, nil, err
	}

	var next *models.GroupMemberCursor
	if len(rows) > limit {
		last := rows[limit-1]
		next = &models.GroupMemberCursor{CreatedAt: last.CreatedAt, UserID: last.UserID}
		rows = rows[:limit]
	}

	out := make([]models.GroupMemberProfile, len(rows))
	for i, row := range rows {
		out[i] = models.GroupMemberProfile{
			UserID:               row.UserID,
			DisplayName:          row.DisplayName,
			Username:             row.Username,
			AvatarURL:            row.AvatarURL,
			BannerColor:          row.BannerColor,
			Bio:                  row.Bio,
			Pronouns:             row.Pronouns,
			UserCreatedAt:        row.UserCreatedAt,
			StatusNoteText:       row.StatusNoteText,
			StatusNoteEmoji:      row.StatusNoteEmoji,
			StatusNoteExpiresAt:  row.StatusNoteExpiresAt,
			JoinMethod:           row.JoinMethod,
			InviteCode:           row.InviteCode,
			InvitedByUserID:      row.InvitedByUserID,
			InvitedByUsername:    row.InvitedByUsername,
			InvitedByDisplayName: row.InvitedByDisplayName,
			TimeoutUntil:         row.TimeoutUntil,
			Role:                 row.Role,
			JoinedAt:             row.CreatedAt,
		}
	}

	return out, next, nil
}

// IsMemberBanned checks whether a user has an active ban in a group.
func (r pgRepository) IsMemberBanned(ctx context.Context, groupID, userID string) (bool, error) {
	exists, err := r.db.NewSelect().Model((*models.GroupBanRow)(nil)).
		Where("group_id = ? AND user_id = ? AND (expires_at IS NULL OR expires_at > now())", groupID, userID).
		Exists(ctx)
	if err != nil {
		r.logIfInfraErr("IsMemberBanned", err)
		return false, err
	}
	return exists, nil
}

// BanMembers records bans and removes users from group membership.
func (r pgRepository) BanMembers(ctx context.Context, groupID string, userIDs []string, bannedBy, reason string, deleteSeconds int, expiresAt *time.Time) (int, error) {
	bannedCount := 0
	err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		now := time.Now().UTC()
		var bannedByPtr *string
		if bannedBy != "" {
			bannedByPtr = &bannedBy
		}

		for _, uid := range userIDs {
			var role string
			err := tx.NewSelect().Model((*models.GroupMemberRow)(nil)).
				Column("role").
				Where("group_id = ? AND user_id = ?", groupID, uid).
				Scan(ctx, &role)
			if err == nil && role == constants.GroupRoleOwner {
				continue
			}

			ban := models.GroupBanRow{
				GroupID:              groupID,
				UserID:               uid,
				BannedBy:             bannedByPtr,
				Reason:               reason,
				DeleteMessageSeconds: deleteSeconds,
				ExpiresAt:            expiresAt,
				CreatedAt:            now,
			}
			if _, err := tx.NewInsert().Model(&ban).
				On("CONFLICT (group_id, user_id) DO UPDATE SET banned_by = EXCLUDED.banned_by, reason = EXCLUDED.reason, delete_message_seconds = EXCLUDED.delete_message_seconds, expires_at = EXCLUDED.expires_at").
				Exec(ctx); err != nil {
				return err
			}

			_, _ = tx.NewDelete().Model((*models.GroupMemberRoleRow)(nil)).
				Where("group_id = ? AND user_id = ?", groupID, uid).
				Exec(ctx)

			if _, err := tx.NewDelete().Model((*models.GroupMemberRow)(nil)).
				Where("group_id = ? AND user_id = ?", groupID, uid).
				Exec(ctx); err != nil {
				return err
			}

			bannedCount++
		}
		return nil
	})
	if err != nil {
		r.logIfInfraErr("BanMembers", err)
		return 0, err
	}
	return bannedCount, nil
}

type groupBanQueryRow struct {
	UserID              string     `bun:"user_id"`
	Username            string     `bun:"username"`
	DisplayName         string     `bun:"display_name"`
	AvatarURL           string     `bun:"avatar_url"`
	Reason              string     `bun:"reason"`
	BannedByUserID      string     `bun:"banned_by_user_id"`
	BannedByUsername    string     `bun:"banned_by_username"`
	BannedByDisplayName string     `bun:"banned_by_display_name"`
	ExpiresAt           *time.Time `bun:"expires_at"`
	CreatedAt           time.Time  `bun:"created_at"`
}

func escapeILIKEPattern(raw string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return "%" + r.Replace(strings.TrimSpace(raw)) + "%"
}

// ListBansPage returns active group bans ordered by newest ban first with optional search.
func (r pgRepository) ListBansPage(
	ctx context.Context,
	groupID, search string,
	limit int,
	cursor *models.GroupMemberCursor,
) ([]models.GroupBanEntry, *models.GroupMemberCursor, int, error) {
	if limit <= 0 {
		limit = constants.DefaultMemberListLimit
	}
	if limit > constants.MaxMemberListLimit {
		limit = constants.MaxMemberListLimit
	}

	trimmedSearch := strings.TrimSpace(search)
	countQ := r.db.NewSelect().
		TableExpr("group_bans AS b").
		Join("JOIN users AS u ON u.id = b.user_id").
		Where("b.group_id = ? AND (b.expires_at IS NULL OR b.expires_at > now())", groupID)
	if trimmedSearch != "" {
		pat := escapeILIKEPattern(trimmedSearch)
		countQ = countQ.Where("(u.username ILIKE ? OR u.display_name ILIKE ? OR b.reason ILIKE ?)", pat, pat, pat)
	}
	total, err := countQ.Count(ctx)
	if err != nil {
		r.logIfInfraErr("ListBansPage.Count", err)
		return nil, nil, 0, err
	}

	q := r.db.NewSelect().
		TableExpr("group_bans AS b").
		ColumnExpr("b.user_id, u.username, u.display_name, COALESCE(u.avatar_url, '') AS avatar_url, b.reason, COALESCE(b.banned_by::text, '') AS banned_by_user_id, COALESCE(bu.username, '') AS banned_by_username, COALESCE(bu.display_name, '') AS banned_by_display_name, b.expires_at, b.created_at").
		Join("JOIN users AS u ON u.id = b.user_id").
		Join("LEFT JOIN users AS bu ON bu.id = b.banned_by").
		Where("b.group_id = ? AND (b.expires_at IS NULL OR b.expires_at > now())", groupID)
	if trimmedSearch != "" {
		pat := escapeILIKEPattern(trimmedSearch)
		q = q.Where("(u.username ILIKE ? OR u.display_name ILIKE ? OR b.reason ILIKE ?)", pat, pat, pat)
	}
	if cursor != nil {
		q = q.Where("(b.created_at < ? OR (b.created_at = ? AND b.user_id > ?))", cursor.CreatedAt, cursor.CreatedAt, cursor.UserID)
	}

	var rows []groupBanQueryRow
	if err := q.OrderExpr("b.created_at DESC, b.user_id ASC").Limit(int64(limit+1)).Scan(ctx, &rows); err != nil {
		r.logIfInfraErr("ListBansPage", err)
		return nil, nil, 0, err
	}

	var next *models.GroupMemberCursor
	if len(rows) > limit {
		last := rows[limit-1]
		next = &models.GroupMemberCursor{CreatedAt: last.CreatedAt, UserID: last.UserID}
		rows = rows[:limit]
	}

	out := make([]models.GroupBanEntry, 0, len(rows))
	for _, row := range rows {
		out = append(out, models.GroupBanEntry(row))
	}
	return out, next, int(total), nil
}

// UnbanMember removes an active ban record from a group.
func (r pgRepository) UnbanMember(ctx context.Context, groupID, userID string) error {
	res, err := r.db.NewDelete().Model((*models.GroupBanRow)(nil)).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Exec(ctx)
	if err != nil {
		r.logIfInfraErr("UnbanMember", err)
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return constants.ErrNotFound
	}
	return nil
}

// UpdateGroupMemberTimeout sets or clears timeout_until for a group member.
func (r pgRepository) UpdateGroupMemberTimeout(ctx context.Context, groupID, userID string, timeoutUntil *time.Time) error {
	res, err := r.db.NewUpdate().Model((*models.GroupMemberRow)(nil)).
		Set("timeout_until = ?", timeoutUntil).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Exec(ctx)
	if err != nil {
		r.logIfInfraErr("UpdateGroupMemberTimeout", err)
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return constants.ErrNotFound
	}
	return nil
}

func (r pgRepository) logIfInfraErr(op string, err error) {
	if err == nil || errors.Is(err, constants.ErrNotFound) || errors.Is(err, constants.ErrConflict) {
		return
	}
	r.log.Error("group-member repository", "op", op, "err", err)
}
