"use client";

import { cn } from "@/lib/utils";
import { ChevronDown } from "lucide-react";
import { useCallback, useEffect, useLayoutEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";

export interface SelectOption {
	value: string;
	label: string;
	/** Shown on trigger when set; list always uses `label`. */
	triggerLabel?: string;
}

interface CustomSelectProps {
	options: SelectOption[];
	value: string;
	placeholder: string;
	onChange: (value: string) => void;
	className?: string;
	triggerClassName: string;
	/** Extra classes on the listbox (use for portaled menus that leave the auth font subtree). */
	menuClassName?: string;
	/** When open, type letters/digits to jump (e.g. Oct → October). */
	typeahead?: boolean;
	/** Render menu in a portal (for triggers inside overflow containers). */
	portaled?: boolean;
	menuAlign?: "start" | "end";
	/** Inline trigger: tight label + chevron (header locale, etc.). */
	compactTrigger?: boolean;
}

type MenuPosition = { top: number; left: number; width: number };

function findTypeaheadMatch(
	options: SelectOption[],
	query: string,
	currentValue: string,
	sameKeyCycle: boolean,
): SelectOption | null {
	const q = query.toLowerCase().trim();
	if (!q) {
		return null;
	}

	const matches = options.filter(
		(o) => o.label.toLowerCase().startsWith(q) || o.value.toLowerCase().startsWith(q),
	);
	if (matches.length === 0) {
		return null;
	}
	if (matches.length === 1 || !sameKeyCycle) {
		return matches[0];
	}

	const currentIdx = matches.findIndex((o) => o.value === currentValue);
	const nextIdx = currentIdx >= 0 ? (currentIdx + 1) % matches.length : 0;

	return matches[nextIdx];
}

function measureMenuPosition(trigger: HTMLButtonElement, align: "start" | "end"): MenuPosition {
	const rect = trigger.getBoundingClientRect();
	return {
		top: rect.bottom + 4,
		left: align === "end" ? rect.right : rect.left,
		width: Math.max(rect.width, 168),
	};
}

export function CustomSelect({
	options,
	value,
	placeholder,
	onChange,
	className = "",
	triggerClassName,
	menuClassName,
	typeahead = false,
	portaled = false,
	menuAlign = "start",
	compactTrigger = false,
}: CustomSelectProps) {
	const [isOpen, setIsOpen] = useState(false);
	const [highlightedValue, setHighlightedValue] = useState(value);
	const [menuPosition, setMenuPosition] = useState<MenuPosition | null>(null);
	const containerRef = useRef<HTMLDivElement>(null);
	const triggerRef = useRef<HTMLButtonElement>(null);
	const listRef = useRef<HTMLDivElement>(null);
	const typeaheadBufRef = useRef("");
	const typeaheadTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
	const lastTypeaheadKeyRef = useRef("");

	const selectedOption = options.find((o) => o.value === value);
	const highlightedOption = options.find((o) => o.value === highlightedValue) ?? selectedOption;

	const triggerText = selectedOption?.triggerLabel ?? selectedOption?.label ?? placeholder;

	const updateMenuPosition = useCallback(() => {
		const trigger = triggerRef.current;
		if (!trigger) {
			return;
		}
		setMenuPosition(measureMenuPosition(trigger, menuAlign));
	}, [menuAlign]);

	const scrollToHighlighted = useCallback(() => {
		if (!listRef.current) {
			return;
		}
		const el = listRef.current.querySelector('[data-highlighted="true"]');
		el?.scrollIntoView({ block: "nearest" });
	}, []);

	const toggleOpen = () => {
		if (isOpen) {
			setIsOpen(false);
			return;
		}
		setHighlightedValue(value || options[0]?.value || "");
		setIsOpen(true);
		if (portaled) {
			updateMenuPosition();
		}
		requestAnimationFrame(scrollToHighlighted);
	};

	useLayoutEffect(() => {
		if (!isOpen || !portaled) {
			return;
		}
		updateMenuPosition();
		window.addEventListener("resize", updateMenuPosition);
		window.addEventListener("scroll", updateMenuPosition, true);
		return () => {
			window.removeEventListener("resize", updateMenuPosition);
			window.removeEventListener("scroll", updateMenuPosition, true);
		};
	}, [isOpen, portaled, updateMenuPosition]);

	useEffect(() => {
		function handleClickOutside(event: MouseEvent) {
			const target = event.target as Node;
			if (containerRef.current?.contains(target) || listRef.current?.contains(target)) {
				return;
			}
			setIsOpen(false);
		}
		if (isOpen) {
			document.addEventListener("mousedown", handleClickOutside);
		}
		return () => {
			document.removeEventListener("mousedown", handleClickOutside);
		};
	}, [isOpen]);

	useEffect(() => {
		if (isOpen) {
			return;
		}
		typeaheadBufRef.current = "";
		lastTypeaheadKeyRef.current = "";
		if (typeaheadTimerRef.current) {
			clearTimeout(typeaheadTimerRef.current);
		}
	}, [isOpen]);

	useEffect(() => {
		if (isOpen) {
			requestAnimationFrame(scrollToHighlighted);
		}
	}, [highlightedValue, isOpen, scrollToHighlighted]);

	useEffect(() => {
		if (!isOpen || !typeahead) {
			return;
		}

		function onKeyDown(e: KeyboardEvent) {
			if (e.key === "Escape") {
				e.preventDefault();
				setIsOpen(false);
				return;
			}

			if (e.key === "Enter") {
				e.preventDefault();
				if (highlightedValue) {
					onChange(highlightedValue);
					setIsOpen(false);
				}
				return;
			}

			if (e.key === "ArrowDown" || e.key === "ArrowUp") {
				e.preventDefault();
				const idx = options.findIndex((o) => o.value === highlightedValue);
				const next =
					e.key === "ArrowDown"
						? options[(idx + 1 + options.length) % options.length]
						: options[(idx - 1 + options.length) % options.length];
				if (next) {
					setHighlightedValue(next.value);
					typeaheadBufRef.current = "";
				}
				return;
			}

			if (e.key.length !== 1 || e.ctrlKey || e.metaKey || e.altKey) {
				return;
			}

			const sameKeyCycle =
				e.key === lastTypeaheadKeyRef.current && typeaheadBufRef.current.length === 1;
			const nextBuf = sameKeyCycle ? e.key : typeaheadBufRef.current + e.key;
			lastTypeaheadKeyRef.current = e.key;

			const match = findTypeaheadMatch(options, nextBuf, highlightedValue, sameKeyCycle);
			typeaheadBufRef.current = match ? nextBuf : e.key;
			if (!match) {
				const fallback = findTypeaheadMatch(options, e.key, highlightedValue, false);
				if (fallback) {
					typeaheadBufRef.current = e.key;
					setHighlightedValue(fallback.value);
				}
			} else {
				setHighlightedValue(match.value);
			}

			if (typeaheadTimerRef.current) {
				clearTimeout(typeaheadTimerRef.current);
			}
			typeaheadTimerRef.current = setTimeout(() => {
				typeaheadBufRef.current = "";
				lastTypeaheadKeyRef.current = "";
			}, 700);

			e.preventDefault();
		}

		window.addEventListener("keydown", onKeyDown);
		return () => {
			window.removeEventListener("keydown", onKeyDown);
		};
	}, [highlightedValue, isOpen, onChange, options, typeahead]);

	const focused = isOpen;

	const listboxClassName =
		"z-50 max-h-[min(14rem,50dvh)] min-w-0 overflow-y-auto rounded-[10px] border border-border bg-popover p-1 text-popover-foreground shadow-elevated-lg sm:max-h-60 sm:rounded-xl sm:p-1.5";

	const listbox = isOpen ? (
		<div
			ref={listRef}
			role="listbox"
			className={cn(
				listboxClassName,
				menuClassName,
				portaled ? "fixed" : "absolute top-[calc(100%+4px)] left-0 w-full",
			)}
			style={
				portaled && menuPosition
					? {
							top: menuPosition.top,
							left: menuPosition.left,
							width: menuPosition.width,
							transform: menuAlign === "end" ? "translateX(-100%)" : undefined,
							scrollbarWidth: "thin",
						}
					: { scrollbarWidth: "thin" }
			}
		>
			{options.map((opt) => {
				const isSelected = opt.value === value;
				const isHighlighted = opt.value === highlightedOption?.value;
				return (
					<button
						key={opt.value}
						type="button"
						role="option"
						aria-selected={isSelected}
						data-selected={isSelected}
						data-highlighted={isHighlighted}
						onMouseEnter={() => setHighlightedValue(opt.value)}
						onClick={() => {
							onChange(opt.value);
							setIsOpen(false);
						}}
						className={cn(
							"flex w-full items-center rounded-md px-2 py-2 text-left text-sm transition-colors sm:rounded-lg sm:px-2.5 sm:py-1.5",
							isSelected
								? "bg-primary font-medium text-primary-foreground"
								: isHighlighted
									? "bg-foreground/8 text-foreground"
									: "text-muted-foreground hover:bg-foreground/5 hover:text-foreground",
						)}
					>
						<span className="truncate">{opt.label}</span>
					</button>
				);
			})}
		</div>
	) : null;

	return (
		<div ref={containerRef} className={`relative min-w-0 ${className}`}>
			<button
				ref={triggerRef}
				type="button"
				aria-haspopup="listbox"
				aria-expanded={isOpen}
				onClick={toggleOpen}
				className={cn(
					triggerClassName,
					compactTrigger
						? "inline-flex w-auto items-center gap-0.5"
						: "flex w-full items-center justify-between",
					focused && !compactTrigger && "border-foreground/20",
					focused && compactTrigger && "text-foreground",
					!selectedOption ? "text-placeholder-foreground" : "",
				)}
			>
				<span className="truncate">{triggerText}</span>
				<ChevronDown
					className={cn(
						"h-3.5 w-3.5 shrink-0 text-muted-foreground transition-transform duration-200",
						!compactTrigger && "ml-1 h-4 w-4",
						isOpen && "rotate-180",
					)}
				/>
			</button>

			{portaled && listbox ? createPortal(listbox, document.body) : listbox}
		</div>
	);
}
