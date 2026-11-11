package models

import "time"

// Group is a workspace-equivalent chat unit (product name: Group).
type Group struct {
	ID                     string
	PublicID               string
	Name                   string
	LogoURL                string
	Description            string
	IsPrivate              bool
	BannerColor            string
	BannerURL              string
	DefaultChannelPublicID string
	CallerRole             string
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

// GroupProfilePatch is a full scalar profile replace from clients (owner save).
type GroupProfilePatch struct {
	Name             string
	Description      string
	IsPrivate        bool
	BannerColor      string
	BannerURL        string
	LogoURL          string
	ClearBannerAsset bool
}

// PublicGroup is the group view returned to clients.
type PublicGroup struct {
	ID               string
	Name             string
	LogoURL          string
	Description      string
	IsPrivate        bool
	BannerColor      string
	BannerURL        string
	DefaultChannelID string
	Role             string
}

// Public returns client-safe group fields.
func (g Group) Public() PublicGroup {
	return PublicGroup{
		ID:               g.PublicID,
		Name:             g.Name,
		LogoURL:          g.LogoURL,
		Description:      g.Description,
		IsPrivate:        g.IsPrivate,
		BannerColor:      g.BannerColor,
		BannerURL:        g.BannerURL,
		DefaultChannelID: g.DefaultChannelPublicID,
		Role:             g.CallerRole,
	}
}

// GroupMember links a user to a group with a preset role.
type GroupMember struct {
	GroupID      string
	UserID       string
	Role         string
	RailPosition *int
	CreatedAt    time.Time
}
