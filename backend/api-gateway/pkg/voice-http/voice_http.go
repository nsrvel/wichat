// Package voiceHttp maps voice JSON bodies and protobuf summaries for HTTP responses.
package voiceHttp

import (
	corev1 "github.com/wichat/wichat/backend/wi-shared/models/gen/core/v1"
)

// VoiceOccupantBody is one active participant in a voice channel.
type VoiceOccupantBody struct {
	UserID           string `json:"user_id"`
	DisplayName      string `json:"display_name"`
	Username         string `json:"username"`
	AvatarURL        string `json:"avatar_url,omitempty"`
	IsMuted          bool   `json:"is_muted"`
	IsDeafened       bool   `json:"is_deafened"`
	IsServerMuted    bool   `json:"is_server_muted"`
	IsServerDeafened bool   `json:"is_server_deafened"`
	IsCameraOn       bool   `json:"is_camera_on"`
	IsScreenSharing  bool   `json:"is_screen_sharing"`
	JoinedAt         string `json:"joined_at"`
}

// ChannelOccupancyBody summarizes live occupants and recording status for a channel.
type ChannelOccupancyBody struct {
	ChannelID   string              `json:"channel_id"`
	IsRecording bool                `json:"is_recording"`
	Occupants   []VoiceOccupantBody `json:"occupants"`
}

// CallRecordingBody summarizes one call recording.
type CallRecordingBody struct {
	ID              string `json:"id"`
	SessionID       string `json:"session_id"`
	ChannelID       string `json:"channel_id"`
	EgressID        string `json:"egress_id"`
	Status          string `json:"status"`
	FileKey         string `json:"file_key,omitempty"`
	FileSizeBytes   int64  `json:"file_size_bytes"`
	DurationSeconds int32  `json:"duration_seconds"`
	StartedBy       string `json:"started_by"`
	StoppedBy       string `json:"stopped_by,omitempty"`
	StartedAt       string `json:"started_at"`
	EndedAt         string `json:"ended_at,omitempty"`
	DownloadURL     string `json:"download_url,omitempty"`
}

// IssueVoiceTokenBody is returned by POST voice/token.
type IssueVoiceTokenBody struct {
	Token              string              `json:"token"`
	LiveKitURL         string              `json:"livekit_url"`
	RoomName           string              `json:"room_name"`
	ChannelID          string              `json:"channel_id"`
	ExpiresAt          string              `json:"expires_at"`
	InitialServerMuted bool                `json:"initial_server_muted"`
	CanPublishAudio    bool                `json:"can_publish_audio"`
	CanPublishVideo    bool                `json:"can_publish_video"`
	UseVoiceActivity   bool                `json:"use_voice_activity"`
	BitrateKbps        int32               `json:"bitrate_kbps"`
	VideoQuality       string              `json:"video_quality"`
	IsRecording        bool                `json:"is_recording"`
	Occupants          []VoiceOccupantBody `json:"occupants"`
}

// GroupOccupancyBody wraps channel occupancy rows for a group.
type GroupOccupancyBody struct {
	Channels []ChannelOccupancyBody `json:"channels"`
}

// ModerateParticipantRequestBody is the JSON body for POST voice/moderate.
type ModerateParticipantRequestBody struct {
	GroupID      string `json:"group_id,omitempty"`
	TargetUserID string `json:"target_user_id"`
	Action       string `json:"action"`
	Reason       string `json:"reason,omitempty"`
}

// ModerateParticipantResponseBody wraps moderation result.
type ModerateParticipantResponseBody struct {
	OK       bool              `json:"ok"`
	Occupant VoiceOccupantBody `json:"occupant"`
}

// ChannelVoiceContextBody carries optional group_id for channel-scoped POST endpoints.
type ChannelVoiceContextBody struct {
	GroupID     string `json:"group_id,omitempty"`
	RecordingID string `json:"recording_id,omitempty"`
}

// RecordingsListBody wraps a list of call recordings.
type RecordingsListBody struct {
	Recordings []CallRecordingBody `json:"recordings"`
	NextCursor string              `json:"next_cursor,omitempty"`
}

// QuotaUsageBody wraps monthly participant-minutes quota status.
type QuotaUsageBody struct {
	Month                   string  `json:"month"`
	UsedParticipantMinutes  int64   `json:"used_participant_minutes"`
	LimitParticipantMinutes int64   `json:"limit_participant_minutes"`
	UsagePercent            float64 `json:"usage_percent"`
	WarningActive           bool    `json:"warning_active"`
}

// OccupantFromPB converts a protobuf VoiceOccupant to JSON body.
func OccupantFromPB(pb *corev1.VoiceOccupant) VoiceOccupantBody {
	if pb == nil {
		return VoiceOccupantBody{}
	}
	return VoiceOccupantBody{
		UserID:           pb.GetUserId(),
		DisplayName:      pb.GetDisplayName(),
		Username:         pb.GetUsername(),
		AvatarURL:        pb.GetAvatarUrl(),
		IsMuted:          pb.GetIsMuted(),
		IsDeafened:       pb.GetIsDeafened(),
		IsServerMuted:    pb.GetIsServerMuted(),
		IsServerDeafened: pb.GetIsServerDeafened(),
		IsCameraOn:       pb.GetIsCameraOn(),
		IsScreenSharing:  pb.GetIsScreenSharing(),
		JoinedAt:         pb.GetJoinedAt(),
	}
}

// OccupantsFromPB converts a slice of protobuf VoiceOccupant to JSON bodies.
func OccupantsFromPB(in []*corev1.VoiceOccupant) []VoiceOccupantBody {
	out := make([]VoiceOccupantBody, 0, len(in))
	for _, o := range in {
		out = append(out, OccupantFromPB(o))
	}
	return out
}

// RecordingFromPB converts a protobuf CallRecordingSummary to JSON body.
func RecordingFromPB(pb *corev1.CallRecordingSummary) CallRecordingBody {
	if pb == nil {
		return CallRecordingBody{}
	}
	return CallRecordingBody{
		ID:              pb.GetId(),
		SessionID:       pb.GetSessionId(),
		ChannelID:       pb.GetChannelId(),
		EgressID:        pb.GetEgressId(),
		Status:          pb.GetStatus(),
		FileKey:         pb.GetFileKey(),
		FileSizeBytes:   pb.GetFileSizeBytes(),
		DurationSeconds: pb.GetDurationSeconds(),
		StartedBy:       pb.GetStartedBy(),
		StoppedBy:       pb.GetStoppedBy(),
		StartedAt:       pb.GetStartedAt(),
		EndedAt:         pb.GetEndedAt(),
		DownloadURL:     pb.GetDownloadUrl(),
	}
}
