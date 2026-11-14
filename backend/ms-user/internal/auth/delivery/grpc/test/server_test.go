package test

import (
	"context"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
	authgrpc "github.com/wichat/wichat/backend/ms-user/internal/auth/delivery/grpc"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
	testSupport "github.com/wichat/wichat/backend/ms-user/pkg/test-support"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

func TestAuthGRPCServerFlowAndMFA(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, userSvc, _, _ := testSupport.NewAuthService(t, st, nil)
	srv := authgrpc.NewServer(authSvc)
	ctx := context.Background()

	// Setup / Register + VerifyRegister
	_, _ = srv.Setup(ctx, &userv1.SetupRequest{
		Email:       "owner-grpc@example.com",
		Password:    "correct-horse-battery",
		DisplayName: "Owner",
	})

	regRes, err := srv.Register(ctx, &userv1.RegisterRequest{
		Email:       "member-grpc@example.com",
		Username:    "member.grpc",
		Password:    "correct-horse-battery",
		DisplayName: "Member",
		BirthDate:   "2000-01-01",
		Locale:      "en",
	})
	if err != nil || regRes.GetChallenge() == nil {
		t.Fatalf("Register: %v", err)
	}
	if _, err := srv.ResendRegisterCode(ctx, &userv1.ResendRegisterCodeRequest{ChallengeId: "missing"}); err == nil {
		t.Fatal("expected error for missing register challenge")
	}

	verRegRes, err := srv.VerifyRegister(ctx, &userv1.VerifyRegisterRequest{
		ChallengeId: regRes.GetChallenge().GetChallengeId(),
		Code:        notify.LastCode,
	})
	if err != nil || verRegRes.GetSession() == nil {
		t.Fatalf("VerifyRegister: %v", err)
	}
	sess := verRegRes.GetSession()

	// Me & Refresh
	meRes, err := srv.Me(ctx, &userv1.MeRequest{AccessToken: sess.GetAccessToken()})
	if err != nil || meRes.GetUser().GetEmail() != "member-grpc@example.com" {
		t.Fatalf("Me: %v, res=%+v", err, meRes)
	}
	refRes, err := srv.Refresh(ctx, &userv1.RefreshRequest{RefreshToken: sess.GetRefreshToken()})
	if err != nil || refRes.GetSession() == nil {
		t.Fatalf("Refresh: %v", err)
	}

	// Enable TOTP for MFA login test
	beginRes, err := userSvc.BeginTOTPSetup(ctx, refRes.GetSession().GetAccessToken(), "correct-horse-battery")
	if err != nil {
		t.Fatalf("BeginTOTPSetup: %v", err)
	}
	totpCode, err := totp.GenerateCode(beginRes.Secret, time.Now().UTC())
	if err != nil {
		t.Fatalf("totp.GenerateCode: %v", err)
	}
	recoveryCodes, err := userSvc.ConfirmTOTPSetup(ctx, refRes.GetSession().GetAccessToken(), beginRes.SetupToken, totpCode)
	if err != nil || len(recoveryCodes) == 0 {
		t.Fatalf("ConfirmTOTPSetup: %v", err)
	}

	// Login -> MFA Challenge -> VerifyLoginMFA
	loginRes, err := srv.Login(ctx, &userv1.LoginRequest{
		Login:    "member-grpc@example.com",
		Password: "correct-horse-battery",
		Locale:   "en",
	})
	if err != nil || loginRes.GetChallenge() == nil || !loginRes.GetChallenge().GetMfaRequired() {
		t.Fatalf("Login MFA: %v, res=%+v", err, loginRes)
	}

	if _, err := srv.BeginLoginWebAuthn(ctx, &userv1.BeginLoginWebAuthnRequest{
		ChallengeId: loginRes.GetChallenge().GetChallengeId(),
	}); err == nil {
		t.Fatal("expected error when no webauthn credentials enrolled")
	}

	mfaVerifyRes, err := srv.VerifyLoginMFA(ctx, &userv1.VerifyLoginMFARequest{
		ChallengeId: loginRes.GetChallenge().GetChallengeId(),
		Method:      "recovery",
		Code:        recoveryCodes[0],
	})
	if err != nil || mfaVerifyRes.GetSession() == nil {
		t.Fatalf("VerifyLoginMFA: %v", err)
	}

	// Error paths for remaining RPCs so gRPC wrappers are exercised
	_, _ = srv.VerifyLogin(ctx, &userv1.VerifyLoginRequest{ChallengeId: "bad", Code: "000000"})
	_, _ = srv.ResendLoginCode(ctx, &userv1.ResendLoginCodeRequest{ChallengeId: "bad"})
	_, _ = srv.GoogleStart(ctx, &userv1.GoogleStartRequest{})
	_, _ = srv.GoogleFinish(ctx, &userv1.GoogleFinishRequest{Code: "bad"})
	_, _ = srv.RequestPasswordReset(ctx, &userv1.RequestPasswordResetRequest{Email: "member-grpc@example.com"})
	_, _ = srv.VerifyPasswordReset(ctx, &userv1.VerifyPasswordResetRequest{Email: "member-grpc@example.com", Code: "000000"})
	_, _ = srv.CompletePasswordReset(ctx, &userv1.CompletePasswordResetRequest{ResetToken: "bad"})
	_, _ = srv.Logout(ctx, &userv1.LogoutRequest{RefreshToken: mfaVerifyRes.GetSession().GetRefreshToken()})
}
