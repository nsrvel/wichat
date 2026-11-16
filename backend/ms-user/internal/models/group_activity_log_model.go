package models

import "time"

// GroupActivityLogEntry is the input payload to record a group audit event.
type GroupActivityLogEntry struct {
	GroupID    string
	ActorID    string
	Category   string
	Action     string
	TargetType string
	TargetID   string
	TargetName string
	Summary    string
	Changes    map[string]ActivityFieldChange
	Metadata   map[string]any
	CreatedAt  time.Time
}

// GroupActivityLog is a stored audit record for a group with hydrated actor fields.
type GroupActivityLog struct {
	ID               string
	GroupID          string
	ActorID          string
	ActorUsername    string
	ActorDisplayName string
	ActorAvatarURL   string
	Category         string
	Action           string
	TargetType       string
	TargetID         string
	TargetName       string
	Summary          string
	Changes          map[string]ActivityFieldChange
	Metadata         map[string]any
	CreatedAt        time.Time
}

// GroupActivityLogFilter scopes group audit log queries.
type GroupActivityLogFilter struct {
	GroupID  string
	ActorID  string
	Category string
	Action   string
}

// GroupActivityLogPage is a cursor-paginated list of group activity logs.
type GroupActivityLogPage struct {
	Logs       []GroupActivityLog
	NextCursor string
}
