"use client";

import { ArrowLeft, Plus } from "lucide-react";
import { useTranslation } from "react-i18next";
import {
	roleDisplayTitle,
	type GroupSettingsRoleRow,
} from "@/features/group-settings/lib/group-settings-roles-view";
import { cn } from "@/lib/utils";

type RoleDetailSidebarProps = {
	roles: GroupSettingsRoleRow[];
	selectedRoleId: string;
	canManage: boolean;
	onSelectRole: (roleId: string) => void;
	onBack: () => void;
	onCreateRole: () => void;
};

export function RoleDetailSidebar({
	roles,
	selectedRoleId,
	canManage,
	onSelectRole,
	onBack,
	onCreateRole,
}: RoleDetailSidebarProps) {
	const { t } = useTranslation();

	return (
		<aside
			aria-label={t("group_settings.roles.detail.sidebar_aria")}
			className="flex shrink-0 flex-col border-b border-border bg-background/60 md:w-56 md:border-r md:border-b-0 lg:h-full lg:w-60 lg:overflow-hidden"
		>
			<div className="flex items-center justify-between gap-2 px-3 pt-10 pb-3 sm:px-4 sm:pt-12">
				<button
					type="button"
					onClick={onBack}
					className="inline-flex h-8 items-center gap-2 rounded-md px-2 text-sm font-semibold text-foreground/90 transition-colors hover:bg-sidebar-accent hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
				>
					<ArrowLeft className="size-4 shrink-0" strokeWidth={2.2} aria-hidden />
					<span>{t("group_settings.roles.back_to_list")}</span>
				</button>

				{canManage ? (
					<button
						type="button"
						onClick={onCreateRole}
						aria-label={t("group_settings.roles.detail.create_role_aria")}
						className="inline-flex size-8 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-sidebar-accent hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
					>
						<Plus className="size-4" strokeWidth={2.2} aria-hidden />
					</button>
				) : null}
			</div>

			<div className="thin-scroll flex gap-1 overflow-x-auto px-3 pb-3 md:flex-1 md:flex-col md:overflow-x-hidden md:overflow-y-auto md:px-3 md:pb-6">
				{roles.map((item) => {
					const active = item.id === selectedRoleId;
					const label = roleDisplayTitle(item, t);
					return (
						<button
							key={item.id}
							type="button"
							onClick={() => onSelectRole(item.id)}
							aria-current={active ? "true" : undefined}
							className={cn(
								"flex min-h-9 shrink-0 items-center gap-2.5 rounded-lg px-2.5 py-1.5 text-left text-sm transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring md:w-full",
								active
									? "bg-sidebar-accent font-semibold text-foreground"
									: "font-medium text-muted-foreground hover:bg-sidebar-accent/60 hover:text-foreground",
							)}
						>
							<span
								className="size-3 shrink-0 rounded-full ring-1 ring-black/10 dark:ring-white/15"
								style={{ backgroundColor: item.color }}
								aria-hidden
							/>
							<span className="truncate">{label}</span>
						</button>
					);
				})}
			</div>
		</aside>
	);
}
