package models

import "time"

// PresenceStatus values returned to clients or stored as manual preference.
const (
	PresenceStatusOnline    = "online"
	PresenceStatusIdle      = "idle"
	PresenceStatusDND       = "dnd"
	PresenceStatusInvisible = "invisible"
	PresenceStatusOffline   = "offline"
)

// PresenceActivity values reported by WebSocket heartbeats.
const (
	PresenceActivityActive = "active"
	PresenceActivityIdle   = "idle"
)

// ManualPresenceState represents a user's explicit status preference in Redis.
type ManualPresenceState struct {
	Status    string
	ExpiresAt *time.Time
}

// UserPresenceSnapshot is the effective presence state visible to other group members.
type UserPresenceSnapshot struct {
	IsOnline       bool
	PresenceStatus string
}

// PresenceTransition is returned when a heartbeat, disconnect, or manual status update changes state.
type PresenceTransition struct {
	UserID          string
	PreviousStatus  string
	EffectiveStatus string
	IsOnline        bool
	GroupIDs        []string
}

// MyPresenceView is returned by GetMyPresence for the authenticated caller.
type MyPresenceView struct {
	ManualStatus    string
	EffectiveStatus string
	IsOnline        bool
	ExpiresAt       *time.Time
}

// SetManualPresenceResult is returned by SetManualPresence.
type SetManualPresenceResult struct {
	UserID          string
	ManualStatus    string
	EffectiveStatus string
	IsOnline        bool
	ExpiresAt       *time.Time
	GroupIDs        []string
}
