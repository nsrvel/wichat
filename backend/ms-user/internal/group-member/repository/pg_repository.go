package repository

import (
	"context"
	"errors"
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
	InsertGroupMember(ctx context.Context, member models.GroupMember) error
	DeleteGroupMember(ctx context.Context, groupID, userID string) error
	CountGroupMembers(ctx context.Context, groupID string) (int, error)
	ListGroupMembersPage(ctx context.Context, groupID string, limit int, cursor *models.GroupMemberCursor) ([]models.GroupMemberProfile, *models.GroupMemberCursor, error)
	IsMemberBanned(ctx context.Context, groupID, userID string) (bool, error)
	BanMembers(ctx context.Context, groupID string, userIDs []string, bannedBy, reason string, deleteSeconds int) (int, error)
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
		GroupID: row.GroupID, UserID: row.UserID, Role: row.Role, CreatedAt: row.CreatedAt,
	}, nil
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

type groupMemberJoinedRow struct {
	UserID      string    `bun:"user_id"`
	DisplayName string    `bun:"display_name"`
	Email       *string   `bun:"email"`
	Role        string    `bun:"role"`
	CreatedAt   time.Time `bun:"created_at"`
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
		ColumnExpr("m.user_id, m.role, m.created_at").
		ColumnExpr("u.display_name, u.email").
		Join("JOIN users AS u ON u.id = m.user_id").
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
		email := ""
		if row.Email != nil {
			email = *row.Email
		}
		out[i] = models.GroupMemberProfile{
			UserID:      row.UserID,
			DisplayName: row.DisplayName,
			Email:       email,
			Role:        row.Role,
			JoinedAt:    row.CreatedAt,
		}
	}

	return out, next, nil
}

// IsMemberBanned checks whether a user is banned from a group.
func (r pgRepository) IsMemberBanned(ctx context.Context, groupID, userID string) (bool, error) {
	exists, err := r.db.NewSelect().Model((*models.GroupBanRow)(nil)).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Exists(ctx)
	if err != nil {
		r.logIfInfraErr("IsMemberBanned", err)
		return false, err
	}
	return exists, nil
}

// BanMembers records bans and removes users from group membership.
func (r pgRepository) BanMembers(ctx context.Context, groupID string, userIDs []string, bannedBy, reason string, deleteSeconds int) (int, error) {
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
				CreatedAt:            now,
			}
			if _, err := tx.NewInsert().Model(&ban).
				On("CONFLICT (group_id, user_id) DO UPDATE SET reason = EXCLUDED.reason, delete_message_seconds = EXCLUDED.delete_message_seconds").
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

func (r pgRepository) logIfInfraErr(op string, err error) {
	if err == nil || errors.Is(err, constants.ErrNotFound) || errors.Is(err, constants.ErrConflict) {
		return
	}
	r.log.Error("group-member repository", "op", op, "err", err)
}
