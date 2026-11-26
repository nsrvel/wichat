package grpc

import (
	"context"
	"math"
	"time"

	mentionservice "github.com/wichat/wichat/backend/ms-chat/internal/mention/service"
	"github.com/wichat/wichat/backend/ms-chat/internal/models"
	"github.com/wichat/wichat/backend/ms-chat/internal/search/service"
	chatv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/chat/v1"
	responsegrpc "github.com/wichat/wichat/backend/wi-shared/response/grpc"
)

// Server is the search gRPC delivery.
type Server interface {
	chatv1.SearchServiceServer
}

type server struct {
	chatv1.UnimplementedSearchServiceServer
	query    service.QueryService
	mentions mentionservice.Service
}

// NewServer binds search and mention inbox use cases to gRPC.
func NewServer(query service.QueryService, mentions mentionservice.Service) Server {
	return &server{query: query, mentions: mentions}
}

func (s *server) SearchMessages(ctx context.Context, req *chatv1.SearchMessagesRequest) (*chatv1.SearchMessagesResponse, error) {
	// Bind
	limit := int(req.GetLimit())
	// Service
	hits, err := s.query.SearchMessages(ctx, req.GetAccessToken(), req.GetGroupId(), req.GetQuery(), limit)
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}
	// Response
	out := make([]*chatv1.MessageSearchHit, 0, len(hits))
	for _, h := range hits {
		out = append(out, &chatv1.MessageSearchHit{
			MessageId:         h.MessageID,
			GroupId:           h.GroupID,
			ChannelId:         h.ChannelID,
			BodyPreview:       h.BodyPreview,
			AuthorDisplayName: h.AuthorDisplayName,
			CreatedAt:         h.CreatedAt,
		})
	}
	return &chatv1.SearchMessagesResponse{Hits: out}, nil
}

func (s *server) SearchFiles(ctx context.Context, req *chatv1.SearchFilesRequest) (*chatv1.SearchFilesResponse, error) {
	// Bind
	limit := int(req.GetLimit())
	// Service
	hits, err := s.query.SearchFiles(ctx, req.GetAccessToken(), req.GetGroupId(), req.GetQuery(), limit, req.GetImagesOnly())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}
	// Response
	out := make([]*chatv1.FileSearchHit, 0, len(hits))
	for _, h := range hits {
		out = append(out, &chatv1.FileSearchHit{
			AttachmentId: h.AttachmentID,
			MessageId:    h.MessageID,
			GroupId:      h.GroupID,
			ChannelId:    h.ChannelID,
			Filename:     h.Filename,
			ContentType:  h.ContentType,
			CreatedAt:    h.CreatedAt,
		})
	}
	return &chatv1.SearchFilesResponse{Hits: out}, nil
}

func (s *server) ListGroupMentions(ctx context.Context, req *chatv1.ListGroupMentionsRequest) (*chatv1.ListGroupMentionsResponse, error) {
	limit := int(req.GetLimit())
	hits, nextCursor, err := s.mentions.ListGroupMentions(ctx, &models.ListGroupMentionsInput{
		AccessToken:   req.GetAccessToken(),
		GroupPublicID: req.GetGroupId(),
		Limit:         limit,
		Cursor:        req.GetCursor(),
		Kind:          req.GetKind(),
	})
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}
	out := make([]*chatv1.GroupMentionHit, 0, len(hits))
	for _, h := range hits {
		out = append(out, &chatv1.GroupMentionHit{
			MessageId:         h.MessageID,
			ChannelId:         h.ChannelID,
			ChannelName:       h.ChannelName,
			ChannelType:       h.ChannelType,
			AuthorDisplayName: h.AuthorDisplayName,
			BodyPreview:       h.BodyPreview,
			CreatedAt:         h.CreatedAt.UTC().Format(time.RFC3339Nano),
			IsUnread:          h.IsUnread,
		})
	}
	return &chatv1.ListGroupMentionsResponse{Hits: out, NextCursor: nextCursor}, nil
}

func (s *server) MarkGroupMentionsRead(ctx context.Context, req *chatv1.MarkGroupMentionsReadRequest) (*chatv1.MarkGroupMentionsReadResponse, error) {
	err := s.mentions.MarkGroupMentionsRead(ctx, &models.MarkGroupMentionsReadInput{
		AccessToken:      req.GetAccessToken(),
		GroupPublicID:    req.GetGroupId(),
		MessagePublicIDs: req.GetMessageIds(),
	})
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}
	return &chatv1.MarkGroupMentionsReadResponse{}, nil
}

func (s *server) CountGroupMentionsUnread(ctx context.Context, req *chatv1.CountGroupMentionsUnreadRequest) (*chatv1.CountGroupMentionsUnreadResponse, error) {
	count, err := s.mentions.CountGroupMentionsUnread(ctx, req.GetAccessToken(), req.GetGroupId())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}
	var out int32
	if count > 0 {
		if count > math.MaxInt32 {
			out = math.MaxInt32
		} else {
			out = int32(count) //nolint:gosec // clamped to MaxInt32
		}
	}
	return &chatv1.CountGroupMentionsUnreadResponse{Count: out}, nil
}
