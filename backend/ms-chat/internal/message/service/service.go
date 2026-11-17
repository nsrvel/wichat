package service

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"path"
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/wichat/wichat/backend/ms-chat/internal/kafka/producers"
	messagerepo "github.com/wichat/wichat/backend/ms-chat/internal/message/repository"
	"github.com/wichat/wichat/backend/ms-chat/internal/models"
	"github.com/wichat/wichat/backend/ms-chat/pkg/constants"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	userchannel "github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user/grpc/channel"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

// Service is channel message use cases.
type Service interface {
	ListChannelMessages(ctx context.Context, in *models.ListChannelMessagesInput) ([]models.PublicMessage, error)
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
	userChannel userchannel.Client
	events      producers.MessageCreatedPublisher
	log         logger.Logger
	now         func() time.Time
}

// NewService builds message use cases.
func NewService(
	messages messagerepo.PgRepository,
	attachments messagerepo.MinioRepository,
	userChannel userchannel.Client,
	events producers.MessageCreatedPublisher,
	log logger.Logger,
) Service {
	return service{
		messages:    messages,
		attachments: attachments,
		userChannel: userChannel,
		events:      events,
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

func (s service) ListChannelMessages(ctx context.Context, in *models.ListChannelMessagesInput) ([]models.PublicMessage, error) {
	if len(in.CallerPermissions) > 0 && !hasAnyPerm(in.CallerPermissions, "administrator", "read_message_history") {
		return []models.PublicMessage{}, nil
	}

	// Repository
	rows, err := s.messages.ListByChannel(ctx, in.ResolvedChannelID, in.Limit, in.BeforeCreatedAt)
	if err != nil {
		return nil, err
	}

	out := make([]models.PublicMessage, len(rows))
	for i, row := range rows {
		out[i] = toPublicMessage(row, in.CallerUserID)
	}

	return out, nil
}

func (s service) CreateChannelMessage(ctx context.Context, in *models.CreateChannelMessageInput) (models.PublicMessage, error) {
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

	// Repository: resolve author profile before persisting message.
	displayName, avatarURL, err := s.messages.FindUserProfile(ctx, in.CallerUserID)
	if err != nil {
		return models.PublicMessage{}, err
	}

	createdAt := s.now()
	msg, err := s.messages.Insert(ctx, in.ResolvedChannelID, in.CallerUserID, body, replyToID, createdAt)
	if err != nil {
		return models.PublicMessage{}, err
	}

	out := models.PublicMessage{
		ID: msg.PublicID, AuthorID: msg.AuthorID, AuthorDisplayName: displayName, AuthorAvatarURL: avatarURL,
		Body: msg.Body, CreatedAt: msg.CreatedAt, ReplyTo: replyPreview, Mine: true,
	}

	pubCtx, pubCancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer pubCancel()
	if err := s.events.PublishMessageCreated(pubCtx, in.GroupPublicID, in.ChannelPublicID, out); err != nil {
		if s.log != nil {
			s.log.Error("publish message.created", "message_id", msg.ID, "err", err)
		}
	}

	return out, nil
}

func (s service) UpdateChannelMessage(ctx context.Context, in *models.UpdateChannelMessageInput) (models.PublicMessage, error) {
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

	updated, err := s.messages.UpdateBody(ctx, in.ResolvedChannelID, msgPubID, in.CallerUserID, body, s.now())
	if err != nil {
		return models.PublicMessage{}, err
	}

	out := toPublicMessage(updated, in.CallerUserID)
	pubCtx, pubCancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer pubCancel()
	if err := s.events.PublishMessageUpdated(pubCtx, in.GroupPublicID, in.ChannelPublicID, out); err != nil && s.log != nil {
		s.log.Error("publish message.updated", "message_id", updated.ID, "err", err)
	}

	return out, nil
}

func (s service) DeleteChannelMessage(ctx context.Context, in *models.DeleteChannelMessageInput) error {
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

	if err := s.messages.DeleteByPublicID(ctx, in.ResolvedChannelID, msgPubID); err != nil {
		return err
	}

	pubCtx, pubCancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer pubCancel()
	if err := s.events.PublishMessageDeleted(pubCtx, in.GroupPublicID, in.ChannelPublicID, msgPubID); err != nil && s.log != nil {
		s.log.Error("publish message.deleted", "message_id", existing.ID, "err", err)
	}

	return nil
}

func (s service) ToggleMessageReaction(ctx context.Context, in *models.ToggleMessageReactionInput) ([]models.MessageReactionSummary, error) {
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

	reactions, err := s.messages.ToggleReaction(ctx, existing.ID, in.CallerUserID, emoji, s.now())
	if err != nil {
		return nil, err
	}

	pubCtx, pubCancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer pubCancel()
	if err := s.events.PublishMessageReactionsUpdated(pubCtx, in.GroupPublicID, in.ChannelPublicID, msgPubID, reactions); err != nil && s.log != nil {
		s.log.Error("publish message.reactions.updated", "message_id", existing.ID, "err", err)
	}

	return reactions, nil
}

func (s service) SetChannelMessagePinned(ctx context.Context, in *models.SetChannelMessagePinnedInput) (models.PublicMessage, error) {
	msgPubID := strings.TrimSpace(in.MessagePublicID)
	if msgPubID == "" {
		return models.PublicMessage{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "message_id"})
	}
	if len(in.CallerPermissions) > 0 && !hasAnyPerm(in.CallerPermissions, "administrator", "pin_messages", "manage_messages") {
		return models.PublicMessage{}, exception.PermissionDenied(exception.MsgForbidden, nil)
	}

	updated, err := s.messages.SetPinned(ctx, in.ResolvedChannelID, msgPubID, in.CallerUserID, in.Pinned, s.now())
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return models.PublicMessage{}, exception.NotFound(exception.MsgInvalidInput, map[string]any{"field": "message_id"})
		}

		return models.PublicMessage{}, err
	}

	out := toPublicMessage(updated, in.CallerUserID)
	pubCtx, pubCancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer pubCancel()
	if err := s.events.PublishMessageUpdated(pubCtx, in.GroupPublicID, in.ChannelPublicID, out); err != nil && s.log != nil {
		s.log.Error("publish message.updated", "message_id", updated.ID, "err", err)
	}

	return out, nil
}

func (s service) ListPinnedChannelMessages(ctx context.Context, in *models.ListPinnedChannelMessagesInput) ([]models.PublicMessage, error) {
	if len(in.CallerPermissions) > 0 && !hasAnyPerm(in.CallerPermissions, "administrator", "read_message_history") {
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

// ResolveChannelAccess validates the bearer token and resolves internal ids and permissions via ms-user.
func ResolveChannelAccess(
	ctx context.Context,
	userChannel userchannel.Client,
	accessToken, groupID, channelID string,
) (models.ChannelAccessInfo, error) {
	out, err := userChannel.AssertGroupChannelAccess(ctx, &userv1.AssertGroupChannelAccessRequest{
		AccessToken: accessToken,
		GroupId:     groupID,
		ChannelId:   channelID,
	})
	if err != nil {
		return models.ChannelAccessInfo{}, err
	}

	return models.ChannelAccessInfo{
		ChannelUUID:     out.GetChannelUuid(),
		UserUUID:        out.GetUserUuid(),
		Permissions:     out.GetPermissions(),
		ChannelType:     out.GetChannelType(),
		SlowmodeSeconds: out.GetSlowmodeSeconds(),
	}, nil
}

func toPublicMessage(row models.MessageWithAuthor, callerUserID string) models.PublicMessage {
	return models.PublicMessage{
		ID:                row.PublicID,
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
