package service

import (
	"context"
	"errors"
	"io"
	"time"

	authrepo "github.com/wichat/wichat/backend/ms-user/internal/auth/repository"
	grouprepo "github.com/wichat/wichat/backend/ms-user/internal/group/repository"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	userrepo "github.com/wichat/wichat/backend/ms-user/internal/user/repository"
	"github.com/wichat/wichat/backend/ms-user/pkg/access"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	"github.com/wichat/wichat/backend/ms-user/pkg/token"
	"github.com/wichat/wichat/backend/ms-user/pkg/validator"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/infra/minio"
)

// Service is the group membership use-case API.
type Service interface {
	ListMyGroups(ctx context.Context, accessToken string) ([]models.PublicGroup, error)
	ListJoinableGroups(ctx context.Context, accessToken string) ([]models.PublicGroup, error)
	CreateGroup(ctx context.Context, accessToken, name string) (*models.PublicGroup, error)
	JoinGroup(ctx context.Context, accessToken, groupID string) error
	LeaveGroup(ctx context.Context, accessToken, groupID string) error
	UploadGroupIcon(ctx context.Context, accessToken, groupPublicID, contentType string, body io.Reader, size int64) (*models.PublicGroup, error)
}

type service struct {
	groups   grouprepo.PgRepository
	users    userrepo.PgRepository
	objects  grouprepo.MinioRepository
	sessions authrepo.SessionRepository
	secret   []byte
	log      logger.Logger
	now      func() time.Time
}

// NewService builds group use cases.
func NewService(
	groups grouprepo.PgRepository,
	users userrepo.PgRepository,
	objects grouprepo.MinioRepository,
	sessions authrepo.SessionRepository,
	secret []byte,
	log logger.Logger,
) Service {
	copied := make([]byte, len(secret))
	copy(copied, secret)

	return service{
		groups:   groups,
		users:    users,
		objects:  objects,
		sessions: sessions,
		secret:   copied,
		log:      log,
		now:      func() time.Time { return time.Now().UTC() },
	}
}

func (s service) ListMyGroups(ctx context.Context, accessToken string) ([]models.PublicGroup, error) {
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, s.now())
	if err != nil {
		return nil, err
	}

	groups, err := s.groups.ListGroupsForUser(ctx, userID)
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
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, s.now())
	if err != nil {
		return nil, err
	}

	groups, err := s.groups.ListJoinableGroups(ctx, userID)
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
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, s.now())
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
	var group models.Group
	var member models.GroupMember
	for range 5 {
		publicID, err := token.NewPublicID()
		if err != nil {
			return nil, exception.Internal(err)
		}
		group = models.Group{
			ID:        groupID,
			PublicID:  publicID,
			Name:      trimmed,
			LogoURL:   "",
			CreatedAt: now,
			UpdatedAt: now,
		}
		member = models.GroupMember{
			GroupID:   groupID,
			UserID:    userID,
			Role:      constants.GroupRoleOwner,
			CreatedAt: now,
		}
		err = s.groups.CreateGroupWithOwner(ctx, group, member)
		if err == nil {
			break
		}
		if errors.Is(err, constants.ErrConflict) {
			continue
		}

		return nil, err
	}
	if group.PublicID == "" {
		return nil, exception.Internal(errors.New("group public id"))
	}

	saved, err := s.groups.FindGroupByID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	pub := saved.Public()

	return &pub, nil
}

func (s service) JoinGroup(ctx context.Context, accessToken, groupID string) error {
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, s.now())
	if err != nil {
		return err
	}

	publicID, err := validator.GroupPublicID(groupID)
	if err != nil {
		return err
	}

	group, err := s.groups.FindGroupByPublicID(ctx, publicID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "group_id"})
		}

		return err
	}

	if _, err := s.groups.FindGroupMember(ctx, group.ID, userID); err != nil {
		if !errors.Is(err, constants.ErrNotFound) {
			return err
		}
	} else {
		return nil
	}

	now := s.now()
	if err := s.groups.InsertGroupMember(ctx, models.GroupMember{
		GroupID:   group.ID,
		UserID:    userID,
		Role:      constants.GroupRoleMember,
		CreatedAt: now,
	}); err != nil {
		return err
	}

	return nil
}

func (s service) LeaveGroup(ctx context.Context, accessToken, groupID string) error {
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, s.now())
	if err != nil {
		return err
	}

	publicID, err := validator.GroupPublicID(groupID)
	if err != nil {
		return err
	}

	group, err := s.groups.FindGroupByPublicID(ctx, publicID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "group_id"})
		}

		return err
	}

	member, err := s.groups.FindGroupMember(ctx, group.ID, userID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "group_id"})
		}

		return err
	}
	if member.Role == constants.GroupRoleOwner {
		return exception.PermissionDenied(exception.MsgForbidden, nil)
	}

	if err := s.groups.DeleteGroupMember(ctx, group.ID, userID); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "group_id"})
		}

		return err
	}

	return nil
}

func (s service) UploadGroupIcon(
	ctx context.Context,
	accessToken,
	groupPublicID,
	contentType string,
	body io.Reader,
	size int64,
) (*models.PublicGroup, error) {
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, s.now())
	if err != nil {
		return nil, err
	}

	if err := validator.GroupIconUpload(contentType, size); err != nil {
		return nil, err
	}

	publicID, err := validator.GroupPublicID(groupPublicID)
	if err != nil {
		return nil, err
	}

	group, err := s.groups.FindGroupByPublicID(ctx, publicID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "group_id"})
		}

		return nil, err
	}

	member, err := s.groups.FindGroupMember(ctx, group.ID, userID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, exception.PermissionDenied(exception.MsgForbidden, nil)
		}

		return nil, err
	}
	if member.Role != constants.GroupRoleOwner {
		return nil, exception.PermissionDenied(exception.MsgForbidden, nil)
	}

	logoURL, err := s.objects.PutGroupIcon(ctx, group.PublicID, contentType, body, size)
	if err != nil {
		if errors.Is(err, minio.ErrDisabled) {
			return nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "file"})
		}

		return nil, err
	}

	now := s.now()
	if err := s.groups.UpdateGroupLogoURL(ctx, group.ID, logoURL, now); err != nil {
		return nil, err
	}

	saved, err := s.groups.FindGroupByID(ctx, group.ID)
	if err != nil {
		return nil, err
	}

	pub := saved.Public()

	return &pub, nil
}
