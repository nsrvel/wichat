package models

import "time"

// Group is a workspace-equivalent chat unit (product name: Group).
type Group struct {
	ID                     string
	PublicID               string
	Name                   string
	LogoURL                string
	DefaultChannelPublicID string
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

// PublicGroup is the group view returned to clients.
type PublicGroup struct {
	ID               string
	Name             string
	LogoURL          string
	DefaultChannelID string
}

// Public returns client-safe group fields.
func (g Group) Public() PublicGroup {
	return PublicGroup{
		ID: g.PublicID, Name: g.Name, LogoURL: g.LogoURL, DefaultChannelID: g.DefaultChannelPublicID,
	}
}

// GroupMember links a user to a group with a preset role.
type GroupMember struct {
	GroupID   string
	UserID    string
	Role      string
	CreatedAt time.Time
}
