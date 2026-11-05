package testSupport

import (
	"testing"

	authrepo "github.com/wichat/wichat/backend/ms-user/internal/auth/repository"
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

// NewGroupService wires the group service against in-memory repos.
func NewGroupService(t *testing.T, st *usermock.State) (groupservice.Service, grouprepo.PgRepository, userrepo.PgRepository, authrepo.SessionRepository) {
	t.Helper()

	users, groups, sessions := TestRepos(st)
	log := logger.New(logger.Options{Service: "group-test", Env: "test", Level: "error"})

	return groupservice.NewService(groups, users, disabledGroupMinio(log), sessions, []byte(TestJWTSecret), log), groups, users, sessions
}
