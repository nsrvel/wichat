package repository

import (
	"context"
	"errors"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	pgErr "github.com/wichat/wichat/backend/ms-user/pkg/db"
	"github.com/wichat/wichat/backend/ms-user/pkg/token"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

// PgRepository is the Postgres store for group roles and permissions.
type PgRepository interface {
	GetEffectivePermissions(ctx context.Context, groupID, userID string) ([]string, error)
	ListRolesForGroup(ctx context.Context, groupID string) ([]models.GroupRole, error)
	FindRoleByID(ctx context.Context, groupID, roleID string) (models.GroupRole, error)
	UpdateRole(ctx context.Context, groupID, roleID string, patch models.GroupRolePatch, now time.Time) (models.GroupRole, error)
	AddMemberRoles(ctx context.Context, groupID, roleID string, userIDs []string) error
	RemoveMemberRole(ctx context.Context, groupID, roleID, userID string) error
	ListRolesForMembers(ctx context.Context, groupID string, userIDs []string) (map[string][]models.GroupRoleSummary, error)
	SeedDefaultGroupRolesTx(ctx context.Context, tx bun.Tx, groupID, ownerUserID string, now time.Time) error
}

type pgRepository struct {
	db  *bun.DB
	log logger.Logger
}

// NewPgRepository stores group roles in Postgres.
func NewPgRepository(db *bun.DB, log logger.Logger) PgRepository {
	return pgRepository{db: db, log: log}
}

// GetEffectivePermissions returns permissions union for (groupID, userID).
func (r pgRepository) GetEffectivePermissions(ctx context.Context, groupID, userID string) ([]string, error) {
	var memberRole string
	err := r.db.NewSelect().Model((*models.GroupMemberRow)(nil)).
		Column("role").
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Scan(ctx, &memberRole)
	if err != nil {
		err = pgErr.MapNotFound(err)
		r.logIfInfraErr("GetEffectivePermissions.FindMember", err)
		return nil, err
	}
	if memberRole == constants.GroupRoleOwner {
		return append([]string(nil), constants.AllKnownPermissions...), nil
	}

	var perms []string
	query := `
		SELECT DISTINCT unnest(gr.permissions)
		FROM group_roles gr
		WHERE gr.group_id = ? AND gr.is_default = true
		UNION
		SELECT DISTINCT unnest(gr.permissions)
		FROM group_roles gr
		JOIN group_member_roles gmr ON gmr.role_id = gr.id AND gmr.group_id = gr.group_id
		WHERE gmr.group_id = ? AND gmr.user_id = ?
	`
	err = r.db.NewRaw(query, groupID, groupID, userID).Scan(ctx, &perms)
	if err != nil {
		r.logIfInfraErr("GetEffectivePermissions.Union", err)
		return nil, err
	}

	for _, p := range perms {
		if p == constants.PermAdministrator {
			return append([]string(nil), constants.AllKnownPermissions...), nil
		}
	}
	return perms, nil
}

type roleWithCountRow struct {
	models.GroupRoleRow
	MemberCount int `bun:"member_count"`
}

// ListRolesForGroup returns all roles in a group ordered by position.
func (r pgRepository) ListRolesForGroup(ctx context.Context, groupID string) ([]models.GroupRole, error) {
	var rows []roleWithCountRow
	query := `
		SELECT gr.id, gr.public_id, gr.group_id, gr.slug, gr.name, gr.color, gr.hoist, gr.mentionable,
		       gr.is_default, gr.is_system, gr.position, gr.permissions, gr.created_at, gr.updated_at,
		       CASE WHEN gr.is_default THEN (SELECT count(*)::int FROM group_members WHERE group_id = gr.group_id)
		            ELSE COALESCE(mc.cnt, 0)::int
		       END AS member_count
		FROM group_roles gr
		LEFT JOIN (
		    SELECT role_id, count(*)::int AS cnt
		    FROM group_member_roles
		    WHERE group_id = ?
		    GROUP BY role_id
		) mc ON mc.role_id = gr.id
		WHERE gr.group_id = ?
		ORDER BY gr.position ASC, gr.created_at ASC
	`
	err := r.db.NewRaw(query, groupID, groupID).Scan(ctx, &rows)
	if err != nil {
		r.logIfInfraErr("ListRolesForGroup", err)
		return nil, err
	}

	out := make([]models.GroupRole, len(rows))
	for i, row := range rows {
		out[i] = models.GroupRole{
			ID:          row.ID,
			PublicID:    row.PublicID,
			GroupID:     row.GroupID,
			Slug:        row.Slug,
			Name:        row.Name,
			Color:       row.Color,
			Hoist:       row.Hoist,
			Mentionable: row.Mentionable,
			IsDefault:   row.IsDefault,
			IsSystem:    row.IsSystem,
			Position:    row.Position,
			Permissions: row.Permissions,
			MemberCount: row.MemberCount,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
		}
	}
	return out, nil
}

// FindRoleByID finds a role by its UUID or 24-char public ID within a group.
func (r pgRepository) FindRoleByID(ctx context.Context, groupID, roleID string) (models.GroupRole, error) {
	var row roleWithCountRow
	query := `
		SELECT gr.id, gr.public_id, gr.group_id, gr.slug, gr.name, gr.color, gr.hoist, gr.mentionable,
		       gr.is_default, gr.is_system, gr.position, gr.permissions, gr.created_at, gr.updated_at,
		       CASE WHEN gr.is_default THEN (SELECT count(*)::int FROM group_members WHERE group_id = gr.group_id)
		            ELSE COALESCE(mc.cnt, 0)::int
		       END AS member_count
		FROM group_roles gr
		LEFT JOIN (
		    SELECT role_id, count(*)::int AS cnt
		    FROM group_member_roles
		    WHERE group_id = ?
		    GROUP BY role_id
		) mc ON mc.role_id = gr.id
		WHERE gr.group_id = ? AND (gr.id = ? OR gr.public_id = ?)
		LIMIT 1
	`
	err := r.db.NewRaw(query, groupID, groupID, roleID, roleID).Scan(ctx, &row)
	if err != nil {
		err = pgErr.MapNotFound(err)
		r.logIfInfraErr("FindRoleByID", err)
		return models.GroupRole{}, err
	}
	return models.GroupRole{
		ID:          row.ID,
		PublicID:    row.PublicID,
		GroupID:     row.GroupID,
		Slug:        row.Slug,
		Name:        row.Name,
		Color:       row.Color,
		Hoist:       row.Hoist,
		Mentionable: row.Mentionable,
		IsDefault:   row.IsDefault,
		IsSystem:    row.IsSystem,
		Position:    row.Position,
		Permissions: row.Permissions,
		MemberCount: row.MemberCount,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}, nil
}

// UpdateRole updates role scalars and permissions.
func (r pgRepository) UpdateRole(ctx context.Context, groupID, roleID string, patch models.GroupRolePatch, now time.Time) (models.GroupRole, error) {
	role, err := r.FindRoleByID(ctx, groupID, roleID)
	if err != nil {
		return models.GroupRole{}, err
	}

	q := r.db.NewUpdate().Model((*models.GroupRoleRow)(nil)).
		Where("group_id = ? AND (id = ? OR public_id = ?)", groupID, roleID, roleID).
		Set("updated_at = ?", now)

	if patch.Permissions != nil {
		q = q.Set("permissions = ?", pgdialect.Array(patch.Permissions))
	}
	if !role.IsDefault {
		if patch.Name != nil {
			q = q.Set("name = ?", *patch.Name)
		}
		if patch.Color != nil {
			q = q.Set("color = ?", *patch.Color)
		}
		if patch.Hoist != nil {
			q = q.Set("hoist = ?", *patch.Hoist)
		}
		if patch.Mentionable != nil {
			q = q.Set("mentionable = ?", *patch.Mentionable)
		}
	}

	res, err := q.Exec(ctx)
	if err != nil {
		r.logIfInfraErr("UpdateRole", err)
		return models.GroupRole{}, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return models.GroupRole{}, constants.ErrNotFound
	}

	return r.FindRoleByID(ctx, groupID, roleID)
}

// AddMemberRoles adds members to a role.
func (r pgRepository) AddMemberRoles(ctx context.Context, groupID, roleID string, userIDs []string) error {
	role, err := r.FindRoleByID(ctx, groupID, roleID)
	if err != nil {
		return err
	}
	if role.IsDefault {
		return nil
	}

	err = r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		for _, uid := range userIDs {
			var exists bool
			err := tx.NewSelect().Model((*models.GroupMemberRow)(nil)).
				Where("group_id = ? AND user_id = ?", groupID, uid).
				ColumnExpr("1").
				Scan(ctx, &exists)
			if err != nil || !exists {
				continue
			}

			mRole := models.GroupMemberRoleRow{
				GroupID:   groupID,
				UserID:    uid,
				RoleID:    role.ID,
				CreatedAt: time.Now().UTC(),
			}
			if _, err := tx.NewInsert().Model(&mRole).On("CONFLICT DO NOTHING").Exec(ctx); err != nil {
				return err
			}
		}
		return nil
	})
	r.logIfInfraErr("AddMemberRoles", err)
	return err
}

// RemoveMemberRole revokes a role from a member.
func (r pgRepository) RemoveMemberRole(ctx context.Context, groupID, roleID, userID string) error {
	role, err := r.FindRoleByID(ctx, groupID, roleID)
	if err != nil {
		return err
	}
	res, err := r.db.NewDelete().Model((*models.GroupMemberRoleRow)(nil)).
		Where("group_id = ? AND role_id = ? AND user_id = ?", groupID, role.ID, userID).
		Exec(ctx)
	if err != nil {
		r.logIfInfraErr("RemoveMemberRole", err)
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return constants.ErrNotFound
	}
	return nil
}

type memberRoleRow struct {
	UserID string `bun:"user_id"`
	ID     string `bun:"id"`
	Slug   string `bun:"slug"`
	Name   string `bun:"name"`
	Color  string `bun:"color"`
}

// ListRolesForMembers returns mapped roles for members in a group.
func (r pgRepository) ListRolesForMembers(ctx context.Context, groupID string, userIDs []string) (map[string][]models.GroupRoleSummary, error) {
	out := make(map[string][]models.GroupRoleSummary, len(userIDs))
	for _, uid := range userIDs {
		out[uid] = []models.GroupRoleSummary{}
	}
	if len(userIDs) == 0 {
		return out, nil
	}

	var rows []memberRoleRow
	query := `
		SELECT gmr.user_id, gr.id, gr.slug, gr.name, gr.color
		FROM group_member_roles gmr
		JOIN group_roles gr ON gr.id = gmr.role_id AND gr.group_id = gmr.group_id
		WHERE gmr.group_id = ? AND gmr.user_id IN (?)
		ORDER BY gr.position ASC, gr.created_at ASC
	`
	err := r.db.NewRaw(query, groupID, bun.List(userIDs)).Scan(ctx, &rows)
	if err != nil {
		r.logIfInfraErr("ListRolesForMembers", err)
		return out, err
	}

	for _, row := range rows {
		out[row.UserID] = append(out[row.UserID], models.GroupRoleSummary{
			ID:    row.ID,
			Slug:  row.Slug,
			Name:  row.Name,
			Color: row.Color,
		})
	}
	return out, nil
}

// SeedDefaultGroupRolesTx creates the default @everyone and Member roles for a newly created group.
func (r pgRepository) SeedDefaultGroupRolesTx(ctx context.Context, tx bun.Tx, groupID, ownerUserID string, now time.Time) error {
	everyonePubID, err := token.NewPublicID()
	if err != nil {
		return err
	}
	everyoneID, err := token.NewID()
	if err != nil {
		return err
	}
	everyoneRole := models.GroupRoleRow{
		ID:          everyoneID,
		PublicID:    everyonePubID,
		GroupID:     groupID,
		Slug:        constants.RoleSlugEveryone,
		Name:        "@everyone",
		Color:       "#99aab5",
		Hoist:       false,
		Mentionable: false,
		IsDefault:   true,
		IsSystem:    true,
		Position:    0,
		Permissions: constants.DefaultEveryonePermissions,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if _, err := tx.NewInsert().Model(&everyoneRole).Exec(ctx); err != nil {
		return err
	}

	memberPubID, err := token.NewPublicID()
	if err != nil {
		return err
	}
	memberID, err := token.NewID()
	if err != nil {
		return err
	}
	memberRole := models.GroupRoleRow{
		ID:          memberID,
		PublicID:    memberPubID,
		GroupID:     groupID,
		Slug:        constants.RoleSlugMember,
		Name:        "Member",
		Color:       "#5865f2",
		Hoist:       false,
		Mentionable: false,
		IsDefault:   false,
		IsSystem:    true,
		Position:    1,
		Permissions: []string{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if _, err := tx.NewInsert().Model(&memberRole).Exec(ctx); err != nil {
		return err
	}

	ownerMemberRole := models.GroupMemberRoleRow{
		GroupID:   groupID,
		UserID:    ownerUserID,
		RoleID:    memberRole.ID,
		CreatedAt: now,
	}
	if _, err := tx.NewInsert().Model(&ownerMemberRole).On("CONFLICT DO NOTHING").Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (r pgRepository) logIfInfraErr(op string, err error) {
	if err == nil || errors.Is(err, constants.ErrNotFound) || errors.Is(err, constants.ErrConflict) {
		return
	}
	r.log.Error("group-role repository", "op", op, "err", err)
}
