"use client";

import { AtSign, Compass, FileText } from "lucide-react";
import { useTranslation } from "react-i18next";
import { ConversationHeader } from "@/features/chat/components/conversation-header";
import { toQuickViewShellNavTarget } from "@/features/channel/lib/quick-view-shell-nav-target";
import { WorkspaceComingSoonState } from "@/features/shell/components/workspace-coming-soon-state";
import type { QuickViewId } from "@/lib/app-routes";

const QUICK_VIEW_META: Record<
	QuickViewId,
	{ labelKey: string; descriptionKey: string; icon: typeof AtSign }
> = {
	mentions: {
		labelKey: "shell.sidebar.mentions",
		descriptionKey: "shell.quick_view.mentions_description",
		icon: AtSign,
	},
	discover: {
		labelKey: "shell.sidebar.discover",
		descriptionKey: "shell.quick_view.discover_description",
		icon: Compass,
	},
	drafts: {
		labelKey: "shell.sidebar.drafts",
		descriptionKey: "shell.quick_view.drafts_description",
		icon: FileText,
	},
};

type QuickViewRoutePageProps = {
	groupId: string;
	quickView: QuickViewId;
	onOpenSidebar?: () => void;
};

export function QuickViewRoutePage({ groupId, quickView, onOpenSidebar }: QuickViewRoutePageProps) {
	void groupId;
	const { t } = useTranslation();
	const meta = QUICK_VIEW_META[quickView];
	const label = t(meta.labelKey);
	const active = toQuickViewShellNavTarget(quickView, label);
	const Icon = meta.icon;

	return (
		<div className="flex h-full min-w-0 flex-1 flex-col bg-chat-surface text-foreground">
			<div className="lg:hidden">
				<ConversationHeader active={active} onOpenSidebar={onOpenSidebar} />
			</div>
			<div className="flex min-h-0 flex-1 flex-col items-center justify-center px-4 py-8">
				<WorkspaceComingSoonState
					icon={<Icon className="h-6 w-6" strokeWidth={1.75} aria-hidden />}
					descriptionKey={meta.descriptionKey}
				/>
			</div>
		</div>
	);
}
