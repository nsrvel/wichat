"use client";

import { useCallback, useEffect, useId, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { useTranslation } from "react-i18next";
import {
	groupBannerColorSwatchClass,
	groupBannerSwatchSelectedRingClass,
	groupBannerSwatchSurfaceClass,
	type GroupBannerColorId,
} from "@/lib/group-banner-presets";
import { useMounted } from "@/hooks/use-mounted";
import { cn } from "@/lib/utils";

const SHOW_DELAY_MS = 200;

type GroupBannerPaletteSwatchProps = {
	id: GroupBannerColorId;
	selected: boolean;
	onSelect: (id: GroupBannerColorId) => void;
};

type TooltipAnchor = {
	x: number;
	y: number;
};

export function GroupBannerPaletteSwatch({
	id,
	selected,
	onSelect,
}: GroupBannerPaletteSwatchProps) {
	const { t } = useTranslation();
	const label = t(`group_settings.profile.banner_colors.${id}`);
	const tooltipId = useId();
	const mounted = useMounted();
	const showTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
	const [tooltipAnchor, setTooltipAnchor] = useState<TooltipAnchor | null>(null);

	const clearShowTimer = useCallback(() => {
		if (showTimer.current) {
			clearTimeout(showTimer.current);
			showTimer.current = null;
		}
	}, []);

	const hideTooltip = useCallback(() => {
		clearShowTimer();
		setTooltipAnchor(null);
	}, [clearShowTimer]);

	const showTooltipFor = useCallback(
		(button: HTMLButtonElement) => {
			clearShowTimer();
			showTimer.current = setTimeout(() => {
				const rect = button.getBoundingClientRect();
				setTooltipAnchor({
					x: rect.left + rect.width / 2,
					y: rect.top,
				});
			}, SHOW_DELAY_MS);
		},
		[clearShowTimer],
	);

	useEffect(() => () => clearShowTimer(), [clearShowTimer]);

	useEffect(() => {
		if (!tooltipAnchor) {
			return;
		}
		const onScroll = () => hideTooltip();
		window.addEventListener("scroll", onScroll, true);
		return () => window.removeEventListener("scroll", onScroll, true);
	}, [hideTooltip, tooltipAnchor]);

	return (
		<>
			<button
				type="button"
				className={cn(
					groupBannerColorSwatchClass,
					groupBannerSwatchSurfaceClass(id),
					selected && cn("border-transparent", groupBannerSwatchSelectedRingClass),
				)}
				aria-label={label}
				aria-describedby={tooltipAnchor ? tooltipId : undefined}
				aria-pressed={selected}
				onMouseEnter={(event) => showTooltipFor(event.currentTarget)}
				onMouseLeave={hideTooltip}
				onFocus={(event) => showTooltipFor(event.currentTarget)}
				onBlur={hideTooltip}
				onClick={() => {
					hideTooltip();
					onSelect(id);
				}}
			></button>

			{mounted && tooltipAnchor
				? createPortal(
						<div
							id={tooltipId}
							role="tooltip"
							className="pointer-events-none fixed z-[110] -translate-x-1/2 -translate-y-[calc(100%+0.4rem)] whitespace-nowrap rounded-lg border border-white/12 bg-[#111214] px-3 py-1.5 text-sm font-semibold text-white shadow-lg"
							style={{ left: tooltipAnchor.x, top: tooltipAnchor.y }}
						>
							{label}
							<span
								className="absolute top-full left-1/2 size-0 -translate-x-1/2 border-x-[6px] border-t-[6px] border-x-transparent border-t-[#111214]"
								aria-hidden
							/>
						</div>,
						document.body,
					)
				: null}
		</>
	);
}
