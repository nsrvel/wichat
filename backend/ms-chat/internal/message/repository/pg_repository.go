package repository

import (
	"context"
	"errors"
	"time"

	"github.com/uptrace/bun"
	"github.com/wichat/wichat/backend/ms-chat/internal/models"
	outboxrepo "github.com/wichat/wichat/backend/ms-chat/internal/outbox/repository"
	"github.com/wichat/wichat/backend/ms-chat/pkg/constants"
	pgErr "github.com/wichat/wichat/backend/ms-chat/pkg/db"
	"github.com/wichat/wichat/backend/ms-chat/pkg/token"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

// PgRepository persists channel messages and reactions.
type PgRepository interface {
	ListByChannel(ctx context.Context, channelID string, limit int, before *time.Time, beforePublicID string, beforeChannelSeq, afterChannelSeq int64) ([]models.MessageWithAuthor, bool, error)
	ListPinnedByChannel(ctx context.Context, channelID string) ([]models.MessageWithAuthor, error)
	FindByPublicIDInChannel(ctx context.Context, channelID, publicID string) (models.MessageWithAuthor, error)
	FindByClientMsgIDInChannel(ctx context.Context, channelID, authorID, clientMsgID string) (models.MessageWithAuthor, error)
	Insert(ctx context.Context, channelID, authorID, body string, replyToID *string, createdAt time.Time) (models.Message, error)
	InsertWithOutbox(ctx context.Context, in *models.InsertMessageWithOutboxInput) (models.Message, error)
	UpdateBody(ctx context.Context, channelID, publicID, authorID, body string, editedAt time.Time) (models.MessageWithAuthor, error)
	UpdateBodyWithOutbox(ctx context.Context, in *models.UpdateMessageBodyWithOutboxInput) (models.MessageWithAuthor, error)
	DeleteByID(ctx context.Context, id string) error
	DeleteByPublicID(ctx context.Context, channelID, publicID string) error
	DeleteByPublicIDWithOutbox(ctx context.Context, in *models.DeleteMessageWithOutboxInput) error
	ToggleReaction(ctx context.Context, messageID, userID, emoji string, now time.Time) ([]models.MessageReactionSummary, error)
	ToggleReactionWithOutbox(ctx context.Context, in *models.ToggleReactionWithOutboxInput) ([]models.MessageReactionSummary, error)
	SetPinned(ctx context.Context, channelID, publicID, userID string, pinned bool, now time.Time) (models.MessageWithAuthor, error)
	SetPinnedWithOutbox(ctx context.Context, in *models.SetPinnedWithOutboxInput) (models.MessageWithAuthor, error)
	FindLastMessageTimeByAuthor(ctx context.Context, channelID, authorID string) (*time.Time, error)
}

type pgRepository struct {
	db     *bun.DB
	outbox outboxrepo.PgRepository
	log    logger.Logger
}

// NewPgRepository constructs a Postgres message store.
func NewPgRepository(db *bun.DB, outbox outboxrepo.PgRepository, log logger.Logger) PgRepository {
	return pgRepository{db: db, outbox: outbox, log: log}
}

func (r pgRepository) baseMessageSelect(rows *[]models.MessageListRow) *bun.SelectQuery {
	return r.db.NewSelect().Model(rows).
		ColumnExpr("m.id, m.public_id, m.channel_id, m.author_id, m.reply_to_id, m.channel_seq, m.body, m.created_at, m.edited_at, m.is_pinned, m.pinned_at, m.pinned_by").
		ColumnExpr("u.display_name AS author_display_name, COALESCE(u.avatar_url, '') AS author_avatar_url").
		ColumnExpr("rm.public_id AS reply_public_id, rm.author_id AS reply_author_id, ru.display_name AS reply_author_name, rm.body AS reply_body").
		Join("JOIN users AS u ON u.id = m.author_id").
		Join("LEFT JOIN messages AS rm ON rm.id = m.reply_to_id").
		Join("LEFT JOIN users AS ru ON ru.id = rm.author_id")
}

func (r pgRepository) loadReactionsByMessageIDs(ctx context.Context, messageIDs []string) (map[string][]models.MessageReactionSummary, error) {
	out := make(map[string][]models.MessageReactionSummary)
	if len(messageIDs) == 0 {
		return out, nil
	}

	var rows []models.MessageReactionRow
	if err := r.db.NewSelect().Model(&rows).
		Where("mr.message_id IN (?)", bun.List(messageIDs)).
		OrderExpr("mr.created_at ASC").
		Scan(ctx); err != nil {
		return nil, err
	}

	byMsgEmoji := make(map[string]map[string]*models.MessageReactionSummary)
	order := make(map[string][]string)
	for _, row := range rows {
		m, ok := byMsgEmoji[row.MessageID]
		if !ok {
			m = make(map[string]*models.MessageReactionSummary)
			byMsgEmoji[row.MessageID] = m
		}
		summary, exists := m[row.Emoji]
		if !exists {
			summary = &models.MessageReactionSummary{Emoji: row.Emoji}
			m[row.Emoji] = summary
			order[row.MessageID] = append(order[row.MessageID], row.Emoji)
		}
		summary.Count++
		summary.UserIDs = append(summary.UserIDs, row.UserID)
	}

	for msgID, emojis := range order {
		list := make([]models.MessageReactionSummary, 0, len(emojis))
		for _, emoji := range emojis {
			list = append(list, *byMsgEmoji[msgID][emoji])
		}
		out[msgID] = list
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

func (r pgRepository) ListByChannel(
	ctx context.Context,
	channelID string,
	limit int,
	before *time.Time,
	beforePublicID string,
	beforeChannelSeq, afterChannelSeq int64,
) ([]models.MessageWithAuthor, bool, error) {
	if afterChannelSeq > 0 {
		return r.listAfterChannelSeq(ctx, channelID, limit, afterChannelSeq)
	}

	var rows []models.MessageListRow
	q := r.baseMessageSelect(&rows).
		Where("m.channel_id = ?", channelID).
		OrderExpr("m.channel_seq DESC, m.id DESC").
		Limit(int64(limit + 1))
	if beforeChannelSeq > 0 {
		q = q.Where("m.channel_seq < ?", beforeChannelSeq)
	} else if before != nil {
		if beforePublicID != "" {
			q = q.Where(
				"(m.created_at < ? OR (m.created_at = ? AND m.id < (SELECT id FROM messages WHERE channel_id = ? AND public_id = ? LIMIT 1)))",
				*before, *before, channelID, beforePublicID,
			)
		} else {
			q = q.Where("m.created_at < ?", *before)
		}
	}

	if err := q.Scan(ctx); err != nil {
		r.logIfInfraErr("ListByChannel", err)

		return nil, false, err
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	hydrated, err := r.hydrateRows(ctx, rows)
	if err != nil {
		r.logIfInfraErr("ListByChannel.hydrate", err)

		return nil, false, err
	}

	out := make([]models.MessageWithAuthor, len(hydrated))
	for i, item := range hydrated {
		out[len(hydrated)-1-i] = item
	}

	return out, hasMore, nil
}

func (r pgRepository) listAfterChannelSeq(ctx context.Context, channelID string, limit int, afterChannelSeq int64) ([]models.MessageWithAuthor, bool, error) {
	var rows []models.MessageListRow
	err := r.baseMessageSelect(&rows).
		Where("m.channel_id = ? AND m.channel_seq > ?", channelID, afterChannelSeq).
		OrderExpr("m.channel_seq ASC, m.id ASC").
		Limit(int64(limit + 1)).
		Scan(ctx)
	if err != nil {
		r.logIfInfraErr("listAfterChannelSeq", err)

		return nil, false, err
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	hydrated, err := r.hydrateRows(ctx, rows)
	if err != nil {
		return nil, false, err
	}

	return hydrated, hasMore, nil
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

func (r pgRepository) FindByClientMsgIDInChannel(ctx context.Context, channelID, authorID, clientMsgID string) (models.MessageWithAuthor, error) {
	var rows []models.MessageListRow
	err := r.baseMessageSelect(&rows).
		Where("m.channel_id = ? AND m.author_id = ? AND m.client_msg_id = ?", channelID, authorID, clientMsgID).
		Limit(1).
		Scan(ctx)
	if err != nil {
		r.logIfInfraErr("FindByClientMsgIDInChannel", err)

		return models.MessageWithAuthor{}, err
	}
	if len(rows) == 0 {
		return models.MessageWithAuthor{}, constants.ErrNotFound
	}

	hydrated, err := r.hydrateRows(ctx, rows)
	if err != nil {
		r.logIfInfraErr("FindByClientMsgIDInChannel.hydrate", err)

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

func (r pgRepository) UpdateBodyWithOutbox(ctx context.Context, in *models.UpdateMessageBodyWithOutboxInput) (models.MessageWithAuthor, error) {
	if in == nil {
		return models.MessageWithAuthor{}, errors.New("update message input required")
	}

	err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		res, err := tx.NewUpdate().
			Model((*models.MessageRow)(nil)).
			Set("body = ?", in.Body).
			Set("edited_at = ?", in.EditedAt).
			Where("channel_id = ? AND public_id = ? AND author_id = ?", in.ChannelID, in.PublicID, in.AuthorID).
			Exec(ctx)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return constants.ErrNotFound
		}

		return r.outbox.InsertInTx(ctx, tx, in.Outbox)
	})
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return models.MessageWithAuthor{}, constants.ErrNotFound
		}
		r.logIfInfraErr("UpdateBodyWithOutbox", err)

		return models.MessageWithAuthor{}, err
	}

	return r.FindByPublicIDInChannel(ctx, in.ChannelID, in.PublicID)
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

func (r pgRepository) DeleteByPublicIDWithOutbox(ctx context.Context, in *models.DeleteMessageWithOutboxInput) error {
	if in == nil {
		return errors.New("delete message input required")
	}

	err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		res, err := tx.NewDelete().
			Model((*models.MessageRow)(nil)).
			Where("channel_id = ? AND public_id = ?", in.ChannelID, in.PublicID).
			Exec(ctx)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return constants.ErrNotFound
		}

		return r.outbox.InsertInTx(ctx, tx, in.Outbox)
	})
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return constants.ErrNotFound
		}
		r.logIfInfraErr("DeleteByPublicIDWithOutbox", err)

		return err
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

func (r pgRepository) ToggleReactionWithOutbox(ctx context.Context, in *models.ToggleReactionWithOutboxInput) ([]models.MessageReactionSummary, error) {
	if in == nil {
		return nil, errors.New("toggle reaction input required")
	}

	var reactions []models.MessageReactionSummary
	err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		var toggleErr error
		reactions, toggleErr = r.toggleReactionInTx(ctx, tx, in.MessageID, in.UserID, in.Emoji, in.Now)
		if toggleErr != nil {
			return toggleErr
		}

		return r.outbox.InsertInTx(ctx, tx, in.Outbox)
	})
	if err != nil {
		r.logIfInfraErr("ToggleReactionWithOutbox", err)

		return nil, err
	}

	return reactions, nil
}

func (r pgRepository) toggleReactionInTx(ctx context.Context, tx bun.Tx, messageID, userID, emoji string, now time.Time) ([]models.MessageReactionSummary, error) {
	res, err := tx.NewDelete().
		Model((*models.MessageReactionRow)(nil)).
		Where("message_id = ? AND user_id = ? AND emoji = ?", messageID, userID, emoji).
		Exec(ctx)
	if err != nil {
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
		if _, err := tx.NewInsert().Model(&row).On("CONFLICT (message_id, user_id, emoji) DO NOTHING").Exec(ctx); err != nil {
			return nil, err
		}
	}

	m, err := r.loadReactionsByMessageIDsInTx(ctx, tx, []string{messageID})
	if err != nil {
		return nil, err
	}

	return m[messageID], nil
}

func (r pgRepository) loadReactionsByMessageIDsInTx(ctx context.Context, tx bun.Tx, messageIDs []string) (map[string][]models.MessageReactionSummary, error) {
	out := make(map[string][]models.MessageReactionSummary)
	if len(messageIDs) == 0 {
		return out, nil
	}

	var rows []models.MessageReactionRow
	if err := tx.NewSelect().Model(&rows).
		Where("mr.message_id IN (?)", bun.List(messageIDs)).
		OrderExpr("mr.created_at ASC").
		Scan(ctx); err != nil {
		return nil, err
	}

	byMsgEmoji := make(map[string]map[string]*models.MessageReactionSummary)
	order := make(map[string][]string)
	for _, row := range rows {
		m, ok := byMsgEmoji[row.MessageID]
		if !ok {
			m = make(map[string]*models.MessageReactionSummary)
			byMsgEmoji[row.MessageID] = m
		}
		summary, exists := m[row.Emoji]
		if !exists {
			summary = &models.MessageReactionSummary{Emoji: row.Emoji}
			m[row.Emoji] = summary
			order[row.MessageID] = append(order[row.MessageID], row.Emoji)
		}
		summary.Count++
		summary.UserIDs = append(summary.UserIDs, row.UserID)
	}

	for msgID, emojis := range order {
		list := make([]models.MessageReactionSummary, 0, len(emojis))
		for _, emoji := range emojis {
			list = append(list, *byMsgEmoji[msgID][emoji])
		}
		out[msgID] = list
	}

	return out, nil
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

func (r pgRepository) SetPinnedWithOutbox(ctx context.Context, in *models.SetPinnedWithOutboxInput) (models.MessageWithAuthor, error) {
	if in == nil {
		return models.MessageWithAuthor{}, errors.New("set pinned input required")
	}

	err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		q := tx.NewUpdate().
			Model((*models.MessageRow)(nil)).
			Set("is_pinned = ?", in.Pinned).
			Where("channel_id = ? AND public_id = ?", in.ChannelID, in.PublicID)
		if in.Pinned {
			q = q.Set("pinned_at = ?", in.Now).Set("pinned_by = ?", in.UserID)
		} else {
			q = q.Set("pinned_at = NULL").Set("pinned_by = NULL")
		}

		res, err := q.Exec(ctx)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return constants.ErrNotFound
		}

		return r.outbox.InsertInTx(ctx, tx, in.Outbox)
	})
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return models.MessageWithAuthor{}, constants.ErrNotFound
		}
		r.logIfInfraErr("SetPinnedWithOutbox", err)

		return models.MessageWithAuthor{}, err
	}

	return r.FindByPublicIDInChannel(ctx, in.ChannelID, in.PublicID)
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

func (r pgRepository) InsertWithOutbox(ctx context.Context, in *models.InsertMessageWithOutboxInput) (models.Message, error) {
	if in == nil {
		return models.Message{}, errors.New("insert message input required")
	}

	if in.ID == "" || in.PublicID == "" {
		return models.Message{}, errors.New("message id and public_id required")
	}

	var inserted models.Message
	err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		channelSeq, err := r.nextChannelSeq(ctx, tx, in.ChannelID)
		if err != nil {
			return err
		}

		row := models.MessageRowFrom(models.Message{
			ID: in.ID, PublicID: in.PublicID, ChannelID: in.ChannelID, AuthorID: in.AuthorID,
			ClientMsgID: in.ClientMsgID, ReplyToID: in.ReplyToID, ChannelSeq: channelSeq, Body: in.Body, CreatedAt: in.CreatedAt,
		})
		if _, err := tx.NewInsert().Model(&row).Exec(ctx); err != nil {
			return err
		}
		if in.BuildOutbox == nil {
			return errors.New("build outbox required")
		}
		outbox, err := in.BuildOutbox(channelSeq)
		if err != nil {
			return err
		}
		if err := r.outbox.InsertInTx(ctx, tx, outbox); err != nil {
			return err
		}

		inserted = row.ToModel()

		return nil
	})
	if err == nil {
		return inserted, nil
	}
	if errors.Is(pgErr.MapUnique(err), constants.ErrConflict) {
		return models.Message{}, constants.ErrConflict
	}
	r.logIfInfraErr("InsertWithOutbox", err)

	return models.Message{}, err
}

func (r pgRepository) nextChannelSeq(ctx context.Context, tx bun.Tx, channelID string) (int64, error) {
	var channelSeq int64
	err := tx.NewRaw(
		`UPDATE channels SET message_seq = message_seq + 1 WHERE id = ? RETURNING message_seq`,
		channelID,
	).Scan(ctx, &channelSeq)
	if err != nil {
		r.logIfInfraErr("nextChannelSeq", err)

		return 0, err
	}

	return channelSeq, nil
}

func (r pgRepository) logIfInfraErr(op string, err error) {
	if err == nil {
		return
	}
	r.log.Error("message repository", "op", op, "err", err)
}
