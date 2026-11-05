package constants

// Registration modes (instance_settings.registration_mode).
const (
	RegistrationInviteOnly       = "invite_only"
	RegistrationApprovalRequired = "approval_required"
	RegistrationOpen             = "open"
)

// Group member roles (group_members.role).
const (
	GroupRoleOwner  = "owner"
	GroupRoleMember = "member"
)

// Default group_settings retention when creating a workspace.
const (
	DefaultGroupRetentionDays      = 0
	DefaultGroupTrashRetentionDays = 30
)
