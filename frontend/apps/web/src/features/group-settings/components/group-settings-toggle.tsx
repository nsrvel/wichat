"use client";

import type { ButtonHTMLAttributes } from "react";
import { cn } from "@/lib/utils";

type GroupSettingsToggleProps = Omit<ButtonHTMLAttributes<HTMLButtonElement>, "type" | "role"> & {
	checked: boolean;
	onCheckedChange: (checked: boolean) => void;
};

export function GroupSettingsToggle({
	checked,
	onCheckedChange,
	className,
	disabled,
	...props
}: GroupSettingsToggleProps) {
	return (
		<button
			type="button"
			role="switch"
			aria-checked={checked}
			disabled={disabled}
			className={cn(
				"relative inline-flex h-6 w-11 shrink-0 items-center rounded-full border-2 border-transparent transition-colors",
				"focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background",
				"disabled:cursor-not-allowed disabled:opacity-50",
				checked ? "bg-primary" : "bg-muted",
				className,
			)}
			onClick={() => onCheckedChange(!checked)}
			{...props}
		>
			<span
				className={cn(
					"pointer-events-none block size-5 rounded-full bg-background shadow-sm transition-transform",
					checked ? "translate-x-5" : "translate-x-0.5",
				)}
				aria-hidden
			/>
		</button>
	);
}
