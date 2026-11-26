package service_test

import (
	"context"
	"testing"

	searchmock "github.com/wichat/wichat/backend/ms-chat/internal/search/repository/mock"
	"github.com/wichat/wichat/backend/ms-chat/internal/search/service"
	"github.com/wichat/wichat/backend/wi-shared/infra/elasticsearch"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

type fakeChannel struct{}

func (fakeChannel) ListGroupChannels(_ context.Context, _ *userv1.ListGroupChannelsRequest) (*userv1.ListGroupChannelsResponse, error) {
	return &userv1.ListGroupChannelsResponse{
		Channels: []*userv1.ChannelSummary{
			{Id: "ch-1"},
		},
	}, nil
}

func TestSearchMessagesFiltersInaccessibleChannel(t *testing.T) {
	st := searchmock.NewState()
	st.Docs = []elasticsearch.MessageDoc{
		{MessageID: "m1", GroupID: "g1", ChannelID: "ch-1", Body: "hello world"},
		{MessageID: "m2", GroupID: "g1", ChannelID: "ch-secret", Body: "hello secret"},
	}
	svc := service.NewQueryService(searchmock.NewEsRepositoryMock(st), fakeChannel{}, logger.New(logger.Options{}))

	hits, err := svc.SearchMessages(context.Background(), "token", "g1", "hello", 8)
	if err != nil {
		t.Fatalf("SearchMessages: %v", err)
	}
	if len(hits) != 1 || hits[0].MessageID != "m1" {
		t.Fatalf("expected one accessible hit, got %+v", hits)
	}
}

func TestSearchFilesFiltersInaccessibleChannel(t *testing.T) {
	st := searchmock.NewState()
	st.AttachmentDocs = []elasticsearch.AttachmentDoc{
		{AttachmentID: "a1", GroupID: "g1", ChannelID: "ch-1", Filename: "report.pdf"},
		{AttachmentID: "a2", GroupID: "g1", ChannelID: "ch-secret", Filename: "report-secret.pdf"},
	}
	svc := service.NewQueryService(searchmock.NewEsRepositoryMock(st), fakeChannel{}, logger.New(logger.Options{}))

	hits, err := svc.SearchFiles(context.Background(), "token", "g1", "report", 8, false)
	if err != nil {
		t.Fatalf("SearchFiles: %v", err)
	}
	if len(hits) != 1 || hits[0].AttachmentID != "a1" {
		t.Fatalf("expected one accessible file hit, got %+v", hits)
	}
}
