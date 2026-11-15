"use client";

import type { GroupMember } from "@wichat/contracts/group/member";
import { Ellipsis } from "lucide-react";
import { useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { RoleDetailDisplayTab } from "@/features/group-settings/components/roles/detail/role-detail-display-tab";
import { RoleDetailMembersTab } from "@/features/group-settings/components/roles/detail/role-detail-members-tab";
import { RoleDetailPermissionsTab } from "@/features/group-settings/components/roles/detail/role-detail-permissions-tab";
import { RoleDetailSidebar } from "@/features/group-settings/components/roles/detail/role-detail-sidebar";
import {
	RoleActionMenu,
	type RoleActionMenuState,
} from "@/features/group-settings/components/roles/role-action-menu";
import {
	groupSettingsFormFooterFloatAlignClass,
	groupSettingsFormFooterFloatColumnClass,
	groupSettingsFormFooterFloatShellClass,
	groupSettingsMainInnerClass,
	groupSettingsMainScrollDirtyPadClass,
	groupSettingsRoleDetailPanePaddingClass,
	groupSettingsSectionTitleClass,
} from "@/features/group-settings/lib/group-settings-chrome";
import {
	DEFAULT_ROLE_COLOR,
	permissionsArrayToRecord,
	permissionsRecordToArray,
	roleDisplayTitle,
	type GroupSettingsRoleDetailTab,
	type GroupSettingsRoleRow,
} from "@/features/group-settings/lib/group-settings-roles-view";
import { ApiError } from "@wichat/api-client/api-error";
import { cn } from "@/lib/utils";
import { useUpdateGroupRoleMutation } from "@/services/group-role/group-role.mutation";

type GroupSettingsRoleDetailSectionProps = {
	groupId: string;
	roles: GroupSettingsRoleRow[];
	role: GroupSettingsRoleRow;
	members: GroupMember[];
	canManage: boolean;
	canGrantAdministrator: boolean;
	hideRoles?: boolean;
	onSelectRole: (roleId: string) => void;
	onBack: () => void;
	onCreateRole: () => void;
};

function sameStringSet(a: string[], b: string[]): boolean {
	if (a.length !== b.length) {
		return false;
	}
	const setA = new Set(a);
	if (setA.size !== b.length) {
		return false;
	}
	return b.every((item) => setA.has(item));
}

export function GroupSettingsRoleDetailSection({
	groupId,
	roles,
	role,
	members,
	canManage,
	canGrantAdministrator,
	hideRoles = false,
	onSelectRole,
	onBack,
	onCreateRole,
}: GroupSettingsRoleDetailSectionProps) {
	const { t } = useTranslation();
	const [activeTab, setActiveTab] = useState<GroupSettingsRoleDetailTab>("display");
	const [menuState, setMenuState] = useState<RoleActionMenuState | null>(null);
	const updateMutation = useUpdateGroupRoleMutation(groupId);

	const roleLabel = roleDisplayTitle(role, t);
	const isEveryone = role.isDefault || role.slug === "everyone";
	const displayDisabled = isEveryone;
	const currentTab: GroupSettingsRoleDetailTab = isEveryone ? "permissions" : activeTab;

	const [syncedRoleKey, setSyncedRoleKey] = useState(
		`${role.id}:${role.name}:${role.color}:${String(role.hoist)}:${String(role.mentionable)}:${role.permissions.join(",")}`,
	);
	const [roleName, setRoleName] = useState(roleLabel);
	const [selectedColor, setSelectedColor] = useState(role.color || DEFAULT_ROLE_COLOR);
	const [hoist, setHoist] = useState(role.hoist);
	const [mentionable, setMentionable] = useState(role.mentionable);
	const [permissions, setPermissions] = useState<Record<string, boolean>>(() =>
		permissionsArrayToRecord(role.permissions),
	);

	const nextRoleKey = `${role.id}:${role.name}:${role.color}:${String(role.hoist)}:${String(role.mentionable)}:${role.permissions.join(",")}`;
	if (nextRoleKey !== syncedRoleKey) {
		setSyncedRoleKey(nextRoleKey);
		setRoleName(roleLabel);
		setSelectedColor(role.color || DEFAULT_ROLE_COLOR);
		setHoist(role.hoist);
		setMentionable(role.mentionable);
		setPermissions(permissionsArrayToRecord(role.permissions));
	}

	const selectedPerms = useMemo(() => permissionsRecordToArray(permissions), [permissions]);
	const dirty = useMemo(() => {
		const permsChanged = !sameStringSet(selectedPerms, role.permissions);
		if (isEveryone) {
			return permsChanged;
		}
		return (
			permsChanged ||
			roleName.trim() !== roleLabel.trim() ||
			selectedColor.toLowerCase() !== (role.color || DEFAULT_ROLE_COLOR).toLowerCase() ||
			hoist !== role.hoist ||
			mentionable !== role.mentionable
		);
	}, [hoist, isEveryone, mentionable, role, roleLabel, roleName, selectedColor, selectedPerms]);

	const handleReset = () => {
		setRoleName(roleLabel);
		setSelectedColor(role.color || DEFAULT_ROLE_COLOR);
		setHoist(role.hoist);
		setMentionable(role.mentionable);
		setPermissions(permissionsArrayToRecord(role.permissions));
	};

	const handleSave = async () => {
		if (!canManage || !dirty) {
			return;
		}
		try {
			const updated = await updateMutation.mutateAsync({
				roleId: role.id,
				payload: isEveryone
					? { permissions: selectedPerms }
					: {
							name: roleName.trim() || role.name,
							color: selectedColor || DEFAULT_ROLE_COLOR,
							hoist,
							mentionable,
							permissions: selectedPerms,
						},
			});
			toast.success(
				t("group_settings.roles.detail.saved_toast", {
					name: updated.name,
				}),
			);
		} catch (error) {
			toast.error(error instanceof ApiError ? error.message : t("common.error_generic"));
		}
	};

	const tabs: Array<{ id: GroupSettingsRoleDetailTab; label: string; disabled?: boolean }> = [
		{
			id: "display",
			label: t("group_settings.roles.detail.tab_display"),
			disabled: displayDisabled,
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
		<div className="flex min-h-0 flex-1 flex-col lg:h-full lg:overflow-hidden">
			<div className="flex min-h-0 flex-1 flex-col md:flex-row lg:overflow-hidden">
				<RoleDetailSidebar
					groupId={groupId}
					roles={roles}
					selectedRoleId={role.id}
					canManage={canManage}
					onSelectRole={onSelectRole}
					onBack={onBack}
					onCreateRole={onCreateRole}
					onDeletedRole={(deletedId) => {
						if (deletedId === role.id) {
							onBack();
						}
					}}
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
									aria-expanded={menuState !== null}
									onClick={(event) => {
										setMenuState({
											role,
											anchor: event.currentTarget,
											align: "end",
										});
									}}
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

					{/* Scrollable Tab Content */}
					<div
						className={cn(
							"overlay-scroll flex min-h-0 flex-1 flex-col pb-10 lg:overflow-y-auto lg:overscroll-contain",
							groupSettingsRoleDetailPanePaddingClass,
							dirty && canManage && groupSettingsMainScrollDirtyPadClass,
						)}
					>
						<div className="mx-auto flex w-full max-w-2xl flex-col lg:mx-0">
							{currentTab === "display" ? (
								<RoleDetailDisplayTab
									role={role}
									canManage={canManage}
									roleName={roleName}
									onRoleNameChange={setRoleName}
									selectedColor={selectedColor}
									onSelectedColorChange={setSelectedColor}
									hoist={hoist}
									onHoistChange={setHoist}
									mentionable={mentionable}
									onMentionableChange={setMentionable}
								/>
							) : null}

							{currentTab === "permissions" ? (
								<RoleDetailPermissionsTab
									canManage={canManage}
									canGrantAdministrator={canGrantAdministrator}
									permissions={permissions}
									onPermissionsChange={setPermissions}
								/>
							) : null}

							{currentTab === "members" ? (
								<RoleDetailMembersTab
									groupId={groupId}
									role={role}
									allMembers={members}
									canManage={canManage}
									hideRoles={hideRoles}
								/>
							) : null}
						</div>
					</div>
				</div>
			</div>

			{dirty && canManage ? (
				<div className={groupSettingsFormFooterFloatShellClass}>
					<div className={groupSettingsFormFooterFloatAlignClass}>
						<div className={groupSettingsFormFooterFloatColumnClass}>
							<div
								className={cn(
									groupSettingsMainInnerClass,
									"flex items-center justify-end gap-2.5 rounded-xl border border-border bg-card/95 px-4 py-3 shadow-elevated backdrop-blur-md",
								)}
							>
								<Button
									type="button"
									variant="ghost"
									disabled={updateMutation.isPending}
									onClick={handleReset}
								>
									{t("group_settings.roles.detail.reset_changes")}
								</Button>
								<Button
									type="button"
									disabled={updateMutation.isPending || (!isEveryone && !roleName.trim())}
									onClick={() => void handleSave()}
								>
									{updateMutation.isPending
										? t("group_settings.roles.detail.saving_changes")
										: t("group_settings.roles.detail.save_changes")}
								</Button>
							</div>
						</div>
					</div>
				</div>
			) : null}

			<RoleActionMenu
				groupId={groupId}
				menuState={menuState}
				onCloseMenu={() => setMenuState(null)}
				canManage={canManage}
				onDuplicated={onSelectRole}
				onDeleted={(deletedId) => {
					if (deletedId === role.id) {
						onBack();
					}
				}}
			/>
		</div>
	);
}
