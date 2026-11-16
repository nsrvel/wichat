package constants

// Group invite row kinds (group_invites.kind).
const (
	GroupInviteKindEmail     = "email"
	GroupInviteKindShareLink = "share_link"
)

// Group member join methods (group_members.join_method).
const (
	JoinMethodOwner      = "owner"
	JoinMethodOpen       = "open"
	JoinMethodInvite     = "invite"
	JoinMethodInviteLink = "invite_link"
	JoinMethodUnknown    = "unknown"
)

// GroupInviteTokenLen is the canonical share-link token length (6 random bytes, base64url).
const GroupInviteTokenLen = 8

// Invite preview invalid reasons (public API).
const (
	InvitePreviewNotFound = "not_found"
	InvitePreviewExpired  = "expired"
	InvitePreviewRedeemed = "redeemed"
)
