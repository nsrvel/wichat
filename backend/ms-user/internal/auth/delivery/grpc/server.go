package grpc

import (
	"context"

	"github.com/wichat/wichat/backend/ms-user/internal/auth/service"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	authPb "github.com/wichat/wichat/backend/ms-user/pkg/auth-pb"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	responsegrpc "github.com/wichat/wichat/backend/wi-shared/response/grpc"
)

// Server is the auth gRPC delivery (proto service + orchestration).
type Server interface {
	userv1.AuthServiceServer
}

type server struct {
	userv1.UnimplementedAuthServiceServer
	svc service.Service
}

// NewServer binds the auth service to gRPC.
func NewServer(svc service.Service) Server {
	return &server{svc: svc}
}

func (s *server) SetupStatus(ctx context.Context, _ *userv1.SetupStatusRequest) (*userv1.SetupStatusResponse, error) {
	// Service
	needsSetup, err := s.svc.SetupStatus(ctx)
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &userv1.SetupStatusResponse{NeedsSetup: needsSetup}, nil
}

func (s *server) Setup(ctx context.Context, req *userv1.SetupRequest) (*userv1.SetupResponse, error) {
	// Bind
	in := models.SetupInput{
		Email:       req.GetEmail(),
		Password:    req.GetPassword(),
		DisplayName: req.GetDisplayName(),
		ClientMeta:  authPb.Meta(req.GetUserAgent(), req.GetClientIp()),
	}

	// Service
	out, err := s.svc.Setup(ctx, in)
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &userv1.SetupResponse{Session: authPb.Session(out)}, nil
}

func (s *server) Login(ctx context.Context, req *userv1.LoginRequest) (*userv1.LoginResponse, error) {
	// Bind
	in := models.LoginInput{
		Email:      req.GetEmail(),
		Password:   req.GetPassword(),
		ClientMeta: authPb.Meta(req.GetUserAgent(), req.GetClientIp()),
	}

	// Service
	out, err := s.svc.Login(ctx, in)
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &userv1.LoginResponse{Session: authPb.Session(out)}, nil
}

func (s *server) Refresh(ctx context.Context, req *userv1.RefreshRequest) (*userv1.RefreshResponse, error) {
	// Bind
	meta := authPb.Meta(req.GetUserAgent(), req.GetClientIp())

	// Service
	out, err := s.svc.Refresh(ctx, req.GetRefreshToken(), meta)
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &userv1.RefreshResponse{Session: authPb.Session(out)}, nil
}

func (s *server) Logout(ctx context.Context, req *userv1.LogoutRequest) (*userv1.LogoutResponse, error) {
	// Service
	if err := s.svc.Logout(ctx, req.GetRefreshToken()); err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &userv1.LogoutResponse{Ok: true}, nil
}

func (s *server) GoogleStart(ctx context.Context, _ *userv1.GoogleStartRequest) (*userv1.GoogleStartResponse, error) {
	// Service
	authURL, state, verifier, err := s.svc.GoogleStart(ctx)
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &userv1.GoogleStartResponse{
		AuthorizationUrl: authURL,
		State:            state,
		CodeVerifier:     verifier,
	}, nil
}

func (s *server) GoogleFinish(ctx context.Context, req *userv1.GoogleFinishRequest) (*userv1.GoogleFinishResponse, error) {
	// Service
	out, err := s.svc.GoogleFinish(ctx, req.GetCode(), req.GetCodeVerifier(), authPb.Meta(req.GetUserAgent(), req.GetClientIp()))
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &userv1.GoogleFinishResponse{Session: authPb.Session(out)}, nil
}

func (s *server) Me(ctx context.Context, req *userv1.MeRequest) (*userv1.MeResponse, error) {
	// Service
	user, err := s.svc.Me(ctx, req.GetAccessToken())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &userv1.MeResponse{User: authPb.User(*user)}, nil
}
