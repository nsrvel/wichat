package repository

import (
	"context"
	"errors"
	"time"

	"github.com/uptrace/bun"
	channelrepo "github.com/wichat/wichat/backend/ms-user/internal/channel/repository"
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
	UpdateGroupDisplayName(ctx context.Context, groupID, name string, now time.Time) error
	UpdateGroupLogoURL(ctx context.Context, groupID, logoURL string, now time.Time) error
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

	return groupFromRows(joined.Group, joined.Settings, joined.DefaultChannelPublicID), nil
}

// ListGroupsForUser returns groups the user belongs to, ordered by name.
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

	return groupFromRows(joined.Group, joined.Settings, joined.DefaultChannelPublicID), nil
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

	return groupFromRows(joined.Group, joined.Settings, joined.DefaultChannelPublicID), nil
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

type groupJoinedRow struct {
	Group                  models.GroupRow         `bun:"embed:g_"`
	Settings               models.GroupSettingsRow `bun:"embed:s_"`
	DefaultChannelPublicID string                  `bun:"ch_public_id"`
}

func groupJoinSelect(q *bun.SelectQuery) *bun.SelectQuery {
	return q.
		ColumnExpr("g.id AS g_id, g.public_id AS g_public_id, g.created_at AS g_created_at, g.updated_at AS g_updated_at").
		ColumnExpr("s.group_id AS s_group_id, s.name AS s_name, s.logo_url AS s_logo_url").
		ColumnExpr("s.retention_days AS s_retention_days, s.trash_retention_days AS s_trash_retention_days").
		ColumnExpr("s.created_at AS s_created_at, s.updated_at AS s_updated_at").
		ColumnExpr("ch.public_id AS ch_public_id").
		Join("JOIN group_settings AS s ON s.group_id = g.id").
		Join("LEFT JOIN channels AS ch ON ch.group_id = g.id AND ch.name = ?", constants.DefaultChannelName)
}

func groupFromRows(g models.GroupRow, s models.GroupSettingsRow, defaultChannelPublicID string) models.Group {
	return models.Group{
		ID: g.ID, PublicID: g.PublicID, Name: s.Name, LogoURL: s.LogoURL,
		DefaultChannelPublicID: defaultChannelPublicID,
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
		out[i] = groupFromRows(row.Group, row.Settings, row.DefaultChannelPublicID)
	}

	return out, nil
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

	return channelrepo.InsertGeneralChannelTx(ctx, tx, group.ID, owner.UserID, group.CreatedAt)
}
