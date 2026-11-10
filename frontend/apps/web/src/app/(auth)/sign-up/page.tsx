import { redirect } from "next/navigation";
import { RegisterPage } from "@/features/auth/register/register-page";
import { inviteLandingPath, pickInviteToken, pickServerName, type AuthSearchParams } from "@/lib/auth-invite";
import { safeNotice } from "@/lib/auth-notice";

export default async function Page({
	searchParams,
}: {
	searchParams: Promise<AuthSearchParams & { error?: string; flow?: string }>;
}) {
	const params = await searchParams;
	const { error, flow, ...rest } = params;
	const inviteToken = pickInviteToken(rest);
	const serverName = pickServerName(rest);

	if (inviteToken && flow !== "register") {
		redirect(inviteLandingPath(inviteToken));
	}

	return (
		<RegisterPage inviteToken={inviteToken} serverName={serverName} notice={safeNotice(error)} />
	);
}
