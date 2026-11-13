package grpc

import (
	"context"

	authservice "github.com/wichat/wichat/backend/ms-user/internal/auth/service"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	authPb "github.com/wichat/wichat/backend/ms-user/pkg/auth-pb"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	responsegrpc "github.com/wichat/wichat/backend/wi-shared/response/grpc"
)

// Server implements userv1.AuthServiceServer.
type Server interface {
	userv1.AuthServiceServer
}

type server struct {
	userv1.UnimplementedAuthServiceServer
	auth authservice.Service
}

// NewServer binds the auth service to the Auth gRPC surface.
func NewServer(auth authservice.Service) Server {
	return &server{
		auth: auth,
	}
}

func (s *server) Setup(ctx context.Context, req *userv1.SetupRequest) (*userv1.SetupResponse, error) {
	in := models.RegisterInput{
		Email:       req.GetEmail(),
		Password:    req.GetPassword(),
		DisplayName: req.GetDisplayName(),
		ClientMeta:  authPb.Meta(req.GetUserAgent(), req.GetClientIp()),
	}

	out, err := s.auth.Register(ctx, in)
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	sess, _ := authPb.PasswordAuth(out)
	return &userv1.SetupResponse{Session: sess}, nil
}

func (s *server) Register(ctx context.Context, req *userv1.RegisterRequest) (*userv1.RegisterResponse, error) {
	in := models.RegisterInput{
		Email:       req.GetEmail(),
		Username:    req.GetUsername(),
		Password:    req.GetPassword(),
		DisplayName: req.GetDisplayName(),
		BirthDate:   req.GetBirthDate(),
		Locale:      req.GetLocale(),
		ClientMeta:  authPb.Meta(req.GetUserAgent(), req.GetClientIp()),
	}

	out, err := s.auth.Register(ctx, in)
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	sess, ch := authPb.PasswordAuth(out)
	return &userv1.RegisterResponse{Session: sess, Challenge: ch}, nil
}

func (s *server) Login(ctx context.Context, req *userv1.LoginRequest) (*userv1.LoginResponse, error) {
	in := models.LoginInput{
		Login:      req.GetLogin(),
		Password:   req.GetPassword(),
		Locale:     req.GetLocale(),
		ClientMeta: authPb.Meta(req.GetUserAgent(), req.GetClientIp()),
	}

	out, err := s.auth.Login(ctx, in)
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	sess, ch := authPb.PasswordAuth(out)
	return &userv1.LoginResponse{Session: sess, Challenge: ch}, nil
}

func (s *server) VerifyRegister(ctx context.Context, req *userv1.VerifyRegisterRequest) (*userv1.VerifyRegisterResponse, error) {
	out, err := s.auth.VerifyRegister(ctx, req.GetChallengeId(), req.GetCode())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.VerifyRegisterResponse{Session: authPb.Session(out)}, nil
}

func (s *server) VerifyLogin(ctx context.Context, req *userv1.VerifyLoginRequest) (*userv1.VerifyLoginResponse, error) {
	out, err := s.auth.VerifyLogin(ctx, req.GetChallengeId(), req.GetCode())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.VerifyLoginResponse{Session: authPb.Session(out)}, nil
}

func (s *server) ResendRegisterCode(ctx context.Context, req *userv1.ResendRegisterCodeRequest) (*userv1.ResendRegisterCodeResponse, error) {
	cooldown, err := s.auth.ResendRegisterCode(ctx, req.GetChallengeId())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.ResendRegisterCodeResponse{ResendCooldownSeconds: cooldown}, nil
}

func (s *server) ResendLoginCode(ctx context.Context, req *userv1.ResendLoginCodeRequest) (*userv1.ResendLoginCodeResponse, error) {
	cooldown, err := s.auth.ResendLoginCode(ctx, req.GetChallengeId())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.ResendLoginCodeResponse{ResendCooldownSeconds: cooldown}, nil
}

func (s *server) Refresh(ctx context.Context, req *userv1.RefreshRequest) (*userv1.RefreshResponse, error) {
	meta := authPb.Meta(req.GetUserAgent(), req.GetClientIp())

	out, err := s.auth.Refresh(ctx, req.GetRefreshToken(), meta)
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.RefreshResponse{Session: authPb.Session(out)}, nil
}

func (s *server) Logout(ctx context.Context, req *userv1.LogoutRequest) (*userv1.LogoutResponse, error) {
	if err := s.auth.Logout(ctx, req.GetRefreshToken()); err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.LogoutResponse{Ok: true}, nil
}

func (s *server) GoogleStart(ctx context.Context, _ *userv1.GoogleStartRequest) (*userv1.GoogleStartResponse, error) {
	authURL, state, verifier, err := s.auth.GoogleStart(ctx)
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.GoogleStartResponse{
		AuthorizationUrl: authURL,
		State:            state,
		CodeVerifier:     verifier,
	}, nil
}

func (s *server) GoogleFinish(ctx context.Context, req *userv1.GoogleFinishRequest) (*userv1.GoogleFinishResponse, error) {
	out, err := s.auth.GoogleFinish(ctx, req.GetCode(), req.GetCodeVerifier(), authPb.Meta(req.GetUserAgent(), req.GetClientIp()))
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.GoogleFinishResponse{Session: authPb.Session(out)}, nil
}

func (s *server) RequestPasswordReset(ctx context.Context, req *userv1.RequestPasswordResetRequest) (*userv1.RequestPasswordResetResponse, error) {
	secs, err := s.auth.RequestPasswordReset(ctx, req.GetEmail(), req.GetLocale())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.RequestPasswordResetResponse{
		Ok:                    true,
		ResendCooldownSeconds: int32(secs), //nolint:gosec // G115 — service returns bounded cooldown seconds
	}, nil
}

func (s *server) VerifyPasswordReset(ctx context.Context, req *userv1.VerifyPasswordResetRequest) (*userv1.VerifyPasswordResetResponse, error) {
	token, err := s.auth.VerifyPasswordReset(ctx, req.GetEmail(), req.GetCode())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.VerifyPasswordResetResponse{ResetToken: token}, nil
}

func (s *server) CompletePasswordReset(ctx context.Context, req *userv1.CompletePasswordResetRequest) (*userv1.CompletePasswordResetResponse, error) {
	if err := s.auth.CompletePasswordReset(ctx, req.GetResetToken(), req.GetPassword()); err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.CompletePasswordResetResponse{Ok: true}, nil
}

func (s *server) Me(ctx context.Context, req *userv1.MeRequest) (*userv1.MeResponse, error) {
	user, err := s.auth.Me(ctx, req.GetAccessToken())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.MeResponse{User: authPb.User(*user)}, nil
}
