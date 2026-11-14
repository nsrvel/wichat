"use client";

import { Check, Copy, Download } from "lucide-react";
import { useId, useState } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import {
	WiChatDialog,
	WiChatDialogBody,
	WiChatDialogFooter,
	WiChatDialogHeader,
} from "@/components/ui/wichat-dialog";
import { authFormFooterButtonClass, authSubmitClass } from "@/lib/form-field-styles";
import { cn } from "@/lib/utils";

type RecoveryCodesDialogProps = {
	open: boolean;
	codes: string[];
	onClose: () => void;
};

export function RecoveryCodesDialog({ open, codes, onClose }: RecoveryCodesDialogProps) {
	const { t } = useTranslation();
	const titleId = useId();
	const [copied, setCopied] = useState(false);

	const handleCopyAll = async () => {
		try {
			await navigator.clipboard.writeText(codes.join("\n"));
			setCopied(true);
			setTimeout(() => setCopied(false), 2000);
		} catch {
			// ignore clipboard failures
		}
	};

	const handleDownload = () => {
		const content = [
			"WiChat Backup Recovery Codes",
			"Each code can only be used once.",
			"",
			...codes,
			"",
		].join("\n");
		const blob = new Blob([content], { type: "text/plain;charset=utf-8" });
		const url = URL.createObjectURL(blob);
		const a = document.createElement("a");
		a.href = url;
		a.download = "wichat-recovery-codes.txt";
		a.click();
		URL.revokeObjectURL(url);
	};

	return (
		<WiChatDialog
			elevated
			open={open}
			titleId={titleId}
			className="max-w-md"
			onOpenChange={(next) => {
				if (!next) {
					onClose();
				}
			}}
		>
			<WiChatDialogHeader
				titleId={titleId}
				title={t("user_settings.account.mfa_modal.recovery_save_title")}
				subtitle={t("user_settings.account.mfa_modal.recovery_save_subtitle")}
			/>
			<WiChatDialogBody className="space-y-4 pb-2">
				<div className="grid grid-cols-2 gap-2 rounded-xl border border-border bg-muted/40 p-3">
					{codes.map((code) => (
						<div
							key={code}
							className="rounded-lg bg-background px-2.5 py-1.5 text-center font-mono text-sm tracking-wider text-foreground select-all"
						>
							{code}
						</div>
					))}
				</div>

				<div className="grid grid-cols-2 gap-2.5">
					<Button
						type="button"
						variant="secondary"
						size="lg"
						className="w-full text-[13px]"
						onClick={() => void handleCopyAll()}
					>
						{copied ? (
							<Check className="mr-1.5 size-4" />
						) : (
							<Copy className="mr-1.5 size-4" />
						)}
						{copied
							? t("user_settings.account.mfa_modal.totp_copied")
							: t("user_settings.account.mfa_modal.recovery_copy_all")}
					</Button>
					<Button
						type="button"
						variant="secondary"
						size="lg"
						className="w-full text-[13px]"
						onClick={handleDownload}
					>
						<Download className="mr-1.5 size-4" />
						{t("user_settings.account.mfa_modal.recovery_download")}
					</Button>
				</div>
			</WiChatDialogBody>
			<WiChatDialogFooter>
				<div className="flex justify-end">
					<Button
						type="button"
						className={cn(authSubmitClass, authFormFooterButtonClass)}
						onClick={onClose}
					>
						{t("user_settings.account.mfa_modal.recovery_saved_confirm")}
					</Button>
				</div>
			</WiChatDialogFooter>
		</WiChatDialog>
	);
}
