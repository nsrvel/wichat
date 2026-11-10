import { InviteLandingPage } from "@/features/invite/components/invite-landing-page";

export default async function Page({ params }: { params: Promise<{ token: string }> }) {
	const { token } = await params;

	return <InviteLandingPage inviteToken={decodeURIComponent(token)} />;
}
