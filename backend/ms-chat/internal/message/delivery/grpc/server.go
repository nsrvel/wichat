package grpc

import (
	"context"

	messageservice "github.com/wichat/wichat/backend/ms-chat/internal/message/service"
	"github.com/wichat/wichat/backend/ms-chat/internal/models"
	chatPb "github.com/wichat/wichat/backend/ms-chat/pkg/chat-pb"
	"github.com/wichat/wichat/backend/ms-chat/pkg/constants"
	chatv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/chat/v1"
	responsegrpc "github.com/wichat/wichat/backend/wi-shared/response/grpc"
)

// Server is the message gRPC delivery.
type Server interface {
	chatv1.MessageServiceServer
}

type server struct {
	chatv1.UnimplementedMessageServiceServer
	svc messageservice.Service
}

// NewServer binds the message service to gRPC.
func NewServer(svc messageservice.Service) Server {
	return &server{svc: svc}
}

func (s *server) ListChannelMessages(ctx context.Context, req *chatv1.ListChannelMessagesRequest) (*chatv1.ListChannelMessagesResponse, error) {
	// Bind
	limit := int(req.GetLimit())
	if limit <= 0 {
		limit = constants.DefaultMessageListLimit
	}
	// Service
	messages, hasMore, err := s.svc.ListChannelMessages(ctx, &models.ListChannelMessagesInput{
		AccessToken:      req.GetAccessToken(),
		GroupPublicID:    req.GetGroupId(),
		ChannelPublicID:  req.GetChannelId(),
		Limit:            limit,
		BeforeChannelSeq: req.GetBeforeChannelSeq(),
		AfterChannelSeq:  req.GetAfterChannelSeq(),
	})
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &chatv1.ListChannelMessagesResponse{
		Messages: chatPb.MessageSummaries(messages),
		HasMore:  hasMore,
	}, nil
}

func (s *server) CreateChannelMessage(ctx context.Context, req *chatv1.CreateChannelMessageRequest) (*chatv1.CreateChannelMessageResponse, error) {
	// Service
	msg, err := s.svc.CreateChannelMessage(ctx, &models.CreateChannelMessageInput{
		AccessToken:      req.GetAccessToken(),
		GroupPublicID:    req.GetGroupId(),
		ChannelPublicID:  req.GetChannelId(),
		Body:             req.GetBody(),
		ReplyToPublicID:  req.GetReplyToId(),
		ClientMsgID:      req.GetClientMsgId(),
		MentionedUserIDs: req.GetMentionedUserIds(),
	})
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &chatv1.CreateChannelMessageResponse{Message: chatPb.MessageSummary(msg)}, nil
}

func (s *server) UpdateChannelMessage(ctx context.Context, req *chatv1.UpdateChannelMessageRequest) (*chatv1.UpdateChannelMessageResponse, error) {
	msg, err := s.svc.UpdateChannelMessage(ctx, &models.UpdateChannelMessageInput{
		AccessToken:     req.GetAccessToken(),
		GroupPublicID:   req.GetGroupId(),
		ChannelPublicID: req.GetChannelId(),
		MessagePublicID: req.GetMessageId(),
		Body:            req.GetBody(),
	})
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &chatv1.UpdateChannelMessageResponse{Message: chatPb.MessageSummary(msg)}, nil
}

func (s *server) DeleteChannelMessage(ctx context.Context, req *chatv1.DeleteChannelMessageRequest) (*chatv1.DeleteChannelMessageResponse, error) {
	if err := s.svc.DeleteChannelMessage(ctx, &models.DeleteChannelMessageInput{
		AccessToken:     req.GetAccessToken(),
		GroupPublicID:   req.GetGroupId(),
		ChannelPublicID: req.GetChannelId(),
		MessagePublicID: req.GetMessageId(),
	}); err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &chatv1.DeleteChannelMessageResponse{Ok: true}, nil
}

func (s *server) ToggleMessageReaction(ctx context.Context, req *chatv1.ToggleMessageReactionRequest) (*chatv1.ToggleMessageReactionResponse, error) {
	reactions, err := s.svc.ToggleMessageReaction(ctx, &models.ToggleMessageReactionInput{
		AccessToken:     req.GetAccessToken(),
		GroupPublicID:   req.GetGroupId(),
		ChannelPublicID: req.GetChannelId(),
		MessagePublicID: req.GetMessageId(),
		Emoji:           req.GetEmoji(),
	})
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &chatv1.ToggleMessageReactionResponse{Reactions: chatPb.ReactionSummaries(reactions)}, nil
}

func (s *server) SetChannelMessagePinned(ctx context.Context, req *chatv1.SetChannelMessagePinnedRequest) (*chatv1.SetChannelMessagePinnedResponse, error) {
	msg, err := s.svc.SetChannelMessagePinned(ctx, &models.SetChannelMessagePinnedInput{
		AccessToken:     req.GetAccessToken(),
		GroupPublicID:   req.GetGroupId(),
		ChannelPublicID: req.GetChannelId(),
		MessagePublicID: req.GetMessageId(),
		Pinned:          req.GetPinned(),
	})
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &chatv1.SetChannelMessagePinnedResponse{Message: chatPb.MessageSummary(msg)}, nil
}

func (s *server) ListPinnedChannelMessages(ctx context.Context, req *chatv1.ListPinnedChannelMessagesRequest) (*chatv1.ListPinnedChannelMessagesResponse, error) {
	messages, err := s.svc.ListPinnedChannelMessages(ctx, &models.ListPinnedChannelMessagesInput{
		AccessToken:     req.GetAccessToken(),
		GroupPublicID:   req.GetGroupId(),
		ChannelPublicID: req.GetChannelId(),
	})
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &chatv1.ListPinnedChannelMessagesResponse{Messages: chatPb.MessageSummaries(messages)}, nil
}

func (s *server) ResolveMessageChannelSeq(ctx context.Context, req *chatv1.ResolveMessageChannelSeqRequest) (*chatv1.ResolveMessageChannelSeqResponse, error) {
	out, err := s.svc.ResolveMessageChannelSeq(ctx, &models.ResolveMessageChannelSeqInput{
		AccessToken:     req.GetAccessToken(),
		GroupPublicID:   req.GetGroupId(),
		ChannelPublicID: req.GetChannelId(),
		MessagePublicID: req.GetMessageId(),
	})
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &chatv1.ResolveMessageChannelSeqResponse{
		MessageId:  out.MessagePublicID,
		GroupId:    out.GroupPublicID,
		ChannelId:  out.ChannelPublicID,
		ChannelSeq: out.ChannelSeq,
	}, nil
}

func (s *server) UploadChannelAttachment(ctx context.Context, req *chatv1.UploadChannelAttachmentRequest) (*chatv1.UploadChannelAttachmentResponse, error) {
	att, err := s.svc.UploadChannelAttachment(ctx, &models.UploadChannelAttachmentInput{
		AccessToken:     req.GetAccessToken(),
		GroupPublicID:   req.GetGroupId(),
		ChannelPublicID: req.GetChannelId(),
		Filename:        req.GetFilename(),
		ContentType:     req.GetContentType(),
		Data:            req.GetData(),
	})
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &chatv1.UploadChannelAttachmentResponse{Attachment: chatPb.MessageAttachment(att)}, nil
}
