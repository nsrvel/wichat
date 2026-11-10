package constants

// Group invite row kinds (group_invites.kind).
const (
	GroupInviteKindEmail     = "email"
	GroupInviteKindShareLink = "share_link"
)

// GroupInviteTokenLen is the canonical share-link token length (6 random bytes, base64url).
const GroupInviteTokenLen = 8

// Invite preview invalid reasons (public API).
const (
	InvitePreviewNotFound = "not_found"
	InvitePreviewExpired  = "expired"
	InvitePreviewRedeemed = "redeemed"
)
