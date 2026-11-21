package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	userdelivery "github.com/wichat/wichat/backend/api-gateway/internal/user/delivery/http"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices"
	msUser "github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeUser struct {
	setupStatus              func(context.Context, *userv1.SetupStatusRequest) (*userv1.SetupStatusResponse, error)
	checkUsername            func(context.Context, *userv1.CheckUsernameRequest) (*userv1.CheckUsernameResponse, error)
	uploadAvatar             func(context.Context, *userv1.UploadUserAvatarRequest) (*userv1.UploadUserAvatarResponse, error)
	updateProfile            func(context.Context, *userv1.UpdateUserProfileRequest) (*userv1.UpdateUserProfileResponse, error)
	updateStatusNote         func(context.Context, *userv1.UpdateUserStatusNoteRequest) (*userv1.UpdateUserStatusNoteResponse, error)
	updatePreferences        func(context.Context, *userv1.UpdateUserPreferencesRequest) (*userv1.UpdateUserPreferencesResponse, error)
	updateUsername           func(context.Context, *userv1.UpdateUsernameRequest) (*userv1.UpdateUsernameResponse, error)
	requestEmailChange       func(context.Context, *userv1.RequestEmailChangeRequest) (*userv1.RequestEmailChangeResponse, error)
	verifyEmailChange        func(context.Context, *userv1.VerifyEmailChangeRequest) (*userv1.VerifyEmailChangeResponse, error)
	completeEmailChange      func(context.Context, *userv1.CompleteEmailChangeRequest) (*userv1.CompleteEmailChangeResponse, error)
	updatePassword           func(context.Context, *userv1.UpdatePasswordRequest) (*userv1.UpdatePasswordResponse, error)
	getMFAStatus             func(context.Context, *userv1.GetMFAStatusRequest) (*userv1.GetMFAStatusResponse, error)
	beginTOTPSetup           func(context.Context, *userv1.BeginTOTPSetupRequest) (*userv1.BeginTOTPSetupResponse, error)
	confirmTOTPSetup         func(context.Context, *userv1.ConfirmTOTPSetupRequest) (*userv1.ConfirmTOTPSetupResponse, error)
	disableTOTP              func(context.Context, *userv1.DisableTOTPRequest) (*userv1.DisableTOTPResponse, error)
	beginWebAuthnRegister    func(context.Context, *userv1.BeginWebAuthnRegisterRequest) (*userv1.BeginWebAuthnRegisterResponse, error)
	finishWebAuthnRegister   func(context.Context, *userv1.FinishWebAuthnRegisterRequest) (*userv1.FinishWebAuthnRegisterResponse, error)
	renameWebAuthnCredential func(context.Context, *userv1.RenameWebAuthnCredentialRequest) (*userv1.RenameWebAuthnCredentialResponse, error)
	deleteWebAuthnCredential func(context.Context, *userv1.DeleteWebAuthnCredentialRequest) (*userv1.DeleteWebAuthnCredentialResponse, error)
	regenerateRecoveryCodes  func(context.Context, *userv1.RegenerateRecoveryCodesRequest) (*userv1.RegenerateRecoveryCodesResponse, error)
	disableAccount           func(context.Context, *userv1.DisableAccountRequest) (*userv1.DisableAccountResponse, error)
	deleteAccount            func(context.Context, *userv1.DeleteAccountRequest) (*userv1.DeleteAccountResponse, error)
	listSessions             func(context.Context, *userv1.ListSessionsRequest) (*userv1.ListSessionsResponse, error)
	listRelationships        func(context.Context, *userv1.ListUserRelationshipsRequest) (*userv1.ListUserRelationshipsResponse, error)
	setRelationship          func(context.Context, *userv1.SetUserRelationshipRequest) (*userv1.SetUserRelationshipResponse, error)
	revokeSession            func(context.Context, *userv1.RevokeSessionRequest) (*userv1.RevokeSessionResponse, error)
	revokeOtherSessions      func(context.Context, *userv1.RevokeOtherSessionsRequest) (*userv1.RevokeOtherSessionsResponse, error)
	listUserActivityLogs     func(context.Context, *userv1.ListUserActivityLogsRequest) (*userv1.ListUserActivityLogsResponse, error)
}

func (f fakeUser) SetupStatus(ctx context.Context, in *userv1.SetupStatusRequest) (*userv1.SetupStatusResponse, error) {
	if f.setupStatus != nil {
		return f.setupStatus(ctx, in)
	}

	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (fakeUser) CompleteInstanceSetup(context.Context, *userv1.CompleteInstanceSetupRequest) (*userv1.CompleteInstanceSetupResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeUser) CheckUsername(ctx context.Context, in *userv1.CheckUsernameRequest) (*userv1.CheckUsernameResponse, error) {
	if f.checkUsername != nil {
		return f.checkUsername(ctx, in)
	}

	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (fakeUser) SearchUsers(context.Context, *userv1.SearchUsersRequest) (*userv1.SearchUsersResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeUser) UploadUserAvatar(ctx context.Context, in *userv1.UploadUserAvatarRequest) (*userv1.UploadUserAvatarResponse, error) {
	if f.uploadAvatar != nil {
		return f.uploadAvatar(ctx, in)
	}

	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeUser) UpdateUserProfile(ctx context.Context, in *userv1.UpdateUserProfileRequest) (*userv1.UpdateUserProfileResponse, error) {
	if f.updateProfile != nil {
		return f.updateProfile(ctx, in)
	}

	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeUser) UpdateUserStatusNote(ctx context.Context, in *userv1.UpdateUserStatusNoteRequest) (*userv1.UpdateUserStatusNoteResponse, error) {
	if f.updateStatusNote != nil {
		return f.updateStatusNote(ctx, in)
	}

	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeUser) UpdateUserPreferences(ctx context.Context, in *userv1.UpdateUserPreferencesRequest) (*userv1.UpdateUserPreferencesResponse, error) {
	if f.updatePreferences != nil {
		return f.updatePreferences(ctx, in)
	}

	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeUser) UpdateUsername(ctx context.Context, in *userv1.UpdateUsernameRequest) (*userv1.UpdateUsernameResponse, error) {
	if f.updateUsername != nil {
		return f.updateUsername(ctx, in)
	}

	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeUser) RequestEmailChange(ctx context.Context, in *userv1.RequestEmailChangeRequest) (*userv1.RequestEmailChangeResponse, error) {
	if f.requestEmailChange != nil {
		return f.requestEmailChange(ctx, in)
	}

	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeUser) VerifyEmailChange(ctx context.Context, in *userv1.VerifyEmailChangeRequest) (*userv1.VerifyEmailChangeResponse, error) {
	if f.verifyEmailChange != nil {
		return f.verifyEmailChange(ctx, in)
	}

	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeUser) CompleteEmailChange(ctx context.Context, in *userv1.CompleteEmailChangeRequest) (*userv1.CompleteEmailChangeResponse, error) {
	if f.completeEmailChange != nil {
		return f.completeEmailChange(ctx, in)
	}

	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeUser) UpdatePassword(ctx context.Context, in *userv1.UpdatePasswordRequest) (*userv1.UpdatePasswordResponse, error) {
	if f.updatePassword != nil {
		return f.updatePassword(ctx, in)
	}

	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeUser) GetMFAStatus(ctx context.Context, in *userv1.GetMFAStatusRequest) (*userv1.GetMFAStatusResponse, error) {
	if f.getMFAStatus != nil {
		return f.getMFAStatus(ctx, in)
	}
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeUser) BeginTOTPSetup(ctx context.Context, in *userv1.BeginTOTPSetupRequest) (*userv1.BeginTOTPSetupResponse, error) {
	if f.beginTOTPSetup != nil {
		return f.beginTOTPSetup(ctx, in)
	}
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeUser) ConfirmTOTPSetup(ctx context.Context, in *userv1.ConfirmTOTPSetupRequest) (*userv1.ConfirmTOTPSetupResponse, error) {
	if f.confirmTOTPSetup != nil {
		return f.confirmTOTPSetup(ctx, in)
	}
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeUser) DisableTOTP(ctx context.Context, in *userv1.DisableTOTPRequest) (*userv1.DisableTOTPResponse, error) {
	if f.disableTOTP != nil {
		return f.disableTOTP(ctx, in)
	}
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeUser) BeginWebAuthnRegister(ctx context.Context, in *userv1.BeginWebAuthnRegisterRequest) (*userv1.BeginWebAuthnRegisterResponse, error) {
	if f.beginWebAuthnRegister != nil {
		return f.beginWebAuthnRegister(ctx, in)
	}
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeUser) FinishWebAuthnRegister(ctx context.Context, in *userv1.FinishWebAuthnRegisterRequest) (*userv1.FinishWebAuthnRegisterResponse, error) {
	if f.finishWebAuthnRegister != nil {
		return f.finishWebAuthnRegister(ctx, in)
	}
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeUser) RenameWebAuthnCredential(ctx context.Context, in *userv1.RenameWebAuthnCredentialRequest) (*userv1.RenameWebAuthnCredentialResponse, error) {
	if f.renameWebAuthnCredential != nil {
		return f.renameWebAuthnCredential(ctx, in)
	}
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeUser) DeleteWebAuthnCredential(ctx context.Context, in *userv1.DeleteWebAuthnCredentialRequest) (*userv1.DeleteWebAuthnCredentialResponse, error) {
	if f.deleteWebAuthnCredential != nil {
		return f.deleteWebAuthnCredential(ctx, in)
	}
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeUser) RegenerateRecoveryCodes(ctx context.Context, in *userv1.RegenerateRecoveryCodesRequest) (*userv1.RegenerateRecoveryCodesResponse, error) {
	if f.regenerateRecoveryCodes != nil {
		return f.regenerateRecoveryCodes(ctx, in)
	}
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeUser) DisableAccount(ctx context.Context, in *userv1.DisableAccountRequest) (*userv1.DisableAccountResponse, error) {
	if f.disableAccount != nil {
		return f.disableAccount(ctx, in)
	}
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeUser) DeleteAccount(ctx context.Context, in *userv1.DeleteAccountRequest) (*userv1.DeleteAccountResponse, error) {
	if f.deleteAccount != nil {
		return f.deleteAccount(ctx, in)
	}
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeUser) ListUserRelationships(ctx context.Context, in *userv1.ListUserRelationshipsRequest) (*userv1.ListUserRelationshipsResponse, error) {
	if f.listRelationships != nil {
		return f.listRelationships(ctx, in)
	}
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeUser) SetUserRelationship(ctx context.Context, in *userv1.SetUserRelationshipRequest) (*userv1.SetUserRelationshipResponse, error) {
	if f.setRelationship != nil {
		return f.setRelationship(ctx, in)
	}
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeUser) ListSessions(ctx context.Context, in *userv1.ListSessionsRequest) (*userv1.ListSessionsResponse, error) {
	if f.listSessions != nil {
		return f.listSessions(ctx, in)
	}
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeUser) RevokeSession(ctx context.Context, in *userv1.RevokeSessionRequest) (*userv1.RevokeSessionResponse, error) {
	if f.revokeSession != nil {
		return f.revokeSession(ctx, in)
	}
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeUser) RevokeOtherSessions(ctx context.Context, in *userv1.RevokeOtherSessionsRequest) (*userv1.RevokeOtherSessionsResponse, error) {
	if f.revokeOtherSessions != nil {
		return f.revokeOtherSessions(ctx, in)
	}
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeUser) ListUserActivityLogs(ctx context.Context, in *userv1.ListUserActivityLogsRequest) (*userv1.ListUserActivityLogsResponse, error) {
	if f.listUserActivityLogs != nil {
		return f.listUserActivityLogs(ctx, in)
	}
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func TestSetupStatusReturnsInstancePolicy(t *testing.T) {
	clients := &microservices.Clients{User: msUser.Service{GRPC: msUser.GRPC{
		User: fakeUser{
			setupStatus: func(context.Context, *userv1.SetupStatusRequest) (*userv1.SetupStatusResponse, error) {
				return &userv1.SetupStatusResponse{
					RegistrationMode: "invite_only",
				}, nil
			},
		},
	}}}
	handler := userdelivery.NewHandler(clients, logger.New(logger.Options{Service: "gateway-test", Env: "test", Level: "error"})).
		WithInviteOnly(true)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/setup-status", nil)
	rec := httptest.NewRecorder()
	handler.SetupStatus(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Data struct {
			RegistrationMode string `json:"registration_mode"`
			InviteOnly       bool   `json:"invite_only"`
			NeedsAccount     *bool  `json:"needs_account"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Data.RegistrationMode != "invite_only" {
		t.Fatalf("policy: %+v", body.Data)
	}
	if !body.Data.InviteOnly {
		t.Fatalf("expected invite_only from env wiring, got %+v", body.Data)
	}
	if body.Data.NeedsAccount != nil {
		t.Fatalf("deprecated field leaked: %+v", body.Data)
	}
}

func TestCheckUsernameReturnsAvailability(t *testing.T) {
	clients := &microservices.Clients{User: msUser.Service{GRPC: msUser.GRPC{
		User: fakeUser{
			checkUsername: func(_ context.Context, in *userv1.CheckUsernameRequest) (*userv1.CheckUsernameResponse, error) {
				return &userv1.CheckUsernameResponse{
					Available:  true,
					Normalized: "ada.lovelace",
				}, nil
			},
		},
	}}}
	handler := userdelivery.NewHandler(clients, logger.New(logger.Options{Service: "gateway-test", Env: "test", Level: "error"}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/check-username?username=Ada.Lovelace", nil)
	rec := httptest.NewRecorder()
	handler.CheckUsername(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestUploadAvatarRequiresBearerAndForwardsFile(t *testing.T) {
	handler := userdelivery.NewHandler(&microservices.Clients{}, logger.New(logger.Options{Service: "gateway-test", Env: "test", Level: "error"}))

	unauthReq := httptest.NewRequest(http.MethodPost, "/api/v1/users/me/avatar", nil)
	unauthRec := httptest.NewRecorder()
	handler.UploadAvatar(unauthRec, unauthReq)
	if unauthRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", unauthRec.Code)
	}

	var called bool
	clients := &microservices.Clients{User: msUser.Service{GRPC: msUser.GRPC{
		User: fakeUser{
			uploadAvatar: func(_ context.Context, in *userv1.UploadUserAvatarRequest) (*userv1.UploadUserAvatarResponse, error) {
				called = true
				if in.GetAccessToken() != "tok-1" || len(in.GetData()) == 0 {
					t.Fatalf("unexpected upload request: %+v", in)
				}
				return &userv1.UploadUserAvatarResponse{
					User: &userv1.User{
						Id:          "u-1",
						Email:       "ada@example.com",
						DisplayName: "Ada",
						Username:    "ada",
						AvatarUrl:   "/media/avatars/u-1.png",
					},
				}, nil
			},
		},
	}}}
	authedHandler := userdelivery.NewHandler(clients, logger.New(logger.Options{Service: "gateway-test", Env: "test", Level: "error"}))

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	part, err := mw.CreateFormFile("file", "avatar.png")
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := part.Write([]byte("png-bytes")); err != nil {
		t.Fatalf("part.Write: %v", err)
	}
	if err := mw.Close(); err != nil {
		t.Fatalf("mw.Close: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/me/avatar", &body)
	req.Header.Set("Authorization", "Bearer tok-1")
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rec := httptest.NewRecorder()
	authedHandler.UploadAvatar(rec, req)

	if rec.Code != http.StatusOK || !called {
		t.Fatalf("status %d called=%v body %s", rec.Code, called, rec.Body.String())
	}
}

func TestUpdateProfileForwardsFields(t *testing.T) {
	var got *userv1.UpdateUserProfileRequest
	clients := &microservices.Clients{User: msUser.Service{GRPC: msUser.GRPC{
		User: fakeUser{
			updateProfile: func(_ context.Context, in *userv1.UpdateUserProfileRequest) (*userv1.UpdateUserProfileResponse, error) {
				got = in
				return &userv1.UpdateUserProfileResponse{
					User: &userv1.User{
						Id:          "u-1",
						Email:       "ada@example.com",
						DisplayName: "Ada Lovelace",
						Username:    "ada",
						Pronouns:    "she/her",
						Bio:         "Mathematician",
						BannerColor: "#336699",
					},
				}, nil
			},
		},
	}}}
	handler := userdelivery.NewHandler(clients, logger.New(logger.Options{Service: "gateway-test", Env: "test", Level: "error"}))

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/me/profile", strings.NewReader(`{"display_name":"Ada Lovelace","pronouns":"she/her","bio":"Mathematician","banner_color":"#336699"}`))
	req.Header.Set("Authorization", "Bearer tok-2")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.UpdateProfile(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	if got == nil || got.GetAccessToken() != "tok-2" || got.GetPronouns() != "she/her" || got.GetBannerColor() != "#336699" {
		t.Fatalf("unexpected grpc payload: %+v", got)
	}
}

func TestUpdateStatusNoteForwardsFields(t *testing.T) {
	var got *userv1.UpdateUserStatusNoteRequest
	clients := &microservices.Clients{User: msUser.Service{GRPC: msUser.GRPC{
		User: fakeUser{
			updateStatusNote: func(_ context.Context, in *userv1.UpdateUserStatusNoteRequest) (*userv1.UpdateUserStatusNoteResponse, error) {
				got = in
				return &userv1.UpdateUserStatusNoteResponse{
					User: &userv1.User{
						Id:                  "u-1",
						Email:               "ada@example.com",
						DisplayName:         "Ada Lovelace",
						Username:            "ada",
						StatusNoteText:      in.GetText(),
						StatusNoteEmoji:     in.GetEmoji(),
						StatusNoteExpiresAt: in.GetExpiresAt(),
					},
				}, nil
			},
		},
	}}}
	handler := userdelivery.NewHandler(clients, logger.New(logger.Options{Service: "gateway-test", Env: "test", Level: "error"}))

	unauthReq := httptest.NewRequest(http.MethodPatch, "/api/v1/users/me/status-note", strings.NewReader(`{}`))
	unauthRec := httptest.NewRecorder()
	handler.UpdateStatusNote(unauthRec, unauthReq)
	if unauthRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", unauthRec.Code)
	}

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/me/status-note", strings.NewReader(`{"text":"Deep work","emoji":"🎧","expires_at":"2026-11-15T12:00:00Z"}`))
	req.Header.Set("Authorization", "Bearer tok-note")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.UpdateStatusNote(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	if got == nil || got.GetAccessToken() != "tok-note" || got.GetText() != "Deep work" || got.GetEmoji() != "🎧" || got.GetExpiresAt() != "2026-11-15T12:00:00Z" {
		t.Fatalf("unexpected grpc payload: %+v", got)
	}
}

func TestUpdatePreferencesForwardsFields(t *testing.T) {
	var got *userv1.UpdateUserPreferencesRequest
	clients := &microservices.Clients{User: msUser.Service{GRPC: msUser.GRPC{
		User: fakeUser{
			updatePreferences: func(_ context.Context, in *userv1.UpdateUserPreferencesRequest) (*userv1.UpdateUserPreferencesResponse, error) {
				got = in
				return &userv1.UpdateUserPreferencesResponse{
					User: &userv1.User{
						Id:         "u-1",
						Email:      "ada@example.com",
						Username:   "ada",
						Locale:     in.GetLocale(),
						TimeFormat: in.GetTimeFormat(),
					},
				}, nil
			},
		},
	}}}
	handler := userdelivery.NewHandler(clients, logger.New(logger.Options{Service: "gateway-test", Env: "test", Level: "error"}))

	unauthReq := httptest.NewRequest(http.MethodPatch, "/api/v1/users/me/preferences", strings.NewReader(`{}`))
	unauthRec := httptest.NewRecorder()
	handler.UpdatePreferences(unauthRec, unauthReq)
	if unauthRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", unauthRec.Code)
	}

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/me/preferences", strings.NewReader(`{"locale":"id","time_format":"24h"}`))
	req.Header.Set("Authorization", "Bearer tok-prefs")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.UpdatePreferences(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	if got == nil || got.GetAccessToken() != "tok-prefs" || got.GetLocale() != "id" || got.GetTimeFormat() != "24h" {
		t.Fatalf("unexpected grpc payload: %+v", got)
	}
}

func TestUpdateUsernameForwardsFields(t *testing.T) {
	var got *userv1.UpdateUsernameRequest
	clients := &microservices.Clients{User: msUser.Service{GRPC: msUser.GRPC{
		User: fakeUser{
			updateUsername: func(_ context.Context, in *userv1.UpdateUsernameRequest) (*userv1.UpdateUsernameResponse, error) {
				got = in
				return &userv1.UpdateUsernameResponse{
					User: &userv1.User{
						Id:          "u-1",
						Email:       "ada@example.com",
						DisplayName: "Ada Lovelace",
						Username:    "ada.new",
					},
				}, nil
			},
		},
	}}}
	handler := userdelivery.NewHandler(clients, logger.New(logger.Options{Service: "gateway-test", Env: "test", Level: "error"}))

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/me/username", strings.NewReader(`{"username":"ada.new","current_password":"StrongPass123!"}`))
	req.Header.Set("Authorization", "Bearer tok-3")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.UpdateUsername(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	if got == nil || got.GetAccessToken() != "tok-3" || got.GetUsername() != "ada.new" || got.GetCurrentPassword() != "StrongPass123!" {
		t.Fatalf("unexpected grpc payload: %+v", got)
	}
}

func TestEmailChangeHandlersForwardFields(t *testing.T) {
	var gotReq *userv1.RequestEmailChangeRequest
	var gotVer *userv1.VerifyEmailChangeRequest
	var gotComp *userv1.CompleteEmailChangeRequest
	clients := &microservices.Clients{User: msUser.Service{GRPC: msUser.GRPC{
		User: fakeUser{
			requestEmailChange: func(_ context.Context, in *userv1.RequestEmailChangeRequest) (*userv1.RequestEmailChangeResponse, error) {
				gotReq = in
				return &userv1.RequestEmailChangeResponse{ChallengeId: "ch-1", EmailHint: "a***a@example.com", ResendCooldownSeconds: 60}, nil
			},
			verifyEmailChange: func(_ context.Context, in *userv1.VerifyEmailChangeRequest) (*userv1.VerifyEmailChangeResponse, error) {
				gotVer = in
				return &userv1.VerifyEmailChangeResponse{ChangeToken: "ct-1"}, nil
			},
			completeEmailChange: func(_ context.Context, in *userv1.CompleteEmailChangeRequest) (*userv1.CompleteEmailChangeResponse, error) {
				gotComp = in
				return &userv1.CompleteEmailChangeResponse{
					User: &userv1.User{Id: "u-1", Email: "new@example.com"},
				}, nil
			},
		},
	}}}
	handler := userdelivery.NewHandler(clients, logger.New(logger.Options{Service: "gateway-test", Env: "test", Level: "error"}))

	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/users/me/email/request", strings.NewReader(`{"locale":"id"}`))
	req1.Header.Set("Authorization", "Bearer tok-4")
	req1.Header.Set("Content-Type", "application/json")
	rec1 := httptest.NewRecorder()
	handler.RequestEmailChange(rec1, req1)
	if rec1.Code != http.StatusOK || gotReq.GetLocale() != "id" {
		t.Fatalf("request email change: %d %s", rec1.Code, rec1.Body.String())
	}

	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/users/me/email/verify", strings.NewReader(`{"challenge_id":"ch-1","code":"ABC123"}`))
	req2.Header.Set("Authorization", "Bearer tok-4")
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	handler.VerifyEmailChange(rec2, req2)
	if rec2.Code != http.StatusOK || gotVer.GetChallengeId() != "ch-1" || gotVer.GetCode() != "ABC123" {
		t.Fatalf("verify email change: %d %s", rec2.Code, rec2.Body.String())
	}

	req3 := httptest.NewRequest(http.MethodPatch, "/api/v1/users/me/email", strings.NewReader(`{"change_token":"ct-1","new_email":"new@example.com","current_password":"StrongPass123!","reason":"new_email"}`))
	req3.Header.Set("Authorization", "Bearer tok-4")
	req3.Header.Set("Content-Type", "application/json")
	rec3 := httptest.NewRecorder()
	handler.CompleteEmailChange(rec3, req3)
	if rec3.Code != http.StatusOK || gotComp.GetChangeToken() != "ct-1" || gotComp.GetNewEmail() != "new@example.com" || gotComp.GetReason() != "new_email" {
		t.Fatalf("complete email change: %d %s", rec3.Code, rec3.Body.String())
	}
}

func TestUpdatePasswordForwardsFields(t *testing.T) {
	var got *userv1.UpdatePasswordRequest
	clients := &microservices.Clients{User: msUser.Service{GRPC: msUser.GRPC{
		User: fakeUser{
			updatePassword: func(_ context.Context, in *userv1.UpdatePasswordRequest) (*userv1.UpdatePasswordResponse, error) {
				got = in
				return &userv1.UpdatePasswordResponse{}, nil
			},
		},
	}}}
	handler := userdelivery.NewHandler(clients, logger.New(logger.Options{Service: "gateway-test", Env: "test", Level: "error"}))

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/me/password", strings.NewReader(`{"current_password":"OldPass123!","new_password":"NewStrongPassword123!"}`))
	req.Header.Set("Authorization", "Bearer tok-5")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.UpdatePassword(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	if got == nil || got.GetAccessToken() != "tok-5" || got.GetCurrentPassword() != "OldPass123!" || got.GetNewPassword() != "NewStrongPassword123!" {
		t.Fatalf("unexpected grpc payload: %+v", got)
	}
}

func TestMFAHandlersForwardFields(t *testing.T) {
	clients := &microservices.Clients{User: msUser.Service{GRPC: msUser.GRPC{
		User: fakeUser{
			getMFAStatus: func(_ context.Context, in *userv1.GetMFAStatusRequest) (*userv1.GetMFAStatusResponse, error) {
				return &userv1.GetMFAStatusResponse{
					TotpEnabled:            true,
					RecoveryCodesRemaining: 10,
					SecurityKeys: []*userv1.SecurityKey{
						{Id: "k-1", Name: "YubiKey", CreatedAt: "2026-01-01T00:00:00Z", Transports: []string{"usb"}},
					},
				}, nil
			},
			beginTOTPSetup: func(_ context.Context, in *userv1.BeginTOTPSetupRequest) (*userv1.BeginTOTPSetupResponse, error) {
				return &userv1.BeginTOTPSetupResponse{SetupToken: "st-1", Secret: "BASE32", OtpauthUri: "otpauth://totp/WiChat"}, nil
			},
			confirmTOTPSetup: func(_ context.Context, in *userv1.ConfirmTOTPSetupRequest) (*userv1.ConfirmTOTPSetupResponse, error) {
				return &userv1.ConfirmTOTPSetupResponse{RecoveryCodes: []string{"AAAA-BBBB"}}, nil
			},
			disableTOTP: func(_ context.Context, in *userv1.DisableTOTPRequest) (*userv1.DisableTOTPResponse, error) {
				return &userv1.DisableTOTPResponse{}, nil
			},
			beginWebAuthnRegister: func(_ context.Context, in *userv1.BeginWebAuthnRegisterRequest) (*userv1.BeginWebAuthnRegisterResponse, error) {
				return &userv1.BeginWebAuthnRegisterResponse{ChallengeId: "wc-1", OptionsJson: `{"publicKey":{}}`}, nil
			},
			finishWebAuthnRegister: func(_ context.Context, in *userv1.FinishWebAuthnRegisterRequest) (*userv1.FinishWebAuthnRegisterResponse, error) {
				return &userv1.FinishWebAuthnRegisterResponse{
					SecurityKey:   &userv1.SecurityKey{Id: "k-2", Name: "Touch ID"},
					RecoveryCodes: []string{"CCCC-DDDD"},
				}, nil
			},
			renameWebAuthnCredential: func(_ context.Context, in *userv1.RenameWebAuthnCredentialRequest) (*userv1.RenameWebAuthnCredentialResponse, error) {
				return &userv1.RenameWebAuthnCredentialResponse{SecurityKey: &userv1.SecurityKey{Id: in.GetCredentialId(), Name: in.GetName()}}, nil
			},
			deleteWebAuthnCredential: func(_ context.Context, in *userv1.DeleteWebAuthnCredentialRequest) (*userv1.DeleteWebAuthnCredentialResponse, error) {
				return &userv1.DeleteWebAuthnCredentialResponse{}, nil
			},
			regenerateRecoveryCodes: func(_ context.Context, in *userv1.RegenerateRecoveryCodesRequest) (*userv1.RegenerateRecoveryCodesResponse, error) {
				return &userv1.RegenerateRecoveryCodesResponse{RecoveryCodes: []string{"EEEE-FFFF"}}, nil
			},
			disableAccount: func(_ context.Context, _ *userv1.DisableAccountRequest) (*userv1.DisableAccountResponse, error) {
				return &userv1.DisableAccountResponse{Ok: true}, nil
			},
			deleteAccount: func(_ context.Context, _ *userv1.DeleteAccountRequest) (*userv1.DeleteAccountResponse, error) {
				return &userv1.DeleteAccountResponse{Ok: true}, nil
			},
			listSessions: func(_ context.Context, _ *userv1.ListSessionsRequest) (*userv1.ListSessionsResponse, error) {
				return &userv1.ListSessionsResponse{Sessions: []*userv1.UserSession{{Id: "s-1", Platform: "web", DeviceLabel: "Chrome · macOS", IsCurrent: true}}}, nil
			},
			revokeSession: func(_ context.Context, _ *userv1.RevokeSessionRequest) (*userv1.RevokeSessionResponse, error) {
				return &userv1.RevokeSessionResponse{Ok: true}, nil
			},
			revokeOtherSessions: func(_ context.Context, _ *userv1.RevokeOtherSessionsRequest) (*userv1.RevokeOtherSessionsResponse, error) {
				return &userv1.RevokeOtherSessionsResponse{Ok: true}, nil
			},
		},
	}}}
	handler := userdelivery.NewHandler(clients, logger.New(logger.Options{Service: "gateway-test", Env: "test", Level: "error"}))

	doReq := func(fn func(http.ResponseWriter, *http.Request), method, target, body, pathID string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, target, strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer tok-mfa")
		req.Header.Set("Content-Type", "application/json")
		if pathID != "" {
			req.SetPathValue("id", pathID)
		}
		rec := httptest.NewRecorder()
		fn(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s %s: status %d body %s", method, target, rec.Code, rec.Body.String())
		}
		return rec
	}

	doReq(handler.GetMFAStatus, http.MethodGet, "/api/v1/users/me/mfa", "", "")
	doReq(handler.BeginTOTPSetup, http.MethodPost, "/api/v1/users/me/mfa/totp/begin", `{"current_password":"pw"}`, "")
	doReq(handler.ConfirmTOTPSetup, http.MethodPost, "/api/v1/users/me/mfa/totp/confirm", `{"setup_token":"st-1","code":"123456"}`, "")
	doReq(handler.DisableTOTP, http.MethodDelete, "/api/v1/users/me/mfa/totp", `{"current_password":"pw"}`, "")
	doReq(handler.BeginWebAuthnRegister, http.MethodPost, "/api/v1/users/me/mfa/security-keys/begin", `{"current_password":"pw","name":"Touch ID"}`, "")
	doReq(handler.FinishWebAuthnRegister, http.MethodPost, "/api/v1/users/me/mfa/security-keys/finish", `{"challenge_id":"wc-1","credential":{"id":"abc"}}`, "")
	doReq(handler.RenameWebAuthnCredential, http.MethodPatch, "/api/v1/users/me/mfa/security-keys/k-2", `{"name":"MacBook"}`, "k-2")
	doReq(handler.DeleteWebAuthnCredential, http.MethodDelete, "/api/v1/users/me/mfa/security-keys/k-2", `{"current_password":"pw"}`, "k-2")
	doReq(handler.RegenerateRecoveryCodes, http.MethodPost, "/api/v1/users/me/mfa/recovery-codes/regenerate", `{"current_password":"pw"}`, "")
	doReq(handler.DisableAccount, http.MethodPost, "/api/v1/users/me/disable", `{"current_password":"pw"}`, "")
	doReq(handler.DeleteAccount, http.MethodPost, "/api/v1/users/me/delete", `{"current_password":"pw"}`, "")
	doReq(handler.ListDevices, http.MethodGet, "/api/v1/users/me/devices", "", "")
	doReq(handler.RevokeDevice, http.MethodDelete, "/api/v1/users/me/devices/s-2", `{"current_password":"pw"}`, "s-2")
	doReq(handler.RevokeOtherDevices, http.MethodPost, "/api/v1/users/me/devices/revoke-others", `{"current_password":"pw"}`, "")
}

func TestListActivityLogsForwardsQueryAndMapsEnvelope(t *testing.T) {
	var got *userv1.ListUserActivityLogsRequest
	clients := &microservices.Clients{User: msUser.Service{GRPC: msUser.GRPC{
		User: fakeUser{
			listUserActivityLogs: func(_ context.Context, in *userv1.ListUserActivityLogsRequest) (*userv1.ListUserActivityLogsResponse, error) {
				got = in
				return &userv1.ListUserActivityLogsResponse{
					Logs: []*userv1.UserActivityLog{
						{
							Id:           "log-1",
							UserId:       "u-1",
							ActorId:      "u-1",
							Category:     "profile",
							Action:       "user.status_note.updated",
							TargetType:   "user",
							TargetId:     "u-1",
							Summary:      "Updated status note",
							ChangesJson:  `{"status_note_text":{"new":"Focus"}}`,
							MetadataJson: `{}`,
							CreatedAt:    "2026-11-15T01:00:00Z",
						},
					},
					NextCursor: "cur-2",
				}, nil
			},
		},
	}}}
	h := userdelivery.NewHandler(clients, logger.New(logger.Options{Service: "gateway-test", Env: "test", Level: "error"}))

	unauthReq := httptest.NewRequest(http.MethodGet, "/api/v1/users/me/activity-logs", nil)
	unauthRec := httptest.NewRecorder()
	h.ListActivityLogs(unauthRec, unauthReq)
	if unauthRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without bearer token, got %d", unauthRec.Code)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me/activity-logs?category=profile&action=user.status_note.updated&limit=15&cursor=cur-1", nil)
	req.Header.Set("Authorization", "Bearer tok-act")
	rec := httptest.NewRecorder()
	h.ListActivityLogs(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body %s", rec.Code, rec.Body.String())
	}
	if got == nil || got.GetAccessToken() != "tok-act" || got.GetCategory() != "profile" || got.GetAction() != "user.status_note.updated" || got.GetLimit() != 15 || got.GetCursor() != "cur-1" {
		t.Fatalf("unexpected grpc query: %+v", got)
	}
	if !strings.Contains(rec.Body.String(), `"user.status_note.updated"`) || !strings.Contains(rec.Body.String(), `"next_cursor":"cur-2"`) {
		t.Fatalf("unexpected json body: %s", rec.Body.String())
	}
}

func TestRelationshipsEndpoints(t *testing.T) {
	clients := &microservices.Clients{User: msUser.Service{GRPC: msUser.GRPC{
		User: fakeUser{
			listRelationships: func(_ context.Context, in *userv1.ListUserRelationshipsRequest) (*userv1.ListUserRelationshipsResponse, error) {
				if in.GetAccessToken() != "tok-rel" {
					t.Fatalf("unexpected token %q", in.GetAccessToken())
				}
				return &userv1.ListUserRelationshipsResponse{
					Relationships: []*userv1.UserRelationshipEntry{
						{TargetUserId: "u2", Kind: "ignore", CreatedAt: "2026-01-01T00:00:00Z"},
					},
				}, nil
			},
			setRelationship: func(_ context.Context, in *userv1.SetUserRelationshipRequest) (*userv1.SetUserRelationshipResponse, error) {
				if in.GetTargetUserId() != "u2" || in.GetKind() != "ignore" {
					t.Fatalf("unexpected req %+v", in)
				}
				return &userv1.SetUserRelationshipResponse{Ok: true, Kind: "ignore"}, nil
			},
		},
	}}}
	h := userdelivery.NewHandler(clients, logger.New(logger.Options{Service: "gateway-test", Env: "test", Level: "error"}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me/relationships", nil)
	req.Header.Set("Authorization", "Bearer tok-rel")
	rec := httptest.NewRecorder()
	h.ListRelationships(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"target_user_id":"u2"`) {
		t.Fatalf("list relationships: %d %s", rec.Code, rec.Body.String())
	}

	putReq := httptest.NewRequest(http.MethodPut, "/api/v1/users/me/relationships/u2", strings.NewReader(`{"kind":"ignore"}`))
	putReq.SetPathValue("userId", "u2")
	putReq.Header.Set("Authorization", "Bearer tok-rel")
	putRec := httptest.NewRecorder()
	h.SetRelationship(putRec, putReq)
	if putRec.Code != http.StatusOK || !strings.Contains(putRec.Body.String(), `"kind":"ignore"`) {
		t.Fatalf("set relationship: %d %s", putRec.Code, putRec.Body.String())
	}
}

type fakePresence struct {
	getMy func(context.Context, *userv1.GetMyPresenceRequest) (*userv1.GetMyPresenceResponse, error)
}

func (fakePresence) HeartbeatPresence(context.Context, *userv1.HeartbeatPresenceRequest) (*userv1.HeartbeatPresenceResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (fakePresence) DisconnectPresence(context.Context, *userv1.DisconnectPresenceRequest) (*userv1.DisconnectPresenceResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (fakePresence) SetManualPresence(context.Context, *userv1.SetManualPresenceRequest) (*userv1.SetManualPresenceResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakePresence) GetMyPresence(ctx context.Context, in *userv1.GetMyPresenceRequest) (*userv1.GetMyPresenceResponse, error) {
	if f.getMy != nil {
		return f.getMy(ctx, in)
	}
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func TestGetMyPresenceDegradesWhenMSUserUnavailable(t *testing.T) {
	clients := &microservices.Clients{User: msUser.Service{GRPC: msUser.GRPC{
		Presence: fakePresence{
			getMy: func(context.Context, *userv1.GetMyPresenceRequest) (*userv1.GetMyPresenceResponse, error) {
				return nil, status.Error(codes.Unavailable, "ms-user down")
			},
		},
	}}}
	h := userdelivery.NewHandler(clients, logger.New(logger.Options{Service: "gateway-test", Env: "test", Level: "error"}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me/presence", nil)
	req.Header.Set("Authorization", "Bearer tok")
	rec := httptest.NewRecorder()
	h.GetMyPresence(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"effective_status":"offline"`) || strings.Contains(rec.Body.String(), `"error"`) {
		t.Fatalf("expected degraded envelope, got %s", rec.Body.String())
	}
}
