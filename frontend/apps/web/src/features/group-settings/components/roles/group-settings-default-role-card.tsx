"use client";

import { ChevronRight } from "lucide-react";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { SettingsToggle } from "@/components/ui/settings-toggle";
import type { GroupSettingsRoleRow } from "@/features/group-settings/lib/group-settings-roles-view";
import { cn } from "@/lib/utils";

type GroupSettingsDefaultRoleCardProps = {
	role: GroupSettingsRoleRow;
	hideRoles: boolean;
	canManage: boolean;
	onHideRolesChange: (next: boolean) => void;
	onOpenRole: (roleId: string) => void;
};

export function GroupSettingsDefaultRoleCard({
	role,
	hideRoles,
	canManage,
	onHideRolesChange,
	onOpenRole,
}: GroupSettingsDefaultRoleCardProps) {
	const { t } = useTranslation();

	return (
		<div className="overflow-hidden rounded-xl border border-border/80 bg-field/30">
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
				className="group flex cursor-pointer flex-row items-center justify-between gap-4 p-4 transition-colors duration-150 hover:bg-field/50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-ring"
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

				<div className="flex shrink-0 items-center">
					<Button
						type="button"
						variant="ghost"
						size="icon"
						aria-label={t("group_settings.roles.default_role_edit")}
						className="size-8 text-muted-foreground transition-colors group-hover:bg-sidebar-accent group-hover:text-foreground"
						onClick={(e) => {
							e.stopPropagation();
							onOpenRole(role.id);
						}}
					>
						<ChevronRight className="size-4" aria-hidden />
					</Button>
				</div>
			</div>

			<div className="border-t border-border/70" />

			<div className="flex items-center justify-between gap-4 p-4">
				<div className="flex min-w-0 flex-col gap-1">
					<span className="text-sm font-semibold text-foreground">
						{t("group_settings.roles.hide_roles_title")}
					</span>
					<p className="text-xs leading-relaxed text-muted-foreground">
						{t("group_settings.roles.hide_roles_description")}
					</p>
				</div>
				<SettingsToggle
					checked={hideRoles}
					disabled={!canManage}
					onCheckedChange={onHideRolesChange}
					aria-label={t("group_settings.roles.hide_roles_title")}
				/>
			</div>
		</div>
	);
}
