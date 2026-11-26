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

// PgRepository loads and persists channel metadata.
type PgRepository interface {
	ListByGroupID(ctx context.Context, groupID string) ([]models.Channel, error)
	CreateChannel(ctx context.Context, groupID, categoryID, channelType, name string, isPrivate bool, createdBy string, now time.Time) (models.Channel, error)
	UpdateChannel(ctx context.Context, groupID, channelPublicID string, input models.UpdateChannelInput, now time.Time) (models.Channel, error)
	DeleteChannel(ctx context.Context, groupID, channelPublicID string) error
	ListChannelPermissionOverrides(ctx context.Context, groupID, channelPublicID string) ([]models.PublicChannelCategoryPermissionOverride, error)
	ListGroupChannelPermissionOverrides(ctx context.Context, groupID string) ([]models.ChannelPermissionOverride, error)
	UpsertChannelPermissionOverride(ctx context.Context, groupID, channelPublicID, targetType, targetPublicID string, allow, deny []string, now time.Time) (models.PublicChannelCategoryPermissionOverride, error)
	DeleteChannelPermissionOverride(ctx context.Context, groupID, channelPublicID, targetType, targetPublicID string) error
	SyncChannelPermissionsWithCategory(ctx context.Context, groupID, channelPublicID string, now time.Time) (models.Channel, error)
	SetChannelPositions(ctx context.Context, groupID string, sidebarSection, categoryID *string, orderedChannelIDs []string) error
	FindByPublicIDInGroup(ctx context.Context, groupID, channelPublicID string) (models.Channel, error)
	FindByID(ctx context.Context, channelID string) (models.Channel, error)
	FindGroupDmChannelBetweenUsers(ctx context.Context, groupID, userID, peerUserID string) (models.Channel, error)
	CreateGroupDmChannel(ctx context.Context, groupID, name, userID, peerUserID, createdBy string, now time.Time) (models.Channel, error)
	DmChannelIDSetForUser(ctx context.Context, groupID, userID string) (map[string]struct{}, error)
	DmPeerUserIDsForCaller(ctx context.Context, groupID, callerUserID string, channelIDs []string) (map[string]string, error)
	SeedDefaultChannels(ctx context.Context, groupID, createdBy string, createdAt time.Time) error
	SeedDefaultChannelsWithTemplate(ctx context.Context, groupID, createdBy, template string, createdAt time.Time) error
}

type pgRepository struct {
	db  *bun.DB
	log logger.Logger
}

// NewPgRepository constructs a Postgres channel store.
func NewPgRepository(db *bun.DB, log logger.Logger) PgRepository {
	return pgRepository{db: db, log: log}
}

func (r pgRepository) ListByGroupID(ctx context.Context, groupID string) ([]models.Channel, error) {
	var rows []models.ChannelRow
	err := r.db.NewSelect().Model(&rows).
		ColumnExpr("channel_row.*").
		ColumnExpr("COALESCE(cat.public_id, '') AS category_public_id").
		Join("LEFT JOIN channel_categories AS cat ON cat.id = channel_row.category_id").
		Where("channel_row.group_id = ?", groupID).
		OrderExpr("channel_row.position ASC, channel_row.name ASC").
		Scan(ctx)
	if err != nil {
		r.logIfInfraErr("ListByGroupID", err)

		return nil, err
	}

	out := make([]models.Channel, len(rows))
	for i, row := range rows {
		out[i] = row.ToModel()
	}

	return out, nil
}

func (r pgRepository) CreateChannel(
	ctx context.Context,
	groupID, categoryID, channelType, name string,
	isPrivate bool,
	createdBy string,
	now time.Time,
) (models.Channel, error) {
	var maxPos int32
	posQuery := r.db.NewSelect().
		TableExpr("channels").
		ColumnExpr("COALESCE(MAX(position), -1)").
		Where("group_id = ?", groupID)
	if categoryID != "" {
		posQuery = posQuery.Where("category_id = ?", categoryID)
	} else {
		posQuery = posQuery.Where("category_id IS NULL")
	}
	if err := posQuery.Scan(ctx, &maxPos); err != nil {
		r.logIfInfraErr("CreateChannel", err)

		return models.Channel{}, err
	}

	var categoryPublicID string
	if categoryID != "" {
		err := r.db.NewSelect().
			TableExpr("channel_categories").
			Column("public_id").
			Where("group_id = ? AND id = ?", groupID, categoryID).
			Scan(ctx, &categoryPublicID)
		if err != nil {
			r.logIfInfraErr("CreateChannel", err)

			return models.Channel{}, pgErr.MapNotFound(err)
		}
	}

	for range 5 {
		channelID, err := token.NewID()
		if err != nil {
			return models.Channel{}, err
		}
		publicID, err := token.NewPublicID()
		if err != nil {
			return models.Channel{}, err
		}
		ch := models.Channel{
			ID:               channelID,
			PublicID:         publicID,
			GroupID:          groupID,
			CategoryID:       categoryID,
			CategoryPublicID: categoryPublicID,
			Type:             channelType,
			IsPrivate:        isPrivate,
			Name:             name,
			SidebarSection:   constants.SidebarSectionChannels,
			Position:         maxPos + 1,
			CreatedBy:        createdBy,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		row := models.ChannelRowFrom(ch)
		_, err = r.db.NewInsert().Model(&row).Exec(ctx)
		if err == nil {
			return ch, nil
		}
		if !errors.Is(pgErr.MapUnique(err), constants.ErrConflict) {
			r.logIfInfraErr("CreateChannel", err)

			return models.Channel{}, err
		}
	}

	return models.Channel{}, constants.ErrConflict
}

func (r pgRepository) SetChannelPositions(ctx context.Context, groupID string, sidebarSection, categoryID *string, orderedChannelIDs []string) error {
	return r.setPositions(ctx, "SetChannelPositions", (*models.ChannelRow)(nil), groupID, sidebarSection, categoryID, orderedChannelIDs)
}

func (r pgRepository) UpdateChannel(
	ctx context.Context,
	groupID, channelPublicID string,
	input models.UpdateChannelInput,
	now time.Time,
) (models.Channel, error) {
	q := r.db.NewUpdate().Model((*models.ChannelRow)(nil)).
		Set("updated_at = ?", now).
		Where("group_id = ? AND public_id = ?", groupID, channelPublicID)

	if input.Name != nil {
		q = q.Set("name = ?", *input.Name)
	}
	if input.IsPrivate != nil {
		q = q.Set("is_private = ?", *input.IsPrivate)
	}
	if input.SlowmodeSeconds != nil {
		q = q.Set("slowmode_seconds = ?", *input.SlowmodeSeconds)
	}
	if input.ThreadSlowmodeSeconds != nil {
		q = q.Set("thread_slowmode_seconds = ?", *input.ThreadSlowmodeSeconds)
	}
	if input.ContentVisibility != nil {
		q = q.Set("content_visibility = ?", *input.ContentVisibility)
	}
	if input.ForumDefaultLayout != nil {
		q = q.Set("forum_default_layout = ?", *input.ForumDefaultLayout)
	}
	if input.ForumSortOrder != nil {
		q = q.Set("forum_sort_order = ?", *input.ForumSortOrder)
	}
	if input.ForumTagMatching != nil {
		q = q.Set("forum_tag_matching = ?", *input.ForumTagMatching)
	}
	if input.BitrateKbps != nil {
		q = q.Set("bitrate_kbps = ?", *input.BitrateKbps)
	}
	if input.VideoQuality != nil {
		q = q.Set("video_quality = ?", *input.VideoQuality)
	}
	if input.UserLimit != nil {
		q = q.Set("user_limit = ?", *input.UserLimit)
	}
	if input.RTCRegion != nil {
		q = q.Set("rtc_region = ?", *input.RTCRegion)
	}

	res, err := q.Exec(ctx)
	if err != nil {
		r.logIfInfraErr("UpdateChannel", err)
		return models.Channel{}, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return models.Channel{}, err
	}
	if n == 0 {
		return models.Channel{}, constants.ErrNotFound
	}

	return r.FindByPublicIDInGroup(ctx, groupID, channelPublicID)
}

func (r pgRepository) DeleteChannel(ctx context.Context, groupID, channelPublicID string) error {
	res, err := r.db.NewDelete().Model((*models.ChannelRow)(nil)).
		Where("group_id = ? AND public_id = ?", groupID, channelPublicID).
		Exec(ctx)
	if err != nil {
		r.logIfInfraErr("DeleteChannel", err)
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return constants.ErrNotFound
	}

	return nil
}

func (r pgRepository) ListChannelPermissionOverrides(
	ctx context.Context,
	groupID, channelPublicID string,
) ([]models.PublicChannelCategoryPermissionOverride, error) {
	ch, err := r.FindByPublicIDInGroup(ctx, groupID, channelPublicID)
	if err != nil {
		return nil, err
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
	err = r.db.NewSelect().
		TableExpr("channel_permission_overrides AS o").
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
		Where("o.channel_id = ?", ch.ID).
		Where("(gr.id IS NOT NULL OR u.id IS NOT NULL)").
		OrderExpr("is_everyone DESC, o.target_type ASC, o.created_at ASC").
		Scan(ctx, &rows)
	if err != nil {
		r.logIfInfraErr("ListChannelPermissionOverrides", err)
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

func (r pgRepository) ListGroupChannelPermissionOverrides(
	ctx context.Context,
	groupID string,
) ([]models.ChannelPermissionOverride, error) {
	type overrideRow struct {
		ChannelID        string   `bun:"channel_id"`
		TargetType       string   `bun:"target_type"`
		TargetID         string   `bun:"target_id"`
		IsEveryone       bool     `bun:"is_everyone"`
		AllowPermissions []string `bun:"allow_permissions,array"`
		DenyPermissions  []string `bun:"deny_permissions,array"`
	}

	var rows []overrideRow
	err := r.db.NewSelect().
		TableExpr("channel_permission_overrides AS o").
		ColumnExpr("o.channel_id").
		ColumnExpr("o.target_type").
		ColumnExpr("o.target_id").
		ColumnExpr("COALESCE(gr.is_default, false) AS is_everyone").
		ColumnExpr("o.allow_permissions").
		ColumnExpr("o.deny_permissions").
		Join("JOIN channels AS ch ON ch.id = o.channel_id").
		Join("LEFT JOIN group_roles AS gr ON o.target_type = 'role' AND gr.id = o.target_id AND gr.group_id = ?", groupID).
		Where("ch.group_id = ?", groupID).
		Scan(ctx, &rows)
	if err != nil {
		r.logIfInfraErr("ListGroupChannelPermissionOverrides", err)

		return nil, err
	}

	out := make([]models.ChannelPermissionOverride, len(rows))
	for i, row := range rows {
		out[i] = models.ChannelPermissionOverride{
			ChannelID:        row.ChannelID,
			TargetType:       row.TargetType,
			TargetID:         row.TargetID,
			IsEveryone:       row.IsEveryone,
			AllowPermissions: row.AllowPermissions,
			DenyPermissions:  row.DenyPermissions,
		}
	}

	return out, nil
}

func (r pgRepository) UpsertChannelPermissionOverride(
	ctx context.Context,
	groupID, channelPublicID, targetType, targetPublicID string,
	allow, deny []string,
	now time.Time,
) (models.PublicChannelCategoryPermissionOverride, error) {
	ch, err := r.FindByPublicIDInGroup(ctx, groupID, channelPublicID)
	if err != nil {
		return models.PublicChannelCategoryPermissionOverride{}, err
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
	row := models.ChannelPermissionOverrideRow{
		ID:               id,
		ChannelID:        ch.ID,
		TargetType:       targetType,
		TargetID:         internalTargetID,
		AllowPermissions: allow,
		DenyPermissions:  deny,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	_, err = r.db.NewInsert().Model(&row).
		On("CONFLICT (channel_id, target_type, target_id) DO UPDATE").
		Set("allow_permissions = EXCLUDED.allow_permissions").
		Set("deny_permissions = EXCLUDED.deny_permissions").
		Set("updated_at = EXCLUDED.updated_at").
		Exec(ctx)
	if err != nil {
		r.logIfInfraErr("UpsertChannelPermissionOverride", err)
		return models.PublicChannelCategoryPermissionOverride{}, err
	}

	return out, nil
}

func (r pgRepository) DeleteChannelPermissionOverride(
	ctx context.Context,
	groupID, channelPublicID, targetType, targetPublicID string,
) error {
	ch, err := r.FindByPublicIDInGroup(ctx, groupID, channelPublicID)
	if err != nil {
		return err
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

	_, err = r.db.NewDelete().Model((*models.ChannelPermissionOverrideRow)(nil)).
		Where("channel_id = ? AND target_type = ? AND target_id = ?", ch.ID, targetType, internalTargetID).
		Exec(ctx)
	if err != nil {
		r.logIfInfraErr("DeleteChannelPermissionOverride", err)
	}

	return err
}

func (r pgRepository) SyncChannelPermissionsWithCategory(
	ctx context.Context,
	groupID, channelPublicID string,
	now time.Time,
) (models.Channel, error) {
	ch, err := r.FindByPublicIDInGroup(ctx, groupID, channelPublicID)
	if err != nil {
		return models.Channel{}, err
	}
	if ch.CategoryID == "" {
		return models.Channel{}, constants.ErrNotFound
	}

	var cat models.ChannelCategoryRow
	if err := r.db.NewSelect().Model(&cat).
		Where("group_id = ? AND id = ?", groupID, ch.CategoryID).
		Scan(ctx); err != nil {
		return models.Channel{}, pgErr.MapNotFound(err)
	}

	err = r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewUpdate().Model((*models.ChannelRow)(nil)).
			Set("is_private = ?", cat.IsPrivate).
			Set("updated_at = ?", now).
			Where("group_id = ? AND id = ?", groupID, ch.ID).
			Exec(ctx); err != nil {
			return err
		}
		if _, err := tx.NewDelete().Model((*models.ChannelPermissionOverrideRow)(nil)).
			Where("channel_id = ?", ch.ID).
			Exec(ctx); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		r.logIfInfraErr("SyncChannelPermissionsWithCategory", err)
		return models.Channel{}, err
	}

	return r.FindByPublicIDInGroup(ctx, groupID, channelPublicID)
}

func (r pgRepository) setPositions(
	ctx context.Context,
	op string,
	model any,
	groupID string,
	sidebarSection *string,
	categoryID *string,
	orderedIDs []string,
) error {
	err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		for i, id := range orderedIDs {
			q := tx.NewUpdate().Model(model).
				Set("position = ?", i).
				Where("group_id = ? AND id = ?", groupID, id)
			if sidebarSection != nil {
				q = q.Set("sidebar_section = ?", *sidebarSection)
				if *sidebarSection == constants.SidebarSectionFavorites {
					q = q.Set("category_id = NULL")
				}
			}
			if categoryID != nil {
				q = q.Set("category_id = ?", *categoryID)
			}
			res, err := q.Exec(ctx)
			if err != nil {
				return err
			}
			if n, _ := res.RowsAffected(); n != 1 {
				return constants.ErrNotFound
			}
		}

		return nil
	})
	r.logIfInfraErr(op, err)

	return err
}

func (r pgRepository) FindByPublicIDInGroup(ctx context.Context, groupID, channelPublicID string) (models.Channel, error) {
	var row models.ChannelRow
	err := r.db.NewSelect().Model(&row).
		Where("group_id = ? AND public_id = ?", groupID, channelPublicID).
		Scan(ctx)
	if err != nil {
		r.logIfInfraErr("FindByPublicIDInGroup", err)

		return models.Channel{}, pgErr.MapNotFound(err)
	}

	return row.ToModel(), nil
}

// FindByID loads one channel by internal id.
func (r pgRepository) FindByID(ctx context.Context, channelID string) (models.Channel, error) {
	var row models.ChannelRow
	err := r.db.NewSelect().Model(&row).
		Where("id = ?", channelID).
		Scan(ctx)
	if err != nil {
		r.logIfInfraErr("FindByID", err)

		return models.Channel{}, pgErr.MapNotFound(err)
	}

	return row.ToModel(), nil
}

func (r pgRepository) FindGroupDmChannelBetweenUsers(ctx context.Context, groupID, userID, peerUserID string) (models.Channel, error) {
	var row models.ChannelRow
	err := r.db.NewSelect().Model(&row).
		Join("JOIN dm_participants AS p1 ON p1.channel_id = channel_row.id AND p1.user_id = ?", userID).
		Join("JOIN dm_participants AS p2 ON p2.channel_id = channel_row.id AND p2.user_id = ?", peerUserID).
		Where("channel_row.group_id = ? AND channel_row.type = ?", groupID, constants.ChannelKindDM).
		Limit(1).
		Scan(ctx)
	if err != nil {
		mapped := pgErr.MapNotFound(err)
		if !errors.Is(mapped, constants.ErrNotFound) {
			r.logIfInfraErr("FindGroupDmChannelBetweenUsers", err)
		}

		return models.Channel{}, mapped
	}

	return row.ToModel(), nil
}

func (r pgRepository) CreateGroupDmChannel(
	ctx context.Context,
	groupID, name, userID, peerUserID, createdBy string,
	now time.Time,
) (models.Channel, error) {
	var maxPos int32
	err := r.db.NewSelect().
		TableExpr("channels").
		ColumnExpr("COALESCE(MAX(position), -1)").
		Where("group_id = ? AND sidebar_section = ?", groupID, constants.SidebarSectionChats).
		Scan(ctx, &maxPos)
	if err != nil {
		r.logIfInfraErr("CreateGroupDmChannel", err)

		return models.Channel{}, err
	}

	var out models.Channel
	err = r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		for range 5 {
			channelID, idErr := token.NewID()
			if idErr != nil {
				return idErr
			}
			publicID, idErr := token.NewPublicID()
			if idErr != nil {
				return idErr
			}
			ch := models.Channel{
				ID:             channelID,
				PublicID:       publicID,
				GroupID:        groupID,
				Type:           constants.ChannelKindDM,
				IsPrivate:      false,
				Name:           name,
				SidebarSection: constants.SidebarSectionChats,
				Position:       maxPos + 1,
				CreatedBy:      createdBy,
				CreatedAt:      now,
				UpdatedAt:      now,
			}
			row := models.ChannelRowFrom(ch)
			if _, err := tx.NewInsert().Model(&row).Exec(ctx); err != nil {
				if errors.Is(pgErr.MapUnique(err), constants.ErrConflict) {
					continue
				}
				return err
			}
			for _, uid := range []string{userID, peerUserID} {
				part := models.DmParticipantRow{ChannelID: channelID, UserID: uid}
				if _, err := tx.NewInsert().Model(&part).Exec(ctx); err != nil {
					return err
				}
			}
			out = ch
			return nil
		}
		return constants.ErrConflict
	})
	if err != nil {
		r.logIfInfraErr("CreateGroupDmChannel", err)

		return models.Channel{}, err
	}

	return out, nil
}

func (r pgRepository) DmPeerUserIDsForCaller(
	ctx context.Context,
	groupID, callerUserID string,
	channelIDs []string,
) (map[string]string, error) {
	out := make(map[string]string)
	if len(channelIDs) == 0 {
		return out, nil
	}

	type peerRow struct {
		ChannelID string `bun:"channel_id"`
		UserID    string `bun:"user_id"`
	}
	var rows []peerRow
	err := r.db.NewSelect().
		TableExpr("dm_participants AS p").
		ColumnExpr("p.channel_id, p.user_id").
		Join("JOIN channels AS c ON c.id = p.channel_id").
		Where("c.group_id = ?", groupID).
		Where("c.type = ?", constants.ChannelKindDM).
		Where("p.channel_id IN (?)", bun.List(channelIDs)).
		Where("p.user_id <> ?", callerUserID).
		Scan(ctx, &rows)
	if err != nil {
		r.logIfInfraErr("DmPeerUserIDsForCaller", err)

		return nil, err
	}
	for _, row := range rows {
		out[row.ChannelID] = row.UserID
	}

	return out, nil
}

func (r pgRepository) DmChannelIDSetForUser(ctx context.Context, groupID, userID string) (map[string]struct{}, error) {
	var ids []string
	err := r.db.NewSelect().
		TableExpr("dm_participants AS p").
		ColumnExpr("p.channel_id").
		Join("JOIN channels AS c ON c.id = p.channel_id").
		Where("c.group_id = ? AND c.type = ?", groupID, constants.ChannelKindDM).
		Where("p.user_id = ?", userID).
		Scan(ctx, &ids)
	if err != nil {
		r.logIfInfraErr("DmChannelIDSetForUser", err)

		return nil, err
	}
	out := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		out[id] = struct{}{}
	}

	return out, nil
}

func (r pgRepository) SeedDefaultChannels(ctx context.Context, groupID, createdBy string, createdAt time.Time) error {
	return r.SeedDefaultChannelsWithTemplate(ctx, groupID, createdBy, "", createdAt)
}

func (r pgRepository) SeedDefaultChannelsWithTemplate(
	ctx context.Context,
	groupID, createdBy, template string,
	createdAt time.Time,
) error {
	err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		return InsertDefaultChannelsWithTemplateTx(ctx, tx, groupID, createdBy, template, createdAt)
	})
	r.logIfInfraErr("SeedDefaultChannelsWithTemplate", err)

	return err
}

func (r pgRepository) logIfInfraErr(op string, err error) {
	if err == nil {
		return
	}
	r.log.Error("channel repository", "op", op, "err", err)
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

// InsertDefaultChannelTx creates a seeded public channel inside an existing transaction.
func InsertDefaultChannelTx(
	ctx context.Context,
	tx bun.Tx,
	groupID, categoryID, createdBy, channelType, name, sidebarSection string,
	now time.Time,
) error {
	for range 5 {
		channelID, err := token.NewID()
		if err != nil {
			return err
		}
		publicId, err := token.NewPublicID()
		if err != nil {
			return err
		}
		row := models.ChannelRowFrom(models.Channel{
			ID: channelID, PublicID: publicId, GroupID: groupID, CategoryID: categoryID,
			Type: channelType, IsPrivate: false, Name: name, SidebarSection: sidebarSection,
			CreatedBy: createdBy, CreatedAt: now, UpdatedAt: now,
		})
		_, err = tx.NewInsert().Model(&row).Exec(ctx)
		if err == nil {
			return nil
		}
		if errors.Is(pgErr.MapUnique(err), constants.ErrConflict) {
			continue
		}

		return err
	}

	return constants.ErrConflict
}

// InsertAnnouncementChannelTx creates the default announcements channel inside an existing transaction.
func InsertAnnouncementChannelTx(ctx context.Context, tx bun.Tx, groupID, createdBy string, now time.Time) error {
	return InsertDefaultChannelTx(
		ctx, tx, groupID, "", createdBy, constants.ChannelKindBroadcast,
		constants.DefaultAnnouncementChannelName, constants.SidebarSectionFavorites, now,
	)
}

// TemplateChannelSpec defines a seeded channel inside a template category.
type TemplateChannelSpec struct {
	Type string
	Name string
}

// TemplateCategorySpec defines a seeded category and its channels for a group template.
type TemplateCategorySpec struct {
	Name     string
	Channels []TemplateChannelSpec
}

// TemplateCategories returns the category and channel blueprint for a group template key.
func TemplateCategories(template string) []TemplateCategorySpec {
	switch template {
	case "office":
		return []TemplateCategorySpec{
			{
				Name: "Office",
				Channels: []TemplateChannelSpec{
					{Type: constants.ChannelKindBroadcast, Name: "announcements"},
					{Type: constants.ChannelKindTextOnly, Name: "general"},
					{Type: constants.ChannelKindForum, Name: "help-desk"},
				},
			},
			{
				Name: "Workspaces",
				Channels: []TemplateChannelSpec{
					{Type: constants.ChannelKindTextOnly, Name: "operations"},
					{Type: constants.ChannelKindTextOnly, Name: "projects"},
					{Type: constants.ChannelKindTextOnly, Name: "watercooler"},
				},
			},
			{
				Name: "Meeting Rooms",
				Channels: []TemplateChannelSpec{
					{Type: constants.ChannelKindVoiceOnly, Name: "Standup Room"},
					{Type: constants.ChannelKindVoiceOnly, Name: "Conference Room"},
				},
			},
		}
	case "project":
		return []TemplateCategorySpec{
			{
				Name: "Overview",
				Channels: []TemplateChannelSpec{
					{Type: constants.ChannelKindBroadcast, Name: "updates"},
					{Type: constants.ChannelKindTextOnly, Name: "general"},
				},
			},
			{
				Name: "Execution",
				Channels: []TemplateChannelSpec{
					{Type: constants.ChannelKindTextOnly, Name: "tasks-and-progress"},
					{Type: constants.ChannelKindForum, Name: "feedback-and-bugs"},
					{Type: constants.ChannelKindTextOnly, Name: "resources"},
				},
			},
			{
				Name: "Calls",
				Channels: []TemplateChannelSpec{
					{Type: constants.ChannelKindVoiceOnly, Name: "Sprint Sync"},
					{Type: constants.ChannelKindVoiceOnly, Name: "Coworking"},
				},
			},
		}
	case "study":
		return []TemplateCategorySpec{
			{
				Name: "Class Info",
				Channels: []TemplateChannelSpec{
					{Type: constants.ChannelKindBroadcast, Name: "announcements"},
					{Type: constants.ChannelKindTextOnly, Name: "introductions"},
				},
			},
			{
				Name: "Study & Discussion",
				Channels: []TemplateChannelSpec{
					{Type: constants.ChannelKindTextOnly, Name: "general"},
					{Type: constants.ChannelKindForum, Name: "q-and-a"},
					{Type: constants.ChannelKindTextOnly, Name: "notes-and-links"},
				},
			},
			{
				Name: "Study Rooms",
				Channels: []TemplateChannelSpec{
					{Type: constants.ChannelKindVoiceOnly, Name: "Silent Study"},
					{Type: constants.ChannelKindVoiceOnly, Name: "Group Discussion"},
				},
			},
		}
	case "gaming":
		return []TemplateCategorySpec{
			{
				Name: "Lobby",
				Channels: []TemplateChannelSpec{
					{Type: constants.ChannelKindBroadcast, Name: "announcements"},
					{Type: constants.ChannelKindTextOnly, Name: "welcome-and-rules"},
				},
			},
			{
				Name: "Community",
				Channels: []TemplateChannelSpec{
					{Type: constants.ChannelKindTextOnly, Name: "general"},
					{Type: constants.ChannelKindForum, Name: "looking-for-group"},
					{Type: constants.ChannelKindTextOnly, Name: "clips-and-screenshots"},
				},
			},
			{
				Name: "Voice Channels",
				Channels: []TemplateChannelSpec{
					{Type: constants.ChannelKindVoiceOnly, Name: "Main Lobby"},
					{Type: constants.ChannelKindVoiceOnly, Name: "Squad 1"},
					{Type: constants.ChannelKindVoiceOnly, Name: "Squad 2"},
				},
			},
		}
	default:
		return []TemplateCategorySpec{
			{
				Name: constants.DefaultFavoritesCategoryName,
				Channels: []TemplateChannelSpec{
					{Type: constants.ChannelKindBroadcast, Name: constants.DefaultAnnouncementChannelName},
				},
			},
			{
				Name: constants.DefaultTextCategoryName,
				Channels: []TemplateChannelSpec{
					{Type: constants.ChannelKindTextOnly, Name: constants.DefaultChannelName},
				},
			},
			{
				Name: constants.DefaultVoiceCategoryName,
				Channels: []TemplateChannelSpec{
					{Type: constants.ChannelKindVoiceOnly, Name: constants.DefaultChannelName},
				},
			},
		}
	}
}

// InsertDefaultChannelsTx seeds Favorites (Announcements), Text Channels (#General), and Voice Channels (General).
func InsertDefaultChannelsTx(ctx context.Context, tx bun.Tx, groupID, createdBy string, now time.Time) error {
	return InsertDefaultChannelsWithTemplateTx(ctx, tx, groupID, createdBy, "", now)
}

// InsertDefaultChannelsWithTemplateTx seeds categories and channels for the requested group template.
func InsertDefaultChannelsWithTemplateTx(
	ctx context.Context,
	tx bun.Tx,
	groupID, createdBy, template string,
	now time.Time,
) error {
	specs := TemplateCategories(template)
	var catPos int32
	for _, catSpec := range specs {
		cat, err := insertCategoryWithRetry(
			ctx, tx, groupID, catSpec.Name, false, catPos, createdBy, now,
		)
		if err != nil {
			return err
		}
		catPos++
		var chPos int32
		for _, chSpec := range catSpec.Channels {
			if err := insertDefaultChannelWithPosTx(
				ctx, tx, groupID, cat.ID, createdBy, chSpec.Type, chSpec.Name,
				constants.SidebarSectionChannels, chPos, now,
			); err != nil {
				return err
			}
			chPos++
		}
	}

	return nil
}

func insertDefaultChannelWithPosTx(
	ctx context.Context,
	tx bun.Tx,
	groupID, categoryID, createdBy, channelType, name, sidebarSection string,
	position int32,
	now time.Time,
) error {
	for range 5 {
		channelID, err := token.NewID()
		if err != nil {
			return err
		}
		publicId, err := token.NewPublicID()
		if err != nil {
			return err
		}
		row := models.ChannelRowFrom(models.Channel{
			ID: channelID, PublicID: publicId, GroupID: groupID, CategoryID: categoryID,
			Type: channelType, IsPrivate: false, Name: name, SidebarSection: sidebarSection,
			Position: position, CreatedBy: createdBy, CreatedAt: now, UpdatedAt: now,
		})
		_, err = tx.NewInsert().Model(&row).Exec(ctx)
		if err == nil {
			return nil
		}
		if errors.Is(pgErr.MapUnique(err), constants.ErrConflict) {
			continue
		}

		return err
	}

	return constants.ErrConflict
}
