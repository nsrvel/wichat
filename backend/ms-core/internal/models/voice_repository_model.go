package models

import (
	"time"

	"github.com/uptrace/bun"
)

// CallSessionRow maps the call_sessions table.
type CallSessionRow struct {
	bun.BaseModel           `bun:"table:call_sessions,alias:cs"`
	ID                      string     `bun:"id,pk,type:uuid"`
	PublicID                string     `bun:"public_id"`
	GroupID                 *string    `bun:"group_id,type:uuid"`
	ChannelID               string     `bun:"channel_id,type:uuid"`
	RoomName                string     `bun:"room_name"`
	StartedAt               time.Time  `bun:"started_at"`
	EndedAt                 *time.Time `bun:"ended_at"`
	PeakParticipants        int32      `bun:"peak_participants"`
	TotalParticipantSeconds int64      `bun:"total_participant_seconds"`
}

// ToModel converts CallSessionRow to domain CallSession.
func (r CallSessionRow) ToModel() CallSession {
	return CallSession{
		ID:                      r.ID,
		PublicID:                r.PublicID,
		GroupID:                 r.GroupID,
		ChannelID:               r.ChannelID,
		RoomName:                r.RoomName,
		StartedAt:               r.StartedAt,
		EndedAt:                 r.EndedAt,
		PeakParticipants:        r.PeakParticipants,
		TotalParticipantSeconds: r.TotalParticipantSeconds,
	}
}

// CallParticipantRow maps the call_participants table.
type CallParticipantRow struct {
	bun.BaseModel    `bun:"table:call_participants,alias:cp"`
	ID               string     `bun:"id,pk,type:uuid"`
	SessionID        string     `bun:"session_id,type:uuid"`
	UserID           string     `bun:"user_id,type:uuid"`
	JoinedAt         time.Time  `bun:"joined_at"`
	LeftAt           *time.Time `bun:"left_at"`
	DurationSeconds  int32      `bun:"duration_seconds"`
	ServerMuted      bool       `bun:"server_muted"`
	ServerDeafened   bool       `bun:"server_deafened"`
	DisconnectReason string     `bun:"disconnect_reason"`
}

// ToModel converts CallParticipantRow to domain CallParticipant.
func (r CallParticipantRow) ToModel() CallParticipant {
	return CallParticipant{
		ID:               r.ID,
		SessionID:        r.SessionID,
		UserID:           r.UserID,
		JoinedAt:         r.JoinedAt,
		LeftAt:           r.LeftAt,
		DurationSeconds:  r.DurationSeconds,
		ServerMuted:      r.ServerMuted,
		ServerDeafened:   r.ServerDeafened,
		DisconnectReason: r.DisconnectReason,
	}
}

// CallRecordingRow maps the call_recordings table.
type CallRecordingRow struct {
	bun.BaseModel   `bun:"table:call_recordings,alias:cr"`
	ID              string     `bun:"id,pk,type:uuid"`
	PublicID        string     `bun:"public_id"`
	SessionID       string     `bun:"session_id,type:uuid"`
	ChannelID       string     `bun:"channel_id,type:uuid"`
	EgressID        string     `bun:"egress_id"`
	Status          string     `bun:"status"`
	FileKey         string     `bun:"file_key"`
	FileSizeBytes   int64      `bun:"file_size_bytes"`
	DurationSeconds int32      `bun:"duration_seconds"`
	StartedBy       string     `bun:"started_by,type:uuid"`
	StoppedBy       *string    `bun:"stopped_by,type:uuid"`
	StartedAt       time.Time  `bun:"started_at"`
	EndedAt         *time.Time `bun:"ended_at"`
}

// ToModel converts CallRecordingRow to domain CallRecording.
func (r CallRecordingRow) ToModel() CallRecording {
	return CallRecording{
		ID:              r.ID,
		PublicID:        r.PublicID,
		SessionID:       r.SessionID,
		ChannelID:       r.ChannelID,
		EgressID:        r.EgressID,
		Status:          r.Status,
		FileKey:         r.FileKey,
		FileSizeBytes:   r.FileSizeBytes,
		DurationSeconds: r.DurationSeconds,
		StartedBy:       r.StartedBy,
		StoppedBy:       r.StoppedBy,
		StartedAt:       r.StartedAt,
		EndedAt:         r.EndedAt,
	}
}
