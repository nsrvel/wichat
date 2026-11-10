package mock_test

import (
	"context"
	"testing"
	"time"

	authmock "github.com/wichat/wichat/backend/ms-user/internal/auth/repository/mock"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
	"github.com/wichat/wichat/backend/ms-user/pkg/token"
)

func TestSessionMockLifecycle(t *testing.T) {
	st := usermock.NewState()
	repo := authmock.NewPgRepositoryMock(st)
	ctx := context.Background()
	now := time.Now().UTC()

	plain, hash, err := token.NewRefresh()
	if err != nil {
		t.Fatal(err)
	}
	sess := models.Session{
		ID: "s1", UserID: "u1", RefreshTokenHash: hash, ExpiresAt: now.Add(time.Hour), CreatedAt: now,
	}
	if err := repo.InsertSession(ctx, sess); err != nil {
		t.Fatal(err)
	}

	got, err := repo.FindSession(ctx, "s1")
	if err != nil || got.ID != "s1" {
		t.Fatalf("%+v %v", got, err)
	}
	got2, err := repo.FindSessionByRefreshHash(ctx, hash, now)
	if err != nil || got2.ID != "s1" {
		t.Fatalf("%+v %v", got2, err)
	}

	_, hash2, _ := token.NewRefresh()
	_ = plain
	if err := repo.TouchSession(ctx, "s1", "r2", hash2, now.Add(2*time.Hour), now); err != nil {
		t.Fatal(err)
	}
	got3, err := repo.FindSessionByRefreshHash(ctx, hash2, now)
	if err != nil || got3.ID != "s1" {
		t.Fatalf("%+v %v", got3, err)
	}

	ids, err := repo.ListSessionIDsForUser(ctx, "u1")
	if err != nil || len(ids) != 1 {
		t.Fatalf("%v %v", ids, err)
	}
	if err := repo.RevokeSession(ctx, "s1", now); err != nil {
		t.Fatal(err)
	}
	if err := repo.RevokeAllSessionsForUser(ctx, "u1", now); err != nil {
		t.Fatal(err)
	}
	if err := repo.DeleteExpired(ctx, now.Add(-time.Minute)); err != nil {
		t.Fatal(err)
	}
}
