package authHttp_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	authHttp "github.com/wichat/wichat/backend/api-gateway/pkg/auth-http"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

func TestChallengeFromPBStripsEmailHint(t *testing.T) {
	body := authHttp.ChallengeFromPB(&userv1.AuthChallenge{
		ChallengeId:           "ch-1",
		EmailHint:             "a***@example.com",
		ResendCooldownSeconds: 30,
		OtpRequired:           true,
	})
	if body.EmailHint != "" {
		t.Fatalf("email_hint leaked: %q", body.EmailHint)
	}
	if body.ChallengeID != "ch-1" || body.ResendAfterSeconds != 30 {
		t.Fatalf("body: %+v", body)
	}
}

func TestWriteChallengeSetsOTPRequired(t *testing.T) {
	rec := httptest.NewRecorder()
	authHttp.WriteChallenge(rec, &userv1.AuthChallenge{ChallengeId: "ch-2", OtpRequired: false})

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	var parsed struct {
		Data struct {
			OTPRequired bool   `json:"otp_required"`
			EmailHint   string `json:"email_hint"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &parsed); err != nil {
		t.Fatal(err)
	}
	if !parsed.Data.OTPRequired || parsed.Data.EmailHint != "" {
		t.Fatalf("data: %+v", parsed.Data)
	}
}
