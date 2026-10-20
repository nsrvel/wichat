"use client";

import { cn } from "@/lib/utils";
import { ChevronDown } from "lucide-react";
import { useCallback, useEffect, useRef, useState } from "react";

interface Option {
	value: string;
	label: string;
}

interface CustomSelectProps {
	options: Option[];
	value: string;
	placeholder: string;
	onChange: (value: string) => void;
	className?: string;
	triggerClassName: string;
	/** When open, type letters/digits to jump (e.g. Oct → October). */
	typeahead?: boolean;
}

function findTypeaheadMatch(
	options: Option[],
	query: string,
	currentValue: string,
	sameKeyCycle: boolean,
): Option | null {
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

export function CustomSelect({
	options,
	value,
	placeholder,
	onChange,
	className = "",
	triggerClassName,
	typeahead = false,
}: CustomSelectProps) {
	const [isOpen, setIsOpen] = useState(false);
	const [highlightedValue, setHighlightedValue] = useState(value);
	const containerRef = useRef<HTMLDivElement>(null);
	const listRef = useRef<HTMLDivElement>(null);
	const typeaheadBufRef = useRef("");
	const typeaheadTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
	const lastTypeaheadKeyRef = useRef("");

	const selectedOption = options.find((o) => o.value === value);
	const highlightedOption =
		options.find((o) => o.value === highlightedValue) ?? selectedOption;

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
		requestAnimationFrame(scrollToHighlighted);
	};

	useEffect(() => {
		function handleClickOutside(event: MouseEvent) {
			if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
				setIsOpen(false);
			}
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

	return (
		<div ref={containerRef} className={`relative min-w-0 ${className}`}>
			<button
				type="button"
				aria-haspopup="listbox"
				aria-expanded={isOpen}
				onClick={toggleOpen}
				className={cn(
					triggerClassName,
					"flex w-full items-center justify-between",
					focused && "border-black/20 dark:border-white/22",
					!selectedOption ? "text-[#9a9a9a] dark:text-[#8d8d8d]" : "",
				)}
			>
				<span className="truncate">
					{selectedOption ? selectedOption.label : placeholder}
				</span>
				<ChevronDown
					className={`ml-1 h-4 w-4 shrink-0 text-[#6d6d6d] transition-transform duration-200 dark:text-[#a3a3a3] ${
						isOpen ? "rotate-180" : ""
					}`}
				/>
			</button>

			{isOpen ? (
				<div
					ref={listRef}
					role="listbox"
					className="absolute top-[calc(100%+4px)] left-0 z-50 max-h-[min(14rem,50dvh)] w-full min-w-0 overflow-y-auto rounded-[10px] border border-black/10 bg-white p-1 text-[#171717] shadow-[0_8px_30px_rgba(0,0,0,0.12)] sm:max-h-60 sm:rounded-xl sm:p-1.5 dark:border-white/10 dark:bg-[#2a2a2a] dark:text-white dark:shadow-[0_8px_30px_rgba(0,0,0,0.4)]"
					style={{ scrollbarWidth: "thin" }}
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
										? "bg-[#171717] font-medium text-white dark:bg-white dark:text-[#171717]"
										: isHighlighted
											? "bg-black/8 text-[#171717] dark:bg-white/15 dark:text-white"
											: "text-[#6d6d6d] hover:bg-black/5 hover:text-[#171717] dark:text-[#a3a3a3] dark:hover:bg-white/10 dark:hover:text-white",
								)}
							>
								<span className="truncate">{opt.label}</span>
							</button>
						);
					})}
				</div>
			) : null}
		</div>
	);
}
