package service

import (
	"context"
	"errors"
	"strings"
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

// Service is the group membership use-case API.
type Service interface {
	ListMyGroups(ctx context.Context, accessToken string) ([]models.PublicGroup, error)
	ListJoinableGroups(ctx context.Context, accessToken string) ([]models.PublicGroup, error)
	CreateGroup(ctx context.Context, accessToken, name string) (*models.PublicGroup, error)
	JoinGroup(ctx context.Context, accessToken, groupID string) error
}

type service struct {
	users    userrepo.PgRepository
	sessions authrepo.PgRepository
	secret   []byte
	log      logger.Logger
	now      func() time.Time
}

// NewService builds group use cases.
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

func (s service) ListMyGroups(ctx context.Context, accessToken string) ([]models.PublicGroup, error) {
	user, err := access.UserFromAccessToken(ctx, accessToken, s.secret, s.sessions, s.users, s.now())
	if err != nil {
		return nil, err
	}

	groups, err := s.users.ListGroupsForUser(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	out := make([]models.PublicGroup, len(groups))
	for i, g := range groups {
		out[i] = g.Public()
	}

	return out, nil
}

func (s service) ListJoinableGroups(ctx context.Context, accessToken string) ([]models.PublicGroup, error) {
	user, err := access.UserFromAccessToken(ctx, accessToken, s.secret, s.sessions, s.users, s.now())
	if err != nil {
		return nil, err
	}

	groups, err := s.users.ListJoinableGroups(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	out := make([]models.PublicGroup, len(groups))
	for i, g := range groups {
		out[i] = g.Public()
	}

	return out, nil
}

func (s service) CreateGroup(ctx context.Context, accessToken, name string) (*models.PublicGroup, error) {
	user, err := access.UserFromAccessToken(ctx, accessToken, s.secret, s.sessions, s.users, s.now())
	if err != nil {
		return nil, err
	}

	trimmed, err := validator.GroupName(name)
	if err != nil {
		return nil, err
	}

	groupID, err := token.NewID()
	if err != nil {
		return nil, exception.Internal(err)
	}

	now := s.now()
	group := models.Group{
		ID:        groupID,
		Name:      trimmed,
		LogoURL:   "",
		CreatedAt: now,
		UpdatedAt: now,
	}
	member := models.GroupMember{
		GroupID:   groupID,
		UserID:    user.ID,
		Role:      constants.GroupRoleOwner,
		CreatedAt: now,
	}
	if err := s.users.CreateGroupWithOwner(ctx, group, member); err != nil {
		return nil, err
	}

	pub := group.Public()

	return &pub, nil
}

func (s service) JoinGroup(ctx context.Context, accessToken, groupID string) error {
	user, err := access.UserFromAccessToken(ctx, accessToken, s.secret, s.sessions, s.users, s.now())
	if err != nil {
		return err
	}

	groupID = strings.TrimSpace(groupID)
	if groupID == "" {
		return exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "group_id"})
	}

	if _, err := s.users.FindGroupByID(ctx, groupID); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "group_id"})
		}

		return err
	}

	if _, err := s.users.FindGroupMember(ctx, groupID, user.ID); err != nil {
		if !errors.Is(err, constants.ErrNotFound) {
			return err
		}
	} else {
		return nil
	}

	now := s.now()
	if err := s.users.InsertGroupMember(ctx, models.GroupMember{
		GroupID:   groupID,
		UserID:    user.ID,
		Role:      constants.GroupRoleMember,
		CreatedAt: now,
	}); err != nil {
		return err
	}

	return nil
}
