"use client";

import { useQuery } from "@tanstack/react-query";
import { useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { GroupSettingsRoleDetailSection } from "@/features/group-settings/components/roles/group-settings-role-detail-section";
import { GroupSettingsRolesTable } from "@/features/group-settings/components/roles/group-settings-roles-table";
import { GroupSettingsRolesToolbar } from "@/features/group-settings/components/roles/group-settings-roles-toolbar";
import {
	groupSettingsMainInnerClass,
	groupSettingsMainScrollClass,
	groupSettingsSectionHeaderClass,
	groupSettingsSectionLeadClass,
	groupSettingsSectionTitleClass,
} from "@/features/group-settings/lib/group-settings-chrome";
import type { GroupSettingsTarget } from "@/features/group-settings/lib/group-settings-target";
import {
	builtinRoleRows,
	matchesRoleSearch,
} from "@/features/group-settings/lib/group-settings-roles-view";
import { cn } from "@/lib/utils";
import { myGroupsQueryOpts } from "@/services/group/group.query";
import { groupMembersQueryOpts } from "@/services/invite/invite.query";

type GroupSettingsRolesSectionProps = {
	target: GroupSettingsTarget;
};

export function GroupSettingsRolesSection({ target }: GroupSettingsRolesSectionProps) {
	const { t } = useTranslation();
	const groupId = target.groupId;
	const [search, setSearch] = useState("");
	const [selectedRoleId, setSelectedRoleId] = useState<string | null>(null);
	const [pageIndex, setPageIndex] = useState(1);

	const ROLE_PAGE_SIZE = 10;

	const { data: myGroups = [] } = useQuery(myGroupsQueryOpts());
	const isOwner = useMemo(
		() => myGroups.some((g) => g.id === groupId && g.role === "owner"),
		[groupId, myGroups],
	);
	const targetIsOwner = target.role === "owner";
	const canManage = isOwner || targetIsOwner;

	const membersQuery = useQuery({
		...groupMembersQueryOpts(groupId, ""),
		enabled: Boolean(groupId),
	});

	const memberTotal = membersQuery.data?.total ?? 0;
	const roleRows = useMemo(() => builtinRoleRows(memberTotal), [memberTotal]);

	const visibleRows = useMemo(() => {
		return roleRows.filter((row) => {
			const label = row.isSystem ? t(`invite.role_${row.name}`) : row.name;
			return matchesRoleSearch(row, search, label);
		});
	}, [roleRows, search, t]);

	const roleTotal = visibleRows.length;
	const totalPages = Math.max(1, Math.ceil(roleTotal / ROLE_PAGE_SIZE));
	const hasPrev = pageIndex > 1;
	const hasNext = pageIndex < totalPages;
	const pageRows = useMemo(
		() => visibleRows.slice((pageIndex - 1) * ROLE_PAGE_SIZE, pageIndex * ROLE_PAGE_SIZE),
		[pageIndex, visibleRows],
	);

	useEffect(() => {
		// eslint-disable-next-line react-hooks/set-state-in-effect -- reset pagination when search query changes
		setPageIndex(1);
	}, [search]);

	useEffect(() => {
		if (pageIndex > totalPages) {
			// eslint-disable-next-line react-hooks/set-state-in-effect -- clamp page within range
			setPageIndex(totalPages);
		}
	}, [pageIndex, totalPages]);

	const selectedRole = useMemo(
		() => roleRows.find((r) => r.id === selectedRoleId) ?? null,
		[roleRows, selectedRoleId],
	);

	const onCreateRole = () => {
		toast.message(t("group_settings.roles.create_coming_soon"));
	};

	const loading = membersQuery.isPending && !membersQuery.data;

	return (
		<div className="flex min-h-0 flex-1 flex-col">
			<div className={groupSettingsMainScrollClass}>
				<div
					className={cn(
						groupSettingsMainInnerClass,
						"flex max-w-none flex-col gap-6 pb-4 lg:max-w-5xl",
					)}
				>
					{selectedRole ? (
						<GroupSettingsRoleDetailSection
							role={selectedRole}
							onBack={() => setSelectedRoleId(null)}
						/>
					) : (
						<>
							<div className={groupSettingsSectionHeaderClass}>
								<h2 className={groupSettingsSectionTitleClass}>
									{t("group_settings.roles.title")}
								</h2>
								<p className={groupSettingsSectionLeadClass}>{t("group_settings.roles.lead")}</p>
							</div>

							<GroupSettingsRolesToolbar
								search={search}
								onSearchChange={setSearch}
								onCreateRole={onCreateRole}
								canManage={canManage}
							/>

							<GroupSettingsRolesTable
								rows={pageRows}
								loading={loading}
								total={roleTotal}
								pageIndex={pageIndex}
								hasPrev={hasPrev}
								hasNext={hasNext}
								maxVisitedPage={pageIndex}
								onOpenRole={setSelectedRoleId}
								onPrev={() => setPageIndex((p) => Math.max(1, p - 1))}
								onNext={() => setPageIndex((p) => Math.min(totalPages, p + 1))}
								onGoToPage={(page) => {
									if (page >= 1 && page <= totalPages) {
										setPageIndex(page);
									}
								}}
							/>
						</>
					)}
				</div>
			</div>
		</div>
	);
}
