package testSupport

import (
	"testing"

	authrepo "github.com/wichat/wichat/backend/ms-user/internal/auth/repository"
	groupservice "github.com/wichat/wichat/backend/ms-user/internal/group/service"
	userrepo "github.com/wichat/wichat/backend/ms-user/internal/user/repository"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

// NewGroupService wires the group service against in-memory repos.
func NewGroupService(t *testing.T, st *usermock.State) (groupservice.Service, userrepo.PgRepository, authrepo.PgRepository) {
	t.Helper()

	users, sessions := TestRepos(st)
	log := logger.New(logger.Options{Service: "group-test", Env: "test", Level: "error"})

	return groupservice.NewService(users, sessions, []byte(TestJWTSecret), log), users, sessions
}
