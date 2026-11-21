import { z } from "zod";

export const ForgotPasswordRequestSchema = z.object({
	email: z.email(),
	locale: z.string().trim().optional(),
	turnstile_token: z.string().optional(),
});

export const ForgotPasswordRequestResponseSchema = z.object({
	ok: z.boolean(),
	resend_cooldown_seconds: z.number().int().nonnegative(),
});

export const ForgotPasswordVerifySchema = z.object({
	email: z.email(),
	code: z.string().trim().min(1).max(16),
	turnstile_token: z.string().optional(),
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
