import { z } from "zod";

export const MfaMethodSchema = z.enum(["webauthn", "totp", "recovery"]);
export type MfaMethod = z.infer<typeof MfaMethodSchema>;

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
	turnstile_token: z.string().optional(),
	challenge_id: z.string().trim().min(1),
	method: MfaMethodSchema,
	code: z.string().optional(),
	assertion: z.unknown().optional(),
});
export type VerifyLoginMfaRequest = z.infer<typeof VerifyLoginMfaRequestSchema>;
