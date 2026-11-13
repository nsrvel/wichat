"use client";

import { ChevronRight } from "lucide-react";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import type { GroupSettingsRoleRow } from "@/features/group-settings/lib/group-settings-roles-view";
import { cn } from "@/lib/utils";

type GroupSettingsDefaultRoleCardProps = {
	role: GroupSettingsRoleRow;
	onOpenRole: (roleId: string) => void;
};

export function GroupSettingsDefaultRoleCard({
	role,
	onOpenRole,
}: GroupSettingsDefaultRoleCardProps) {
	const { t } = useTranslation();

	return (
		<div
			role="button"
			tabIndex={0}
			onClick={() => onOpenRole(role.id)}
			onKeyDown={(e) => {
				if (e.key === "Enter" || e.key === " ") {
					e.preventDefault();
					onOpenRole(role.id);
				}
			}}
			className="group relative flex flex-row items-center justify-between gap-4 rounded-xl border border-border/80 bg-field/30 p-4 transition-all duration-150 hover:border-primary/40 hover:bg-field/50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring cursor-pointer"
		>
			<div className="flex min-w-0 flex-col gap-1">
				<div className="flex flex-wrap items-center gap-2">
					<span className="font-semibold text-foreground text-base tracking-tight">
						{t("group_settings.roles.default_role_title")}
					</span>
					<span
						className={cn(
							"inline-flex items-center rounded-full px-2 py-0.5 text-[11px] font-medium tracking-normal",
							"bg-secondary text-secondary-foreground border border-border/60",
						)}
					>
						{t("group_settings.roles.default_role_badge")}
					</span>
				</div>
				<p className="text-xs text-muted-foreground leading-relaxed">
					{t("group_settings.roles.default_role_description")}
				</p>
			</div>

			<div className="flex items-center shrink-0">
				<Button
					type="button"
					variant="ghost"
					size="icon"
					aria-label={t("group_settings.roles.default_role_edit")}
					className="size-8 text-muted-foreground transition-colors group-hover:text-foreground group-hover:bg-sidebar-accent"
					onClick={(e) => {
						e.stopPropagation();
						onOpenRole(role.id);
					}}
				>
					<ChevronRight className="size-4" aria-hidden />
				</Button>
			</div>
		</div>
	);
}
