package test

import (
	"context"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	activitymock "github.com/wichat/wichat/backend/ms-user/internal/user-activity-log/repository/mock"
	activityservice "github.com/wichat/wichat/backend/ms-user/internal/user-activity-log/service"
	usergrpc "github.com/wichat/wichat/backend/ms-user/internal/user/delivery/grpc"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
	testSupport "github.com/wichat/wichat/backend/ms-user/pkg/test-support"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

func TestUserGRPCServer(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, userSvc, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	_, _, sessions := testSupport.TestRepos(st)
	actState := activitymock.NewState()
	actSvc := activityservice.NewService(
		activitymock.NewPgRepositoryMock(actState),
		sessions,
		[]byte(testSupport.TestJWTSecret),
		logger.New(logger.Options{Service: "user-grpc-test", Env: "test", Level: "error"}),
	)
	userSvc = userSvc.WithActivityLog(actSvc)
	srv := usergrpc.NewServer(userSvc).WithActivityLog(actSvc)
	ctx := context.Background()

	statusRes, err := srv.SetupStatus(ctx, &userv1.SetupStatusRequest{})
	if err != nil || statusRes.GetRegistrationMode() == "" {
		t.Fatalf("SetupStatus: %v, res: %+v", err, statusRes)
	}

	checkRes, err := srv.CheckUsername(ctx, &userv1.CheckUsernameRequest{Username: "fresh.user"})
	if err != nil || !checkRes.GetAvailable() || checkRes.GetNormalized() != "fresh.user" {
		t.Fatalf("CheckUsername: %v, res: %+v", err, checkRes)
	}

	owner := testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email: "user-grpc@example.com", Username: "user.grpc", Password: "correct-horse-battery", DisplayName: "Owner", BirthDate: "2000-01-01",
	})
	_ = testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email: "findme@example.com", Username: "findme", Password: "correct-horse-battery", DisplayName: "FindMe", BirthDate: "2000-01-01",
	})

	group, err := groupSvc.CreateGroup(ctx, owner.AccessToken, "User Search Group")
	if err != nil {
		t.Fatalf("CreateGroup: %v", err)
	}

	searchRes, err := srv.SearchUsers(ctx, &userv1.SearchUsersRequest{
		AccessToken: owner.AccessToken,
		Query:       "findme",
		Limit:       10,
		GroupId:     group.ID,
	})
	if err != nil || len(searchRes.GetUsers()) != 1 {
		t.Fatalf("SearchUsers: %v, len: %d", err, len(searchRes.GetUsers()))
	}

	reqEmailRes, err := srv.RequestEmailChange(ctx, &userv1.RequestEmailChangeRequest{
		AccessToken: owner.AccessToken,
		Locale:      "en",
	})
	if err != nil || reqEmailRes.GetChallengeId() == "" {
		t.Fatalf("RequestEmailChange: %v, res: %+v", err, reqEmailRes)
	}

	if _, err := srv.VerifyEmailChange(ctx, &userv1.VerifyEmailChangeRequest{
		AccessToken: owner.AccessToken,
		ChallengeId: reqEmailRes.GetChallengeId(),
		Code:        "000000",
	}); err == nil {
		t.Fatal("expected VerifyEmailChange to fail on wrong code")
	}

	verEmailRes, err := srv.VerifyEmailChange(ctx, &userv1.VerifyEmailChangeRequest{
		AccessToken: owner.AccessToken,
		ChallengeId: reqEmailRes.GetChallengeId(),
		Code:        notify.LastCode,
	})
	if err != nil || verEmailRes.GetChangeToken() == "" {
		t.Fatalf("VerifyEmailChange: %v, res: %+v", err, verEmailRes)
	}

	compEmailRes, err := srv.CompleteEmailChange(ctx, &userv1.CompleteEmailChangeRequest{
		AccessToken:     owner.AccessToken,
		ChangeToken:     verEmailRes.GetChangeToken(),
		NewEmail:        "owner-new@example.com",
		CurrentPassword: "correct-horse-battery",
		Reason:          "new_email",
	})
	if err != nil || compEmailRes.GetUser().GetEmail() != "owner-new@example.com" {
		t.Fatalf("CompleteEmailChange: %v, res: %+v", err, compEmailRes)
	}

	if _, err := srv.UpdatePassword(ctx, &userv1.UpdatePasswordRequest{
		AccessToken:     owner.AccessToken,
		CurrentPassword: "correct-horse-battery",
		NewPassword:     "new-strong-password-123",
	}); err != nil {
		t.Fatalf("UpdatePassword: %v", err)
	}

	mfaStatus, err := srv.GetMFAStatus(ctx, &userv1.GetMFAStatusRequest{AccessToken: owner.AccessToken})
	if err != nil || mfaStatus.GetTotpEnabled() {
		t.Fatalf("GetMFAStatus: %v, res: %+v", err, mfaStatus)
	}

	totpStart, err := srv.BeginTOTPSetup(ctx, &userv1.BeginTOTPSetupRequest{
		AccessToken:     owner.AccessToken,
		CurrentPassword: "new-strong-password-123",
	})
	if err != nil || totpStart.GetSetupToken() == "" {
		t.Fatalf("BeginTOTPSetup: %v", err)
	}

	totpCode, err := totp.GenerateCode(totpStart.GetSecret(), time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	totpConfirm, err := srv.ConfirmTOTPSetup(ctx, &userv1.ConfirmTOTPSetupRequest{
		AccessToken: owner.AccessToken,
		SetupToken:  totpStart.GetSetupToken(),
		Code:        totpCode,
	})
	if err != nil || len(totpConfirm.GetRecoveryCodes()) == 0 {
		t.Fatalf("ConfirmTOTPSetup: %v", err)
	}

	if _, err := srv.RegenerateRecoveryCodes(ctx, &userv1.RegenerateRecoveryCodesRequest{
		AccessToken:     owner.AccessToken,
		CurrentPassword: "new-strong-password-123",
	}); err != nil {
		t.Fatalf("RegenerateRecoveryCodes: %v", err)
	}

	waStart, err := srv.BeginWebAuthnRegister(ctx, &userv1.BeginWebAuthnRegisterRequest{
		AccessToken:     owner.AccessToken,
		CurrentPassword: "new-strong-password-123",
		Name:            "Test Key",
	})
	if err != nil || waStart.GetChallengeId() == "" {
		t.Fatalf("BeginWebAuthnRegister: %v", err)
	}

	if _, err := srv.DisableTOTP(ctx, &userv1.DisableTOTPRequest{
		AccessToken:     owner.AccessToken,
		CurrentPassword: "new-strong-password-123",
	}); err != nil {
		t.Fatalf("DisableTOTP: %v", err)
	}

	if _, err := srv.FinishWebAuthnRegister(ctx, &userv1.FinishWebAuthnRegisterRequest{
		AccessToken:    owner.AccessToken,
		ChallengeId:    waStart.GetChallengeId(),
		CredentialJson: "{}",
	}); err == nil {
		t.Fatal("expected FinishWebAuthnRegister to fail with empty credential JSON")
	}

	now := time.Now().UTC()
	st.Lock()
	st.WebAuthnCreds = append(st.WebAuthnCreds, models.WebAuthnCredential{
		ID:           "grpc-key-1",
		UserID:       owner.User.ID,
		Name:         "Touch ID",
		CredentialID: []byte("cred-1"),
		PublicKey:    []byte("pk-1"),
		Transports:   []string{"internal"},
		CreatedAt:    now,
		LastUsedAt:   &now,
	})
	st.Unlock()

	if statusWithKey, err := srv.GetMFAStatus(ctx, &userv1.GetMFAStatusRequest{AccessToken: owner.AccessToken}); err != nil || len(statusWithKey.GetSecurityKeys()) != 1 {
		t.Fatalf("GetMFAStatus with key: %v, res=%+v", err, statusWithKey)
	}
	if _, err := srv.RenameWebAuthnCredential(ctx, &userv1.RenameWebAuthnCredentialRequest{
		AccessToken:  owner.AccessToken,
		CredentialId: "grpc-key-1",
		Name:         "Touch ID Pro",
	}); err != nil {
		t.Fatalf("RenameWebAuthnCredential: %v", err)
	}
	if _, err := srv.DeleteWebAuthnCredential(ctx, &userv1.DeleteWebAuthnCredentialRequest{
		AccessToken:     owner.AccessToken,
		CredentialId:    "grpc-key-1",
		CurrentPassword: "new-strong-password-123",
	}); err != nil {
		t.Fatalf("DeleteWebAuthnCredential: %v", err)
	}
	profResp, err := srv.UpdateUserProfile(ctx, &userv1.UpdateUserProfileRequest{
		AccessToken: owner.AccessToken,
		DisplayName: "Owner Updated",
		Pronouns:    "they/them",
		Bio:         "Workspace admin",
		BannerColor: "#336699",
	})
	if err != nil {
		t.Fatalf("UpdateUserProfile: %v", err)
	}
	if got := profResp.GetUser(); got.GetDisplayName() != "Owner Updated" ||
		got.GetPronouns() != "they/them" ||
		got.GetBio() != "Workspace admin" ||
		got.GetBannerColor() != "#336699" ||
		got.GetCreatedAt() == "" {
		t.Fatalf("unexpected UpdateUserProfile response user: %+v", got)
	}
	noteResp, err := srv.UpdateUserStatusNote(ctx, &userv1.UpdateUserStatusNoteRequest{
		AccessToken: owner.AccessToken,
		Text:        "Focused",
		Emoji:       "🎧",
	})
	if err != nil {
		t.Fatalf("UpdateUserStatusNote: %v", err)
	}
	if got := noteResp.GetUser(); got.GetStatusNoteText() != "Focused" || got.GetStatusNoteEmoji() != "🎧" {
		t.Fatalf("unexpected UpdateUserStatusNote response user: %+v", got)
	}
	prefsResp, err := srv.UpdateUserPreferences(ctx, &userv1.UpdateUserPreferencesRequest{
		AccessToken: owner.AccessToken,
		Locale:      "id",
		TimeFormat:  "24h",
	})
	if err != nil {
		t.Fatalf("UpdateUserPreferences: %v", err)
	}
	if got := prefsResp.GetUser(); got.GetLocale() != "id" || got.GetTimeFormat() != "24h" {
		t.Fatalf("unexpected UpdateUserPreferences response user: %+v", got)
	}
	if _, err := srv.UpdateUsername(ctx, &userv1.UpdateUsernameRequest{
		AccessToken:     owner.AccessToken,
		Username:        "owner.updated",
		CurrentPassword: "new-strong-password-123",
	}); err != nil {
		t.Fatalf("UpdateUsername: %v", err)
	}
	if _, err := srv.DisableAccount(ctx, &userv1.DisableAccountRequest{
		AccessToken:     owner.AccessToken,
		CurrentPassword: "new-strong-password-123",
	}); err == nil {
		t.Fatal("expected DisableAccount to fail while user owns a group")
	}
	if _, err := srv.DeleteAccount(ctx, &userv1.DeleteAccountRequest{
		AccessToken:     owner.AccessToken,
		CurrentPassword: "new-strong-password-123",
	}); err == nil {
		t.Fatal("expected DeleteAccount to fail while user owns a group")
	}

	sessionsResp, err := srv.ListSessions(ctx, &userv1.ListSessionsRequest{AccessToken: owner.AccessToken})
	if err != nil || len(sessionsResp.GetSessions()) == 0 || !sessionsResp.GetSessions()[0].GetIsCurrent() {
		t.Fatalf("unexpected ListSessions response: %+v %v", sessionsResp, err)
	}
	if _, err := srv.RevokeSession(ctx, &userv1.RevokeSessionRequest{
		AccessToken:     owner.AccessToken,
		SessionId:       sessionsResp.GetSessions()[0].GetId(),
		CurrentPassword: "new-strong-password-123",
	}); err == nil {
		t.Fatal("expected RevokeSession on current session to fail")
	}
	if revAll, err := srv.RevokeOtherSessions(ctx, &userv1.RevokeOtherSessionsRequest{
		AccessToken:     owner.AccessToken,
		CurrentPassword: "new-strong-password-123",
	}); err != nil || !revAll.GetOk() {
		t.Fatalf("RevokeOtherSessions: %+v %v", revAll, err)
	}

	logsResp, err := srv.ListUserActivityLogs(ctx, &userv1.ListUserActivityLogsRequest{
		AccessToken: owner.AccessToken,
		Limit:       20,
	})
	if err != nil || len(logsResp.GetLogs()) == 0 {
		t.Fatalf("ListUserActivityLogs: %+v %v", logsResp, err)
	}

	other := testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email: "rel-other@example.com", Username: "rel.other", Password: "correct-horse-battery", DisplayName: "RelOther", BirthDate: "2000-01-01",
	})
	if setResp, err := srv.SetUserRelationship(ctx, &userv1.SetUserRelationshipRequest{
		AccessToken: owner.AccessToken, TargetUserId: other.User.ID, Kind: "ignore",
	}); err != nil || !setResp.GetOk() || setResp.GetKind() != "ignore" {
		t.Fatalf("SetUserRelationship: %+v %v", setResp, err)
	}
	if relResp, err := srv.ListUserRelationships(ctx, &userv1.ListUserRelationshipsRequest{
		AccessToken: owner.AccessToken,
	}); err != nil || len(relResp.GetRelationships()) != 1 || relResp.GetRelationships()[0].GetTargetUserId() != other.User.ID {
		t.Fatalf("ListUserRelationships: %+v %v", relResp, err)
	}
}
