package service

import (
	"bytes"
	"context"
	"errors"
	"math"
	"net/http"
	"path"
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
	messagerepo "github.com/wichat/wichat/backend/ms-chat/internal/message/repository"
	"github.com/wichat/wichat/backend/ms-chat/internal/models"
	outboxrepo "github.com/wichat/wichat/backend/ms-chat/internal/outbox/repository"
	pkgChatEvents "github.com/wichat/wichat/backend/ms-chat/pkg/chat-events"
	"github.com/wichat/wichat/backend/ms-chat/pkg/constants"
	"github.com/wichat/wichat/backend/ms-chat/pkg/token"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	userauth "github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user/grpc/auth"
	userchannel "github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user/grpc/channel"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

// Service is channel message use cases.
type Service interface {
	ListChannelMessages(ctx context.Context, in *models.ListChannelMessagesInput) ([]models.PublicMessage, bool, error)
	CreateChannelMessage(ctx context.Context, in *models.CreateChannelMessageInput) (models.PublicMessage, error)
	UpdateChannelMessage(ctx context.Context, in *models.UpdateChannelMessageInput) (models.PublicMessage, error)
	DeleteChannelMessage(ctx context.Context, in *models.DeleteChannelMessageInput) error
	ToggleMessageReaction(ctx context.Context, in *models.ToggleMessageReactionInput) ([]models.MessageReactionSummary, error)
	SetChannelMessagePinned(ctx context.Context, in *models.SetChannelMessagePinnedInput) (models.PublicMessage, error)
	ListPinnedChannelMessages(ctx context.Context, in *models.ListPinnedChannelMessagesInput) ([]models.PublicMessage, error)
	UploadChannelAttachment(ctx context.Context, in *models.UploadChannelAttachmentInput) (models.MessageAttachment, error)
}

type service struct {
	messages    messagerepo.PgRepository
	attachments messagerepo.MinioRepository
	userAuth    userauth.Client
	userChannel userchannel.Client
	outbox      outboxrepo.PgRepository
	log         logger.Logger
	now         func() time.Time
}

// NewService builds message use cases.
func NewService(
	messages messagerepo.PgRepository,
	attachments messagerepo.MinioRepository,
	userAuth userauth.Client,
	userChannel userchannel.Client,
	outbox outboxrepo.PgRepository,
	log logger.Logger,
) Service {
	return service{
		messages:    messages,
		attachments: attachments,
		userAuth:    userAuth,
		userChannel: userChannel,
		outbox:      outbox,
		log:         log,
		now:         func() time.Time { return time.Now().UTC() },
	}
}

func hasAnyPerm(perms []string, targets ...string) bool {
	for _, t := range targets {
		if slices.Contains(perms, t) {
			return true
		}
	}

	return false
}

func validateMessageBody(raw string) (string, error) {
	body := strings.TrimSpace(raw)
	if body == "" || utf8.RuneCountInString(body) > constants.MaxMessageBodyRunes || hasDangerousControlChars(body) {
		return "", exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "body"})
	}

	return body, nil
}

func (s service) ListChannelMessages(ctx context.Context, in *models.ListChannelMessagesInput) ([]models.PublicMessage, bool, error) {
	if err := s.applyChannelAccess(ctx, in.AccessToken, in.GroupPublicID, in.ChannelPublicID, channelAccessInput{
		callerUserID:      &in.CallerUserID,
		resolvedChannelID: &in.ResolvedChannelID,
		callerPermissions: &in.CallerPermissions,
	}); err != nil {
		return nil, false, err
	}

	if !hasAnyPerm(in.CallerPermissions, "administrator", "read_message_history") {
		return []models.PublicMessage{}, false, nil
	}

	limit := in.Limit
	if in.AfterChannelSeq > 0 {
		if limit <= 0 {
			limit = constants.DefaultMessageListLimit
		}
		if limit > constants.MaxGapFillListLimit {
			limit = constants.MaxGapFillListLimit
		}
	} else {
		if limit <= 0 {
			limit = constants.DefaultMessageListLimit
		}
		if limit > constants.MaxMessageListLimit {
			limit = constants.MaxMessageListLimit
		}
	}

	// Repository
	beforeSeq := in.BeforeChannelSeq
	if in.AfterChannelSeq > 0 {
		beforeSeq = 0
	}
	rows, hasMore, err := s.messages.ListByChannel(ctx, in.ResolvedChannelID, limit, beforeSeq, in.AfterChannelSeq)
	if err != nil {
		return nil, false, err
	}

	out := make([]models.PublicMessage, len(rows))
	for i, row := range rows {
		out[i] = toPublicMessage(row, in.CallerUserID)
	}

	return out, hasMore, nil
}

func (s service) CreateChannelMessage(ctx context.Context, in *models.CreateChannelMessageInput) (models.PublicMessage, error) {
	if err := s.applyChannelAccess(ctx, in.AccessToken, in.GroupPublicID, in.ChannelPublicID, channelAccessInput{
		callerUserID:      &in.CallerUserID,
		resolvedChannelID: &in.ResolvedChannelID,
		callerPermissions: &in.CallerPermissions,
		channelType:       &in.ChannelType,
		slowmodeSeconds:   &in.SlowmodeSeconds,
	}); err != nil {
		return models.PublicMessage{}, err
	}

	// Permissions & channel policy
	if len(in.CallerPermissions) > 0 {
		if !hasAnyPerm(in.CallerPermissions, "administrator", "send_messages") {
			return models.PublicMessage{}, exception.PermissionDenied(exception.MsgForbidden, nil)
		}
		if in.ChannelType == "broadcast" && !hasAnyPerm(in.CallerPermissions, "administrator", "manage_channels", "manage_messages") {
			return models.PublicMessage{}, exception.PermissionDenied(exception.MsgForbidden, nil)
		}
	}

	// Validation
	body, err := validateMessageBody(in.Body)
	if err != nil {
		return models.PublicMessage{}, err
	}

	var clientMsgIDPtr *string
	clientMsgID := strings.TrimSpace(in.ClientMsgID)
	if clientMsgID != "" {
		if _, err := uuid.Parse(clientMsgID); err != nil {
			return models.PublicMessage{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "client_msg_id"})
		}
		clientMsgIDPtr = &clientMsgID
		existing, err := s.messages.FindByClientMsgIDInChannel(ctx, in.ResolvedChannelID, in.CallerUserID, clientMsgID)
		if err != nil && !errors.Is(err, constants.ErrNotFound) {
			return models.PublicMessage{}, err
		}
		if err == nil {
			out := toPublicMessage(existing, in.CallerUserID)
			displayName, avatarURL, pErr := s.callerProfile(ctx, in.AccessToken, in.CallerUserID)
			if pErr == nil {
				out.AuthorDisplayName = displayName
				out.AuthorAvatarURL = avatarURL
			}

			return out, nil
		}
	}

	// Slowmode enforcement
	if in.SlowmodeSeconds > 0 && !hasAnyPerm(in.CallerPermissions, "administrator", "bypass_slowmode", "manage_channels", "manage_messages") {
		lastAt, err := s.messages.FindLastMessageTimeByAuthor(ctx, in.ResolvedChannelID, in.CallerUserID)
		if err != nil {
			return models.PublicMessage{}, err
		}
		if lastAt != nil {
			cooldown := time.Duration(in.SlowmodeSeconds) * time.Second
			elapsed := s.now().Sub(*lastAt)
			if elapsed < cooldown {
				remaining := int((cooldown - elapsed + time.Second - 1) / time.Second)
				if remaining < 1 {
					remaining = 1
				}

				return models.PublicMessage{}, exception.RateLimited(exception.MsgRateLimited, map[string]any{
					"retry_after_seconds": remaining,
				})
			}
		}
	}

	// Optional reply target resolution within same channel
	var replyToID *string
	var replyPreview *models.MessageReplyPreview
	if replyPubID := strings.TrimSpace(in.ReplyToPublicID); replyPubID != "" {
		parent, err := s.messages.FindByPublicIDInChannel(ctx, in.ResolvedChannelID, replyPubID)
		if err != nil {
			if errors.Is(err, constants.ErrNotFound) {
				return models.PublicMessage{}, exception.NotFound(exception.MsgInvalidInput, map[string]any{"field": "reply_to_id"})
			}

			return models.PublicMessage{}, err
		}
		replyToID = &parent.ID
		previewBody := parent.Body
		if utf8.RuneCountInString(previewBody) > 120 {
			previewBody = string([]rune(previewBody)[:120]) + "…"
		}
		replyPreview = &models.MessageReplyPreview{
			ID:                parent.PublicID,
			AuthorID:          parent.AuthorID,
			AuthorDisplayName: parent.AuthorDisplayName,
			BodyPreview:       previewBody,
		}
	}

	displayName, avatarURL, err := s.callerProfile(ctx, in.AccessToken, in.CallerUserID)
	if err != nil {
		return models.PublicMessage{}, err
	}

	createdAt := s.now()
	preview := models.PublicMessage{
		AuthorID: in.CallerUserID, AuthorDisplayName: displayName, AuthorAvatarURL: avatarURL,
		ClientMsgID: clientMsgID, Body: body, CreatedAt: createdAt, ReplyTo: replyPreview, Mine: true,
	}

	var msg models.Message
	for range 5 {
		id, err := token.NewID()
		if err != nil {
			return models.PublicMessage{}, err
		}
		publicID, err := token.NewPublicID()
		if err != nil {
			return models.PublicMessage{}, err
		}
		preview.ID = publicID
		wire := toPublicMessageWire(preview)

		msg, err = s.messages.InsertWithOutbox(ctx, &models.InsertMessageWithOutboxInput{
			ID: id, PublicID: publicID,
			ChannelID:   in.ResolvedChannelID,
			AuthorID:    in.CallerUserID,
			ClientMsgID: clientMsgIDPtr,
			Body:        body,
			ReplyToID:   replyToID,
			CreatedAt:   createdAt,
			BuildOutbox: func(channelSeq int64) (models.EventOutboxInsert, error) {
				wire.ChannelSeq = channelSeq
				outboxRow, err := pkgChatEvents.EncodeMessageCreated(
					in.GroupPublicID,
					in.ChannelPublicID,
					wire,
					createdAt,
				)
				if err != nil {
					return models.EventOutboxInsert{}, err
				}

				return toEventOutboxInsert(outboxRow), nil
			},
		})
		if err == nil {
			break
		}
		if errors.Is(err, constants.ErrConflict) && clientMsgID != "" {
			existing, fErr := s.messages.FindByClientMsgIDInChannel(ctx, in.ResolvedChannelID, in.CallerUserID, clientMsgID)
			if fErr == nil {
				out := toPublicMessage(existing, in.CallerUserID)
				out.AuthorDisplayName = displayName
				out.AuthorAvatarURL = avatarURL

				return out, nil
			}
		}
		if !errors.Is(err, constants.ErrConflict) {
			return models.PublicMessage{}, err
		}
	}
	if msg.PublicID == "" {
		return models.PublicMessage{}, constants.ErrConflict
	}

	out := models.PublicMessage{
		ID: msg.PublicID, ChannelSeq: msg.ChannelSeq, AuthorID: msg.AuthorID,
		AuthorDisplayName: displayName, AuthorAvatarURL: avatarURL,
		Body: msg.Body, CreatedAt: msg.CreatedAt, ReplyTo: replyPreview, Mine: true,
	}
	if msg.ClientMsgID != nil {
		out.ClientMsgID = *msg.ClientMsgID
	}

	return out, nil
}

func (s service) UpdateChannelMessage(ctx context.Context, in *models.UpdateChannelMessageInput) (models.PublicMessage, error) {
	if err := s.applyChannelAccess(ctx, in.AccessToken, in.GroupPublicID, in.ChannelPublicID, channelAccessInput{
		callerUserID:      &in.CallerUserID,
		resolvedChannelID: &in.ResolvedChannelID,
	}); err != nil {
		return models.PublicMessage{}, err
	}

	msgPubID := strings.TrimSpace(in.MessagePublicID)
	if msgPubID == "" {
		return models.PublicMessage{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "message_id"})
	}
	body, err := validateMessageBody(in.Body)
	if err != nil {
		return models.PublicMessage{}, err
	}

	existing, err := s.messages.FindByPublicIDInChannel(ctx, in.ResolvedChannelID, msgPubID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return models.PublicMessage{}, exception.NotFound(exception.MsgInvalidInput, map[string]any{"field": "message_id"})
		}

		return models.PublicMessage{}, err
	}
	if existing.AuthorID != in.CallerUserID {
		return models.PublicMessage{}, exception.PermissionDenied(exception.MsgForbidden, nil)
	}

	editedAt := s.now()
	preview := toPublicMessage(existing, in.CallerUserID)
	preview.Body = body
	preview.EditedAt = &editedAt
	outboxRow, err := pkgChatEvents.EncodeMessageUpdated(
		in.GroupPublicID,
		in.ChannelPublicID,
		toPublicMessageWire(preview),
		editedAt,
	)
	if err != nil {
		return models.PublicMessage{}, err
	}

	updated, err := s.messages.UpdateBodyWithOutbox(ctx, &models.UpdateMessageBodyWithOutboxInput{
		ChannelID: in.ResolvedChannelID,
		PublicID:  msgPubID,
		AuthorID:  in.CallerUserID,
		Body:      body,
		EditedAt:  editedAt,
		Outbox:    toEventOutboxInsert(outboxRow),
	})
	if err != nil {
		return models.PublicMessage{}, err
	}

	return toPublicMessage(updated, in.CallerUserID), nil
}

func (s service) DeleteChannelMessage(ctx context.Context, in *models.DeleteChannelMessageInput) error {
	if err := s.applyChannelAccess(ctx, in.AccessToken, in.GroupPublicID, in.ChannelPublicID, channelAccessInput{
		callerUserID:      &in.CallerUserID,
		resolvedChannelID: &in.ResolvedChannelID,
		callerPermissions: &in.CallerPermissions,
	}); err != nil {
		return err
	}

	msgPubID := strings.TrimSpace(in.MessagePublicID)
	if msgPubID == "" {
		return exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "message_id"})
	}

	existing, err := s.messages.FindByPublicIDInChannel(ctx, in.ResolvedChannelID, msgPubID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return exception.NotFound(exception.MsgInvalidInput, map[string]any{"field": "message_id"})
		}

		return err
	}
	if existing.AuthorID != in.CallerUserID && !hasAnyPerm(in.CallerPermissions, "administrator", "manage_messages") {
		return exception.PermissionDenied(exception.MsgForbidden, nil)
	}

	at := s.now()
	outboxRow, err := pkgChatEvents.EncodeMessageDeleted(in.GroupPublicID, in.ChannelPublicID, msgPubID, at)
	if err != nil {
		return err
	}

	return s.messages.DeleteByPublicIDWithOutbox(ctx, &models.DeleteMessageWithOutboxInput{
		ChannelID: in.ResolvedChannelID,
		PublicID:  msgPubID,
		Outbox:    toEventOutboxInsert(outboxRow),
	})
}

func (s service) ToggleMessageReaction(ctx context.Context, in *models.ToggleMessageReactionInput) ([]models.MessageReactionSummary, error) {
	if err := s.applyChannelAccess(ctx, in.AccessToken, in.GroupPublicID, in.ChannelPublicID, channelAccessInput{
		callerUserID:      &in.CallerUserID,
		resolvedChannelID: &in.ResolvedChannelID,
		callerPermissions: &in.CallerPermissions,
	}); err != nil {
		return nil, err
	}

	msgPubID := strings.TrimSpace(in.MessagePublicID)
	emoji := strings.TrimSpace(in.Emoji)
	if msgPubID == "" || emoji == "" || utf8.RuneCountInString(emoji) > 64 {
		return nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "emoji"})
	}

	existing, err := s.messages.FindByPublicIDInChannel(ctx, in.ResolvedChannelID, msgPubID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, exception.NotFound(exception.MsgInvalidInput, map[string]any{"field": "message_id"})
		}

		return nil, err
	}

	alreadyReacted := false
	for _, r := range existing.Reactions {
		if r.Emoji == emoji && slices.Contains(r.UserIDs, in.CallerUserID) {
			alreadyReacted = true
			break
		}
	}
	if !alreadyReacted && len(in.CallerPermissions) > 0 && !hasAnyPerm(in.CallerPermissions, "administrator", "add_reactions") {
		return nil, exception.PermissionDenied(exception.MsgForbidden, nil)
	}

	at := s.now()
	predicted := predictToggledReactions(existing.Reactions, in.CallerUserID, emoji)
	outboxRow, err := pkgChatEvents.EncodeMessageReactionsUpdated(
		in.GroupPublicID,
		in.ChannelPublicID,
		msgPubID,
		toReactionWire(predicted),
		at,
	)
	if err != nil {
		return nil, err
	}

	reactions, err := s.messages.ToggleReactionWithOutbox(ctx, &models.ToggleReactionWithOutboxInput{
		MessageID: existing.ID,
		UserID:    in.CallerUserID,
		Emoji:     emoji,
		Now:       at,
		Outbox:    toEventOutboxInsert(outboxRow),
	})
	if err != nil {
		return nil, err
	}

	return reactions, nil
}

func (s service) SetChannelMessagePinned(ctx context.Context, in *models.SetChannelMessagePinnedInput) (models.PublicMessage, error) {
	if err := s.applyChannelAccess(ctx, in.AccessToken, in.GroupPublicID, in.ChannelPublicID, channelAccessInput{
		callerUserID:      &in.CallerUserID,
		resolvedChannelID: &in.ResolvedChannelID,
		callerPermissions: &in.CallerPermissions,
	}); err != nil {
		return models.PublicMessage{}, err
	}

	msgPubID := strings.TrimSpace(in.MessagePublicID)
	if msgPubID == "" {
		return models.PublicMessage{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "message_id"})
	}
	if len(in.CallerPermissions) > 0 && !hasAnyPerm(in.CallerPermissions, "administrator", "pin_messages", "manage_messages") {
		return models.PublicMessage{}, exception.PermissionDenied(exception.MsgForbidden, nil)
	}

	existing, err := s.messages.FindByPublicIDInChannel(ctx, in.ResolvedChannelID, msgPubID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return models.PublicMessage{}, exception.NotFound(exception.MsgInvalidInput, map[string]any{"field": "message_id"})
		}

		return models.PublicMessage{}, err
	}

	at := s.now()
	preview := toPublicMessage(existing, in.CallerUserID)
	preview.IsPinned = in.Pinned
	outboxRow, err := pkgChatEvents.EncodeMessageUpdated(
		in.GroupPublicID,
		in.ChannelPublicID,
		toPublicMessageWire(preview),
		at,
	)
	if err != nil {
		return models.PublicMessage{}, err
	}

	updated, err := s.messages.SetPinnedWithOutbox(ctx, &models.SetPinnedWithOutboxInput{
		ChannelID: in.ResolvedChannelID,
		PublicID:  msgPubID,
		UserID:    in.CallerUserID,
		Pinned:    in.Pinned,
		Now:       at,
		Outbox:    toEventOutboxInsert(outboxRow),
	})
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return models.PublicMessage{}, exception.NotFound(exception.MsgInvalidInput, map[string]any{"field": "message_id"})
		}

		return models.PublicMessage{}, err
	}

	return toPublicMessage(updated, in.CallerUserID), nil
}

func (s service) ListPinnedChannelMessages(ctx context.Context, in *models.ListPinnedChannelMessagesInput) ([]models.PublicMessage, error) {
	if err := s.applyChannelAccess(ctx, in.AccessToken, in.GroupPublicID, in.ChannelPublicID, channelAccessInput{
		callerUserID:      &in.CallerUserID,
		resolvedChannelID: &in.ResolvedChannelID,
		callerPermissions: &in.CallerPermissions,
	}); err != nil {
		return nil, err
	}

	if !hasAnyPerm(in.CallerPermissions, "administrator", "read_message_history") {
		return []models.PublicMessage{}, nil
	}

	rows, err := s.messages.ListPinnedByChannel(ctx, in.ResolvedChannelID)
	if err != nil {
		return nil, err
	}

	out := make([]models.PublicMessage, len(rows))
	for i, row := range rows {
		out[i] = toPublicMessage(row, in.CallerUserID)
	}

	return out, nil
}

func hasDangerousControlChars(s string) bool {
	for _, r := range s {
		if r == '\n' || r == '\r' || r == '\t' {
			continue
		}
		if unicode.Is(unicode.Cc, r) {
			return true
		}
	}

	return false
}

type channelAccessInput struct {
	callerUserID      *string
	resolvedChannelID *string
	callerPermissions *[]string
	channelType       *string
	slowmodeSeconds   *int32
}

func (s service) applyChannelAccess(ctx context.Context, accessToken, groupID, channelID string, in channelAccessInput) error {
	if s.userChannel == nil {
		return errors.New("channel client required")
	}

	out, err := s.userChannel.AssertGroupChannelAccess(ctx, &userv1.AssertGroupChannelAccessRequest{
		AccessToken: accessToken,
		GroupId:     groupID,
		ChannelId:   channelID,
	})
	if err != nil {
		return err
	}

	resolved := strings.TrimSpace(out.GetChannelUuid())
	if in.resolvedChannelID != nil {
		if prev := strings.TrimSpace(*in.resolvedChannelID); prev != "" && prev != resolved {
			return exception.PermissionDenied(exception.MsgForbidden, nil)
		}
		*in.resolvedChannelID = resolved
	}
	if in.callerUserID != nil {
		*in.callerUserID = out.GetUserUuid()
	}
	if in.callerPermissions != nil {
		*in.callerPermissions = out.GetPermissions()
	}
	if in.channelType != nil {
		*in.channelType = out.GetChannelType()
	}
	if in.slowmodeSeconds != nil {
		*in.slowmodeSeconds = out.GetSlowmodeSeconds()
	}

	return nil
}

func (s service) callerProfile(ctx context.Context, accessToken, callerUserID string) (displayName, avatarURL string, err error) {
	if s.userAuth == nil {
		return "", "", nil
	}
	token := strings.TrimSpace(accessToken)
	if token == "" {
		return "", "", nil
	}

	me, err := s.userAuth.Me(ctx, &userv1.MeRequest{AccessToken: token})
	if err != nil {
		return "", "", err
	}
	u := me.GetUser()
	if u == nil || u.GetId() != callerUserID {
		return "", "", exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}

	return u.GetDisplayName(), u.GetAvatarUrl(), nil
}

func toPublicMessage(row models.MessageWithAuthor, callerUserID string) models.PublicMessage {
	clientMsgID := ""
	if row.ClientMsgID != nil {
		clientMsgID = *row.ClientMsgID
	}

	return models.PublicMessage{
		ID:                row.PublicID,
		ChannelSeq:        row.ChannelSeq,
		ClientMsgID:       clientMsgID,
		AuthorID:          row.AuthorID,
		AuthorDisplayName: row.AuthorDisplayName,
		AuthorAvatarURL:   row.AuthorAvatarURL,
		Body:              row.Body,
		CreatedAt:         row.CreatedAt,
		EditedAt:          row.EditedAt,
		IsPinned:          row.IsPinned,
		ReplyTo:           row.ReplyTo,
		Reactions:         row.Reactions,
		Mine:              row.AuthorID == callerUserID,
	}
}

var allowedAttachmentMIMEs = map[string]bool{
	"image/jpeg":                    true,
	"image/png":                     true,
	"image/gif":                     true,
	"image/webp":                    true,
	"image/avif":                    true,
	"application/pdf":               true,
	"text/plain":                    true,
	"text/markdown":                 true,
	"text/csv":                      true,
	"application/json":              true,
	"application/zip":               true,
	"application/x-zip-compressed":  true,
	"application/gzip":              true,
	"application/x-gzip":            true,
	"application/msword":            true,
	"application/vnd.ms-excel":      true,
	"application/vnd.ms-powerpoint": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         true,
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
	"video/mp4":  true,
	"video/webm": true,
	"audio/mpeg": true,
	"audio/ogg":  true,
	"audio/wav":  true,
	"audio/webm": true,
}

var dangerousAttachmentMIMEs = map[string]bool{
	"text/html":                true,
	"application/xhtml+xml":    true,
	"image/svg+xml":            true,
	"application/javascript":   true,
	"text/javascript":          true,
	"application/x-msdownload": true,
	"application/x-executable": true,
	"application/x-sh":         true,
}

func (s service) UploadChannelAttachment(ctx context.Context, in *models.UploadChannelAttachmentInput) (models.MessageAttachment, error) {
	if err := s.applyChannelAccess(ctx, in.AccessToken, in.GroupPublicID, in.ChannelPublicID, channelAccessInput{
		callerUserID:      &in.CallerUserID,
		resolvedChannelID: &in.ResolvedChannelID,
		callerPermissions: &in.CallerPermissions,
		channelType:       &in.ChannelType,
	}); err != nil {
		return models.MessageAttachment{}, err
	}

	// Permissions & channel policy
	if len(in.CallerPermissions) > 0 {
		if !hasAnyPerm(in.CallerPermissions, "administrator", "send_messages") ||
			!hasAnyPerm(in.CallerPermissions, "administrator", "attach_files") {
			return models.MessageAttachment{}, exception.PermissionDenied(exception.MsgForbidden, nil)
		}
		if in.ChannelType == "broadcast" && !hasAnyPerm(in.CallerPermissions, "administrator", "manage_channels", "manage_messages") {
			return models.MessageAttachment{}, exception.PermissionDenied(exception.MsgForbidden, nil)
		}
	}

	if s.attachments == nil {
		return models.MessageAttachment{}, exception.Internal(errors.New("attachment storage is not configured"))
	}

	size := int64(len(in.Data))
	if size <= 0 || size > constants.MaxAttachmentBytes {
		return models.MessageAttachment{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "file"})
	}

	contentType, ok := resolveSafeAttachmentMIME(in.ContentType, in.Data)
	if !ok {
		return models.MessageAttachment{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "content_type"})
	}

	safeFilename := sanitizeAttachmentFilename(in.Filename)
	attachmentID := uuid.NewString()

	publicURL, err := s.attachments.PutChannelAttachment(
		ctx,
		in.ResolvedChannelID,
		attachmentID,
		safeFilename,
		contentType,
		bytes.NewReader(in.Data),
		size,
	)
	if err != nil {
		return models.MessageAttachment{}, err
	}

	return models.MessageAttachment{
		ID:          attachmentID,
		URL:         publicURL,
		Filename:    safeFilename,
		ContentType: contentType,
		SizeBytes:   size,
	}, nil
}

func normalizeMIME(raw string) string {
	mime := strings.ToLower(strings.TrimSpace(raw))
	if idx := strings.IndexByte(mime, ';'); idx >= 0 {
		mime = strings.TrimSpace(mime[:idx])
	}
	return mime
}

func resolveSafeAttachmentMIME(declared string, data []byte) (string, bool) {
	sniffed := normalizeMIME(http.DetectContentType(data))
	decl := normalizeMIME(declared)

	if dangerousAttachmentMIMEs[sniffed] || dangerousAttachmentMIMEs[decl] {
		return "", false
	}

	// Text formats (markdown, csv, json) sniff as text/plain; trust safe declared subtype.
	if sniffed == "text/plain" && (decl == "text/markdown" || decl == "text/csv" || decl == "application/json") {
		return decl, true
	}

	// Office OpenXML (.docx, .xlsx, .pptx) sniff as application/zip; allow safe declared Office MIME.
	if sniffed == "application/zip" && strings.HasPrefix(decl, "application/vnd.openxmlformats-officedocument.") && allowedAttachmentMIMEs[decl] {
		return decl, true
	}

	// WebP/AVIF or other media where declared MIME is safe and sniffed is generic octet-stream.
	if sniffed == "application/octet-stream" && allowedAttachmentMIMEs[decl] {
		return decl, true
	}

	if allowedAttachmentMIMEs[sniffed] {
		return sniffed, true
	}

	return "", false
}

func sanitizeAttachmentFilename(raw string) string {
	base := path.Base(strings.ReplaceAll(strings.TrimSpace(raw), "\\", "/"))
	var b strings.Builder
	for _, r := range base {
		if unicode.IsControl(r) {
			continue
		}
		switch r {
		case '<', '>', ':', '"', '/', '\\', '|', '?', '*':
			b.WriteByte('_')
		default:
			b.WriteRune(r)
		}
	}
	clean := strings.TrimSpace(b.String())
	if clean == "" || clean == "." || clean == ".." {
		return "file"
	}
	if utf8.RuneCountInString(clean) > 180 {
		ext := path.Ext(clean)
		runes := []rune(clean)
		maxStem := 180 - utf8.RuneCountInString(ext)
		if maxStem < 1 {
			return string(runes[:180])
		}
		return string(runes[:maxStem]) + ext
	}
	return clean
}

func toEventOutboxInsert(in pkgChatEvents.OutboxInsert) models.EventOutboxInsert {
	return models.EventOutboxInsert{
		EventID:      in.EventID,
		EventType:    in.EventType,
		PartitionKey: in.PartitionKey,
		Payload:      in.Payload,
		CreatedAt:    in.CreatedAt,
	}
}

func toPublicMessageWire(msg models.PublicMessage) pkgChatEvents.PublicMessageWire {
	var reply *pkgChatEvents.ReplyPreviewWire
	if msg.ReplyTo != nil {
		reply = &pkgChatEvents.ReplyPreviewWire{
			ID:                msg.ReplyTo.ID,
			AuthorID:          msg.ReplyTo.AuthorID,
			AuthorDisplayName: msg.ReplyTo.AuthorDisplayName,
			BodyPreview:       msg.ReplyTo.BodyPreview,
		}
	}
	reactions := make([]pkgChatEvents.ReactionSummaryWire, len(msg.Reactions))
	for i, r := range msg.Reactions {
		reactions[i] = pkgChatEvents.ReactionSummaryWire{
			Emoji:   r.Emoji,
			Count:   r.Count,
			UserIDs: r.UserIDs,
		}
	}

	return pkgChatEvents.PublicMessageWire{
		ID: msg.ID, ChannelSeq: msg.ChannelSeq, ClientMsgID: msg.ClientMsgID, AuthorID: msg.AuthorID,
		AuthorDisplayName: msg.AuthorDisplayName, AuthorAvatarURL: msg.AuthorAvatarURL,
		Body: msg.Body, CreatedAt: msg.CreatedAt, EditedAt: msg.EditedAt, IsPinned: msg.IsPinned,
		ReplyTo: reply, Reactions: reactions,
	}
}

func toReactionWire(in []models.MessageReactionSummary) []pkgChatEvents.ReactionSummaryWire {
	out := make([]pkgChatEvents.ReactionSummaryWire, len(in))
	for i, r := range in {
		out[i] = pkgChatEvents.ReactionSummaryWire{Emoji: r.Emoji, Count: r.Count, UserIDs: r.UserIDs}
	}

	return out
}

func predictToggledReactions(
	current []models.MessageReactionSummary,
	userID, emoji string,
) []models.MessageReactionSummary {
	removed := false
	out := make([]models.MessageReactionSummary, 0, len(current)+1)
	for _, r := range current {
		if r.Emoji != emoji {
			out = append(out, r)
			continue
		}
		filtered := make([]string, 0, len(r.UserIDs))
		for _, uid := range r.UserIDs {
			if uid == userID {
				removed = true
				continue
			}
			filtered = append(filtered, uid)
		}
		if len(filtered) > 0 {
			out = append(out, models.MessageReactionSummary{
				Emoji:   r.Emoji,
				Count:   reactionCount(len(filtered)),
				UserIDs: filtered,
			})
		}
	}
	if !removed {
		found := false
		for i, r := range out {
			if r.Emoji == emoji {
				users := append(append([]string{}, r.UserIDs...), userID)
				out[i] = models.MessageReactionSummary{
					Emoji: emoji, Count: reactionCount(len(users)), UserIDs: users,
				}
				found = true
				break
			}
		}
		if !found {
			out = append(out, models.MessageReactionSummary{Emoji: emoji, Count: 1, UserIDs: []string{userID}})
		}
	}

	return out
}

func reactionCount(n int) int32 {
	if n <= 0 {
		return 0
	}
	if n > math.MaxInt32 {
		return math.MaxInt32
	}

	return int32(n)
}
