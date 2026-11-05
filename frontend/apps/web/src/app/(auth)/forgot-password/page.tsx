import { ForgotPasswordPage } from "@/features/auth/forgot-password/forgot-password-page";
import { safeNotice } from "@/lib/auth-notice";

export default async function Page({
	searchParams,
}: {
	searchParams: Promise<{ error?: string }>;
}) {
	const params = await searchParams;
	const notice = safeNotice(params.error);

	return <ForgotPasswordPage notice={notice} />;
}
