// Integration tests for auth Postgres repository. Requires Docker; skipped when unavailable.

package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	authrepo "github.com/wichat/wichat/backend/ms-user/internal/auth/repository"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	userrepo "github.com/wichat/wichat/backend/ms-user/internal/user/repository"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
)

func TestCreateFirstUserConflict(t *testing.T) {
	users, _, _, cleanup := newPgRepositories(t)
	defer cleanup()
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)

	user := models.User{
		ID:           "11111111-1111-4111-8111-111111111111",
		Email:        "ada@example.com",
		PasswordHash: "hash",
		DisplayName:  "Ada",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := users.CreateFirstUser(ctx, user); err != nil {
		t.Fatalf("CreateFirstUser: %v", err)
	}

	got, err := users.FindUserByEmail(ctx, "ada@example.com")
	if err != nil {
		t.Fatalf("FindUserByEmail: %v", err)
	}
	if got.ID != user.ID || got.Email != user.Email || got.PasswordHash != "hash" {
		t.Fatalf("user mismatch: %+v", got)
	}

	dup := user
	dup.ID = "22222222-2222-4222-8222-222222222222"
	if err := users.CreateFirstUser(ctx, dup); !errors.Is(err, constants.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestRefreshRotateTwice(t *testing.T) {
	users, sessions, refresh, cleanup := newPgRepositories(t)
	defer cleanup()
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)

	user := models.User{
		ID:           "33333333-3333-4333-8333-333333333333",
		Email:        "refresh@example.com",
		PasswordHash: "hash",
		DisplayName:  "Refresh",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := users.CreateFirstUser(ctx, user); err != nil {
		t.Fatalf("CreateFirstUser: %v", err)
	}

	hash := []byte("refresh-hash-one")
	session := models.Session{
		ID:               "44444444-4444-4444-8444-444444444444",
		UserID:           user.ID,
		ExpiresAt:        now.Add(24 * time.Hour),
		UserAgent:        "test",
		ClientIP:         "127.0.0.1",
		RefreshID:        "rt-1",
		RefreshTokenHash: hash,
		LastUsedAt:       now,
		CreatedAt:        now,
	}
	refreshRec := models.RefreshToken{
		ID:        "55555555-5555-4555-8555-555555555555",
		SessionID: session.ID,
		TokenHash: hash,
		CreatedAt: now,
	}
	if err := sessions.InsertSession(ctx, session); err != nil {
		t.Fatalf("InsertSession: %v", err)
	}
	if err := refresh.Put(ctx, refreshRec, user.ID, session.ExpiresAt); err != nil {
		t.Fatalf("Put refresh: %v", err)
	}

	nextHash := []byte("refresh-hash-two")
	expires := now.Add(30 * 24 * time.Hour)
	nextRefresh := models.RefreshToken{ID: "rt-2", SessionID: session.ID, TokenHash: nextHash, CreatedAt: now}
	if err := refresh.Rotate(ctx, hash, nextHash, nextRefresh, user.ID, expires, now); err != nil {
		t.Fatalf("first Rotate: %v", err)
	}
	err := refresh.Rotate(ctx, hash, []byte("refresh-hash-three"), models.RefreshToken{
		ID: "rt-3", SessionID: session.ID, TokenHash: []byte("refresh-hash-three"), CreatedAt: now,
	}, user.ID, expires, now.Add(time.Second))
	if !errors.Is(err, constants.ErrAlreadyUsed) {
		t.Fatalf("expected ErrAlreadyUsed, got %v", err)
	}
}

func TestSaveGoogleUserLinksExistingEmail(t *testing.T) {
	users, _, _, cleanup := newPgRepositories(t)
	defer cleanup()
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)

	existing := models.User{
		ID:           "66666666-6666-4666-8666-666666666666",
		Email:        "link@example.com",
		PasswordHash: "hash",
		DisplayName:  "Link",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := users.CreateFirstUser(ctx, existing); err != nil {
		t.Fatalf("CreateFirstUser: %v", err)
	}

	incoming := models.User{
		ID:           "77777777-7777-4777-8777-777777777777",
		Email:        "link@example.com",
		PasswordHash: "",
		DisplayName:  "Google Name",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	got, err := users.SaveGoogleUser(ctx, incoming, "88888888-8888-4888-8888-888888888888", "google", "google-subject-1")
	if err != nil {
		t.Fatalf("SaveGoogleUser: %v", err)
	}
	if got.ID != existing.ID {
		t.Fatalf("expected linked user %q, got %q", existing.ID, got.ID)
	}

	again, err := users.SaveGoogleUser(ctx, incoming, "88888888-8888-4888-8888-888888888888", "google", "google-subject-1")
	if err != nil {
		t.Fatalf("SaveGoogleUser again: %v", err)
	}
	if again.ID != existing.ID {
		t.Fatalf("expected same user on repeat subject")
	}
}

func newPgRepositories(t *testing.T) (userrepo.PgRepository, authrepo.SessionRepository, authrepo.RefreshStore, func()) {
	t.Helper()
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.BasicWaitStrategies(),
		postgres.WithDatabase("wichat"),
		postgres.WithUsername("wichat"),
		postgres.WithPassword("wichat"),
	)
	if err != nil {
		t.Skipf("docker postgres unavailable: %v", err)
	}

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = testcontainers.TerminateContainer(pgContainer)
		t.Fatalf("connection string: %v", err)
	}

	if err := runMigrations(dsn); err != nil {
		_ = testcontainers.TerminateContainer(pgContainer)
		t.Fatalf("migrations: %v", err)
	}

	sqldb, err := sql.Open("pgx", dsn)
	if err != nil {
		_ = testcontainers.TerminateContainer(pgContainer)
		t.Fatalf("sql open: %v", err)
	}

	db := bun.NewDB(sqldb, pgdialect.New())

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	cleanup := func() {
		_ = db.Close()
		_ = sqldb.Close()
		mr.Close()
		if err := testcontainers.TerminateContainer(pgContainer); err != nil {
			t.Logf("terminate container: %v", err)
		}
	}

	return userrepo.NewPgRepository(db), authrepo.NewSessionRepository(db), authrepo.NewRedisRefreshStore(rdb), cleanup
}

func runMigrations(dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	dir := migrationsDir()
	return goose.Up(db, dir)
}

func migrationsDir() string {
	if dir := os.Getenv("MS_USER_MIGRATIONS_DIR"); dir != "" {
		return dir
	}
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "internal/migrations"
	}
	return filepath.Join(filepath.Dir(file), "..", "..", "..", "migrations")
}
