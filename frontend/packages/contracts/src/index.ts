import { z } from "zod";
import { displayNameSchema } from "./display-name";

export {
	displayNameSchema,
	unicodeRuneCount,
	MAX_DISPLAY_NAME_RUNES,
} from "./display-name";

export const UserSchema = z.object({
	id: z.string(),
	email: z.string(),
	display_name: z.string(),
	username: z.string().optional().default(""),
	is_guest: z.boolean().optional().default(false),
});

export const SessionSchema = z.object({
	access_token: z.string(),
	expires_in: z.number(),
	user: UserSchema,
});

export const AuthChallengeSchema = z.object({
	otp_required: z.boolean(),
	challenge_id: z.string(),
	email_hint: z.string(),
	resend_after_seconds: z.number().int().nonnegative(),
});

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
	guest_enabled: z.boolean(),
	username_enabled: z.boolean(),
});

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

export const groupMembershipRoleSchema = z.enum(["owner", "member"]);

export const GroupSchema = z.object({
	id: publicEntityIdSchema,
	name: z.string(),
	logo_url: z.string(),
	default_channel_id: z.union([publicEntityIdSchema, z.literal("")]),
	role: z.preprocess(
		(v) => (v === "" || v == null ? undefined : v),
		groupMembershipRoleSchema.or(z.enum(["admin", "guest"])).optional(),
	),
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
	name: displayNameSchema,
});

export const ForgotPasswordRequestResponseSchema = z.object({
	ok: z.boolean(),
	resend_cooldown_seconds: z.number().int().nonnegative(),
});

export const LoginRequestSchema = z.object({
	login: z.string().trim().min(1).max(254),
	password: z.string().min(12).max(128),
	locale: z.string().trim().optional(),
});

const birthDateSchema = z.string().regex(/^\d{4}-\d{2}-\d{2}$/);

export const RegisterRequestSchema = z.object({
	email: z.email(),
	password: z.string().min(12).max(128),
	display_name: displayNameSchema,
	birth_date: birthDateSchema,
	locale: z.string().trim().optional(),
});

export const GuestRegisterRequestSchema = z.object({
	display_name: displayNameSchema,
	birth_date: birthDateSchema,
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
	| { kind: "challenge"; challenge: AuthChallenge };

export function parsePasswordAuthStep(data: unknown): PasswordAuthStep {
	const record =
		data && typeof data === "object" ? (data as Record<string, unknown>) : {};
	if (
		typeof record.challenge_id === "string" &&
		record.challenge_id.length > 0
	) {
		return { kind: "challenge", challenge: AuthChallengeSchema.parse(data) };
	}

	return { kind: "session", session: SessionSchema.parse(data) };
}
export type InstancePolicy = z.infer<typeof InstancePolicySchema>;
/** @deprecated Use InstancePolicy */
export type SetupStatus = InstancePolicy;
export type RegistrationMode = z.infer<typeof RegistrationModeSchema>;
export type LoginRequest = z.infer<typeof LoginRequestSchema>;
export type RegisterRequest = z.infer<typeof RegisterRequestSchema>;
export type GuestRegisterRequest = z.infer<typeof GuestRegisterRequestSchema>;
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

export const GroupMemberSchema = z.object({
	user_id: z.string(),
	display_name: z.string(),
	email: z.string(),
	role: groupMembershipRoleSchema,
	joined_at: z.string(),
});

export const GroupMembersPageSchema = z.object({
	members: z.array(GroupMemberSchema),
	next_cursor: z.string().optional().default(""),
	total: z.number().int().nonnegative(),
});

export const UserSearchHitSchema = z.object({
	user_id: z.string(),
	display_name: z.string(),
	email: z.string(),
	username: z.string(),
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

export const InvitePreviewSchema = z.object({
	valid: z.boolean(),
	group_name: z.string().optional(),
	channel_name: z.string().optional(),
	expires_at: z.string().optional(),
	reason: z.string().optional(),
});

export const RedeemInviteRequestSchema = z.object({
	token: z.string().trim().min(1),
});

export const RedeemInviteResultSchema = z.object({
	group_id: publicEntityIdSchema,
	channel_id: z.union([publicEntityIdSchema, z.literal("")]).optional(),
});
