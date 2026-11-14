"use client";

import { Search, X } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Input } from "@/components/ui/input";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { cn } from "@/lib/utils";

const SKIN_TONE_OPTIONS = [
	{ tone: 0, sample: "👋", label: "Default" },
	{ tone: 1, sample: "👋🏻", label: "Light" },
	{ tone: 2, sample: "👋🏼", label: "Medium-Light" },
	{ tone: 3, sample: "👋🏽", label: "Medium" },
	{ tone: 4, sample: "👋🏾", label: "Medium-Dark" },
	{ tone: 5, sample: "👋🏿", label: "Dark" },
] as const;

type ExpressionSearchBarProps = {
	value: string;
	onChange: (value: string) => void;
	showSkinTonePicker?: boolean;
	skinTone?: number;
	onSkinToneChange?: (tone: number) => void;
	className?: string;
};

export function ExpressionSearchBar({
	value,
	onChange,
	showSkinTonePicker = false,
	skinTone = 0,
	onSkinToneChange,
	className,
}: ExpressionSearchBarProps) {
	const { t } = useTranslation();
	const [skinPopoverOpen, setSkinPopoverOpen] = useState(false);

	const activeOption =
		SKIN_TONE_OPTIONS.find((opt) => opt.tone === skinTone) ?? SKIN_TONE_OPTIONS[0];

	return (
		<div className={cn("flex items-center gap-2 px-3.5 pt-3.5 pb-4", className)}>
			<div className="relative flex flex-1 items-center">
				<Input
					type="text"
					value={value}
					onChange={(event) => onChange(event.target.value)}
					placeholder={t("expression.search_placeholder")}
					aria-label={t("expression.search_placeholder")}
					className="h-9 w-full rounded-md border-0 bg-[#f5f5f6] pl-3.5 pr-9 text-[13px] text-foreground shadow-none placeholder:text-muted-foreground focus-visible:border-0 focus-visible:bg-[#f5f5f6] focus-visible:ring-0 dark:bg-[#242424] dark:focus-visible:bg-[#242424]"
				/>
				{value ? (
					<button
						type="button"
						onClick={() => onChange("")}
						aria-label={t("expression.clear_search")}
						className="absolute right-2.5 flex size-5 items-center justify-center rounded text-muted-foreground hover:text-foreground"
					>
						<X className="size-3.5" />
					</button>
				) : (
					<Search
						className="pointer-events-none absolute right-3 size-4 text-muted-foreground"
						aria-hidden
					/>
				)}
			</div>

			{showSkinTonePicker && onSkinToneChange ? (
				<Popover open={skinPopoverOpen} onOpenChange={setSkinPopoverOpen}>
					<PopoverTrigger
						render={
							<button
								type="button"
								title={t("expression.skin_tone")}
								aria-label={t("expression.skin_tone")}
								className="flex size-9 shrink-0 items-center justify-center rounded-md bg-[#f5f5f6] text-[18px] leading-none transition-colors hover:bg-accent data-[popup-open]:bg-accent dark:bg-[#242424] dark:hover:bg-[#303030] dark:data-[popup-open]:bg-[#303030]"
							>
								{activeOption.sample}
							</button>
						}
					/>
					<PopoverContent
						side="bottom"
						align="end"
						sideOffset={6}
						className="w-[224px] border-border bg-card p-2.5 text-card-foreground shadow-lg"
					>
						<p className="mb-2 px-0.5 text-[11px] font-medium text-muted-foreground">
							{t("expression.skin_tone_hint")}
						</p>
						<div
							className="grid grid-cols-6 gap-1"
							role="radiogroup"
							aria-label={t("expression.skin_tone")}
						>
							{SKIN_TONE_OPTIONS.map((option) => {
								const selected = skinTone === option.tone;
								return (
									<button
										key={option.tone}
										type="button"
										role="radio"
										aria-checked={selected}
										aria-label={option.label}
										title={option.label}
										onClick={() => {
											onSkinToneChange(option.tone);
											setSkinPopoverOpen(false);
										}}
										className={cn(
											"flex size-8 items-center justify-center rounded-md text-[18px] leading-none transition-colors hover:bg-accent",
											selected &&
												"bg-primary/15 ring-1 ring-primary/50",
										)}
									>
										{option.sample}
									</button>
								);
							})}
						</div>
					</PopoverContent>
				</Popover>
			) : null}
		</div>
	);
}
