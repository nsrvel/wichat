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

export const GroupSchema = z.object({
	id: z.string(),
	name: z.string(),
	logo_url: z.string(),
});

export const GroupsListSchema = z.object({
	groups: z.array(GroupSchema),
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
export type Group = z.infer<typeof GroupSchema>;
export type GroupsList = z.infer<typeof GroupsListSchema>;
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
