// Package authHttp maps auth JSON bodies and protobuf sessions for HTTP responses (dir: auth-http, file: auth_http.go).
package authHttp

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/wichat/wichat/backend/api-gateway/pkg/cookie"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	"github.com/wichat/wichat/backend/wi-shared/response"
	responsegrpc "github.com/wichat/wichat/backend/wi-shared/response/grpc"
	responsehttp "github.com/wichat/wichat/backend/wi-shared/response/http"
)

const maxBodyBytes = 1 << 20

// MaxGroupIconBytes is the largest accepted group icon upload.
const MaxGroupIconBytes = 5 * 1024 * 1024

// MaxUserAvatarBytes is the largest accepted user avatar upload.
const MaxUserAvatarBytes = 5 * 1024 * 1024

// MaxGroupBannerBytes is the largest accepted group banner upload.
const MaxGroupBannerBytes = 8 * 1024 * 1024

// SetupBody is the deprecated JSON body for POST /api/v1/auth/setup.
type SetupBody = RegisterBody

// RegisterBody is the JSON body for POST /api/v1/auth/register.
type RegisterBody struct {
	Email          string `json:"email"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	DisplayName    string `json:"display_name"`
	BirthDate      string `json:"birth_date"`
	Locale         string `json:"locale"`
	InviteToken    string `json:"invite_token"`
	TurnstileToken string `json:"turnstile_token"`
}

// LoginBody is the JSON body for POST /api/v1/auth/login.
type LoginBody struct {
	Login          string `json:"login"`
	Email          string `json:"email"`
	Password       string `json:"password"`
	Locale         string `json:"locale"`
	TurnstileToken string `json:"turnstile_token"`
}

// AuthChallengeBody is the email OTP or MFA step before a session is issued.
type AuthChallengeBody struct {
	OTPRequired        bool            `json:"otp_required"`
	MFARequired        bool            `json:"mfa_required,omitempty"`
	MFAMethods         []string        `json:"mfa_methods,omitempty"`
	WebAuthnOptions    json.RawMessage `json:"webauthn_options,omitempty"`
	ChallengeID        string          `json:"challenge_id"`
	EmailHint          string          `json:"email_hint,omitempty"`
	ResendAfterSeconds int32           `json:"resend_after_seconds"`
}

// VerifyEmailCodeBody is POST /api/v1/auth/register/verify and login/verify.
type VerifyEmailCodeBody struct {
	ChallengeID    string `json:"challenge_id"`
	Code           string `json:"code"`
	TurnstileToken string `json:"turnstile_token"`
}

// ResendEmailCodeBody is POST resend endpoints.
type ResendEmailCodeBody struct {
	ChallengeID    string `json:"challenge_id"`
	TurnstileToken string `json:"turnstile_token"`
}

// ResendEmailCodeData is resend success payload.
type ResendEmailCodeData struct {
	ResendAfterSeconds int32 `json:"resend_after_seconds"`
}

// SetupStatusBody is the JSON data for GET /api/v1/auth/setup-status (instance policy).
type SetupStatusBody struct {
	RegistrationMode  string `json:"registration_mode"`
	InviteOnly        bool   `json:"invite_only"`
	TurnstileSiteKey  string `json:"turnstile_site_key,omitempty"`
	TurnstileRequired bool   `json:"turnstile_required,omitempty"`
	MfaEnabled        bool   `json:"mfa_enabled"`
}

// InstanceSetupBody is POST /api/v1/auth/instance-setup.
type InstanceSetupBody struct {
	Name             string `json:"name"`
	LogoURL          string `json:"logo_url"`
	RegistrationMode string `json:"registration_mode"`
}

// CheckUsernameBody is GET /api/v1/users/check-username.
type CheckUsernameBody struct {
	Available  bool   `json:"available"`
	Normalized string `json:"normalized"`
	Reason     string `json:"reason,omitempty"`
}

// SessionBody is the public session JSON (refresh token is HttpOnly cookie only).
type SessionBody struct {
	AccessToken string   `json:"access_token"`
	ExpiresIn   int32    `json:"expires_in"`
	User        UserBody `json:"user"`
}

// UserBody is the public user slice in auth responses.
type UserBody struct {
	ID                  string  `json:"id"`
	Email               string  `json:"email"`
	DisplayName         string  `json:"display_name"`
	Username            string  `json:"username"`
	AvatarURL           string  `json:"avatar_url"`
	Pronouns            string  `json:"pronouns"`
	Bio                 string  `json:"bio"`
	BannerColor         string  `json:"banner_color"`
	StatusNoteText      string  `json:"status_note_text"`
	StatusNoteEmoji     string  `json:"status_note_emoji"`
	StatusNoteExpiresAt *string `json:"status_note_expires_at"`
	Locale              string  `json:"locale"`
	TimeFormat          string  `json:"time_format"`
	CreatedAt           string  `json:"created_at"`
	HasPassword         bool    `json:"has_password"`
}

// UpdateUserProfileBody is PATCH /api/v1/users/me/profile.
type UpdateUserProfileBody struct {
	DisplayName string `json:"display_name"`
	Pronouns    string `json:"pronouns"`
	Bio         string `json:"bio"`
	BannerColor string `json:"banner_color"`
	ClearAvatar bool   `json:"clear_avatar"`
}

// UpdateUserStatusNoteBody is PATCH /api/v1/users/me/status-note.
type UpdateUserStatusNoteBody struct {
	Text      string  `json:"text"`
	Emoji     string  `json:"emoji"`
	ExpiresAt *string `json:"expires_at"`
}

// UpdateUserPreferencesBody is PATCH /api/v1/users/me/preferences.
type UpdateUserPreferencesBody struct {
	Locale     string `json:"locale"`
	TimeFormat string `json:"time_format"`
}

// UpdateUsernameBody is PATCH /api/v1/users/me/username.
type UpdateUsernameBody struct {
	Username        string `json:"username"`
	CurrentPassword string `json:"current_password"`
}

// RequestEmailChangeBody is POST /api/v1/users/me/email/request.
type RequestEmailChangeBody struct {
	Locale string `json:"locale"`
}

// VerifyEmailChangeBody is POST /api/v1/users/me/email/verify.
type VerifyEmailChangeBody struct {
	ChallengeID string `json:"challenge_id"`
	Code        string `json:"code"`
}

// VerifyEmailChangeData is the payload returned by POST /api/v1/users/me/email/verify.
type VerifyEmailChangeData struct {
	ChangeToken string `json:"change_token"`
}

// CompleteEmailChangeBody is PATCH /api/v1/users/me/email.
type CompleteEmailChangeBody struct {
	ChangeToken     string `json:"change_token"`
	NewEmail        string `json:"new_email"`
	CurrentPassword string `json:"current_password"`
	Reason          string `json:"reason"`
}

// UpdatePasswordBody is PATCH /api/v1/users/me/password.
type UpdatePasswordBody struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// VerifyLoginMFABody is POST /api/v1/auth/login/mfa/verify.
type VerifyLoginMFABody struct {
	ChallengeID    string          `json:"challenge_id"`
	Method         string          `json:"method"`
	Code           string          `json:"code"`
	Assertion      json.RawMessage `json:"assertion"`
	TurnstileToken string          `json:"turnstile_token"`
}

// BeginLoginWebAuthnBody is POST /api/v1/auth/login/mfa/webauthn/begin.
type BeginLoginWebAuthnBody struct {
	ChallengeID string `json:"challenge_id"`
}

// WebAuthnOptionsData wraps WebAuthn options JSON.
type WebAuthnOptionsData struct {
	WebAuthnOptions json.RawMessage `json:"webauthn_options"`
}

// SecurityKeyBody is a registered WebAuthn security key.
type SecurityKeyBody struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	CreatedAt  string   `json:"created_at"`
	LastUsedAt string   `json:"last_used_at,omitempty"`
	Transports []string `json:"transports"`
}

// MFAStatusBody is GET /api/v1/users/me/mfa.
type MFAStatusBody struct {
	TOTPEnabled            bool              `json:"totp_enabled"`
	SecurityKeys           []SecurityKeyBody `json:"security_keys"`
	RecoveryCodesRemaining int32             `json:"recovery_codes_remaining"`
}

// CurrentPasswordBody is used for password-gated MFA actions.
type CurrentPasswordBody struct {
	CurrentPassword string `json:"current_password"`
}

// TOTPSetupBeginData is POST /api/v1/users/me/mfa/totp/begin.
type TOTPSetupBeginData struct {
	SetupToken string `json:"setup_token"`
	Secret     string `json:"secret"`
	OtpauthURI string `json:"otpauth_uri"`
}

// TOTPSetupConfirmBody is POST /api/v1/users/me/mfa/totp/confirm.
type TOTPSetupConfirmBody struct {
	SetupToken string `json:"setup_token"`
	Code       string `json:"code"`
}

// RecoveryCodesData wraps one-time recovery codes.
type RecoveryCodesData struct {
	RecoveryCodes []string `json:"recovery_codes"`
}

// BeginWebAuthnRegisterBody is POST /api/v1/users/me/mfa/security-keys/begin.
type BeginWebAuthnRegisterBody struct {
	CurrentPassword string `json:"current_password"`
	Name            string `json:"name"`
}

// BeginWebAuthnRegisterData is returned when beginning security key registration.
type BeginWebAuthnRegisterData struct {
	ChallengeID string          `json:"challenge_id"`
	Options     json.RawMessage `json:"options"`
}

// FinishWebAuthnRegisterBody is POST /api/v1/users/me/mfa/security-keys/finish.
type FinishWebAuthnRegisterBody struct {
	ChallengeID string          `json:"challenge_id"`
	Credential  json.RawMessage `json:"credential"`
}

// FinishWebAuthnRegisterData is returned after saving a security key.
type FinishWebAuthnRegisterData struct {
	SecurityKey   SecurityKeyBody `json:"security_key"`
	RecoveryCodes []string        `json:"recovery_codes,omitempty"`
}

// RenameSecurityKeyBody is PATCH /api/v1/users/me/mfa/security-keys/{id}.
type RenameSecurityKeyBody struct {
	Name string `json:"name"`
}

// OKBody is a minimal success payload.
type OKBody struct {
	OK bool `json:"ok"`
}

// GroupBody is a public group in JSON responses.
type GroupBody struct {
	ID                   string   `json:"id"`
	Name                 string   `json:"name"`
	LogoURL              string   `json:"logo_url"`
	Description          string   `json:"description"`
	IsPrivate            bool     `json:"is_private"`
	BannerColor          string   `json:"banner_color"`
	BannerURL            string   `json:"banner_url"`
	HideRoles            bool     `json:"hide_roles"`
	DefaultChannelID     string   `json:"default_channel_id"`
	Role                 string   `json:"role,omitempty"`
	EffectivePermissions []string `json:"effective_permissions"`
}

// UpdateGroupProfileBody is PATCH /api/v1/groups/{id}/profile.
type UpdateGroupProfileBody struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	IsPrivate        bool   `json:"is_private"`
	BannerColor      string `json:"banner_color"`
	ClearLogo        bool   `json:"clear_logo"`
	ClearBannerImage bool   `json:"clear_banner_image"`
}

// ChannelBody is channel metadata in JSON responses.
type ChannelBody struct {
	ID                    string   `json:"id"`
	Name                  string   `json:"name"`
	Type                  string   `json:"type"`
	SidebarSection        string   `json:"sidebar_section"`
	ImageURL              string   `json:"image_url"`
	IsPrivate             bool     `json:"is_private"`
	CategoryID            string   `json:"category_id"`
	Position              int      `json:"position"`
	SlowmodeSeconds       int      `json:"slowmode_seconds"`
	ThreadSlowmodeSeconds int      `json:"thread_slowmode_seconds"`
	ContentVisibility     string   `json:"content_visibility"`
	ForumDefaultLayout    string   `json:"forum_default_layout"`
	ForumSortOrder        string   `json:"forum_sort_order"`
	ForumTagMatching      string   `json:"forum_tag_matching"`
	BitrateKbps           int      `json:"bitrate_kbps"`
	VideoQuality          string   `json:"video_quality"`
	UserLimit             int      `json:"user_limit"`
	RTCRegion             string   `json:"rtc_region"`
	CallerPermissions     []string `json:"caller_permissions"`
	DmPeerUserID          string   `json:"dm_peer_user_id,omitempty"`
}

// UpdateGroupChannelBody is PATCH /api/v1/groups/{id}/channels/{channelId}.
type UpdateGroupChannelBody struct {
	Name                  *string `json:"name,omitempty"`
	IsPrivate             *bool   `json:"is_private,omitempty"`
	SlowmodeSeconds       *int32  `json:"slowmode_seconds,omitempty"`
	ThreadSlowmodeSeconds *int32  `json:"thread_slowmode_seconds,omitempty"`
	ContentVisibility     *string `json:"content_visibility,omitempty"`
	ForumDefaultLayout    *string `json:"forum_default_layout,omitempty"`
	ForumSortOrder        *string `json:"forum_sort_order,omitempty"`
	ForumTagMatching      *string `json:"forum_tag_matching,omitempty"`
	BitrateKbps           *int32  `json:"bitrate_kbps,omitempty"`
	VideoQuality          *string `json:"video_quality,omitempty"`
	UserLimit             *int32  `json:"user_limit,omitempty"`
	RTCRegion             *string `json:"rtc_region,omitempty"`
}

// ChannelCategoryBody is category metadata in JSON responses.
type ChannelCategoryBody struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	IsPrivate bool   `json:"is_private"`
	Position  int    `json:"position"`
}

// CreateChannelCategoryBody is POST /api/v1/groups/{id}/categories.
type CreateChannelCategoryBody struct {
	Name      string `json:"name"`
	IsPrivate bool   `json:"is_private"`
}

// UpdateChannelCategoryBody is PATCH /api/v1/groups/{id}/categories/{categoryId}.
type UpdateChannelCategoryBody struct {
	Name      *string `json:"name,omitempty"`
	IsPrivate *bool   `json:"is_private,omitempty"`
}

// ChannelCategoryPermissionOverrideBody is a role or member permission override on a category.
type ChannelCategoryPermissionOverrideBody struct {
	TargetType       string   `json:"target_type"`
	TargetID         string   `json:"target_id"`
	TargetName       string   `json:"target_name"`
	TargetColor      string   `json:"target_color"`
	TargetAvatarURL  string   `json:"target_avatar_url"`
	IsEveryone       bool     `json:"is_everyone"`
	AllowPermissions []string `json:"allow_permissions"`
	DenyPermissions  []string `json:"deny_permissions"`
}

// CategoryPermissionsListResponse is GET /api/v1/groups/{id}/categories/{categoryId}/permissions.
type CategoryPermissionsListResponse struct {
	Overrides []ChannelCategoryPermissionOverrideBody `json:"overrides"`
}

// UpsertCategoryPermissionOverrideBody is PUT /api/v1/groups/{id}/categories/{categoryId}/permissions.
type UpsertCategoryPermissionOverrideBody struct {
	TargetType       string   `json:"target_type"`
	TargetID         string   `json:"target_id"`
	AllowPermissions []string `json:"allow_permissions"`
	DenyPermissions  []string `json:"deny_permissions"`
}

// CategoryPermissionOverrideResponse wraps one updated permission override.
type CategoryPermissionOverrideResponse struct {
	Override ChannelCategoryPermissionOverrideBody `json:"override"`
}

// CategoryPermissionOverrideFromPB maps a proto override to JSON.
func CategoryPermissionOverrideFromPB(o *userv1.ChannelCategoryPermissionOverride) ChannelCategoryPermissionOverrideBody {
	if o == nil {
		return ChannelCategoryPermissionOverrideBody{AllowPermissions: []string{}, DenyPermissions: []string{}}
	}
	allow := o.GetAllowPermissions()
	if allow == nil {
		allow = []string{}
	}
	deny := o.GetDenyPermissions()
	if deny == nil {
		deny = []string{}
	}
	return ChannelCategoryPermissionOverrideBody{
		TargetType:       o.GetTargetType(),
		TargetID:         o.GetTargetId(),
		TargetName:       o.GetTargetName(),
		TargetColor:      o.GetTargetColor(),
		TargetAvatarURL:  o.GetTargetAvatarUrl(),
		IsEveryone:       o.GetIsEveryone(),
		AllowPermissions: allow,
		DenyPermissions:  deny,
	}
}

// CreateGroupChannelBody is POST /api/v1/groups/{id}/channels.
type CreateGroupChannelBody struct {
	Name       string  `json:"name"`
	Type       string  `json:"type"`
	CategoryID *string `json:"category_id,omitempty"`
	IsPrivate  bool    `json:"is_private"`
}

// ChannelResponse wraps a single created channel.
type ChannelResponse struct {
	Channel ChannelBody `json:"channel"`
}

// EnsureGroupDmBody is POST /api/v1/groups/{id}/dm.
type EnsureGroupDmBody struct {
	UserID string `json:"user_id"`
}

// EnsureGroupDmResponse is the ensure DM result.
type EnsureGroupDmResponse struct {
	Channel ChannelBody `json:"channel"`
	Created bool        `json:"created"`
}

// ReorderChannelCategoriesBody is PUT /api/v1/groups/{id}/categories/order.
type ReorderChannelCategoriesBody struct {
	CategoryIDs []string `json:"category_ids"`
}

// ReorderGroupChannelsBody is PUT /api/v1/groups/{id}/channels/order.
type ReorderGroupChannelsBody struct {
	ChannelIDs     []string `json:"channel_ids"`
	CategoryID     string   `json:"category_id,omitempty"`
	SidebarSection string   `json:"sidebar_section,omitempty"`
}

// ChannelCategoryResponse wraps a single created category.
type ChannelCategoryResponse struct {
	Category ChannelCategoryBody `json:"category"`
}

// ChannelsListBody wraps a channel and category list.
type ChannelsListBody struct {
	Channels   []ChannelBody         `json:"channels"`
	Categories []ChannelCategoryBody `json:"categories"`
}

// GroupsListBody wraps a group list.
type GroupsListBody struct {
	Groups []GroupBody `json:"groups"`
}

// CreateGroupBody is POST /api/v1/groups.
type CreateGroupBody struct {
	Name     string `json:"name"`
	Template string `json:"template,omitempty"`
}

// ReorderMyGroupsBody is PUT /api/v1/groups/mine/order.
type ReorderMyGroupsBody struct {
	GroupIDs []string `json:"group_ids"`
}

// GroupMemberRoleSummaryBody summarizes an assigned role for a group member.
type GroupMemberRoleSummaryBody struct {
	ID       string `json:"id"`
	Slug     string `json:"slug"`
	Name     string `json:"name"`
	Color    string `json:"color"`
	Hoist    bool   `json:"hoist"`
	Position int32  `json:"position"`
}

// GroupMemberBody is one row in the share access list.
type GroupMemberBody struct {
	UserID               string                       `json:"user_id"`
	DisplayName          string                       `json:"display_name"`
	Username             string                       `json:"username"`
	AvatarURL            string                       `json:"avatar_url"`
	Role                 string                       `json:"role"`
	JoinedAt             string                       `json:"joined_at"`
	Roles                []GroupMemberRoleSummaryBody `json:"roles,omitempty"`
	IsOnline             bool                         `json:"is_online"`
	PresenceStatus       string                       `json:"presence_status"`
	BannerColor          string                       `json:"banner_color,omitempty"`
	Bio                  string                       `json:"bio,omitempty"`
	Pronouns             string                       `json:"pronouns,omitempty"`
	UserCreatedAt        string                       `json:"user_created_at,omitempty"`
	StatusNoteText       string                       `json:"status_note_text,omitempty"`
	StatusNoteEmoji      string                       `json:"status_note_emoji,omitempty"`
	StatusNoteExpiresAt  string                       `json:"status_note_expires_at,omitempty"`
	JoinMethod           string                       `json:"join_method,omitempty"`
	InviteCode           string                       `json:"invite_code,omitempty"`
	InvitedByUserID      string                       `json:"invited_by_user_id,omitempty"`
	InvitedByUsername    string                       `json:"invited_by_username,omitempty"`
	InvitedByDisplayName string                       `json:"invited_by_display_name,omitempty"`
	TimeoutUntil         string                       `json:"timeout_until,omitempty"`
}

// MyPresenceBody is GET/PUT /api/v1/users/me/presence.
type MyPresenceBody struct {
	ManualStatus    string `json:"manual_status"`
	EffectiveStatus string `json:"effective_status"`
	IsOnline        bool   `json:"is_online"`
	ExpiresAt       string `json:"expires_at,omitempty"`
}

// DegradedMyPresenceBody is the safe fallback when live presence cannot be loaded.
func DegradedMyPresenceBody() MyPresenceBody {
	return MyPresenceBody{
		ManualStatus:    "online",
		EffectiveStatus: "offline",
		IsOnline:        false,
	}
}

// UpdateMyPresenceRequestBody is PUT /api/v1/users/me/presence.
type UpdateMyPresenceRequestBody struct {
	Status     string `json:"status"`
	TTLSeconds int64  `json:"ttl_seconds,omitempty"`
}

// GroupMembersPageBody is GET /api/v1/groups/{id}/members.
type GroupMembersPageBody struct {
	Members    []GroupMemberBody `json:"members"`
	NextCursor string            `json:"next_cursor,omitempty"`
	Total      int               `json:"total"`
	Degraded   bool              `json:"degraded,omitempty"`
}

// DegradedGroupMembersPageBody keeps HTTP 200 when ms-user is unavailable so clients can retain cached rows.
func DegradedGroupMembersPageBody() GroupMembersPageBody {
	return GroupMembersPageBody{
		Members:  []GroupMemberBody{},
		Total:    0,
		Degraded: true,
	}
}

// UserSearchHitBody is one autocomplete row.
type UserSearchHitBody struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
	Username    string `json:"username"`
	AvatarURL   string `json:"avatar_url"`
}

// UserSearchBody is GET /api/v1/users/search.
type UserSearchBody struct {
	Users []UserSearchHitBody `json:"users"`
}

// GroupInviteLinkBody is GET /api/v1/groups/{id}/invite-link.
type GroupInviteLinkBody struct {
	Token     string `json:"token"`
	URL       string `json:"url"`
	ExpiresAt string `json:"expires_at"`
}

// SendGroupInvitesBody is POST /api/v1/groups/{id}/invites.
type SendGroupInvitesBody struct {
	UserIDs        []string `json:"user_ids"`
	Emails         []string `json:"emails"`
	ChannelID      string   `json:"channel_id,omitempty"`
	TurnstileToken string   `json:"turnstile_token"`
}

// SendGroupInvitesResultBody is the invite batch outcome.
type SendGroupInvitesResultBody struct {
	Added   int `json:"added"`
	Emailed int `json:"emailed"`
	Skipped int `json:"skipped"`
}

// InvitePreviewBody is GET /api/v1/invites/preview.
type InvitePreviewBody struct {
	Valid            bool   `json:"valid"`
	IsPrivate        bool   `json:"is_private,omitempty"`
	GroupName        string `json:"group_name,omitempty"`
	GroupDescription string `json:"group_description,omitempty"`
	ChannelName      string `json:"channel_name,omitempty"`
	GroupLogoURL     string `json:"group_logo_url"`
	GroupBannerURL   string `json:"group_banner_url,omitempty"`
	GroupBannerColor string `json:"group_banner_color,omitempty"`
	MemberCount      int    `json:"member_count"`
	OnlineCount      int    `json:"online_count"`
	ExpiresAt        string `json:"expires_at,omitempty"`
	Reason           string `json:"reason,omitempty"`
}

// RedeemInviteBody is POST /api/v1/invites/redeem.
type RedeemInviteBody struct {
	Token          string `json:"token"`
	TurnstileToken string `json:"turnstile_token"`
}

// RedeemInviteResultBody is membership after redeem.
type RedeemInviteResultBody struct {
	GroupID   string `json:"group_id"`
	ChannelID string `json:"channel_id,omitempty"`
}

// ForgotPasswordRequestData is POST /api/v1/auth/forgot-password success body.
type ForgotPasswordRequestData struct {
	OK                    bool `json:"ok"`
	ResendCooldownSeconds int  `json:"resend_cooldown_seconds"`
}

// ForgotPasswordRequestBody is POST /api/v1/auth/forgot-password.
type ForgotPasswordRequestBody struct {
	Email          string `json:"email"`
	Locale         string `json:"locale"`
	TurnstileToken string `json:"turnstile_token"`
}

// ForgotPasswordVerifyBody is POST /api/v1/auth/forgot-password/verify.
type ForgotPasswordVerifyBody struct {
	Email          string `json:"email"`
	Code           string `json:"code"`
	TurnstileToken string `json:"turnstile_token"`
}

// ForgotPasswordResetBody is POST /api/v1/auth/forgot-password/reset.
type ForgotPasswordResetBody struct {
	ResetToken string `json:"reset_token"`
	Password   string `json:"password"`
}

// ForgotPasswordVerifyData is the verify step response data.
type ForgotPasswordVerifyData struct {
	ResetToken string `json:"reset_token"`
}

// DecodeJSON reads a JSON body with a size limit.
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return exception.InvalidInput(exception.MsgInvalidInput, nil)
	}

	return nil
}

// Bearer returns the token from Authorization: Bearer.
func Bearer(r *http.Request) string {
	header := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if len(header) < len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return ""
	}

	return strings.TrimSpace(header[len(prefix):])
}

// GroupFromPB maps a group summary proto to JSON.
func GroupFromPB(g *userv1.GroupSummary) GroupBody {
	if g == nil {
		return GroupBody{}
	}

	return GroupBody{
		ID:                   g.GetId(),
		Name:                 g.GetName(),
		LogoURL:              g.GetLogoUrl(),
		Description:          g.GetDescription(),
		IsPrivate:            g.GetIsPrivate(),
		BannerColor:          g.GetBannerColor(),
		BannerURL:            g.GetBannerUrl(),
		HideRoles:            g.GetHideRoles(),
		DefaultChannelID:     g.GetDefaultChannelId(),
		Role:                 g.GetRole(),
		EffectivePermissions: nonNilStrings(g.GetEffectivePermissions()),
	}
}

// GroupRoleBody is a group role in JSON responses.
type GroupRoleBody struct {
	ID          string   `json:"id"`
	Slug        string   `json:"slug"`
	Name        string   `json:"name"`
	Color       string   `json:"color"`
	Hoist       bool     `json:"hoist"`
	Mentionable bool     `json:"mentionable"`
	IsDefault   bool     `json:"is_default"`
	IsSystem    bool     `json:"is_system"`
	Position    int32    `json:"position"`
	Permissions []string `json:"permissions"`
	MemberCount int32    `json:"member_count"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

// GroupRolesListBody is GET /api/v1/groups/{id}/roles.
type GroupRolesListBody struct {
	Roles     []GroupRoleBody `json:"roles"`
	HideRoles bool            `json:"hide_roles"`
}

// GroupRoleResponseBody wraps a single GroupRoleBody.
type GroupRoleResponseBody struct {
	Role GroupRoleBody `json:"role"`
}

// CreateGroupRoleBody is POST /api/v1/groups/{id}/roles.
type CreateGroupRoleBody struct {
	Name string `json:"name"`
}

// UpdateGroupRoleBody is PATCH /api/v1/groups/{id}/roles/{roleId}.
type UpdateGroupRoleBody struct {
	Name        *string  `json:"name"`
	Color       *string  `json:"color"`
	Hoist       *bool    `json:"hoist"`
	Mentionable *bool    `json:"mentionable"`
	Permissions []string `json:"permissions"`
}

// AddGroupRoleMembersBody is POST /api/v1/groups/{id}/roles/{roleId}/members.
type AddGroupRoleMembersBody struct {
	UserIDs []string `json:"user_ids"`
}

// GroupRoleMembersBody is GET /api/v1/groups/{id}/roles/{roleId}/members.
type GroupRoleMembersBody struct {
	Members []GroupMemberBody `json:"members"`
}

// UpdateGroupHideRolesBody is PATCH /api/v1/groups/{id}/roles/settings.
type UpdateGroupHideRolesBody struct {
	HideRoles bool `json:"hide_roles"`
}

// GroupRoleFromPB maps a userv1.GroupRole to JSON.
func GroupRoleFromPB(r *userv1.GroupRole) GroupRoleBody {
	if r == nil {
		return GroupRoleBody{Permissions: []string{}}
	}

	return GroupRoleBody{
		ID:          r.GetId(),
		Slug:        r.GetSlug(),
		Name:        r.GetName(),
		Color:       r.GetColor(),
		Hoist:       r.GetHoist(),
		Mentionable: r.GetMentionable(),
		IsDefault:   r.GetIsDefault(),
		IsSystem:    r.GetIsSystem(),
		Position:    r.GetPosition(),
		Permissions: nonNilStrings(r.GetPermissions()),
		MemberCount: r.GetMemberCount(),
		CreatedAt:   r.GetCreatedAt(),
		UpdatedAt:   r.GetUpdatedAt(),
	}
}

// GroupRolesFromPB maps role protos to JSON bodies.
func GroupRolesFromPB(roles []*userv1.GroupRole) []GroupRoleBody {
	out := make([]GroupRoleBody, len(roles))
	for i, r := range roles {
		out[i] = GroupRoleFromPB(r)
	}

	return out
}

func nonNilStrings(in []string) []string {
	if in == nil {
		return []string{}
	}
	return in
}

// ChannelFromPB maps a channel summary proto to JSON.
func ChannelFromPB(c *userv1.ChannelSummary) ChannelBody {
	if c == nil {
		return ChannelBody{}
	}

	callerPerms := c.GetCallerPermissions()
	if callerPerms == nil {
		callerPerms = []string{}
	}

	return ChannelBody{
		ID:                    c.GetId(),
		Name:                  c.GetName(),
		Type:                  c.GetType(),
		SidebarSection:        c.GetSidebarSection(),
		ImageURL:              c.GetImageUrl(),
		IsPrivate:             c.GetIsPrivate(),
		CategoryID:            c.GetCategoryId(),
		Position:              int(c.GetPosition()),
		SlowmodeSeconds:       int(c.GetSlowmodeSeconds()),
		ThreadSlowmodeSeconds: int(c.GetThreadSlowmodeSeconds()),
		ContentVisibility:     c.GetContentVisibility(),
		ForumDefaultLayout:    c.GetForumDefaultLayout(),
		ForumSortOrder:        c.GetForumSortOrder(),
		ForumTagMatching:      c.GetForumTagMatching(),
		BitrateKbps:           int(c.GetBitrateKbps()),
		VideoQuality:          c.GetVideoQuality(),
		UserLimit:             int(c.GetUserLimit()),
		RTCRegion:             c.GetRtcRegion(),
		CallerPermissions:     callerPerms,
		DmPeerUserID:          c.GetDmPeerUserId(),
	}
}

// ChannelCategoryFromPB maps a channel category proto to JSON.
func ChannelCategoryFromPB(c *userv1.ChannelCategory) ChannelCategoryBody {
	if c == nil {
		return ChannelCategoryBody{}
	}

	return ChannelCategoryBody{
		ID:        c.GetId(),
		Name:      c.GetName(),
		IsPrivate: c.GetIsPrivate(),
		Position:  int(c.GetPosition()),
	}
}

// ChannelCategoriesFromPB maps channel category protos to JSON bodies.
func ChannelCategoriesFromPB(categories []*userv1.ChannelCategory) []ChannelCategoryBody {
	out := make([]ChannelCategoryBody, len(categories))
	for i, c := range categories {
		out[i] = ChannelCategoryFromPB(c)
	}

	return out
}

// ChannelsFromPB maps channel summary protos to JSON bodies.
func ChannelsFromPB(channels []*userv1.ChannelSummary) []ChannelBody {
	out := make([]ChannelBody, len(channels))
	for i, c := range channels {
		out[i] = ChannelFromPB(c)
	}

	return out
}

// GroupsFromPB maps group summary protos to JSON bodies.
func GroupsFromPB(groups []*userv1.GroupSummary) []GroupBody {
	out := make([]GroupBody, len(groups))
	for i, g := range groups {
		out[i] = GroupFromPB(g)
	}

	return out
}

// GroupMembersFromPB maps member entries to JSON.
func GroupMembersFromPB(members []*userv1.GroupMemberEntry) []GroupMemberBody {
	out := make([]GroupMemberBody, len(members))
	for i, m := range members {
		if m == nil {
			continue
		}
		pbRoles := m.GetRoles()
		roles := make([]GroupMemberRoleSummaryBody, 0, len(pbRoles))
		for _, r := range pbRoles {
			if r == nil {
				continue
			}
			roles = append(roles, GroupMemberRoleSummaryBody{
				ID:       r.GetId(),
				Slug:     r.GetSlug(),
				Name:     r.GetName(),
				Color:    r.GetColor(),
				Hoist:    r.GetHoist(),
				Position: r.GetPosition(),
			})
		}
		presenceStatus := m.GetPresenceStatus()
		if presenceStatus == "" {
			if m.GetIsOnline() {
				presenceStatus = "online"
			} else {
				presenceStatus = "offline"
			}
		}
		out[i] = GroupMemberBody{
			UserID:               m.GetUserId(),
			DisplayName:          m.GetDisplayName(),
			Username:             m.GetUsername(),
			AvatarURL:            m.GetAvatarUrl(),
			Role:                 m.GetRole(),
			JoinedAt:             m.GetJoinedAt(),
			Roles:                roles,
			IsOnline:             m.GetIsOnline(),
			PresenceStatus:       presenceStatus,
			BannerColor:          m.GetBannerColor(),
			Bio:                  m.GetBio(),
			Pronouns:             m.GetPronouns(),
			UserCreatedAt:        m.GetUserCreatedAt(),
			StatusNoteText:       m.GetStatusNoteText(),
			StatusNoteEmoji:      m.GetStatusNoteEmoji(),
			StatusNoteExpiresAt:  m.GetStatusNoteExpiresAt(),
			JoinMethod:           m.GetJoinMethod(),
			InviteCode:           m.GetInviteCode(),
			InvitedByUserID:      m.GetInvitedByUserId(),
			InvitedByUsername:    m.GetInvitedByUsername(),
			InvitedByDisplayName: m.GetInvitedByDisplayName(),
			TimeoutUntil:         m.GetTimeoutUntil(),
		}
	}

	return out
}

// UserSearchFromPB maps search hits to JSON.
func UserSearchFromPB(users []*userv1.UserSearchHit) []UserSearchHitBody {
	out := make([]UserSearchHitBody, len(users))
	for i, u := range users {
		if u == nil {
			continue
		}
		out[i] = UserSearchHitBody{
			UserID:      u.GetUserId(),
			DisplayName: u.GetDisplayName(),
			Username:    u.GetUsername(),
			AvatarURL:   u.GetAvatarUrl(),
		}
	}

	return out
}

// UserFromPB maps auth user proto to JSON body.
func UserFromPB(user *userv1.User) UserBody {
	var expiresAt *string
	if raw := user.GetStatusNoteExpiresAt(); raw != "" {
		expiresAt = &raw
	}

	return UserBody{
		ID:                  user.GetId(),
		Email:               user.GetEmail(),
		DisplayName:         user.GetDisplayName(),
		Username:            user.GetUsername(),
		AvatarURL:           user.GetAvatarUrl(),
		Pronouns:            user.GetPronouns(),
		Bio:                 user.GetBio(),
		BannerColor:         user.GetBannerColor(),
		StatusNoteText:      user.GetStatusNoteText(),
		StatusNoteEmoji:     user.GetStatusNoteEmoji(),
		StatusNoteExpiresAt: expiresAt,
		Locale:              user.GetLocale(),
		TimeFormat:          user.GetTimeFormat(),
		CreatedAt:           user.GetCreatedAt(),
		HasPassword:         user.GetHasPassword(),
	}
}

// ChallengeFromPB maps an auth challenge proto to JSON.
func ChallengeFromPB(ch *userv1.AuthChallenge) AuthChallengeBody {
	if ch == nil {
		return AuthChallengeBody{}
	}

	var waOpts json.RawMessage
	if raw := strings.TrimSpace(ch.GetWebauthnOptionsJson()); raw != "" {
		waOpts = json.RawMessage(raw)
	}

	return AuthChallengeBody{
		OTPRequired:        ch.GetOtpRequired(),
		MFARequired:        ch.GetMfaRequired(),
		MFAMethods:         ch.GetMfaMethods(),
		WebAuthnOptions:    waOpts,
		ChallengeID:        ch.GetChallengeId(),
		ResendAfterSeconds: ch.GetResendCooldownSeconds(),
	}
}

// SecurityKeyFromPB maps a SecurityKey proto to JSON body.
func SecurityKeyFromPB(k *userv1.SecurityKey) SecurityKeyBody {
	if k == nil {
		return SecurityKeyBody{}
	}
	transports := k.GetTransports()
	if transports == nil {
		transports = []string{}
	}
	return SecurityKeyBody{
		ID:         k.GetId(),
		Name:       k.GetName(),
		CreatedAt:  k.GetCreatedAt(),
		LastUsedAt: k.GetLastUsedAt(),
		Transports: transports,
	}
}

// WriteChallenge returns the email OTP or MFA step without setting a session cookie.
func WriteChallenge(w http.ResponseWriter, ch *userv1.AuthChallenge) {
	body := ChallengeFromPB(ch)
	if body.ChallengeID == "" {
		responsehttp.WriteError(w, nil, exception.Internal(nil))
		return
	}
	if !body.MFARequired {
		body.OTPRequired = true
	}
	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(body))
}

// WritePasswordAuth writes either a challenge or a session from register/login RPCs.
func WritePasswordAuth(w http.ResponseWriter, sess *userv1.Session, ch *userv1.AuthChallenge, secure bool) {
	if ch != nil && ch.GetChallengeId() != "" {
		WriteChallenge(w, ch)
		return
	}

	WriteSession(w, sess, secure)
}

// WriteSession sets the refresh cookie and returns access token + user in JSON.
func WriteSession(w http.ResponseWriter, sess *userv1.Session, secure bool) {
	if sess == nil || sess.GetAccessToken() == "" || sess.GetRefreshToken() == "" {
		responsehttp.WriteError(w, nil, exception.Internal(nil))
		return
	}

	cookie.SetRefresh(w, sess.GetRefreshToken(), secure)
	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(SessionBody{
		AccessToken: sess.GetAccessToken(),
		ExpiresIn:   sess.GetExpiresIn(),
		User:        UserFromPB(sess.GetUser()),
	}))
}

// WriteOutbound maps downstream gRPC/transport errors to JSON (optional app log on 5xx).
func WriteOutbound(w http.ResponseWriter, r *http.Request, log logger.Logger, err error) {
	var opts []responsehttp.WriteOption
	if log != nil {
		opts = append(opts, responsehttp.WithLogger(logger.HTTPWriteErrorLogger(log)))
	}
	responsehttp.WriteOutboundError(w, r, err, opts...)
}

// GRPCMessageCode extracts an i18n error code from a gRPC status for OAuth redirects.
func GRPCMessageCode(err error) string {
	parsed, ok := responsegrpc.FromError(err)
	if !ok {
		return string(exception.MsgInternal)
	}

	return string(parsed.Code)
}

// OAuthSignupReturnBlocked reports whether Google OAuth must not start a new signup when invite-only is on.
func OAuthSignupReturnBlocked(returnPath string) bool {
	path := SafeReturnPath(returnPath)
	if strings.HasPrefix(path, "/invite/") {
		return false
	}
	return path == "/sign-up"
}

// SafeReturnPath allows only same-app absolute paths (no open redirect).
func SafeReturnPath(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "/"
	}
	if !strings.HasPrefix(raw, "/") || strings.HasPrefix(raw, "//") {
		return "/"
	}

	return raw
}

// RedirectSignIn sends the browser to the web sign-in page with an error query param.
func RedirectSignIn(w http.ResponseWriter, r *http.Request, webOrigin, code string) {
	target := strings.TrimRight(strings.TrimSpace(webOrigin), "/") + "/sign-in?error=" + url.QueryEscape(code)
	http.Redirect(w, r, target, http.StatusFound)
}

// RedirectForbidden sends the browser to the registration-closed status page.
func RedirectForbidden(w http.ResponseWriter, r *http.Request, webOrigin string) {
	target := strings.TrimRight(strings.TrimSpace(webOrigin), "/") + "/forbidden"
	http.Redirect(w, r, target, http.StatusFound)
}

// UserActivityLogBody is the JSON representation of a universal user audit event.
type UserActivityLogBody struct {
	ID         string          `json:"id"`
	UserID     string          `json:"user_id"`
	ActorID    string          `json:"actor_id"`
	Category   string          `json:"category"`
	Action     string          `json:"action"`
	TargetType string          `json:"target_type"`
	TargetID   string          `json:"target_id"`
	Summary    string          `json:"summary"`
	Changes    json.RawMessage `json:"changes"`
	Metadata   json.RawMessage `json:"metadata"`
	CreatedAt  string          `json:"created_at"`
}

// UserActivityLogsFromPB maps proto UserActivityLog items to JSON bodies.
func UserActivityLogsFromPB(logs []*userv1.UserActivityLog) []UserActivityLogBody {
	out := make([]UserActivityLogBody, 0, len(logs))
	for _, entry := range logs {
		if entry == nil {
			continue
		}
		changes := json.RawMessage("{}")
		if raw := strings.TrimSpace(entry.GetChangesJson()); raw != "" {
			changes = json.RawMessage(raw)
		}
		meta := json.RawMessage("{}")
		if raw := strings.TrimSpace(entry.GetMetadataJson()); raw != "" {
			meta = json.RawMessage(raw)
		}
		out = append(out, UserActivityLogBody{
			ID:         entry.GetId(),
			UserID:     entry.GetUserId(),
			ActorID:    entry.GetActorId(),
			Category:   entry.GetCategory(),
			Action:     entry.GetAction(),
			TargetType: entry.GetTargetType(),
			TargetID:   entry.GetTargetId(),
			Summary:    entry.GetSummary(),
			Changes:    changes,
			Metadata:   meta,
			CreatedAt:  entry.GetCreatedAt(),
		})
	}
	return out
}

// GroupActivityLogBody is the JSON representation of a group audit log event.
type GroupActivityLogBody struct {
	ID               string          `json:"id"`
	GroupID          string          `json:"group_id"`
	ActorID          string          `json:"actor_id"`
	ActorUsername    string          `json:"actor_username"`
	ActorDisplayName string          `json:"actor_display_name"`
	ActorAvatarURL   string          `json:"actor_avatar_url"`
	Category         string          `json:"category"`
	Action           string          `json:"action"`
	TargetType       string          `json:"target_type"`
	TargetID         string          `json:"target_id"`
	TargetName       string          `json:"target_name"`
	Summary          string          `json:"summary"`
	Changes          json.RawMessage `json:"changes"`
	Metadata         json.RawMessage `json:"metadata"`
	CreatedAt        string          `json:"created_at"`
}

// GroupActivityLogsFromPB maps proto GroupActivityLog items to JSON bodies.
func GroupActivityLogsFromPB(logs []*userv1.GroupActivityLog) []GroupActivityLogBody {
	out := make([]GroupActivityLogBody, 0, len(logs))
	for _, entry := range logs {
		if entry == nil {
			continue
		}
		changes := json.RawMessage("{}")
		if raw := strings.TrimSpace(entry.GetChangesJson()); raw != "" {
			changes = json.RawMessage(raw)
		}
		meta := json.RawMessage("{}")
		if raw := strings.TrimSpace(entry.GetMetadataJson()); raw != "" {
			meta = json.RawMessage(raw)
		}
		out = append(out, GroupActivityLogBody{
			ID:               entry.GetId(),
			GroupID:          entry.GetGroupId(),
			ActorID:          entry.GetActorId(),
			ActorUsername:    entry.GetActorUsername(),
			ActorDisplayName: entry.GetActorDisplayName(),
			ActorAvatarURL:   entry.GetActorAvatarUrl(),
			Category:         entry.GetCategory(),
			Action:           entry.GetAction(),
			TargetType:       entry.GetTargetType(),
			TargetID:         entry.GetTargetId(),
			TargetName:       entry.GetTargetName(),
			Summary:          entry.GetSummary(),
			Changes:          changes,
			Metadata:         meta,
			CreatedAt:        entry.GetCreatedAt(),
		})
	}
	return out
}

// UserRelationshipBody is one ignore/block relationship entry.
type UserRelationshipBody struct {
	TargetUserID string `json:"target_user_id"`
	Kind         string `json:"kind"`
	CreatedAt    string `json:"created_at"`
}

// SetUserRelationshipBody is PUT /api/v1/users/me/relationships/{userId}.
type SetUserRelationshipBody struct {
	Kind string `json:"kind"`
}

// SetUserRelationshipResponseBody is the response for PUT /api/v1/users/me/relationships/{userId}.
type SetUserRelationshipResponseBody struct {
	OK   bool   `json:"ok"`
	Kind string `json:"kind"`
}
