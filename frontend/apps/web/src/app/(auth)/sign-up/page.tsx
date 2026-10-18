import { safeNotice } from "@/features/auth/notice";
import { SetupPage } from "@/features/auth/setup";

export default async function Page({
	searchParams,
}: {
	searchParams: Promise<{ error?: string }>;
}) {
	const { error } = await searchParams;

	return <SetupPage notice={safeNotice(error)} />;
}
