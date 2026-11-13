package grpc

import (
	"context"

	"github.com/wichat/wichat/backend/ms-user/internal/models"
	usersvc "github.com/wichat/wichat/backend/ms-user/internal/user/service"
	authPb "github.com/wichat/wichat/backend/ms-user/pkg/auth-pb"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	responsegrpc "github.com/wichat/wichat/backend/wi-shared/response/grpc"
)

// Server implements userv1.UserServiceServer.
type Server struct {
	userv1.UnimplementedUserServiceServer
	svc usersvc.Service
}

// NewServer creates a new UserServiceServer.
func NewServer(svc usersvc.Service) *Server {
	return &Server{svc: svc}
}

func (s *Server) SetupStatus(ctx context.Context, _ *userv1.SetupStatusRequest) (*userv1.SetupStatusResponse, error) {
	out, err := s.svc.SetupStatus(ctx)
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.SetupStatusResponse{
		RegistrationMode: out.RegistrationMode,
		UsernameEnabled:  out.UsernameEnabled,
	}, nil
}

func (s *Server) CompleteInstanceSetup(ctx context.Context, req *userv1.CompleteInstanceSetupRequest) (*userv1.CompleteInstanceSetupResponse, error) {
	in := models.CompleteInstanceSetupInput{
		Name:             req.GetName(),
		LogoURL:          req.GetLogoUrl(),
		UsernameEnabled:  req.GetUsernameEnabled(),
		RegistrationMode: req.GetRegistrationMode(),
	}

	if err := s.svc.CompleteInstanceSetup(ctx, req.GetAccessToken(), in); err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.CompleteInstanceSetupResponse{Ok: true}, nil
}

func (s *Server) SearchUsers(ctx context.Context, req *userv1.SearchUsersRequest) (*userv1.SearchUsersResponse, error) {
	users, err := s.svc.SearchUsers(ctx, req.GetAccessToken(), req.GetQuery(), int(req.GetLimit()), req.GetGroupId())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.SearchUsersResponse{Users: authPb.UserSearchHits(users)}, nil
}
