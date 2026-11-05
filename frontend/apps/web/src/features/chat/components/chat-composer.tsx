"use client";

import { AtSign, Paperclip, Plus, Send, Smile, Type } from "lucide-react";
import { cn } from "@/lib/utils";

const COMPOSER_ACTIONS = [
	{ icon: Type, label: "Formatting" },
	{ icon: Smile, label: "Emoji" },
	{ icon: Paperclip, label: "Attach file" },
	{ icon: AtSign, label: "Mention" },
	{ icon: Plus, label: "More" },
] as const;

type ChatComposerProps = {
	className?: string;
	placeholder?: string;
};

export function ChatComposer({ className, placeholder = "Type a message" }: ChatComposerProps) {
	return (
		<div
			className={cn(
				"mx-auto w-full max-w-[56rem] shrink-0 px-4 pb-4 pt-1 sm:px-5 lg:max-w-[60rem] lg:px-6 lg:pb-5",
				className,
			)}
		>
			<div className="flex min-h-11 items-center gap-2 rounded-xl border border-chat-bubble-border bg-chat-bubble px-3 py-1.5">
				<p className="min-w-0 flex-1 truncate text-[15px] leading-6 text-muted-foreground">
					{placeholder}
				</p>
				<div className="flex shrink-0 items-center">
					{COMPOSER_ACTIONS.map(({ icon: Icon, label }) => (
						<button
							key={label}
							type="button"
							className="flex h-8 w-8 shrink-0 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-sidebar-accent hover:text-sidebar-foreground"
							aria-label={label}
						>
							<Icon className="h-[18px] w-[18px]" strokeWidth={1.75} />
						</button>
					))}
					<span className="mx-1.5 h-6 w-px shrink-0 bg-chat-bubble-border" aria-hidden />
					<button
						type="button"
						className="flex h-8 w-8 shrink-0 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-sidebar-accent hover:text-sidebar-foreground"
						aria-label="Send message"
					>
						<Send className="h-[18px] w-[18px]" strokeWidth={1.75} />
					</button>
				</div>
			</div>
		</div>
	);
}
