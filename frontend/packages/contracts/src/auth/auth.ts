import { z } from "zod";
import { displayNameSchema, usernameSchema } from "../common";
import { UserSchema } from "../user/user";
import { MfaMethodSchema } from "./mfa";

export const SessionSchema = z.object({
	access_token: z.string(),
	expires_in: z.number(),
	user: UserSchema,
});
export type Session = z.infer<typeof SessionSchema>;

export const AuthChallengeSchema = z.object({
	otp_required: z.boolean().optional().default(false),
	mfa_required: z.boolean().optional().default(false),
	mfa_methods: z.array(MfaMethodSchema).optional().default([]),
	webauthn_options: z.unknown().optional(),
	challenge_id: z.string(),
	email_hint: z.string().optional().default(""),
	resend_after_seconds: z.number().int().nonnegative().optional().default(0),
});
export type AuthChallenge = z.infer<typeof AuthChallengeSchema>;

export const VerifyEmailCodeSchema = z.object({
	challenge_id: z.string().trim().min(1),
	code: z.string().trim().min(1).max(16),
	turnstile_token: z.string().optional(),
});
export type VerifyEmailCode = z.infer<typeof VerifyEmailCodeSchema>;

export const ResendEmailCodeSchema = z.object({
	challenge_id: z.string().trim().min(1),
	turnstile_token: z.string().optional(),
});
export type ResendEmailCode = z.infer<typeof ResendEmailCodeSchema>;

export const ResendEmailCodeResponseSchema = z.object({
	resend_after_seconds: z.number().int().nonnegative(),
});
export type ResendEmailCodeResponse = z.infer<
	typeof ResendEmailCodeResponseSchema
>;

export const RegistrationModeSchema = z.enum([
	"invite_only",
	"approval_required",
	"open",
]);
export type RegistrationMode = z.infer<typeof RegistrationModeSchema>;

export const InstancePolicySchema = z.object({
	registration_mode: RegistrationModeSchema,
	invite_only: z.boolean().optional().default(false),
	turnstile_site_key: z.string().optional().default(""),
	turnstile_required: z.boolean().optional().default(false),
});
export type InstancePolicy = z.infer<typeof InstancePolicySchema>;

/** @deprecated Use InstancePolicySchema */
export const SetupStatusSchema = InstancePolicySchema;
/** @deprecated Use InstancePolicy */
export type SetupStatus = InstancePolicy;

export const CheckUsernameResponseSchema = z.object({
	available: z.boolean(),
	normalized: z.string(),
	reason: z.string().optional().default(""),
});
export type CheckUsernameResponse = z.infer<typeof CheckUsernameResponseSchema>;

export const LoginRequestSchema = z.object({
	login: z.string().trim().min(1).max(254),
	password: z.string().min(1).max(128),
	locale: z.string().trim().optional(),
	turnstile_token: z.string().optional(),
});
export type LoginRequest = z.infer<typeof LoginRequestSchema>;

const birthDateSchema = z.string().regex(/^\d{4}-\d{2}-\d{2}$/);

export const RegisterRequestSchema = z.object({
	email: z.email(),
	username: usernameSchema,
	password: z.string().min(12).max(128),
	display_name: displayNameSchema,
	birth_date: birthDateSchema,
	locale: z.string().trim().optional(),
	invite_token: z.string().trim().optional(),
	turnstile_token: z.string().optional(),
});
export type RegisterRequest = z.infer<typeof RegisterRequestSchema>;

/** @deprecated Use RegisterRequestSchema */
export const SetupRequestSchema = RegisterRequestSchema;
export type SetupRequest = RegisterRequest;

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
