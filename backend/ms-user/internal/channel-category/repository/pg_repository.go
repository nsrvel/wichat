package repository

import (
	"context"
	"errors"
	"time"

	"github.com/uptrace/bun"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	pgErr "github.com/wichat/wichat/backend/ms-user/pkg/db"
	"github.com/wichat/wichat/backend/ms-user/pkg/token"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

// PgRepository loads and persists channel category metadata and permission overrides.
type PgRepository interface {
	ListCategoriesByGroupID(ctx context.Context, groupID string) ([]models.ChannelCategory, error)
	CreateCategory(ctx context.Context, groupID, name string, isPrivate bool, createdBy string, now time.Time) (models.ChannelCategory, error)
	UpdateCategory(ctx context.Context, groupID, categoryPublicID string, name *string, isPrivate *bool, now time.Time) (models.ChannelCategory, error)
	DeleteCategory(ctx context.Context, groupID, categoryPublicID string) error
	ListCategoryPermissionOverrides(ctx context.Context, groupID, categoryPublicID string) ([]models.PublicChannelCategoryPermissionOverride, error)
	ListGroupCategoryPermissionOverrides(ctx context.Context, groupID string) ([]models.ChannelCategoryPermissionOverride, error)
	UpsertCategoryPermissionOverride(ctx context.Context, groupID, categoryPublicID, targetType, targetPublicID string, allow, deny []string, now time.Time) (models.PublicChannelCategoryPermissionOverride, error)
	DeleteCategoryPermissionOverride(ctx context.Context, groupID, categoryPublicID, targetType, targetPublicID string) error
	SetCategoryPositions(ctx context.Context, groupID string, orderedCategoryIDs []string) error
}

type pgRepository struct {
	db  *bun.DB
	log logger.Logger
}

// NewPgRepository constructs a Postgres channel category store.
func NewPgRepository(db *bun.DB, log logger.Logger) PgRepository {
	return pgRepository{db: db, log: log}
}

func (r pgRepository) ListCategoriesByGroupID(ctx context.Context, groupID string) ([]models.ChannelCategory, error) {
	var rows []models.ChannelCategoryRow
	err := r.db.NewSelect().Model(&rows).
		Where("group_id = ?", groupID).
		OrderExpr("position ASC, created_at ASC").
		Scan(ctx)
	if err != nil {
		r.logIfInfraErr("ListCategoriesByGroupID", err)

		return nil, err
	}

	out := make([]models.ChannelCategory, len(rows))
	for i, row := range rows {
		out[i] = row.ToModel()
	}

	return out, nil
}

func (r pgRepository) CreateCategory(
	ctx context.Context,
	groupID, name string,
	isPrivate bool,
	createdBy string,
	now time.Time,
) (models.ChannelCategory, error) {
	var maxPos int32
	err := r.db.NewSelect().
		TableExpr("channel_categories").
		ColumnExpr("COALESCE(MAX(position), -1)").
		Where("group_id = ?", groupID).
		Scan(ctx, &maxPos)
	if err != nil {
		r.logIfInfraErr("CreateCategory", err)

		return models.ChannelCategory{}, err
	}

	cat, err := insertCategoryWithRetry(ctx, r.db, groupID, name, isPrivate, maxPos+1, createdBy, now)
	r.logIfInfraErr("CreateCategory", err)

	return cat, err
}

func (r pgRepository) UpdateCategory(
	ctx context.Context,
	groupID, categoryPublicID string,
	name *string,
	isPrivate *bool,
	now time.Time,
) (models.ChannelCategory, error) {
	q := r.db.NewUpdate().Model((*models.ChannelCategoryRow)(nil)).
		Set("updated_at = ?", now).
		Where("group_id = ? AND public_id = ?", groupID, categoryPublicID)
	if name != nil {
		q = q.Set("name = ?", *name)
	}
	if isPrivate != nil {
		q = q.Set("is_private = ?", *isPrivate)
	}
	var row models.ChannelCategoryRow
	res, err := q.Returning("*").Exec(ctx, &row)
	if err != nil {
		if mapped := pgErr.MapNotFound(err); errors.Is(mapped, constants.ErrNotFound) {
			return models.ChannelCategory{}, constants.ErrNotFound
		}
		r.logIfInfraErr("UpdateCategory", err)

		return models.ChannelCategory{}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return models.ChannelCategory{}, constants.ErrNotFound
	}

	return row.ToModel(), nil
}

func (r pgRepository) DeleteCategory(ctx context.Context, groupID, categoryPublicID string) error {
	err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		var row models.ChannelCategoryRow
		if err := tx.NewSelect().Model(&row).
			Where("group_id = ? AND public_id = ?", groupID, categoryPublicID).
			Scan(ctx); err != nil {
			return pgErr.MapNotFound(err)
		}

		// Move channels inside this category to uncategorized before deleting the category.
		if _, err := tx.NewUpdate().Model((*models.ChannelRow)(nil)).
			Set("category_id = NULL").
			Where("group_id = ? AND category_id = ?", groupID, row.ID).
			Exec(ctx); err != nil {
			return err
		}

		res, err := tx.NewDelete().Model((*models.ChannelCategoryRow)(nil)).
			Where("id = ? AND group_id = ?", row.ID, groupID).
			Exec(ctx)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n != 1 {
			return constants.ErrNotFound
		}

		return nil
	})
	if err != nil && !errors.Is(err, constants.ErrNotFound) {
		r.logIfInfraErr("DeleteCategory", err)
	}

	return err
}

func (r pgRepository) ListCategoryPermissionOverrides(
	ctx context.Context,
	groupID, categoryPublicID string,
) ([]models.PublicChannelCategoryPermissionOverride, error) {
	var cat models.ChannelCategoryRow
	if err := r.db.NewSelect().Model(&cat).
		Where("group_id = ? AND public_id = ?", groupID, categoryPublicID).
		Scan(ctx); err != nil {
		return nil, pgErr.MapNotFound(err)
	}

	type overrideJoinRow struct {
		TargetType       string   `bun:"target_type"`
		TargetID         string   `bun:"target_public_id"`
		TargetName       string   `bun:"target_name"`
		TargetColor      string   `bun:"target_color"`
		TargetAvatarURL  string   `bun:"target_avatar_url"`
		IsEveryone       bool     `bun:"is_everyone"`
		AllowPermissions []string `bun:"allow_permissions,array"`
		DenyPermissions  []string `bun:"deny_permissions,array"`
	}

	var rows []overrideJoinRow
	err := r.db.NewSelect().
		TableExpr("channel_category_permission_overrides AS o").
		ColumnExpr("o.target_type").
		ColumnExpr("COALESCE(gr.public_id, u.public_id, '') AS target_public_id").
		ColumnExpr("COALESCE(gr.name, NULLIF(u.display_name, ''), u.username, '') AS target_name").
		ColumnExpr("COALESCE(gr.color, '') AS target_color").
		ColumnExpr("COALESCE(u.avatar_url, '') AS target_avatar_url").
		ColumnExpr("COALESCE(gr.slug = ?, false) AS is_everyone", constants.RoleSlugEveryone).
		ColumnExpr("o.allow_permissions").
		ColumnExpr("o.deny_permissions").
		Join("LEFT JOIN group_roles AS gr ON o.target_type = 'role' AND gr.id = o.target_id AND gr.group_id = ?", groupID).
		Join("LEFT JOIN users AS u ON o.target_type = 'member' AND u.id = o.target_id").
		Where("o.category_id = ?", cat.ID).
		Where("(gr.id IS NOT NULL OR u.id IS NOT NULL)").
		OrderExpr("is_everyone DESC, o.target_type ASC, o.created_at ASC").
		Scan(ctx, &rows)
	if err != nil {
		r.logIfInfraErr("ListCategoryPermissionOverrides", err)

		return nil, err
	}

	out := make([]models.PublicChannelCategoryPermissionOverride, len(rows))
	for i, row := range rows {
		out[i] = models.PublicChannelCategoryPermissionOverride{
			TargetType:       row.TargetType,
			TargetID:         row.TargetID,
			TargetName:       row.TargetName,
			TargetColor:      row.TargetColor,
			TargetAvatarURL:  row.TargetAvatarURL,
			IsEveryone:       row.IsEveryone,
			AllowPermissions: row.AllowPermissions,
			DenyPermissions:  row.DenyPermissions,
		}
	}

	return out, nil
}

func (r pgRepository) ListGroupCategoryPermissionOverrides(
	ctx context.Context,
	groupID string,
) ([]models.ChannelCategoryPermissionOverride, error) {
	type overrideRow struct {
		CategoryID       string   `bun:"category_id"`
		TargetType       string   `bun:"target_type"`
		TargetID         string   `bun:"target_id"`
		IsEveryone       bool     `bun:"is_everyone"`
		AllowPermissions []string `bun:"allow_permissions,array"`
		DenyPermissions  []string `bun:"deny_permissions,array"`
	}

	var rows []overrideRow
	err := r.db.NewSelect().
		TableExpr("channel_category_permission_overrides AS o").
		ColumnExpr("o.category_id").
		ColumnExpr("o.target_type").
		ColumnExpr("o.target_id").
		ColumnExpr("COALESCE(gr.is_default, false) AS is_everyone").
		ColumnExpr("o.allow_permissions").
		ColumnExpr("o.deny_permissions").
		Join("JOIN channel_categories AS cat ON cat.id = o.category_id").
		Join("LEFT JOIN group_roles AS gr ON o.target_type = 'role' AND gr.id = o.target_id AND gr.group_id = ?", groupID).
		Where("cat.group_id = ?", groupID).
		Scan(ctx, &rows)
	if err != nil {
		r.logIfInfraErr("ListGroupCategoryPermissionOverrides", err)

		return nil, err
	}

	out := make([]models.ChannelCategoryPermissionOverride, len(rows))
	for i, row := range rows {
		out[i] = models.ChannelCategoryPermissionOverride{
			CategoryID:       row.CategoryID,
			TargetType:       row.TargetType,
			TargetID:         row.TargetID,
			IsEveryone:       row.IsEveryone,
			AllowPermissions: row.AllowPermissions,
			DenyPermissions:  row.DenyPermissions,
		}
	}

	return out, nil
}

func (r pgRepository) UpsertCategoryPermissionOverride(
	ctx context.Context,
	groupID, categoryPublicID, targetType, targetPublicID string,
	allow, deny []string,
	now time.Time,
) (models.PublicChannelCategoryPermissionOverride, error) {
	var cat models.ChannelCategoryRow
	if err := r.db.NewSelect().Model(&cat).
		Where("group_id = ? AND public_id = ?", groupID, categoryPublicID).
		Scan(ctx); err != nil {
		return models.PublicChannelCategoryPermissionOverride{}, pgErr.MapNotFound(err)
	}

	out := models.PublicChannelCategoryPermissionOverride{
		TargetType:       targetType,
		TargetID:         targetPublicID,
		AllowPermissions: allow,
		DenyPermissions:  deny,
	}

	var internalTargetID string
	switch targetType {
	case "role":
		var role models.GroupRoleRow
		if err := r.db.NewSelect().Model(&role).
			Where("group_id = ? AND public_id = ?", groupID, targetPublicID).
			Scan(ctx); err != nil {
			return models.PublicChannelCategoryPermissionOverride{}, pgErr.MapNotFound(err)
		}
		internalTargetID = role.ID
		out.TargetName = role.Name
		out.TargetColor = role.Color
		out.IsEveryone = role.Slug == constants.RoleSlugEveryone
	case "member":
		var userRow struct {
			ID          string `bun:"id"`
			DisplayName string `bun:"display_name"`
			Username    string `bun:"username"`
			AvatarURL   string `bun:"avatar_url"`
		}
		if err := r.db.NewSelect().
			TableExpr("users AS u").
			ColumnExpr("u.id, u.display_name, u.username, u.avatar_url").
			Join("JOIN group_members AS gm ON gm.user_id = u.id AND gm.group_id = ?", groupID).
			Where("u.public_id = ?", targetPublicID).
			Scan(ctx, &userRow); err != nil {
			return models.PublicChannelCategoryPermissionOverride{}, pgErr.MapNotFound(err)
		}
		internalTargetID = userRow.ID
		out.TargetName = userRow.DisplayName
		if out.TargetName == "" {
			out.TargetName = userRow.Username
		}
		out.TargetAvatarURL = userRow.AvatarURL
	default:
		return models.PublicChannelCategoryPermissionOverride{}, constants.ErrNotFound
	}

	id, err := token.NewID()
	if err != nil {
		return models.PublicChannelCategoryPermissionOverride{}, err
	}
	row := models.ChannelCategoryPermissionOverrideRow{
		ID:               id,
		CategoryID:       cat.ID,
		TargetType:       targetType,
		TargetID:         internalTargetID,
		AllowPermissions: allow,
		DenyPermissions:  deny,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	_, err = r.db.NewInsert().Model(&row).
		On("CONFLICT (category_id, target_type, target_id) DO UPDATE").
		Set("allow_permissions = EXCLUDED.allow_permissions").
		Set("deny_permissions = EXCLUDED.deny_permissions").
		Set("updated_at = EXCLUDED.updated_at").
		Exec(ctx)
	if err != nil {
		r.logIfInfraErr("UpsertCategoryPermissionOverride", err)

		return models.PublicChannelCategoryPermissionOverride{}, err
	}

	return out, nil
}

func (r pgRepository) DeleteCategoryPermissionOverride(
	ctx context.Context,
	groupID, categoryPublicID, targetType, targetPublicID string,
) error {
	var cat models.ChannelCategoryRow
	if err := r.db.NewSelect().Model(&cat).
		Where("group_id = ? AND public_id = ?", groupID, categoryPublicID).
		Scan(ctx); err != nil {
		return pgErr.MapNotFound(err)
	}

	var internalTargetID string
	switch targetType {
	case "role":
		var role models.GroupRoleRow
		if err := r.db.NewSelect().Model(&role).
			Where("group_id = ? AND public_id = ?", groupID, targetPublicID).
			Scan(ctx); err != nil {
			return pgErr.MapNotFound(err)
		}
		internalTargetID = role.ID
	case "member":
		if err := r.db.NewSelect().
			TableExpr("users").
			ColumnExpr("id").
			Where("public_id = ?", targetPublicID).
			Scan(ctx, &internalTargetID); err != nil {
			return pgErr.MapNotFound(err)
		}
	default:
		return constants.ErrNotFound
	}

	_, err := r.db.NewDelete().Model((*models.ChannelCategoryPermissionOverrideRow)(nil)).
		Where("category_id = ? AND target_type = ? AND target_id = ?", cat.ID, targetType, internalTargetID).
		Exec(ctx)
	if err != nil {
		r.logIfInfraErr("DeleteCategoryPermissionOverride", err)
	}

	return err
}

func (r pgRepository) SetCategoryPositions(ctx context.Context, groupID string, orderedCategoryIDs []string) error {
	err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		for i, id := range orderedCategoryIDs {
			res, err := tx.NewUpdate().Model((*models.ChannelCategoryRow)(nil)).
				Set("position = ?", i).
				Where("group_id = ? AND id = ?", groupID, id).
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
	r.logIfInfraErr("SetCategoryPositions", err)

	return err
}

func (r pgRepository) logIfInfraErr(op string, err error) {
	if err != nil && !errors.Is(err, constants.ErrNotFound) && !errors.Is(err, constants.ErrConflict) {
		r.log.Error("channel category repository", "op", op, "err", err)
	}
}

func insertCategoryWithRetry(
	ctx context.Context,
	db bun.IDB,
	groupID, name string,
	isPrivate bool,
	position int32,
	createdBy string,
	now time.Time,
) (models.ChannelCategory, error) {
	for range 5 {
		categoryID, err := token.NewID()
		if err != nil {
			return models.ChannelCategory{}, err
		}
		publicID, err := token.NewPublicID()
		if err != nil {
			return models.ChannelCategory{}, err
		}
		cat := models.ChannelCategory{
			ID: categoryID, PublicID: publicID, GroupID: groupID,
			Name: name, IsPrivate: isPrivate, Position: position,
			CreatedBy: createdBy, CreatedAt: now, UpdatedAt: now,
		}
		row := models.ChannelCategoryRowFrom(cat)
		_, err = db.NewInsert().Model(&row).Exec(ctx)
		if err == nil {
			return cat, nil
		}
		if errors.Is(pgErr.MapUnique(err), constants.ErrConflict) {
			continue
		}

		return models.ChannelCategory{}, err
	}

	return models.ChannelCategory{}, constants.ErrConflict
}
