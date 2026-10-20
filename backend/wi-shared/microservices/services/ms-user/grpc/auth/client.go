// Package auth is the gRPC client for ms-user AuthService.
package auth

import (
	"context"

	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	"google.golang.org/grpc"
)

// Client calls ms-user AuthService with shared timeouts, retries, and circuit breaking.
type Client interface {
	SetupStatus(ctx context.Context, in *userv1.SetupStatusRequest) (*userv1.SetupStatusResponse, error)
	Setup(ctx context.Context, in *userv1.SetupRequest) (*userv1.SetupResponse, error)
	Register(ctx context.Context, in *userv1.RegisterRequest) (*userv1.RegisterResponse, error)
	RegisterGuest(ctx context.Context, in *userv1.RegisterGuestRequest) (*userv1.RegisterGuestResponse, error)
	Login(ctx context.Context, in *userv1.LoginRequest) (*userv1.LoginResponse, error)
	VerifyRegister(ctx context.Context, in *userv1.VerifyRegisterRequest) (*userv1.VerifyRegisterResponse, error)
	VerifyLogin(ctx context.Context, in *userv1.VerifyLoginRequest) (*userv1.VerifyLoginResponse, error)
	ResendRegisterCode(ctx context.Context, in *userv1.ResendRegisterCodeRequest) (*userv1.ResendRegisterCodeResponse, error)
	ResendLoginCode(ctx context.Context, in *userv1.ResendLoginCodeRequest) (*userv1.ResendLoginCodeResponse, error)
	Refresh(ctx context.Context, in *userv1.RefreshRequest) (*userv1.RefreshResponse, error)
	Logout(ctx context.Context, in *userv1.LogoutRequest) (*userv1.LogoutResponse, error)
	Me(ctx context.Context, in *userv1.MeRequest) (*userv1.MeResponse, error)
	CompleteInstanceSetup(ctx context.Context, in *userv1.CompleteInstanceSetupRequest) (*userv1.CompleteInstanceSetupResponse, error)
	GoogleStart(ctx context.Context, in *userv1.GoogleStartRequest) (*userv1.GoogleStartResponse, error)
	GoogleFinish(ctx context.Context, in *userv1.GoogleFinishRequest) (*userv1.GoogleFinishResponse, error)
	RequestPasswordReset(ctx context.Context, in *userv1.RequestPasswordResetRequest) (*userv1.RequestPasswordResetResponse, error)
	VerifyPasswordReset(ctx context.Context, in *userv1.VerifyPasswordResetRequest) (*userv1.VerifyPasswordResetResponse, error)
	CompletePasswordReset(ctx context.Context, in *userv1.CompletePasswordResetRequest) (*userv1.CompletePasswordResetResponse, error)
	ListMyGroups(ctx context.Context, in *userv1.ListMyGroupsRequest) (*userv1.ListMyGroupsResponse, error)
	ListJoinableGroups(ctx context.Context, in *userv1.ListJoinableGroupsRequest) (*userv1.ListJoinableGroupsResponse, error)
	CreateGroup(ctx context.Context, in *userv1.CreateGroupRequest) (*userv1.CreateGroupResponse, error)
	JoinGroup(ctx context.Context, in *userv1.JoinGroupRequest) (*userv1.JoinGroupResponse, error)
}

type client struct {
	stub userv1.AuthServiceClient
	log  logger.Logger
	none utils.CallConfig
	read utils.CallConfig
}

// New wraps an existing gRPC connection (tests or custom dial). Caller owns conn lifetime.
func New(conn *grpc.ClientConn, settings utils.Settings, log logger.Logger) Client {
	breaker := utils.NewBreaker(settings)
	none := utils.CallConfigFromSettings(settings, utils.RetryNone)
	none.Breaker = breaker
	read := utils.CallConfigFromSettings(settings, utils.RetryIdempotent)
	read.Breaker = breaker

	return &client{
		stub: userv1.NewAuthServiceClient(conn),
		log:  log,
		none: none,
		read: read,
	}
}

func (c *client) SetupStatus(ctx context.Context, in *userv1.SetupStatusRequest) (*userv1.SetupStatusResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "auth.AuthService/SetupStatus", func(callCtx context.Context) (*userv1.SetupStatusResponse, error) {
		return c.stub.SetupStatus(callCtx, in)
	})
}

func (c *client) Setup(ctx context.Context, in *userv1.SetupRequest) (*userv1.SetupResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "auth.AuthService/Setup", func(callCtx context.Context) (*userv1.SetupResponse, error) {
		return c.stub.Setup(callCtx, in)
	})
}

func (c *client) Register(ctx context.Context, in *userv1.RegisterRequest) (*userv1.RegisterResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "auth.AuthService/Register", func(callCtx context.Context) (*userv1.RegisterResponse, error) {
		return c.stub.Register(callCtx, in)
	})
}

func (c *client) RegisterGuest(ctx context.Context, in *userv1.RegisterGuestRequest) (*userv1.RegisterGuestResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "auth.AuthService/RegisterGuest", func(callCtx context.Context) (*userv1.RegisterGuestResponse, error) {
		return c.stub.RegisterGuest(callCtx, in)
	})
}

func (c *client) Login(ctx context.Context, in *userv1.LoginRequest) (*userv1.LoginResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "auth.AuthService/Login", func(callCtx context.Context) (*userv1.LoginResponse, error) {
		return c.stub.Login(callCtx, in)
	})
}

func (c *client) VerifyRegister(ctx context.Context, in *userv1.VerifyRegisterRequest) (*userv1.VerifyRegisterResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "auth.AuthService/VerifyRegister", func(callCtx context.Context) (*userv1.VerifyRegisterResponse, error) {
		return c.stub.VerifyRegister(callCtx, in)
	})
}

func (c *client) VerifyLogin(ctx context.Context, in *userv1.VerifyLoginRequest) (*userv1.VerifyLoginResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "auth.AuthService/VerifyLogin", func(callCtx context.Context) (*userv1.VerifyLoginResponse, error) {
		return c.stub.VerifyLogin(callCtx, in)
	})
}

func (c *client) ResendRegisterCode(ctx context.Context, in *userv1.ResendRegisterCodeRequest) (*userv1.ResendRegisterCodeResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "auth.AuthService/ResendRegisterCode", func(callCtx context.Context) (*userv1.ResendRegisterCodeResponse, error) {
		return c.stub.ResendRegisterCode(callCtx, in)
	})
}

func (c *client) ResendLoginCode(ctx context.Context, in *userv1.ResendLoginCodeRequest) (*userv1.ResendLoginCodeResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "auth.AuthService/ResendLoginCode", func(callCtx context.Context) (*userv1.ResendLoginCodeResponse, error) {
		return c.stub.ResendLoginCode(callCtx, in)
	})
}

func (c *client) Refresh(ctx context.Context, in *userv1.RefreshRequest) (*userv1.RefreshResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "auth.AuthService/Refresh", func(callCtx context.Context) (*userv1.RefreshResponse, error) {
		return c.stub.Refresh(callCtx, in)
	})
}

func (c *client) Logout(ctx context.Context, in *userv1.LogoutRequest) (*userv1.LogoutResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "auth.AuthService/Logout", func(callCtx context.Context) (*userv1.LogoutResponse, error) {
		return c.stub.Logout(callCtx, in)
	})
}

func (c *client) Me(ctx context.Context, in *userv1.MeRequest) (*userv1.MeResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "auth.AuthService/Me", func(callCtx context.Context) (*userv1.MeResponse, error) {
		return c.stub.Me(callCtx, in)
	})
}

func (c *client) CompleteInstanceSetup(ctx context.Context, in *userv1.CompleteInstanceSetupRequest) (*userv1.CompleteInstanceSetupResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "auth.AuthService/CompleteInstanceSetup", func(callCtx context.Context) (*userv1.CompleteInstanceSetupResponse, error) {
		return c.stub.CompleteInstanceSetup(callCtx, in)
	})
}

func (c *client) GoogleStart(ctx context.Context, in *userv1.GoogleStartRequest) (*userv1.GoogleStartResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "auth.AuthService/GoogleStart", func(callCtx context.Context) (*userv1.GoogleStartResponse, error) {
		return c.stub.GoogleStart(callCtx, in)
	})
}

func (c *client) GoogleFinish(ctx context.Context, in *userv1.GoogleFinishRequest) (*userv1.GoogleFinishResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "auth.AuthService/GoogleFinish", func(callCtx context.Context) (*userv1.GoogleFinishResponse, error) {
		return c.stub.GoogleFinish(callCtx, in)
	})
}

func (c *client) RequestPasswordReset(ctx context.Context, in *userv1.RequestPasswordResetRequest) (*userv1.RequestPasswordResetResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "auth.AuthService/RequestPasswordReset", func(callCtx context.Context) (*userv1.RequestPasswordResetResponse, error) {
		return c.stub.RequestPasswordReset(callCtx, in)
	})
}

func (c *client) VerifyPasswordReset(ctx context.Context, in *userv1.VerifyPasswordResetRequest) (*userv1.VerifyPasswordResetResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "auth.AuthService/VerifyPasswordReset", func(callCtx context.Context) (*userv1.VerifyPasswordResetResponse, error) {
		return c.stub.VerifyPasswordReset(callCtx, in)
	})
}

func (c *client) CompletePasswordReset(ctx context.Context, in *userv1.CompletePasswordResetRequest) (*userv1.CompletePasswordResetResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "auth.AuthService/CompletePasswordReset", func(callCtx context.Context) (*userv1.CompletePasswordResetResponse, error) {
		return c.stub.CompletePasswordReset(callCtx, in)
	})
}

func (c *client) ListMyGroups(ctx context.Context, in *userv1.ListMyGroupsRequest) (*userv1.ListMyGroupsResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "auth.AuthService/ListMyGroups", func(callCtx context.Context) (*userv1.ListMyGroupsResponse, error) {
		return c.stub.ListMyGroups(callCtx, in)
	})
}

func (c *client) ListJoinableGroups(ctx context.Context, in *userv1.ListJoinableGroupsRequest) (*userv1.ListJoinableGroupsResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "auth.AuthService/ListJoinableGroups", func(callCtx context.Context) (*userv1.ListJoinableGroupsResponse, error) {
		return c.stub.ListJoinableGroups(callCtx, in)
	})
}

func (c *client) CreateGroup(ctx context.Context, in *userv1.CreateGroupRequest) (*userv1.CreateGroupResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "auth.AuthService/CreateGroup", func(callCtx context.Context) (*userv1.CreateGroupResponse, error) {
		return c.stub.CreateGroup(callCtx, in)
	})
}

func (c *client) JoinGroup(ctx context.Context, in *userv1.JoinGroupRequest) (*userv1.JoinGroupResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "auth.AuthService/JoinGroup", func(callCtx context.Context) (*userv1.JoinGroupResponse, error) {
		return c.stub.JoinGroup(callCtx, in)
	})
}
