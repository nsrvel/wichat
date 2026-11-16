"use client";

import { useQuery } from "@tanstack/react-query";
import type {
	CategoryPermissionTargetType,
	ChannelCategoryPermissionOverride,
} from "@wichat/contracts/chat/channel-category";
import { Check, Lock, Search, Slash, Users, X } from "lucide-react";
import { useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { UserAvatar } from "@/components/ui/user-avatar";
import { CategoryPermissionTargetPopover } from "@/features/channel-category/category-settings/category-permission-target-popover";
import type { CategorySettingsTarget } from "@/features/channel-category/category-settings/category-settings-sidebar";
import {
	groupSettingsFormFooterFloatAlignClass,
	groupSettingsFormFooterFloatColumnClass,
	groupSettingsFormFooterFloatShellClass,
	groupSettingsMainInnerClass,
	groupSettingsMainScrollClass,
	groupSettingsMainScrollDirtyPadClass,
	groupSettingsSectionHeaderClass,
	groupSettingsSectionLeadClass,
	groupSettingsSectionTitleClass,
	groupSettingsToolbarSearchInputClass,
} from "@/features/group-settings/lib/group-settings-chrome";
import { ROLE_PERMISSION_GROUPS } from "@/features/group-settings/lib/group-settings-roles-view";
import {
	useGroupSettingsUnsaved,
	useReportGroupSettingsUnsaved,
} from "@/features/group-settings/lib/group-settings-unsaved-guard";
import { translateError } from "@/integrations/i18n/translate-error";
import { authInputClass } from "@/lib/form-field-styles";
import { hasCustomRoleColor } from "@/lib/member-role-display";
import { settingsFormFooterBarClass, settingsFormFooterBarNudgeClass } from "@/lib/settings-chrome";
import { userInitials } from "@/lib/user-initials";
import { cn } from "@/lib/utils";
import {
	useDeleteCategoryPermissionMutation,
	useUpdateGroupCategoryMutation,
	useUpsertCategoryPermissionMutation,
} from "@/services/channel-category/channel-category.mutation";
import { groupChannelsQueryOpts } from "@/services/channel/channel.query";
import { categoryPermissionsQueryOpts } from "@/services/channel-category/channel-category.query";
import { groupRolesQueryOpts } from "@/services/group-role/group-role.query";
import { groupMembersQueryOpts } from "@/services/invite/invite.query";

type TriState = "deny" | "inherit" | "allow";

function targetKey(targetType: CategoryPermissionTargetType, targetId: string) {
	return `${targetType}:${targetId}`;
}

function getTriState(allow: readonly string[], deny: readonly string[], permId: string): TriState {
	if (allow.includes(permId)) {
		return "allow";
	}
	if (deny.includes(permId)) {
		return "deny";
	}
	return "inherit";
}

function withTriState(
	override: ChannelCategoryPermissionOverride,
	permId: string,
	next: TriState,
): ChannelCategoryPermissionOverride {
	const allow = override.allow_permissions.filter((p) => p !== permId);
	const deny = override.deny_permissions.filter((p) => p !== permId);
	if (next === "allow") {
		allow.push(permId);
	} else if (next === "deny") {
		deny.push(permId);
	}
	return {
		...override,
		allow_permissions: allow,
		deny_permissions: deny,
	};
}

function normalizeSorted(list: readonly string[]) {
	return [...list].sort().join(",");
}

function TargetOverrideRow({
	item,
	effectiveItem,
	selected,
	defaultBadgeLabel,
	removeLabel,
	onSelect,
	onRemove,
}: {
	item: ChannelCategoryPermissionOverride;
	effectiveItem: ChannelCategoryPermissionOverride;
	selected: boolean;
	defaultBadgeLabel: string;
	removeLabel: string;
	onSelect: () => void;
	onRemove?: () => void;
}) {
	const displayName = item.is_everyone ? "@everyone" : item.target_name;
	const allowCount = effectiveItem.allow_permissions.length;
	const denyCount = effectiveItem.deny_permissions.length;
	const hasRoleColor =
		item.target_type === "role" && !item.is_everyone && hasCustomRoleColor(item.target_color);

	return (
		<div
			className={cn(
				"group/target flex h-8 items-center gap-2 rounded-lg px-2 text-xs transition-colors",
				selected
					? "bg-accent text-foreground font-medium shadow-xs ring-1 ring-border/70"
					: "text-muted-foreground hover:bg-accent/50 hover:text-foreground",
			)}
		>
			<button
				type="button"
				onClick={onSelect}
				className="flex min-w-0 flex-1 items-center gap-2 text-left"
			>
				{item.is_everyone ? (
					<span className="flex size-5 shrink-0 items-center justify-center rounded-full bg-muted text-muted-foreground">
						<Users className="size-3" aria-hidden />
					</span>
				) : item.target_type === "role" ? (
					<span className="flex size-5 shrink-0 items-center justify-center">
						<span
							className="size-2.5 rounded-full bg-muted-foreground/60"
							style={hasRoleColor ? { backgroundColor: item.target_color } : undefined}
						/>
					</span>
				) : (
					<UserAvatar
						src={item.target_avatar_url}
						initials={userInitials(displayName)}
						size="xs"
						className="size-5 text-[10px]"
					/>
				)}

				<span className="min-w-0 flex-1 truncate">{displayName}</span>

				{item.is_everyone && allowCount === 0 && denyCount === 0 ? (
					<span className="shrink-0 rounded bg-muted px-1.5 py-0.5 text-[10px] font-medium text-muted-foreground">
						{defaultBadgeLabel}
					</span>
				) : allowCount > 0 || denyCount > 0 ? (
					<span className="flex shrink-0 items-center gap-1 text-[10px] font-medium">
						{allowCount > 0 ? (
							<span className="rounded bg-emerald-500/15 px-1 py-0.5 text-emerald-500">
								+{allowCount}
							</span>
						) : null}
						{denyCount > 0 ? (
							<span className="rounded bg-destructive/15 px-1 py-0.5 text-destructive">
								-{denyCount}
							</span>
						) : null}
					</span>
				) : null}
			</button>

			{onRemove ? (
				<button
					type="button"
					onClick={onRemove}
					aria-label={removeLabel}
					className={cn(
						"rounded p-0.5 text-muted-foreground transition-opacity hover:text-destructive",
						selected ? "opacity-70 hover:opacity-100" : "opacity-0 group-hover/target:opacity-100",
					)}
				>
					<X className="size-3.5" aria-hidden />
				</button>
			) : null}
		</div>
	);
}

export function CategorySettingsPermissionsSection({ target }: { target: CategorySettingsTarget }) {
	const { t } = useTranslation();
	const { footerNudgeKey } = useGroupSettingsUnsaved();
	const updateCategoryMutation = useUpdateGroupCategoryMutation(target.groupId);
	const upsertPermMutation = useUpsertCategoryPermissionMutation(target.groupId, target.categoryId);
	const deletePermMutation = useDeleteCategoryPermissionMutation(target.groupId, target.categoryId);

	const channelsQuery = useQuery(groupChannelsQueryOpts(target.groupId));
	const permissionsQuery = useQuery(
		categoryPermissionsQueryOpts(target.groupId, target.categoryId),
	);
	const rolesQuery = useQuery(groupRolesQueryOpts(target.groupId));
	const membersQuery = useQuery({
		...groupMembersQueryOpts(target.groupId, ""),
		enabled: Boolean(target.groupId),
	});

	const liveCategory = channelsQuery.data?.categories.find((c) => c.id === target.categoryId);
	const isPrivate = liveCategory?.is_private ?? target.isPrivate ?? false;

	const [searchQuery, setSearchQuery] = useState("");
	const [selectedKey, setSelectedKey] = useState<string | null>(null);
	const [addedTargets, setAddedTargets] = useState<ChannelCategoryPermissionOverride[]>([]);
	const [draftMap, setDraftMap] = useState<Record<string, ChannelCategoryPermissionOverride>>({});

	const everyoneRole = useMemo(
		() => (rolesQuery.data?.roles ?? []).find((r) => r.is_default || r.slug === "everyone"),
		[rolesQuery.data?.roles],
	);

	const baseTargets = useMemo(() => {
		const saved = permissionsQuery.data?.overrides ?? [];
		const map = new Map<string, ChannelCategoryPermissionOverride>();

		if (everyoneRole) {
			map.set(targetKey("role", everyoneRole.id), {
				target_type: "role",
				target_id: everyoneRole.id,
				target_name: everyoneRole.name,
				target_color: everyoneRole.color,
				target_avatar_url: "",
				is_everyone: true,
				allow_permissions: [],
				deny_permissions: [],
			});
		}

		for (const item of saved) {
			map.set(targetKey(item.target_type, item.target_id), item);
		}
		for (const item of addedTargets) {
			const k = targetKey(item.target_type, item.target_id);
			if (!map.has(k)) {
				map.set(k, item);
			}
		}
		return Array.from(map.values());
	}, [addedTargets, everyoneRole, permissionsQuery.data?.overrides]);

	const groupedTargets = useMemo(() => {
		const everyone = baseTargets.filter((item) => item.is_everyone);
		const roles = baseTargets.filter((item) => item.target_type === "role" && !item.is_everyone);
		const members = baseTargets.filter((item) => item.target_type === "member");
		return { everyone, roles, members };
	}, [baseTargets]);

	const activeKey =
		selectedKey &&
		baseTargets.some((tItem) => targetKey(tItem.target_type, tItem.target_id) === selectedKey)
			? selectedKey
			: baseTargets[0]
				? targetKey(baseTargets[0].target_type, baseTargets[0].target_id)
				: null;

	const activeOverride = useMemo(() => {
		if (!activeKey) {
			return null;
		}
		if (draftMap[activeKey]) {
			return draftMap[activeKey];
		}
		return (
			baseTargets.find((item) => targetKey(item.target_type, item.target_id) === activeKey) ?? null
		);
	}, [activeKey, baseTargets, draftMap]);

	const activeSaved = useMemo(() => {
		if (!activeKey) {
			return null;
		}
		return (
			baseTargets.find((item) => targetKey(item.target_type, item.target_id) === activeKey) ?? null
		);
	}, [activeKey, baseTargets]);

	const isDirty = useMemo(() => {
		if (!activeOverride || !activeSaved) {
			return false;
		}
		return (
			normalizeSorted(activeOverride.allow_permissions) !==
				normalizeSorted(activeSaved.allow_permissions) ||
			normalizeSorted(activeOverride.deny_permissions) !==
				normalizeSorted(activeSaved.deny_permissions)
		);
	}, [activeOverride, activeSaved]);

	useReportGroupSettingsUnsaved(isDirty);

	const availableRolesToAdd = useMemo(() => {
		const existing = new Set(
			baseTargets.map((item) => targetKey(item.target_type, item.target_id)),
		);
		return (rolesQuery.data?.roles ?? []).filter((r) => !existing.has(targetKey("role", r.id)));
	}, [baseTargets, rolesQuery.data?.roles]);

	const availableMembersToAdd = useMemo(() => {
		const existing = new Set(
			baseTargets.map((item) => targetKey(item.target_type, item.target_id)),
		);
		return (membersQuery.data?.members ?? []).filter(
			(m) => !existing.has(targetKey("member", m.user_id)),
		);
	}, [baseTargets, membersQuery.data?.members]);

	const filteredGroups = useMemo(() => {
		const q = searchQuery.trim().toLowerCase();
		if (!q) {
			return ROLE_PERMISSION_GROUPS;
		}
		return ROLE_PERMISSION_GROUPS.map((group) => ({
			...group,
			items: group.items.filter((item) => {
				const title = t(item.titleKey).toLowerCase();
				const desc = t(item.descriptionKey).toLowerCase();
				return title.includes(q) || desc.includes(q);
			}),
		})).filter((group) => group.items.length > 0);
	}, [searchQuery, t]);

	const handleTogglePrivate = (nextValue: boolean) => {
		updateCategoryMutation.mutate(
			{
				categoryId: target.categoryId,
				payload: { is_private: nextValue },
			},
			{
				onSuccess: () => {
					toast.success(t("channel_category.settings.updated_toast"));
				},
				onError: (err) => {
					toast.error(translateError(t, err));
				},
			},
		);
	};

	const handleSetTriState = (permId: string, nextState: TriState) => {
		if (!activeKey || !activeOverride) {
			return;
		}
		setDraftMap((prev) => ({
			...prev,
			[activeKey]: withTriState(activeOverride, permId, nextState),
		}));
	};

	const handleSave = () => {
		if (!activeKey || !activeOverride) {
			return;
		}
		upsertPermMutation.mutate(
			{
				target_type: activeOverride.target_type,
				target_id: activeOverride.target_id,
				allow_permissions: activeOverride.allow_permissions,
				deny_permissions: activeOverride.deny_permissions,
			},
			{
				onSuccess: () => {
					setDraftMap((prev) => {
						const next = { ...prev };
						delete next[activeKey];
						return next;
					});
					toast.success(t("channel_category.settings.updated_toast"));
				},
				onError: (err) => {
					toast.error(translateError(t, err));
				},
			},
		);
	};

	const handleReset = () => {
		if (!activeKey) {
			return;
		}
		setDraftMap((prev) => {
			const next = { ...prev };
			delete next[activeKey];
			return next;
		});
	};

	const handleRemoveOverride = (item: ChannelCategoryPermissionOverride) => {
		const key = targetKey(item.target_type, item.target_id);
		const isSaved = (permissionsQuery.data?.overrides ?? []).some(
			(o) => targetKey(o.target_type, o.target_id) === key,
		);
		if (!isSaved) {
			setAddedTargets((prev) =>
				prev.filter((entry) => targetKey(entry.target_type, entry.target_id) !== key),
			);
			setDraftMap((prev) => {
				const next = { ...prev };
				delete next[key];
				return next;
			});
			if (selectedKey === key) {
				setSelectedKey(null);
			}
			return;
		}
		deletePermMutation.mutate(
			{ targetType: item.target_type, targetId: item.target_id },
			{
				onSuccess: () => {
					setDraftMap((prev) => {
						const next = { ...prev };
						delete next[key];
						return next;
					});
					if (selectedKey === key) {
						setSelectedKey(null);
					}
					toast.success(t("channel_category.settings.updated_toast"));
				},
				onError: (err) => {
					toast.error(translateError(t, err));
				},
			},
		);
	};

	const renderTargetItem = (item: ChannelCategoryPermissionOverride) => {
		const key = targetKey(item.target_type, item.target_id);
		return (
			<TargetOverrideRow
				key={key}
				item={item}
				effectiveItem={draftMap[key] ?? item}
				selected={key === activeKey}
				defaultBadgeLabel={t("channel_category.settings.default_role_badge")}
				removeLabel={t("channel_category.settings.remove_override")}
				onSelect={() => setSelectedKey(key)}
				onRemove={!item.is_everyone ? () => handleRemoveOverride(item) : undefined}
			/>
		);
	};

	return (
		<div className="flex min-h-0 flex-1 flex-col">
			<div
				className={cn(
					groupSettingsMainScrollClass,
					isDirty && groupSettingsMainScrollDirtyPadClass,
				)}
			>
				<div className={groupSettingsMainInnerClass}>
					<header className={groupSettingsSectionHeaderClass}>
						<h2 className={groupSettingsSectionTitleClass}>
							{t("channel_category.settings.permissions_title")}
						</h2>
						<p className={groupSettingsSectionLeadClass}>
							{t("channel_category.settings.permissions_subtitle")}
						</p>
					</header>

					<div className="space-y-6">
						<div className="flex items-start justify-between gap-4 rounded-xl border border-border bg-card/60 p-4">
							<div className="flex items-start gap-3">
								<Lock className="mt-0.5 size-4 shrink-0 text-muted-foreground" aria-hidden />
								<div className="space-y-1">
									<p className="text-sm font-medium text-foreground">
										{t("channel_category.settings.private_category_title")}
									</p>
									<p className="text-xs leading-relaxed text-muted-foreground">
										{t("channel_category.settings.private_category_description")}
									</p>
								</div>
							</div>
							<Switch
								checked={isPrivate}
								disabled={updateCategoryMutation.isPending}
								onCheckedChange={handleTogglePrivate}
								aria-label={t("channel_category.settings.private_category_title")}
							/>
						</div>

						<div className="grid gap-6 border-t border-border pt-6 md:grid-cols-[240px_minmax(0,1fr)] md:items-start">
							<div>
								<div className="mb-2">
									<CategoryPermissionTargetPopover
										roles={availableRolesToAdd}
										members={availableMembersToAdd}
										onSelectTarget={(nextItem) => {
											setAddedTargets((prev) => [...prev, nextItem]);
											setSelectedKey(targetKey(nextItem.target_type, nextItem.target_id));
										}}
									/>
								</div>

								<div className="space-y-3">
									{groupedTargets.everyone.length > 0 ? (
										<div className="space-y-0.5">
											{groupedTargets.everyone.map(renderTargetItem)}
										</div>
									) : null}

									{groupedTargets.roles.length > 0 ? (
										<div className="space-y-1 border-t border-border/60 pt-2.5">
											<p className="px-2 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
												{t("channel_category.settings.roles_heading")}
											</p>
											<div className="space-y-0.5">
												{groupedTargets.roles.map(renderTargetItem)}
											</div>
										</div>
									) : null}

									{groupedTargets.members.length > 0 ? (
										<div className="space-y-1 border-t border-border/60 pt-2.5">
											<p className="px-2 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
												{t("channel_category.settings.members_heading")}
											</p>
											<div className="space-y-0.5">
												{groupedTargets.members.map(renderTargetItem)}
											</div>
										</div>
									) : null}
								</div>
							</div>

							<div className="space-y-6">
								<label className="relative block min-w-0">
									<Search
										className="pointer-events-none absolute top-1/2 left-3.5 z-[1] size-4 -translate-y-1/2 text-muted-foreground"
										aria-hidden
									/>
									<Input
										value={searchQuery}
										onChange={(e) => setSearchQuery(e.target.value)}
										placeholder={t("channel_category.settings.search_permissions_placeholder")}
										className={cn(authInputClass, groupSettingsToolbarSearchInputClass)}
										autoComplete="off"
									/>
								</label>

								{filteredGroups.length === 0 ? (
									<p className="py-6 text-center text-sm text-muted-foreground">
										{t("channel_category.settings.no_permissions_match")}
									</p>
								) : (
									<div className="space-y-6">
										{filteredGroups.map((group) => (
											<div key={group.id} className="space-y-3">
												<h3 className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
													{t(group.titleKey)}
												</h3>
												<div className="divide-y divide-border/60">
													{group.items.map((perm) => {
														const currentState = activeOverride
															? getTriState(
																	activeOverride.allow_permissions,
																	activeOverride.deny_permissions,
																	perm.id,
																)
															: "inherit";
														return (
															<div
																key={perm.id}
																className="flex items-start justify-between gap-4 py-3 first:pt-1"
															>
																<div className="min-w-0 space-y-0.5">
																	<p className="text-sm font-medium text-foreground">
																		{t(perm.titleKey)}
																	</p>
																	<p className="text-xs leading-relaxed text-muted-foreground">
																		{t(perm.descriptionKey)}
																	</p>
																</div>

																<div
																	role="group"
																	aria-label={t(perm.titleKey)}
																	className="inline-flex shrink-0 items-center gap-0.5"
																>
																	<button
																		type="button"
																		onClick={() => handleSetTriState(perm.id, "deny")}
																		aria-label={t("channel_category.settings.perm_deny")}
																		aria-pressed={currentState === "deny"}
																		className={cn(
																			"flex h-7 w-7 items-center justify-center rounded-md transition-colors",
																			currentState === "deny"
																				? "bg-destructive/20 text-destructive"
																				: "text-destructive/70 hover:bg-destructive/10 hover:text-destructive",
																		)}
																	>
																		<X className="size-3.5" strokeWidth={2.5} />
																	</button>
																	<button
																		type="button"
																		onClick={() => handleSetTriState(perm.id, "inherit")}
																		aria-label={t("channel_category.settings.perm_inherit")}
																		aria-pressed={currentState === "inherit"}
																		className={cn(
																			"flex h-7 w-7 items-center justify-center rounded-md transition-colors",
																			currentState === "inherit"
																				? "bg-accent text-foreground"
																				: "text-muted-foreground hover:bg-accent/50",
																		)}
																	>
																		<Slash className="size-3" strokeWidth={2.25} />
																	</button>
																	<button
																		type="button"
																		onClick={() => handleSetTriState(perm.id, "allow")}
																		aria-label={t("channel_category.settings.perm_allow")}
																		aria-pressed={currentState === "allow"}
																		className={cn(
																			"flex h-7 w-7 items-center justify-center rounded-md transition-colors",
																			currentState === "allow"
																				? "bg-emerald-500/20 text-emerald-500"
																				: "text-emerald-500/70 hover:bg-emerald-500/10 hover:text-emerald-500",
																		)}
																	>
																		<Check className="size-3.5" strokeWidth={2.5} />
																	</button>
																</div>
															</div>
														);
													})}
												</div>
											</div>
										))}
									</div>
								)}
							</div>
						</div>
					</div>
				</div>
			</div>

			{isDirty ? (
				<div className={groupSettingsFormFooterFloatShellClass}>
					<div className={groupSettingsFormFooterFloatAlignClass}>
						<div className={groupSettingsFormFooterFloatColumnClass}>
							<div
								className={cn(
									settingsFormFooterBarClass,
									footerNudgeKey > 0 && settingsFormFooterBarNudgeClass,
								)}
							>
								<p className="flex min-w-0 flex-1 items-center text-sm text-muted-foreground">
									{footerNudgeKey > 0 ? t("group_settings.unsaved_block") : "\u00a0"}
								</p>
								<div className="flex shrink-0 items-center gap-2">
									<Button
										type="button"
										variant="ghost"
										size="sm"
										disabled={upsertPermMutation.isPending}
										onClick={handleReset}
									>
										{t("group_settings.reset")}
									</Button>
									<Button
										type="button"
										size="sm"
										disabled={upsertPermMutation.isPending}
										onClick={handleSave}
									>
										{t("group_settings.save_changes")}
									</Button>
								</div>
							</div>
						</div>
					</div>
				</div>
			) : null}
		</div>
	);
}
