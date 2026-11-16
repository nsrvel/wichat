package models

import (
	"encoding/json"
	"time"

	"github.com/uptrace/bun"
)

// GroupActivityLogRow maps public.group_activity_logs for Bun.
type GroupActivityLogRow struct {
	bun.BaseModel `bun:"table:group_activity_logs,alias:gal"`

	ID         string          `bun:"id,pk,type:uuid"`
	GroupID    string          `bun:"group_id,type:uuid"`
	ActorID    *string         `bun:"actor_id,type:uuid"`
	Category   string          `bun:"category"`
	Action     string          `bun:"action"`
	TargetType string          `bun:"target_type"`
	TargetID   string          `bun:"target_id"`
	TargetName string          `bun:"target_name"`
	Summary    string          `bun:"summary"`
	Changes    json.RawMessage `bun:"changes,type:jsonb"`
	Metadata   json.RawMessage `bun:"metadata,type:jsonb"`
	CreatedAt  time.Time       `bun:"created_at"`
}

// GroupActivityLogReadRow extends GroupActivityLogRow with joined actor columns.
type GroupActivityLogReadRow struct {
	bun.BaseModel `bun:"table:group_activity_logs,alias:gal"`

	ID               string          `bun:"id,pk,type:uuid"`
	GroupID          string          `bun:"group_id,type:uuid"`
	ActorID          *string         `bun:"actor_id,type:uuid"`
	ActorUsername    string          `bun:"actor_username"`
	ActorDisplayName string          `bun:"actor_display_name"`
	ActorAvatarURL   string          `bun:"actor_avatar_url"`
	Category         string          `bun:"category"`
	Action           string          `bun:"action"`
	TargetType       string          `bun:"target_type"`
	TargetID         string          `bun:"target_id"`
	TargetName       string          `bun:"target_name"`
	Summary          string          `bun:"summary"`
	Changes          json.RawMessage `bun:"changes,type:jsonb"`
	Metadata         json.RawMessage `bun:"metadata,type:jsonb"`
	CreatedAt        time.Time       `bun:"created_at"`
}

// ToModel converts a read row into a domain GroupActivityLog.
func (r GroupActivityLogReadRow) ToModel() GroupActivityLog {
	actorID := ""
	if r.ActorID != nil {
		actorID = *r.ActorID
	}

	var changes map[string]ActivityFieldChange
	if len(r.Changes) > 0 {
		_ = json.Unmarshal(r.Changes, &changes)
	}
	if changes == nil {
		changes = map[string]ActivityFieldChange{}
	}

	var metadata map[string]any
	if len(r.Metadata) > 0 {
		_ = json.Unmarshal(r.Metadata, &metadata)
	}
	if metadata == nil {
		metadata = map[string]any{}
	}

	return GroupActivityLog{
		ID:               r.ID,
		GroupID:          r.GroupID,
		ActorID:          actorID,
		ActorUsername:    r.ActorUsername,
		ActorDisplayName: r.ActorDisplayName,
		ActorAvatarURL:   r.ActorAvatarURL,
		Category:         r.Category,
		Action:           r.Action,
		TargetType:       r.TargetType,
		TargetID:         r.TargetID,
		TargetName:       r.TargetName,
		Summary:          r.Summary,
		Changes:          changes,
		Metadata:         metadata,
		CreatedAt:        r.CreatedAt.UTC(),
	}
}
