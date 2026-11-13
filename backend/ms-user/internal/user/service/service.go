package service

import (
	"context"
	"errors"
	"time"

	authrepo "github.com/wichat/wichat/backend/ms-user/internal/auth/repository"
	groupmemberrepo "github.com/wichat/wichat/backend/ms-user/internal/group-member/repository"
	grouprepo "github.com/wichat/wichat/backend/ms-user/internal/group/repository"
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
	UsernameEnabledForNewUsers(ctx context.Context) (bool, error)
	SearchUsers(ctx context.Context, accessToken, query string, limit int, groupPublicID string) ([]models.UserSearchHit, error)
}

type service struct {
	users    userrepo.PgRepository
	groups   grouprepo.PgRepository
	members  groupmemberrepo.PgRepository
	sessions authrepo.SessionRepository
	secret   []byte
	log      logger.Logger
	now      func() time.Time
}

// NewService builds user/workspace use cases.
func NewService(
	users userrepo.PgRepository,
	groups grouprepo.PgRepository,
	members groupmemberrepo.PgRepository,
	sessions authrepo.SessionRepository,
	secret []byte,
	log logger.Logger,
) Service {
	copied := make([]byte, len(secret))
	copy(copied, secret)

	return service{
		users:    users,
		groups:   groups,
		members:  members,
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
	publicID, err := token.NewPublicID()
	if err != nil {
		return exception.Internal(err)
	}
	payload.GroupID = groupID
	payload.GroupPublicID = publicID

	now := s.now()
	if err := s.users.CompleteInstanceSetup(ctx, user.ID, payload, now); err != nil {
		if errors.Is(err, constants.ErrNotFound) || errors.Is(err, constants.ErrConflict) {
			return exception.Conflict(exception.MsgSetupClosed, nil)
		}

		return err
	}

	return nil
}

func (s service) UsernameEnabledForNewUsers(ctx context.Context) (bool, error) {
	settings, err := s.users.GetInstanceSettings(ctx)
	if err != nil {
		return false, err
	}

	return settings.UsernameEnabled, nil
}

func (s service) SearchUsers(
	ctx context.Context,
	accessToken, query string,
	limit int,
	groupPublicID string,
) ([]models.UserSearchHit, error) {
	now := s.now()
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, now)
	if err != nil {
		return nil, err
	}

	publicID, err := validator.GroupPublicID(groupPublicID)
	if err != nil {
		return nil, err
	}

	group, err := s.groups.FindGroupByPublicID(ctx, publicID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, exception.NotFound(exception.MsgInvalidInput, map[string]any{"field": "group_id"})
		}
		return nil, err
	}

	if _, err := s.members.FindGroupMember(ctx, group.ID, userID); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, exception.PermissionDenied(exception.MsgForbidden, nil)
		}
		return nil, err
	}

	return s.users.SearchUsers(ctx, query, limit, group.ID)
}
