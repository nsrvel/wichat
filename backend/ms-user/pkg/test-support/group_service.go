package testSupport

import (
	"context"
	"io"
	"testing"

	authrepo "github.com/wichat/wichat/backend/ms-user/internal/auth/repository"
	channelmock "github.com/wichat/wichat/backend/ms-user/internal/channel/repository/mock"
	channelservice "github.com/wichat/wichat/backend/ms-user/internal/channel/service"
	groupmembermock "github.com/wichat/wichat/backend/ms-user/internal/group-member/repository/mock"
	groupmemberservice "github.com/wichat/wichat/backend/ms-user/internal/group-member/service"
	grouprolemock "github.com/wichat/wichat/backend/ms-user/internal/group-role/repository/mock"
	grouproleservice "github.com/wichat/wichat/backend/ms-user/internal/group-role/service"
	grouprepo "github.com/wichat/wichat/backend/ms-user/internal/group/repository"
	groupservice "github.com/wichat/wichat/backend/ms-user/internal/group/service"
	userrepo "github.com/wichat/wichat/backend/ms-user/internal/user/repository"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/infra/minio"
)

func disabledGroupMinio(log logger.Logger) grouprepo.MinioRepository {
	client, err := minio.New(minio.Config{}, log)
	if err != nil {
		panic(err)
	}

	return grouprepo.NewMinioRepository(client)
}

// FakeGroupMinio is a stub object store for group icon upload tests.
type FakeGroupMinio struct{}

func (FakeGroupMinio) PutGroupIcon(_ context.Context, groupPublicID, _ string, _ io.Reader, _ int64) (string, error) {
	return "https://cdn.example/groups/" + groupPublicID + "/icon.png", nil
}

func (FakeGroupMinio) RemoveGroupIcon(context.Context, string) error { return nil }

func (FakeGroupMinio) PutGroupBanner(_ context.Context, groupPublicID, _ string, _ io.Reader, _ int64) (string, error) {
	return "https://cdn.example/groups/" + groupPublicID + "/banner.png", nil
}

func (FakeGroupMinio) RemoveGroupBanner(context.Context, string) error { return nil }

// NewGroupService wires the group service against in-memory repos.
func NewGroupService(t *testing.T, st *usermock.State) (groupservice.Service, grouprepo.PgRepository, userrepo.PgRepository, authrepo.SessionRepository) {
	t.Helper()

	users, groups, sessions := TestRepos(st)
	members := groupmembermock.NewPgRepositoryMock(st)
	roles := grouprolemock.NewPgRepositoryMock(st)
	log := logger.New(logger.Options{Service: "group-test", Env: "test", Level: "error"})

	channelSvc := channelservice.NewService(channelmock.NewPgRepositoryMock(st), groups, members, roles, sessions, []byte(TestJWTSecret), log)

	return groupservice.NewService(groups, members, roles, channelSvc, users, disabledGroupMinio(log), sessions, []byte(TestJWTSecret), log), groups, users, sessions
}

// NewGroupServiceWithFakeMinio wires the group service with a fake object store.
func NewGroupServiceWithFakeMinio(t *testing.T, st *usermock.State) groupservice.Service {
	t.Helper()

	users, groups, sessions := TestRepos(st)
	members := groupmembermock.NewPgRepositoryMock(st)
	roles := grouprolemock.NewPgRepositoryMock(st)
	log := logger.New(logger.Options{Service: "group-test", Env: "test", Level: "error"})
	channelSvc := channelservice.NewService(channelmock.NewPgRepositoryMock(st), groups, members, roles, sessions, []byte(TestJWTSecret), log)

	return groupservice.NewService(groups, members, roles, channelSvc, users, FakeGroupMinio{}, sessions, []byte(TestJWTSecret), log)
}

// NewGroupMemberService wires the group-member service against in-memory repos.
func NewGroupMemberService(t *testing.T, st *usermock.State) groupmemberservice.Service {
	t.Helper()

	_, groups, sessions := TestRepos(st)
	members := groupmembermock.NewPgRepositoryMock(st)
	roles := grouprolemock.NewPgRepositoryMock(st)
	log := logger.New(logger.Options{Service: "group-member-test", Env: "test", Level: "error"})

	return groupmemberservice.NewService(members, groups, roles, sessions, []byte(TestJWTSecret), log)
}

// NewGroupRoleService wires the group-role service against in-memory repos.
func NewGroupRoleService(t *testing.T, st *usermock.State) grouproleservice.Service {
	t.Helper()

	_, groups, sessions := TestRepos(st)
	members := groupmembermock.NewPgRepositoryMock(st)
	roles := grouprolemock.NewPgRepositoryMock(st)
	log := logger.New(logger.Options{Service: "group-role-test", Env: "test", Level: "error"})

	return grouproleservice.NewService(roles, groups, members, sessions, []byte(TestJWTSecret), log)
}
