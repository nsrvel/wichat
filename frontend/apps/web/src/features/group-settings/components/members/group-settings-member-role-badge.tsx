"use client";

import { Crown } from "lucide-react";
import { useTranslation } from "react-i18next";
import { hasCustomRoleColor } from "@/lib/member-role-display";
import { cn } from "@/lib/utils";

export function GroupSettingsMemberRoleBadge({
	role,
	color,
	hideColor = false,
}: {
	role: string;
	color?: string;
	hideColor?: boolean;
}) {
	const { t } = useTranslation();
	const isOwner = role === "owner";
	const label = t(`invite.role_${role}`, { defaultValue: role });
	const showDot = !isOwner && !hideColor && hasCustomRoleColor(color);

	const baseClass =
		"inline-flex h-6 max-w-[160px] items-center gap-1.5 rounded-full border px-2 text-[11px] font-medium leading-none tracking-tight transition-colors";

	if (isOwner) {
		return (
			<span
				className={cn(
					baseClass,
					"border-amber-500/30 bg-amber-500/10 text-amber-700 dark:text-amber-300",
				)}
			>
				<Crown className="size-3 shrink-0" strokeWidth={2} aria-hidden />
				<span className="truncate">{label}</span>
			</span>
		);
	}

	return (
		<span
			className={cn(
				baseClass,
				!showDot && "border-border/75 bg-muted/55 text-foreground/90",
			)}
			style={
				showDot && color
					? {
							backgroundColor: `color-mix(in oklab, ${color} 11%, var(--card))`,
							borderColor: `color-mix(in oklab, ${color} 26%, var(--border))`,
							color: "var(--foreground)",
						}
					: undefined
			}
		>
			{showDot ? (
				<span
					className="size-1.5 shrink-0 rounded-full"
					style={{
						backgroundColor: color,
						boxShadow: `0 0 0 2px color-mix(in oklab, ${color} 18%, transparent)`,
					}}
					aria-hidden
				/>
			) : null}
			<span className="truncate">{label}</span>
		</span>
	);
}
