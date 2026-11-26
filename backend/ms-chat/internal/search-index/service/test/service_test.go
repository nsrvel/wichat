package service_test

import (
	"context"
	"encoding/json"
	"testing"

	searchindexservice "github.com/wichat/wichat/backend/ms-chat/internal/search-index/service"
	searchmock "github.com/wichat/wichat/backend/ms-chat/internal/search/repository/mock"
	chatEvents "github.com/wichat/wichat/backend/wi-shared/chat-events"
)

func TestIndexKafkaValue_deleted(t *testing.T) {
	st := searchmock.NewState()
	svc := searchindexservice.NewService(searchmock.NewEsRepositoryMock(st), nil)

	payload, err := json.Marshal(chatEvents.MessageDeleted{
		Type:      chatEvents.TypeMessageDeleted,
		GroupID:   "g1",
		ChannelID: "c1",
		MessageID: "m-public",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.IndexKafkaValue(context.Background(), payload); err != nil {
		t.Fatal(err)
	}
	if len(st.DeletedIDs) != 1 || st.DeletedIDs[0] != "m-public" {
		t.Fatalf("deleted ids: %v", st.DeletedIDs)
	}
}

func TestIndexKafkaValue_poisonPayload(t *testing.T) {
	svc := searchindexservice.NewService(searchmock.NewEsRepositoryMock(searchmock.NewState()), nil)
	err := svc.IndexKafkaValue(context.Background(), []byte(`{`))
	if err == nil || !searchindexservice.InvalidEventPayload(err) {
		t.Fatalf("expected invalid payload, got %v", err)
	}
}
