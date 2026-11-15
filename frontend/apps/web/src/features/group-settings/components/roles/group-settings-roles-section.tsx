"use client";

import { useQuery } from "@tanstack/react-query";
import { useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { GroupSettingsDefaultRoleCard } from "@/features/group-settings/components/roles/group-settings-default-role-card";
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
import {
	DEFAULT_ROLE_COLOR,
	matchesRoleSearch,
	roleDisplayTitle,
	toGroupSettingsRoleRow,
	type GroupSettingsRoleRow,
} from "@/features/group-settings/lib/group-settings-roles-view";
import type { GroupSettingsTarget } from "@/features/group-settings/lib/group-settings-target";
import { ApiError } from "@wichat/api-client/api-error";
import { cn } from "@/lib/utils";
import { myGroupsQueryOpts } from "@/services/group/group.query";
import {
	useCreateGroupRoleMutation,
	useUpdateGroupHideRolesMutation,
} from "@/services/group-role/group-role.mutation";
import { groupRolesQueryOpts } from "@/services/group-role/group-role.query";
import { groupMembersQueryOpts } from "@/services/invite/invite.query";

type GroupSettingsRolesSectionProps = {
	target: GroupSettingsTarget;
};

const ROLE_PAGE_SIZE = 10;

export function GroupSettingsRolesSection({ target }: GroupSettingsRolesSectionProps) {
	const { t } = useTranslation();
	const groupId = target.groupId;
	const [search, setSearch] = useState("");
	const [selectedRoleId, setSelectedRoleId] = useState<string | null>(null);
	const [pageIndex, setPageIndex] = useState(1);

	const { data: myGroups = [] } = useQuery(myGroupsQueryOpts());
	const currentGroup = useMemo(
		() => myGroups.find((g) => g.id === groupId),
		[groupId, myGroups],
	);

	const isOwner = currentGroup?.role === "owner" || target.role === "owner";
	const effectivePerms = useMemo(
		() => currentGroup?.effective_permissions ?? [],
		[currentGroup?.effective_permissions],
	);

	const canGrantAdministrator = isOwner || effectivePerms.includes("administrator");
	const canManage =
		isOwner ||
		effectivePerms.includes("administrator") ||
		effectivePerms.includes("manage_roles");

	const rolesQuery = useQuery(groupRolesQueryOpts(groupId));
	const membersQuery = useQuery({
		...groupMembersQueryOpts(groupId, ""),
		enabled: Boolean(groupId),
	});
	const createMutation = useCreateGroupRoleMutation(groupId);
	const hideRolesMutation = useUpdateGroupHideRolesMutation(groupId);

	const memberTotal = membersQuery.data?.total ?? 0;
	const hideRoles = rolesQuery.data?.hide_roles ?? currentGroup?.hide_roles ?? false;

	const allRoles = useMemo<GroupSettingsRoleRow[]>(
		() => (rolesQuery.data?.roles ?? []).map(toGroupSettingsRoleRow),
		[rolesQuery.data?.roles],
	);

	const defaultRole = useMemo<GroupSettingsRoleRow>(() => {
		const found = allRoles.find((r) => r.isDefault || r.slug === "everyone");
		if (found) {
			return found;
		}
		return {
			id: "everyone",
			slug: "everyone",
			name: "@everyone",
			memberCount: memberTotal,
			isDefault: true,
			isSystem: true,
			color: DEFAULT_ROLE_COLOR,
			hoist: false,
			mentionable: false,
			permissions: [],
		};
	}, [allRoles, memberTotal]);

	const roleRows = useMemo(
		() => allRoles.filter((r) => !(r.isDefault || r.slug === "everyone")),
		[allRoles],
	);

	const sidebarRoles = useMemo(
		() => (allRoles.length > 0 ? [...roleRows, defaultRole] : []),
		[allRoles.length, defaultRole, roleRows],
	);

	const visibleRows = useMemo(() => {
		return roleRows.filter((row) => {
			const label = roleDisplayTitle(row, t);
			return matchesRoleSearch(row, search, label);
		});
	}, [roleRows, search, t]);

	const roleTotal = visibleRows.length;
	const totalPages = Math.max(1, Math.ceil(roleTotal / ROLE_PAGE_SIZE));
	const currentPageIndex = Math.min(pageIndex, totalPages);
	const hasPrev = currentPageIndex > 1;
	const hasNext = currentPageIndex < totalPages;
	const pageRows = useMemo(
		() =>
			visibleRows.slice((currentPageIndex - 1) * ROLE_PAGE_SIZE, currentPageIndex * ROLE_PAGE_SIZE),
		[currentPageIndex, visibleRows],
	);

	const handleSearchChange = (next: string) => {
		setSearch(next);
		setPageIndex(1);
	};

	const selectedRole = useMemo(
		() => allRoles.find((r) => r.id === selectedRoleId) ?? null,
		[allRoles, selectedRoleId],
	);

	const handleCreateRole = async () => {
		if (!canManage || createMutation.isPending) {
			return;
		}
		try {
			const created = await createMutation.mutateAsync("new role");
			toast.success(t("group_settings.roles.detail.created_toast", { name: created.name }));
			setSelectedRoleId(created.id);
		} catch (error) {
			toast.error(error instanceof ApiError ? error.message : t("common.error_generic"));
		}
	};

	const handleHideRolesChange = async (next: boolean) => {
		if (!canManage) {
			return;
		}
		try {
			await hideRolesMutation.mutateAsync(next);
		} catch (error) {
			toast.error(error instanceof ApiError ? error.message : t("common.error_generic"));
		}
	};

	const loading = rolesQuery.isPending && !rolesQuery.data;

	if (selectedRole) {
		return (
			<GroupSettingsRoleDetailSection
				groupId={groupId}
				roles={sidebarRoles}
				role={selectedRole}
				members={membersQuery.data?.members ?? []}
				canManage={canManage}
				canGrantAdministrator={canGrantAdministrator}
				hideRoles={hideRoles}
				onSelectRole={setSelectedRoleId}
				onBack={() => setSelectedRoleId(null)}
				onCreateRole={() => void handleCreateRole()}
			/>
		);
	}

	return (
		<div className="flex min-h-0 flex-1 flex-col">
			<div className={groupSettingsMainScrollClass}>
				<div
					className={cn(
						groupSettingsMainInnerClass,
						"flex max-w-none flex-col gap-6 pb-4 lg:max-w-5xl",
					)}
				>
					<div className={groupSettingsSectionHeaderClass}>
						<h2 className={groupSettingsSectionTitleClass}>{t("group_settings.roles.title")}</h2>
						<p className={groupSettingsSectionLeadClass}>{t("group_settings.roles.lead")}</p>
					</div>

					<GroupSettingsDefaultRoleCard
						role={defaultRole}
						hideRoles={hideRoles}
						canManage={canManage}
						onHideRolesChange={(next) => void handleHideRolesChange(next)}
						onOpenRole={setSelectedRoleId}
					/>

					<GroupSettingsRolesToolbar
						search={search}
						onSearchChange={handleSearchChange}
						onCreateRole={() => void handleCreateRole()}
						canManage={canManage}
					/>

					<GroupSettingsRolesTable
						groupId={groupId}
						rows={pageRows}
						loading={loading}
						canManage={canManage}
						total={roleTotal}
						pageIndex={currentPageIndex}
						hasPrev={hasPrev}
						hasNext={hasNext}
						maxVisitedPage={currentPageIndex}
						onOpenRole={setSelectedRoleId}
						onPrev={() => setPageIndex(Math.max(1, currentPageIndex - 1))}
						onNext={() => setPageIndex(Math.min(totalPages, currentPageIndex + 1))}
						onGoToPage={(page) => {
							if (page >= 1 && page <= totalPages) {
								setPageIndex(page);
							}
						}}
					/>
				</div>
			</div>
		</div>
	);
}
