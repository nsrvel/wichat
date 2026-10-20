import { RegisterPage } from "@/features/auth/register/register-page";
import {
	pickInviteToken,
	pickServerName,
	type AuthSearchParams,
} from "@/lib/auth-invite";
import { safeNotice } from "@/lib/auth-notice";

export default async function Page({
	searchParams,
}: {
	searchParams: Promise<AuthSearchParams & { error?: string }>;
}) {
	const params = await searchParams;
	const { error, ...rest } = params;
	const inviteToken = pickInviteToken(rest);
	const serverName = pickServerName(rest);

	return (
		<RegisterPage
			inviteToken={inviteToken}
			serverName={serverName}
			notice={safeNotice(error)}
		/>
	);
}
