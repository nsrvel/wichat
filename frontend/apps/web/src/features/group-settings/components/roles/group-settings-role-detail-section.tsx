"use client";

import type { GroupMember } from "@wichat/contracts";
import { Ellipsis } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { RoleDetailDisplayTab } from "@/features/group-settings/components/roles/detail/role-detail-display-tab";
import { RoleDetailMembersTab } from "@/features/group-settings/components/roles/detail/role-detail-members-tab";
import { RoleDetailPermissionsTab } from "@/features/group-settings/components/roles/detail/role-detail-permissions-tab";
import { RoleDetailSidebar } from "@/features/group-settings/components/roles/detail/role-detail-sidebar";
import {
	groupSettingsRoleDetailPanePaddingClass,
	groupSettingsSectionTitleClass,
} from "@/features/group-settings/lib/group-settings-chrome";
import {
	roleDisplayTitle,
	type GroupSettingsRoleDetailTab,
	type GroupSettingsRoleRow,
} from "@/features/group-settings/lib/group-settings-roles-view";
import { cn } from "@/lib/utils";

type GroupSettingsRoleDetailSectionProps = {
	roles: GroupSettingsRoleRow[];
	role: GroupSettingsRoleRow;
	members: GroupMember[];
	canManage: boolean;
	onSelectRole: (roleId: string) => void;
	onBack: () => void;
	onCreateRole: () => void;
};

export function GroupSettingsRoleDetailSection({
	roles,
	role,
	members,
	canManage,
	onSelectRole,
	onBack,
	onCreateRole,
}: GroupSettingsRoleDetailSectionProps) {
	const { t } = useTranslation();
	const [activeTab, setActiveTab] = useState<GroupSettingsRoleDetailTab>("display");
	const roleLabel = roleDisplayTitle(role, t);
	const isEveryone = role.id === "everyone";
	const currentTab: GroupSettingsRoleDetailTab = isEveryone ? "permissions" : activeTab;

	const tabs: Array<{ id: GroupSettingsRoleDetailTab; label: string; disabled?: boolean }> = [
		{
			id: "display",
			label: t("group_settings.roles.detail.tab_display"),
			disabled: isEveryone,
		},
		{
			id: "permissions",
			label: t("group_settings.roles.detail.tab_permissions"),
		},
		{
			id: "members",
			label: t("group_settings.roles.detail.tab_members", {
				count: role.memberCount,
			}),
			disabled: isEveryone,
		},
	];

	return (
		<div className="flex min-h-0 flex-1 flex-col md:flex-row lg:h-full lg:overflow-hidden">
			<RoleDetailSidebar
				roles={roles}
				selectedRoleId={role.id}
				canManage={canManage}
				onSelectRole={onSelectRole}
				onBack={onBack}
				onCreateRole={onCreateRole}
			/>

			<div className="flex min-h-0 min-w-0 flex-1 flex-col lg:h-full lg:overflow-hidden">
				{/* Fixed Header & Tab Bar (never scrolls) */}
				<div className={cn("shrink-0 pt-10 sm:pt-12", groupSettingsRoleDetailPanePaddingClass)}>
					<div className="mx-auto w-full max-w-2xl lg:mx-0">
						{/* Header */}
						<div className="flex items-center justify-between gap-3">
							<h2 className={cn(groupSettingsSectionTitleClass, "truncate")}>
								{t("group_settings.roles.detail.edit_title", { name: roleLabel })}
							</h2>
							<button
								type="button"
								aria-label={t("group_settings.roles.detail.more_actions_aria")}
								className="inline-flex size-8 shrink-0 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-sidebar-accent hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
							>
								<Ellipsis className="size-5" strokeWidth={2} aria-hidden />
							</button>
						</div>

						{/* Tab bar */}
						<div
							role="tablist"
							aria-label={t("group_settings.roles.detail.tabs_aria")}
							className="mt-5 flex items-center gap-6 border-b border-border"
						>
							{tabs.map((tab) => {
								const active = currentTab === tab.id;
								return (
									<button
										key={tab.id}
										type="button"
										role="tab"
										disabled={tab.disabled}
										aria-disabled={tab.disabled}
										aria-selected={active}
										onClick={() => setActiveTab(tab.id)}
										className={cn(
											"relative -mb-px pb-3 text-sm font-semibold transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
											tab.disabled && "cursor-not-allowed opacity-40 hover:text-muted-foreground",
											active && !tab.disabled
												? "border-b-2 border-primary text-primary"
												: "border-b-2 border-transparent text-muted-foreground hover:text-foreground",
										)}
									>
										{tab.label}
									</button>
								);
							})}
						</div>
					</div>
				</div>

				{/* Scrollable Tab Content (scrollbar sits at far-right edge of web viewport) */}
				<div
					className={cn(
						"overlay-scroll flex min-h-0 flex-1 flex-col pb-10 lg:overflow-y-auto lg:overscroll-contain",
						groupSettingsRoleDetailPanePaddingClass,
					)}
				>
					<div className="mx-auto flex w-full max-w-2xl flex-col lg:mx-0">
						{currentTab === "display" ? (
							<RoleDetailDisplayTab
								key={role.id}
								role={role}
								roleLabel={roleLabel}
								canManage={canManage}
							/>
						) : null}

						{currentTab === "permissions" ? (
							<RoleDetailPermissionsTab key={role.id} role={role} canManage={canManage} />
						) : null}

						{currentTab === "members" ? (
							<RoleDetailMembersTab
								key={role.id}
								role={role}
								members={members}
								canManage={canManage}
							/>
						) : null}
					</div>
				</div>
			</div>
		</div>
	);
}
