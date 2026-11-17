package models

import "time"

// VoiceOccupant represents one live participant in a voice channel.
type VoiceOccupant struct {
	UserID           string    `json:"user_id"`
	DisplayName      string    `json:"display_name"`
	Username         string    `json:"username"`
	AvatarURL        string    `json:"avatar_url,omitempty"`
	IsMuted          bool      `json:"is_muted"`
	IsDeafened       bool      `json:"is_deafened"`
	IsServerMuted    bool      `json:"is_server_muted"`
	IsServerDeafened bool      `json:"is_server_deafened"`
	IsCameraOn       bool      `json:"is_camera_on"`
	IsScreenSharing  bool      `json:"is_screen_sharing"`
	JoinedAt         string    `json:"joined_at"`
	LastActiveAt     int64     `json:"last_active_at"`
	IdleWarned       bool      `json:"idle_warned"`
	JoinedAtTime     time.Time `json:"-"`
}

// RoomMeta stores active voice room metadata in Redis.
type RoomMeta struct {
	SessionID           string
	GroupPublicID       string
	ChannelPublicID     string
	RoomName            string
	StartedAt           string
	IsRecording         bool
	ActiveRecordingID   string
	VideoPublisherCount int
}

// ChannelOccupancy summarizes live occupants and recording state for one channel.
type ChannelOccupancy struct {
	ChannelID   string
	IsRecording bool
	Occupants   []VoiceOccupant
}

// CallSession is a voice/video call session stored in PostgreSQL.
type CallSession struct {
	ID                      string
	PublicID                string
	GroupID                 *string
	ChannelID               string
	RoomName                string
	StartedAt               time.Time
	EndedAt                 *time.Time
	PeakParticipants        int32
	TotalParticipantSeconds int64
}

// CallParticipant is a participant join/leave record stored in PostgreSQL.
type CallParticipant struct {
	ID               string
	SessionID        string
	UserID           string
	JoinedAt         time.Time
	LeftAt           *time.Time
	DurationSeconds  int32
	ServerMuted      bool
	ServerDeafened   bool
	DisconnectReason string
}

// CallRecording is a LiveKit Egress recording stored in PostgreSQL.
type CallRecording struct {
	ID              string
	PublicID        string
	SessionID       string
	ChannelID       string
	EgressID        string
	Status          string
	FileKey         string
	FileSizeBytes   int64
	DurationSeconds int32
	StartedBy       string
	StoppedBy       *string
	StartedAt       time.Time
	EndedAt         *time.Time
	DownloadURL     string
}

// IssueVoiceTokenInput holds parameters for IssueVoiceToken.
type IssueVoiceTokenInput struct {
	AccessToken     string
	GroupPublicID   string
	ChannelPublicID string
}

// IssueVoiceTokenResult is returned by IssueVoiceToken.
type IssueVoiceTokenResult struct {
	Token              string
	LiveKitURL         string
	RoomName           string
	ChannelID          string
	ExpiresAt          time.Time
	InitialServerMuted bool
	CanPublishAudio    bool
	CanPublishVideo    bool
	UseVoiceActivity   bool
	BitrateKbps        int32
	VideoQuality       string
	IsRecording        bool
	Occupants          []VoiceOccupant
}

// ModerateParticipantInput holds parameters for ModerateParticipant.
type ModerateParticipantInput struct {
	AccessToken     string
	GroupPublicID   string
	ChannelPublicID string
	TargetUserID    string
	Action          string
	Reason          string
}

// QuotaUsageResult summarizes monthly LiveKit participant-minutes usage.
type QuotaUsageResult struct {
	Month                   string
	UsedParticipantMinutes  int64
	LimitParticipantMinutes int64
	UsagePercent            float64
	WarningActive           bool
}
