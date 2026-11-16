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
	HideRoles          bool      `bun:"hide_roles"`
	RetentionDays      int       `bun:"retention_days"`
	TrashRetentionDays int       `bun:"trash_retention_days"`
	CreatedAt          time.Time `bun:"created_at"`
	UpdatedAt          time.Time `bun:"updated_at"`
}

// GroupMemberRow maps group_members for Bun.
type GroupMemberRow struct {
	bun.BaseModel `bun:"table:group_members"`

	GroupID         string     `bun:"group_id,pk,type:uuid"`
	UserID          string     `bun:"user_id,pk,type:uuid"`
	Role            string     `bun:"role"`
	RailPosition    *int       `bun:"rail_position"`
	JoinMethod      string     `bun:"join_method"`
	InviteCode      string     `bun:"invite_code"`
	InvitedByUserID *string    `bun:"invited_by_user_id,type:uuid"`
	TimeoutUntil    *time.Time `bun:"timeout_until"`
	CreatedAt       time.Time  `bun:"created_at"`
}

// GroupMemberRowFrom maps domain membership to a Bun row.
func GroupMemberRowFrom(m GroupMember) GroupMemberRow {
	joinMethod := m.JoinMethod
	if joinMethod == "" {
		joinMethod = "unknown"
	}
	return GroupMemberRow{
		GroupID: m.GroupID, UserID: m.UserID, Role: m.Role,
		RailPosition: m.RailPosition, JoinMethod: joinMethod,
		InviteCode: m.InviteCode, InvitedByUserID: m.InvitedByUserID,
		TimeoutUntil: m.TimeoutUntil, CreatedAt: m.CreatedAt,
	}
}

// GroupRoleRow maps group_roles for Bun.
type GroupRoleRow struct {
	bun.BaseModel `bun:"table:group_roles"`

	ID          string    `bun:"id,pk,type:uuid"`
	PublicID    string    `bun:"public_id"`
	GroupID     string    `bun:"group_id,type:uuid"`
	Slug        string    `bun:"slug"`
	Name        string    `bun:"name"`
	Color       string    `bun:"color"`
	Hoist       bool      `bun:"hoist"`
	Mentionable bool      `bun:"mentionable"`
	IsDefault   bool      `bun:"is_default"`
	IsSystem    bool      `bun:"is_system"`
	Position    int       `bun:"position"`
	Permissions []string  `bun:"permissions,array"`
	CreatedAt   time.Time `bun:"created_at"`
	UpdatedAt   time.Time `bun:"updated_at"`
}

// GroupMemberRoleRow maps group_member_roles for Bun.
type GroupMemberRoleRow struct {
	bun.BaseModel `bun:"table:group_member_roles"`

	GroupID   string    `bun:"group_id,pk,type:uuid"`
	UserID    string    `bun:"user_id,pk,type:uuid"`
	RoleID    string    `bun:"role_id,pk,type:uuid"`
	CreatedAt time.Time `bun:"created_at"`
}

// GroupBanRow maps group_bans for Bun.
type GroupBanRow struct {
	bun.BaseModel `bun:"table:group_bans"`

	GroupID              string     `bun:"group_id,pk,type:uuid"`
	UserID               string     `bun:"user_id,pk,type:uuid"`
	BannedBy             *string    `bun:"banned_by,type:uuid"`
	Reason               string     `bun:"reason"`
	DeleteMessageSeconds int        `bun:"delete_message_seconds"`
	ExpiresAt            *time.Time `bun:"expires_at"`
	CreatedAt            time.Time  `bun:"created_at"`
}
