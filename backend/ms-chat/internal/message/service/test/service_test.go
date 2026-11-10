package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/wichat/wichat/backend/ms-chat/internal/message/service"
	"github.com/wichat/wichat/backend/ms-chat/internal/models"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

type fakeRepo struct{}

func (fakeRepo) ListByChannel(context.Context, string, int, *time.Time) ([]models.MessageWithAuthor, error) {
	return nil, nil
}

func (fakeRepo) Insert(context.Context, string, string, string, time.Time) (models.Message, error) {
	return models.Message{}, nil
}

func (fakeRepo) FindUserDisplayName(context.Context, string) (string, error) {
	return "", nil
}

func (fakeRepo) DeleteByID(context.Context, string) error {
	return nil
}

type fakeChannel struct{}

func (fakeChannel) ListGroupChannels(context.Context, *userv1.ListGroupChannelsRequest) (*userv1.ListGroupChannelsResponse, error) {
	return nil, nil
}

func (fakeChannel) AssertGroupChannelAccess(context.Context, *userv1.AssertGroupChannelAccessRequest) (*userv1.AssertGroupChannelAccessResponse, error) {
	return &userv1.AssertGroupChannelAccessResponse{
		ChannelUuid: "channel-uuid",
		UserUuid:    "user-uuid",
	}, nil
}

type noopPublisher struct{}

func (noopPublisher) PublishMessageCreated(context.Context, string, string, models.PublicMessage) error {
	return nil
}

func TestCreateChannelMessageRejectsEmptyBody(t *testing.T) {
	svc := service.NewService(fakeRepo{}, fakeChannel{}, noopPublisher{}, nil)
	_, err := svc.CreateChannelMessage(context.Background(), &models.CreateChannelMessageInput{
		ResolvedChannelID: "channel-uuid",
		CallerUserID:      "user-uuid",
		Body:              "   ",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	ex, ok := exception.As(err)
	if !ok || ex.Kind != exception.KindInvalidInput {
		t.Fatalf("expected invalid input, got %v", err)
	}
}
