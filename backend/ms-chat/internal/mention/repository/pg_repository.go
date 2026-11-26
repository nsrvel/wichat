package repository

import (
	"context"
	"strings"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/wichat/wichat/backend/ms-chat/internal/models"
	"github.com/wichat/wichat/backend/ms-chat/pkg/constants"
	"github.com/wichat/wichat/backend/ms-chat/pkg/mention"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

// PgRepository persists and lists message mention inbox rows.
type PgRepository interface {
	InsertMentions(ctx context.Context, tx bun.Tx, rows []models.MessageMentionRow) error
	ResolveUserPublicIDsForChannel(ctx context.Context, channelID string, publicIDs []string) ([]string, error)
	LookupGroupInternalID(ctx context.Context, groupPublicID string) (string, error)
	LookupUserInternalID(ctx context.Context, userID string) (string, error)
	ListForMentionedUser(
		ctx context.Context,
		groupID, mentionedUserID string,
		limit int,
		cursor time.Time,
		channelKind string,
		unreadOnly bool,
	) ([]models.GroupMentionHit, error)
	MarkRead(ctx context.Context, groupID, mentionedUserID string, messagePublicIDs []string, at time.Time) error
	CountUnread(ctx context.Context, groupID, mentionedUserID string) (int, error)
	MentionedPublicUserIDsByMessageIDs(ctx context.Context, messageInternalIDs []string) (map[string][]string, error)
}

type pgRepository struct {
	db  *bun.DB
	log logger.Logger
}

// NewPgRepository constructs a Postgres mention store.
func NewPgRepository(db *bun.DB, log logger.Logger) PgRepository {
	return pgRepository{db: db, log: log}
}

func (r pgRepository) InsertMentions(ctx context.Context, tx bun.Tx, rows []models.MessageMentionRow) error {
	if len(rows) == 0 {
		return nil
	}
	_, err := tx.NewInsert().Model(&rows).Exec(ctx)
	if err != nil {
		r.logIfInfraErr("InsertMentions", err)
	}
	return err
}

func (r pgRepository) ResolveUserPublicIDsForChannel(
	ctx context.Context,
	channelID string,
	publicIDs []string,
) ([]string, error) {
	if len(publicIDs) == 0 {
		return nil, nil
	}
	var ids []string
	err := r.db.NewSelect().
		TableExpr("users AS u").
		ColumnExpr("u.id").
		Join("JOIN channels AS c ON c.id = ?", channelID).
		Join("JOIN group_members AS gm ON gm.user_id = u.id AND gm.group_id = c.group_id").
		Where("u.id = ANY(?)", pgdialect.Array(publicIDs)).
		Scan(ctx, &ids)
	if err != nil {
		r.logIfInfraErr("ResolveUserPublicIDsForChannel", err)
		return nil, err
	}
	return ids, nil
}

func (r pgRepository) LookupGroupInternalID(ctx context.Context, groupPublicID string) (string, error) {
	var id string
	err := r.db.NewSelect().
		TableExpr("groups").
		ColumnExpr("id").
		Where("public_id = ?", strings.TrimSpace(groupPublicID)).
		Scan(ctx, &id)
	if err != nil {
		r.logIfInfraErr("LookupGroupInternalID", err)
		return "", err
	}
	return id, nil
}

func (r pgRepository) LookupUserInternalID(ctx context.Context, userID string) (string, error) {
	var id string
	err := r.db.NewSelect().
		TableExpr("users").
		ColumnExpr("id").
		Where("id = ?", strings.TrimSpace(userID)).
		Scan(ctx, &id)
	if err != nil {
		r.logIfInfraErr("LookupUserInternalID", err)
		return "", err
	}
	return id, nil
}

func (r pgRepository) ListForMentionedUser(
	ctx context.Context,
	groupID, mentionedUserID string,
	limit int,
	cursor time.Time,
	channelKind string,
	unreadOnly bool,
) ([]models.GroupMentionHit, error) {
	if limit <= 0 {
		limit = constants.DefaultMentionListLimit
	}
	if limit > 150 {
		limit = 150
	}

	type row struct {
		MessagePublicID string     `bun:"message_public_id"`
		ChannelPublicID string     `bun:"channel_public_id"`
		ChannelName     string     `bun:"channel_name"`
		ChannelType     string     `bun:"channel_type"`
		AuthorName      string     `bun:"author_name"`
		Body            string     `bun:"body"`
		CreatedAt       time.Time  `bun:"created_at"`
		ReadAt          *time.Time `bun:"read_at"`
	}

	var rows []row
	q := r.db.NewSelect().
		TableExpr("message_mentions AS mm").
		ColumnExpr("m.public_id AS message_public_id").
		ColumnExpr("c.public_id AS channel_public_id").
		ColumnExpr("c.name AS channel_name").
		ColumnExpr("c.type AS channel_type").
		ColumnExpr("u.display_name AS author_name").
		ColumnExpr("m.body AS body").
		ColumnExpr("mm.created_at AS created_at").
		ColumnExpr("mm.read_at AS read_at").
		Join("JOIN messages AS m ON m.id = mm.message_id").
		Join("JOIN channels AS c ON c.id = mm.channel_id").
		Join("JOIN users AS u ON u.id = m.author_id").
		Where("mm.group_id = ?", groupID).
		Where("mm.mentioned_user_id = ?", mentionedUserID).
		OrderExpr("mm.created_at DESC").
		Limit(int64(limit))

	if !cursor.IsZero() {
		q = q.Where("mm.created_at < ?", cursor)
	}
	if unreadOnly {
		q = q.Where("mm.read_at IS NULL")
	}
	kind := strings.TrimSpace(strings.ToLower(channelKind))
	switch kind {
	case "channels":
		q = q.Where("c.type <> ?", "dm")
	case "chats":
		q = q.Where("c.type = ?", "dm")
	}

	if err := q.Scan(ctx, &rows); err != nil {
		r.logIfInfraErr("ListForMentionedUser", err)
		return nil, err
	}

	out := make([]models.GroupMentionHit, len(rows))
	for i, row := range rows {
		body := mention.StripTokensForPreview(row.Body)
		out[i] = models.GroupMentionHit{
			MessageID:         row.MessagePublicID,
			ChannelID:         row.ChannelPublicID,
			ChannelName:       row.ChannelName,
			ChannelType:       row.ChannelType,
			AuthorDisplayName: row.AuthorName,
			BodyPreview:       previewBody(body, 160),
			CreatedAt:         row.CreatedAt,
			IsUnread:          row.ReadAt == nil,
		}
	}
	return out, nil
}

func (r pgRepository) MarkRead(
	ctx context.Context,
	groupID, mentionedUserID string,
	messagePublicIDs []string,
	at time.Time,
) error {
	if len(messagePublicIDs) == 0 {
		return nil
	}
	var messageIDs []string
	err := r.db.NewSelect().
		TableExpr("messages").
		ColumnExpr("id").
		Where("public_id = ANY(?)", pgdialect.Array(messagePublicIDs)).
		Scan(ctx, &messageIDs)
	if err != nil {
		r.logIfInfraErr("MarkRead.lookup", err)
		return err
	}
	if len(messageIDs) == 0 {
		return nil
	}
	_, err = r.db.NewUpdate().
		TableExpr("message_mentions").
		Set("read_at = ?", at.UTC()).
		Where("group_id = ?", groupID).
		Where("mentioned_user_id = ?", mentionedUserID).
		Where("message_id = ANY(?)", pgdialect.Array(messageIDs)).
		Where("read_at IS NULL").
		Exec(ctx)
	if err != nil {
		r.logIfInfraErr("MarkRead", err)
	}
	return err
}

func (r pgRepository) CountUnread(ctx context.Context, groupID, mentionedUserID string) (int, error) {
	count, err := r.db.NewSelect().
		TableExpr("message_mentions").
		ColumnExpr("count(*)").
		Where("group_id = ?", groupID).
		Where("mentioned_user_id = ?", mentionedUserID).
		Where("read_at IS NULL").
		Count(ctx)
	if err != nil {
		r.logIfInfraErr("CountUnread", err)
		return 0, err
	}
	return int(count), nil
}

func (r pgRepository) MentionedPublicUserIDsByMessageIDs(
	ctx context.Context,
	messageInternalIDs []string,
) (map[string][]string, error) {
	if len(messageInternalIDs) == 0 {
		return map[string][]string{}, nil
	}
	type row struct {
		MessageID string `bun:"message_id"`
		PublicID  string `bun:"public_id"`
	}
	var rows []row
	err := r.db.NewSelect().
		TableExpr("message_mentions AS mm").
		ColumnExpr("mm.message_id").
		ColumnExpr("u.id AS public_id").
		Join("JOIN users AS u ON u.id = mm.mentioned_user_id").
		Where("mm.message_id = ANY(?)", pgdialect.Array(messageInternalIDs)).
		OrderExpr("mm.created_at ASC").
		Scan(ctx, &rows)
	if err != nil {
		r.logIfInfraErr("MentionedPublicUserIDsByMessageIDs", err)
		return nil, err
	}
	out := make(map[string][]string)
	seen := make(map[string]map[string]struct{})
	for _, row := range rows {
		if row.MessageID == "" || row.PublicID == "" {
			continue
		}
		if seen[row.MessageID] == nil {
			seen[row.MessageID] = make(map[string]struct{})
		}
		if _, ok := seen[row.MessageID][row.PublicID]; ok {
			continue
		}
		seen[row.MessageID][row.PublicID] = struct{}{}
		out[row.MessageID] = append(out[row.MessageID], row.PublicID)
	}
	return out, nil
}

func previewBody(body string, maxRunes int) string {
	runes := []rune(strings.TrimSpace(body))
	if len(runes) <= maxRunes {
		return string(runes)
	}
	return string(runes[:maxRunes]) + "…"
}

func (r pgRepository) logIfInfraErr(op string, err error) {
	if err != nil {
		r.log.Error("mention repository "+op, "err", err)
	}
}
