package service

import (
	"context"
	"errors"
	"time"

	authrepo "github.com/wichat/wichat/backend/ms-user/internal/auth/repository"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	userrepo "github.com/wichat/wichat/backend/ms-user/internal/user/repository"
	"github.com/wichat/wichat/backend/ms-user/pkg/access"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	"github.com/wichat/wichat/backend/ms-user/pkg/token"
	"github.com/wichat/wichat/backend/ms-user/pkg/validator"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

// Service is the user/workspace use-case API.
type Service interface {
	SetupStatus(ctx context.Context) (*models.SetupStatus, error)
	CompleteInstanceSetup(ctx context.Context, accessToken string, in models.CompleteInstanceSetupInput) error
	CreateGuest(ctx context.Context, in models.GuestRegisterInput) (models.User, error)
	UsernameEnabledForNewUsers(ctx context.Context) (bool, error)
}

type service struct {
	users    userrepo.PgRepository
	sessions authrepo.PgRepository
	secret   []byte
	log      logger.Logger
	now      func() time.Time
}

// NewService builds user/workspace use cases.
func NewService(users userrepo.PgRepository, sessions authrepo.PgRepository, secret []byte, log logger.Logger) Service {
	copied := make([]byte, len(secret))
	copy(copied, secret)

	return service{
		users:    users,
		sessions: sessions,
		secret:   copied,
		log:      log,
		now:      func() time.Time { return time.Now().UTC() },
	}
}

func (s service) SetupStatus(ctx context.Context) (*models.SetupStatus, error) {
	settings, err := s.users.GetInstanceSettings(ctx)
	if err != nil {
		return nil, err
	}

	return &models.SetupStatus{
		RegistrationMode: settings.RegistrationMode,
		GuestEnabled:     settings.GuestEnabled,
		UsernameEnabled:  settings.UsernameEnabled,
	}, nil
}

func (s service) CompleteInstanceSetup(ctx context.Context, accessToken string, in models.CompleteInstanceSetupInput) error {
	user, err := access.UserFromAccessToken(ctx, accessToken, s.secret, s.sessions, s.users, s.now())
	if err != nil {
		return err
	}
	if !user.IsInstanceOwner {
		return exception.PermissionDenied(exception.MsgForbidden, nil)
	}

	settings, err := s.users.GetInstanceSettings(ctx)
	if err != nil {
		return err
	}
	if settings.SetupCompleted {
		return exception.Conflict(exception.MsgSetupClosed, nil)
	}

	payload, err := validator.InstanceSetup(in)
	if err != nil {
		return err
	}

	groupID, err := token.NewID()
	if err != nil {
		return exception.Internal(err)
	}
	payload.GroupID = groupID

	now := s.now()
	if err := s.users.CompleteInstanceSetup(ctx, user.ID, payload, now); err != nil {
		if errors.Is(err, constants.ErrNotFound) || errors.Is(err, constants.ErrConflict) {
			return exception.Conflict(exception.MsgSetupClosed, nil)
		}

		return err
	}

	return nil
}

func (s service) CreateGuest(ctx context.Context, in models.GuestRegisterInput) (models.User, error) {
	name, birth, err := validator.GuestRegister(in)
	if err != nil {
		return models.User{}, err
	}

	if err := s.allowGuestRegistration(ctx); err != nil {
		return models.User{}, err
	}

	now := s.now()
	id, err := token.NewID()
	if err != nil {
		return models.User{}, exception.Internal(err)
	}

	user := models.User{
		ID:          id,
		DisplayName: name,
		BirthDate:   birth,
		IsGuest:     true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.users.CreateUser(ctx, user); err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (s service) UsernameEnabledForNewUsers(ctx context.Context) (bool, error) {
	settings, err := s.users.GetInstanceSettings(ctx)
	if err != nil {
		return false, err
	}

	return settings.UsernameEnabled, nil
}

func (s service) allowGuestRegistration(ctx context.Context) error {
	settings, err := s.users.GetInstanceSettings(ctx)
	if err != nil {
		return err
	}
	if !settings.GuestEnabled {
		return exception.PermissionDenied(exception.MsgGuestDisabled, nil)
	}

	return nil
}
