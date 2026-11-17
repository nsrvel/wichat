package http

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	authHttp "github.com/wichat/wichat/backend/api-gateway/pkg/auth-http"
	chatHttp "github.com/wichat/wichat/backend/api-gateway/pkg/chat-http"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices"
	chatv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/chat/v1"
	"github.com/wichat/wichat/backend/wi-shared/response"
	responsehttp "github.com/wichat/wichat/backend/wi-shared/response/http"
)

const (
	maxMessageBodyBytes      = 1 << 14
	maxAttachmentUploadBytes = 25 << 20
	maxMultipartOverhead     = 1 << 20
)

// Handler is the public chat message HTTP delivery.
type Handler interface {
	ListChannelMessages(w http.ResponseWriter, r *http.Request)
	CreateChannelMessage(w http.ResponseWriter, r *http.Request)
	UploadChannelAttachment(w http.ResponseWriter, r *http.Request)
	UpdateChannelMessage(w http.ResponseWriter, r *http.Request)
	DeleteChannelMessage(w http.ResponseWriter, r *http.Request)
	ToggleMessageReaction(w http.ResponseWriter, r *http.Request)
	ListPinnedChannelMessages(w http.ResponseWriter, r *http.Request)
	PinChannelMessage(w http.ResponseWriter, r *http.Request)
	UnpinChannelMessage(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	ms  *microservices.Clients
	log logger.Logger
}

// NewHandler constructs chat HTTP handlers.
func NewHandler(ms *microservices.Clients, log logger.Logger) Handler {
	return &handler{ms: ms, log: log}
}

func (h handler) requireChannelContext(w http.ResponseWriter, r *http.Request) (token, groupID, channelID string, ok bool) {
	token = authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return "", "", "", false
	}

	groupID = strings.TrimSpace(r.PathValue("groupId"))
	channelID = strings.TrimSpace(r.PathValue("channelId"))
	if groupID == "" || channelID == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "group_id"}))
		return "", "", "", false
	}

	return token, groupID, channelID, true
}

func (h handler) requireMessageContext(w http.ResponseWriter, r *http.Request) (token, groupID, channelID, messageID string, ok bool) {
	token, groupID, channelID, ok = h.requireChannelContext(w, r)
	if !ok {
		return "", "", "", "", false
	}
	messageID = strings.TrimSpace(r.PathValue("messageId"))
	if messageID == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "message_id"}))
		return "", "", "", "", false
	}

	return token, groupID, channelID, messageID, true
}

func (h handler) ListChannelMessages(w http.ResponseWriter, r *http.Request) {
	token, groupID, channelID, ok := h.requireChannelContext(w, r)
	if !ok {
		return
	}

	limit := int32(0)
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 32)
		if err != nil || parsed < 0 {
			authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "limit"}))
			return
		}
		limit = int32(parsed)
	}

	out, err := h.ms.Chat.GRPC.Message.ListChannelMessages(r.Context(), &chatv1.ListChannelMessagesRequest{
		AccessToken:     token,
		GroupId:         groupID,
		ChannelId:       channelID,
		Limit:           limit,
		BeforeCreatedAt: strings.TrimSpace(r.URL.Query().Get("before_created_at")),
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(chatHttp.MessagesListBody{
		Messages: chatHttp.MessagesFromPB(out.GetMessages()),
	}))
}

func (h handler) CreateChannelMessage(w http.ResponseWriter, r *http.Request) {
	token, groupID, channelID, ok := h.requireChannelContext(w, r)
	if !ok {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxMessageBodyBytes)
	var body chatHttp.CreateMessageBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, nil))
		return
	}

	out, err := h.ms.Chat.GRPC.Message.CreateChannelMessage(r.Context(), &chatv1.CreateChannelMessageRequest{
		AccessToken: token,
		GroupId:     groupID,
		ChannelId:   channelID,
		Body:        body.Body,
		ReplyToId:   body.ReplyToID,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusCreated, response.NewDataEnvelope(chatHttp.CreateMessageResponseBody{
		Message: chatHttp.MessageFromPB(out.GetMessage()),
	}))
}

func (h handler) UpdateChannelMessage(w http.ResponseWriter, r *http.Request) {
	token, groupID, channelID, messageID, ok := h.requireMessageContext(w, r)
	if !ok {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxMessageBodyBytes)
	var body chatHttp.UpdateMessageBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, nil))
		return
	}

	out, err := h.ms.Chat.GRPC.Message.UpdateChannelMessage(r.Context(), &chatv1.UpdateChannelMessageRequest{
		AccessToken: token,
		GroupId:     groupID,
		ChannelId:   channelID,
		MessageId:   messageID,
		Body:        body.Body,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(chatHttp.CreateMessageResponseBody{
		Message: chatHttp.MessageFromPB(out.GetMessage()),
	}))
}

func (h handler) DeleteChannelMessage(w http.ResponseWriter, r *http.Request) {
	token, groupID, channelID, messageID, ok := h.requireMessageContext(w, r)
	if !ok {
		return
	}

	_, err := h.ms.Chat.GRPC.Message.DeleteChannelMessage(r.Context(), &chatv1.DeleteChannelMessageRequest{
		AccessToken: token,
		GroupId:     groupID,
		ChannelId:   channelID,
		MessageId:   messageID,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.OKBody{OK: true}))
}

func (h handler) ToggleMessageReaction(w http.ResponseWriter, r *http.Request) {
	token, groupID, channelID, messageID, ok := h.requireMessageContext(w, r)
	if !ok {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1024)
	var body chatHttp.ToggleReactionBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, nil))
		return
	}

	out, err := h.ms.Chat.GRPC.Message.ToggleMessageReaction(r.Context(), &chatv1.ToggleMessageReactionRequest{
		AccessToken: token,
		GroupId:     groupID,
		ChannelId:   channelID,
		MessageId:   messageID,
		Emoji:       body.Emoji,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(chatHttp.ReactionsResponseBody{
		Reactions: chatHttp.ReactionsFromPB(out.GetReactions()),
	}))
}

func (h handler) ListPinnedChannelMessages(w http.ResponseWriter, r *http.Request) {
	token, groupID, channelID, ok := h.requireChannelContext(w, r)
	if !ok {
		return
	}

	out, err := h.ms.Chat.GRPC.Message.ListPinnedChannelMessages(r.Context(), &chatv1.ListPinnedChannelMessagesRequest{
		AccessToken: token,
		GroupId:     groupID,
		ChannelId:   channelID,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(chatHttp.MessagesListBody{
		Messages: chatHttp.MessagesFromPB(out.GetMessages()),
	}))
}

func (h handler) setChannelMessagePinned(w http.ResponseWriter, r *http.Request, pinned bool) {
	token, groupID, channelID, messageID, ok := h.requireMessageContext(w, r)
	if !ok {
		return
	}

	out, err := h.ms.Chat.GRPC.Message.SetChannelMessagePinned(r.Context(), &chatv1.SetChannelMessagePinnedRequest{
		AccessToken: token,
		GroupId:     groupID,
		ChannelId:   channelID,
		MessageId:   messageID,
		Pinned:      pinned,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(chatHttp.CreateMessageResponseBody{
		Message: chatHttp.MessageFromPB(out.GetMessage()),
	}))
}

func (h handler) PinChannelMessage(w http.ResponseWriter, r *http.Request) {
	h.setChannelMessagePinned(w, r, true)
}

func (h handler) UnpinChannelMessage(w http.ResponseWriter, r *http.Request) {
	h.setChannelMessagePinned(w, r, false)
}

func (h handler) UploadChannelAttachment(w http.ResponseWriter, r *http.Request) {
	token, groupID, channelID, ok := h.requireChannelContext(w, r)
	if !ok {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxAttachmentUploadBytes+maxMultipartOverhead)
	if err := r.ParseMultipartForm(maxAttachmentUploadBytes); err != nil {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "file"}))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "file"}))
		return
	}
	defer func() { _ = file.Close() }()

	data, err := io.ReadAll(io.LimitReader(file, maxAttachmentUploadBytes+1))
	if err != nil || len(data) == 0 || len(data) > maxAttachmentUploadBytes {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "file"}))
		return
	}

	contentType := ""
	filename := "file"
	if header != nil {
		if ct := strings.TrimSpace(header.Header.Get("Content-Type")); ct != "" {
			contentType = ct
		}
		if fn := strings.TrimSpace(header.Filename); fn != "" {
			filename = fn
		}
	}
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = http.DetectContentType(data)
	}

	out, err := h.ms.Chat.GRPC.Message.UploadChannelAttachment(r.Context(), &chatv1.UploadChannelAttachmentRequest{
		AccessToken: token,
		GroupId:     groupID,
		ChannelId:   channelID,
		Filename:    filename,
		ContentType: contentType,
		Data:        data,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusCreated, response.NewDataEnvelope(chatHttp.UploadAttachmentResponseBody{
		Attachment: chatHttp.AttachmentFromPB(out.GetAttachment()),
	}))
}
