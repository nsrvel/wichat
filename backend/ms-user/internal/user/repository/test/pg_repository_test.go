// Integration tests for user Postgres repository. Requires Docker; skipped when unavailable.

package repository_test

import (
	"context"
	"database/sql"
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
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	userrepo "github.com/wichat/wichat/backend/ms-user/internal/user/repository"
)

func TestFindUserByEmailAfterCreate(t *testing.T) {
	users, cleanup := newUserPgRepository(t)
	defer cleanup()
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)

	user := models.User{
		ID:           "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
		Email:        "user-repo@example.com",
		PasswordHash: "hash",
		DisplayName:  "Repo",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := users.CreateFirstUser(ctx, user); err != nil {
		t.Fatalf("CreateFirstUser: %v", err)
	}

	got, err := users.FindUserByEmail(ctx, "user-repo@example.com")
	if err != nil {
		t.Fatalf("FindUserByEmail: %v", err)
	}
	if got.ID != user.ID {
		t.Fatalf("got %+v", got)
	}

	// TOTP CRUD
	if err := users.UpsertMFATOTP(ctx, models.MFATOTP{
		UserID:          user.ID,
		SecretEncrypted: []byte("sealed-secret"),
		LastUsedStep:    100,
		EnabledAt:       now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}); err != nil {
		t.Fatalf("UpsertMFATOTP: %v", err)
	}
	if err := users.UpdateMFATOTPLastUsedStep(ctx, user.ID, 101, now); err != nil {
		t.Fatalf("UpdateMFATOTPLastUsedStep: %v", err)
	}
	totpRec, err := users.GetMFATOTP(ctx, user.ID)
	if err != nil || totpRec.LastUsedStep != 101 {
		t.Fatalf("GetMFATOTP: %+v %v", totpRec, err)
	}
	if err := users.DeleteMFATOTP(ctx, user.ID); err != nil {
		t.Fatalf("DeleteMFATOTP: %v", err)
	}

	// WebAuthn credentials CRUD
	cred := models.WebAuthnCredential{
		ID:              "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
		UserID:          user.ID,
		Name:            "YubiKey",
		CredentialID:    []byte("cred-1"),
		PublicKey:       []byte("pk-1"),
		AttestationType: "none",
		AAGUID:          []byte{1, 2, 3, 4},
		SignCount:       1,
		Transports:      []string{"usb"},
		CreatedAt:       now,
	}
	if err := users.InsertWebAuthnCredential(ctx, cred); err != nil {
		t.Fatalf("InsertWebAuthnCredential: %v", err)
	}
	if err := users.UpdateWebAuthnCredentialUsage(ctx, cred.CredentialID, 2, false, true, now); err != nil {
		t.Fatalf("UpdateWebAuthnCredentialUsage: %v", err)
	}
	renamed, err := users.RenameWebAuthnCredential(ctx, user.ID, cred.ID, "Backup Key")
	if err != nil || renamed.Name != "Backup Key" {
		t.Fatalf("RenameWebAuthnCredential: %+v %v", renamed, err)
	}
	creds, err := users.ListWebAuthnCredentials(ctx, user.ID)
	if err != nil || len(creds) != 1 || creds[0].SignCount != 2 {
		t.Fatalf("ListWebAuthnCredentials: %+v %v", creds, err)
	}
	if err := users.DeleteWebAuthnCredential(ctx, user.ID, cred.ID); err != nil {
		t.Fatalf("DeleteWebAuthnCredential: %v", err)
	}

	// Recovery codes CRUD
	hashes := [][]byte{[]byte("hash-1"), []byte("hash-2")}
	if err := users.ReplaceRecoveryCodes(ctx, user.ID, hashes, now); err != nil {
		t.Fatalf("ReplaceRecoveryCodes: %v", err)
	}
	cnt, err := users.CountUnusedRecoveryCodes(ctx, user.ID)
	if err != nil || cnt != 2 {
		t.Fatalf("CountUnusedRecoveryCodes: %d %v", cnt, err)
	}
	ok, err := users.ConsumeRecoveryCode(ctx, user.ID, []byte("hash-1"), now)
	if err != nil || !ok {
		t.Fatalf("ConsumeRecoveryCode: %v %v", ok, err)
	}
	cnt2, err := users.CountUnusedRecoveryCodes(ctx, user.ID)
	if err != nil || cnt2 != 1 {
		t.Fatalf("CountUnusedRecoveryCodes after consume: %d %v", cnt2, err)
	}
}

func newUserPgRepository(t *testing.T) (userrepo.PgRepository, func()) {
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
		t.Skipf("docker unavailable: %v", err)
	}

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = testcontainers.TerminateContainer(pgContainer)
		t.Fatalf("dsn: %v", err)
	}

	if err := runUserMigrations(dsn); err != nil {
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

	return userrepo.NewPgRepository(db), cleanup
}

func runUserMigrations(dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	return goose.Up(db, userMigrationsDir())
}

func userMigrationsDir() string {
	if dir := os.Getenv("MS_USER_MIGRATIONS_DIR"); dir != "" {
		return dir
	}
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "internal/migrations"
	}
	return filepath.Join(filepath.Dir(file), "..", "..", "..", "migrations")
}
