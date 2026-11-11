package models

import (
	"time"

	"github.com/uptrace/bun"
)

// GroupRow maps groups for Bun.
type GroupRow struct {
	bun.BaseModel `bun:"table:groups"`

	ID        string    `bun:"id,pk,type:uuid"`
	PublicID  string    `bun:"public_id"`
	CreatedAt time.Time `bun:"created_at"`
	UpdatedAt time.Time `bun:"updated_at"`
}

// GroupSettingsRow maps group_settings for Bun.
type GroupSettingsRow struct {
	bun.BaseModel `bun:"table:group_settings"`

	GroupID            string    `bun:"group_id,pk,type:uuid"`
	Name               string    `bun:"name"`
	LogoURL            string    `bun:"logo_url"`
	Description        string    `bun:"description"`
	IsPrivate          bool      `bun:"is_private"`
	BannerColor        string    `bun:"banner_color"`
	BannerURL          string    `bun:"banner_url"`
	RetentionDays      int       `bun:"retention_days"`
	TrashRetentionDays int       `bun:"trash_retention_days"`
	CreatedAt          time.Time `bun:"created_at"`
	UpdatedAt          time.Time `bun:"updated_at"`
}

// GroupMemberRow maps group_members for Bun.
type GroupMemberRow struct {
	bun.BaseModel `bun:"table:group_members"`

	GroupID      string    `bun:"group_id,pk,type:uuid"`
	UserID       string    `bun:"user_id,pk,type:uuid"`
	Role         string    `bun:"role"`
	RailPosition *int      `bun:"rail_position"`
	CreatedAt    time.Time `bun:"created_at"`
}

// GroupMemberRowFrom maps domain membership to a Bun row.
func GroupMemberRowFrom(m GroupMember) GroupMemberRow {
	return GroupMemberRow{
		GroupID: m.GroupID, UserID: m.UserID, Role: m.Role,
		RailPosition: m.RailPosition, CreatedAt: m.CreatedAt,
	}
}
