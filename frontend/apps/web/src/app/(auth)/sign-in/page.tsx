import { LoginPage } from "@/features/auth/login/login-page";
import { pickAuthReturnTo, type AuthSearchParams } from "@/lib/auth-invite";
import { safeNotice } from "@/lib/auth-notice";

export default async function Page({
	searchParams,
}: {
	searchParams: Promise<AuthSearchParams & { error?: string }>;
}) {
	const params = await searchParams;
	const { error, ...rest } = params;
	const authReturnTo = pickAuthReturnTo(rest);

	return <LoginPage authReturnTo={authReturnTo} notice={safeNotice(error)} />;
}
