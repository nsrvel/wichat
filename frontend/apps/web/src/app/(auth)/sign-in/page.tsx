import { LoginPage } from "@/features/auth/login";
import { safeNotice } from "@/features/auth/notice";

export default async function Page({
	searchParams,
}: {
	searchParams: Promise<{ error?: string }>;
}) {
	const { error } = await searchParams;

	return <LoginPage notice={safeNotice(error)} />;
}
