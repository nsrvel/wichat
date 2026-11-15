package models

import (
	"encoding/json"
	"time"

	"github.com/uptrace/bun"
)

// UserActivityLogRow maps public.user_activity_logs for Bun.
type UserActivityLogRow struct {
	bun.BaseModel `bun:"table:user_activity_logs"`

	ID         string          `bun:"id,pk,type:uuid"`
	UserID     string          `bun:"user_id,type:uuid"`
	ActorID    *string         `bun:"actor_id,type:uuid"`
	Category   string          `bun:"category"`
	Action     string          `bun:"action"`
	TargetType string          `bun:"target_type"`
	TargetID   string          `bun:"target_id"`
	Summary    string          `bun:"summary"`
	Changes    json.RawMessage `bun:"changes,type:jsonb"`
	Metadata   json.RawMessage `bun:"metadata,type:jsonb"`
	CreatedAt  time.Time       `bun:"created_at"`
}

// ToModel converts a database row into a domain UserActivityLog.
func (r UserActivityLogRow) ToModel() UserActivityLog {
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

	return UserActivityLog{
		ID:         r.ID,
		UserID:     r.UserID,
		ActorID:    actorID,
		Category:   r.Category,
		Action:     r.Action,
		TargetType: r.TargetType,
		TargetID:   r.TargetID,
		Summary:    r.Summary,
		Changes:    changes,
		Metadata:   metadata,
		CreatedAt:  r.CreatedAt,
	}
}
