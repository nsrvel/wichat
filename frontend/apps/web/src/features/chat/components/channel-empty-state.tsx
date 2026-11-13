"use client";

import { MessageSquare } from "lucide-react";
import { useTranslation } from "react-i18next";

export function ChannelEmptyState() {
	const { t } = useTranslation();
	return (
		<div className="flex flex-col items-center justify-center gap-3 text-center">
			<div className="flex size-12 items-center justify-center rounded-2xl border border-chat-bubble-border bg-chat-bubble text-muted-foreground">
				<MessageSquare className="size-6" strokeWidth={1.75} aria-hidden />
			</div>
			<div className="max-w-sm space-y-1">
				<p className="text-[15px] font-semibold text-sidebar-foreground">{t("chat.empty_title")}</p>
				<p className="text-[13px] leading-relaxed text-muted-foreground">
					{t("chat.empty_description")}
				</p>
			</div>
		</div>
	);
}
