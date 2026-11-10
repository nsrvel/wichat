// Integration tests for group Postgres repository. Requires Docker; skipped when unavailable.

package repository_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	channelrepo "github.com/wichat/wichat/backend/ms-user/internal/channel/repository"
	grouprepo "github.com/wichat/wichat/backend/ms-user/internal/group/repository"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	userrepo "github.com/wichat/wichat/backend/ms-user/internal/user/repository"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

func TestCreateGroupWithOwnerListAndJoinable(t *testing.T) {
	users, groups, channels, cleanup := newGroupPgRepositories(t)
	defer cleanup()
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)

	ownerID := "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	memberID := "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
	owner := models.User{
		ID: ownerID, Email: "owner@example.com", PasswordHash: "hash", DisplayName: "Owner",
		CreatedAt: now, UpdatedAt: now,
	}
	member := models.User{
		ID: memberID, Email: "member@example.com", PasswordHash: "hash", DisplayName: "Member",
		CreatedAt: now, UpdatedAt: now,
	}
	if err := users.CreateFirstUser(ctx, owner); err != nil {
		t.Fatalf("CreateFirstUser owner: %v", err)
	}
	if err := users.CreateUser(ctx, member); err != nil {
		t.Fatalf("CreateUser member: %v", err)
	}

	groupID := "cccccccc-cccc-4ccc-8ccc-cccccccccccc"
	publicID := "507f1f77bcf86cd799439011"
	group := models.Group{
		ID: groupID, PublicID: publicID, Name: "Acme", LogoURL: "",
		CreatedAt: now, UpdatedAt: now,
	}
	ownerMember := models.GroupMember{
		GroupID: groupID, UserID: ownerID, Role: constants.GroupRoleOwner, CreatedAt: now,
	}
	if err := groups.CreateGroupWithOwner(ctx, group, ownerMember); err != nil {
		t.Fatalf("CreateGroupWithOwner: %v", err)
	}
	if err := channels.SeedDefaultChannels(ctx, groupID, ownerID, now); err != nil {
		t.Fatalf("SeedDefaultChannels: %v", err)
	}

	listed, err := channels.ListByGroupID(ctx, groupID)
	if err != nil {
		t.Fatalf("ListByGroupID: %v", err)
	}
	if len(listed) != 2 {
		t.Fatalf("channels after create: got %d want 2: %+v", len(listed), listed)
	}
	names := map[string]struct{}{listed[0].Name: {}, listed[1].Name: {}}
	if _, ok := names[constants.DefaultChannelName]; !ok {
		t.Fatalf("missing %q: %+v", constants.DefaultChannelName, listed)
	}
	if _, ok := names[constants.DefaultAnnouncementChannelName]; !ok {
		t.Fatalf("missing %q: %+v", constants.DefaultAnnouncementChannelName, listed)
	}
	for _, ch := range listed {
		switch ch.Name {
		case constants.DefaultChannelName:
			if ch.SidebarSection != constants.SidebarSectionChannels {
				t.Fatalf("general sidebar_section: %+v", ch)
			}
		case constants.DefaultAnnouncementChannelName:
			if ch.SidebarSection != constants.SidebarSectionFavorites {
				t.Fatalf("announcements sidebar_section: %+v", ch)
			}
		}
	}

	mine, err := groups.ListGroupsForUser(ctx, ownerID)
	if err != nil || len(mine) != 1 {
		t.Fatalf("ListGroupsForUser owner: %+v err %v", mine, err)
	}
	if mine[0].Name != "Acme" || mine[0].PublicID != publicID {
		t.Fatalf("group: %+v", mine[0])
	}
	if mine[0].CallerRole != constants.GroupRoleOwner {
		t.Fatalf("caller role: %q", mine[0].CallerRole)
	}
	if !regexp.MustCompile(`^[0-9a-f]{24}$`).MatchString(mine[0].DefaultChannelPublicID) {
		t.Fatalf("default channel id: %q", mine[0].DefaultChannelPublicID)
	}
	if !regexp.MustCompile(`^[0-9a-f]{24}$`).MatchString(mine[0].PublicID) {
		t.Fatalf("public id format: %q", mine[0].PublicID)
	}

	joinable, err := groups.ListJoinableGroups(ctx, memberID)
	if err != nil || len(joinable) != 1 || joinable[0].PublicID != publicID {
		t.Fatalf("ListJoinableGroups: %+v err %v", joinable, err)
	}

	if err := groups.InsertGroupMember(ctx, models.GroupMember{
		GroupID: groupID, UserID: memberID, Role: constants.GroupRoleMember, CreatedAt: now,
	}); err != nil {
		t.Fatalf("InsertGroupMember: %v", err)
	}

	joinable, err = groups.ListJoinableGroups(ctx, memberID)
	if err != nil || len(joinable) != 0 {
		t.Fatalf("joinable after join: %+v err %v", joinable, err)
	}

	memberMine, err := groups.ListGroupsForUser(ctx, memberID)
	if err != nil || len(memberMine) != 1 || memberMine[0].PublicID != publicID {
		t.Fatalf("member groups: %+v err %v", memberMine, err)
	}
}

func newGroupPgRepositories(t *testing.T) (userrepo.PgRepository, grouprepo.PgRepository, channelrepo.PgRepository, func()) {
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

	log := logger.New(logger.Options{Service: "group-repo-test", Env: "test", Level: "error"})

	return userrepo.NewPgRepository(db), grouprepo.NewPgRepository(db, log), channelrepo.NewPgRepository(db, log), cleanup
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
