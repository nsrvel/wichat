"use client";

import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";

type WorkspaceComingSoonStateProps = {
	icon: ReactNode;
	descriptionKey: string;
};

export function WorkspaceComingSoonState({ icon, descriptionKey }: WorkspaceComingSoonStateProps) {
	const { t } = useTranslation();
	return (
		<div className="flex flex-col items-center justify-center gap-3 text-center">
			<div className="flex h-12 w-12 items-center justify-center rounded-2xl border border-chat-bubble-border bg-chat-bubble text-muted-foreground">
				{icon}
			</div>
			<div className="max-w-sm space-y-1">
				<p className="text-[15px] font-semibold text-sidebar-foreground">
					{t("shell.quick_view.coming_soon_title")}
				</p>
				<p className="text-[13px] leading-relaxed text-muted-foreground">{t(descriptionKey)}</p>
			</div>
		</div>
	);
}
