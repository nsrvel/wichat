package models

import "time"

// ChannelCategory groups channels inside a workspace sidebar.
type ChannelCategory struct {
	ID        string
	PublicID  string
	GroupID   string
	Name      string
	IsPrivate bool
	Position  int32
	CreatedBy string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// PublicChannelCategory is returned to clients (JSON id is public_id).
type PublicChannelCategory struct {
	ID        string
	Name      string
	IsPrivate bool
	Position  int32
}

// Public returns client-safe category fields.
func (c ChannelCategory) Public() PublicChannelCategory {
	return PublicChannelCategory{
		ID: c.PublicID, Name: c.Name, IsPrivate: c.IsPrivate, Position: c.Position,
	}
}

// PublicChannelCategoryPermissionOverride is a role or member permission override on a category or channel.
type PublicChannelCategoryPermissionOverride struct {
	TargetType       string
	TargetID         string
	TargetName       string
	TargetColor      string
	TargetAvatarURL  string
	IsEveryone       bool
	AllowPermissions []string
	DenyPermissions  []string
}

// ChannelCategoryPermissionOverride is internal category permission override metadata.
type ChannelCategoryPermissionOverride struct {
	CategoryID       string
	TargetType       string
	TargetID         string
	IsEveryone       bool
	AllowPermissions []string
	DenyPermissions  []string
}
