import { expect, type APIRequestContext, type BrowserContext } from "@playwright/test";

export const E2E_API_BASE = process.env.E2E_API_URL ?? "http://localhost:8080/api/v1";

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
	const data = await parseEnvelope<{ access_token: string; user: { id: string } }>(res);
	return {
		id: data.user.id,
		username,
		displayName: name,
		email,
		password,
		accessToken: data.access_token,
	};
}

export async function loginInBrowser(context: BrowserContext, user: TestUser): Promise<void> {
	const res = await context.request.post(`${E2E_API_BASE}/auth/login`, {
		data: {
			login: user.username,
			password: user.password,
			locale: "en",
		},
	});
	expect(res.ok(), `Browser login failed (${res.status()})`).toBeTruthy();
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
