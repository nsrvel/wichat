package access_test

import (
	"context"
	"testing"
	"time"

	authmock "github.com/wichat/wichat/backend/ms-user/internal/auth/repository/mock"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
	"github.com/wichat/wichat/backend/ms-user/pkg/access"
	"github.com/wichat/wichat/backend/ms-user/pkg/jwt"
	testSupport "github.com/wichat/wichat/backend/ms-user/pkg/test-support"
)

func TestUserIDFromAccessToken(t *testing.T) {
	t.Parallel()

	st := usermock.NewState()
	now := time.Now().UTC()
	secret := []byte(testSupport.TestJWTSecret)
	sessionID := "sess-1"
	userID := "user-1"
	st.Sessions = []models.Session{{ID: sessionID, UserID: userID, ExpiresAt: now.Add(time.Hour)}}

	raw, _, err := jwt.Issue(secret, userID, sessionID, now, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := access.UserIDFromAccessToken(context.Background(), raw, secret, authmock.NewPgRepositoryMock(st), now)
	if err != nil || got != userID {
		t.Fatalf("got %q %v", got, err)
	}
}

func TestUserIDFromAccessTokenRejectsEmpty(t *testing.T) {
	t.Parallel()

	_, err := access.UserIDFromAccessToken(context.Background(), "", []byte("x"), authmock.NewPgRepositoryMock(usermock.NewState()), time.Now().UTC())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUserIDFromAccessTokenRejectsBadJWT(t *testing.T) {
	t.Parallel()

	_, err := access.UserIDFromAccessToken(context.Background(), "not.a.jwt", []byte(testSupport.TestJWTSecret), authmock.NewPgRepositoryMock(usermock.NewState()), time.Now().UTC())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUserIDFromAccessTokenRejectsExpiredSession(t *testing.T) {
	t.Parallel()

	st := usermock.NewState()
	now := time.Now().UTC()
	secret := []byte(testSupport.TestJWTSecret)
	sessionID := "sess-exp"
	userID := "user-exp"
	st.Sessions = []models.Session{{ID: sessionID, UserID: userID, ExpiresAt: now.Add(-time.Minute)}}

	raw, _, err := jwt.Issue(secret, userID, sessionID, now, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	_, err = access.UserIDFromAccessToken(context.Background(), raw, secret, authmock.NewPgRepositoryMock(st), now)
	if err == nil {
		t.Fatal("expired session")
	}
}

func TestUserFromAccessTokenRejectsMissingUser(t *testing.T) {
	t.Parallel()

	st := usermock.NewState()
	now := time.Now().UTC()
	secret := []byte(testSupport.TestJWTSecret)
	sessionID := "sess-3"
	userID := "ghost"
	st.Sessions = []models.Session{{ID: sessionID, UserID: userID, ExpiresAt: now.Add(time.Hour)}}

	raw, _, err := jwt.Issue(secret, userID, sessionID, now, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	_, err = access.UserFromAccessToken(context.Background(), raw, secret, authmock.NewPgRepositoryMock(st), usermock.NewPgRepositoryMock(st), now)
	if err == nil {
		t.Fatal("missing user")
	}
}

func TestUserFromAccessToken(t *testing.T) {
	t.Parallel()

	st := usermock.NewState()
	now := time.Now().UTC()
	secret := []byte(testSupport.TestJWTSecret)
	sessionID := "sess-2"
	userID := "user-2"
	st.Sessions = []models.Session{{ID: sessionID, UserID: userID, ExpiresAt: now.Add(time.Hour)}}
	st.Users = []models.User{{ID: userID, Email: "a@x.com", DisplayName: "Ada"}}

	raw, _, err := jwt.Issue(secret, userID, sessionID, now, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	user, err := access.UserFromAccessToken(context.Background(), raw, secret, authmock.NewPgRepositoryMock(st), usermock.NewPgRepositoryMock(st), now)
	if err != nil || user.ID != userID {
		t.Fatalf("%+v %v", user, err)
	}
}

func TestUserFromAccessTokenMissingUser(t *testing.T) {
	st := usermock.NewState()
	users, _, sessions := testSupport.TestRepos(st)
	now := time.Now().UTC()
	secret := []byte(testSupport.TestJWTSecret)

	sess := models.Session{
		ID: "s1", UserID: "gone", ExpiresAt: now.Add(time.Hour), CreatedAt: now, LastUsedAt: now,
	}
	_ = sessions.InsertSession(context.Background(), sess)
	tok, _, err := jwt.Issue(secret, "gone", "s1", now, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	_, err = access.UserFromAccessToken(context.Background(), tok, secret, sessions, users, now)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUserIDRejectsRevokedSession(t *testing.T) {
	t.Parallel()

	st := usermock.NewState()
	now := time.Now().UTC()
	secret := []byte(testSupport.TestJWTSecret)
	sessionID := "sess-r"
	userID := "user-r"
	revoked := now
	st.Sessions = []models.Session{{ID: sessionID, UserID: userID, ExpiresAt: now.Add(time.Hour), RevokedAt: &revoked}}

	raw, _, err := jwt.Issue(secret, userID, sessionID, now, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	_, err = access.UserIDFromAccessToken(context.Background(), raw, secret, authmock.NewPgRepositoryMock(st), now)
	if err == nil {
		t.Fatal("revoked")
	}
}

func TestUserIDRejectsMismatchedJWTUser(t *testing.T) {
	t.Parallel()

	st := usermock.NewState()
	now := time.Now().UTC()
	secret := []byte(testSupport.TestJWTSecret)
	sessionID := "sess-m"
	st.Sessions = []models.Session{{ID: sessionID, UserID: "real-user", ExpiresAt: now.Add(time.Hour)}}

	raw, _, err := jwt.Issue(secret, "wrong-user", sessionID, now, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	_, err = access.UserIDFromAccessToken(context.Background(), raw, secret, authmock.NewPgRepositoryMock(st), now)
	if err == nil {
		t.Fatal("mismatch")
	}
}
