package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	chathttp "github.com/wichat/wichat/backend/api-gateway/internal/chat/delivery/http"
)

type chatRouteStub struct{}

func (chatRouteStub) ListChannelMessages(http.ResponseWriter, *http.Request)            {}
func (chatRouteStub) CreateChannelMessage(http.ResponseWriter, *http.Request)          {}
func (chatRouteStub) UploadChannelAttachment(http.ResponseWriter, *http.Request)     {}
func (chatRouteStub) UpdateChannelMessage(http.ResponseWriter, *http.Request)          {}
func (chatRouteStub) DeleteChannelMessage(http.ResponseWriter, *http.Request)          {}
func (chatRouteStub) ToggleMessageReaction(http.ResponseWriter, *http.Request)       {}
func (chatRouteStub) ListPinnedChannelMessages(http.ResponseWriter, *http.Request)     {}
func (chatRouteStub) PinChannelMessage(http.ResponseWriter, *http.Request)             {}
func (chatRouteStub) UnpinChannelMessage(http.ResponseWriter, *http.Request)           {}
func (chatRouteStub) ResolveMessageChannelSeq(http.ResponseWriter, *http.Request)       {}
func (chatRouteStub) ResolveMessageChannelSeqGlobal(http.ResponseWriter, *http.Request) {}

func TestRegisterMountsResolveMessageSeqRoutes(t *testing.T) {
	mux := http.NewServeMux()
	chathttp.Register(mux, chatRouteStub{})

	groupID := "aaaaaaaaaaaaaaaaaaaaaaaa"
	channelID := "bbbbbbbbbbbbbbbbbbbbbbbb"
	messageID := "cccccccccccccccccccccccc"

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/groups/"+groupID+"/channels/"+channelID+"/messages/"+messageID+"/seq",
		nil,
	)
	req.SetPathValue("groupId", groupID)
	req.SetPathValue("channelId", channelID)
	req.SetPathValue("messageId", messageID)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code == http.StatusNotFound {
		t.Fatal("nested resolve route not registered")
	}

	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/messages/"+messageID+"/seq?group_id="+groupID+"&channel_id="+channelID, nil)
	req2.SetPathValue("messageId", messageID)
	rec2 := httptest.NewRecorder()
	mux.ServeHTTP(rec2, req2)
	if rec2.Code == http.StatusNotFound {
		t.Fatal("global resolve route not registered")
	}
}