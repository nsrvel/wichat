import { redirect } from "next/navigation";
import { InviteLandingPage } from "@/features/invite/components/invite-landing-page";
import { inviteLandingPath, pickInviteToken, type AuthSearchParams } from "@/lib/auth-invite";

export default async function Page({
	searchParams,
}: {
	searchParams: Promise<AuthSearchParams>;
}) {
	const params = await searchParams;
	const inviteToken = pickInviteToken(params);

	if (inviteToken) {
		redirect(inviteLandingPath(inviteToken));
	}

	return <InviteLandingPage inviteToken="" />;
}
