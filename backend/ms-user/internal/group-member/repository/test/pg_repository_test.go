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
	groupmemberrepo "github.com/wichat/wichat/backend/ms-user/internal/group-member/repository"
	grouprepo "github.com/wichat/wichat/backend/ms-user/internal/group/repository"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	userrepo "github.com/wichat/wichat/backend/ms-user/internal/user/repository"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

func TestGroupMemberPgRepository(t *testing.T) {
	users, groups, members, cleanup := newGroupMemberPgRepositories(t)
	defer cleanup()
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)

	ownerID := "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	memberID := "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
	_ = users.CreateFirstUser(ctx, models.User{
		ID: ownerID, Email: "owner-mem@example.com", PasswordHash: "hash", DisplayName: "Owner",
		CreatedAt: now, UpdatedAt: now,
	})
	_ = users.CreateUser(ctx, models.User{
		ID: memberID, Email: "member-mem@example.com", PasswordHash: "hash", DisplayName: "Member",
		CreatedAt: now, UpdatedAt: now,
	})

	groupID := "cccccccc-cccc-4ccc-8ccc-cccccccccccc"
	publicID := "507f1f77bcf86cd799439088"
	_ = groups.CreateGroupWithOwner(ctx, models.Group{
		ID: groupID, PublicID: publicID, Name: "Member Group",
		CreatedAt: now, UpdatedAt: now,
	}, models.GroupMember{
		GroupID: groupID, UserID: ownerID, Role: constants.GroupRoleOwner, CreatedAt: now,
	})

	if err := members.InsertGroupMember(ctx, models.GroupMember{
		GroupID: groupID, UserID: memberID, Role: constants.GroupRoleMember, CreatedAt: now.Add(time.Second),
	}); err != nil {
		t.Fatalf("InsertGroupMember: %v", err)
	}

	cnt, err := members.CountGroupMembers(ctx, groupID)
	if err != nil || cnt != 2 {
		t.Fatalf("CountGroupMembers: %d, err %v", cnt, err)
	}

	page, _, err := members.ListGroupMembersPage(ctx, groupID, 10, nil)
	if err != nil || len(page) != 2 {
		t.Fatalf("ListGroupMembersPage: %d, err %v", len(page), err)
	}

	bannedCount, err := members.BanMembers(ctx, groupID, []string{memberID}, ownerID, "spam", 0)
	if err != nil || bannedCount != 1 {
		t.Fatalf("BanMembers: %d, err %v", bannedCount, err)
	}

	banned, err := members.IsMemberBanned(ctx, groupID, memberID)
	if err != nil || !banned {
		t.Fatalf("IsMemberBanned: %v, err %v", banned, err)
	}
}

func newGroupMemberPgRepositories(t *testing.T) (userrepo.PgRepository, grouprepo.PgRepository, groupmemberrepo.PgRepository, func()) {
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

	log := logger.New(logger.Options{Service: "group-member-repo-test", Env: "test", Level: "error"})

	return userrepo.NewPgRepository(db), grouprepo.NewPgRepository(db, log), groupmemberrepo.NewPgRepository(db, log), cleanup
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

	return goose.Up(db, migrationsDir())
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
