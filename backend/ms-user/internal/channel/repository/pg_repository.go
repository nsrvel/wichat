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
		Where("group_id = ?", groupID).
		OrderExpr("name ASC").
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

// InsertDefaultChannelTx creates a seeded public channel inside an existing transaction.
func InsertDefaultChannelTx(
	ctx context.Context,
	tx bun.Tx,
	groupID, createdBy, channelType, name, sidebarSection string,
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
			ID: channelID, PublicID: publicId, GroupID: groupID,
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

// InsertGeneralChannelTx creates the default #general channel inside an existing transaction.
func InsertGeneralChannelTx(ctx context.Context, tx bun.Tx, groupID, createdBy string, now time.Time) error {
	return InsertDefaultChannelTx(
		ctx, tx, groupID, createdBy, constants.ChannelKindTextAndVoice, constants.DefaultChannelName, constants.SidebarSectionChannels, now,
	)
}

// InsertAnnouncementChannelTx creates the default announcements channel inside an existing transaction.
func InsertAnnouncementChannelTx(ctx context.Context, tx bun.Tx, groupID, createdBy string, now time.Time) error {
	return InsertDefaultChannelTx(
		ctx, tx, groupID, createdBy, constants.ChannelKindBroadcast,
		constants.DefaultAnnouncementChannelName, constants.SidebarSectionFavorites, now,
	)
}

// InsertDefaultChannelsTx seeds general and announcements — same path as new group create.
func InsertDefaultChannelsTx(ctx context.Context, tx bun.Tx, groupID, createdBy string, now time.Time) error {
	if err := InsertGeneralChannelTx(ctx, tx, groupID, createdBy, now); err != nil {
		return err
	}

	return InsertAnnouncementChannelTx(ctx, tx, groupID, createdBy, now)
}
