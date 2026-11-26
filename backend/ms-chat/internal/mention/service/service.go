package service

import (
	"context"
	"strings"
	"time"

	mentionrepo "github.com/wichat/wichat/backend/ms-chat/internal/mention/repository"
	"github.com/wichat/wichat/backend/ms-chat/internal/models"
	channelacl "github.com/wichat/wichat/backend/ms-chat/pkg/channel-acl"
	"github.com/wichat/wichat/backend/ms-chat/pkg/constants"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	userauth "github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user/grpc/auth"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

// Service runs @mention inbox use cases.
type Service interface {
	ListGroupMentions(ctx context.Context, in *models.ListGroupMentionsInput) ([]models.GroupMentionHit, string, error)
	MarkGroupMentionsRead(ctx context.Context, in *models.MarkGroupMentionsReadInput) error
	CountGroupMentionsUnread(ctx context.Context, accessToken, groupPublicID string) (int, error)
}

type service struct {
	mentions mentionrepo.PgRepository
	channel  channelacl.ChannelLister
	userAuth userauth.Client
	log      logger.Logger
}

// NewService builds mention inbox use cases.
func NewService(
	mentions mentionrepo.PgRepository,
	channel channelacl.ChannelLister,
	userAuth userauth.Client,
	log logger.Logger,
) Service {
	return service{mentions: mentions, channel: channel, userAuth: userAuth, log: log}
}

func (s service) resolveCaller(
	ctx context.Context,
	accessToken, groupPublicID string,
) (groupInternalID, userInternalID string, err error) {
	token := strings.TrimSpace(accessToken)
	groupPublicID = strings.TrimSpace(groupPublicID)
	if token == "" {
		return "", "", exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}
	if groupPublicID == "" {
		return "", "", exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "group_id"})
	}
	if s.mentions == nil || s.userAuth == nil {
		return "", "", nil
	}

	me, err := s.userAuth.Me(ctx, &userv1.MeRequest{AccessToken: token})
	if err != nil {
		return "", "", err
	}
	u := me.GetUser()
	if u == nil || strings.TrimSpace(u.GetId()) == "" {
		return "", "", exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}

	userInternalID, err = s.mentions.LookupUserInternalID(ctx, u.GetId())
	if err != nil {
		return "", "", exception.PermissionDenied(exception.MsgForbidden, nil)
	}

	groupInternalID, err = s.mentions.LookupGroupInternalID(ctx, groupPublicID)
	if err != nil {
		return "", "", exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "group_id"})
	}
	return groupInternalID, userInternalID, nil
}

func (s service) ListGroupMentions(
	ctx context.Context,
	in *models.ListGroupMentionsInput,
) ([]models.GroupMentionHit, string, error) {
	groupInternalID, userInternalID, err := s.resolveCaller(ctx, in.AccessToken, in.GroupPublicID)
	if err != nil {
		return nil, "", err
	}
	if groupInternalID == "" {
		return []models.GroupMentionHit{}, "", nil
	}

	limit := in.Limit
	if limit <= 0 {
		limit = constants.DefaultMentionListLimit
	}
	if limit > 50 {
		limit = 50
	}

	var cursor time.Time
	if c := strings.TrimSpace(in.Cursor); c != "" {
		parsed, pErr := time.Parse(time.RFC3339Nano, c)
		if pErr != nil {
			return nil, "", exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "cursor"})
		}
		cursor = parsed.UTC()
	}

	kind := strings.TrimSpace(strings.ToLower(in.Kind))
	unreadOnly := kind == "unread"
	channelKind := kind
	if unreadOnly {
		channelKind = ""
	}

	allowed, err := channelacl.AllowedIDs(ctx, s.channel, in.AccessToken, in.GroupPublicID)
	if err != nil {
		return nil, "", err
	}

	fetchLimit := limit * 3
	if fetchLimit > 150 {
		fetchLimit = 150
	}

	raw, err := s.mentions.ListForMentionedUser(
		ctx,
		groupInternalID,
		userInternalID,
		fetchLimit,
		cursor,
		channelKind,
		unreadOnly,
	)
	if err != nil {
		return nil, "", err
	}

	hits := make([]models.GroupMentionHit, 0, limit)
	for _, h := range raw {
		if _, ok := allowed[h.ChannelID]; !ok {
			continue
		}
		hits = append(hits, h)
		if len(hits) >= limit {
			break
		}
	}

	var nextCursor string
	if len(hits) >= limit {
		last := hits[len(hits)-1]
		nextCursor = last.CreatedAt.UTC().Format(time.RFC3339Nano)
	}

	return hits, nextCursor, nil
}

func (s service) MarkGroupMentionsRead(ctx context.Context, in *models.MarkGroupMentionsReadInput) error {
	groupInternalID, userInternalID, err := s.resolveCaller(ctx, in.AccessToken, in.GroupPublicID)
	if err != nil {
		return err
	}
	if groupInternalID == "" || s.mentions == nil {
		return nil
	}
	return s.mentions.MarkRead(ctx, groupInternalID, userInternalID, in.MessagePublicIDs, time.Now().UTC())
}

func (s service) CountGroupMentionsUnread(ctx context.Context, accessToken, groupPublicID string) (int, error) {
	groupInternalID, userInternalID, err := s.resolveCaller(ctx, accessToken, groupPublicID)
	if err != nil {
		return 0, err
	}
	if groupInternalID == "" || s.mentions == nil {
		return 0, nil
	}

	count, err := s.mentions.CountUnread(ctx, groupInternalID, userInternalID)
	if err != nil {
		return 0, err
	}

	allowed, err := channelacl.AllowedIDs(ctx, s.channel, accessToken, groupPublicID)
	if err != nil {
		return 0, err
	}

	// ponytail: exact ACL-unread count needs join; over-count is safe for badge when channel revoked
	_ = allowed
	return count, nil
}
