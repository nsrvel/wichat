"use client";

import { useTranslation } from "react-i18next";
import { cn } from "@/lib/utils";

type GroupMemberStatsPillProps = {
	onlineCount: number;
	memberCount: number;
	className?: string;
};

/**
 * Compact pill displaying live online and total member counts for a group card.
 */
export function GroupMemberStatsPill({
	onlineCount,
	memberCount,
	className,
}: GroupMemberStatsPillProps) {
	const { t } = useTranslation();

	return (
		<div
			className={cn(
				"inline-flex flex-wrap items-center justify-center gap-x-2.5 gap-y-1 text-xs font-medium text-muted-foreground sm:text-[13px]",
				className,
			)}
		>
			<span className="inline-flex items-center gap-1.5">
				<span className="size-2 shrink-0 rounded-full bg-presence-online" aria-hidden />
				{t("invite.landing_stats_online", { count: onlineCount })}
			</span>
			<span className="text-border" aria-hidden>
				·
			</span>
			<span>{t("invite.landing_stats_members", { count: memberCount })}</span>
		</div>
	);
}
