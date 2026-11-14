"use client";

import type { ButtonHTMLAttributes } from "react";
import { UserAvatar } from "@/components/ui/user-avatar";
import type { UserPresenceStatus } from "@/lib/user-status";
import { cn } from "@/lib/utils";

type UserProfileButtonProps = Omit<ButtonHTMLAttributes<HTMLButtonElement>, "children"> & {
	displayName: string;
	secondaryText: string;
	avatarUrl: string;
	initials: string;
	status: UserPresenceStatus;
	statusRingClassName?: string;
};

export function UserProfileButton({
	displayName,
	secondaryText,
	avatarUrl,
	initials,
	status,
	statusRingClassName = "ring-chat-surface",
	className,
	type = "button",
	...props
}: UserProfileButtonProps) {
	return (
		<button
			type={type}
			className={cn(
				"flex min-h-9 min-w-0 flex-1 items-center gap-2 rounded-lg px-1.5 py-1 text-left transition-colors hover:bg-sidebar-accent",
				className,
			)}
			{...props}
		>
			<UserAvatar
				src={avatarUrl}
				initials={initials}
				size="sm"
				status={status}
				statusRingClassName={statusRingClassName}
			/>
			<div className="flex min-w-0 flex-1 flex-col gap-0.5">
				<p className="truncate text-[13px] font-semibold leading-tight text-foreground">
					{displayName}
				</p>
				<p
					className="truncate text-[11px] leading-tight text-muted-foreground"
					suppressHydrationWarning
				>
					{secondaryText}
				</p>
			</div>
		</button>
	);
}
