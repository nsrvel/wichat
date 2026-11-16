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
	ListCategoriesByGroupID(ctx context.Context, groupID string) ([]models.ChannelCategory, error)
	CreateCategory(ctx context.Context, groupID, name string, isPrivate bool, createdBy string, now time.Time) (models.ChannelCategory, error)
	CreateChannel(ctx context.Context, groupID, categoryID, channelType, name string, isPrivate bool, createdBy string, now time.Time) (models.Channel, error)
	SetCategoryPositions(ctx context.Context, groupID string, orderedCategoryIDs []string) error
	SetChannelPositions(ctx context.Context, groupID string, sidebarSection, categoryID *string, orderedChannelIDs []string) error
	FindByPublicIDInGroup(ctx context.Context, groupID, channelPublicID string) (models.Channel, error)
	FindByID(ctx context.Context, channelID string) (models.Channel, error)
	SeedDefaultChannels(ctx context.Context, groupID, createdBy string, createdAt time.Time) error
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

func (r pgRepository) SetCategoryPositions(ctx context.Context, groupID string, orderedCategoryIDs []string) error {
	return r.setPositions(ctx, "SetCategoryPositions", (*models.ChannelCategoryRow)(nil), groupID, nil, nil, orderedCategoryIDs)
}

func (r pgRepository) SetChannelPositions(ctx context.Context, groupID string, sidebarSection, categoryID *string, orderedChannelIDs []string) error {
	return r.setPositions(ctx, "SetChannelPositions", (*models.ChannelRow)(nil), groupID, sidebarSection, categoryID, orderedChannelIDs)
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

func (r pgRepository) SeedDefaultChannels(ctx context.Context, groupID, createdBy string, createdAt time.Time) error {
	err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		return InsertDefaultChannelsTx(ctx, tx, groupID, createdBy, createdAt)
	})
	r.logIfInfraErr("SeedDefaultChannels", err)

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

// InsertDefaultChannelsTx seeds Text Channels (#General), Voice Channels (General), and Announcements.
func InsertDefaultChannelsTx(ctx context.Context, tx bun.Tx, groupID, createdBy string, now time.Time) error {
	textCat, err := insertCategoryWithRetry(
		ctx, tx, groupID, constants.DefaultTextCategoryName, false, 0, createdBy, now,
	)
	if err != nil {
		return err
	}
	if err := InsertDefaultChannelTx(
		ctx, tx, groupID, textCat.ID, createdBy, constants.ChannelKindTextOnly,
		constants.DefaultChannelName, constants.SidebarSectionChannels, now,
	); err != nil {
		return err
	}

	voiceCat, err := insertCategoryWithRetry(
		ctx, tx, groupID, constants.DefaultVoiceCategoryName, false, 1, createdBy, now,
	)
	if err != nil {
		return err
	}
	if err := InsertDefaultChannelTx(
		ctx, tx, groupID, voiceCat.ID, createdBy, constants.ChannelKindVoiceOnly,
		constants.DefaultChannelName, constants.SidebarSectionChannels, now,
	); err != nil {
		return err
	}

	return InsertAnnouncementChannelTx(ctx, tx, groupID, createdBy, now)
}
