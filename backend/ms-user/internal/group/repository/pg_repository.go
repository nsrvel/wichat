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

// PgRepository is the Postgres store for groups and membership.
type PgRepository interface {
	GetFirstGroup(ctx context.Context) (models.Group, error)
	ListGroupsForUser(ctx context.Context, userID string) ([]models.Group, error)
	ListJoinableGroups(ctx context.Context, userID string) ([]models.Group, error)
	FindGroupByID(ctx context.Context, id string) (models.Group, error)
	FindGroupByPublicID(ctx context.Context, publicID string) (models.Group, error)
	FindGroupMember(ctx context.Context, groupID, userID string) (models.GroupMember, error)
	CreateGroupWithOwner(ctx context.Context, group models.Group, owner models.GroupMember) error
	InsertGroupMember(ctx context.Context, member models.GroupMember) error
	DeleteGroupMember(ctx context.Context, groupID, userID string) error
	DeleteGroup(ctx context.Context, groupID string) error
	UpdateGroupDisplayName(ctx context.Context, groupID, name string, now time.Time) error
	UpdateGroupLogoURL(ctx context.Context, groupID, logoURL string, now time.Time) error
	UpdateGroupProfile(ctx context.Context, groupID string, patch models.GroupProfilePatch, now time.Time) error
	SetRailPositions(ctx context.Context, userID string, orderedGroupIDs []string) error
	CountGroupMembers(ctx context.Context, groupID string) (int, error)
	ListGroupMembersPage(ctx context.Context, groupID string, limit int, cursor *models.GroupMemberCursor) ([]models.GroupMemberProfile, *models.GroupMemberCursor, error)
}

type pgRepository struct {
	db  *bun.DB
	log logger.Logger
}

// NewPgRepository stores groups and membership in Postgres.
func NewPgRepository(db *bun.DB, log logger.Logger) PgRepository {
	return pgRepository{db: db, log: log}
}

// GetFirstGroup returns the earliest group (bootstrap / setup status).
func (r pgRepository) GetFirstGroup(ctx context.Context) (models.Group, error) {
	var joined groupJoinedRow
	err := groupJoinSelect(r.db.NewSelect().TableExpr("groups AS g")).
		Order("g.created_at ASC").
		Limit(1).
		Scan(ctx, &joined)
	if err != nil {
		err = pgErr.MapNotFound(err)
		r.logIfInfraErr("GetFirstGroup", err)

		return models.Group{}, err
	}

	return groupFromRows(joined.Group, joined.Settings, joined.DefaultChannelPublicID, ""), nil
}

// ListGroupsForUser returns groups the user belongs to, ordered by rail preference then name.
func (r pgRepository) ListGroupsForUser(ctx context.Context, userID string) ([]models.Group, error) {
	return r.listGroupsRaw(ctx, constants.SQLListGroupsForUser, userID, "ListGroupsForUser")
}

// ListJoinableGroups returns instance groups the user is not in.
func (r pgRepository) ListJoinableGroups(ctx context.Context, userID string) ([]models.Group, error) {
	return r.listGroupsRaw(ctx, constants.SQLListJoinableGroups, userID, "ListJoinableGroups")
}

// FindGroupByID loads one group by id.
func (r pgRepository) FindGroupByID(ctx context.Context, id string) (models.Group, error) {
	var joined groupJoinedRow
	err := groupJoinSelect(r.db.NewSelect().TableExpr("groups AS g")).
		Where("g.id = ?", id).
		Scan(ctx, &joined)
	if err != nil {
		err = pgErr.MapNotFound(err)
		r.logIfInfraErr("FindGroupByID", err)

		return models.Group{}, err
	}

	return groupFromRows(joined.Group, joined.Settings, joined.DefaultChannelPublicID, ""), nil
}

// FindGroupByPublicID loads one group by its client-facing public id.
func (r pgRepository) FindGroupByPublicID(ctx context.Context, publicID string) (models.Group, error) {
	var joined groupJoinedRow
	err := groupJoinSelect(r.db.NewSelect().TableExpr("groups AS g")).
		Where("g.public_id = ?", publicID).
		Scan(ctx, &joined)
	if err != nil {
		err = pgErr.MapNotFound(err)
		r.logIfInfraErr("FindGroupByPublicID", err)

		return models.Group{}, err
	}

	return groupFromRows(joined.Group, joined.Settings, joined.DefaultChannelPublicID, ""), nil
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

// CreateGroupWithOwner inserts a group and owner membership in one transaction.
func (r pgRepository) CreateGroupWithOwner(ctx context.Context, group models.Group, owner models.GroupMember) error {
	err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		return InsertGroupWithOwnerTx(ctx, tx, group, owner)
	})
	r.logIfInfraErr("CreateGroupWithOwner", err)

	return err
}

// InsertGroupMember adds a user to a group.
func (r pgRepository) InsertGroupMember(ctx context.Context, member models.GroupMember) error {
	row := models.GroupMemberRowFrom(member)
	_, err := r.db.NewInsert().Model(&row).Exec(ctx)
	err = pgErr.MapUnique(err)
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

// DeleteGroup removes a group; membership and channels cascade in Postgres.
func (r pgRepository) DeleteGroup(ctx context.Context, groupID string) error {
	res, err := r.db.NewDelete().Model((*models.GroupRow)(nil)).
		Where("id = ?", groupID).
		Exec(ctx)
	if err != nil {
		r.logIfInfraErr("DeleteGroup", err)

		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		r.logIfInfraErr("DeleteGroup", err)

		return err
	}
	if n == 0 {
		return constants.ErrNotFound
	}

	return nil
}

// UpdateGroupDisplayName sets group_settings.name for a workspace.
func (r pgRepository) UpdateGroupDisplayName(ctx context.Context, groupID, name string, now time.Time) error {
	_, err := r.db.NewUpdate().Model((*models.GroupSettingsRow)(nil)).
		Set("name = ?", name).
		Set("updated_at = ?", now).
		Where("group_id = ?", groupID).
		Exec(ctx)
	r.logIfInfraErr("UpdateGroupDisplayName", err)

	return err
}

// UpdateGroupLogoURL sets group_settings.logo_url.
func (r pgRepository) UpdateGroupLogoURL(ctx context.Context, groupID, logoURL string, now time.Time) error {
	_, err := r.db.NewUpdate().Model((*models.GroupSettingsRow)(nil)).
		Set("logo_url = ?", logoURL).
		Set("updated_at = ?", now).
		Where("group_id = ?", groupID).
		Exec(ctx)
	r.logIfInfraErr("UpdateGroupLogoURL", err)

	return err
}

// UpdateGroupProfile replaces owner-editable group_settings profile fields.
func (r pgRepository) UpdateGroupProfile(ctx context.Context, groupID string, patch models.GroupProfilePatch, now time.Time) error {
	_, err := r.db.NewUpdate().Model((*models.GroupSettingsRow)(nil)).
		Set("name = ?", patch.Name).
		Set("description = ?", patch.Description).
		Set("is_private = ?", patch.IsPrivate).
		Set("banner_color = ?", patch.BannerColor).
		Set("banner_url = ?", patch.BannerURL).
		Set("logo_url = ?", patch.LogoURL).
		Set("updated_at = ?", now).
		Where("group_id = ?", groupID).
		Exec(ctx)
	r.logIfInfraErr("UpdateGroupProfile", err)

	return err
}

type groupJoinedRow struct {
	Group                  models.GroupRow         `bun:"embed:g_"`
	Settings               models.GroupSettingsRow `bun:"embed:s_"`
	DefaultChannelPublicID string                  `bun:"ch_public_id"`
	MemberRole             string                  `bun:"member_role"`
}

func groupJoinSelect(q *bun.SelectQuery) *bun.SelectQuery {
	return q.
		ColumnExpr("g.id AS g_id, g.public_id AS g_public_id, g.created_at AS g_created_at, g.updated_at AS g_updated_at").
		ColumnExpr("s.group_id AS s_group_id, s.name AS s_name, s.logo_url AS s_logo_url").
		ColumnExpr("s.description AS s_description, s.is_private AS s_is_private").
		ColumnExpr("s.banner_color AS s_banner_color, s.banner_url AS s_banner_url").
		ColumnExpr("s.retention_days AS s_retention_days, s.trash_retention_days AS s_trash_retention_days").
		ColumnExpr("s.created_at AS s_created_at, s.updated_at AS s_updated_at").
		ColumnExpr("ch.public_id AS ch_public_id").
		Join("JOIN group_settings AS s ON s.group_id = g.id").
		Join("LEFT JOIN channels AS ch ON ch.group_id = g.id AND lower(ch.name) = ?", constants.DefaultChannelLookupKey)
}

func groupFromRows(g models.GroupRow, s models.GroupSettingsRow, defaultChannelPublicID, callerRole string) models.Group {
	return models.Group{
		ID: g.ID, PublicID: g.PublicID, Name: s.Name, LogoURL: s.LogoURL,
		Description: s.Description, IsPrivate: s.IsPrivate,
		BannerColor: s.BannerColor, BannerURL: s.BannerURL,
		DefaultChannelPublicID: defaultChannelPublicID,
		CallerRole:             callerRole,
		CreatedAt:              g.CreatedAt, UpdatedAt: g.UpdatedAt,
	}
}

func (r pgRepository) listGroupsRaw(ctx context.Context, query, userID, op string) ([]models.Group, error) {
	var joined []groupJoinedRow
	err := r.db.NewRaw(query, userID).Scan(ctx, &joined)
	if err != nil {
		r.logIfInfraErr(op, err)

		return nil, err
	}

	out := make([]models.Group, len(joined))
	for i, row := range joined {
		out[i] = groupFromRows(row.Group, row.Settings, row.DefaultChannelPublicID, row.MemberRole)
	}

	return out, nil
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

// SetRailPositions sets per-user rail order (0..n-1) for the caller's memberships.
func (r pgRepository) SetRailPositions(ctx context.Context, userID string, orderedGroupIDs []string) error {
	err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		for i, groupID := range orderedGroupIDs {
			res, err := tx.NewUpdate().Model((*models.GroupMemberRow)(nil)).
				Set("rail_position = ?", i).
				Where("user_id = ? AND group_id = ?", userID, groupID).
				Exec(ctx)
			if err != nil {
				return err
			}
			if n, _ := res.RowsAffected(); n != 1 {
				return constants.ErrNotFound
			}
		}
		return nil
	})
	if err != nil {
		r.logIfInfraErr("SetRailPositions", err)
	}
	return err
}

func (r pgRepository) logIfInfraErr(op string, err error) {
	if err == nil || errors.Is(err, constants.ErrNotFound) || errors.Is(err, constants.ErrConflict) {
		return
	}
	r.log.Error("group repository", "op", op, "err", err)
}

// InsertGroupWithOwnerTx inserts groups, group_settings, and owner membership inside an existing transaction.
func InsertGroupWithOwnerTx(ctx context.Context, tx bun.Tx, group models.Group, owner models.GroupMember) error {
	gRow := models.GroupRow{
		ID: group.ID, PublicID: group.PublicID, CreatedAt: group.CreatedAt, UpdatedAt: group.UpdatedAt,
	}
	if _, err := tx.NewInsert().Model(&gRow).Exec(ctx); err != nil {
		return pgErr.MapUnique(err)
	}
	sRow := models.GroupSettingsRow{
		GroupID: group.ID, Name: group.Name, LogoURL: group.LogoURL,
		RetentionDays: constants.DefaultGroupRetentionDays, TrashRetentionDays: constants.DefaultGroupTrashRetentionDays,
		CreatedAt: group.CreatedAt, UpdatedAt: group.UpdatedAt,
	}
	if _, err := tx.NewInsert().Model(&sRow).Exec(ctx); err != nil {
		return err
	}
	mRow := models.GroupMemberRowFrom(owner)
	if _, err := tx.NewInsert().Model(&mRow).Exec(ctx); err != nil {
		return err
	}

	return nil
}
