"use client";

import { Search } from "lucide-react";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";

/** 2× default `Input` horizontal padding (`px-3` → `px-6`). */
const SHELL_SEARCH_PAD_X = "px-6";
const SHELL_SEARCH_ICON_INSET = "left-6";
/** Room for doubled inset + icon + gap (paired with `SHELL_SEARCH_ICON_INSET`). */
const SHELL_SEARCH_PL = "pl-[3.25rem]";

type ShellSearchInputProps = {
	className?: string;
	placeholder?: string;
};

export function ShellSearchInput({
	className,
	placeholder = "Search (⌥ ⌘ E)",
}: ShellSearchInputProps) {
	return (
		<div className={cn("relative w-full", className)}>
			<Search
				className={cn(
					"pointer-events-none absolute top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-[#a1a1aa]",
					SHELL_SEARCH_ICON_INSET,
				)}
				strokeWidth={2}
				aria-hidden
			/>
			<Input
				type="search"
				placeholder={placeholder}
				aria-label="Search"
				className={cn(
					"h-8 rounded-md border-0 bg-[#f5f5f6] text-[13px] shadow-none focus-visible:border-0 focus-visible:bg-[#f5f5f6] focus-visible:ring-0 dark:bg-[#242424] dark:focus-visible:bg-[#242424]",
					SHELL_SEARCH_PAD_X,
					SHELL_SEARCH_PL,
				)}
			/>
		</div>
	);
}
