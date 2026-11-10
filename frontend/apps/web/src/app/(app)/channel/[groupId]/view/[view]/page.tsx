import { notFound } from "next/navigation";
import { QuickViewRoutePage } from "@/features/channel/quick-view-route-page";
import { isQuickViewId } from "@/lib/app-routes";

type PageProps = {
	params: Promise<{ groupId: string; view: string }>;
};

export default async function QuickViewPage({ params }: PageProps) {
	const { groupId, view } = await params;
	if (!isQuickViewId(view)) {
		notFound();
	}
	return <QuickViewRoutePage groupId={groupId} quickView={view} />;
}
