"use client";

import { Menu, MoreHorizontal, PanelRight, Phone, Plus, Search, UserPlus } from "lucide-react";
import { useState } from "react";
import { ConversationNavIcon } from "@/components/layout/channel-nav-icon";
import {
	ConversationTabActionRow,
	ConversationTabNav,
	ConversationTabRow,
} from "@/components/layout/conversation-tab-row";
import {
	WorkspaceHeaderIconButton,
	workspaceHeaderIconClass,
} from "@/components/layout/workspace-header-icon-button";
import { WorkspaceSplitIconControl } from "@/components/layout/workspace-split-icon-control";
import type { ShellNavTarget } from "@/components/layout/shell-navigation";
import { cn } from "@/lib/utils";

const CONVERSATION_TABS = ["Chat", "Pinned", "Shared", "Notes"] as const;
type ConversationTab = (typeof CONVERSATION_TABS)[number];

function isSelfDm(active: ShellNavTarget) {
	return active.id === "fav-self-dm";
}

type ConversationHeaderProps = {
	active: ShellNavTarget;
	onOpenSidebar?: () => void;
	/** Desktop paired row draws one divider on the parent; mobile keeps it here. */
	showBottomDivider?: boolean;
};

export function ConversationHeader({
	active,
	onOpenSidebar,
	showBottomDivider = true,
}: ConversationHeaderProps) {
	const [tab, setTab] = useState<ConversationTab>("Chat");
	const selfDm = isSelfDm(active);
	return (
		<div
			className={cn(
				"flex h-14 shrink-0 items-center gap-2 bg-chat-surface px-3 lg:h-[46px] lg:gap-3 lg:px-4",
				showBottomDivider && "border-b border-black",
			)}
		>
			<button
				type="button"
				className="lg:hidden -ml-1 flex h-11 w-11 shrink-0 items-center justify-center rounded-lg text-muted-foreground hover:bg-sidebar-accent hover:text-sidebar-foreground"
				onClick={onOpenSidebar}
				aria-label="Open channels"
			>
				<Menu className="h-5 w-5" strokeWidth={1.8} />
			</button>

			<div className="flex min-w-0 shrink-0 items-center gap-2.5 sm:max-w-[min(240px,38%)]">
				<span aria-hidden>
					<ConversationNavIcon active={active} variant="header" />
				</span>
				<span
					className={cn(
						"truncate text-[15px] font-semibold leading-tight tracking-tight text-sidebar-foreground lg:text-base",
						active.kind === "dm" && "uppercase tracking-wide",
					)}
				>
					{active.label}
				</span>
			</div>

			<ConversationTabNav>
				{CONVERSATION_TABS.map((label) => {
					const selected = tab === label;
					return (
						<ConversationTabRow key={label} selected={selected}>
							<button
								type="button"
								onClick={() => setTab(label)}
								className={cn(
									"shrink-0 text-xs font-medium transition-colors lg:text-[13px]",
									selected
										? "text-sidebar-foreground"
										: "text-muted-foreground hover:text-sidebar-foreground",
								)}
								aria-current={selected ? "page" : undefined}
							>
								{label}
							</button>
						</ConversationTabRow>
					);
				})}
				<ConversationTabActionRow>
					<WorkspaceHeaderIconButton className="shrink-0" aria-label="Add tab">
						<Plus className="h-4 w-4" strokeWidth={2} />
					</WorkspaceHeaderIconButton>
				</ConversationTabActionRow>
			</ConversationTabNav>

			<div className="ml-auto flex shrink-0 items-center gap-0.5">
				{!selfDm ? (
					<>
						<div className="hidden sm:flex">
							<WorkspaceSplitIconControl mainAriaLabel="Start call" optionsAriaLabel="Call options">
								<Phone className={workspaceHeaderIconClass} strokeWidth={2} />
							</WorkspaceSplitIconControl>
						</div>
						<WorkspaceHeaderIconButton className="hidden sm:flex" aria-label="Add people">
							<UserPlus className={workspaceHeaderIconClass} strokeWidth={2} />
						</WorkspaceHeaderIconButton>
					</>
				) : null}
				<WorkspaceHeaderIconButton aria-label="Search in conversation">
					<Search className={workspaceHeaderIconClass} strokeWidth={2} />
				</WorkspaceHeaderIconButton>
				<WorkspaceHeaderIconButton className="hidden lg:flex" aria-label="Open details panel">
					<PanelRight className={workspaceHeaderIconClass} strokeWidth={2} />
				</WorkspaceHeaderIconButton>
				<WorkspaceHeaderIconButton aria-label="More actions">
					<MoreHorizontal className={workspaceHeaderIconClass} strokeWidth={2} />
				</WorkspaceHeaderIconButton>
			</div>
		</div>
	);
}
