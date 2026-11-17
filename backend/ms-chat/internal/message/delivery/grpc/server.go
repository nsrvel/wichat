package grpc

import (
	"context"
	"time"

	messageservice "github.com/wichat/wichat/backend/ms-chat/internal/message/service"
	"github.com/wichat/wichat/backend/ms-chat/internal/models"
	"github.com/wichat/wichat/backend/ms-chat/pkg/constants"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	userchannel "github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user/grpc/channel"
	chatv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/chat/v1"
	responsegrpc "github.com/wichat/wichat/backend/wi-shared/response/grpc"
)

// Server is the message gRPC delivery.
type Server interface {
	chatv1.MessageServiceServer
}

type server struct {
	chatv1.UnimplementedMessageServiceServer
	svc         messageservice.Service
	userChannel userchannel.Client
}

// NewServer binds the message service to gRPC.
func NewServer(svc messageservice.Service, userChannel userchannel.Client) Server {
	return &server{svc: svc, userChannel: userChannel}
}

func (s *server) ListChannelMessages(ctx context.Context, req *chatv1.ListChannelMessagesRequest) (*chatv1.ListChannelMessagesResponse, error) {
	// Bind
	limit := int(req.GetLimit())
	if limit <= 0 {
		limit = constants.DefaultMessageListLimit
	}
	var before *time.Time
	if raw := req.GetBeforeCreatedAt(); raw != "" {
		parsed, err := time.Parse(time.RFC3339Nano, raw)
		if err != nil {
			parsed, err = time.Parse(time.RFC3339, raw)
			if err != nil {
				return nil, responsegrpc.ToStatus(exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "before_created_at"}))
			}
		}
		before = &parsed
	}

	acc, err := messageservice.ResolveChannelAccess(ctx, s.userChannel, req.GetAccessToken(), req.GetGroupId(), req.GetChannelId())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Service
	messages, err := s.svc.ListChannelMessages(ctx, &models.ListChannelMessagesInput{
		AccessToken:       req.GetAccessToken(),
		GroupPublicID:     req.GetGroupId(),
		ChannelPublicID:   req.GetChannelId(),
		CallerUserID:      acc.UserUUID,
		ResolvedChannelID: acc.ChannelUUID,
		CallerPermissions: acc.Permissions,
		Limit:             limit,
		BeforeCreatedAt:   before,
	})
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	out := make([]*chatv1.MessageSummary, len(messages))
	for i, msg := range messages {
		out[i] = toMessagePB(msg)
	}

	return &chatv1.ListChannelMessagesResponse{Messages: out}, nil
}

func (s *server) CreateChannelMessage(ctx context.Context, req *chatv1.CreateChannelMessageRequest) (*chatv1.CreateChannelMessageResponse, error) {
	acc, err := messageservice.ResolveChannelAccess(ctx, s.userChannel, req.GetAccessToken(), req.GetGroupId(), req.GetChannelId())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Service
	msg, err := s.svc.CreateChannelMessage(ctx, &models.CreateChannelMessageInput{
		AccessToken:       req.GetAccessToken(),
		GroupPublicID:     req.GetGroupId(),
		ChannelPublicID:   req.GetChannelId(),
		CallerUserID:      acc.UserUUID,
		ResolvedChannelID: acc.ChannelUUID,
		CallerPermissions: acc.Permissions,
		ChannelType:       acc.ChannelType,
		SlowmodeSeconds:   acc.SlowmodeSeconds,
		Body:              req.GetBody(),
		ReplyToPublicID:   req.GetReplyToId(),
	})
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &chatv1.CreateChannelMessageResponse{Message: toMessagePB(msg)}, nil
}

func (s *server) UpdateChannelMessage(ctx context.Context, req *chatv1.UpdateChannelMessageRequest) (*chatv1.UpdateChannelMessageResponse, error) {
	acc, err := messageservice.ResolveChannelAccess(ctx, s.userChannel, req.GetAccessToken(), req.GetGroupId(), req.GetChannelId())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	msg, err := s.svc.UpdateChannelMessage(ctx, &models.UpdateChannelMessageInput{
		GroupPublicID:     req.GetGroupId(),
		ChannelPublicID:   req.GetChannelId(),
		CallerUserID:      acc.UserUUID,
		ResolvedChannelID: acc.ChannelUUID,
		MessagePublicID:   req.GetMessageId(),
		Body:              req.GetBody(),
	})
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &chatv1.UpdateChannelMessageResponse{Message: toMessagePB(msg)}, nil
}

func (s *server) DeleteChannelMessage(ctx context.Context, req *chatv1.DeleteChannelMessageRequest) (*chatv1.DeleteChannelMessageResponse, error) {
	acc, err := messageservice.ResolveChannelAccess(ctx, s.userChannel, req.GetAccessToken(), req.GetGroupId(), req.GetChannelId())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	if err := s.svc.DeleteChannelMessage(ctx, &models.DeleteChannelMessageInput{
		GroupPublicID:     req.GetGroupId(),
		ChannelPublicID:   req.GetChannelId(),
		CallerUserID:      acc.UserUUID,
		ResolvedChannelID: acc.ChannelUUID,
		CallerPermissions: acc.Permissions,
		MessagePublicID:   req.GetMessageId(),
	}); err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &chatv1.DeleteChannelMessageResponse{Ok: true}, nil
}

func (s *server) ToggleMessageReaction(ctx context.Context, req *chatv1.ToggleMessageReactionRequest) (*chatv1.ToggleMessageReactionResponse, error) {
	acc, err := messageservice.ResolveChannelAccess(ctx, s.userChannel, req.GetAccessToken(), req.GetGroupId(), req.GetChannelId())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	reactions, err := s.svc.ToggleMessageReaction(ctx, &models.ToggleMessageReactionInput{
		GroupPublicID:     req.GetGroupId(),
		ChannelPublicID:   req.GetChannelId(),
		CallerUserID:      acc.UserUUID,
		ResolvedChannelID: acc.ChannelUUID,
		CallerPermissions: acc.Permissions,
		MessagePublicID:   req.GetMessageId(),
		Emoji:             req.GetEmoji(),
	})
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &chatv1.ToggleMessageReactionResponse{Reactions: toReactionsPB(reactions)}, nil
}

func (s *server) SetChannelMessagePinned(ctx context.Context, req *chatv1.SetChannelMessagePinnedRequest) (*chatv1.SetChannelMessagePinnedResponse, error) {
	acc, err := messageservice.ResolveChannelAccess(ctx, s.userChannel, req.GetAccessToken(), req.GetGroupId(), req.GetChannelId())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	msg, err := s.svc.SetChannelMessagePinned(ctx, &models.SetChannelMessagePinnedInput{
		GroupPublicID:     req.GetGroupId(),
		ChannelPublicID:   req.GetChannelId(),
		CallerUserID:      acc.UserUUID,
		ResolvedChannelID: acc.ChannelUUID,
		CallerPermissions: acc.Permissions,
		MessagePublicID:   req.GetMessageId(),
		Pinned:            req.GetPinned(),
	})
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &chatv1.SetChannelMessagePinnedResponse{Message: toMessagePB(msg)}, nil
}

func (s *server) ListPinnedChannelMessages(ctx context.Context, req *chatv1.ListPinnedChannelMessagesRequest) (*chatv1.ListPinnedChannelMessagesResponse, error) {
	acc, err := messageservice.ResolveChannelAccess(ctx, s.userChannel, req.GetAccessToken(), req.GetGroupId(), req.GetChannelId())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	messages, err := s.svc.ListPinnedChannelMessages(ctx, &models.ListPinnedChannelMessagesInput{
		CallerUserID:      acc.UserUUID,
		ResolvedChannelID: acc.ChannelUUID,
		CallerPermissions: acc.Permissions,
	})
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	out := make([]*chatv1.MessageSummary, len(messages))
	for i, msg := range messages {
		out[i] = toMessagePB(msg)
	}

	return &chatv1.ListPinnedChannelMessagesResponse{Messages: out}, nil
}

func (s *server) UploadChannelAttachment(ctx context.Context, req *chatv1.UploadChannelAttachmentRequest) (*chatv1.UploadChannelAttachmentResponse, error) {
	acc, err := messageservice.ResolveChannelAccess(ctx, s.userChannel, req.GetAccessToken(), req.GetGroupId(), req.GetChannelId())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	att, err := s.svc.UploadChannelAttachment(ctx, &models.UploadChannelAttachmentInput{
		GroupPublicID:     req.GetGroupId(),
		ChannelPublicID:   req.GetChannelId(),
		CallerUserID:      acc.UserUUID,
		ResolvedChannelID: acc.ChannelUUID,
		CallerPermissions: acc.Permissions,
		ChannelType:       acc.ChannelType,
		Filename:          req.GetFilename(),
		ContentType:       req.GetContentType(),
		Data:              req.GetData(),
	})
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &chatv1.UploadChannelAttachmentResponse{
		Attachment: &chatv1.MessageAttachment{
			Id:          att.ID,
			Url:         att.URL,
			Filename:    att.Filename,
			ContentType: att.ContentType,
			SizeBytes:   att.SizeBytes,
		},
	}, nil
}

func toReactionsPB(reactions []models.MessageReactionSummary) []*chatv1.MessageReactionSummary {
	if len(reactions) == 0 {
		return nil
	}
	out := make([]*chatv1.MessageReactionSummary, len(reactions))
	for i, r := range reactions {
		out[i] = &chatv1.MessageReactionSummary{
			Emoji:   r.Emoji,
			Count:   r.Count,
			UserIds: r.UserIDs,
		}
	}

	return out
}

func toMessagePB(msg models.PublicMessage) *chatv1.MessageSummary {
	var editedAt string
	if msg.EditedAt != nil {
		editedAt = msg.EditedAt.UTC().Format(time.RFC3339Nano)
	}
	var reply *chatv1.MessageReplyPreview
	if msg.ReplyTo != nil {
		reply = &chatv1.MessageReplyPreview{
			Id:                msg.ReplyTo.ID,
			AuthorId:          msg.ReplyTo.AuthorID,
			AuthorDisplayName: msg.ReplyTo.AuthorDisplayName,
			BodyPreview:       msg.ReplyTo.BodyPreview,
		}
	}

	return &chatv1.MessageSummary{
		Id:                msg.ID,
		AuthorId:          msg.AuthorID,
		AuthorDisplayName: msg.AuthorDisplayName,
		AuthorAvatarUrl:   msg.AuthorAvatarURL,
		Body:              msg.Body,
		CreatedAt:         msg.CreatedAt.UTC().Format(time.RFC3339Nano),
		EditedAt:          editedAt,
		IsPinned:          msg.IsPinned,
		ReplyTo:           reply,
		Reactions:         toReactionsPB(msg.Reactions),
		Mine:              msg.Mine,
	}
}
