import { z } from "zod";
import {
	displayNameSchema,
	groupNameSchema,
	usernameSchema,
} from "./display-name";

export {
	displayNameSchema,
	groupNameSchema,
	usernameSchema,
	unicodeRuneCount,
	MIN_DISPLAY_NAME_RUNES,
	MAX_DISPLAY_NAME_RUNES,
	MIN_GROUP_NAME_RUNES,
	MAX_GROUP_NAME_RUNES,
	MIN_USERNAME_LEN,
	MAX_USERNAME_LEN,
	RESERVED_USERNAMES,
} from "./display-name";

export const UserSchema = z.object({
	id: z.string(),
	email: z.string(),
	display_name: z.string(),
	username: z.string(),
	avatar_url: z.string().optional().default(""),
	pronouns: z.string().optional().default(""),
	bio: z.string().optional().default(""),
	banner_color: z.string().optional().default(""),
	status_note_text: z.string().optional().default(""),
	status_note_emoji: z.string().optional().default(""),
	status_note_expires_at: z.string().nullable().optional().default(null),
	locale: z.string().optional().default(""),
	time_format: z.enum(["auto", "12h", "24h"]).optional().default("auto"),
	has_password: z.boolean().default(true),
	created_at: z.string().optional().default(""),
});

export const UpdateUserProfileRequestSchema = z.object({
	display_name: z.string(),
	pronouns: z.string().optional().default(""),
	bio: z.string().optional().default(""),
	banner_color: z.string().optional().default(""),
	clear_avatar: z.boolean().optional().default(false),
});

export type UpdateUserProfileRequest = z.infer<
	typeof UpdateUserProfileRequestSchema
>;

export const UpdateUserStatusNoteRequestSchema = z.object({
	text: z.string().max(80).optional().default(""),
	emoji: z.string().max(16).optional().default(""),
	expires_at: z.string().nullable().optional().default(null),
});

export type UpdateUserStatusNoteRequest = z.infer<
	typeof UpdateUserStatusNoteRequestSchema
>;

export const UpdateUserPreferencesRequestSchema = z.object({
	locale: z.enum(["en", "id"]),
	time_format: z.enum(["auto", "12h", "24h"]),
});

export type UpdateUserPreferencesRequest = z.infer<
	typeof UpdateUserPreferencesRequestSchema
>;

export const UpdateUsernameRequestSchema = z.object({
	username: usernameSchema,
	current_password: z.string().min(1),
});

export type UpdateUsernameRequest = z.infer<typeof UpdateUsernameRequestSchema>;

export const RequestEmailChangeRequestSchema = z.object({
	locale: z.string().optional().default("en"),
});

export type RequestEmailChangeRequest = z.infer<
	typeof RequestEmailChangeRequestSchema
>;

export const VerifyEmailChangeRequestSchema = z.object({
	challenge_id: z.string().trim().min(1),
	code: z.string().trim().min(1).max(16),
});

export type VerifyEmailChangeRequest = z.infer<
	typeof VerifyEmailChangeRequestSchema
>;

export const VerifyEmailChangeResponseSchema = z.object({
	change_token: z.string().min(1),
});

export type VerifyEmailChangeResponse = z.infer<
	typeof VerifyEmailChangeResponseSchema
>;

export const CompleteEmailChangeRequestSchema = z.object({
	change_token: z.string().trim().min(1),
	new_email: z.email(),
	current_password: z.string().min(1),
	reason: z.string().trim().optional().default(""),
});

export type CompleteEmailChangeRequest = z.infer<
	typeof CompleteEmailChangeRequestSchema
>;

export const UpdatePasswordRequestSchema = z.object({
	current_password: z.string().optional().default(""),
	new_password: z.string().min(12).max(128),
});

export type UpdatePasswordRequest = z.infer<typeof UpdatePasswordRequestSchema>;

export const AccountActionRequestSchema = z.object({
	current_password: z.string().optional().default(""),
});

export type AccountActionRequest = z.infer<typeof AccountActionRequestSchema>;

export const SessionSchema = z.object({
	access_token: z.string(),
	expires_in: z.number(),
	user: UserSchema,
});

export const MfaMethodSchema = z.enum(["webauthn", "totp", "recovery"]);
export type MfaMethod = z.infer<typeof MfaMethodSchema>;

export const AuthChallengeSchema = z.object({
	otp_required: z.boolean().optional().default(false),
	mfa_required: z.boolean().optional().default(false),
	mfa_methods: z.array(MfaMethodSchema).optional().default([]),
	webauthn_options: z.unknown().optional(),
	challenge_id: z.string(),
	email_hint: z.string().optional().default(""),
	resend_after_seconds: z.number().int().nonnegative().optional().default(0),
});

export const SecurityKeySchema = z.object({
	id: z.string(),
	name: z.string(),
	created_at: z.string(),
	last_used_at: z.string().optional(),
	transports: z.array(z.string()).default([]),
});
export type SecurityKey = z.infer<typeof SecurityKeySchema>;

export const MfaStatusSchema = z.object({
	totp_enabled: z.boolean(),
	security_keys: z.array(SecurityKeySchema).default([]),
	recovery_codes_remaining: z.number().int().nonnegative(),
});
export type MfaStatus = z.infer<typeof MfaStatusSchema>;

export const BeginTotpSetupResponseSchema = z.object({
	setup_token: z.string(),
	secret: z.string(),
	otpauth_uri: z.string(),
});
export type BeginTotpSetupResponse = z.infer<
	typeof BeginTotpSetupResponseSchema
>;

export const RecoveryCodesResponseSchema = z.object({
	recovery_codes: z.array(z.string()),
});
export type RecoveryCodesResponse = z.infer<typeof RecoveryCodesResponseSchema>;

export const BeginWebAuthnRegisterResponseSchema = z.object({
	challenge_id: z.string(),
	options: z.unknown(),
});
export type BeginWebAuthnRegisterResponse = z.infer<
	typeof BeginWebAuthnRegisterResponseSchema
>;

export const FinishWebAuthnRegisterResponseSchema = z.object({
	security_key: SecurityKeySchema,
	recovery_codes: z.array(z.string()).optional(),
});
export type FinishWebAuthnRegisterResponse = z.infer<
	typeof FinishWebAuthnRegisterResponseSchema
>;

export const BeginLoginWebAuthnResponseSchema = z.object({
	webauthn_options: z.unknown(),
});
export type BeginLoginWebAuthnResponse = z.infer<
	typeof BeginLoginWebAuthnResponseSchema
>;

export const VerifyLoginMfaRequestSchema = z.object({
	challenge_id: z.string().trim().min(1),
	method: MfaMethodSchema,
	code: z.string().optional(),
	assertion: z.unknown().optional(),
});
export type VerifyLoginMfaRequest = z.infer<typeof VerifyLoginMfaRequestSchema>;

export const VerifyEmailCodeSchema = z.object({
	challenge_id: z.string().trim().min(1),
	code: z.string().trim().min(1).max(16),
});

export const ResendEmailCodeSchema = z.object({
	challenge_id: z.string().trim().min(1),
});

export const ResendEmailCodeResponseSchema = z.object({
	resend_after_seconds: z.number().int().nonnegative(),
});

export const RegistrationModeSchema = z.enum([
	"invite_only",
	"approval_required",
	"open",
]);

export const InstancePolicySchema = z.object({
	registration_mode: RegistrationModeSchema,
	turnstile_site_key: z.string().optional().default(""),
});

export const CheckUsernameResponseSchema = z.object({
	available: z.boolean(),
	normalized: z.string(),
	reason: z.string().optional().default(""),
});
export type CheckUsernameResponse = z.infer<typeof CheckUsernameResponseSchema>;

/** @deprecated Use InstancePolicySchema */
export const SetupStatusSchema = InstancePolicySchema;

export const OkSchema = z.object({
	ok: z.boolean(),
});

/** Client-facing entity id (24 lowercase hex), used in REST and web URLs. */
export const publicEntityIdSchema = z.string().regex(/^[0-9a-f]{24}$/);

export const channelSidebarSectionSchema = z.enum(["favorites", "channels"]);

export const channelKindSchema = z.enum([
	"broadcast",
	"text_only",
	"voice_only",
	"text_and_voice",
	"forum",
	"dm",
	"public",
	"private",
]);

export const ChannelSchema = z.object({
	id: publicEntityIdSchema,
	name: z.string(),
	type: z.union([channelKindSchema, z.string()]),
	sidebar_section: channelSidebarSectionSchema.default("channels"),
	image_url: z.string().default(""),
	is_private: z.boolean().default(false),
});

export const ChannelsListSchema = z.object({
	channels: z.array(ChannelSchema),
});

export const MessageSchema = z.object({
	id: publicEntityIdSchema,
	author_id: z.string(),
	author_display_name: z.string(),
	body: z.string(),
	created_at: z.string(),
	mine: z.boolean().default(false),
});

export const MessagesListSchema = z.object({
	messages: z.array(MessageSchema),
});

export const CreateMessageBodySchema = z.object({
	body: z.string().trim().min(1).max(4096),
});

export const CreateMessageResponseSchema = z.object({
	message: MessageSchema,
});

export const RealtimeMessageCreatedSchema = z.object({
	type: z.literal("message.created"),
	message: MessageSchema,
});

export const PresenceStatusSchema = z.enum([
	"online",
	"idle",
	"dnd",
	"offline",
]);
export type PresenceStatus = z.infer<typeof PresenceStatusSchema>;

export const ManualPresenceStatusSchema = z.enum([
	"online",
	"idle",
	"dnd",
	"invisible",
]);
export type ManualPresenceStatus = z.infer<typeof ManualPresenceStatusSchema>;

export const MyPresenceResponseSchema = z.object({
	manual_status: ManualPresenceStatusSchema,
	effective_status: PresenceStatusSchema,
	is_online: z.boolean(),
	expires_at: z.string().optional().default(""),
});
export type MyPresenceResponse = z.infer<typeof MyPresenceResponseSchema>;

export const UpdateMyPresenceRequestSchema = z.object({
	status: ManualPresenceStatusSchema,
	ttl_seconds: z.number().int().min(0).optional(),
});
export type UpdateMyPresenceRequest = z.infer<
	typeof UpdateMyPresenceRequestSchema
>;

export const RealtimePresenceUpdatedSchema = z.object({
	type: z.literal("presence.updated"),
	user_id: z.string(),
	presence_status: PresenceStatusSchema,
	is_online: z.boolean(),
});
export type RealtimePresenceUpdated = z.infer<
	typeof RealtimePresenceUpdatedSchema
>;

export const groupMembershipRoleSchema = z.enum(["owner", "member"]);

export const GroupSchema = z.object({
	id: publicEntityIdSchema,
	name: z.string(),
	logo_url: z.string(),
	description: z.string().default(""),
	is_private: z.boolean().default(false),
	banner_color: z.string().default(""),
	banner_url: z.string().default(""),
	hide_roles: z.boolean().default(false),
	default_channel_id: z.union([publicEntityIdSchema, z.literal("")]),
	role: z.preprocess(
		(v) => (v === "" || v == null ? undefined : v),
		groupMembershipRoleSchema.or(z.enum(["admin", "guest"])).optional(),
	),
	effective_permissions: z.array(z.string()).default([]),
});

export const GroupsListSchema = z.object({
	groups: z.array(GroupSchema),
});

/** Groups the user belongs to — membership role is always set. */
export const MyGroupSchema = GroupSchema.extend({
	role: groupMembershipRoleSchema,
});

export const MyGroupsListSchema = z.object({
	groups: z.array(MyGroupSchema),
});

export const ReorderMyGroupsRequestSchema = z.object({
	group_ids: z.array(publicEntityIdSchema).min(1),
});

export const CreateGroupRequestSchema = z.object({
	name: groupNameSchema,
});

export const UpdateGroupProfileBodySchema = z.object({
	name: groupNameSchema,
	description: z.string().trim().max(2000),
	is_private: z.boolean(),
	banner_color: z.string(),
	clear_logo: z.boolean(),
	clear_banner_image: z.boolean(),
});

export const ForgotPasswordRequestResponseSchema = z.object({
	ok: z.boolean(),
	resend_cooldown_seconds: z.number().int().nonnegative(),
});

export const LoginRequestSchema = z.object({
	login: z.string().trim().min(1).max(254),
	password: z.string().min(1).max(128),
	locale: z.string().trim().optional(),
	turnstile_token: z.string().optional(),
});

const birthDateSchema = z.string().regex(/^\d{4}-\d{2}-\d{2}$/);

export const RegisterRequestSchema = z.object({
	email: z.email(),
	username: usernameSchema,
	password: z.string().min(12).max(128),
	display_name: displayNameSchema,
	birth_date: birthDateSchema,
	locale: z.string().trim().optional(),
	turnstile_token: z.string().optional(),
});

/** @deprecated Use RegisterRequestSchema */
export const SetupRequestSchema = RegisterRequestSchema;

export type User = z.infer<typeof UserSchema>;
export type Session = z.infer<typeof SessionSchema>;
export type AuthChallenge = z.infer<typeof AuthChallengeSchema>;
export type VerifyEmailCode = z.infer<typeof VerifyEmailCodeSchema>;
export type ResendEmailCode = z.infer<typeof ResendEmailCodeSchema>;
export type ResendEmailCodeResponse = z.infer<
	typeof ResendEmailCodeResponseSchema
>;

export type PasswordAuthStep =
	| { kind: "session"; session: Session }
	| { kind: "challenge"; challenge: AuthChallenge }
	| { kind: "mfa_challenge"; challenge: AuthChallenge };

export function parsePasswordAuthStep(data: unknown): PasswordAuthStep {
	const record =
		data && typeof data === "object" ? (data as Record<string, unknown>) : {};
	if (
		typeof record.challenge_id === "string" &&
		record.challenge_id.length > 0
	) {
		const challenge = AuthChallengeSchema.parse(data);
		if (challenge.mfa_required) {
			return { kind: "mfa_challenge", challenge };
		}
		return { kind: "challenge", challenge };
	}

	return { kind: "session", session: SessionSchema.parse(data) };
}
export type InstancePolicy = z.infer<typeof InstancePolicySchema>;
/** @deprecated Use InstancePolicy */
export type SetupStatus = InstancePolicy;
export type RegistrationMode = z.infer<typeof RegistrationModeSchema>;
export type LoginRequest = z.infer<typeof LoginRequestSchema>;
export type RegisterRequest = z.infer<typeof RegisterRequestSchema>;
export type SetupRequest = RegisterRequest;
export type Channel = z.infer<typeof ChannelSchema>;
export type ChannelsList = z.infer<typeof ChannelsListSchema>;
export type Message = z.infer<typeof MessageSchema>;
export type MessagesList = z.infer<typeof MessagesListSchema>;
export type CreateMessageBody = z.infer<typeof CreateMessageBodySchema>;
export type CreateMessageResponse = z.infer<typeof CreateMessageResponseSchema>;
export type Group = z.infer<typeof GroupSchema>;
export type GroupsList = z.infer<typeof GroupsListSchema>;

export function isPublicEntityId(value: string): boolean {
	return publicEntityIdSchema.safeParse(value).success;
}
export type CreateGroupRequest = z.infer<typeof CreateGroupRequestSchema>;
export type UpdateGroupProfileBody = z.infer<
	typeof UpdateGroupProfileBodySchema
>;

export const ForgotPasswordRequestSchema = z.object({
	email: z.email(),
	locale: z.string().trim().optional(),
});

export const ForgotPasswordVerifySchema = z.object({
	email: z.email(),
	code: z.string().trim().min(1).max(16),
});

export const ForgotPasswordVerifyResponseSchema = z.object({
	reset_token: z.string(),
});

export const ForgotPasswordResetSchema = z.object({
	reset_token: z.string().trim().min(1),
	password: z.string().min(12).max(128),
});

export type ForgotPasswordRequest = z.infer<typeof ForgotPasswordRequestSchema>;
export type ForgotPasswordRequestResponse = z.infer<
	typeof ForgotPasswordRequestResponseSchema
>;
export type ForgotPasswordVerify = z.infer<typeof ForgotPasswordVerifySchema>;
export type ForgotPasswordReset = z.infer<typeof ForgotPasswordResetSchema>;

export const GroupMemberRoleSummarySchema = z.object({
	id: z.string(),
	slug: z.string(),
	name: z.string(),
	color: z.string(),
	hoist: z.boolean().optional().default(false),
	position: z.number().int().optional().default(0),
});

export type GroupMemberRoleSummary = z.infer<
	typeof GroupMemberRoleSummarySchema
>;

export const GroupMemberSchema = z.object({
	user_id: z.string(),
	display_name: z.string(),
	username: z.string(),
	avatar_url: z.string().optional().default(""),
	role: groupMembershipRoleSchema,
	joined_at: z.string(),
	roles: z.array(GroupMemberRoleSummarySchema).optional().default([]),
	is_online: z.boolean().optional().default(false),
	presence_status: PresenceStatusSchema.optional().default("offline"),
});

export type GroupMember = z.infer<typeof GroupMemberSchema>;

export const GroupMembersPageSchema = z.object({
	members: z.array(GroupMemberSchema),
	next_cursor: z.string().optional().default(""),
	total: z.number().int().nonnegative(),
});
export type GroupMembersPage = z.infer<typeof GroupMembersPageSchema>;

export const UserSearchHitSchema = z.object({
	user_id: z.string(),
	display_name: z.string(),
	username: z.string(),
	avatar_url: z.string().optional().default(""),
});

export const UserSearchSchema = z.object({
	users: z.array(UserSearchHitSchema),
});

export const GroupInviteLinkSchema = z.object({
	token: z.string(),
	url: z.string(),
	expires_at: z.string(),
});

export const SendGroupInvitesRequestSchema = z.object({
	user_ids: z.array(z.string()).default([]),
	emails: z.array(z.email()).default([]),
	channel_id: z.string().optional(),
});

export const SendGroupInvitesResultSchema = z.object({
	added: z.number().int().nonnegative(),
	emailed: z.number().int().nonnegative(),
	skipped: z.number().int().nonnegative(),
});

export const InvitePreviewSchema = z
	.object({
		valid: z.boolean(),
		is_private: z.boolean().optional(),
		group_name: z.string().optional(),
		channel_name: z.string().optional(),
		group_logo_url: z.string().optional(),
		groupLogoUrl: z.string().optional(),
		member_count: z.number().int().nonnegative().optional(),
		online_count: z.number().int().nonnegative().optional(),
		expires_at: z.string().optional(),
		reason: z.string().optional(),
	})
	.transform(({ groupLogoUrl, group_logo_url, ...rest }) => ({
		...rest,
		group_logo_url: group_logo_url ?? groupLogoUrl,
	}));

export type InvitePreview = z.infer<typeof InvitePreviewSchema>;

export const RedeemInviteRequestSchema = z.object({
	token: z.string().trim().min(1),
});

export const RedeemInviteResultSchema = z.object({
	group_id: publicEntityIdSchema,
	channel_id: z.union([publicEntityIdSchema, z.literal("")]).optional(),
});

export const DeviceSessionSchema = z.object({
	id: z.string(),
	platform: z.string().optional().default("web"),
	device_label: z.string().optional().default(""),
	location_label: z.string().optional().default(""),
	user_agent: z.string().optional().default(""),
	last_used_at: z.string(),
	created_at: z.string(),
	is_current: z.boolean(),
});

export type DeviceSession = z.infer<typeof DeviceSessionSchema>;

export const DeviceSessionsResponseSchema = z.object({
	devices: z.array(DeviceSessionSchema),
});

export type DeviceSessionsResponse = z.infer<
	typeof DeviceSessionsResponseSchema
>;

export const RevokeDeviceRequestSchema = z.object({
	current_password: z.string().optional().default(""),
});

export type RevokeDeviceRequest = z.infer<typeof RevokeDeviceRequestSchema>;

export const ActivityFieldChangeSchema = z.object({
	old: z.unknown().optional(),
	new: z.unknown().optional(),
});

export type ActivityFieldChange = z.infer<typeof ActivityFieldChangeSchema>;

export const UserActivityLogSchema = z.object({
	id: z.string(),
	user_id: z.string(),
	actor_id: z.string(),
	category: z.string(),
	action: z.string(),
	target_type: z.string(),
	target_id: z.string(),
	summary: z.string(),
	changes: z.record(z.string(), ActivityFieldChangeSchema).default({}),
	metadata: z.record(z.string(), z.unknown()).default({}),
	created_at: z.string(),
});

export type UserActivityLog = z.infer<typeof UserActivityLogSchema>;

export const ListUserActivityLogsResponseSchema = z.object({
	logs: z.array(UserActivityLogSchema),
	next_cursor: z.string().optional(),
});

export type ListUserActivityLogsResponse = z.infer<
	typeof ListUserActivityLogsResponseSchema
>;

export const GroupRoleSchema = z.object({
	id: z.string(),
	slug: z.string(),
	name: z.string(),
	color: z.string().default("#99aab5"),
	hoist: z.boolean().default(false),
	mentionable: z.boolean().default(false),
	is_default: z.boolean().default(false),
	is_system: z.boolean().default(false),
	position: z.number().int().nonnegative().default(0),
	permissions: z.array(z.string()).default([]),
	member_count: z.number().int().nonnegative().default(0),
	created_at: z.string().optional().default(""),
	updated_at: z.string().optional().default(""),
});

export type GroupRole = z.infer<typeof GroupRoleSchema>;

export const GroupRolesListResponseSchema = z.object({
	roles: z.array(GroupRoleSchema),
	hide_roles: z.boolean().default(false),
});

export type GroupRolesListResponse = z.infer<
	typeof GroupRolesListResponseSchema
>;

export const GroupRoleResponseSchema = z.object({
	role: GroupRoleSchema,
});

export type GroupRoleResponse = z.infer<typeof GroupRoleResponseSchema>;

export const CreateGroupRoleRequestSchema = z.object({
	name: z.string().trim().max(48).optional().default(""),
});

export type CreateGroupRoleRequest = z.infer<
	typeof CreateGroupRoleRequestSchema
>;

export const UpdateGroupRoleRequestSchema = z.object({
	name: z.string().trim().min(1).max(48).optional(),
	color: z.string().trim().optional(),
	hoist: z.boolean().optional(),
	mentionable: z.boolean().optional(),
	permissions: z.array(z.string()).optional(),
});

export type UpdateGroupRoleRequest = z.infer<
	typeof UpdateGroupRoleRequestSchema
>;

export const UpdateGroupHideRolesRequestSchema = z.object({
	hide_roles: z.boolean(),
});

export type UpdateGroupHideRolesRequest = z.infer<
	typeof UpdateGroupHideRolesRequestSchema
>;

export const UpdateGroupHideRolesResponseSchema = z.object({
	hide_roles: z.boolean(),
});

export type UpdateGroupHideRolesResponse = z.infer<
	typeof UpdateGroupHideRolesResponseSchema
>;

export const GroupRoleMembersResponseSchema = z.object({
	members: z.array(GroupMemberSchema),
});

export type GroupRoleMembersResponse = z.infer<
	typeof GroupRoleMembersResponseSchema
>;

export const AddGroupRoleMembersRequestSchema = z.object({
	user_ids: z.array(z.string().min(1)).min(1),
});

export type AddGroupRoleMembersRequest = z.infer<
	typeof AddGroupRoleMembersRequestSchema
>;
