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

	channelUUID, userUUID, err := messageservice.ResolveChannelAccess(ctx, s.userChannel, req.GetAccessToken(), req.GetGroupId(), req.GetChannelId())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Service
	messages, err := s.svc.ListChannelMessages(ctx, &models.ListChannelMessagesInput{
		AccessToken:       req.GetAccessToken(),
		GroupPublicID:     req.GetGroupId(),
		ChannelPublicID:   req.GetChannelId(),
		CallerUserID:      userUUID,
		ResolvedChannelID: channelUUID,
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
	channelUUID, userUUID, err := messageservice.ResolveChannelAccess(ctx, s.userChannel, req.GetAccessToken(), req.GetGroupId(), req.GetChannelId())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Service
	msg, err := s.svc.CreateChannelMessage(ctx, &models.CreateChannelMessageInput{
		AccessToken:       req.GetAccessToken(),
		GroupPublicID:     req.GetGroupId(),
		ChannelPublicID:   req.GetChannelId(),
		CallerUserID:      userUUID,
		ResolvedChannelID: channelUUID,
		Body:              req.GetBody(),
	})
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &chatv1.CreateChannelMessageResponse{Message: toMessagePB(msg)}, nil
}

func toMessagePB(msg models.PublicMessage) *chatv1.MessageSummary {
	return &chatv1.MessageSummary{
		Id:                msg.ID,
		AuthorId:          msg.AuthorID,
		AuthorDisplayName: msg.AuthorDisplayName,
		Body:              msg.Body,
		CreatedAt:         msg.CreatedAt.UTC().Format(time.RFC3339Nano),
		Mine:              msg.Mine,
	}
}
