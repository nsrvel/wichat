export type AuthSearchParams = {
	invite?: string;
	token?: string;
	server?: string;
	serverName?: string;
	as?: string;
};

export function pickInviteToken(params: AuthSearchParams): string | undefined {
	const value = params.invite || params.token;
	return value?.trim() ? value.trim() : undefined;
}

export function pickServerName(params: AuthSearchParams): string | undefined {
	const value = params.serverName || params.server;
	return value?.trim() ? value.trim() : undefined;
}

/** Sign-in: invite link defaults to guest; `as=account` = existing member login. */
export function resolveSignInMode(
	inviteToken: string | undefined,
	as: string | undefined,
): "account" | "guest" {
	if (!inviteToken) {
		return "account";
	}
	return as === "account" ? "account" : "guest";
}

export function authPath(
	path: "/sign-in" | "/sign-up",
	opts?: { invite?: string; as?: "account" },
): string {
	if (!opts?.invite) {
		return path;
	}
	const q = new URLSearchParams();
	q.set("invite", opts.invite);
	if (opts.as === "account") {
		q.set("as", "account");
	}
	return `${path}?${q.toString()}`;
}
