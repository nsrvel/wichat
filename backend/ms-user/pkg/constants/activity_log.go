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
