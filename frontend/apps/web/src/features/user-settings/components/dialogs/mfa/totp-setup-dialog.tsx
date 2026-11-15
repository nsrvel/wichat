"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import type { BeginTotpSetupResponse } from "@wichat/contracts/auth/mfa";
import { Check, Copy } from "lucide-react";
import Image from "next/image";
import QRCode from "qrcode";
import { useId, useState } from "react";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { z } from "zod";
import { OtpField } from "@/components/form/otp-field";
import { PasswordField } from "@/components/form/password-field";
import { Button } from "@/components/ui/button";
import { Form } from "@/components/ui/form";
import {
	WiChatDialog,
	WiChatDialogBody,
	WiChatDialogFooter,
	WiChatDialogFooterActions,
	WiChatDialogHeader,
	WiChatDialogStepCarousel,
} from "@/components/ui/wichat-dialog";
import { applySettingsFormError } from "@/features/user-settings/lib/settings-form-error";
import {
	authFieldLabelClass,
	authFormFooterButtonClass,
	authSubmitClass,
	authSubmitGhostClass,
} from "@/lib/form-field-styles";
import { cn } from "@/lib/utils";
import {
	useBeginTotpSetupMutation,
	useConfirmTotpSetupMutation,
} from "@/services/user/user.mutation";

const TotpBeginSchema = z.object({
	currentPassword: z.string().min(1, "validation.password_required"),
});

const TotpConfirmSchema = z.object({
	code: z.string().regex(/^\d{6}$/, "validation.otp"),
});

type TotpSetupDialogProps = {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	onRecoveryCodes: (codes: string[]) => void;
};

export function TotpSetupDialog({
	open,
	onOpenChange,
	onRecoveryCodes,
}: TotpSetupDialogProps) {
	const { t } = useTranslation();
	const titleId = useId();
	const [step, setStep] = useState<0 | 1 | 2>(0);
	const [setup, setSetup] = useState<BeginTotpSetupResponse | null>(null);
	const [qrDataUrl, setQrDataUrl] = useState<string | null>(null);
	const [copiedSecret, setCopiedSecret] = useState(false);

	const beginMutation = useBeginTotpSetupMutation();
	const confirmMutation = useConfirmTotpSetupMutation();

	const beginForm = useForm<z.infer<typeof TotpBeginSchema>>({
		resolver: zodResolver(TotpBeginSchema),
		defaultValues: { currentPassword: "" },
	});

	const form = useForm<z.infer<typeof TotpConfirmSchema>>({
		resolver: zodResolver(TotpConfirmSchema),
		defaultValues: { code: "" },
	});

	const handleOpenChange = (next: boolean) => {
		if (!next) {
			setStep(0);
			setSetup(null);
			setQrDataUrl(null);
			setCopiedSecret(false);
			beginMutation.reset();
			confirmMutation.reset();
			beginForm.reset({ currentPassword: "" });
			form.reset({ code: "" });
		}
		onOpenChange(next);
	};

	const handleCopySecret = async () => {
		if (!setup?.secret) {
			return;
		}
		try {
			await navigator.clipboard.writeText(setup.secret);
			setCopiedSecret(true);
			setTimeout(() => setCopiedSecret(false), 2000);
		} catch {
			// ignore clipboard errors
		}
	};

	const handleBegin = beginForm.handleSubmit(({ currentPassword }) => {
		beginMutation.mutate(currentPassword, {
			onSuccess: (data) => {
				setSetup(data);
				setStep(1);
				void QRCode.toDataURL(data.otpauth_uri, {
					width: 180,
					margin: 1,
				}).then(setQrDataUrl);
			},
			onError: (err) => {
				applySettingsFormError(beginForm, err, t, {
					passwordField: "currentPassword",
				});
			},
		});
	});

	const handleConfirm = form.handleSubmit(({ code }) => {
		if (!setup) {
			return;
		}
		confirmMutation.mutate(
			{ setup_token: setup.setup_token, code },
			{
				onSuccess: (res) => {
					toast.success(t("user_settings.account.mfa_modal.totp_enabled_success"));
					handleOpenChange(false);
					if (res.recovery_codes.length > 0) {
						onRecoveryCodes(res.recovery_codes);
					}
				},
				onError: (err) => {
					applySettingsFormError(form, err, t, {
						codeMap: {
							"auth.reset_code_invalid": "code",
							"error.invalid_input": "code",
						},
					});
				},
			},
		);
	});

	return (
		<WiChatDialog
			elevated
			open={open}
			onOpenChange={handleOpenChange}
			titleId={titleId}
			className="max-w-md"
		>
			<WiChatDialogHeader
				titleId={titleId}
				title={t("user_settings.account.mfa_modal.totp_setup_title")}
				subtitle={
					step === 0
						? t("user_settings.account.mfa_modal.totp_password_subtitle")
						: step === 1
							? t("user_settings.account.mfa_modal.totp_scan_subtitle")
							: t("user_settings.account.mfa_modal.totp_code_label")
				}
			/>

			<WiChatDialogStepCarousel activeIndex={step}>
				<Form {...beginForm}>
					<form onSubmit={handleBegin}>
						<WiChatDialogBody className="space-y-4">
							<PasswordField
								control={beginForm.control}
								name="currentPassword"
								label={t("user_settings.account.current_password")}
								autoComplete="current-password"
								showLabel={t("login.show_password")}
								hideLabel={t("login.hide_password")}
								required
							/>
						</WiChatDialogBody>
						<WiChatDialogFooter>
							<WiChatDialogFooterActions
								backLabel={t("user_settings.account.mfa_modal.cancel")}
								onBack={() => handleOpenChange(false)}
								primaryLabel={t("user_settings.account.mfa_modal.continue")}
								primaryType="submit"
								primaryDisabled={beginMutation.isPending}
								backClassName={cn(authSubmitGhostClass, authFormFooterButtonClass)}
								primaryClassName={cn(authSubmitClass, authFormFooterButtonClass)}
							/>
						</WiChatDialogFooter>
					</form>
				</Form>

				<div>
					<WiChatDialogBody className="space-y-4">
						{qrDataUrl ? (
							<div className="flex justify-center">
								<div className="rounded-xl border border-border bg-white p-3">
									<Image
										src={qrDataUrl}
										alt="Authenticator QR Code"
										width={180}
										height={180}
										unoptimized
									/>
								</div>
							</div>
						) : null}

						{setup?.secret ? (
							<div className="space-y-1.5">
								<p className={authFieldLabelClass}>
									{t("user_settings.account.mfa_modal.totp_manual_key_label")}
								</p>
								<div className="flex items-center gap-2 rounded-lg border border-border bg-muted/40 py-1.5 pr-1.5 pl-3">
									<code className="flex-1 break-all font-mono text-xs text-foreground">
										{setup.secret}
									</code>
									<Button
										type="button"
										variant="ghost"
										size="icon"
										aria-label={
											copiedSecret
												? t("user_settings.account.mfa_modal.totp_copied")
												: t("user_settings.account.mfa_modal.totp_copy_key")
										}
										onClick={() => void handleCopySecret()}
									>
										{copiedSecret ? (
											<Check className="size-4" aria-hidden />
										) : (
											<Copy className="size-4" aria-hidden />
										)}
									</Button>
								</div>
							</div>
						) : null}
					</WiChatDialogBody>
					<WiChatDialogFooter>
						<WiChatDialogFooterActions
							backLabel={t("user_settings.account.mfa_modal.back")}
							onBack={() => setStep(0)}
							primaryLabel={t("user_settings.account.mfa_modal.continue")}
							onPrimary={() => setStep(2)}
							backClassName={cn(authSubmitGhostClass, authFormFooterButtonClass)}
							primaryClassName={cn(authSubmitClass, authFormFooterButtonClass)}
						/>
					</WiChatDialogFooter>
				</div>

				<Form {...form}>
					<form onSubmit={handleConfirm}>
						<WiChatDialogBody className="space-y-4">
							<OtpField
								control={form.control}
								name="code"
								label={t("user_settings.account.mfa_modal.totp_code_label")}
								length={6}
								hideLabel
								required
							/>
						</WiChatDialogBody>
						<WiChatDialogFooter>
							<WiChatDialogFooterActions
								backLabel={t("user_settings.account.mfa_modal.back")}
								onBack={() => {
									confirmMutation.reset();
									setStep(1);
								}}
								primaryLabel={t("user_settings.account.mfa_modal.totp_verify_enable")}
								primaryType="submit"
								primaryDisabled={confirmMutation.isPending}
								backClassName={cn(authSubmitGhostClass, authFormFooterButtonClass)}
								primaryClassName={cn(authSubmitClass, authFormFooterButtonClass)}
							/>
						</WiChatDialogFooter>
					</form>
				</Form>
			</WiChatDialogStepCarousel>
		</WiChatDialog>
	);
}
