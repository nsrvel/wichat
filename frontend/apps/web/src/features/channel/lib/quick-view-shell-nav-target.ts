import type { ShellNavTarget } from "@/components/layout/shell-navigation";
import type { QuickViewId } from "@/lib/app-routes";

const QUICK_VIEW_NAV_IDS: Record<QuickViewId, string> = {
	mentions: "qv-mentions",
	discover: "qv-discover",
	drafts: "qv-drafts",
};

export function toQuickViewShellNavTarget(view: QuickViewId, label: string): ShellNavTarget {
	return {
		id: QUICK_VIEW_NAV_IDS[view],
		label,
		kind: "quickView",
	};
}
