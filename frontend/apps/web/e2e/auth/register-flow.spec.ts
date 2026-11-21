import { expect, test } from "@playwright/test";
import { createTestUser } from "../fixtures/api-helpers";
import { writeLoginSuiteUser } from "../fixtures/login-suite-user";
import {
	e2eAuthSeedingAvailable,
	expectRegisterRejected,
	expectGoogleStartForbiddenForSignUp,
	expectRegisterUnavailable,
	expectRegisterVerifyRequiresCaptcha,
	expectVerifyRegisterRejected,
	fetchPolicyInviteOnly,
	fetchTurnstileRequired,
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
	const safePrefix = prefix.replace(/[^a-z0-9_]/gi, "_");
	const username = `e2e_${safePrefix}_${suffix}`.toLowerCase().slice(0, 30);
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
		test("sign-up page loads and policy request succeeds", async ({ page, request }) => {
			const inviteOnly = await fetchPolicyInviteOnly(request);
			const setupStatus = page.waitForResponse(
				(res) => res.url().includes("/auth/setup-status") && res.ok(),
			);
			await page.goto("/sign-up");
			await setupStatus;
			if (inviteOnly) {
				await expect(page.getByRole("heading", { name: /^403$/ })).toBeVisible();
				return;
			}
			await expect(
				page.getByRole("heading", { name: /create your account|buat akun|register|daftar/i }),
			).toBeVisible();
		});

		test("register returns email OTP challenge", async ({ request }) => {
			test.skip(!e2eAuthSeedingAvailable(), "Redis required for dev OTP shadow");
			test.skip(await fetchPolicyInviteOnly(request), "INVITE_ONLY enabled on gateway");
			const body = registerPayload("regchal");
			registerChallengeId = await registerPasswordChallenge(request, body);
			expect(registerChallengeId.length).toBeGreaterThan(0);
		});

		test("register rejected without invite when INVITE_ONLY", async ({ request }) => {
			test.skip(!(await fetchPolicyInviteOnly(request)), "INVITE_ONLY not enabled");
			await expectRegisterRejected(request, registerPayload("invclosed"));
		});

		test("google start redirects to forbidden when INVITE_ONLY and sign-up return", async ({
			request,
		}) => {
			test.skip(!(await fetchPolicyInviteOnly(request)), "INVITE_ONLY not enabled");
			await expectGoogleStartForbiddenForSignUp(request);
		});

		test("register verify requires captcha when turnstile enabled", async ({ request }) => {
			test.skip(!(await fetchTurnstileRequired(request)), "Turnstile not enabled on gateway");
			test.skip(await fetchPolicyInviteOnly(request), "INVITE_ONLY enabled on gateway");
			test.skip(!registerChallengeId, "needs register challenge from prior test");
			await expectRegisterVerifyRequiresCaptcha(request, registerChallengeId, "000000");
		});

		test("wrong OTP then same dev shadow code still verifies", async ({ request }) => {
			test.skip(!e2eAuthSeedingAvailable(), "Redis required for dev OTP shadow");
			test.skip(await fetchPolicyInviteOnly(request), "INVITE_ONLY enabled on gateway");
			test.skip(!process.env.E2E_FULL_AUTH, "set E2E_FULL_AUTH=1 to run OTP verify");
			const body = registerPayload("otp-retry");
			const ch = await registerPasswordChallenge(request, body);
			test.skip(!ch, "register rate limited on gateway");
			const code = waitDevOtpForChallenge(ch, 40);
			expect(code, "dev Redis OTP shadow missing").toBeTruthy();
			await expectVerifyRegisterRejected(request, ch, "000000");
			const session = await verifyRegisterCode(request, ch, code!);
			expect(session.access_token.length).toBeGreaterThan(10);
		});

		test("wrong register OTP code is rejected", async ({ request }) => {
			test.skip(await fetchPolicyInviteOnly(request), "INVITE_ONLY enabled on gateway");
			test.skip(!registerChallengeId, "needs register challenge from prior test");
			await expectVerifyRegisterRejected(request, registerChallengeId, "ZZZZZZ");
		});

		test("register verify succeeds when dev Redis OTP shadow is available", async ({ request }) => {
			test.skip(await fetchPolicyInviteOnly(request), "INVITE_ONLY enabled on gateway");
			test.skip(!process.env.E2E_FULL_AUTH, "set E2E_FULL_AUTH=1 to run full OTP verify");
			test.skip(!registerChallengeId, "needs register challenge from prior test");

			const code = waitDevOtpForChallenge(registerChallengeId);
			expect(code, "dev Redis OTP shadow missing — restart ms-user (development)").toBeTruthy();

			const session = await verifyRegisterCode(request, registerChallengeId, code!);
			expect(session.access_token.length).toBeGreaterThan(10);
			expect(session.user.id.length).toBeGreaterThan(0);
		});

		test("duplicate email register returns neutral error", async ({ request }) => {
			test.skip(!e2eAuthSeedingAvailable(), "Redis required for dev OTP shadow");
			test.skip(await fetchPolicyInviteOnly(request), "INVITE_ONLY enabled on gateway");
			test.skip(!process.env.E2E_FULL_AUTH, "set E2E_FULL_AUTH=1 for duplicate-email register");
			const body = registerPayload("neutral-dup");
			const ch = await registerPasswordChallenge(request, body);
			test.skip(!ch, "register rate limited on gateway");
			const code = waitDevOtpForChallenge(ch);
			expect(code, "dev Redis OTP shadow missing").toBeTruthy();
			await verifyRegisterCode(request, ch, code!);
			await expectRegisterUnavailable(request, body);
		});

		test("second register challenge for same email is rejected", async ({ request }) => {
			test.skip(!e2eAuthSeedingAvailable(), "Redis required for dev OTP shadow");
			test.skip(await fetchPolicyInviteOnly(request), "INVITE_ONLY enabled on gateway");
			const body = registerPayload("one-chal");
			const ch = await registerPasswordChallenge(request, body);
			test.skip(!ch, "register rate limited on gateway");
			expect(ch.length).toBeGreaterThan(0);
			await expectRegisterUnavailable(request, body);
		});

		test("seeds login suite user for login-flow e2e", async ({ request }) => {
			test.skip(await fetchPolicyInviteOnly(request), "INVITE_ONLY enabled on gateway");
			test.skip(!e2eAuthSeedingAvailable(), "Redis required for dev OTP shadow");
			const user = await createTestUser(request, "login_suite");
			writeLoginSuiteUser(user);
			expect(user.accessToken.length).toBeGreaterThan(10);
		});
	});
});
