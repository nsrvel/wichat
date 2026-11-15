"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import type { AuthChallenge } from "@wichat/contracts/auth";
import Image from "next/image";
import { useId, useState } from "react";
import { useForm, useWatch } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { z } from "zod";
import { OtpField } from "@/components/form/otp-field";
import { PasswordField } from "@/components/form/password-field";
import { TextField } from "@/components/form/text-field";
import { Form } from "@/components/ui/form";
import { applySettingsFormError } from "@/features/user-settings/lib/settings-form-error";
import {
	WiChatDialog,
	WiChatDialogBody,
	WiChatDialogFooter,
	WiChatDialogFooterActions,
	WiChatDialogHeader,
	WiChatDialogStepCarousel,
} from "@/components/ui/wichat-dialog";
import { ForgotPasswordResendCode } from "@/features/auth/forgot-password/forgot-password-resend-code";
import { translateError } from "@/integrations/i18n/translate-error";
import {
	authFieldLabelClass,
	authFormFooterButtonClass,
	authInputClass,
	authSubmitClass,
	authSubmitGhostClass,
} from "@/lib/form-field-styles";
import { cn } from "@/lib/utils";
import {
	useCompleteMyEmailChangeMutation,
	useRequestMyEmailChangeMutation,
	useVerifyMyEmailChangeMutation,
} from "@/services/user/user.mutation";

const OTP_LENGTH = 6;
const OTP_SCHEMA = z.object({
	code: z.string().regex(new RegExp(`^[A-Z0-9]{${OTP_LENGTH}}$`), "validation.otp"),
});
const EMAIL_SCHEMA = z.email("validation.email");

const CHANGE_EMAIL_REASONS = ["staff_asked", "someone_asked", "new_email", "other"] as const;

type ChangeEmailReason = (typeof CHANGE_EMAIL_REASONS)[number];

type ChangeEmailDialogProps = {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	currentEmail: string;
};

function ChangeEmailDialogBody({
	currentEmail,
	onClose,
	titleId,
}: {
	currentEmail: string;
	onClose: () => void;
	titleId: string;
}) {
	const { t, i18n } = useTranslation();
	const otpFormId = `${titleId}-otp-form`;
	const emailFormId = `${titleId}-email-form`;

	const [stepIndex, setStepIndex] = useState(0);
	const [challenge, setChallenge] = useState<AuthChallenge | null>(null);
	const [changeToken, setChangeToken] = useState("");
	const [reason, setReason] = useState<ChangeEmailReason | null>(null);

	const requestMutation = useRequestMyEmailChangeMutation();
	const verifyMutation = useVerifyMyEmailChangeMutation();
	const completeMutation = useCompleteMyEmailChangeMutation();

	const otpForm = useForm<{ code: string }>({
		resolver: zodResolver(OTP_SCHEMA),
		defaultValues: { code: "" },
	});
	const otpCode = useWatch({ control: otpForm.control, name: "code" });

	const emailFormSchema = z.object({
		newEmail: EMAIL_SCHEMA,
		currentPassword: z.string().min(1, "validation.password_required"),
	});
	type EmailFormValues = z.infer<typeof emailFormSchema>;
	const emailForm = useForm<EmailFormValues>({
		resolver: zodResolver(emailFormSchema),
		defaultValues: { newEmail: "", currentPassword: "" },
	});

	const requestOtp = async (advance: boolean) => {
		try {
			const res = await requestMutation.mutateAsync({ locale: i18n.language });
			setChallenge(res);
			if (advance) {
				setStepIndex(1);
			}
		} catch (err) {
			toast.error(translateError(t, err));
		}
	};

	const handleSendOtp = () => {
		if (challenge && challenge.resend_after_seconds > 0) {
			setStepIndex(1);
			return;
		}
		void requestOtp(true);
	};

	const handleResendOtp = () => {
		void requestOtp(false);
	};

	const handleVerifyOtp = otpForm.handleSubmit(async ({ code }) => {
		if (!challenge) {
			return;
		}
		try {
			const res = await verifyMutation.mutateAsync({
				challenge_id: challenge.challenge_id,
				code,
			});
			setChangeToken(res.change_token);
			setStepIndex(2);
		} catch (err) {
			applySettingsFormError(otpForm, err, t, {
				codeMap: { "auth.reset_code_invalid": "code" },
			});
		}
	});

	const handleSubmitEmail = emailForm.handleSubmit(async (values) => {
		if (!changeToken || completeMutation.isPending) {
			return;
		}
		try {
			await completeMutation.mutateAsync({
				change_token: changeToken,
				new_email: values.newEmail.trim(),
				current_password: values.currentPassword,
				reason: reason ?? "",
			});
			toast.success(t("user_settings.account.change_email_success"));
			onClose();
		} catch (err) {
			applySettingsFormError(emailForm, err, t, {
				passwordField: "currentPassword",
				codeMap: { "auth.email_taken": "newEmail" },
				paramFieldMap: {
					new_email: "newEmail",
					current_password: "currentPassword",
				},
			});
		}
	});

	return (
		<WiChatDialogStepCarousel activeIndex={stepIndex}>
			{/* Step 0: Verify old email */}
			<div>
				<WiChatDialogHeader
					title={t("user_settings.account.change_email_verify_title")}
					titleId={stepIndex === 0 ? titleId : undefined}
					className="pb-2 sm:pb-3"
				/>
				<WiChatDialogBody className="flex flex-col items-center pb-5 text-center">
					<Image
						src="/send-email.png"
						alt=""
						width={160}
						height={120}
						className="mb-3 h-28 w-auto object-contain select-none"
						priority
					/>
					<p className="max-w-[320px] text-sm leading-relaxed text-muted-foreground">
						{t("user_settings.account.change_email_verify_desc", {
							email: currentEmail || t("user_settings.account.not_set"),
						})}
					</p>
				</WiChatDialogBody>
				<WiChatDialogFooter>
					<WiChatDialogFooterActions
						backLabel={t("user_settings.account.change_username_cancel")}
						onBack={onClose}
						primaryLabel={
							requestMutation.isPending
								? t("forgot_password.sending_code")
								: t("user_settings.account.change_email_send_code")
						}
						onPrimary={handleSendOtp}
						primaryDisabled={requestMutation.isPending}
						backClassName={cn(authSubmitGhostClass, authFormFooterButtonClass)}
						primaryClassName={cn(authSubmitClass, authFormFooterButtonClass)}
					/>
				</WiChatDialogFooter>
			</div>

			{/* Step 1: Enter OTP */}
			<div>
				<WiChatDialogHeader
					title={t("user_settings.account.change_email_otp_title")}
					subtitle={t("login.code_lead", {
						email: challenge?.email_hint || currentEmail || t("user_settings.account.not_set"),
					})}
					titleId={stepIndex === 1 ? titleId : undefined}
				/>
				<WiChatDialogBody className="pt-2 pb-5">
					<Form {...otpForm}>
						<form id={otpFormId} onSubmit={handleVerifyOtp} className="space-y-4">
							<OtpField
								control={otpForm.control}
								name="code"
								length={OTP_LENGTH}
								label={t("forgot_password.otp_field_label")}
								required
								hideLabel
							/>
							{challenge ? (
								<ForgotPasswordResendCode
									key={challenge.challenge_id}
									cooldownSec={challenge.resend_after_seconds}
									isResending={requestMutation.isPending}
									onResend={handleResendOtp}
								/>
							) : null}
						</form>
					</Form>
				</WiChatDialogBody>
				<WiChatDialogFooter>
					<WiChatDialogFooterActions
						backLabel={t("user_settings.account.change_email_back")}
						onBack={() => setStepIndex(0)}
						primaryLabel={t("user_settings.account.change_email_verify_action")}
						primaryType="submit"
						primaryForm={otpFormId}
						primaryDisabled={otpCode.length !== OTP_LENGTH || verifyMutation.isPending}
						backClassName={cn(authSubmitGhostClass, authFormFooterButtonClass)}
						primaryClassName={cn(authSubmitClass, authFormFooterButtonClass)}
					/>
				</WiChatDialogFooter>
			</div>

			{/* Step 2: Reason */}
			<div>
				<WiChatDialogHeader
					title={t("user_settings.account.change_email_reason_title")}
					subtitle={t("user_settings.account.change_email_reason_subtitle")}
					titleId={stepIndex === 2 ? titleId : undefined}
				/>
				<WiChatDialogBody className="pb-5">
					<div
						role="radiogroup"
						aria-label={t("user_settings.account.change_email_reason_title")}
						className="space-y-1"
					>
						{CHANGE_EMAIL_REASONS.map((opt) => {
							const selected = reason === opt;
							return (
								<button
									key={opt}
									type="button"
									role="radio"
									aria-checked={selected}
									onClick={() => setReason(opt)}
									className={cn(
										"flex w-full items-center gap-3 rounded-lg px-2 py-2 text-left text-sm transition-colors hover:bg-muted/40",
										selected ? "bg-muted/35 text-foreground" : "text-muted-foreground hover:text-foreground",
									)}
								>
									<span
										className={cn(
											"flex size-[18px] shrink-0 items-center justify-center rounded-full border-2 transition-colors",
											selected ? "border-primary bg-primary" : "border-muted-foreground/50",
										)}
										aria-hidden
									>
										{selected ? (
											<span className="size-1.5 rounded-full bg-primary-foreground" />
										) : null}
									</span>
									<span className="min-w-0 flex-1 font-medium text-foreground">
										{t(`user_settings.account.change_email_reason_${opt}`)}
									</span>
								</button>
							);
						})}
					</div>
				</WiChatDialogBody>
				<WiChatDialogFooter>
					<WiChatDialogFooterActions
						backLabel={t("user_settings.account.change_username_cancel")}
						onBack={onClose}
						primaryLabel={t("user_settings.account.change_email_continue")}
						onPrimary={() => setStepIndex(3)}
						primaryDisabled={reason === null}
						backClassName={cn(authSubmitGhostClass, authFormFooterButtonClass)}
						primaryClassName={cn(authSubmitClass, authFormFooterButtonClass)}
					/>
				</WiChatDialogFooter>
			</div>

			{/* Step 3: Enter new email + current password */}
			<div>
				<WiChatDialogHeader
					title={t("user_settings.account.change_email_new_title")}
					subtitle={t("user_settings.account.change_email_new_subtitle")}
					titleId={stepIndex === 3 ? titleId : undefined}
				/>
				<WiChatDialogBody className="pb-5">
					<Form {...emailForm}>
						<form id={emailFormId} onSubmit={handleSubmitEmail} className="space-y-4">
							<TextField
								control={emailForm.control}
								name="newEmail"
								type="email"
								label={t("user_settings.account.email")}
								placeholder="name@example.com"
								autoComplete="email"
								required
								labelClassName={authFieldLabelClass}
								inputClassName={authInputClass}
							/>

							<PasswordField
								control={emailForm.control}
								name="currentPassword"
								label={t("user_settings.account.current_password")}
								autoComplete="current-password"
								required
								showLabel={t("login.show_password")}
								hideLabel={t("login.hide_password")}
								labelClassName={authFieldLabelClass}
								inputClassName={authInputClass}
							/>
						</form>
					</Form>
				</WiChatDialogBody>
				<WiChatDialogFooter>
					<WiChatDialogFooterActions
						backLabel={t("user_settings.account.change_email_back")}
						onBack={() => setStepIndex(2)}
						primaryLabel={t("user_settings.account.change_username_done")}
						primaryType="submit"
						primaryForm={emailFormId}
						primaryDisabled={!changeToken || completeMutation.isPending}
						backClassName={cn(authSubmitGhostClass, authFormFooterButtonClass)}
						primaryClassName={cn(authSubmitClass, authFormFooterButtonClass)}
					/>
				</WiChatDialogFooter>
			</div>
		</WiChatDialogStepCarousel>
	);
}

export function ChangeEmailDialog({ open, onOpenChange, currentEmail }: ChangeEmailDialogProps) {
	const titleId = useId();

	return (
		<WiChatDialog
			elevated
			open={open}
			onOpenChange={onOpenChange}
			titleId={titleId}
			className="max-w-[440px]"
		>
			<ChangeEmailDialogBody
				currentEmail={currentEmail}
				titleId={titleId}
				onClose={() => onOpenChange(false)}
			/>
		</WiChatDialog>
	);
}
