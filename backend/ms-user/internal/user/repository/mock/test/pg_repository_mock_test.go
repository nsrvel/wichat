package mock_test

import (
	"context"
	"testing"
	"time"

	"github.com/wichat/wichat/backend/ms-user/internal/models"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	"github.com/wichat/wichat/backend/ms-user/pkg/token"
)

func TestUserMockCRUD(t *testing.T) {
	st := usermock.NewState()
	repo := usermock.NewPgRepositoryMock(st)
	ctx := context.Background()
	now := time.Now().UTC()

	u := models.User{ID: "u1", Email: "a@x.com", Username: "ada", PasswordHash: "h", DisplayName: "Ada", CreatedAt: now, UpdatedAt: now}
	if err := repo.CreateFirstUser(ctx, u); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateFirstUser(ctx, models.User{ID: "u2", Email: "b@x.com"}); err == nil {
		t.Fatal("conflict")
	}

	u2 := models.User{ID: "u2", Email: "b@x.com", Username: "bob", DisplayName: "Bob", CreatedAt: now, UpdatedAt: now}
	if err := repo.CreateUser(ctx, u2); err != nil {
		t.Fatal(err)
	}

	if _, err := repo.FindUserByLogin(ctx, "ADA"); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.FindUserByUsername(ctx, "bob"); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.FindUserByID(ctx, "missing"); err == nil {
		t.Fatal("not found")
	}

	if err := repo.UpdateUserPasswordHash(ctx, "u1", "newhash", now); err != nil {
		t.Fatal(err)
	}

	saved, err := repo.SaveGoogleUser(ctx, models.User{ID: "u3", Email: "g@x.com", DisplayName: "G"}, "", "google", "sub-1")
	if err != nil || saved.ID != "u3" {
		t.Fatalf("%+v %v", saved, err)
	}
	again, err := repo.SaveGoogleUser(ctx, models.User{ID: "u3", Email: "g@x.com"}, "", "google", "sub-1")
	if err != nil || again.ID != "u3" {
		t.Fatalf("link %+v %v", again, err)
	}

	st.SetInstanceSettings(models.InstanceSettings{RegistrationMode: constants.RegistrationOpen})
	got, err := repo.GetInstanceSettings(ctx)
	if err != nil || got.RegistrationMode != constants.RegistrationOpen {
		t.Fatalf("%+v %v", got, err)
	}

	hits, err := repo.SearchUsers(ctx, "ada", 10, "")
	if err != nil || len(hits) == 0 {
		t.Fatalf("search %+v %v", hits, err)
	}

	empty, err := repo.SearchUsers(ctx, "", 10, "")
	if err != nil || len(empty) != 0 {
		t.Fatalf("empty search %+v %v", empty, err)
	}

	if _, err := repo.FindUserByLogin(ctx, "ADA"); err != nil {
		t.Fatal(err)
	}
	if err := repo.UpdateUserPasswordHash(ctx, "missing", "h", now); err == nil {
		t.Fatal("not found")
	}
}

func TestCompleteInstanceSetupOnce(t *testing.T) {
	st := usermock.NewState()
	repo := usermock.NewPgRepositoryMock(st)
	ctx := context.Background()
	now := time.Now().UTC()

	_ = repo.CreateFirstUser(ctx, models.User{ID: "owner", Email: "o@x.com", DisplayName: "O", CreatedAt: now, UpdatedAt: now})
	if err := repo.CompleteInstanceSetup(ctx, "owner", models.InstanceSetupPayload{
		GroupID: "g1", GroupPublicID: "bbbbbbbbbbbbbbbbbbbbbbbb", GroupName: "WiChat",
		RegistrationMode: constants.RegistrationOpen,
	}, now); err != nil {
		t.Fatal(err)
	}
	if err := repo.CompleteInstanceSetup(ctx, "owner", models.InstanceSetupPayload{
		GroupID: "g2", GroupPublicID: "cccccccccccccccccccccccc", GroupName: "Again",
		RegistrationMode: constants.RegistrationOpen,
	}, now); err == nil {
		t.Fatal("conflict")
	}
}

func TestUserMockCountFindEmailDefaultSettings(t *testing.T) {
	st := usermock.NewState()
	repo := usermock.NewPgRepositoryMock(st)
	ctx := context.Background()
	now := time.Now().UTC()

	if n, err := repo.CountUsers(ctx); err != nil || n != 0 {
		t.Fatalf("count %d %v", n, err)
	}

	u := models.User{ID: "u1", Email: "findme@example.com", DisplayName: "F", CreatedAt: now, UpdatedAt: now}
	if err := repo.CreateFirstUser(ctx, u); err != nil {
		t.Fatal(err)
	}
	if n, err := repo.CountUsers(ctx); err != nil || n != 1 {
		t.Fatalf("count after %d %v", n, err)
	}

	got, err := repo.FindUserByEmail(ctx, "findme@example.com")
	if err != nil || got.ID != "u1" {
		t.Fatalf("%+v %v", got, err)
	}
	if _, err := repo.FindUserByEmail(ctx, "missing@example.com"); err == nil {
		t.Fatal("not found")
	}

	def, err := repo.GetInstanceSettings(ctx)
	if err != nil || def.SetupCompleted {
		t.Fatalf("defaults %+v %v", def, err)
	}
}

func TestStateSessionHelpers(t *testing.T) {
	st := usermock.NewState()
	now := time.Now().UTC()
	plain, hash, err := token.NewRefresh()
	if err != nil {
		t.Fatal(err)
	}
	_ = plain

	sess := models.Session{ID: "s1", UserID: "u1", RefreshTokenHash: hash, ExpiresAt: now.Add(time.Hour)}
	if err := st.InsertSessionRecord(sess); err != nil {
		t.Fatal(err)
	}
	got, err := st.FindSession("s1")
	if err != nil || got.ID != "s1" {
		t.Fatal(err)
	}
	_, err = st.FindSessionByRefreshHash(hash, now)
	if err != nil {
		t.Fatal(err)
	}
	_, hash2, _ := token.NewRefresh()
	_ = st.TouchSession("s1", "r2", hash2, now.Add(2*time.Hour), now)
	ids, _ := st.ListSessionIDsForUser("u1")
	if len(ids) != 1 {
		t.Fatalf("ids %v", ids)
	}
	_ = st.RevokeSession("s1", now)
	_ = st.RevokeAllSessionsForUser("u1", now)
	sess2 := models.Session{ID: "s2", UserID: "u2", ExpiresAt: now.Add(-time.Minute)}
	_ = st.InsertSessionRecord(sess2)
	if err := st.DeleteExpiredSessions(now); err != nil {
		t.Fatal(err)
	}
}
