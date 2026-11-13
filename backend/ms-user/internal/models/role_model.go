package models

import "time"

// GroupRole represents an RBAC role within a group.
type GroupRole struct {
	ID          string
	PublicID    string
	GroupID     string
	Slug        string
	Name        string
	Color       string
	Hoist       bool
	Mentionable bool
	IsDefault   bool
	IsSystem    bool
	Position    int
	Permissions []string
	MemberCount int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// GroupRoleSummary is a light role summary used in member cards and responses.
type GroupRoleSummary struct {
	ID    string
	Slug  string
	Name  string
	Color string
}

// GroupRolePatch represents editable role properties.
type GroupRolePatch struct {
	Name        *string
	Color       *string
	Hoist       *bool
	Mentionable *bool
	Permissions []string
}

// GroupBan represents a banned user from a group.
type GroupBan struct {
	GroupID              string
	UserID               string
	BannedBy             string
	Reason               string
	DeleteMessageSeconds int
	CreatedAt            time.Time
}
