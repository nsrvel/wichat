import { expect, type APIRequestContext, type BrowserContext } from "@playwright/test";

export const E2E_API_BASE =
	process.env.E2E_API_URL ?? "http://localhost:8080/api/v1";

export type TestUser = {
	id: string;
	username: string;
	displayName: string;
	email: string;
	password: string;
	accessToken: string;
};

export type SeededGroup = {
	id: string;
	name: string;
	inviteCode: string;
};

type Envelope<T> = { data: T };

function uniqueSuffix(): string {
	return `${Date.now().toString(36)}_${Math.random().toString(36).slice(2, 7)}`;
}

async function parseEnvelope<T>(res: Awaited<ReturnType<APIRequestContext["post"]>>): Promise<T> {
	const text = await res.text();
	expect(res.ok(), `API ${res.url()} failed (${res.status()}): ${text}`).toBeTruthy();
	const json = JSON.parse(text) as Envelope<T>;
	return json.data;
}

export async function createTestUser(
	request: APIRequestContext,
	prefix: string,
	displayName?: string,
): Promise<TestUser> {
	const suffix = uniqueSuffix();
	const username = `e2e_${prefix}_${suffix}`.toLowerCase().slice(0, 30);
	const email = `${username}@example.test`;
	const password = `E2ePass!_${suffix}`;
	const name = displayName ?? `E2E ${prefix} ${suffix}`;

	const res = await request.post(`${E2E_API_BASE}/auth/register`, {
		data: {
			email,
			username,
			password,
			display_name: name,
			birth_date: "1998-05-15",
			locale: "en",
		},
	});
	const raw = await res.text();
	expect(res.ok(), `register failed: ${raw}`).toBeTruthy();
	const envelope = JSON.parse(raw) as Envelope<
		| { access_token: string; user: { id: string } }
		| { challenge_id: string; user?: { id: string } }
	>;
	const data = envelope.data;
	let accessToken = "";
	let userId = "";
	if ("access_token" in data && data.access_token) {
		accessToken = data.access_token;
		userId = data.user.id;
	} else if ("challenge_id" in data && data.challenge_id) {
		const { waitDevOtpForChallenge } = await import("./auth-helpers");
		const otp = waitDevOtpForChallenge(data.challenge_id);
		expect(otp, "register verify needs dev Redis OTP shadow (restart ms-user in development)").toBeTruthy();
		const verifyRes = await request.post(`${E2E_API_BASE}/auth/register/verify`, {
			data: { challenge_id: data.challenge_id, code: otp },
		});
		const verified = await parseEnvelope<{ access_token: string; user: { id: string } }>(
			verifyRes,
		);
		accessToken = verified.access_token;
		userId = verified.user.id;
	} else {
		throw new Error(`unexpected register response: ${raw}`);
	}
	return {
		id: userId,
		username,
		displayName: name,
		email,
		password,
		accessToken,
	};
}

export async function loginInBrowser(context: BrowserContext, user: TestUser): Promise<void> {
	const { loginPasswordChallenge, waitDevOtpForChallenge } = await import("./auth-helpers");
	const challengeId = await loginPasswordChallenge(context.request, user.username, user.password);
	const code = waitDevOtpForChallenge(challengeId);
	expect(code, "browser login needs dev Redis OTP shadow (restart ms-user in development)").toBeTruthy();
	const res = await context.request.post(`${E2E_API_BASE}/auth/login/verify`, {
		data: { challenge_id: challengeId, code },
	});
	expect(res.ok(), `Browser login verify failed (${res.status()}): ${await res.text()}`).toBeTruthy();
}

export async function createGroupWithInvite(
	request: APIRequestContext,
	owner: TestUser,
	groupName: string,
): Promise<SeededGroup> {
	const groupRes = await request.post(`${E2E_API_BASE}/groups`, {
		headers: { Authorization: `Bearer ${owner.accessToken}` },
		data: { name: groupName },
	});
	const group = await parseEnvelope<{ id: string; name: string }>(groupRes);

	const inviteRes = await request.post(
		`${E2E_API_BASE}/groups/${encodeURIComponent(group.id)}/invites`,
		{
			headers: { Authorization: `Bearer ${owner.accessToken}` },
			data: {},
		},
	);
	const invite = await parseEnvelope<{ code: string }>(inviteRes);

	return {
		id: group.id,
		name: group.name,
		inviteCode: invite.code,
	};
}

export async function redeemInvite(
	request: APIRequestContext,
	member: TestUser,
	inviteCode: string,
) {
	return request.post(`${E2E_API_BASE}/invites/redeem`, {
		headers: { Authorization: `Bearer ${member.accessToken}` },
		data: { code: inviteCode },
	});
}

export async function createRoleAndAssign(
	request: APIRequestContext,
	owner: TestUser,
	groupId: string,
	roleName: string,
	targetUserId: string,
	permissions: string[] = [],
): Promise<{ id: string; name: string }> {
	const roleRes = await request.post(
		`${E2E_API_BASE}/groups/${encodeURIComponent(groupId)}/roles`,
		{
			headers: { Authorization: `Bearer ${owner.accessToken}` },
			data: {
				name: roleName,
				color: "#5865F2",
				hoist: false,
				mentionable: false,
				permissions,
			},
		},
	);
	const role = await parseEnvelope<{ id: string; name: string }>(roleRes);

	const assignRes = await request.post(
		`${E2E_API_BASE}/groups/${encodeURIComponent(groupId)}/roles/${encodeURIComponent(role.id)}/members`,
		{
			headers: { Authorization: `Bearer ${owner.accessToken}` },
			data: { user_ids: [targetUserId] },
		},
	);
	expect(assignRes.ok(), `Role assign failed (${assignRes.status()})`).toBeTruthy();
	return role;
}
