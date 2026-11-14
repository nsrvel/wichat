"use client";

import type { ReactNode } from "react";
import {
	WiChatDialog,
	WiChatDialogBody,
	WiChatDialogFooter,
	WiChatDialogFooterActions,
	WiChatDialogHeader,
	wichatDialogSubtitleClass,
} from "@/components/ui/wichat-dialog";
import { authSubmitClass, authSubmitGhostClass, authSubmitShapeClass } from "@/lib/form-field-styles";
import { cn } from "@/lib/utils";

const confirmDialogFooterButtonClass = "w-auto min-w-[7.5rem] max-w-[55%] shrink-0 px-8";

const authSubmitDestructiveClass = [
	authSubmitShapeClass,
	confirmDialogFooterButtonClass,
	"bg-destructive text-destructive-foreground hover:bg-destructive/90",
	"focus-visible:border-destructive focus-visible:ring-destructive/30",
].join(" ");

/** Narrower than onboarding/share dialogs — title + body message + footer only. */
export const confirmDialogShellClass =
	"mx-auto w-full min-w-0 max-w-md rounded-[14px] border border-border bg-card py-4 text-left text-card-foreground shadow-elevated sm:rounded-2xl sm:py-6 sm:shadow-elevated-lg max-h-[880px]:py-3.5 max-h-[880px]:sm:py-5";

type ConfirmDialogProps = {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	titleId: string;
	title: ReactNode;
	/** Confirmation copy in the dialog body (no header subtitle). */
	message: ReactNode;
	cancelLabel: ReactNode;
	confirmLabel: ReactNode;
	onConfirm: () => void;
	confirmDisabled?: boolean;
	destructive?: boolean;
};

export function ConfirmDialog({
	open,
	onOpenChange,
	titleId,
	title,
	message,
	cancelLabel,
	confirmLabel,
	onConfirm,
	confirmDisabled,
	destructive,
}: ConfirmDialogProps) {
	return (
		<WiChatDialog
			open={open}
			onOpenChange={onOpenChange}
			titleId={titleId}
			className={confirmDialogShellClass}
		>
			<WiChatDialogHeader title={title} titleId={titleId} className="pb-2 sm:pb-3" />
			<WiChatDialogBody className="pb-2 text-center sm:pb-3">
				<p className={cn(wichatDialogSubtitleClass, "text-pretty")}>{message}</p>
			</WiChatDialogBody>
			<WiChatDialogFooter>
				<WiChatDialogFooterActions
					backLabel={cancelLabel}
					onBack={() => onOpenChange(false)}
					primaryLabel={confirmLabel}
					onPrimary={onConfirm}
					primaryDisabled={confirmDisabled}
					backClassName={cn(authSubmitGhostClass, confirmDialogFooterButtonClass)}
					primaryClassName={
						destructive
							? authSubmitDestructiveClass
							: cn(authSubmitClass, confirmDialogFooterButtonClass)
					}
				/>
			</WiChatDialogFooter>
		</WiChatDialog>
	);
}
