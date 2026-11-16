"use client";

import { Trash2, type LucideIcon } from "lucide-react";
import type { ReactNode } from "react";
import {
	SIDEBAR_INSET_X,
	sidebarRowInnerButtonClass,
	sidebarRowShellClass,
	sidebarRowTextClass,
} from "@/components/layout/channel-sidebar/sidebar-nav-styles";
import { SidebarPlainNavRow } from "@/components/layout/channel-sidebar/sidebar-plain-nav-row";
import { SidebarStaticCategoryRow } from "@/components/layout/channel-sidebar/sidebar-static-category-row";
import { ChannelGlyphSlot, sidebarNavGlyphIconClass } from "@/components/layout/sidebar-list-icon";
import { cn } from "@/lib/utils";

export type FullscreenSettingsNavItem = {
	id: string;
	label: string;
	icon: LucideIcon;
	selected: boolean;
	onSelect: () => void;
	trailingIcon?: ReactNode;
};

export type FullscreenSettingsNavGroup = {
	id: string;
	label?: string;
	items: FullscreenSettingsNavItem[];
};

export type FullscreenSettingsDestructiveAction = {
	label: string;
	onClick: () => void;
};

type FullscreenSettingsSidebarProps = {
	headerTitle: string;
	navLabel: string;
	groups: FullscreenSettingsNavGroup[];
	destructiveAction?: FullscreenSettingsDestructiveAction | null;
};

/** Shared fullscreen settings sidebar matching the Group Settings layout, spacing, and row chrome. */
export function FullscreenSettingsSidebar({
	headerTitle,
	navLabel,
	groups,
	destructiveAction,
}: FullscreenSettingsSidebarProps) {
	return (
		<aside className="flex h-full min-size-full max-h-[100dvh] flex-col">
			<div className={cn("flex min-h-0 flex-1 flex-col pt-12 sm:pt-14", SIDEBAR_INSET_X)}>
				<div className="thin-scroll flex min-h-0 flex-1 flex-col overflow-y-auto pb-3">
					<p className="mb-2 truncate px-0.5 text-sm font-semibold text-sidebar-foreground">
						{headerTitle}
					</p>
					<nav className="flex flex-col" aria-label={navLabel}>
						{groups.map((group, index) => (
							<div
								key={group.id}
								className={cn(
									"group/cat",
									index > 0 && "mt-3 border-t border-sidebar-border/35 pt-3",
								)}
							>
								{group.label ? <SidebarStaticCategoryRow title={group.label} /> : null}
								<div className="flex flex-col gap-0">
									{group.items.map((item) => {
										const Icon = item.icon;
										return (
											<SidebarPlainNavRow
												key={item.id}
												icon={<Icon className={sidebarNavGlyphIconClass} strokeWidth={2} />}
												label={item.label}
												selected={item.selected}
												onSelect={item.onSelect}
												trailingIcon={item.trailingIcon}
											/>
										);
									})}
								</div>
							</div>
						))}

						{destructiveAction ? (
							<div className="mt-3 border-t border-sidebar-border/35 pt-3">
								<div
									className={cn(
										sidebarRowShellClass,
										"text-destructive hover:bg-destructive/10 dark:hover:bg-destructive/15",
									)}
								>
									<button
										type="button"
										className={cn(sidebarRowInnerButtonClass, "text-destructive")}
										onClick={destructiveAction.onClick}
									>
										<ChannelGlyphSlot framed={false} className="text-destructive">
											<Trash2 className={sidebarNavGlyphIconClass} strokeWidth={2} aria-hidden />
										</ChannelGlyphSlot>
										<span
											className={cn(
												sidebarRowTextClass,
												"min-w-0 flex-1 truncate text-left font-medium capitalize",
											)}
										>
											{destructiveAction.label}
										</span>
									</button>
								</div>
							</div>
						) : null}
					</nav>
				</div>
			</div>
		</aside>
	);
}
