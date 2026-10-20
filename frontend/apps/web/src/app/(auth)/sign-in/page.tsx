import { GuestPage } from "@/features/auth/guest/guest-page";
import { LoginPage } from "@/features/auth/login/login-page";
import {
	pickInviteToken,
	pickServerName,
	resolveSignInMode,
	type AuthSearchParams,
} from "@/lib/auth-invite";
import { safeNotice } from "@/lib/auth-notice";

export default async function Page({
	searchParams,
}: {
	searchParams: Promise<AuthSearchParams & { error?: string }>;
}) {
	const params = await searchParams;
	const { error, as, ...rest } = params;
	const inviteToken = pickInviteToken(rest);
	const serverName = pickServerName(rest);
	const mode = resolveSignInMode(inviteToken, as);
	const notice = safeNotice(error);

	if (mode === "guest") {
		return (
			<GuestPage inviteToken={inviteToken} serverName={serverName} notice={notice} />
		);
	}

	return <LoginPage inviteToken={inviteToken} serverName={serverName} notice={notice} />;
}
