package repository

import (
	"context"
	"errors"
	"time"

	"github.com/uptrace/bun"
	"github.com/wichat/wichat/backend/ms-chat/internal/models"
	"github.com/wichat/wichat/backend/ms-chat/pkg/constants"
	pgErr "github.com/wichat/wichat/backend/ms-chat/pkg/db"
	"github.com/wichat/wichat/backend/ms-chat/pkg/token"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

// PgRepository persists channel messages and reactions.
type PgRepository interface {
	ListByChannel(ctx context.Context, channelID string, limit int, before *time.Time) ([]models.MessageWithAuthor, error)
	ListPinnedByChannel(ctx context.Context, channelID string) ([]models.MessageWithAuthor, error)
	FindByPublicIDInChannel(ctx context.Context, channelID, publicID string) (models.MessageWithAuthor, error)
	Insert(ctx context.Context, channelID, authorID, body string, replyToID *string, createdAt time.Time) (models.Message, error)
	UpdateBody(ctx context.Context, channelID, publicID, authorID, body string, editedAt time.Time) (models.MessageWithAuthor, error)
	DeleteByID(ctx context.Context, id string) error
	DeleteByPublicID(ctx context.Context, channelID, publicID string) error
	ToggleReaction(ctx context.Context, messageID, userID, emoji string, now time.Time) ([]models.MessageReactionSummary, error)
	SetPinned(ctx context.Context, channelID, publicID, userID string, pinned bool, now time.Time) (models.MessageWithAuthor, error)
	FindUserDisplayName(ctx context.Context, userID string) (string, error)
	FindUserProfile(ctx context.Context, userID string) (displayName, avatarURL string, err error)
	FindLastMessageTimeByAuthor(ctx context.Context, channelID, authorID string) (*time.Time, error)
}

type pgRepository struct {
	db  *bun.DB
	log logger.Logger
}

// NewPgRepository constructs a Postgres message store.
func NewPgRepository(db *bun.DB, log logger.Logger) PgRepository {
	return pgRepository{db: db, log: log}
}

func (r pgRepository) baseMessageSelect(rows *[]models.MessageListRow) *bun.SelectQuery {
	return r.db.NewSelect().Model(rows).
		ColumnExpr("m.id, m.public_id, m.channel_id, m.author_id, m.reply_to_id, m.body, m.created_at, m.edited_at, m.is_pinned, m.pinned_at, m.pinned_by").
		ColumnExpr("u.display_name AS author_display_name, COALESCE(u.avatar_url, '') AS author_avatar_url").
		ColumnExpr("rm.public_id AS reply_public_id, rm.author_id AS reply_author_id, ru.display_name AS reply_author_name, LEFT(rm.body, 200) AS reply_body").
		Join("JOIN users AS u ON u.id = m.author_id").
		Join("LEFT JOIN messages AS rm ON rm.id = m.reply_to_id").
		Join("LEFT JOIN users AS ru ON ru.id = rm.author_id")
}

func (r pgRepository) loadReactionsByMessageIDs(ctx context.Context, messageIDs []string) (map[string][]models.MessageReactionSummary, error) {
	out := make(map[string][]models.MessageReactionSummary)
	if len(messageIDs) == 0 {
		return out, nil
	}

	type reactionAggRow struct {
		MessageID string   `bun:"message_id"`
		Emoji     string   `bun:"emoji"`
		Count     int32    `bun:"count"`
		UserIDs   []string `bun:"user_ids,array"`
	}

	var rows []reactionAggRow
	if err := r.db.NewSelect().
		TableExpr("message_reactions AS mr").
		ColumnExpr("mr.message_id").
		ColumnExpr("mr.emoji").
		ColumnExpr("COUNT(*)::int AS count").
		ColumnExpr("ARRAY_AGG(mr.user_id::text ORDER BY mr.created_at ASC) AS user_ids").
		Where("mr.message_id IN (?)", bun.List(messageIDs)).
		GroupExpr("mr.message_id, mr.emoji").
		OrderExpr("MIN(mr.created_at) ASC").
		Scan(ctx, &rows); err != nil {
		return nil, err
	}

	for _, row := range rows {
		out[row.MessageID] = append(out[row.MessageID], models.MessageReactionSummary{
			Emoji:   row.Emoji,
			Count:   row.Count,
			UserIDs: row.UserIDs,
		})
	}

	return out, nil
}

func (r pgRepository) hydrateRows(ctx context.Context, rows []models.MessageListRow) ([]models.MessageWithAuthor, error) {
	ids := make([]string, len(rows))
	for i, row := range rows {
		ids[i] = row.ID
	}
	reactionsMap, err := r.loadReactionsByMessageIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	out := make([]models.MessageWithAuthor, len(rows))
	for i, row := range rows {
		item := row.ToMessageWithAuthor()
		item.Reactions = reactionsMap[row.ID]
		out[i] = item
	}

	return out, nil
}

func (r pgRepository) ListByChannel(ctx context.Context, channelID string, limit int, before *time.Time) ([]models.MessageWithAuthor, error) {
	if limit <= 0 {
		limit = constants.DefaultMessageListLimit
	}
	if limit > constants.MaxMessageListLimit {
		limit = constants.MaxMessageListLimit
	}

	var rows []models.MessageListRow
	q := r.baseMessageSelect(&rows).
		Where("m.channel_id = ?", channelID).
		OrderExpr("m.created_at DESC, m.id DESC").
		Limit(int64(limit))
	if before != nil {
		q = q.Where("m.created_at < ?", *before)
	}

	if err := q.Scan(ctx); err != nil {
		r.logIfInfraErr("ListByChannel", err)

		return nil, err
	}

	hydrated, err := r.hydrateRows(ctx, rows)
	if err != nil {
		r.logIfInfraErr("ListByChannel.hydrate", err)

		return nil, err
	}

	out := make([]models.MessageWithAuthor, len(hydrated))
	for i, item := range hydrated {
		out[len(hydrated)-1-i] = item
	}

	return out, nil
}

func (r pgRepository) ListPinnedByChannel(ctx context.Context, channelID string) ([]models.MessageWithAuthor, error) {
	var rows []models.MessageListRow
	if err := r.baseMessageSelect(&rows).
		Where("m.channel_id = ? AND m.is_pinned = true", channelID).
		OrderExpr("m.pinned_at DESC, m.created_at DESC").
		Scan(ctx); err != nil {
		r.logIfInfraErr("ListPinnedByChannel", err)

		return nil, err
	}

	out, err := r.hydrateRows(ctx, rows)
	if err != nil {
		r.logIfInfraErr("ListPinnedByChannel.hydrate", err)

		return nil, err
	}

	return out, nil
}

func (r pgRepository) FindByPublicIDInChannel(ctx context.Context, channelID, publicID string) (models.MessageWithAuthor, error) {
	var rows []models.MessageListRow
	if err := r.baseMessageSelect(&rows).
		Where("m.channel_id = ? AND m.public_id = ?", channelID, publicID).
		Limit(1).
		Scan(ctx); err != nil {
		r.logIfInfraErr("FindByPublicIDInChannel", err)

		return models.MessageWithAuthor{}, err
	}
	if len(rows) == 0 {
		return models.MessageWithAuthor{}, constants.ErrNotFound
	}

	hydrated, err := r.hydrateRows(ctx, rows)
	if err != nil {
		r.logIfInfraErr("FindByPublicIDInChannel.hydrate", err)

		return models.MessageWithAuthor{}, err
	}

	return hydrated[0], nil
}

func (r pgRepository) Insert(ctx context.Context, channelID, authorID, body string, replyToID *string, createdAt time.Time) (models.Message, error) {
	for range 5 {
		id, err := token.NewID()
		if err != nil {
			return models.Message{}, err
		}
		publicID, err := token.NewPublicID()
		if err != nil {
			return models.Message{}, err
		}

		row := models.MessageRowFrom(models.Message{
			ID: id, PublicID: publicID, ChannelID: channelID, AuthorID: authorID,
			ReplyToID: replyToID, Body: body, CreatedAt: createdAt,
		})
		_, err = r.db.NewInsert().Model(&row).Exec(ctx)
		if err == nil {
			return row.ToModel(), nil
		}
		if errors.Is(pgErr.MapUnique(err), constants.ErrConflict) {
			continue
		}

		r.logIfInfraErr("Insert", err)

		return models.Message{}, err
	}

	return models.Message{}, constants.ErrConflict
}

func (r pgRepository) UpdateBody(ctx context.Context, channelID, publicID, authorID, body string, editedAt time.Time) (models.MessageWithAuthor, error) {
	res, err := r.db.NewUpdate().
		Model((*models.MessageRow)(nil)).
		Set("body = ?", body).
		Set("edited_at = ?", editedAt).
		Where("channel_id = ? AND public_id = ? AND author_id = ?", channelID, publicID, authorID).
		Exec(ctx)
	if err != nil {
		r.logIfInfraErr("UpdateBody", err)

		return models.MessageWithAuthor{}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return models.MessageWithAuthor{}, constants.ErrNotFound
	}

	return r.FindByPublicIDInChannel(ctx, channelID, publicID)
}

func (r pgRepository) DeleteByID(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().
		Model((*models.MessageRow)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		r.logIfInfraErr("DeleteByID", err)

		return err
	}

	return nil
}

func (r pgRepository) DeleteByPublicID(ctx context.Context, channelID, publicID string) error {
	res, err := r.db.NewDelete().
		Model((*models.MessageRow)(nil)).
		Where("channel_id = ? AND public_id = ?", channelID, publicID).
		Exec(ctx)
	if err != nil {
		r.logIfInfraErr("DeleteByPublicID", err)

		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return constants.ErrNotFound
	}

	return nil
}

func (r pgRepository) ToggleReaction(ctx context.Context, messageID, userID, emoji string, now time.Time) ([]models.MessageReactionSummary, error) {
	res, err := r.db.NewDelete().
		Model((*models.MessageReactionRow)(nil)).
		Where("message_id = ? AND user_id = ? AND emoji = ?", messageID, userID, emoji).
		Exec(ctx)
	if err != nil {
		r.logIfInfraErr("ToggleReaction.delete", err)

		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		id, err := token.NewID()
		if err != nil {
			return nil, err
		}
		row := models.MessageReactionRow{
			ID:        id,
			MessageID: messageID,
			UserID:    userID,
			Emoji:     emoji,
			CreatedAt: now,
		}
		if _, err := r.db.NewInsert().Model(&row).On("CONFLICT (message_id, user_id, emoji) DO NOTHING").Exec(ctx); err != nil {
			r.logIfInfraErr("ToggleReaction.insert", err)

			return nil, err
		}
	}

	m, err := r.loadReactionsByMessageIDs(ctx, []string{messageID})
	if err != nil {
		r.logIfInfraErr("ToggleReaction.load", err)

		return nil, err
	}

	return m[messageID], nil
}

func (r pgRepository) SetPinned(ctx context.Context, channelID, publicID, userID string, pinned bool, now time.Time) (models.MessageWithAuthor, error) {
	q := r.db.NewUpdate().
		Model((*models.MessageRow)(nil)).
		Set("is_pinned = ?", pinned).
		Where("channel_id = ? AND public_id = ?", channelID, publicID)
	if pinned {
		q = q.Set("pinned_at = ?", now).Set("pinned_by = ?", userID)
	} else {
		q = q.Set("pinned_at = NULL").Set("pinned_by = NULL")
	}

	res, err := q.Exec(ctx)
	if err != nil {
		r.logIfInfraErr("SetPinned", err)

		return models.MessageWithAuthor{}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return models.MessageWithAuthor{}, constants.ErrNotFound
	}

	return r.FindByPublicIDInChannel(ctx, channelID, publicID)
}

func (r pgRepository) FindUserDisplayName(ctx context.Context, userID string) (string, error) {
	name, _, err := r.FindUserProfile(ctx, userID)

	return name, err
}

func (r pgRepository) FindUserProfile(ctx context.Context, userID string) (displayName, avatarURL string, err error) {
	var row struct {
		DisplayName string `bun:"display_name"`
		AvatarURL   string `bun:"avatar_url"`
	}
	err = r.db.NewSelect().
		TableExpr("users").
		ColumnExpr("display_name, COALESCE(avatar_url, '') AS avatar_url").
		Where("id = ?", userID).
		Scan(ctx, &row)
	if err != nil {
		r.logIfInfraErr("FindUserProfile", err)

		return "", "", pgErr.MapNotFound(err)
	}

	return row.DisplayName, row.AvatarURL, nil
}

func (r pgRepository) FindLastMessageTimeByAuthor(ctx context.Context, channelID, authorID string) (*time.Time, error) {
	var createdAt time.Time
	err := r.db.NewSelect().
		TableExpr("messages").
		Column("created_at").
		Where("channel_id = ? AND author_id = ?", channelID, authorID).
		OrderExpr("created_at DESC").
		Limit(1).
		Scan(ctx, &createdAt)
	if err != nil {
		if errors.Is(pgErr.MapNotFound(err), constants.ErrNotFound) {
			return nil, nil
		}
		r.logIfInfraErr("FindLastMessageTimeByAuthor", err)

		return nil, err
	}

	return &createdAt, nil
}

func (r pgRepository) logIfInfraErr(op string, err error) {
	if err == nil {
		return
	}
	r.log.Error("message repository", "op", op, "err", err)
}
