package grpc

import (
	"bytes"
	"context"
	"time"

	"github.com/wichat/wichat/backend/ms-user/internal/models"
	activitysvc "github.com/wichat/wichat/backend/ms-user/internal/user-activity-log/service"
	usersvc "github.com/wichat/wichat/backend/ms-user/internal/user/service"
	authPb "github.com/wichat/wichat/backend/ms-user/pkg/auth-pb"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	responsegrpc "github.com/wichat/wichat/backend/wi-shared/response/grpc"
)

// Server implements userv1.UserServiceServer.
type Server struct {
	userv1.UnimplementedUserServiceServer
	svc         usersvc.Service
	activitySvc activitysvc.Service
}

// NewServer creates a new UserServiceServer.
func NewServer(svc usersvc.Service) *Server {
	return &Server{svc: svc}
}

// WithActivityLog attaches the user activity log service for ListUserActivityLogs.
func (s *Server) WithActivityLog(activitySvc activitysvc.Service) *Server {
	s.activitySvc = activitySvc
	return s
}

func (s *Server) SetupStatus(ctx context.Context, _ *userv1.SetupStatusRequest) (*userv1.SetupStatusResponse, error) {
	out, err := s.svc.SetupStatus(ctx)
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.SetupStatusResponse{
		RegistrationMode: out.RegistrationMode,
	}, nil
}

func (s *Server) CompleteInstanceSetup(ctx context.Context, req *userv1.CompleteInstanceSetupRequest) (*userv1.CompleteInstanceSetupResponse, error) {
	in := models.CompleteInstanceSetupInput{
		Name:             req.GetName(),
		LogoURL:          req.GetLogoUrl(),
		RegistrationMode: req.GetRegistrationMode(),
	}

	if err := s.svc.CompleteInstanceSetup(ctx, req.GetAccessToken(), in); err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.CompleteInstanceSetupResponse{Ok: true}, nil
}

func (s *Server) CheckUsername(ctx context.Context, req *userv1.CheckUsernameRequest) (*userv1.CheckUsernameResponse, error) {
	out, err := s.svc.CheckUsername(ctx, req.GetUsername())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.CheckUsernameResponse{
		Available:  out.Available,
		Normalized: out.Normalized,
		Reason:     out.Reason,
	}, nil
}

func (s *Server) SearchUsers(ctx context.Context, req *userv1.SearchUsersRequest) (*userv1.SearchUsersResponse, error) {
	users, err := s.svc.SearchUsers(ctx, req.GetAccessToken(), req.GetQuery(), int(req.GetLimit()), req.GetGroupId())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.SearchUsersResponse{Users: authPb.UserSearchHits(users)}, nil
}

func (s *Server) UploadUserAvatar(ctx context.Context, req *userv1.UploadUserAvatarRequest) (*userv1.UploadUserAvatarResponse, error) {
	data := req.GetData()
	user, err := s.svc.UploadUserAvatar(
		ctx,
		req.GetAccessToken(),
		req.GetContentType(),
		bytes.NewReader(data),
		int64(len(data)),
	)
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.UploadUserAvatarResponse{User: authPb.User(*user)}, nil
}

func (s *Server) UpdateUserProfile(ctx context.Context, req *userv1.UpdateUserProfileRequest) (*userv1.UpdateUserProfileResponse, error) {
	user, err := s.svc.UpdateUserProfile(
		ctx,
		req.GetAccessToken(),
		req.GetDisplayName(),
		req.GetPronouns(),
		req.GetBio(),
		req.GetBannerColor(),
		req.GetClearAvatar(),
	)
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.UpdateUserProfileResponse{User: authPb.User(*user)}, nil
}

func (s *Server) UpdateUserStatusNote(ctx context.Context, req *userv1.UpdateUserStatusNoteRequest) (*userv1.UpdateUserStatusNoteResponse, error) {
	user, err := s.svc.UpdateUserStatusNote(
		ctx,
		req.GetAccessToken(),
		req.GetText(),
		req.GetEmoji(),
		req.GetExpiresAt(),
	)
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.UpdateUserStatusNoteResponse{User: authPb.User(*user)}, nil
}

func (s *Server) UpdateUserPreferences(ctx context.Context, req *userv1.UpdateUserPreferencesRequest) (*userv1.UpdateUserPreferencesResponse, error) {
	user, err := s.svc.UpdateUserPreferences(
		ctx,
		req.GetAccessToken(),
		req.GetLocale(),
		req.GetTimeFormat(),
	)
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.UpdateUserPreferencesResponse{User: authPb.User(*user)}, nil
}

func (s *Server) ListUserActivityLogs(ctx context.Context, req *userv1.ListUserActivityLogsRequest) (*userv1.ListUserActivityLogsResponse, error) {
	if s.activitySvc == nil {
		return nil, responsegrpc.ToStatus(exception.Unavailable(exception.MsgUnavailable, nil))
	}
	page, err := s.activitySvc.ListMyActivityLogs(
		ctx,
		req.GetAccessToken(),
		req.GetCategory(),
		req.GetAction(),
		int(req.GetLimit()),
		req.GetCursor(),
	)
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.ListUserActivityLogsResponse{
		Logs:       authPb.UserActivityLogs(page.Logs),
		NextCursor: page.NextCursor,
	}, nil
}

func (s *Server) UpdateUsername(ctx context.Context, req *userv1.UpdateUsernameRequest) (*userv1.UpdateUsernameResponse, error) {
	user, err := s.svc.UpdateUsername(
		ctx,
		req.GetAccessToken(),
		req.GetUsername(),
		req.GetCurrentPassword(),
	)
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.UpdateUsernameResponse{User: authPb.User(*user)}, nil
}

func (s *Server) RequestEmailChange(ctx context.Context, req *userv1.RequestEmailChangeRequest) (*userv1.RequestEmailChangeResponse, error) {
	ch, err := s.svc.RequestEmailChange(ctx, req.GetAccessToken(), req.GetLocale())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.RequestEmailChangeResponse{
		ChallengeId:           ch.ChallengeID,
		EmailHint:             ch.EmailHint,
		ResendCooldownSeconds: ch.ResendCooldownSeconds,
	}, nil
}

func (s *Server) VerifyEmailChange(ctx context.Context, req *userv1.VerifyEmailChangeRequest) (*userv1.VerifyEmailChangeResponse, error) {
	token, err := s.svc.VerifyEmailChange(ctx, req.GetAccessToken(), req.GetChallengeId(), req.GetCode())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.VerifyEmailChangeResponse{ChangeToken: token}, nil
}

func (s *Server) CompleteEmailChange(ctx context.Context, req *userv1.CompleteEmailChangeRequest) (*userv1.CompleteEmailChangeResponse, error) {
	user, err := s.svc.CompleteEmailChange(
		ctx,
		req.GetAccessToken(),
		req.GetChangeToken(),
		req.GetNewEmail(),
		req.GetCurrentPassword(),
		req.GetReason(),
	)
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.CompleteEmailChangeResponse{User: authPb.User(*user)}, nil
}

func (s *Server) UpdatePassword(ctx context.Context, req *userv1.UpdatePasswordRequest) (*userv1.UpdatePasswordResponse, error) {
	if err := s.svc.UpdatePassword(
		ctx,
		req.GetAccessToken(),
		req.GetCurrentPassword(),
		req.GetNewPassword(),
	); err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.UpdatePasswordResponse{}, nil
}

func (s *Server) GetMFAStatus(ctx context.Context, req *userv1.GetMFAStatusRequest) (*userv1.GetMFAStatusResponse, error) {
	st, err := s.svc.GetMFAStatus(ctx, req.GetAccessToken())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.GetMFAStatusResponse{
		TotpEnabled:            st.TOTPEnabled,
		SecurityKeys:           authPb.SecurityKeys(st.SecurityKeys),
		RecoveryCodesRemaining: st.RecoveryCodesRemaining,
	}, nil
}

func (s *Server) BeginTOTPSetup(ctx context.Context, req *userv1.BeginTOTPSetupRequest) (*userv1.BeginTOTPSetupResponse, error) {
	out, err := s.svc.BeginTOTPSetup(ctx, req.GetAccessToken(), req.GetCurrentPassword())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.BeginTOTPSetupResponse{
		SetupToken: out.SetupToken,
		Secret:     out.Secret,
		OtpauthUri: out.OtpauthURI,
	}, nil
}

func (s *Server) ConfirmTOTPSetup(ctx context.Context, req *userv1.ConfirmTOTPSetupRequest) (*userv1.ConfirmTOTPSetupResponse, error) {
	codes, err := s.svc.ConfirmTOTPSetup(ctx, req.GetAccessToken(), req.GetSetupToken(), req.GetCode())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.ConfirmTOTPSetupResponse{RecoveryCodes: codes}, nil
}

func (s *Server) DisableTOTP(ctx context.Context, req *userv1.DisableTOTPRequest) (*userv1.DisableTOTPResponse, error) {
	if err := s.svc.DisableTOTP(ctx, req.GetAccessToken(), req.GetCurrentPassword()); err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.DisableTOTPResponse{}, nil
}

func (s *Server) BeginWebAuthnRegister(ctx context.Context, req *userv1.BeginWebAuthnRegisterRequest) (*userv1.BeginWebAuthnRegisterResponse, error) {
	out, err := s.svc.BeginWebAuthnRegister(ctx, req.GetAccessToken(), req.GetCurrentPassword(), req.GetName())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.BeginWebAuthnRegisterResponse{
		ChallengeId: out.ChallengeID,
		OptionsJson: out.OptionsJSON,
	}, nil
}

func (s *Server) FinishWebAuthnRegister(ctx context.Context, req *userv1.FinishWebAuthnRegisterRequest) (*userv1.FinishWebAuthnRegisterResponse, error) {
	key, codes, err := s.svc.FinishWebAuthnRegister(ctx, req.GetAccessToken(), req.GetChallengeId(), req.GetCredentialJson())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.FinishWebAuthnRegisterResponse{
		SecurityKey:   authPb.SecurityKey(*key),
		RecoveryCodes: codes,
	}, nil
}

func (s *Server) RenameWebAuthnCredential(ctx context.Context, req *userv1.RenameWebAuthnCredentialRequest) (*userv1.RenameWebAuthnCredentialResponse, error) {
	key, err := s.svc.RenameWebAuthnCredential(ctx, req.GetAccessToken(), req.GetCredentialId(), req.GetName())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.RenameWebAuthnCredentialResponse{SecurityKey: authPb.SecurityKey(*key)}, nil
}

func (s *Server) DeleteWebAuthnCredential(ctx context.Context, req *userv1.DeleteWebAuthnCredentialRequest) (*userv1.DeleteWebAuthnCredentialResponse, error) {
	if err := s.svc.DeleteWebAuthnCredential(ctx, req.GetAccessToken(), req.GetCredentialId(), req.GetCurrentPassword()); err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.DeleteWebAuthnCredentialResponse{}, nil
}

func (s *Server) RegenerateRecoveryCodes(ctx context.Context, req *userv1.RegenerateRecoveryCodesRequest) (*userv1.RegenerateRecoveryCodesResponse, error) {
	codes, err := s.svc.RegenerateRecoveryCodes(ctx, req.GetAccessToken(), req.GetCurrentPassword())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.RegenerateRecoveryCodesResponse{RecoveryCodes: codes}, nil
}

func (s *Server) DisableAccount(ctx context.Context, req *userv1.DisableAccountRequest) (*userv1.DisableAccountResponse, error) {
	if err := s.svc.DisableAccount(ctx, req.GetAccessToken(), req.GetCurrentPassword()); err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.DisableAccountResponse{Ok: true}, nil
}

func (s *Server) DeleteAccount(ctx context.Context, req *userv1.DeleteAccountRequest) (*userv1.DeleteAccountResponse, error) {
	if err := s.svc.DeleteAccount(ctx, req.GetAccessToken(), req.GetCurrentPassword()); err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.DeleteAccountResponse{Ok: true}, nil
}

func (s *Server) ListSessions(ctx context.Context, req *userv1.ListSessionsRequest) (*userv1.ListSessionsResponse, error) {
	sessions, currentID, err := s.svc.ListSessions(ctx, req.GetAccessToken())
	if err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	out := make([]*userv1.UserSession, 0, len(sessions))
	for _, sess := range sessions {
		out = append(out, &userv1.UserSession{
			Id:            sess.ID,
			Platform:      sess.Platform,
			DeviceLabel:   sess.DeviceLabel,
			LocationLabel: sess.LocationLabel,
			UserAgent:     sess.UserAgent,
			LastUsedAt:    sess.LastUsedAt.UTC().Format(time.RFC3339),
			CreatedAt:     sess.CreatedAt.UTC().Format(time.RFC3339),
			IsCurrent:     sess.ID == currentID,
		})
	}

	return &userv1.ListSessionsResponse{Sessions: out}, nil
}

func (s *Server) RevokeSession(ctx context.Context, req *userv1.RevokeSessionRequest) (*userv1.RevokeSessionResponse, error) {
	if err := s.svc.RevokeSession(ctx, req.GetAccessToken(), req.GetSessionId(), req.GetCurrentPassword()); err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.RevokeSessionResponse{Ok: true}, nil
}

func (s *Server) RevokeOtherSessions(ctx context.Context, req *userv1.RevokeOtherSessionsRequest) (*userv1.RevokeOtherSessionsResponse, error) {
	if err := s.svc.RevokeOtherSessions(ctx, req.GetAccessToken(), req.GetCurrentPassword()); err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	return &userv1.RevokeOtherSessionsResponse{Ok: true}, nil
}
