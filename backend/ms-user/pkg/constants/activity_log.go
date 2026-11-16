package constants

// Universal user activity log categories.
const (
	ActivityCategoryAuth        = "auth"
	ActivityCategoryProfile     = "profile"
	ActivityCategoryPreferences = "preferences"
	ActivityCategoryAccount     = "account"
	ActivityCategorySecurity    = "security"
	ActivityCategorySession     = "session"
	ActivityCategoryGroup       = "group"
)

// Canonical user activity log actions.
const (
	ActivityUserRegistered             = "user.auth.registered"
	ActivityUserLoggedIn               = "user.auth.logged_in"
	ActivityUserLoggedOut              = "user.auth.logged_out"
	ActivityUserPasswordReset          = "user.auth.password_reset"
	ActivityUserAvatarUpdated          = "user.avatar.updated"
	ActivityUserProfileUpdated         = "user.profile.updated"
	ActivityUserStatusNoteUpdated      = "user.status_note.updated"
	ActivityUserStatusNoteCleared      = "user.status_note.cleared"
	ActivityUserPreferencesUpdated     = "user.preferences.updated"
	ActivityUserUsernameUpdated        = "user.username.updated"
	ActivityUserEmailUpdated           = "user.email.updated"
	ActivityUserPasswordUpdated        = "user.password.updated"
	ActivityUserMFATOTPEnabled         = "user.mfa.totp.enabled"
	ActivityUserMFATOTPDisabled        = "user.mfa.totp.disabled"
	ActivityUserMFAWebAuthnRegistered  = "user.mfa.webauthn.registered"
	ActivityUserMFAWebAuthnRenamed     = "user.mfa.webauthn.renamed"
	ActivityUserMFAWebAuthnDeleted     = "user.mfa.webauthn.deleted"
	ActivityUserMFARecoveryRegenerated = "user.mfa.recovery_codes.regenerated"
	ActivityUserSessionRevoked         = "user.session.revoked"
	ActivityUserSessionsRevokedOthers  = "user.sessions.revoked_others"
	ActivityUserAccountDisabled        = "user.account.disabled"
	ActivityUserAccountDeleted         = "user.account.deleted"
	ActivityGroupCreated               = "group.created"
	ActivityGroupJoined                = "group.joined"
	ActivityGroupLeft                  = "group.left"
	ActivityGroupKicked                = "group.kicked"
	ActivityGroupBanned                = "group.banned"
	ActivityGroupUnbanned              = "group.unbanned"
)

// ActivityTarget types.
const (
	ActivityTargetUser        = "user"
	ActivityTargetGroup       = "group"
	ActivityTargetSession     = "session"
	ActivityTargetSecurityKey = "security_key"
	ActivityTargetInvite      = "invite"
)

// Activity log query pagination limits.
const (
	DefaultActivityLogLimit = 30
	MaxActivityLogLimit     = 100
)

// Group activity log categories.
const (
	GroupAuditCategoryProfile    = "group"
	GroupAuditCategoryMember     = "member"
	GroupAuditCategoryRole       = "role"
	GroupAuditCategoryModeration = "moderation"
	GroupAuditCategoryInvite     = "invite"
)

// Canonical group activity log actions.
const (
	GroupAuditGroupCreated          = "group.created"
	GroupAuditGroupProfileUpdated   = "group.profile.updated"
	GroupAuditGroupIconUpdated      = "group.icon.updated"
	GroupAuditGroupBannerUpdated    = "group.banner.updated"
	GroupAuditGroupHideRolesUpdated = "group.settings.hide_roles_updated"
	GroupAuditGroupDeleted          = "group.deleted"
	GroupAuditMemberJoined          = "group.member.joined"
	GroupAuditMemberLeft            = "group.member.left"
	GroupAuditMemberKicked          = "group.member.kicked"
	GroupAuditMemberBanned          = "group.member.banned"
	GroupAuditMemberUnbanned        = "group.member.unbanned"
	GroupAuditMemberTimedOut        = "group.member.timed_out"
	GroupAuditMemberTimeoutRemoved  = "group.member.timeout_removed"
	GroupAuditRoleCreated           = "group.role.created"
	GroupAuditRoleUpdated           = "group.role.updated"
	GroupAuditRoleDuplicated        = "group.role.duplicated"
	GroupAuditRoleDeleted           = "group.role.deleted"
	GroupAuditMemberRoleAdded       = "group.member.role_added"
	GroupAuditMemberRoleRemoved     = "group.member.role_removed"
	GroupAuditInviteCreated         = "group.invite.created"
	GroupAuditInviteSent            = "group.invite.sent"
)

// Group activity target types.
const (
	GroupAuditTargetGroup  = "group"
	GroupAuditTargetMember = "member"
	GroupAuditTargetRole   = "role"
	GroupAuditTargetInvite = "invite"
)
