import { notFound } from "next/navigation";
import { QuickViewRoutePageClient } from "@/features/channel/quick-view-route-page-client";
import { isQuickViewId } from "@/lib/app-routes";

type PageProps = {
	params: Promise<{ groupId: string; view: string }>;
};

export default async function QuickViewPage({ params }: PageProps) {
	const { groupId, view } = await params;
	if (!isQuickViewId(view)) {
		notFound();
	}
	return <QuickViewRoutePageClient groupId={groupId} quickView={view} />;
}
