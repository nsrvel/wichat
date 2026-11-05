package models

import "time"

// Channel is workspace channel metadata (ms-user).
type Channel struct {
	ID        string
	PublicID  string
	GroupID   string
	Type      string
	Name      string
	CreatedBy string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// PublicChannel is returned to clients (JSON id is public_id).
type PublicChannel struct {
	ID   string
	Name string
	Type string
}

// Public returns client-safe channel fields.
func (c Channel) Public() PublicChannel {
	return PublicChannel{ID: c.PublicID, Name: c.Name, Type: c.Type}
}
