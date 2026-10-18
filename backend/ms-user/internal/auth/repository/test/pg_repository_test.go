// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Garam

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

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/wichat/wichat/backend/ms-user/internal/auth/repository"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
)

func TestCreateFirstUserConflict(t *testing.T) {
	repo, cleanup := newPgRepository(t)
	defer cleanup()
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)

	user := models.User{
		ID:           "user-1",
		Email:        "ada@example.com",
		PasswordHash: "hash",
		DisplayName:  "Ada",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := repo.CreateFirstUser(ctx, user); err != nil {
		t.Fatalf("CreateFirstUser: %v", err)
	}

	got, err := repo.FindUserByEmail(ctx, "ada@example.com")
	if err != nil {
		t.Fatalf("FindUserByEmail: %v", err)
	}
	if got.ID != user.ID || got.Email != user.Email {
		t.Fatalf("user mismatch: %+v", got)
	}

	dup := user
	dup.ID = "user-2"
	if err := repo.CreateFirstUser(ctx, dup); !errors.Is(err, repository.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestMarkRefreshUsedTwice(t *testing.T) {
	repo, cleanup := newPgRepository(t)
	defer cleanup()
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)

	user := models.User{
		ID:           "user-refresh",
		Email:        "refresh@example.com",
		PasswordHash: "hash",
		DisplayName:  "Refresh",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := repo.CreateFirstUser(ctx, user); err != nil {
		t.Fatalf("CreateFirstUser: %v", err)
	}

	session := models.Session{
		ID:         "sess-1",
		UserID:     user.ID,
		ExpiresAt:  now.Add(24 * time.Hour),
		UserAgent:  "test",
		ClientIP:   "127.0.0.1",
		LastUsedAt: now,
		CreatedAt:  now,
	}
	hash := []byte("refresh-hash-one")
	refresh := models.RefreshToken{
		ID:        "rt-1",
		SessionID: session.ID,
		TokenHash: hash,
		CreatedAt: now,
	}
	if err := repo.InsertSession(ctx, session, refresh); err != nil {
		t.Fatalf("InsertSession: %v", err)
	}

	nextHash := []byte("refresh-hash-two")
	expires := now.Add(30 * 24 * time.Hour)
	if err := repo.MarkRefreshUsed(ctx, hash, nextHash, "rt-2", session.ID, expires, now); err != nil {
		t.Fatalf("first MarkRefreshUsed: %v", err)
	}
	err := repo.MarkRefreshUsed(ctx, hash, []byte("refresh-hash-three"), "rt-3", session.ID, expires, now.Add(time.Second))
	if !errors.Is(err, repository.ErrAlreadyUsed) {
		t.Fatalf("expected ErrAlreadyUsed, got %v", err)
	}
}

func TestSaveGoogleUserLinksExistingEmail(t *testing.T) {
	repo, cleanup := newPgRepository(t)
	defer cleanup()
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)

	existing := models.User{
		ID:           "user-google",
		Email:        "link@example.com",
		PasswordHash: "hash",
		DisplayName:  "Link",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := repo.CreateFirstUser(ctx, existing); err != nil {
		t.Fatalf("CreateFirstUser: %v", err)
	}

	incoming := models.User{
		ID:           "user-should-not-insert",
		Email:        "link@example.com",
		PasswordHash: "",
		DisplayName:  "Google Name",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	got, err := repo.SaveGoogleUser(ctx, incoming, "id-1", "google", "google-subject-1")
	if err != nil {
		t.Fatalf("SaveGoogleUser: %v", err)
	}
	if got.ID != existing.ID {
		t.Fatalf("expected linked user %q, got %q", existing.ID, got.ID)
	}

	again, err := repo.SaveGoogleUser(ctx, incoming, "id-1", "google", "google-subject-1")
	if err != nil {
		t.Fatalf("SaveGoogleUser again: %v", err)
	}
	if again.ID != existing.ID {
		t.Fatalf("expected same user on repeat subject")
	}
}

func newPgRepository(t *testing.T) (repository.PgRepository, func()) {
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

	cleanup := func() {
		_ = db.Close()
		_ = sqldb.Close()
		if err := testcontainers.TerminateContainer(pgContainer); err != nil {
			t.Logf("terminate container: %v", err)
		}
	}

	return repository.NewPgRepository(db), cleanup
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
