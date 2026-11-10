"use client";

import { SmilePlus } from "lucide-react";
import { useState } from "react";
import { cn } from "@/lib/utils";

export function DemoReaction({
	emoji,
	count,
	active,
}: {
	emoji: string;
	count?: number;
	active?: boolean;
}) {
	const [n, setN] = useState(count ?? 0);
	const [on, setOn] = useState(!!active);
	return (
		<button
			type="button"
			onClick={() => {
				setN(on ? n - 1 : n + 1);
				setOn(!on);
			}}
			className={cn(
				"flex h-6 min-w-[30px] items-center justify-center gap-1 rounded-full border px-1.5 text-xs transition-colors",
				on
					? "border-[#c9d6ec] bg-[#f0f4fa] dark:border-[#3d4f6f] dark:bg-[#1e2a3d]"
					: "border-chat-surface-border bg-chat-composer-field hover:bg-sidebar-accent",
			)}
		>
			<span className="text-[13px] leading-none">{emoji}</span>
			{count !== undefined ? (
				<span className="text-[11px] font-medium text-muted-foreground">{n}</span>
			) : null}
		</button>
	);
}

export function DemoAddReaction() {
	return (
		<button
			type="button"
			className="flex h-6 w-8 items-center justify-center rounded-full border border-chat-surface-border bg-chat-surface text-muted-foreground transition-colors hover:bg-sidebar-accent"
		>
			<SmilePlus className="h-3.5 w-3.5" strokeWidth={1.8} />
		</button>
	);
}
