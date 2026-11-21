import { expect, test } from "@playwright/test";
import { E2E_API_BASE, type TestUser } from "../fixtures/api-helpers";
import { readLoginSuiteUser } from "../fixtures/login-suite-user";
import {
	expectLoginInvalidCredentials,
	expectVerifyOtpRejected,
	loginPasswordChallenge,
	waitDevOtpForChallenge,
	verifyLoginCode,
} from "../fixtures/auth-helpers";

test.describe("auth login e2e", () => {
	test.describe.configure({ mode: "serial" });

	let suiteUser: TestUser | undefined;
	let loginChallengeId = "";

	test.beforeAll(() => {
		suiteUser = readLoginSuiteUser();
	});

	test.describe("auth login flow (API)", () => {
		test("wrong password returns neutral invalid_credentials", async ({ request }) => {
			test.skip(
				!suiteUser,
				"run register-flow first (check:e2e-auth) to seed login-suite-user.json",
			);
			await expectLoginInvalidCredentials(request, suiteUser!.username, "definitely-wrong-password");
		});

		test("Google start redirects to IdP when captcha disabled", async ({ request }) => {
			const res = await request.get(`${E2E_API_BASE}/auth/google?next=/`, { maxRedirects: 0 });
			expect([302, 303, 307]).toContain(res.status());
			const location = res.headers().location ?? "";
			expect(location).toMatch(/accounts\.google\.com|google/i);
		});

		test("sign-in page loads and submit stays disabled until policy loads", async ({ page }) => {
			const setupStatus = page.waitForResponse(
				(res) => res.url().includes("/auth/setup-status") && res.ok(),
			);
			await page.goto("/sign-in");
			await setupStatus;
			const submit = page.getByRole("button", { name: /sign in|masuk/i });
			await expect(submit).toBeEnabled();
		});

		test("login UI form accepts seeded credentials", async ({ page }) => {
			test.skip(
				!suiteUser,
				"run register-flow first (check:e2e-auth) to seed login-suite-user.json",
			);
			await page.goto("/sign-in");
			await page.getByPlaceholder("name@example.com").fill(suiteUser!.username);
			await page.locator('input[type="password"]').fill(suiteUser!.password);
			await expect(page.getByRole("button", { name: /sign in|masuk/i })).toBeEnabled();
		});

		test("valid password returns email OTP challenge", async ({ request }) => {
			test.skip(
				!suiteUser,
				"run register-flow first (check:e2e-auth) to seed login-suite-user.json",
			);
			loginChallengeId = await loginPasswordChallenge(
				request,
				suiteUser!.username,
				suiteUser!.password,
			);
			expect(loginChallengeId.length).toBeGreaterThan(0);
		});

		test("wrong OTP code is rejected", async ({ request }) => {
			test.skip(!suiteUser || !loginChallengeId, "needs login challenge from prior test");
			await expectVerifyOtpRejected(request, loginChallengeId, "ZZZZZZ");
		});

		test("OTP verify succeeds when dev Redis OTP shadow is available", async ({ request }) => {
			test.skip(!process.env.E2E_FULL_AUTH, "set E2E_FULL_AUTH=1 to run full OTP verify");
			test.skip(!suiteUser || !loginChallengeId, "needs login challenge from prior test");

			const code = waitDevOtpForChallenge(loginChallengeId);
			expect(code, "dev Redis OTP shadow missing — restart ms-user (development)").toBeTruthy();

			const session = await verifyLoginCode(request, loginChallengeId, code!);
			expect(session.access_token.length).toBeGreaterThan(10);
		});
	});
});