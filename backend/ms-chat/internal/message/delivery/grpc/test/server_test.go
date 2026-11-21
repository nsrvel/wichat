package grpc_test

import (
	"context"
	"testing"

	messagegrpc "github.com/wichat/wichat/backend/ms-chat/internal/message/delivery/grpc"
	"github.com/wichat/wichat/backend/ms-chat/internal/message/service"
	"github.com/wichat/wichat/backend/ms-chat/internal/models"
	chatv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/chat/v1"
)

type stubService struct {
	service.Service
}

func (stubService) ListChannelMessages(context.Context, *models.ListChannelMessagesInput) ([]models.PublicMessage, bool, error) {
	return []models.PublicMessage{{ID: "m1", Body: "hello"}}, false, nil
}

func TestListChannelMessages_deliveryBind(t *testing.T) {
	srv := messagegrpc.NewServer(stubService{})
	res, err := srv.ListChannelMessages(context.Background(), &chatv1.ListChannelMessagesRequest{
		AccessToken: "tok",
		GroupId:     "aaaaaaaaaaaaaaaaaaaaaaaa",
		ChannelId:   "bbbbbbbbbbbbbbbbbbbbbbbb",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.GetMessages()) != 1 || res.GetMessages()[0].GetBody() != "hello" {
		t.Fatalf("unexpected response: %+v", res)
	}
}
