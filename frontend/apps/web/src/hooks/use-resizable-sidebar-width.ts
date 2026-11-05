"use client";

import { useCallback, useEffect, useRef, useState } from "react";

export const CHANNEL_SIDEBAR_MIN_WIDTH = 280;
export const CHANNEL_SIDEBAR_MAX_WIDTH = 400;
const STORAGE_KEY = "wichat-channel-sidebar-width";

function clampWidth(n: number, minWidth: number) {
	return Math.min(CHANNEL_SIDEBAR_MAX_WIDTH, Math.max(minWidth, n));
}

export function useResizableSidebarWidth(enabled: boolean, minWidth = CHANNEL_SIDEBAR_MIN_WIDTH) {
	const [width, setWidth] = useState(() => clampWidth(CHANNEL_SIDEBAR_MIN_WIDTH, minWidth));
	const widthRef = useRef(width);
	const minWidthRef = useRef(minWidth);

	useEffect(() => {
		widthRef.current = width;
	}, [width]);

	useEffect(() => {
		minWidthRef.current = minWidth;
	}, [minWidth]);

	useEffect(() => {
		if (!enabled) return;
		const raw = localStorage.getItem(STORAGE_KEY);
		if (!raw) return;
		const parsed = Number.parseInt(raw, 10);
		if (!Number.isNaN(parsed)) {
			const next = clampWidth(parsed, minWidthRef.current);
			setWidth(next);
			widthRef.current = next;
		}
	}, [enabled]);

	const onResizePointerDown = useCallback(
		(event: React.PointerEvent<HTMLDivElement>) => {
			if (!enabled) return;
			event.preventDefault();
			const startX = event.clientX;
			const startWidth = widthRef.current;

			const onMove = (moveEvent: PointerEvent) => {
				const next = clampWidth(startWidth + (moveEvent.clientX - startX), minWidthRef.current);
				widthRef.current = next;
				setWidth(next);
			};

			const onUp = () => {
				document.removeEventListener("pointermove", onMove);
				document.removeEventListener("pointerup", onUp);
				document.body.style.cursor = "";
				document.body.style.userSelect = "";
				localStorage.setItem(STORAGE_KEY, String(widthRef.current));
			};

			document.body.style.cursor = "col-resize";
			document.body.style.userSelect = "none";
			document.addEventListener("pointermove", onMove);
			document.addEventListener("pointerup", onUp);
		},
		[enabled],
	);

	return { width: clampWidth(width, minWidth), onResizePointerDown };
}
