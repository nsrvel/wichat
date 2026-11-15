package models

import "time"

// ActivityFieldChange represents a field-level before/after diff in an audit log.
type ActivityFieldChange struct {
	Old any `json:"old,omitempty"`
	New any `json:"new,omitempty"`
}

// UserActivityLogEntry is the input payload to record a universal user audit event.
type UserActivityLogEntry struct {
	UserID     string
	ActorID    string
	Category   string
	Action     string
	TargetType string
	TargetID   string
	Summary    string
	Changes    map[string]ActivityFieldChange
	Metadata   map[string]any
	CreatedAt  time.Time
}

// UserActivityLog is a stored audit record for a user.
type UserActivityLog struct {
	ID         string
	UserID     string
	ActorID    string
	Category   string
	Action     string
	TargetType string
	TargetID   string
	Summary    string
	Changes    map[string]ActivityFieldChange
	Metadata   map[string]any
	CreatedAt  time.Time
}

// UserActivityLogPage is a cursor-paginated list of user activity logs.
type UserActivityLogPage struct {
	Logs       []UserActivityLog
	NextCursor string
}
