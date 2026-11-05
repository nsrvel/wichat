import { ChevronDown, MessageSquarePlus, Search } from "lucide-react";
import {
	WorkspaceHeaderIconButton,
	workspaceHeaderIconClass,
} from "@/components/layout/workspace-header-icon-button";
import { cn } from "@/lib/utils";

type WorkspaceSidebarHeaderProps = {
	title: string;
	className?: string;
	onTitleClick?: () => void;
	onSearchClick?: () => void;
	onNewMessageClick?: () => void;
};

function HeaderIconButtons({
	onSearchClick,
	onNewMessageClick,
}: {
	onSearchClick?: () => void;
	onNewMessageClick?: () => void;
}) {
	return (
		<div className="ml-auto flex shrink-0 items-center gap-0.5">
			<WorkspaceHeaderIconButton aria-label="Search" onClick={onSearchClick}>
				<Search className={workspaceHeaderIconClass} strokeWidth={2} />
			</WorkspaceHeaderIconButton>
			<WorkspaceHeaderIconButton aria-label="New message" onClick={onNewMessageClick}>
				<MessageSquarePlus className={workspaceHeaderIconClass} strokeWidth={2} />
			</WorkspaceHeaderIconButton>
		</div>
	);
}

export function WorkspaceSidebarHeader({
	title,
	className,
	onTitleClick,
	onSearchClick,
	onNewMessageClick,
}: WorkspaceSidebarHeaderProps) {
	return (
		<div className={cn("flex w-full min-w-0 items-center gap-2", className)}>
			<button
				type="button"
				className="flex min-w-0 flex-1 items-center gap-1 text-left"
				onClick={onTitleClick}
			>
				<span className="truncate text-[16px] font-bold tracking-tight text-sidebar-foreground lg:text-[15px]">
					{title}
				</span>
				<ChevronDown className="h-3.5 w-3.5 shrink-0 text-muted-foreground" strokeWidth={2.4} />
			</button>
			<HeaderIconButtons onSearchClick={onSearchClick} onNewMessageClick={onNewMessageClick} />
		</div>
	);
}
