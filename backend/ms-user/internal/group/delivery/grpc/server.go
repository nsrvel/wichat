package grpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"time"

	groupactivitysvc "github.com/wichat/wichat/backend/ms-user/internal/group-activity-log/service"
	groupsvc "github.com/wichat/wichat/backend/ms-user/internal/group/service"
	authPb "github.com/wichat/wichat/backend/ms-user/pkg/auth-pb"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	responsegrpc "github.com/wichat/wichat/backend/wi-shared/response/grpc"
)

// Server implements userv1.GroupServiceServer.
type Server struct {
	userv1.UnimplementedGroupServiceServer
	svc           groupsvc.Service
	groupActivity groupactivitysvc.Service
}

// NewServer creates a new GroupServiceServer.
func NewServer(svc groupsvc.Service) *Server {
	return &Server{svc: svc}
}

// WithGroupActivityLog attaches the group activity log service for ListGroupActivityLogs.
func (s *Server) WithGroupActivityLog(svc groupactivitysvc.Service) *Server {
	s.groupActivity = svc
	return s
}

func (s *Server) ListMyGroups(ctx context.Context, req *userv1.ListMyGroupsRequest) (*userv1.ListMyGroupsResponse, error) {
	groups, err := s.svc.ListMyGroups(ctx, req.GetAccessToken())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.ListMyGroupsResponse{Groups: authPb.GroupSummaries(groups)}, nil
}

func (s *Server) CreateGroup(ctx context.Context, req *userv1.CreateGroupRequest) (*userv1.CreateGroupResponse, error) {
	group, err := s.svc.CreateGroup(ctx, req.GetAccessToken(), req.GetName())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.CreateGroupResponse{Group: authPb.GroupSummary(*group)}, nil
}

func (s *Server) ReorderMyGroups(ctx context.Context, req *userv1.ReorderMyGroupsRequest) (*userv1.ReorderMyGroupsResponse, error) {
	if err := s.svc.ReorderMyGroups(ctx, req.GetAccessToken(), req.GetGroupIds()); err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.ReorderMyGroupsResponse{Ok: true}, nil
}

func (s *Server) DeleteGroup(ctx context.Context, req *userv1.DeleteGroupRequest) (*userv1.DeleteGroupResponse, error) {
	if err := s.svc.DeleteGroup(ctx, req.GetAccessToken(), req.GetGroupId()); err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.DeleteGroupResponse{Ok: true}, nil
}

func (s *Server) UploadGroupIcon(ctx context.Context, req *userv1.UploadGroupIconRequest) (*userv1.UploadGroupIconResponse, error) {
	group, err := s.svc.UploadGroupIcon(
		ctx,
		req.GetAccessToken(),
		req.GetGroupId(),
		req.GetContentType(),
		bytes.NewReader(req.GetData()),
		int64(len(req.GetData())),
	)
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.UploadGroupIconResponse{Group: authPb.GroupSummary(*group)}, nil
}

func (s *Server) UpdateGroupProfile(ctx context.Context, req *userv1.UpdateGroupProfileRequest) (*userv1.UpdateGroupProfileResponse, error) {
	group, err := s.svc.UpdateGroupProfile(
		ctx,
		req.GetAccessToken(),
		req.GetGroupId(),
		req.GetName(),
		req.GetDescription(),
		req.GetIsPrivate(),
		req.GetBannerColor(),
		req.GetClearLogo(),
		req.GetClearBannerImage(),
	)
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.UpdateGroupProfileResponse{Group: authPb.GroupSummary(*group)}, nil
}

func (s *Server) UploadGroupBanner(ctx context.Context, req *userv1.UploadGroupBannerRequest) (*userv1.UploadGroupBannerResponse, error) {
	group, err := s.svc.UploadGroupBanner(
		ctx,
		req.GetAccessToken(),
		req.GetGroupId(),
		req.GetContentType(),
		bytes.NewReader(req.GetData()),
		int64(len(req.GetData())),
	)
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.UploadGroupBannerResponse{Group: authPb.GroupSummary(*group)}, nil
}

func (s *Server) ListGroupActivityLogs(ctx context.Context, req *userv1.ListGroupActivityLogsRequest) (*userv1.ListGroupActivityLogsResponse, error) {
	if s.groupActivity == nil {
		return nil, responsegrpc.ToStatus(exception.Internal(errors.New("group activity log service not configured")))
	}
	page, err := s.groupActivity.ListGroupActivityLogs(
		ctx,
		req.GetAccessToken(),
		req.GetGroupId(),
		req.GetActorId(),
		req.GetCategory(),
		req.GetAction(),
		int(req.GetLimit()),
		req.GetCursor(),
	)
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	out := make([]*userv1.GroupActivityLog, 0, len(page.Logs))
	for _, entry := range page.Logs {
		changesJSON := "{}"
		if len(entry.Changes) > 0 {
			if raw, err := json.Marshal(entry.Changes); err == nil {
				changesJSON = string(raw)
			}
		}
		metadataJSON := "{}"
		if len(entry.Metadata) > 0 {
			if raw, err := json.Marshal(entry.Metadata); err == nil {
				metadataJSON = string(raw)
			}
		}
		out = append(out, &userv1.GroupActivityLog{
			Id:               entry.ID,
			GroupId:          entry.GroupID,
			ActorId:          entry.ActorID,
			ActorUsername:    entry.ActorUsername,
			ActorDisplayName: entry.ActorDisplayName,
			ActorAvatarUrl:   entry.ActorAvatarURL,
			Category:         entry.Category,
			Action:           entry.Action,
			TargetType:       entry.TargetType,
			TargetId:         entry.TargetID,
			TargetName:       entry.TargetName,
			Summary:          entry.Summary,
			ChangesJson:      changesJSON,
			MetadataJson:     metadataJSON,
			CreatedAt:        entry.CreatedAt.UTC().Format(time.RFC3339Nano),
		})
	}
	return &userv1.ListGroupActivityLogsResponse{
		Logs:       out,
		NextCursor: page.NextCursor,
	}, nil
}
