"use client";

import { Search } from "lucide-react";
import { forwardRef, type KeyboardEvent as ReactKeyboardEvent } from "react";
import { useTranslation } from "react-i18next";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";

/** Shared chrome for header trigger and popover search — one visual, two behaviors. */
const SHELL_SEARCH_PAD_X = "px-6";
const SHELL_SEARCH_ICON_INSET = "left-6";
const SHELL_SEARCH_PL = "pl-[3.25rem]";

const shellSearchFieldChrome = cn(
	"h-9 w-full bg-[#f5f5f6] text-[13px] shadow-none",
	"focus-visible:bg-[#f5f5f6] focus-visible:ring-0 dark:bg-[#242424] dark:focus-visible:bg-[#242424]",
	SHELL_SEARCH_PAD_X,
	SHELL_SEARCH_PL,
);

/** Bar only (trigger / idle). */
export const shellSearchFieldClass = cn(
	shellSearchFieldChrome,
	"rounded-md border-0 focus-visible:border-0",
);

/** Open search: flat bottom — panel continues below as one block (matches context menu surface). */
export const shellSearchFieldAttachedClass = cn(
	"h-9 w-full rounded-t-md rounded-b-none border-0 bg-card text-[13px] text-card-foreground shadow-none",
	"focus-visible:border-0 focus-visible:ring-0",
	SHELL_SEARCH_PAD_X,
	SHELL_SEARCH_PL,
);

type ShellSearchFieldBaseProps = {
	id?: string;
	className?: string;
	placeholder?: string;
};

type ShellSearchFieldTriggerProps = ShellSearchFieldBaseProps & {
	mode: "trigger";
};

type ShellSearchFieldActiveProps = ShellSearchFieldBaseProps & {
	mode: "active";
	value: string;
	onValueChange: (value: string) => void;
	onKeyDown?: (event: ReactKeyboardEvent<HTMLInputElement>) => void;
	"aria-expanded"?: boolean;
	"aria-controls"?: string;
};

export type ShellSearchFieldProps = ShellSearchFieldTriggerProps | ShellSearchFieldActiveProps;

export const ShellSearchField = forwardRef<HTMLInputElement, ShellSearchFieldProps>(
	function ShellSearchField(props, ref) {
		const { t } = useTranslation();
		const placeholder =
			props.placeholder ??
			(props.mode === "trigger" ? t("shell.search.shortcut_hint") : t("shell.search.placeholder"));

		const icon = (
			<Search
				className={cn(
					"pointer-events-none absolute top-1/2 z-10 size-3.5 -translate-y-1/2 text-[#a1a1aa]",
					SHELL_SEARCH_ICON_INSET,
				)}
				strokeWidth={2}
				aria-hidden
			/>
		);

		if (props.mode === "trigger") {
			return (
				<span className={cn("relative block w-full text-left", props.className)}>
					{icon}
					<span
						className={cn(
							shellSearchFieldClass,
							"flex cursor-pointer items-center text-muted-foreground",
						)}
					>
						<span className="truncate">{placeholder}</span>
					</span>
				</span>
			);
		}

		return (
			<div className={cn("relative w-full", props.className)}>
				{icon}
				<Input
					ref={ref}
					id={props.id}
					type="text"
					value={props.value}
					onChange={(e) => props.onValueChange(e.target.value)}
					onKeyDown={props.onKeyDown}
					placeholder={placeholder}
					aria-label={t("shell.search.placeholder")}
					aria-expanded={props["aria-expanded"]}
					aria-controls={props["aria-controls"]}
					autoComplete="off"
					className={cn(shellSearchFieldAttachedClass, "cursor-text")}
				/>
			</div>
		);
	},
);
