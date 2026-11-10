package models

import "time"

// Channel is workspace channel metadata (ms-user).
type Channel struct {
	ID             string
	PublicID       string
	GroupID        string
	Type           string
	IsPrivate      bool
	Name           string
	ImageURL       string
	SidebarSection string
	CreatedBy      string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// ChannelAccess holds internal ids after token, membership, and channel resolution.
type ChannelAccess struct {
	ChannelID string
	UserID    string
}

// PublicChannel is returned to clients (JSON id is public_id).
type PublicChannel struct {
	ID             string
	Name           string
	Type           string
	IsPrivate      bool
	ImageURL       string
	SidebarSection string
}

// Public returns client-safe channel fields.
func (c Channel) Public() PublicChannel {
	return PublicChannel{
		ID: c.PublicID, Name: c.Name, Type: c.Type, IsPrivate: c.IsPrivate,
		ImageURL: c.ImageURL, SidebarSection: c.SidebarSection,
	}
}
