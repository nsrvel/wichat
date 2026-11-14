"use client";

import { RotateCcw } from "lucide-react";
import { useRef, useState, type KeyboardEvent, type PointerEvent } from "react";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { cn } from "@/lib/utils";

type HsvColor = {
	h: number;
	s: number;
	v: number;
};

const HEX_6_RE = /^#?[0-9a-fA-F]{6}$/;

export function normalizeHexColor(raw: string, fallback = "#313338"): string {
	const trimmed = raw.trim();
	if (!HEX_6_RE.test(trimmed)) {
		return fallback.toUpperCase();
	}
	return (trimmed.startsWith("#") ? trimmed : `#${trimmed}`).toUpperCase();
}

function hexToRgb(hex: string): { r: number; g: number; b: number } {
	const clean = normalizeHexColor(hex).slice(1);
	const num = Number.parseInt(clean, 16);
	return {
		r: (num >> 16) & 255,
		g: (num >> 8) & 255,
		b: num & 255,
	};
}

function rgbToHex(r: number, g: number, b: number): string {
	const toByte = (n: number) => Math.max(0, Math.min(255, Math.round(n))).toString(16).padStart(2, "0");
	return `#${toByte(r)}${toByte(g)}${toByte(b)}`.toUpperCase();
}

function hexToHsv(hex: string): HsvColor {
	const { r, g, b } = hexToRgb(hex);
	const rn = r / 255;
	const gn = g / 255;
	const bn = b / 255;
	const max = Math.max(rn, gn, bn);
	const min = Math.min(rn, gn, bn);
	const delta = max - min;

	let h = 0;
	if (delta !== 0) {
		if (max === rn) {
			h = 60 * (((gn - bn) / delta) % 6);
		} else if (max === gn) {
			h = 60 * ((bn - rn) / delta + 2);
		} else {
			h = 60 * ((rn - gn) / delta + 4);
		}
	}
	if (h < 0) {
		h += 360;
	}
	const s = max === 0 ? 0 : (delta / max) * 100;
	const v = max * 100;
	return { h, s, v };
}

function hsvToHex({ h, s, v }: HsvColor): string {
	const sn = Math.max(0, Math.min(100, s)) / 100;
	const vn = Math.max(0, Math.min(100, v)) / 100;
	const c = vn * sn;
	const x = c * (1 - Math.abs(((h / 60) % 2) - 1));
	const m = vn - c;

	let rp = 0;
	let gp = 0;
	let bp = 0;
	if (h < 60) {
		rp = c;
		gp = x;
	} else if (h < 120) {
		rp = x;
		gp = c;
	} else if (h < 180) {
		gp = c;
		bp = x;
	} else if (h < 240) {
		gp = x;
		bp = c;
	} else if (h < 300) {
		rp = x;
		bp = c;
	} else {
		rp = c;
		bp = x;
	}
	return rgbToHex((rp + m) * 255, (gp + m) * 255, (bp + m) * 255);
}

type HexColorFieldProps = {
	id?: string;
	value: string;
	onChange: (hex: string) => void;
	defaultColor?: string;
	resetLabel?: string;
	pickerAriaLabel?: string;
	inputClassName?: string;
	className?: string;
};

/**
 * Manual `#RRGGBB` hex field paired with an interactive Saturation/Brightness + Hue popover picker.
 */
export function HexColorField({
	id,
	value,
	onChange,
	defaultColor = "#313338",
	resetLabel = "Reset",
	pickerAriaLabel = "Pick color",
	inputClassName,
	className,
}: HexColorFieldProps) {
	const normalized = normalizeHexColor(value, defaultColor);
	const [draft, setDraft] = useState<string | null>(null);
	const [open, setOpen] = useState(false);
	const [hsv, setHsv] = useState<HsvColor>(() => hexToHsv(normalized));
	const svBoxRef = useRef<HTMLDivElement>(null);
	const hueBarRef = useRef<HTMLDivElement>(null);

	const activeHsv = open ? hsv : hexToHsv(normalized);
	const displayText = draft ?? normalized;
	const isDefault = normalized === normalizeHexColor(defaultColor, defaultColor);

	const commitHsv = (next: HsvColor) => {
		setHsv(next);
		setDraft(null);
		onChange(hsvToHex(next));
	};

	const handleTextChange = (raw: string) => {
		setDraft(raw);
		if (HEX_6_RE.test(raw.trim())) {
			const nextHex = normalizeHexColor(raw, normalized);
			setHsv(hexToHsv(nextHex));
			onChange(nextHex);
		}
	};

	const handleTextBlur = () => {
		if (draft !== null && HEX_6_RE.test(draft.trim())) {
			const nextHex = normalizeHexColor(draft, normalized);
			setHsv(hexToHsv(nextHex));
			onChange(nextHex);
		}
		setDraft(null);
	};

	const updateSvFromPointer = (clientX: number, clientY: number) => {
		const el = svBoxRef.current;
		if (!el) {
			return;
		}
		const rect = el.getBoundingClientRect();
		const s = Math.max(0, Math.min(1, (clientX - rect.left) / rect.width)) * 100;
		const v = (1 - Math.max(0, Math.min(1, (clientY - rect.top) / rect.height))) * 100;
		commitHsv({ h: activeHsv.h, s, v });
	};

	const updateHueFromPointer = (clientX: number) => {
		const el = hueBarRef.current;
		if (!el) {
			return;
		}
		const rect = el.getBoundingClientRect();
		const ratio = Math.max(0, Math.min(1, (clientX - rect.left) / rect.width));
		commitHsv({ ...activeHsv, h: ratio * 360 });
	};

	const onSvPointerDown = (e: PointerEvent<HTMLDivElement>) => {
		e.currentTarget.setPointerCapture(e.pointerId);
		updateSvFromPointer(e.clientX, e.clientY);
	};

	const onSvPointerMove = (e: PointerEvent<HTMLDivElement>) => {
		if (e.buttons !== 1 && !e.currentTarget.hasPointerCapture(e.pointerId)) {
			return;
		}
		updateSvFromPointer(e.clientX, e.clientY);
	};

	const onHuePointerDown = (e: PointerEvent<HTMLDivElement>) => {
		e.currentTarget.setPointerCapture(e.pointerId);
		updateHueFromPointer(e.clientX);
	};

	const onHuePointerMove = (e: PointerEvent<HTMLDivElement>) => {
		if (e.buttons !== 1 && !e.currentTarget.hasPointerCapture(e.pointerId)) {
			return;
		}
		updateHueFromPointer(e.clientX);
	};

	const onSvKeyDown = (e: KeyboardEvent<HTMLDivElement>) => {
		const step = e.shiftKey ? 10 : 2;
		if (e.key === "ArrowLeft") {
			e.preventDefault();
			commitHsv({ ...activeHsv, s: Math.max(0, activeHsv.s - step) });
		} else if (e.key === "ArrowRight") {
			e.preventDefault();
			commitHsv({ ...activeHsv, s: Math.min(100, activeHsv.s + step) });
		} else if (e.key === "ArrowUp") {
			e.preventDefault();
			commitHsv({ ...activeHsv, v: Math.min(100, activeHsv.v + step) });
		} else if (e.key === "ArrowDown") {
			e.preventDefault();
			commitHsv({ ...activeHsv, v: Math.max(0, activeHsv.v - step) });
		}
	};

	const onHueKeyDown = (e: KeyboardEvent<HTMLDivElement>) => {
		const step = e.shiftKey ? 15 : 5;
		if (e.key === "ArrowLeft" || e.key === "ArrowDown") {
			e.preventDefault();
			commitHsv({ ...activeHsv, h: (activeHsv.h - step + 360) % 360 });
		} else if (e.key === "ArrowRight" || e.key === "ArrowUp") {
			e.preventDefault();
			commitHsv({ ...activeHsv, h: (activeHsv.h + step) % 360 });
		}
	};

	return (
		<div className={cn("flex items-center gap-2", className)}>
			<Popover
				open={open}
				onOpenChange={(nextOpen) => {
					if (nextOpen) {
						setHsv(hexToHsv(normalized));
					}
					setOpen(nextOpen);
				}}
			>
				<div
					className={cn(
						"flex h-10 w-full items-center gap-2.5 rounded-[10px] border border-border bg-field px-3 text-sm text-foreground transition-colors focus-within:border-ring focus-within:ring-2 focus-within:ring-ring/20",
						inputClassName,
						"px-2.5 sm:px-3",
					)}
				>
					<PopoverTrigger
						render={
							<button
								type="button"
								aria-label={pickerAriaLabel}
								style={{ backgroundColor: normalized }}
								className="size-6 shrink-0 rounded-md shadow-xs transition-transform hover:scale-105 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
							/>
						}
					/>
					<input
						id={id}
						value={displayText}
						onChange={(e) => handleTextChange(e.target.value)}
						onBlur={handleTextBlur}
						maxLength={7}
						spellCheck={false}
						autoComplete="off"
						placeholder={defaultColor}
						className="min-w-0 flex-1 bg-transparent font-mono uppercase text-foreground placeholder:text-placeholder-foreground focus:outline-none"
					/>
					{!isDefault ? (
						<button
							type="button"
							aria-label={resetLabel}
							title={resetLabel}
							onClick={() => {
								const fallback = normalizeHexColor(defaultColor, defaultColor);
								setDraft(null);
								setHsv(hexToHsv(fallback));
								onChange(fallback);
							}}
							className="flex size-7 shrink-0 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-foreground/10 hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
						>
							<RotateCcw className="size-3.5" aria-hidden />
						</button>
					) : null}
				</div>

				<PopoverContent side="bottom" align="start" sideOffset={8} className="w-[248px] p-3">
					<div className="space-y-3">
						{/* 2D Saturation / Brightness selector */}
						<div
							ref={svBoxRef}
							role="slider"
							tabIndex={0}
							aria-label="Saturation and brightness"
							aria-valuemin={0}
							aria-valuemax={100}
							aria-valuenow={Math.round(activeHsv.v)}
							aria-valuetext={`Saturation ${Math.round(activeHsv.s)}%, Brightness ${Math.round(activeHsv.v)}%`}
							onPointerDown={onSvPointerDown}
							onPointerMove={onSvPointerMove}
							onKeyDown={onSvKeyDown}
							style={{ backgroundColor: `hsl(${activeHsv.h}, 100%, 50%)` }}
							className="relative h-36 w-full cursor-crosshair touch-none select-none overflow-hidden rounded-lg border border-border/80 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
						>
							<div
								className="pointer-events-none absolute inset-0"
								style={{ background: "linear-gradient(to right, #fff, transparent)" }}
							/>
							<div
								className="pointer-events-none absolute inset-0"
								style={{ background: "linear-gradient(to top, #000, transparent)" }}
							/>
							<span
								aria-hidden
								style={{
									left: `${activeHsv.s}%`,
									top: `${100 - activeHsv.v}%`,
									backgroundColor: normalized,
								}}
								className="pointer-events-none absolute size-3.5 -translate-x-1/2 -translate-y-1/2 rounded-full border-2 border-white shadow-[0_0_0_1px_rgba(0,0,0,0.45)]"
							/>
						</div>

						{/* Hue slider */}
						<div
							ref={hueBarRef}
							role="slider"
							tabIndex={0}
							aria-label="Hue"
							aria-valuemin={0}
							aria-valuemax={360}
							aria-valuenow={Math.round(activeHsv.h)}
							onPointerDown={onHuePointerDown}
							onPointerMove={onHuePointerMove}
							onKeyDown={onHueKeyDown}
							style={{
								background:
									"linear-gradient(to right, #ff0000 0%, #ffff00 17%, #00ff00 33%, #00ffff 50%, #0000ff 67%, #ff00ff 83%, #ff0000 100%)",
							}}
							className="relative h-3.5 w-full cursor-pointer touch-none select-none rounded-md border border-border/70 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
						>
							<span
								aria-hidden
								style={{
									left: `${(activeHsv.h / 360) * 100}%`,
									backgroundColor: `hsl(${activeHsv.h}, 100%, 50%)`,
								}}
								className="pointer-events-none absolute top-1/2 size-4 -translate-x-1/2 -translate-y-1/2 rounded-full border-2 border-white shadow-[0_0_0_1px_rgba(0,0,0,0.45)]"
							/>
						</div>

						{/* Synced Hex readout + quick input (same field style & size, non-clickable swatch) */}
						<div
							className={cn(
								"flex h-10 w-full items-center gap-2.5 rounded-[10px] border border-border bg-field px-3 text-sm text-foreground transition-colors focus-within:border-ring focus-within:ring-2 focus-within:ring-ring/20",
								inputClassName,
								"px-2.5 sm:px-3",
							)}
						>
							<span
								aria-hidden
								style={{ backgroundColor: normalized }}
								className="size-6 shrink-0 rounded-md"
							/>
							<input
								value={displayText}
								onChange={(e) => handleTextChange(e.target.value)}
								onBlur={handleTextBlur}
								maxLength={7}
								spellCheck={false}
								autoComplete="off"
								aria-label="Hex color value"
								className="min-w-0 flex-1 bg-transparent font-mono uppercase text-foreground placeholder:text-placeholder-foreground focus:outline-none"
							/>
						</div>
					</div>
				</PopoverContent>
			</Popover>
		</div>
	);
}
