package models

import (
	"time"

	"github.com/uptrace/bun"
)

// GroupInviteRow maps group_invites for Bun.
type GroupInviteRow struct {
	bun.BaseModel `bun:"table:group_invites"`

	ID               string     `bun:"id,pk,type:uuid"`
	Token            string     `bun:"token"`
	GroupID          string     `bun:"group_id,type:uuid"`
	ChannelID        *string    `bun:"channel_id,type:uuid,nullzero"`
	Email            *string    `bun:"email,nullzero"`
	InvitedByUserID  string     `bun:"invited_by_user_id,type:uuid"`
	Kind             string     `bun:"kind"`
	ExpiresAt        time.Time  `bun:"expires_at"`
	RedeemedAt       *time.Time `bun:"redeemed_at,nullzero"`
	RedeemedByUserID *string    `bun:"redeemed_by_user_id,type:uuid,nullzero"`
	CreatedAt        time.Time  `bun:"created_at"`
}
