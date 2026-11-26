import { floatingCollisionPadding } from "@/lib/floating-position";

export type ViewportFloaterAlign = "start" | "end";
export type ViewportFloaterSide = "bottom" | "top";

export type ViewportFloaterInput = {
	anchor: Pick<DOMRect, "top" | "bottom" | "left" | "right">;
	width: number;
	height: number;
	side?: ViewportFloaterSide;
	align?: ViewportFloaterAlign;
	gap?: number;
	padding?: number;
	viewportWidth?: number;
	viewportHeight?: number;
};

export type ViewportFloaterPosition = {
	top: number;
	left: number;
	side: ViewportFloaterSide;
};

/**
 * Flip/shift a fixed-position panel against the viewport (manual `position: fixed` UIs).
 */
export function positionWithinViewport(input: ViewportFloaterInput): ViewportFloaterPosition {
	const pad = input.padding ?? floatingCollisionPadding;
	const gap = input.gap ?? 4;
	const vw = input.viewportWidth ?? (typeof window !== "undefined" ? window.innerWidth : 0);
	const vh = input.viewportHeight ?? (typeof window !== "undefined" ? window.innerHeight : 0);
	const preferredSide = input.side ?? "bottom";
	const align = input.align ?? "start";
	const { anchor, width, height } = input;

	const belowTop = anchor.bottom + gap;
	const aboveTop = anchor.top - gap - height;
	const fitsBelow = belowTop + height <= vh - pad;
	const fitsAbove = aboveTop >= pad;

	let side: ViewportFloaterSide = preferredSide;
	let top: number;

	if (preferredSide === "bottom") {
		if (fitsBelow) {
			top = belowTop;
		} else if (fitsAbove) {
			top = aboveTop;
			side = "top";
		} else {
			top = Math.max(pad, Math.min(belowTop, vh - pad - height));
		}
	} else if (fitsAbove) {
		top = aboveTop;
	} else if (fitsBelow) {
		top = belowTop;
		side = "bottom";
	} else {
		top = Math.max(pad, Math.min(aboveTop, vh - pad - height));
	}

	let left = align === "end" ? anchor.right - width : anchor.left;
	const maxLeft = Math.max(pad, vw - pad - width);
	left = Math.min(Math.max(left, pad), maxLeft);

	return { top, left, side };
}