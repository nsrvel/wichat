"use client";

import { useQuery } from "@tanstack/react-query";
import { Shield, SquareArrowOutUpRight, User, Users } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { useTranslation } from "react-i18next";
import { myGroupsQueryOpts } from "@/services/group/group.query";
import { SidebarPlainNavRow } from "@/components/layout/channel-sidebar/sidebar-plain-nav-row";
import { SIDEBAR_INSET_X } from "@/components/layout/channel-sidebar/sidebar-nav-styles";
import { SidebarStaticCategoryRow } from "@/components/layout/channel-sidebar/sidebar-static-category-row";
import { sidebarNavGlyphIconClass } from "@/components/layout/sidebar-list-icon";
import {
	groupSettingsNavGroups,
	type GroupSettingsNavItem,
} from "@/features/group-settings/lib/group-settings-nav";
import type {
	GroupSettingsSection,
	GroupSettingsTarget,
} from "@/features/group-settings/lib/group-settings-target";
import { cn } from "@/lib/utils";

const NAV_ICON: Record<GroupSettingsSection, LucideIcon> = {
	profile: User,
	members: Users,
	roles: Shield,
};

type GroupSettingsSidebarProps = {
	target: GroupSettingsTarget;
	activeSection: GroupSettingsSection;
	onSectionChange: (section: GroupSettingsSection) => void;
	onOpenMembersInWorkspace: () => void;
};

function SettingsNavItem({
	item,
	activeSection,
	onSectionChange,
	onOpenMembersInWorkspace,
}: {
	item: GroupSettingsNavItem;
	activeSection: GroupSettingsSection;
	onSectionChange: (section: GroupSettingsSection) => void;
	onOpenMembersInWorkspace: () => void;
}) {
	const { t } = useTranslation();
	const active = item.section === activeSection;
	const Icon = NAV_ICON[item.id];

	return (
		<SidebarPlainNavRow
			icon={<Icon className={sidebarNavGlyphIconClass} strokeWidth={2} />}
			label={t(item.sectionKey)}
			selected={active}
			onSelect={() => onSectionChange(item.section)}
			trailingAction={
				item.id === "members"
					? {
							label: t("group_settings.members.open_in_workspace"),
							visible: "always",
							onClick: onOpenMembersInWorkspace,
							icon: <SquareArrowOutUpRight className="h-3.5 w-3.5" strokeWidth={2} aria-hidden />,
						}
					: undefined
			}
		/>
	);
}

export function GroupSettingsSidebar({
	target,
	activeSection,
	onSectionChange,
	onOpenMembersInWorkspace,
}: GroupSettingsSidebarProps) {
	const { t } = useTranslation();
	const { data: myGroups = [] } = useQuery(myGroupsQueryOpts());
	const isOwner =
		target.role === "owner" || myGroups.some((g) => g.id === target.groupId && g.role === "owner");

	return (
		<aside className="flex h-full min-h-full w-full max-h-[100dvh] flex-col">
			<div className={cn("flex min-h-0 flex-1 flex-col pt-12 sm:pt-14", SIDEBAR_INSET_X)}>
				<div className="thin-scroll flex min-h-0 flex-1 flex-col overflow-y-auto pb-3">
					<p className="mb-2 truncate px-0.5 text-sm font-semibold text-sidebar-foreground">
						{target.groupName}
					</p>
					<nav className="flex flex-col" aria-label={t("group_settings.nav_label")}>
						{groupSettingsNavGroups.map((group, index) => (
							<div
								key={group.labelKey}
								className={cn(
									"group/cat",
									index > 0 && "mt-3 border-t border-sidebar-border/35 pt-3",
								)}
							>
								<SidebarStaticCategoryRow title={t(group.labelKey)} />
								<div className="flex flex-col gap-0">
									{group.items
										.filter((item) => (item.id !== "members" && item.id !== "roles") || isOwner)
										.map((item) => (
											<SettingsNavItem
												key={item.id}
												item={item}
												activeSection={activeSection}
												onSectionChange={onSectionChange}
												onOpenMembersInWorkspace={onOpenMembersInWorkspace}
											/>
										))}
								</div>
							</div>
						))}
					</nav>
				</div>
			</div>
		</aside>
	);
}
