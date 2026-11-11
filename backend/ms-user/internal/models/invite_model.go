package models

import "time"

// GroupMemberProfile is a member row with account fields for the share UI.
type GroupMemberProfile struct {
	UserID      string
	DisplayName string
	Email       string
	Role        string
	JoinedAt    time.Time
}

// GroupMemberCursor pages group_members by (created_at, user_id).
type GroupMemberCursor struct {
	CreatedAt time.Time
	UserID    string
}

// GroupMemberPage is a paginated member list.
type GroupMemberPage struct {
	Members    []GroupMemberProfile
	NextCursor *GroupMemberCursor
	Total      int
}

// UserSearchHit is one row from instance user search.
type UserSearchHit struct {
	UserID      string
	DisplayName string
	Email       string
	Username    string
}

// SendGroupInvitesInput is the batch share payload.
type SendGroupInvitesInput struct {
	GroupPublicID   string
	ChannelPublicID string
	UserIDs         []string
	Emails          []string
}

// SendGroupInvitesResult counts per SendGroupInvites.
type SendGroupInvitesResult struct {
	Added   int
	Emailed int
	Skipped int
}

// InviteLinkResult is the reusable copy-link token.
type InviteLinkResult struct {
	Token     string
	ExpiresAt time.Time
}

// RedeemInviteResult is membership outcome after redeem.
type RedeemInviteResult struct {
	GroupPublicID   string
	ChannelPublicID string
}

// InvitePreviewResult is public invite metadata for the landing page.
type InvitePreviewResult struct {
	Valid        bool
	GroupName    string
	ChannelName  string
	GroupLogoURL string
	MemberCount  int
	OnlineCount  int
	ExpiresAt    time.Time
	Reason       string
}
