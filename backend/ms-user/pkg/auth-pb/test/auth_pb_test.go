package authPb_test

import (
	"testing"
	"time"

	"github.com/wichat/wichat/backend/ms-user/internal/models"
	authPb "github.com/wichat/wichat/backend/ms-user/pkg/auth-pb"
)

func TestGroupMemberEntries(t *testing.T) {
	t.Parallel()

	joined := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	out := authPb.GroupMemberEntries([]models.GroupMemberProfile{
		{UserID: "u1", DisplayName: "Ada", Email: "a@x.com", Role: "owner", JoinedAt: joined},
	})
	if len(out) != 1 || out[0].UserId != "u1" || out[0].JoinedAt != joined.Format(time.RFC3339Nano) {
		t.Fatalf("%+v", out)
	}
}

func TestMappersNilAndValues(t *testing.T) {
	t.Parallel()

	if authPb.Challenge(nil) != nil {
		t.Fatal("challenge nil")
	}
	if authPb.Session(nil) != nil {
		t.Fatal("session nil")
	}
	sessOut, chNil := authPb.PasswordAuth(nil)
	if sessOut != nil || chNil != nil {
		t.Fatal("password auth nil")
	}

	ch := authPb.Challenge(&models.AuthChallenge{ChallengeID: "c1", EmailHint: "a@x.com", ResendCooldownSeconds: 30})
	if ch == nil || ch.ChallengeId != "c1" || !ch.OtpRequired {
		t.Fatalf("%+v", ch)
	}

	user := models.PublicUser{ID: "u1", Email: "a@x.com", DisplayName: "Ada", Username: "ada"}
	if authPb.User(user).Username != "ada" {
		t.Fatal("user")
	}

	group := models.PublicGroup{ID: "g1", Name: "Team", LogoURL: "https://x/l.png", Role: "owner"}
	if authPb.GroupSummary(group).Name != "Team" {
		t.Fatal("group")
	}

	channels := authPb.ChannelSummaries([]models.PublicChannel{{ID: "c1", Name: "general", Type: "text"}})
	if len(channels) != 1 || channels[0].Name != "general" {
		t.Fatal("channels")
	}

	groups := authPb.GroupSummaries([]models.PublicGroup{group})
	if len(groups) != 1 {
		t.Fatal("groups")
	}

	hits := authPb.UserSearchHits([]models.UserSearchHit{{UserID: "u1", Email: "a@x.com"}})
	if len(hits) != 1 || hits[0].UserId != "u1" {
		t.Fatal("hits")
	}

	sess := models.SessionResult{AccessToken: "at", RefreshToken: "rt", ExpiresIn: 60, User: user}
	out, chOut := authPb.PasswordAuth(&models.PasswordAuthResult{Session: &sess})
	if chOut != nil || out.AccessToken != "at" {
		t.Fatalf("password auth session: %+v %+v", out, chOut)
	}
	step, chStep := authPb.PasswordAuth(&models.PasswordAuthResult{Challenge: &models.AuthChallenge{ChallengeID: "c2"}})
	if step != nil || chStep.ChallengeId != "c2" {
		t.Fatalf("password auth challenge: %+v %+v", step, chStep)
	}

	meta := authPb.Meta("Mozilla", "127.0.0.1")
	if meta.UserAgent != "Mozilla" || meta.ClientIP != "127.0.0.1" {
		t.Fatalf("%+v", meta)
	}
}
