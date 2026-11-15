// Integration tests for user-activity-log Postgres repository. Requires Docker; skipped when unavailable.

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
	activityrepo "github.com/wichat/wichat/backend/ms-user/internal/user-activity-log/repository"
	userrepo "github.com/wichat/wichat/backend/ms-user/internal/user/repository"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
)

func TestInsertAndListByUser(t *testing.T) {
	users, logsRepo, cleanup := newActivityLogPgRepos(t)
	defer cleanup()
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)

	userID := "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
	if err := users.CreateFirstUser(ctx, models.User{
		ID:           userID,
		Email:        "audit-pg@example.com",
		PasswordHash: "hash",
		DisplayName:  "Audit PG",
		CreatedAt:    now,
		UpdatedAt:    now,
	}); err != nil {
		t.Fatalf("CreateFirstUser: %v", err)
	}

	err := logsRepo.Insert(ctx, models.UserActivityLog{
		ID:         "11111111-1111-4111-8111-111111111111",
		UserID:     userID,
		ActorID:    userID,
		Category:   constants.ActivityCategoryProfile,
		Action:     constants.ActivityUserProfileUpdated,
		TargetType: constants.ActivityTargetUser,
		TargetID:   userID,
		Changes:    map[string]models.ActivityFieldChange{"display_name": {Old: "Audit PG", New: "Audit Two"}},
		Metadata:   map[string]any{"source": "web"},
		CreatedAt:  now,
	})
	if err != nil {
		t.Fatalf("Insert: %v", err)
	}

	got, err := logsRepo.ListByUser(ctx, userID, constants.ActivityCategoryProfile, "", 10, nil, "")
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}
	if len(got) != 1 || got[0].Action != constants.ActivityUserProfileUpdated || got[0].Changes["display_name"].New != "Audit Two" {
		t.Fatalf("unexpected logs: %+v", got)
	}
}

func newActivityLogPgRepos(t *testing.T) (userrepo.PgRepository, activityrepo.PgRepository, func()) {
	t.Helper()
	if os.Getenv("TESTCONTAINERS_SKIP") == "1" {
		t.Skip("TESTCONTAINERS_SKIP=1")
	}

	ctx := context.Background()
	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("wichat_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Skipf("postgres container unavailable: %v", err)
	}

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = testcontainers.TerminateContainer(pgContainer)
		t.Fatalf("dsn: %v", err)
	}

	if err := runActivityMigrations(dsn); err != nil {
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

	return userrepo.NewPgRepository(db), activityrepo.NewPgRepository(db), cleanup
}

func runActivityMigrations(dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	return goose.Up(db, activityMigrationsDir())
}

func activityMigrationsDir() string {
	if dir := os.Getenv("MS_USER_MIGRATIONS_DIR"); dir != "" {
		return dir
	}
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "../../../../migrations"
	}
	return filepath.Join(filepath.Dir(file), "..", "..", "..", "migrations")
}
