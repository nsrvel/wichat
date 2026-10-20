package grpc

import (
	"context"

	authservice "github.com/wichat/wichat/backend/ms-user/internal/auth/service"
	groupservice "github.com/wichat/wichat/backend/ms-user/internal/group/service"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	userservice "github.com/wichat/wichat/backend/ms-user/internal/user/service"
	authPb "github.com/wichat/wichat/backend/ms-user/pkg/converter"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	responsegrpc "github.com/wichat/wichat/backend/wi-shared/response/grpc"
)

// Server is the auth gRPC delivery (proto service + orchestration).
type Server interface {
	userv1.AuthServiceServer
}

type server struct {
	userv1.UnimplementedAuthServiceServer
	auth  authservice.Service
	user  userservice.Service
	group groupservice.Service
}

// NewServer binds auth, user, and group services to the Auth gRPC surface.
func NewServer(auth authservice.Service, user userservice.Service, group groupservice.Service) Server {
	return &server{auth: auth, user: user, group: group}
}

func (s *server) SetupStatus(ctx context.Context, _ *userv1.SetupStatusRequest) (*userv1.SetupStatusResponse, error) {
	// Service
	out, err := s.user.SetupStatus(ctx)
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &userv1.SetupStatusResponse{
		RegistrationMode: out.RegistrationMode,
		GuestEnabled:     out.GuestEnabled,
		UsernameEnabled:  out.UsernameEnabled,
	}, nil
}

func (s *server) CompleteInstanceSetup(ctx context.Context, req *userv1.CompleteInstanceSetupRequest) (*userv1.CompleteInstanceSetupResponse, error) {
	// Bind
	in := models.CompleteInstanceSetupInput{
		Name:             req.GetName(),
		LogoURL:          req.GetLogoUrl(),
		UsernameEnabled:  req.GetUsernameEnabled(),
		RegistrationMode: req.GetRegistrationMode(),
		GuestEnabled:     req.GetGuestEnabled(),
	}

	// Service
	if err := s.user.CompleteInstanceSetup(ctx, req.GetAccessToken(), in); err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &userv1.CompleteInstanceSetupResponse{Ok: true}, nil
}

func (s *server) Setup(ctx context.Context, req *userv1.SetupRequest) (*userv1.SetupResponse, error) {
	// Bind
	in := models.RegisterInput{
		Email:       req.GetEmail(),
		Password:    req.GetPassword(),
		DisplayName: req.GetDisplayName(),
		ClientMeta:  authPb.Meta(req.GetUserAgent(), req.GetClientIp()),
	}

	// Service
	out, err := s.auth.Register(ctx, in)
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	sess, _ := authPb.PasswordAuth(out)
	return &userv1.SetupResponse{Session: sess}, nil
}

func (s *server) Register(ctx context.Context, req *userv1.RegisterRequest) (*userv1.RegisterResponse, error) {
	// Bind
	in := models.RegisterInput{
		Email:       req.GetEmail(),
		Password:    req.GetPassword(),
		DisplayName: req.GetDisplayName(),
		BirthDate:   req.GetBirthDate(),
		Locale:      req.GetLocale(),
		ClientMeta:  authPb.Meta(req.GetUserAgent(), req.GetClientIp()),
	}

	// Service
	out, err := s.auth.Register(ctx, in)
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	sess, ch := authPb.PasswordAuth(out)
	return &userv1.RegisterResponse{Session: sess, Challenge: ch}, nil
}

func (s *server) RegisterGuest(ctx context.Context, req *userv1.RegisterGuestRequest) (*userv1.RegisterGuestResponse, error) {
	// Bind
	in := models.GuestRegisterInput{
		DisplayName: req.GetDisplayName(),
		BirthDate:   req.GetBirthDate(),
		ClientMeta:  authPb.Meta(req.GetUserAgent(), req.GetClientIp()),
	}

	// Service
	out, err := s.auth.RegisterGuest(ctx, in)
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &userv1.RegisterGuestResponse{Session: authPb.Session(out)}, nil
}

func (s *server) Login(ctx context.Context, req *userv1.LoginRequest) (*userv1.LoginResponse, error) {
	// Bind
	in := models.LoginInput{
		Login:      req.GetLogin(),
		Password:   req.GetPassword(),
		Locale:     req.GetLocale(),
		ClientMeta: authPb.Meta(req.GetUserAgent(), req.GetClientIp()),
	}

	// Service
	out, err := s.auth.Login(ctx, in)
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
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
	// Bind
	meta := authPb.Meta(req.GetUserAgent(), req.GetClientIp())

	// Service
	out, err := s.auth.Refresh(ctx, req.GetRefreshToken(), meta)
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &userv1.RefreshResponse{Session: authPb.Session(out)}, nil
}

func (s *server) Logout(ctx context.Context, req *userv1.LogoutRequest) (*userv1.LogoutResponse, error) {
	// Service
	if err := s.auth.Logout(ctx, req.GetRefreshToken()); err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &userv1.LogoutResponse{Ok: true}, nil
}

func (s *server) GoogleStart(ctx context.Context, _ *userv1.GoogleStartRequest) (*userv1.GoogleStartResponse, error) {
	// Service
	authURL, state, verifier, err := s.auth.GoogleStart(ctx)
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
	out, err := s.auth.GoogleFinish(ctx, req.GetCode(), req.GetCodeVerifier(), authPb.Meta(req.GetUserAgent(), req.GetClientIp()))
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &userv1.GoogleFinishResponse{Session: authPb.Session(out)}, nil
}

func (s *server) RequestPasswordReset(ctx context.Context, req *userv1.RequestPasswordResetRequest) (*userv1.RequestPasswordResetResponse, error) {
	// Service
	secs, err := s.auth.RequestPasswordReset(ctx, req.GetEmail(), req.GetLocale())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &userv1.RequestPasswordResetResponse{
		Ok:                    true,
		ResendCooldownSeconds: int32(secs), //nolint:gosec // G115 — service returns bounded cooldown seconds
	}, nil
}

func (s *server) VerifyPasswordReset(ctx context.Context, req *userv1.VerifyPasswordResetRequest) (*userv1.VerifyPasswordResetResponse, error) {
	// Service
	token, err := s.auth.VerifyPasswordReset(ctx, req.GetEmail(), req.GetCode())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &userv1.VerifyPasswordResetResponse{ResetToken: token}, nil
}

func (s *server) CompletePasswordReset(ctx context.Context, req *userv1.CompletePasswordResetRequest) (*userv1.CompletePasswordResetResponse, error) {
	// Service
	if err := s.auth.CompletePasswordReset(ctx, req.GetResetToken(), req.GetPassword()); err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &userv1.CompletePasswordResetResponse{Ok: true}, nil
}

func (s *server) Me(ctx context.Context, req *userv1.MeRequest) (*userv1.MeResponse, error) {
	// Service
	user, err := s.auth.Me(ctx, req.GetAccessToken())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &userv1.MeResponse{User: authPb.User(*user)}, nil
}

func (s *server) ListMyGroups(ctx context.Context, req *userv1.ListMyGroupsRequest) (*userv1.ListMyGroupsResponse, error) {
	groups, err := s.group.ListMyGroups(ctx, req.GetAccessToken())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.ListMyGroupsResponse{Groups: authPb.GroupSummaries(groups)}, nil
}

func (s *server) ListJoinableGroups(ctx context.Context, req *userv1.ListJoinableGroupsRequest) (*userv1.ListJoinableGroupsResponse, error) {
	groups, err := s.group.ListJoinableGroups(ctx, req.GetAccessToken())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.ListJoinableGroupsResponse{Groups: authPb.GroupSummaries(groups)}, nil
}

func (s *server) CreateGroup(ctx context.Context, req *userv1.CreateGroupRequest) (*userv1.CreateGroupResponse, error) {
	group, err := s.group.CreateGroup(ctx, req.GetAccessToken(), req.GetName())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.CreateGroupResponse{Group: authPb.GroupSummary(*group)}, nil
}

func (s *server) JoinGroup(ctx context.Context, req *userv1.JoinGroupRequest) (*userv1.JoinGroupResponse, error) {
	if err := s.group.JoinGroup(ctx, req.GetAccessToken(), req.GetGroupId()); err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.JoinGroupResponse{Ok: true}, nil
}
