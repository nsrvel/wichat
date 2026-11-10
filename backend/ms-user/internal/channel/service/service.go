package service

import (
	"context"
	"errors"
	"time"

	authrepo "github.com/wichat/wichat/backend/ms-user/internal/auth/repository"
	channelrepo "github.com/wichat/wichat/backend/ms-user/internal/channel/repository"
	grouprepo "github.com/wichat/wichat/backend/ms-user/internal/group/repository"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	"github.com/wichat/wichat/backend/ms-user/pkg/access"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	"github.com/wichat/wichat/backend/ms-user/pkg/validator"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

// Service is channel metadata for a group workspace.
type Service interface {
	ListGroupChannels(ctx context.Context, accessToken, groupPublicID string) ([]models.PublicChannel, error)
	AssertGroupChannelAccess(ctx context.Context, accessToken, groupPublicID, channelPublicID string) (models.ChannelAccess, error)
	SeedDefaultChannels(ctx context.Context, groupID, ownerUserID string, createdAt time.Time) error
}

type service struct {
	channels channelrepo.PgRepository
	groups   grouprepo.PgRepository
	sessions authrepo.SessionRepository
	secret   []byte
	log      logger.Logger
	now      func() time.Time
}

// NewService builds channel use cases.
func NewService(
	channels channelrepo.PgRepository,
	groups grouprepo.PgRepository,
	sessions authrepo.SessionRepository,
	secret []byte,
	log logger.Logger,
) Service {
	copied := make([]byte, len(secret))
	copy(copied, secret)

	return service{
		channels: channels,
		groups:   groups,
		sessions: sessions,
		secret:   copied,
		log:      log,
		now:      func() time.Time { return time.Now().UTC() },
	}
}

func (s service) ListGroupChannels(ctx context.Context, accessToken, groupPublicID string) ([]models.PublicChannel, error) {
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, s.now())
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
			return nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "group_id"})
		}

		return nil, err
	}

	if _, err := s.groups.FindGroupMember(ctx, group.ID, userID); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, exception.PermissionDenied(exception.MsgForbidden, nil)
		}

		return nil, err
	}

	channels, err := s.channels.ListByGroupID(ctx, group.ID)
	if err != nil {
		return nil, err
	}

	out := make([]models.PublicChannel, len(channels))
	for i, ch := range channels {
		out[i] = ch.Public()
	}

	return out, nil
}

func (s service) AssertGroupChannelAccess(ctx context.Context, accessToken, groupPublicID, channelPublicID string) (models.ChannelAccess, error) {
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, s.now())
	if err != nil {
		return models.ChannelAccess{}, err
	}

	groupID, err := validator.GroupPublicID(groupPublicID)
	if err != nil {
		return models.ChannelAccess{}, err
	}

	channelID, err := validator.ChannelPublicID(channelPublicID)
	if err != nil {
		return models.ChannelAccess{}, err
	}

	group, err := s.groups.FindGroupByPublicID(ctx, groupID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return models.ChannelAccess{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "group_id"})
		}

		return models.ChannelAccess{}, err
	}

	if _, err := s.groups.FindGroupMember(ctx, group.ID, userID); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return models.ChannelAccess{}, exception.PermissionDenied(exception.MsgForbidden, nil)
		}

		return models.ChannelAccess{}, err
	}

	ch, err := s.channels.FindByPublicIDInGroup(ctx, group.ID, channelID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return models.ChannelAccess{}, exception.PermissionDenied(exception.MsgForbidden, nil)
		}

		return models.ChannelAccess{}, err
	}

	return models.ChannelAccess{ChannelID: ch.ID, UserID: userID}, nil
}

func (s service) SeedDefaultChannels(ctx context.Context, groupID, ownerUserID string, createdAt time.Time) error {
	return s.channels.SeedDefaultChannels(ctx, groupID, ownerUserID, createdAt)
}
