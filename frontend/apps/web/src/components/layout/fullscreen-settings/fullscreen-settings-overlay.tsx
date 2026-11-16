"use client";

import { useEffect, type ReactNode } from "react";
import { createPortal } from "react-dom";
import { groupSettingsOverlayClass } from "@/features/group-settings/lib/group-settings-chrome";
import { useMounted } from "@/hooks/use-mounted";

type FullscreenSettingsOverlayProps = {
	open: boolean;
	titleId: string;
	onRequestClose: () => void;
	children: ReactNode;
};

/** Fullscreen portal dialog with body scroll lock and Escape-key close handling. */
export function FullscreenSettingsOverlay({
	open,
	titleId,
	onRequestClose,
	children,
}: FullscreenSettingsOverlayProps) {
	const mounted = useMounted();

	useEffect(() => {
		if (!open) {
			return;
		}
		const prev = document.body.style.overflow;
		document.body.style.overflow = "hidden";
		return () => {
			document.body.style.overflow = prev;
		};
	}, [open]);

	useEffect(() => {
		if (!open) {
			return;
		}
		const onKeyDown = (event: KeyboardEvent) => {
			if (event.key === "Escape") {
				event.preventDefault();
				onRequestClose();
			}
		};
		window.addEventListener("keydown", onKeyDown);
		return () => window.removeEventListener("keydown", onKeyDown);
	}, [onRequestClose, open]);

	if (!open || !mounted) {
		return null;
	}

	return createPortal(
		<div
			className={groupSettingsOverlayClass}
			role="dialog"
			aria-modal="true"
			aria-labelledby={titleId}
		>
			{children}
		</div>,
		document.body,
	);
}
