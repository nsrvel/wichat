import { expect, test } from "@playwright/test";
import { readLoginSuiteUser } from "../fixtures/login-suite-user";
import {
	e2eAuthSeedingAvailable,
	expectForgotPasswordVerifyRequiresCaptcha,
	expectForgotPasswordWrongCode,
	fetchTurnstileRequired,
	forgotPasswordRequest,
	forgotPasswordReset,
	forgotPasswordVerify,
	waitDevOtpForPasswordReset,
} from "../fixtures/auth-helpers";

test.describe("auth forgot-password e2e", () => {
	test.describe.configure({ mode: "serial" });

	const suiteUser = readLoginSuiteUser();
	const resetEmail = suiteUser?.email ?? "";
	let resetToken = "";

	test.describe("forgot password flow", () => {
		test("forgot-password page loads and setup-status succeeds", async ({ page }) => {
			const setupStatus = page.waitForResponse(
				(res) => res.url().includes("/auth/setup-status") && res.ok(),
			);
			await page.goto("/forgot-password");
			await setupStatus;
			await expect(
				page.getByRole("heading", { name: /reset|forgot|lupa|password|kata sandi/i }),
			).toBeVisible();
		});

		test("unknown email still returns ok (anti-enumeration)", async ({ request }) => {
			const ghost = `ghost_reset_${Date.now()}@example.test`;
			const data = await forgotPasswordRequest(request, ghost);
			expect(data.resend_cooldown_seconds).toBeGreaterThanOrEqual(0);
		});

		test("request sends OTP for seeded user", async ({ request }) => {
			test.skip(!suiteUser, "run register-flow first to seed login-suite-user.json");
			test.skip(!e2eAuthSeedingAvailable(), "Redis required for dev OTP shadow");
			const data = await forgotPasswordRequest(request, resetEmail);
			expect(data.resend_cooldown_seconds).toBeGreaterThanOrEqual(0);
		});

		test("wrong reset code is rejected", async ({ request }) => {
			test.skip(!suiteUser, "needs seeded user");
			await expectForgotPasswordWrongCode(request, resetEmail, "ZZZZZZ");
		});

		test("verify requires captcha when turnstile enabled", async ({ request }) => {
			test.skip(!suiteUser, "needs seeded user");
			test.skip(!(await fetchTurnstileRequired(request)), "Turnstile not enabled on gateway");
			const code = waitDevOtpForPasswordReset(resetEmail) ?? "ABCDEF";
			await expectForgotPasswordVerifyRequiresCaptcha(request, resetEmail, code);
		});

		test("full reset when dev Redis OTP shadow is available", async ({ request }) => {
			test.skip(!process.env.E2E_FULL_AUTH, "set E2E_FULL_AUTH=1 to run full OTP verify");
			test.skip(!suiteUser, "needs seeded user");
			test.skip(!e2eAuthSeedingAvailable(), "Redis required for dev OTP shadow");
			test.skip(await fetchTurnstileRequired(request), "Turnstile blocks API-only verify in this spec");

			const code = waitDevOtpForPasswordReset(resetEmail);
			expect(code, "dev Redis OTP shadow missing — restart ms-user (development)").toBeTruthy();

			const verified = await forgotPasswordVerify(request, resetEmail, code!);
			expect(verified.reset_token.length).toBeGreaterThan(10);
			resetToken = verified.reset_token;

			const originalPass = suiteUser!.password;
			const newPass = `${originalPass}_reset_${Date.now().toString(36)}`.slice(0, 64);
			await forgotPasswordReset(request, resetToken, newPass);

			await forgotPasswordRequest(request, resetEmail);
			const restoreCode = waitDevOtpForPasswordReset(resetEmail);
			expect(restoreCode).toBeTruthy();
			const restore = await forgotPasswordVerify(request, resetEmail, restoreCode!);
			await forgotPasswordReset(request, restore.reset_token, originalPass);
		});
	});
});