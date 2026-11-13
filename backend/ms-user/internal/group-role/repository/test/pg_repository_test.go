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
	grouprolerepo "github.com/wichat/wichat/backend/ms-user/internal/group-role/repository"
	grouprepo "github.com/wichat/wichat/backend/ms-user/internal/group/repository"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	userrepo "github.com/wichat/wichat/backend/ms-user/internal/user/repository"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

func TestGroupRolePgRepository(t *testing.T) {
	users, groups, roles, cleanup := newGroupRolePgRepositories(t)
	defer cleanup()
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)

	ownerID := "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	memberID := "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
	_ = users.CreateFirstUser(ctx, models.User{
		ID: ownerID, Email: "owner-role@example.com", PasswordHash: "hash", DisplayName: "Owner",
		CreatedAt: now, UpdatedAt: now,
	})
	_ = users.CreateUser(ctx, models.User{
		ID: memberID, Email: "member-role@example.com", PasswordHash: "hash", DisplayName: "Member",
		CreatedAt: now, UpdatedAt: now,
	})

	groupID := "cccccccc-cccc-4ccc-8ccc-cccccccccccc"
	publicID := "507f1f77bcf86cd799439099"
	_ = groups.CreateGroupWithOwner(ctx, models.Group{
		ID: groupID, PublicID: publicID, Name: "Role Group",
		CreatedAt: now, UpdatedAt: now,
	}, models.GroupMember{
		GroupID: groupID, UserID: ownerID, Role: constants.GroupRoleOwner, CreatedAt: now,
	})

	list, err := roles.ListRolesForGroup(ctx, groupID)
	if err != nil || len(list) != 2 {
		t.Fatalf("expected 2 roles (@everyone and Member), got %d, err: %v", len(list), err)
	}

	perms, err := roles.GetEffectivePermissions(ctx, groupID, ownerID)
	if err != nil || len(perms) != len(constants.AllKnownPermissions) {
		t.Fatalf("owner should have all permissions: %v, count %d", err, len(perms))
	}

	role, err := roles.FindRoleByID(ctx, groupID, list[1].ID)
	if err != nil {
		t.Fatalf("FindRoleByID: %v", err)
	}

	name := "Moderator"
	patch := models.GroupRolePatch{
		Name:        &name,
		Permissions: []string{constants.PermManageRoles},
	}
	updated, err := roles.UpdateRole(ctx, groupID, role.ID, patch, now)
	if err != nil || updated.Name != "Moderator" {
		t.Fatalf("UpdateRole: %v, updated: %+v", err, updated)
	}

	_ = roles.AddMemberRoles(ctx, groupID, role.ID, []string{memberID})
	_ = roles.RemoveMemberRole(ctx, groupID, role.ID, memberID)
}

func newGroupRolePgRepositories(t *testing.T) (userrepo.PgRepository, grouprepo.PgRepository, grouprolerepo.PgRepository, func()) {
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

	log := logger.New(logger.Options{Service: "group-role-repo-test", Env: "test", Level: "error"})

	return userrepo.NewPgRepository(db), grouprepo.NewPgRepository(db, log), grouprolerepo.NewPgRepository(db, log), cleanup
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
