package service

import (
	"context"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/wichat/wichat/backend/ms-chat/internal/kafka/producers"
	messagerepo "github.com/wichat/wichat/backend/ms-chat/internal/message/repository"
	"github.com/wichat/wichat/backend/ms-chat/internal/models"
	"github.com/wichat/wichat/backend/ms-chat/pkg/constants"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	userchannel "github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user/grpc/channel"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

// Service is channel message use cases.
type Service interface {
	ListChannelMessages(ctx context.Context, in *models.ListChannelMessagesInput) ([]models.PublicMessage, error)
	CreateChannelMessage(ctx context.Context, in *models.CreateChannelMessageInput) (models.PublicMessage, error)
}

type service struct {
	messages    messagerepo.PgRepository
	userChannel userchannel.Client
	events      producers.MessageCreatedPublisher
	log         logger.Logger
	now         func() time.Time
}

// NewService builds message use cases.
func NewService(
	messages messagerepo.PgRepository,
	userChannel userchannel.Client,
	events producers.MessageCreatedPublisher,
	log logger.Logger,
) Service {
	return service{
		messages:    messages,
		userChannel: userChannel,
		events:      events,
		log:         log,
		now:         func() time.Time { return time.Now().UTC() },
	}
}

func (s service) ListChannelMessages(ctx context.Context, in *models.ListChannelMessagesInput) ([]models.PublicMessage, error) {
	// Repository
	rows, err := s.messages.ListByChannel(ctx, in.ResolvedChannelID, in.Limit, in.BeforeCreatedAt)
	if err != nil {
		return nil, err
	}

	out := make([]models.PublicMessage, len(rows))
	for i, row := range rows {
		out[i] = toPublicMessage(row, in.CallerUserID)
	}

	return out, nil
}

func (s service) CreateChannelMessage(ctx context.Context, in *models.CreateChannelMessageInput) (models.PublicMessage, error) {
	// Validation
	body := strings.TrimSpace(in.Body)
	if body == "" || utf8.RuneCountInString(body) > constants.MaxMessageBodyRunes || hasDangerousControlChars(body) {
		return models.PublicMessage{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "body"})
	}

	// Repository
	createdAt := s.now()
	msg, err := s.messages.Insert(ctx, in.ResolvedChannelID, in.CallerUserID, body, createdAt)
	if err != nil {
		return models.PublicMessage{}, err
	}

	displayName, err := s.messages.FindUserDisplayName(ctx, in.CallerUserID)
	if err != nil {
		return models.PublicMessage{}, err
	}

	out := models.PublicMessage{
		ID: msg.PublicID, AuthorID: msg.AuthorID, AuthorDisplayName: displayName,
		Body: msg.Body, CreatedAt: msg.CreatedAt, Mine: true,
	}

	if err := s.events.PublishMessageCreated(ctx, in.GroupPublicID, in.ChannelPublicID, out); err != nil {
		s.log.Error("publish message.created", "err", err)
		// ponytail: compensating delete until transactional outbox ships
		if delErr := s.messages.DeleteByID(ctx, msg.ID); delErr != nil {
			s.log.Error("compensating delete after publish failure", "message_id", msg.ID, "err", delErr)
		}
		return models.PublicMessage{}, err
	}

	return out, nil
}

func hasDangerousControlChars(s string) bool {
	for _, r := range s {
		if r == '\n' || r == '\r' || r == '\t' {
			continue
		}
		if unicode.Is(unicode.Cc, r) {
			return true
		}
	}

	return false
}

// ResolveChannelAccess validates the bearer token and resolves internal ids via ms-user.
func ResolveChannelAccess(
	ctx context.Context,
	userChannel userchannel.Client,
	accessToken, groupID, channelID string,
) (channelUUID, userUUID string, err error) {
	out, err := userChannel.AssertGroupChannelAccess(ctx, &userv1.AssertGroupChannelAccessRequest{
		AccessToken: accessToken,
		GroupId:     groupID,
		ChannelId:   channelID,
	})
	if err != nil {
		return "", "", err
	}

	return out.GetChannelUuid(), out.GetUserUuid(), nil
}

func toPublicMessage(row models.MessageWithAuthor, callerUserID string) models.PublicMessage {
	return models.PublicMessage{
		ID:                row.PublicID,
		AuthorID:          row.AuthorID,
		AuthorDisplayName: row.AuthorDisplayName,
		Body:              row.Body,
		CreatedAt:         row.CreatedAt,
		Mine:              row.AuthorID == callerUserID,
	}
}
