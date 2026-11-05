"use client";

import { X } from "lucide-react";
import { Children, type ReactNode, useLayoutEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import {
	authCardInsetXClass,
	groupOnboardingCardShellClass,
} from "@/features/auth/components/auth-shell";
import { cn } from "@/lib/utils";

export const wichatDialogTitleClass =
	"text-pretty text-xl font-bold tracking-tight text-foreground sm:text-2xl";

export const wichatDialogSubtitleClass = "text-sm leading-relaxed text-muted-foreground";

const wichatDialogCloseButtonClass =
	"absolute top-3 right-3 z-20 flex h-9 w-9 items-center justify-center rounded-lg text-muted-foreground transition-colors hover:bg-foreground/5 hover:text-foreground sm:top-4 sm:right-4";

type WiChatDialogProps = {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	/** Element id for `aria-labelledby` on the dialog panel. */
	titleId?: string;
	closeLabel?: string;
	children: ReactNode;
	className?: string;
};

export function WiChatDialog({
	open,
	onOpenChange,
	titleId,
	closeLabel,
	children,
	className,
}: WiChatDialogProps) {
	const { t } = useTranslation();
	const label = closeLabel ?? t("dialog.close");

	if (!open) {
		return null;
	}

	return (
		<div className="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6">
			<button
				type="button"
				className="absolute inset-0 bg-black/50"
				aria-label={label}
				onClick={() => onOpenChange(false)}
			/>
			<div
				role="dialog"
				aria-modal="true"
				aria-labelledby={titleId}
				className={cn(
					groupOnboardingCardShellClass,
					"relative z-10 flex h-auto w-full max-h-[min(90vh,44rem)] flex-col overflow-hidden",
					className,
				)}
			>
				<button
					type="button"
					className={wichatDialogCloseButtonClass}
					aria-label={label}
					onClick={() => onOpenChange(false)}
				>
					<X className="h-5 w-5" strokeWidth={2} aria-hidden />
				</button>
				{children}
			</div>
		</div>
	);
}

type WiChatDialogStepCarouselProps = {
	activeIndex: number;
	children: ReactNode;
	className?: string;
};

/** Horizontal step track inside a dialog (header/body/footer per panel). Slide only; card height snaps to the active step (centering handles backdrop space). */
export function WiChatDialogStepCarousel({
	activeIndex,
	children,
	className,
}: WiChatDialogStepCarouselProps) {
	const panels = Children.toArray(children);
	const count = panels.length;
	const safeIndex = count === 0 ? 0 : Math.min(Math.max(activeIndex, 0), count - 1);
	const trackWidthPercent = count * 100;
	const panelWidthPercent = count === 0 ? 100 : 100 / count;
	const translatePercent = safeIndex * panelWidthPercent;
	const panelRefs = useRef<(HTMLDivElement | null)[]>([]);
	const [viewportHeight, setViewportHeight] = useState<number | null>(null);

	const measurePanel = (index: number) => panelRefs.current[index]?.offsetHeight ?? 0;

	useLayoutEffect(() => {
		if (count === 0) {
			return;
		}

		const frame = requestAnimationFrame(() => {
			const h = measurePanel(safeIndex);
			if (h > 0) {
				setViewportHeight(h);
			}
		});

		return () => cancelAnimationFrame(frame);
	}, [safeIndex, count]);

	useLayoutEffect(() => {
		const onResize = () => {
			const h = measurePanel(safeIndex);
			if (h > 0) {
				setViewportHeight(h);
			}
		};

		window.addEventListener("resize", onResize);

		return () => window.removeEventListener("resize", onResize);
	}, [safeIndex]);

	return (
		<div
			className={cn("w-full overflow-hidden", className)}
			style={viewportHeight === null ? undefined : { height: viewportHeight }}
		>
			<div
				className="flex items-start transition-transform duration-[var(--shell-rail-motion-duration)] ease-[var(--shell-rail-motion-ease)] motion-reduce:transition-none"
				style={{
					width: `${trackWidthPercent}%`,
					transform: `translateX(-${translatePercent}%)`,
				}}
			>
				{panels.map((panel, index) => (
					<div
						key={index}
						ref={(node) => {
							panelRefs.current[index] = node;
						}}
						className="flex min-w-0 shrink-0 flex-col"
						style={{ width: `${panelWidthPercent}%` }}
						aria-hidden={index !== safeIndex}
						{...(index !== safeIndex ? { inert: true as const } : {})}
					>
						{panel}
					</div>
				))}
			</div>
		</div>
	);
}

type WiChatDialogHeaderProps = {
	title: ReactNode;
	subtitle?: ReactNode;
	titleId?: string;
	className?: string;
};

export function WiChatDialogHeader({
	title,
	subtitle,
	titleId,
	className,
}: WiChatDialogHeaderProps) {
	return (
		<header className={cn(authCardInsetXClass, "shrink-0 pb-4 text-center sm:pb-5", className)}>
			<div className="space-y-2">
				<h2 id={titleId} className={wichatDialogTitleClass}>
					{title}
				</h2>
				{subtitle ? <p className={wichatDialogSubtitleClass}>{subtitle}</p> : null}
			</div>
		</header>
	);
}

type WiChatDialogBodyProps = {
	children: ReactNode;
	className?: string;
};

export function WiChatDialogBody({ children, className }: WiChatDialogBodyProps) {
	return (
		<div
			className={cn(
				authCardInsetXClass,
				"min-h-0 max-h-[calc(min(90vh,44rem)-12rem)] overflow-y-auto",
				className,
			)}
		>
			{children}
		</div>
	);
}

type WiChatDialogFooterProps = {
	children: ReactNode;
	className?: string;
};

/** Sticky dialog footer — actions, secondary CTAs, etc. */
export function WiChatDialogFooter({ children, className }: WiChatDialogFooterProps) {
	return (
		<footer className={cn(authCardInsetXClass, "shrink-0 pt-4 pb-1 sm:pt-5", className)}>
			{children}
		</footer>
	);
}

/** Back + primary submit row (create / customize steps). */
export function WiChatDialogFooterActions({
	backLabel,
	onBack,
	primaryLabel,
	onPrimary,
	primaryType = "button",
	primaryDisabled,
	primaryForm,
	backClassName,
	primaryClassName,
}: {
	backLabel: ReactNode;
	onBack: () => void;
	primaryLabel: ReactNode;
	onPrimary?: () => void;
	primaryType?: "button" | "submit";
	primaryDisabled?: boolean;
	primaryForm?: string;
	backClassName?: string;
	primaryClassName?: string;
}) {
	return (
		<div className="flex items-center justify-between gap-3">
			<Button type="button" variant="ghost" className={backClassName} onClick={onBack}>
				{backLabel}
			</Button>
			<Button
				type={primaryType}
				form={primaryForm}
				className={primaryClassName}
				disabled={primaryDisabled}
				onClick={onPrimary}
			>
				{primaryLabel}
			</Button>
		</div>
	);
}
