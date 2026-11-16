package models

import "time"

// Channel is workspace channel metadata (ms-user).
type Channel struct {
	ID               string
	PublicID         string
	GroupID          string
	CategoryID       string
	CategoryPublicID string
	Type             string
	IsPrivate        bool
	Name             string
	ImageURL         string
	SidebarSection   string
	Position         int32
	CreatedBy        string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

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

// ChannelAccess holds internal ids after token, membership, and channel resolution.
type ChannelAccess struct {
	ChannelID   string
	UserID      string
	Permissions []string
}

// PublicChannel is returned to clients (JSON id is public_id).
type PublicChannel struct {
	ID             string
	Name           string
	Type           string
	IsPrivate      bool
	ImageURL       string
	SidebarSection string
	CategoryID     string
	Position       int32
}

// PublicChannelCategory is returned to clients (JSON id is public_id).
type PublicChannelCategory struct {
	ID        string
	Name      string
	IsPrivate bool
	Position  int32
}

// Public returns client-safe channel fields.
func (c Channel) Public() PublicChannel {
	return PublicChannel{
		ID: c.PublicID, Name: c.Name, Type: c.Type, IsPrivate: c.IsPrivate,
		ImageURL: c.ImageURL, SidebarSection: c.SidebarSection, CategoryID: c.CategoryPublicID,
		Position: c.Position,
	}
}

// Public returns client-safe category fields.
func (c ChannelCategory) Public() PublicChannelCategory {
	return PublicChannelCategory{
		ID: c.PublicID, Name: c.Name, IsPrivate: c.IsPrivate, Position: c.Position,
	}
}
