package user_test

import (
	"context"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user/grpc/user"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type setupStatusServer struct {
	userv1.UnimplementedUserServiceServer
	calls atomic.Int32
}

func (s *setupStatusServer) SetupStatus(context.Context, *userv1.SetupStatusRequest) (*userv1.SetupStatusResponse, error) {
	if s.calls.Add(1) == 1 {
		return nil, status.Error(codes.Unavailable, "try again")
	}

	return &userv1.SetupStatusResponse{RegistrationMode: "invite_only"}, nil
}

func (s *setupStatusServer) CheckUsername(_ context.Context, in *userv1.CheckUsernameRequest) (*userv1.CheckUsernameResponse, error) {
	return &userv1.CheckUsernameResponse{Available: true, Normalized: in.GetUsername()}, nil
}

func (s *setupStatusServer) UploadUserAvatar(_ context.Context, in *userv1.UploadUserAvatarRequest) (*userv1.UploadUserAvatarResponse, error) {
	return &userv1.UploadUserAvatarResponse{User: &userv1.User{Id: "u-1", AvatarUrl: "/media/u-1.png"}}, nil
}

func (s *setupStatusServer) UpdateUserProfile(_ context.Context, in *userv1.UpdateUserProfileRequest) (*userv1.UpdateUserProfileResponse, error) {
	return &userv1.UpdateUserProfileResponse{User: &userv1.User{Id: "u-1", DisplayName: in.GetDisplayName()}}, nil
}

func (s *setupStatusServer) UpdateUsername(_ context.Context, in *userv1.UpdateUsernameRequest) (*userv1.UpdateUsernameResponse, error) {
	return &userv1.UpdateUsernameResponse{User: &userv1.User{Id: "u-1", Username: in.GetUsername()}}, nil
}

func (s *setupStatusServer) RequestEmailChange(_ context.Context, _ *userv1.RequestEmailChangeRequest) (*userv1.RequestEmailChangeResponse, error) {
	return &userv1.RequestEmailChangeResponse{ChallengeId: "ch-1", EmailHint: "a***z@x.com", ResendCooldownSeconds: 60}, nil
}

func (s *setupStatusServer) VerifyEmailChange(_ context.Context, _ *userv1.VerifyEmailChangeRequest) (*userv1.VerifyEmailChangeResponse, error) {
	return &userv1.VerifyEmailChangeResponse{ChangeToken: "tok-1"}, nil
}

func (s *setupStatusServer) CompleteEmailChange(_ context.Context, in *userv1.CompleteEmailChangeRequest) (*userv1.CompleteEmailChangeResponse, error) {
	return &userv1.CompleteEmailChangeResponse{User: &userv1.User{Id: "u-1", Email: in.GetNewEmail()}}, nil
}

func (s *setupStatusServer) UpdatePassword(_ context.Context, _ *userv1.UpdatePasswordRequest) (*userv1.UpdatePasswordResponse, error) {
	return &userv1.UpdatePasswordResponse{}, nil
}

func (s *setupStatusServer) GetMFAStatus(_ context.Context, _ *userv1.GetMFAStatusRequest) (*userv1.GetMFAStatusResponse, error) {
	return &userv1.GetMFAStatusResponse{TotpEnabled: true, RecoveryCodesRemaining: 10}, nil
}

func (s *setupStatusServer) BeginTOTPSetup(_ context.Context, _ *userv1.BeginTOTPSetupRequest) (*userv1.BeginTOTPSetupResponse, error) {
	return &userv1.BeginTOTPSetupResponse{SetupToken: "st-1", Secret: "SEC", OtpauthUri: "otpauth://totp/WiChat:u"}, nil
}

func (s *setupStatusServer) ConfirmTOTPSetup(_ context.Context, _ *userv1.ConfirmTOTPSetupRequest) (*userv1.ConfirmTOTPSetupResponse, error) {
	return &userv1.ConfirmTOTPSetupResponse{RecoveryCodes: []string{"AAAA-BBBB"}}, nil
}

func (s *setupStatusServer) DisableTOTP(_ context.Context, _ *userv1.DisableTOTPRequest) (*userv1.DisableTOTPResponse, error) {
	return &userv1.DisableTOTPResponse{}, nil
}

func (s *setupStatusServer) BeginWebAuthnRegister(_ context.Context, _ *userv1.BeginWebAuthnRegisterRequest) (*userv1.BeginWebAuthnRegisterResponse, error) {
	return &userv1.BeginWebAuthnRegisterResponse{ChallengeId: "wc-1", OptionsJson: `{}`}, nil
}

func (s *setupStatusServer) FinishWebAuthnRegister(_ context.Context, _ *userv1.FinishWebAuthnRegisterRequest) (*userv1.FinishWebAuthnRegisterResponse, error) {
	return &userv1.FinishWebAuthnRegisterResponse{SecurityKey: &userv1.SecurityKey{Id: "k-1", Name: "YubiKey"}}, nil
}

func (s *setupStatusServer) RenameWebAuthnCredential(_ context.Context, in *userv1.RenameWebAuthnCredentialRequest) (*userv1.RenameWebAuthnCredentialResponse, error) {
	return &userv1.RenameWebAuthnCredentialResponse{SecurityKey: &userv1.SecurityKey{Id: in.GetCredentialId(), Name: in.GetName()}}, nil
}

func (s *setupStatusServer) DeleteWebAuthnCredential(_ context.Context, _ *userv1.DeleteWebAuthnCredentialRequest) (*userv1.DeleteWebAuthnCredentialResponse, error) {
	return &userv1.DeleteWebAuthnCredentialResponse{}, nil
}

func (s *setupStatusServer) RegenerateRecoveryCodes(_ context.Context, _ *userv1.RegenerateRecoveryCodesRequest) (*userv1.RegenerateRecoveryCodesResponse, error) {
	return &userv1.RegenerateRecoveryCodesResponse{RecoveryCodes: []string{"CCCC-DDDD"}}, nil
}

func (s *setupStatusServer) DisableAccount(_ context.Context, _ *userv1.DisableAccountRequest) (*userv1.DisableAccountResponse, error) {
	return &userv1.DisableAccountResponse{Ok: true}, nil
}

func (s *setupStatusServer) DeleteAccount(_ context.Context, _ *userv1.DeleteAccountRequest) (*userv1.DeleteAccountResponse, error) {
	return &userv1.DeleteAccountResponse{Ok: true}, nil
}

func (s *setupStatusServer) ListSessions(_ context.Context, _ *userv1.ListSessionsRequest) (*userv1.ListSessionsResponse, error) {
	return &userv1.ListSessionsResponse{Sessions: []*userv1.UserSession{{Id: "sess-1", IsCurrent: true}}}, nil
}

func (s *setupStatusServer) RevokeSession(_ context.Context, _ *userv1.RevokeSessionRequest) (*userv1.RevokeSessionResponse, error) {
	return &userv1.RevokeSessionResponse{Ok: true}, nil
}

func (s *setupStatusServer) RevokeOtherSessions(_ context.Context, _ *userv1.RevokeOtherSessionsRequest) (*userv1.RevokeOtherSessionsResponse, error) {
	return &userv1.RevokeOtherSessionsResponse{Ok: true}, nil
}

func TestSetupStatusRetriesIdempotent(t *testing.T) {
	const bufSize = 1024 * 1024
	lis := bufconn.Listen(bufSize)
	srv := grpc.NewServer()
	userv1.RegisterUserServiceServer(srv, &setupStatusServer{})
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(func() { srv.Stop() })

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	settings := utils.Settings{
		AttemptTimeoutMs:        500,
		OverallTimeoutMs:        2000,
		MaxRetries:              2,
		RetryBackoffMs:          1,
		BreakerFailureThreshold: 0,
	}
	grpcClient := user.New(userv1.NewUserServiceClient(conn), settings, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	out, err := grpcClient.SetupStatus(ctx, &userv1.SetupStatusRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if out.GetRegistrationMode() != "invite_only" {
		t.Fatalf("unexpected response: %+v", out)
	}

	chk, err := grpcClient.CheckUsername(ctx, &userv1.CheckUsernameRequest{Username: "alice"})
	if err != nil || !chk.GetAvailable() || chk.GetNormalized() != "alice" {
		t.Fatalf("unexpected check username: %+v %v", chk, err)
	}

	av, err := grpcClient.UploadUserAvatar(ctx, &userv1.UploadUserAvatarRequest{AccessToken: "tok"})
	if err != nil || av.GetUser().GetAvatarUrl() != "/media/u-1.png" {
		t.Fatalf("unexpected upload avatar: %+v %v", av, err)
	}

	prof, err := grpcClient.UpdateUserProfile(ctx, &userv1.UpdateUserProfileRequest{AccessToken: "tok", DisplayName: "Alice"})
	if err != nil || prof.GetUser().GetDisplayName() != "Alice" {
		t.Fatalf("unexpected update profile: %+v %v", prof, err)
	}

	un, err := grpcClient.UpdateUsername(ctx, &userv1.UpdateUsernameRequest{AccessToken: "tok", Username: "alice.new"})
	if err != nil || un.GetUser().GetUsername() != "alice.new" {
		t.Fatalf("unexpected update username: %+v %v", un, err)
	}

	reqEmail, err := grpcClient.RequestEmailChange(ctx, &userv1.RequestEmailChangeRequest{AccessToken: "tok", Locale: "en"})
	if err != nil || reqEmail.GetChallengeId() != "ch-1" {
		t.Fatalf("unexpected request email change: %+v %v", reqEmail, err)
	}

	verEmail, err := grpcClient.VerifyEmailChange(ctx, &userv1.VerifyEmailChangeRequest{AccessToken: "tok", ChallengeId: "ch-1", Code: "123456"})
	if err != nil || verEmail.GetChangeToken() != "tok-1" {
		t.Fatalf("unexpected verify email change: %+v %v", verEmail, err)
	}

	compEmail, err := grpcClient.CompleteEmailChange(ctx, &userv1.CompleteEmailChangeRequest{AccessToken: "tok", ChangeToken: "tok-1", NewEmail: "new@x.com"})
	if err != nil || compEmail.GetUser().GetEmail() != "new@x.com" {
		t.Fatalf("unexpected complete email change: %+v %v", compEmail, err)
	}

	if _, err := grpcClient.UpdatePassword(ctx, &userv1.UpdatePasswordRequest{AccessToken: "tok", CurrentPassword: "old", NewPassword: "new-password-12"}); err != nil {
		t.Fatalf("unexpected update password: %v", err)
	}

	mfaStatus, err := grpcClient.GetMFAStatus(ctx, &userv1.GetMFAStatusRequest{AccessToken: "tok"})
	if err != nil || !mfaStatus.GetTotpEnabled() || mfaStatus.GetRecoveryCodesRemaining() != 10 {
		t.Fatalf("unexpected get mfa status: %+v %v", mfaStatus, err)
	}

	totpBegin, err := grpcClient.BeginTOTPSetup(ctx, &userv1.BeginTOTPSetupRequest{AccessToken: "tok", CurrentPassword: "pw"})
	if err != nil || totpBegin.GetSetupToken() != "st-1" {
		t.Fatalf("unexpected begin totp: %+v %v", totpBegin, err)
	}

	totpConfirm, err := grpcClient.ConfirmTOTPSetup(ctx, &userv1.ConfirmTOTPSetupRequest{AccessToken: "tok", SetupToken: "st-1", Code: "123456"})
	if err != nil || len(totpConfirm.GetRecoveryCodes()) != 1 {
		t.Fatalf("unexpected confirm totp: %+v %v", totpConfirm, err)
	}

	if _, err := grpcClient.DisableTOTP(ctx, &userv1.DisableTOTPRequest{AccessToken: "tok", CurrentPassword: "pw"}); err != nil {
		t.Fatalf("unexpected disable totp: %v", err)
	}

	waBegin, err := grpcClient.BeginWebAuthnRegister(ctx, &userv1.BeginWebAuthnRegisterRequest{AccessToken: "tok", CurrentPassword: "pw", Name: "Key"})
	if err != nil || waBegin.GetChallengeId() != "wc-1" {
		t.Fatalf("unexpected begin webauthn: %+v %v", waBegin, err)
	}

	waFinish, err := grpcClient.FinishWebAuthnRegister(ctx, &userv1.FinishWebAuthnRegisterRequest{AccessToken: "tok", ChallengeId: "wc-1", CredentialJson: `{}`})
	if err != nil || waFinish.GetSecurityKey().GetId() != "k-1" {
		t.Fatalf("unexpected finish webauthn: %+v %v", waFinish, err)
	}

	waRename, err := grpcClient.RenameWebAuthnCredential(ctx, &userv1.RenameWebAuthnCredentialRequest{AccessToken: "tok", CredentialId: "k-1", Name: "Backup Key"})
	if err != nil || waRename.GetSecurityKey().GetName() != "Backup Key" {
		t.Fatalf("unexpected rename webauthn: %+v %v", waRename, err)
	}

	if _, err := grpcClient.DeleteWebAuthnCredential(ctx, &userv1.DeleteWebAuthnCredentialRequest{AccessToken: "tok", CredentialId: "k-1", CurrentPassword: "pw"}); err != nil {
		t.Fatalf("unexpected delete webauthn: %v", err)
	}

	regen, err := grpcClient.RegenerateRecoveryCodes(ctx, &userv1.RegenerateRecoveryCodesRequest{AccessToken: "tok", CurrentPassword: "pw"})
	if err != nil || len(regen.GetRecoveryCodes()) != 1 {
		t.Fatalf("unexpected regenerate recovery codes: %+v %v", regen, err)
	}

	dis, err := grpcClient.DisableAccount(ctx, &userv1.DisableAccountRequest{AccessToken: "tok", CurrentPassword: "pw"})
	if err != nil || !dis.GetOk() {
		t.Fatalf("unexpected disable account: %+v %v", dis, err)
	}

	del, err := grpcClient.DeleteAccount(ctx, &userv1.DeleteAccountRequest{AccessToken: "tok", CurrentPassword: "pw"})
	if err != nil || !del.GetOk() {
		t.Fatalf("unexpected delete account: %+v %v", del, err)
	}

	listSess, err := grpcClient.ListSessions(ctx, &userv1.ListSessionsRequest{AccessToken: "tok"})
	if err != nil || len(listSess.GetSessions()) != 1 || !listSess.GetSessions()[0].GetIsCurrent() {
		t.Fatalf("unexpected list sessions: %+v %v", listSess, err)
	}

	revSess, err := grpcClient.RevokeSession(ctx, &userv1.RevokeSessionRequest{AccessToken: "tok", SessionId: "sess-2", CurrentPassword: "pw"})
	if err != nil || !revSess.GetOk() {
		t.Fatalf("unexpected revoke session: %+v %v", revSess, err)
	}

	revOther, err := grpcClient.RevokeOtherSessions(ctx, &userv1.RevokeOtherSessionsRequest{AccessToken: "tok", CurrentPassword: "pw"})
	if err != nil || !revOther.GetOk() {
		t.Fatalf("unexpected revoke other sessions: %+v %v", revOther, err)
	}
}
