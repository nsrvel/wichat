"use client";

import { LogOut, Search, X } from "lucide-react";
import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import {
	SIDEBAR_INSET_X,
	sidebarRowInnerButtonClass,
	sidebarRowShellClass,
	sidebarRowTextClass,
} from "@/components/layout/channel-sidebar/sidebar-nav-styles";
import { SidebarPlainNavRow } from "@/components/layout/channel-sidebar/sidebar-plain-nav-row";
import { SidebarStaticCategoryRow } from "@/components/layout/channel-sidebar/sidebar-static-category-row";
import { ChannelGlyphSlot, sidebarNavGlyphIconClass } from "@/components/layout/sidebar-list-icon";
import { UserProfileButton } from "@/components/layout/user-profile-button";
import { Input } from "@/components/ui/input";
import {
	USER_SETTINGS_NAV_GROUPS,
	type UserSettingsSectionId,
} from "@/features/user-settings/lib/user-settings-nav";
import type { UserPresenceStatus } from "@/lib/user-status";
import { cn } from "@/lib/utils";

type UserSettingsSidebarProps = {
	displayName: string;
	secondaryText: string;
	avatarUrl: string;
	initials: string;
	status: UserPresenceStatus;
	activeSection: UserSettingsSectionId;
	activeSubAnchor: string | null;
	searchQuery: string;
	onSearchChange: (value: string) => void;
	onSelectSection: (section: UserSettingsSectionId, anchorId?: string) => void;
	onSignOut: () => void;
	signingOut?: boolean;
};

export function UserSettingsSidebar({
	displayName,
	secondaryText,
	avatarUrl,
	initials,
	status,
	activeSection,
	activeSubAnchor,
	searchQuery,
	onSearchChange,
	onSelectSection,
	onSignOut,
	signingOut,
}: UserSettingsSidebarProps) {
	const { t } = useTranslation();

	const trimmedQuery = searchQuery.trim().toLowerCase();

	const filteredGroups = useMemo(() => {
		if (!trimmedQuery) {
			return USER_SETTINGS_NAV_GROUPS.map((group) => ({
				...group,
				items: group.items.map((item) => ({
					...item,
					matchedFields: [] as NonNullable<typeof item.fields>,
				})),
			}));
		}
		return USER_SETTINGS_NAV_GROUPS.map((group) => ({
			...group,
			items: group.items
				.map((item) => {
					const label = t(item.labelKey).toLowerCase();
					const matchedFields = (item.fields ?? []).filter((field) => {
						const fieldLabel = t(field.labelKey).toLowerCase();
						return (
							fieldLabel.includes(trimmedQuery) ||
							(field.keywords ?? []).some((kw) => kw.includes(trimmedQuery))
						);
					});
					const matchesItem =
						label.includes(trimmedQuery) ||
						item.keywords.some((kw) => kw.includes(trimmedQuery)) ||
						item.subAnchors?.some((sub) => t(sub.labelKey).toLowerCase().includes(trimmedQuery)) ||
						matchedFields.length > 0;
					return { ...item, matchedFields, matchesItem };
				})
				.filter((item) => item.matchesItem),
		})).filter((group) => group.items.length > 0);
	}, [trimmedQuery, t]);

	return (
		<aside className="flex h-full w-[15rem] shrink-0 flex-col border-r border-sidebar-border/60 bg-sidebar text-sidebar-foreground">
			{/* Same profile button as ShellUserDock + emoji-style search bar */}
			<div className={cn("shrink-0 pt-3.5 pb-2", SIDEBAR_INSET_X)}>
				<div className="flex">
					<UserProfileButton
						displayName={displayName}
						secondaryText={secondaryText}
						avatarUrl={avatarUrl}
						initials={initials}
						status={status}
						statusRingClassName="ring-sidebar"
						aria-label={t("user_settings.edit_profiles")}
						onClick={() => onSelectSection("profiles")}
					/>
				</div>

				<div className="relative mt-2.5 flex items-center">
					<Input
						type="text"
						value={searchQuery}
						onChange={(e) => onSearchChange(e.target.value)}
						placeholder={t("user_settings.search_placeholder")}
						aria-label={t("user_settings.search_placeholder")}
						className="h-9 w-full rounded-md border-0 bg-[#f5f5f6] pl-3.5 pr-9 text-[13px] text-foreground shadow-none placeholder:text-muted-foreground focus-visible:border-0 focus-visible:bg-[#f5f5f6] focus-visible:ring-0 dark:bg-[#242424] dark:focus-visible:bg-[#242424]"
					/>
					{searchQuery ? (
						<button
							type="button"
							onClick={() => onSearchChange("")}
							aria-label={t("expression.clear_search")}
							className="absolute right-2.5 flex size-5 items-center justify-center rounded text-muted-foreground hover:text-foreground"
						>
							<X className="size-3.5" aria-hidden />
						</button>
					) : (
						<Search
							className="pointer-events-none absolute right-3 size-4 text-muted-foreground"
							aria-hidden
						/>
					)}
				</div>
			</div>

			{/* Scrollbar sits flush at the sidebar right edge; SIDEBAR_INSET_X is on inner nav */}
			<div className="thin-scroll flex min-h-0 flex-1 flex-col overflow-y-auto pt-3 pb-3">
				<nav
					className={cn("flex flex-col", SIDEBAR_INSET_X)}
					aria-label={t("user_settings.title")}
				>
					{filteredGroups.map((group, index) => (
						<div
							key={group.id}
							className={cn(
								"group/cat",
								index > 0 && "mt-3 border-t border-sidebar-border/35 pt-3",
							)}
						>
							<SidebarStaticCategoryRow title={t(group.labelKey)} />
							<div className="flex flex-col gap-0">
								{group.items.map((item) => {
									const Icon = item.icon;
									const showSubAnchors = item.id === activeSection && Boolean(item.subAnchors);
									const isParentSelected = item.id === activeSection && !activeSubAnchor;
									const searchTargetId =
										trimmedQuery.length > 0 ? item.matchedFields[0]?.elementId : undefined;
									return (
										<div key={item.id} className="flex flex-col gap-0">
											<SidebarPlainNavRow
												icon={<Icon className={sidebarNavGlyphIconClass} strokeWidth={2} />}
												label={t(item.labelKey)}
												selected={isParentSelected}
												onSelect={() => onSelectSection(item.id, searchTargetId)}
											/>
											{showSubAnchors ? (
												<div className="flex flex-col gap-0">
													{item.subAnchors!.map((sub) => {
														const subSelected = activeSubAnchor === sub.id;
														return (
															<div
																key={sub.id}
																className={cn(
																	sidebarRowShellClass,
																	subSelected && "bg-sidebar-row-selected",
																)}
																data-selected={subSelected ? "true" : undefined}
															>
																<button
																	type="button"
																	onClick={() => onSelectSection(item.id, sub.id)}
																	aria-current={subSelected ? "true" : undefined}
																	className={cn(
																		sidebarRowInnerButtonClass,
																		subSelected
																			? "font-semibold text-sidebar-foreground"
																			: "font-medium text-muted-foreground",
																	)}
																>
																	<ChannelGlyphSlot framed={false} aria-hidden>
																		{null}
																	</ChannelGlyphSlot>
																	<span
																		className={cn(
																			sidebarRowTextClass,
																			"min-w-0 flex-1 truncate text-left",
																		)}
																	>
																		{t(sub.labelKey)}
																	</span>
																</button>
															</div>
														);
													})}
												</div>
											) : null}
										</div>
									);
								})}
							</div>
						</div>
					))}

					{/* Sign out row matching GroupSettings destructive row */}
					<div className="mt-3 border-t border-sidebar-border/35 pt-3">
						<div
							className={cn(
								sidebarRowShellClass,
								"text-destructive hover:bg-destructive/10 dark:hover:bg-destructive/15",
							)}
						>
							<button
								type="button"
								disabled={signingOut}
								onClick={onSignOut}
								className={cn(
									sidebarRowInnerButtonClass,
									"text-destructive disabled:opacity-50",
								)}
							>
								<ChannelGlyphSlot framed={false} className="text-destructive">
									<LogOut className={sidebarNavGlyphIconClass} strokeWidth={2} aria-hidden />
								</ChannelGlyphSlot>
								<span
									className={cn(
										sidebarRowTextClass,
										"min-w-0 flex-1 truncate text-left font-medium",
									)}
								>
									{t("user_settings.sign_out")}
								</span>
							</button>
						</div>
					</div>
				</nav>
			</div>
		</aside>
	);
}
