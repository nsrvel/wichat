// Package user is the gRPC client for ms-user user-domain RPCs on UserService.
package user

import (
	"context"

	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

// Client calls user-domain UserService RPCs with shared timeouts, retries, and circuit breaking.
type Client interface {
	SetupStatus(ctx context.Context, in *userv1.SetupStatusRequest) (*userv1.SetupStatusResponse, error)
	CompleteInstanceSetup(ctx context.Context, in *userv1.CompleteInstanceSetupRequest) (*userv1.CompleteInstanceSetupResponse, error)
	SearchUsers(ctx context.Context, in *userv1.SearchUsersRequest) (*userv1.SearchUsersResponse, error)
	CheckUsername(ctx context.Context, in *userv1.CheckUsernameRequest) (*userv1.CheckUsernameResponse, error)
	UploadUserAvatar(ctx context.Context, in *userv1.UploadUserAvatarRequest) (*userv1.UploadUserAvatarResponse, error)
	UpdateUserProfile(ctx context.Context, in *userv1.UpdateUserProfileRequest) (*userv1.UpdateUserProfileResponse, error)
	UpdateUserStatusNote(ctx context.Context, in *userv1.UpdateUserStatusNoteRequest) (*userv1.UpdateUserStatusNoteResponse, error)
	UpdateUserPreferences(ctx context.Context, in *userv1.UpdateUserPreferencesRequest) (*userv1.UpdateUserPreferencesResponse, error)
	ListUserActivityLogs(ctx context.Context, in *userv1.ListUserActivityLogsRequest) (*userv1.ListUserActivityLogsResponse, error)
	UpdateUsername(ctx context.Context, in *userv1.UpdateUsernameRequest) (*userv1.UpdateUsernameResponse, error)
	RequestEmailChange(ctx context.Context, in *userv1.RequestEmailChangeRequest) (*userv1.RequestEmailChangeResponse, error)
	VerifyEmailChange(ctx context.Context, in *userv1.VerifyEmailChangeRequest) (*userv1.VerifyEmailChangeResponse, error)
	CompleteEmailChange(ctx context.Context, in *userv1.CompleteEmailChangeRequest) (*userv1.CompleteEmailChangeResponse, error)
	UpdatePassword(ctx context.Context, in *userv1.UpdatePasswordRequest) (*userv1.UpdatePasswordResponse, error)
	GetMFAStatus(ctx context.Context, in *userv1.GetMFAStatusRequest) (*userv1.GetMFAStatusResponse, error)
	BeginTOTPSetup(ctx context.Context, in *userv1.BeginTOTPSetupRequest) (*userv1.BeginTOTPSetupResponse, error)
	ConfirmTOTPSetup(ctx context.Context, in *userv1.ConfirmTOTPSetupRequest) (*userv1.ConfirmTOTPSetupResponse, error)
	DisableTOTP(ctx context.Context, in *userv1.DisableTOTPRequest) (*userv1.DisableTOTPResponse, error)
	BeginWebAuthnRegister(ctx context.Context, in *userv1.BeginWebAuthnRegisterRequest) (*userv1.BeginWebAuthnRegisterResponse, error)
	FinishWebAuthnRegister(ctx context.Context, in *userv1.FinishWebAuthnRegisterRequest) (*userv1.FinishWebAuthnRegisterResponse, error)
	RenameWebAuthnCredential(ctx context.Context, in *userv1.RenameWebAuthnCredentialRequest) (*userv1.RenameWebAuthnCredentialResponse, error)
	DeleteWebAuthnCredential(ctx context.Context, in *userv1.DeleteWebAuthnCredentialRequest) (*userv1.DeleteWebAuthnCredentialResponse, error)
	RegenerateRecoveryCodes(ctx context.Context, in *userv1.RegenerateRecoveryCodesRequest) (*userv1.RegenerateRecoveryCodesResponse, error)
	DisableAccount(ctx context.Context, in *userv1.DisableAccountRequest) (*userv1.DisableAccountResponse, error)
	DeleteAccount(ctx context.Context, in *userv1.DeleteAccountRequest) (*userv1.DeleteAccountResponse, error)
	ListSessions(ctx context.Context, in *userv1.ListSessionsRequest) (*userv1.ListSessionsResponse, error)
	RevokeSession(ctx context.Context, in *userv1.RevokeSessionRequest) (*userv1.RevokeSessionResponse, error)
	RevokeOtherSessions(ctx context.Context, in *userv1.RevokeOtherSessionsRequest) (*userv1.RevokeOtherSessionsResponse, error)
}

type client struct {
	stub userv1.UserServiceClient
	log  logger.Logger
	none utils.CallConfig
	read utils.CallConfig
}

// New wraps an existing UserService client stub.
func New(stub userv1.UserServiceClient, settings utils.Settings, log logger.Logger) Client {
	breaker := utils.NewBreaker(settings)
	none := utils.CallConfigFromSettings(settings, utils.RetryNone)
	none.Breaker = breaker
	read := utils.CallConfigFromSettings(settings, utils.RetryIdempotent)
	read.Breaker = breaker

	return &client{
		stub: stub,
		log:  log,
		none: none,
		read: read,
	}
}

func (c *client) SetupStatus(ctx context.Context, in *userv1.SetupStatusRequest) (*userv1.SetupStatusResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "user.UserService/SetupStatus", func(callCtx context.Context) (*userv1.SetupStatusResponse, error) {
		return c.stub.SetupStatus(callCtx, in)
	})
}

func (c *client) CompleteInstanceSetup(ctx context.Context, in *userv1.CompleteInstanceSetupRequest) (*userv1.CompleteInstanceSetupResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "user.UserService/CompleteInstanceSetup", func(callCtx context.Context) (*userv1.CompleteInstanceSetupResponse, error) {
		return c.stub.CompleteInstanceSetup(callCtx, in)
	})
}

func (c *client) SearchUsers(ctx context.Context, in *userv1.SearchUsersRequest) (*userv1.SearchUsersResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "user.UserService/SearchUsers", func(callCtx context.Context) (*userv1.SearchUsersResponse, error) {
		return c.stub.SearchUsers(callCtx, in)
	})
}

func (c *client) CheckUsername(ctx context.Context, in *userv1.CheckUsernameRequest) (*userv1.CheckUsernameResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "user.UserService/CheckUsername", func(callCtx context.Context) (*userv1.CheckUsernameResponse, error) {
		return c.stub.CheckUsername(callCtx, in)
	})
}

func (c *client) UploadUserAvatar(ctx context.Context, in *userv1.UploadUserAvatarRequest) (*userv1.UploadUserAvatarResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "user.UserService/UploadUserAvatar", func(callCtx context.Context) (*userv1.UploadUserAvatarResponse, error) {
		return c.stub.UploadUserAvatar(callCtx, in)
	})
}

func (c *client) UpdateUserProfile(ctx context.Context, in *userv1.UpdateUserProfileRequest) (*userv1.UpdateUserProfileResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "user.UserService/UpdateUserProfile", func(callCtx context.Context) (*userv1.UpdateUserProfileResponse, error) {
		return c.stub.UpdateUserProfile(callCtx, in)
	})
}

func (c *client) UpdateUserStatusNote(ctx context.Context, in *userv1.UpdateUserStatusNoteRequest) (*userv1.UpdateUserStatusNoteResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "user.UserService/UpdateUserStatusNote", func(callCtx context.Context) (*userv1.UpdateUserStatusNoteResponse, error) {
		return c.stub.UpdateUserStatusNote(callCtx, in)
	})
}

func (c *client) UpdateUserPreferences(ctx context.Context, in *userv1.UpdateUserPreferencesRequest) (*userv1.UpdateUserPreferencesResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "user.UserService/UpdateUserPreferences", func(callCtx context.Context) (*userv1.UpdateUserPreferencesResponse, error) {
		return c.stub.UpdateUserPreferences(callCtx, in)
	})
}

func (c *client) ListUserActivityLogs(ctx context.Context, in *userv1.ListUserActivityLogsRequest) (*userv1.ListUserActivityLogsResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "user.UserService/ListUserActivityLogs", func(callCtx context.Context) (*userv1.ListUserActivityLogsResponse, error) {
		return c.stub.ListUserActivityLogs(callCtx, in)
	})
}

func (c *client) UpdateUsername(ctx context.Context, in *userv1.UpdateUsernameRequest) (*userv1.UpdateUsernameResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "user.UserService/UpdateUsername", func(callCtx context.Context) (*userv1.UpdateUsernameResponse, error) {
		return c.stub.UpdateUsername(callCtx, in)
	})
}

func (c *client) RequestEmailChange(ctx context.Context, in *userv1.RequestEmailChangeRequest) (*userv1.RequestEmailChangeResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "user.UserService/RequestEmailChange", func(callCtx context.Context) (*userv1.RequestEmailChangeResponse, error) {
		return c.stub.RequestEmailChange(callCtx, in)
	})
}

func (c *client) VerifyEmailChange(ctx context.Context, in *userv1.VerifyEmailChangeRequest) (*userv1.VerifyEmailChangeResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "user.UserService/VerifyEmailChange", func(callCtx context.Context) (*userv1.VerifyEmailChangeResponse, error) {
		return c.stub.VerifyEmailChange(callCtx, in)
	})
}

func (c *client) CompleteEmailChange(ctx context.Context, in *userv1.CompleteEmailChangeRequest) (*userv1.CompleteEmailChangeResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "user.UserService/CompleteEmailChange", func(callCtx context.Context) (*userv1.CompleteEmailChangeResponse, error) {
		return c.stub.CompleteEmailChange(callCtx, in)
	})
}

func (c *client) UpdatePassword(ctx context.Context, in *userv1.UpdatePasswordRequest) (*userv1.UpdatePasswordResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "user.UserService/UpdatePassword", func(callCtx context.Context) (*userv1.UpdatePasswordResponse, error) {
		return c.stub.UpdatePassword(callCtx, in)
	})
}

func (c *client) GetMFAStatus(ctx context.Context, in *userv1.GetMFAStatusRequest) (*userv1.GetMFAStatusResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "user.UserService/GetMFAStatus", func(callCtx context.Context) (*userv1.GetMFAStatusResponse, error) {
		return c.stub.GetMFAStatus(callCtx, in)
	})
}

func (c *client) BeginTOTPSetup(ctx context.Context, in *userv1.BeginTOTPSetupRequest) (*userv1.BeginTOTPSetupResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "user.UserService/BeginTOTPSetup", func(callCtx context.Context) (*userv1.BeginTOTPSetupResponse, error) {
		return c.stub.BeginTOTPSetup(callCtx, in)
	})
}

func (c *client) ConfirmTOTPSetup(ctx context.Context, in *userv1.ConfirmTOTPSetupRequest) (*userv1.ConfirmTOTPSetupResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "user.UserService/ConfirmTOTPSetup", func(callCtx context.Context) (*userv1.ConfirmTOTPSetupResponse, error) {
		return c.stub.ConfirmTOTPSetup(callCtx, in)
	})
}

func (c *client) DisableTOTP(ctx context.Context, in *userv1.DisableTOTPRequest) (*userv1.DisableTOTPResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "user.UserService/DisableTOTP", func(callCtx context.Context) (*userv1.DisableTOTPResponse, error) {
		return c.stub.DisableTOTP(callCtx, in)
	})
}

func (c *client) BeginWebAuthnRegister(ctx context.Context, in *userv1.BeginWebAuthnRegisterRequest) (*userv1.BeginWebAuthnRegisterResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "user.UserService/BeginWebAuthnRegister", func(callCtx context.Context) (*userv1.BeginWebAuthnRegisterResponse, error) {
		return c.stub.BeginWebAuthnRegister(callCtx, in)
	})
}

func (c *client) FinishWebAuthnRegister(ctx context.Context, in *userv1.FinishWebAuthnRegisterRequest) (*userv1.FinishWebAuthnRegisterResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "user.UserService/FinishWebAuthnRegister", func(callCtx context.Context) (*userv1.FinishWebAuthnRegisterResponse, error) {
		return c.stub.FinishWebAuthnRegister(callCtx, in)
	})
}

func (c *client) RenameWebAuthnCredential(ctx context.Context, in *userv1.RenameWebAuthnCredentialRequest) (*userv1.RenameWebAuthnCredentialResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "user.UserService/RenameWebAuthnCredential", func(callCtx context.Context) (*userv1.RenameWebAuthnCredentialResponse, error) {
		return c.stub.RenameWebAuthnCredential(callCtx, in)
	})
}

func (c *client) DeleteWebAuthnCredential(ctx context.Context, in *userv1.DeleteWebAuthnCredentialRequest) (*userv1.DeleteWebAuthnCredentialResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "user.UserService/DeleteWebAuthnCredential", func(callCtx context.Context) (*userv1.DeleteWebAuthnCredentialResponse, error) {
		return c.stub.DeleteWebAuthnCredential(callCtx, in)
	})
}

func (c *client) RegenerateRecoveryCodes(ctx context.Context, in *userv1.RegenerateRecoveryCodesRequest) (*userv1.RegenerateRecoveryCodesResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "user.UserService/RegenerateRecoveryCodes", func(callCtx context.Context) (*userv1.RegenerateRecoveryCodesResponse, error) {
		return c.stub.RegenerateRecoveryCodes(callCtx, in)
	})
}

func (c *client) DisableAccount(ctx context.Context, in *userv1.DisableAccountRequest) (*userv1.DisableAccountResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "user.UserService/DisableAccount", func(callCtx context.Context) (*userv1.DisableAccountResponse, error) {
		return c.stub.DisableAccount(callCtx, in)
	})
}

func (c *client) DeleteAccount(ctx context.Context, in *userv1.DeleteAccountRequest) (*userv1.DeleteAccountResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "user.UserService/DeleteAccount", func(callCtx context.Context) (*userv1.DeleteAccountResponse, error) {
		return c.stub.DeleteAccount(callCtx, in)
	})
}

func (c *client) ListSessions(ctx context.Context, in *userv1.ListSessionsRequest) (*userv1.ListSessionsResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "user.UserService/ListSessions", func(callCtx context.Context) (*userv1.ListSessionsResponse, error) {
		return c.stub.ListSessions(callCtx, in)
	})
}

func (c *client) RevokeSession(ctx context.Context, in *userv1.RevokeSessionRequest) (*userv1.RevokeSessionResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "user.UserService/RevokeSession", func(callCtx context.Context) (*userv1.RevokeSessionResponse, error) {
		return c.stub.RevokeSession(callCtx, in)
	})
}

func (c *client) RevokeOtherSessions(ctx context.Context, in *userv1.RevokeOtherSessionsRequest) (*userv1.RevokeOtherSessionsResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "user.UserService/RevokeOtherSessions", func(callCtx context.Context) (*userv1.RevokeOtherSessionsResponse, error) {
		return c.stub.RevokeOtherSessions(callCtx, in)
	})
}
