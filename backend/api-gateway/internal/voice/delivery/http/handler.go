package http

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	authHttp "github.com/wichat/wichat/backend/api-gateway/pkg/auth-http"
	voiceHttp "github.com/wichat/wichat/backend/api-gateway/pkg/voice-http"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices"
	corev1 "github.com/wichat/wichat/backend/wi-shared/models/gen/core/v1"
	"github.com/wichat/wichat/backend/wi-shared/response"
	responsehttp "github.com/wichat/wichat/backend/wi-shared/response/http"
)

const maxWebhookBodyBytes = 1 << 20 // 1 MiB

// Handler is the public voice HTTP delivery.
type Handler interface {
	IssueVoiceToken(w http.ResponseWriter, r *http.Request)
	ListGroupOccupancy(w http.ResponseWriter, r *http.Request)
	ModerateParticipant(w http.ResponseWriter, r *http.Request)
	AcknowledgeIdleWarning(w http.ResponseWriter, r *http.Request)
	StartRecording(w http.ResponseWriter, r *http.Request)
	StopRecording(w http.ResponseWriter, r *http.Request)
	ListChannelRecordings(w http.ResponseWriter, r *http.Request)
	GetQuotaUsage(w http.ResponseWriter, r *http.Request)
	HandleWebhook(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	ms  *microservices.Clients
	log logger.Logger
}

// NewHandler wires voice HTTP to the ms-core client.
func NewHandler(ms *microservices.Clients, log logger.Logger) Handler {
	return handler{ms: ms, log: log}
}

func (h handler) requireCoreReady(w http.ResponseWriter, r *http.Request) bool {
	if h.ms == nil || h.ms.Core.GRPC.Voice == nil {
		authHttp.WriteOutbound(w, r, h.log, exception.Internal(errors.New("ms-core unavailable")))
		return false
	}
	return true
}

func (h handler) requireBearerAndChannel(w http.ResponseWriter, r *http.Request) (token, channelID string, ok bool) {
	token = authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return "", "", false
	}
	channelID = strings.TrimSpace(r.PathValue("channelId"))
	if channelID == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "channelId"}))
		return "", "", false
	}
	if !h.requireCoreReady(w, r) {
		return "", "", false
	}
	return token, channelID, true
}

func (h handler) IssueVoiceToken(w http.ResponseWriter, r *http.Request) {
	token, channelID, ok := h.requireBearerAndChannel(w, r)
	if !ok {
		return
	}

	groupID := strings.TrimSpace(r.PathValue("groupId"))
	if groupID == "" {
		groupID = strings.TrimSpace(r.URL.Query().Get("group_id"))
	}
	if groupID == "" && r.ContentLength > 0 {
		var body voiceHttp.ChannelVoiceContextBody
		if err := authHttp.DecodeJSON(w, r, &body); err != nil {
			authHttp.WriteOutbound(w, r, h.log, err)
			return
		}
		groupID = strings.TrimSpace(body.GroupID)
	}

	out, err := h.ms.Core.GRPC.Voice.IssueVoiceToken(r.Context(), &corev1.IssueVoiceTokenRequest{
		AccessToken: token,
		GroupId:     groupID,
		ChannelId:   channelID,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(voiceHttp.IssueVoiceTokenBody{
		Token:              out.GetToken(),
		LiveKitURL:         out.GetLivekitUrl(),
		RoomName:           out.GetRoomName(),
		ChannelID:          out.GetChannelId(),
		ExpiresAt:          out.GetExpiresAt(),
		InitialServerMuted: out.GetInitialServerMuted(),
		CanPublishAudio:    out.GetCanPublishAudio(),
		CanPublishVideo:    out.GetCanPublishVideo(),
		UseVoiceActivity:   out.GetUseVoiceActivity(),
		BitrateKbps:        out.GetBitrateKbps(),
		VideoQuality:       out.GetVideoQuality(),
		IsRecording:        out.GetIsRecording(),
		Occupants:          voiceHttp.OccupantsFromPB(out.GetOccupants()),
	}))
}

func (h handler) ListGroupOccupancy(w http.ResponseWriter, r *http.Request) {
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}
	groupID := strings.TrimSpace(r.PathValue("groupId"))
	if groupID == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "groupId"}))
		return
	}
	if !h.requireCoreReady(w, r) {
		return
	}

	out, err := h.ms.Core.GRPC.Voice.ListGroupOccupancy(r.Context(), &corev1.ListGroupOccupancyRequest{
		AccessToken: token,
		GroupId:     groupID,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	channels := make([]voiceHttp.ChannelOccupancyBody, 0, len(out.GetChannels()))
	for _, ch := range out.GetChannels() {
		channels = append(channels, voiceHttp.ChannelOccupancyBody{
			ChannelID:   ch.GetChannelId(),
			IsRecording: ch.GetIsRecording(),
			Occupants:   voiceHttp.OccupantsFromPB(ch.GetOccupants()),
		})
	}
	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(voiceHttp.GroupOccupancyBody{
		Channels: channels,
	}))
}

func (h handler) ModerateParticipant(w http.ResponseWriter, r *http.Request) {
	token, channelID, ok := h.requireBearerAndChannel(w, r)
	if !ok {
		return
	}

	var body voiceHttp.ModerateParticipantRequestBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	groupID := strings.TrimSpace(body.GroupID)
	if groupID == "" {
		groupID = strings.TrimSpace(r.URL.Query().Get("group_id"))
	}

	out, err := h.ms.Core.GRPC.Voice.ModerateParticipant(r.Context(), &corev1.ModerateParticipantRequest{
		AccessToken:  token,
		GroupId:      groupID,
		ChannelId:    channelID,
		TargetUserId: strings.TrimSpace(body.TargetUserID),
		Action:       strings.TrimSpace(body.Action),
		Reason:       strings.TrimSpace(body.Reason),
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(voiceHttp.ModerateParticipantResponseBody{
		OK:       out.GetOk(),
		Occupant: voiceHttp.OccupantFromPB(out.GetOccupant()),
	}))
}

func (h handler) AcknowledgeIdleWarning(w http.ResponseWriter, r *http.Request) {
	token, channelID, ok := h.requireBearerAndChannel(w, r)
	if !ok {
		return
	}

	groupID := strings.TrimSpace(r.URL.Query().Get("group_id"))
	if groupID == "" && r.ContentLength > 0 {
		var body voiceHttp.ChannelVoiceContextBody
		if err := authHttp.DecodeJSON(w, r, &body); err != nil {
			authHttp.WriteOutbound(w, r, h.log, err)
			return
		}
		groupID = strings.TrimSpace(body.GroupID)
	}

	out, err := h.ms.Core.GRPC.Voice.AcknowledgeIdleWarning(r.Context(), &corev1.AcknowledgeIdleWarningRequest{
		AccessToken: token,
		GroupId:     groupID,
		ChannelId:   channelID,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(map[string]bool{"ok": out.GetOk()}))
}

func (h handler) StartRecording(w http.ResponseWriter, r *http.Request) {
	token, channelID, ok := h.requireBearerAndChannel(w, r)
	if !ok {
		return
	}

	groupID := strings.TrimSpace(r.URL.Query().Get("group_id"))
	if groupID == "" && r.ContentLength > 0 {
		var body voiceHttp.ChannelVoiceContextBody
		if err := authHttp.DecodeJSON(w, r, &body); err != nil {
			authHttp.WriteOutbound(w, r, h.log, err)
			return
		}
		groupID = strings.TrimSpace(body.GroupID)
	}

	out, err := h.ms.Core.GRPC.Voice.StartRecording(r.Context(), &corev1.StartRecordingRequest{
		AccessToken: token,
		GroupId:     groupID,
		ChannelId:   channelID,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(voiceHttp.RecordingFromPB(out.GetRecording())))
}

func (h handler) StopRecording(w http.ResponseWriter, r *http.Request) {
	token, channelID, ok := h.requireBearerAndChannel(w, r)
	if !ok {
		return
	}

	groupID := strings.TrimSpace(r.URL.Query().Get("group_id"))
	recordingID := strings.TrimSpace(r.URL.Query().Get("recording_id"))
	if r.ContentLength > 0 {
		var body voiceHttp.ChannelVoiceContextBody
		if err := authHttp.DecodeJSON(w, r, &body); err != nil {
			authHttp.WriteOutbound(w, r, h.log, err)
			return
		}
		if groupID == "" {
			groupID = strings.TrimSpace(body.GroupID)
		}
		if recordingID == "" {
			recordingID = strings.TrimSpace(body.RecordingID)
		}
	}

	out, err := h.ms.Core.GRPC.Voice.StopRecording(r.Context(), &corev1.StopRecordingRequest{
		AccessToken: token,
		GroupId:     groupID,
		ChannelId:   channelID,
		RecordingId: recordingID,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(voiceHttp.RecordingFromPB(out.GetRecording())))
}

func (h handler) ListChannelRecordings(w http.ResponseWriter, r *http.Request) {
	token, channelID, ok := h.requireBearerAndChannel(w, r)
	if !ok {
		return
	}

	groupID := strings.TrimSpace(r.URL.Query().Get("group_id"))
	var limit int32 = 25
	if rawLimit := strings.TrimSpace(r.URL.Query().Get("limit")); rawLimit != "" {
		if parsed, parseErr := strconv.ParseInt(rawLimit, 10, 32); parseErr == nil && parsed > 0 {
			if parsed > 100 {
				parsed = 100
			}
			limit = int32(parsed)
		}
	}

	out, err := h.ms.Core.GRPC.Voice.ListChannelRecordings(r.Context(), &corev1.ListChannelRecordingsRequest{
		AccessToken: token,
		GroupId:     groupID,
		ChannelId:   channelID,
		Limit:       limit,
		Cursor:      strings.TrimSpace(r.URL.Query().Get("cursor")),
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	recordings := make([]voiceHttp.CallRecordingBody, 0, len(out.GetRecordings()))
	for _, rec := range out.GetRecordings() {
		recordings = append(recordings, voiceHttp.RecordingFromPB(rec))
	}
	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(voiceHttp.RecordingsListBody{
		Recordings: recordings,
		NextCursor: out.GetNextCursor(),
	}))
}

func (h handler) GetQuotaUsage(w http.ResponseWriter, r *http.Request) {
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}
	if !h.requireCoreReady(w, r) {
		return
	}

	out, err := h.ms.Core.GRPC.Voice.GetQuotaUsage(r.Context(), &corev1.GetQuotaUsageRequest{
		AccessToken: token,
		GroupId:     strings.TrimSpace(r.URL.Query().Get("group_id")),
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(voiceHttp.QuotaUsageBody{
		Month:                   out.GetMonth(),
		UsedParticipantMinutes:  out.GetUsedParticipantMinutes(),
		LimitParticipantMinutes: out.GetLimitParticipantMinutes(),
		UsagePercent:            out.GetUsagePercent(),
		WarningActive:           out.GetWarningActive(),
	}))
}

func (h handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if !h.requireCoreReady(w, r) {
		return
	}

	rawBody, err := io.ReadAll(io.LimitReader(r.Body, maxWebhookBodyBytes))
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, nil))
		return
	}

	authHeader := r.Header.Get("Authorization")
	if _, err := h.ms.Core.GRPC.Voice.HandleWebhook(r.Context(), &corev1.HandleWebhookRequest{
		AuthHeader: authHeader,
		RawBody:    rawBody,
	}); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(map[string]bool{"ok": true}))
}
