"use client";

import { Crown } from "lucide-react";
import { useTranslation } from "react-i18next";
import { cn } from "@/lib/utils";

export function GroupSettingsMemberRoleBadge({ role }: { role: string }) {
	const { t } = useTranslation();
	const label = t(`invite.role_${role}`, { defaultValue: role });

	if (role === "owner") {
		return (
			<span
				className={cn(
					"inline-flex items-center gap-1.5 rounded-full px-2.5 py-1",
					"bg-amber-500/12 text-xs font-semibold tracking-tight text-amber-950",
					"ring-1 ring-inset ring-amber-500/22",
					"dark:bg-amber-400/14 dark:text-amber-50 dark:ring-amber-400/28",
				)}
			>
				<Crown
					className="size-3.5 shrink-0 text-amber-500 dark:text-amber-400"
					strokeWidth={2.2}
					aria-hidden
				/>
				<span className="truncate">{label}</span>
			</span>
		);
	}

	return (
		<span
			className={cn(
				"inline-flex items-center rounded-full px-2.5 py-1",
				"bg-background/90 text-xs font-medium tracking-tight text-foreground",
				"ring-1 ring-inset ring-border/70 shadow-[0_1px_0_0_color-mix(in_oklab,var(--foreground)_6%,transparent)]",
			)}
		>
			<span className="truncate">{label}</span>
		</span>
	);
}
