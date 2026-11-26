package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	authHttp "github.com/wichat/wichat/backend/api-gateway/pkg/auth-http"
	chatHttp "github.com/wichat/wichat/backend/api-gateway/pkg/chat-http"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	chatv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/chat/v1"
	"github.com/wichat/wichat/backend/wi-shared/response"
	responsehttp "github.com/wichat/wichat/backend/wi-shared/response/http"
)

func (h handler) ListGroupMentions(w http.ResponseWriter, r *http.Request) {
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	groupID := strings.TrimSpace(r.PathValue("groupId"))
	if groupID == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, nil))
		return
	}

	limit := int32(20)
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		n, err := strconv.ParseInt(raw, 10, 32)
		if err != nil || n <= 0 {
			authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "limit"}))
			return
		}
		limit = int32(n)
	}

	kind := strings.TrimSpace(r.URL.Query().Get("kind"))
	cursor := strings.TrimSpace(r.URL.Query().Get("cursor"))

	out, err := h.ms.Chat.GRPC.Search.ListGroupMentions(r.Context(), &chatv1.ListGroupMentionsRequest{
		AccessToken: token,
		GroupId:     groupID,
		Limit:       limit,
		Cursor:      cursor,
		Kind:        kind,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(chatHttp.GroupMentionsListBody{
		Hits:       chatHttp.GroupMentionsFromPB(out.GetHits()),
		NextCursor: out.GetNextCursor(),
	}))
}

func (h handler) MarkGroupMentionsRead(w http.ResponseWriter, r *http.Request) {
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	groupID := strings.TrimSpace(r.PathValue("groupId"))
	if groupID == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, nil))
		return
	}

	var body chatHttp.MarkMentionsReadBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, nil))
		return
	}

	_, err := h.ms.Chat.GRPC.Search.MarkGroupMentionsRead(r.Context(), &chatv1.MarkGroupMentionsReadRequest{
		AccessToken: token,
		GroupId:     groupID,
		MessageIds:  body.MessageIDs,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.OKBody{OK: true}))
}

func (h handler) CountGroupMentionsUnread(w http.ResponseWriter, r *http.Request) {
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	groupID := strings.TrimSpace(r.PathValue("groupId"))
	if groupID == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, nil))
		return
	}

	out, err := h.ms.Chat.GRPC.Search.CountGroupMentionsUnread(r.Context(), &chatv1.CountGroupMentionsUnreadRequest{
		AccessToken: token,
		GroupId:     groupID,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(chatHttp.MentionsUnreadCountBody{
		Count: out.GetCount(),
	}))
}
