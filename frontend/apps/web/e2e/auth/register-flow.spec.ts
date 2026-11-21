import { expect, test } from "@playwright/test";
import { createTestUser } from "../fixtures/api-helpers";
import { writeLoginSuiteUser } from "../fixtures/login-suite-user";
import {
	e2eAuthSeedingAvailable,
	expectVerifyRegisterRejected,
	registerPasswordChallenge,
	verifyRegisterCode,
	waitDevOtpForChallenge,
	type RegisterPayload,
} from "../fixtures/auth-helpers";

function uniqueSuffix(): string {
	return `${Date.now().toString(36)}_${Math.random().toString(36).slice(2, 7)}`;
}

function registerPayload(prefix: string): RegisterPayload {
	const suffix = uniqueSuffix();
	const username = `e2e_${prefix}_${suffix}`.toLowerCase().slice(0, 30);
	return {
		email: `${username}@example.test`,
		username,
		password: `E2ePass!_${suffix}`,
		display_name: `E2E ${prefix} ${suffix}`,
		birth_date: "1998-05-15",
		locale: "en",
	};
}

test.describe("auth register e2e", () => {
	test.describe.configure({ mode: "serial" });

	let registerChallengeId = "";

	test.describe("auth register flow", () => {
		test("sign-up page loads and policy request succeeds", async ({ page }) => {
			const setupStatus = page.waitForResponse(
				(res) => res.url().includes("/auth/setup-status") && res.ok(),
			);
			await page.goto("/sign-up");
			await setupStatus;
			await expect(
				page.getByRole("heading", { name: /create your account|buat akun|register|daftar/i }),
			).toBeVisible();
		});

		test("register returns email OTP challenge", async ({ request }) => {
			test.skip(!e2eAuthSeedingAvailable(), "Redis required for dev OTP shadow");
			const body = registerPayload("regchal");
			registerChallengeId = await registerPasswordChallenge(request, body);
			expect(registerChallengeId.length).toBeGreaterThan(0);
		});

		test("wrong register OTP code is rejected", async ({ request }) => {
			test.skip(!registerChallengeId, "needs register challenge from prior test");
			await expectVerifyRegisterRejected(request, registerChallengeId, "ZZZZZZ");
		});

		test("register verify succeeds when dev Redis OTP shadow is available", async ({
			request,
		}) => {
			test.skip(!process.env.E2E_FULL_AUTH, "set E2E_FULL_AUTH=1 to run full OTP verify");
			test.skip(!registerChallengeId, "needs register challenge from prior test");

			const code = waitDevOtpForChallenge(registerChallengeId);
			expect(code, "dev Redis OTP shadow missing — restart ms-user (development)").toBeTruthy();

			const session = await verifyRegisterCode(request, registerChallengeId, code!);
			expect(session.access_token.length).toBeGreaterThan(10);
			expect(session.user.id.length).toBeGreaterThan(0);
		});

		test("seeds login suite user for login-flow e2e", async ({ request }) => {
			test.skip(!e2eAuthSeedingAvailable(), "Redis required for dev OTP shadow");
			const user = await createTestUser(request, "login_suite");
			writeLoginSuiteUser(user);
			expect(user.accessToken.length).toBeGreaterThan(10);
		});
	});
});