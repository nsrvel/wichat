export type AuthSearchParams = {
	invite?: string;
	token?: string;
	server?: string;
	serverName?: string;
	next?: string;
	flow?: string;
};

export function pickInviteToken(params: AuthSearchParams): string | undefined {
	const value = params.invite || params.token;
	return value?.trim() ? value.trim() : undefined;
}

export function pickServerName(params: AuthSearchParams): string | undefined {
	const value = params.serverName || params.server;
	return value?.trim() ? value.trim() : undefined;
}

/** Safe in-app return path for post-auth redirect (invite landing, etc.). */
export function pickAuthReturnTo(
	params: Pick<AuthSearchParams, "next" | "invite" | "token">,
): string {
	const next = params.next?.trim();
	if (next?.startsWith("/") && !next.startsWith("//")) {
		return next;
	}
	const invite = pickInviteToken(params);
	if (invite) {
		return inviteLandingPath(invite);
	}
	return "/";
}

export function inviteLandingPath(invite: string): string {
	return `/invite/${encodeURIComponent(invite.trim())}`;
}

export function postAuthRedirectPath(returnTo?: string): string {
	const path = returnTo?.trim();
	if (path?.startsWith("/") && !path.startsWith("//")) {
		return path;
	}
	return "/";
}

export function authPathWithReturn(
	path: "/sign-in" | "/sign-up",
	returnTo: string,
	opts?: { flow?: "register"; serverName?: string },
): string {
	const q = new URLSearchParams();
	q.set("next", returnTo);
	if (opts?.flow === "register") {
		q.set("flow", "register");
	}
	const serverName = opts?.serverName?.trim();
	if (serverName) {
		q.set("serverName", serverName);
	}
	return `${path}?${q.toString()}`;
}

/** @deprecated Prefer authPathWithReturn + next= */
export function authPath(
	path: "/sign-in" | "/sign-up",
	opts?: { invite?: string; flow?: "register" },
): string {
	if (!opts?.invite) {
		return path;
	}
	return authPathWithReturn(path, inviteLandingPath(opts.invite), { flow: opts.flow });
}
