package testSupport

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	authrepo "github.com/wichat/wichat/backend/ms-user/internal/auth/repository"
	authmock "github.com/wichat/wichat/backend/ms-user/internal/auth/repository/mock"
	"github.com/wichat/wichat/backend/ms-user/internal/auth/service"
	groupservice "github.com/wichat/wichat/backend/ms-user/internal/group/service"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	userrepo "github.com/wichat/wichat/backend/ms-user/internal/user/repository"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
	userservice "github.com/wichat/wichat/backend/ms-user/internal/user/service"
	googleoauth "github.com/wichat/wichat/backend/ms-user/pkg/oauth"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	notificationv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/notification/v1"
)

const TestJWTSecret = "test-secret-must-be-32-characters!!"

// FakeNotifier records outbound notification sends in tests.
type FakeNotifier struct {
	Sends    int
	LastCode string
}

func (f *FakeNotifier) Send(_ context.Context, in *notificationv1.SendRequest) (*notificationv1.SendResponse, error) {
	f.Sends++
	if in != nil && in.Vars != nil {
		f.LastCode = in.Vars["code"]
	}

	return &notificationv1.SendResponse{Ok: true}, nil
}

// TestRepos returns in-memory Postgres repository mocks sharing state.
func TestRepos(st *usermock.State) (userrepo.PgRepository, authrepo.PgRepository) {
	return usermock.NewPgRepositoryMock(st), authmock.NewPgRepositoryMock(st)
}

// RedisClient starts miniredis and returns a go-redis client tied to test cleanup.
func RedisClient(t *testing.T) *redis.Client {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(mr.Close)

	return redis.NewClient(&redis.Options{Addr: mr.Addr()})
}

// NewAuthService wires auth + user services with miniredis OTP stores.
// NewAuthService wires auth, user, and group services with a shared miniredis OTP backend.
// The returned Miniredis handle supports FastForward in rate-limit tests.
func NewAuthService(t *testing.T, st *usermock.State) (service.Service, *FakeNotifier, userrepo.PgRepository, userservice.Service, groupservice.Service, *miniredis.Miniredis) {
	t.Helper()

	users, sessions := TestRepos(st)
	notify := &FakeNotifier{}

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(mr.Close)

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	log := logger.New(logger.Options{Service: "auth-test", Env: "test", Level: "error"})
	userSvc := userservice.NewService(users, sessions, []byte(TestJWTSecret), log)
	groupSvc := groupservice.NewService(users, sessions, []byte(TestJWTSecret), log)
	svc, err := service.NewService(users, sessions, []byte(TestJWTSecret), log, googleoauth.Config{}, authrepo.NewRedisPasswordResetStore(rdb), authrepo.NewRedisEmailAuthStore(rdb), notify, userSvc)
	if err != nil {
		t.Fatal(err)
	}

	return svc, notify, users, userSvc, groupSvc, mr
}

// RegisterSession completes register + email OTP in tests.
func RegisterSession(t *testing.T, svc service.Service, notify *FakeNotifier, in models.RegisterInput) *models.SessionResult {
	t.Helper()
	ctx := context.Background()

	step, err := svc.Register(ctx, in)
	if err != nil {
		t.Fatal(err)
	}
	if step.Challenge == nil {
		t.Fatal("expected email challenge")
	}
	code := notify.LastCode
	if code == "" {
		t.Fatal("expected emailed code")
	}

	sess, err := svc.VerifyRegister(ctx, step.Challenge.ChallengeID, code)
	if err != nil {
		t.Fatal(err)
	}

	return sess
}

// LoginSession completes login + email OTP in tests.
func LoginSession(t *testing.T, svc service.Service, notify *FakeNotifier, in models.LoginInput) *models.SessionResult {
	t.Helper()
	ctx := context.Background()

	step, err := svc.Login(ctx, in)
	if err != nil {
		t.Fatal(err)
	}
	if step.Challenge == nil {
		t.Fatal("expected email challenge")
	}
	code := notify.LastCode
	if code == "" {
		t.Fatal("expected emailed code")
	}

	sess, err := svc.VerifyLogin(ctx, step.Challenge.ChallengeID, code)
	if err != nil {
		t.Fatal(err)
	}

	return sess
}
