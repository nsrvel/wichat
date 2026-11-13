package service_test

import (
	"context"
	"testing"

	"github.com/wichat/wichat/backend/ms-chat/internal/message/repository/mock"
	"github.com/wichat/wichat/backend/ms-chat/internal/message/service"
	"github.com/wichat/wichat/backend/ms-chat/internal/models"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

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
	st := mock.NewState()
	svc := service.NewService(mock.NewPgRepositoryMock(st), fakeChannel{}, noopPublisher{}, nil)
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

func TestCreateChannelMessageRejectsControlChars(t *testing.T) {
	st := mock.NewState()
	svc := service.NewService(mock.NewPgRepositoryMock(st), fakeChannel{}, noopPublisher{}, nil)
	_, err := svc.CreateChannelMessage(context.Background(), &models.CreateChannelMessageInput{
		ResolvedChannelID: "channel-uuid",
		CallerUserID:      "user-uuid",
		Body:              "hello\x00world",
	})
	if err == nil {
		t.Fatal("expected error for null byte")
	}
}
