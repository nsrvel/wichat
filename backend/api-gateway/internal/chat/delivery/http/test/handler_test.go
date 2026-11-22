package http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	chatdelivery "github.com/wichat/wichat/backend/api-gateway/internal/chat/delivery/http"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices"
	msChat "github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-chat"
	chatv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/chat/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeMessageGRPC struct {
	resolve func(context.Context, *chatv1.ResolveMessageChannelSeqRequest) (*chatv1.ResolveMessageChannelSeqResponse, error)
}

func (f fakeMessageGRPC) ListChannelMessages(context.Context, *chatv1.ListChannelMessagesRequest) (*chatv1.ListChannelMessagesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeMessageGRPC) CreateChannelMessage(context.Context, *chatv1.CreateChannelMessageRequest) (*chatv1.CreateChannelMessageResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeMessageGRPC) UpdateChannelMessage(context.Context, *chatv1.UpdateChannelMessageRequest) (*chatv1.UpdateChannelMessageResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeMessageGRPC) DeleteChannelMessage(context.Context, *chatv1.DeleteChannelMessageRequest) (*chatv1.DeleteChannelMessageResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeMessageGRPC) ToggleMessageReaction(context.Context, *chatv1.ToggleMessageReactionRequest) (*chatv1.ToggleMessageReactionResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeMessageGRPC) SetChannelMessagePinned(context.Context, *chatv1.SetChannelMessagePinnedRequest) (*chatv1.SetChannelMessagePinnedResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeMessageGRPC) ListPinnedChannelMessages(context.Context, *chatv1.ListPinnedChannelMessagesRequest) (*chatv1.ListPinnedChannelMessagesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeMessageGRPC) UploadChannelAttachment(context.Context, *chatv1.UploadChannelAttachmentRequest) (*chatv1.UploadChannelAttachmentResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeMessageGRPC) ResolveMessageChannelSeq(ctx context.Context, in *chatv1.ResolveMessageChannelSeqRequest) (*chatv1.ResolveMessageChannelSeqResponse, error) {
	if f.resolve != nil {
		return f.resolve(ctx, in)
	}
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func testChatHandler(msg fakeMessageGRPC) chatdelivery.Handler {
	return chatdelivery.NewHandler(&microservices.Clients{
		Chat: msChat.Service{GRPC: msChat.GRPC{Message: msg}},
	}, logger.New(logger.Options{Service: "gateway-test", Env: "test", Level: "error"}))
}

func TestResolveMessageChannelSeqHTTP(t *testing.T) {
	groupID := "aaaaaaaaaaaaaaaaaaaaaaaa"
	channelID := "bbbbbbbbbbbbbbbbbbbbbbbb"
	messageID := "cccccccccccccccccccccccc"

	handler := testChatHandler(fakeMessageGRPC{
		resolve: func(_ context.Context, in *chatv1.ResolveMessageChannelSeqRequest) (*chatv1.ResolveMessageChannelSeqResponse, error) {
			if in.GetMessageId() != messageID {
				t.Fatalf("message id %q", in.GetMessageId())
			}
			return &chatv1.ResolveMessageChannelSeqResponse{
				MessageId:  messageID,
				GroupId:    groupID,
				ChannelId:  channelID,
				ChannelSeq: 501,
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups/"+groupID+"/channels/"+channelID+"/messages/"+messageID+"/seq", nil)
	req.SetPathValue("groupId", groupID)
	req.SetPathValue("channelId", channelID)
	req.SetPathValue("messageId", messageID)
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()
	handler.ResolveMessageChannelSeq(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Data struct {
			ChannelSeq int64  `json:"channel_seq"`
			MessageID  string `json:"message_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Data.ChannelSeq != 501 || body.Data.MessageID != messageID {
		t.Fatalf("unexpected %+v", body.Data)
	}
}
