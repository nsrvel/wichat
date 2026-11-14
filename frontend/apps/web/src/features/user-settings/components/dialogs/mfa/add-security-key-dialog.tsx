"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import {
	browserSupportsWebAuthn,
	startRegistration,
	type PublicKeyCredentialCreationOptionsJSON,
} from "@simplewebauthn/browser";
import { useId, useState } from "react";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { z } from "zod";
import { PasswordField } from "@/components/form/password-field";
import { TextField } from "@/components/form/text-field";
import { Form } from "@/components/ui/form";
import {
	WiChatDialog,
	WiChatDialogBody,
	WiChatDialogFooter,
	WiChatDialogFooterActions,
	WiChatDialogHeader,
} from "@/components/ui/wichat-dialog";
import { applySettingsFormError } from "@/features/user-settings/lib/settings-form-error";
import { translateError } from "@/integrations/i18n/translate-error";
import {
	authFormFooterButtonClass,
	authSubmitClass,
	authSubmitGhostClass,
} from "@/lib/form-field-styles";
import { cn } from "@/lib/utils";
import {
	useBeginWebAuthnRegisterMutation,
	useFinishWebAuthnRegisterMutation,
} from "@/services/user/user.mutation";

const AddSecurityKeySchema = z.object({
	name: z.string().trim().min(1, "validation.security_key_name").max(64),
	currentPassword: z.string().min(1, "validation.password_required"),
});

type AddSecurityKeyDialogProps = {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	onRecoveryCodes: (codes: string[]) => void;
};

function unwrapCreationOptions(raw: unknown): PublicKeyCredentialCreationOptionsJSON | null {
	if (!raw || typeof raw !== "object") {
		return null;
	}
	const record = raw as Record<string, unknown>;
	if (record.publicKey && typeof record.publicKey === "object") {
		return record.publicKey as PublicKeyCredentialCreationOptionsJSON;
	}
	return record as unknown as PublicKeyCredentialCreationOptionsJSON;
}

export function AddSecurityKeyDialog({
	open,
	onOpenChange,
	onRecoveryCodes,
}: AddSecurityKeyDialogProps) {
	const { t } = useTranslation();
	const titleId = useId();
	const [localError, setLocalError] = useState<string | null>(null);
	const [waitingBrowser, setWaitingBrowser] = useState(false);

	const beginMutation = useBeginWebAuthnRegisterMutation();
	const finishMutation = useFinishWebAuthnRegisterMutation();

	const form = useForm<z.infer<typeof AddSecurityKeySchema>>({
		resolver: zodResolver(AddSecurityKeySchema),
		defaultValues: { name: "", currentPassword: "" },
	});

	const handleOpenChange = (next: boolean) => {
		if (!next) {
			setLocalError(null);
			setWaitingBrowser(false);
			beginMutation.reset();
			finishMutation.reset();
			form.reset({ name: "", currentPassword: "" });
		}
		onOpenChange(next);
	};

	const handleSubmit = form.handleSubmit(async ({ name, currentPassword }) => {
		setLocalError(null);
		if (!browserSupportsWebAuthn()) {
			setLocalError(t("user_settings.account.mfa_modal.webauthn_unsupported"));
			return;
		}

		try {
			const beginRes = await beginMutation.mutateAsync({
				current_password: currentPassword,
				name,
			});
			const options = unwrapCreationOptions(beginRes.options);
			if (!options) {
				setLocalError(t("user_settings.account.mfa_modal.webauthn_cancelled"));
				return;
			}

			setWaitingBrowser(true);
			const credential = await startRegistration({ optionsJSON: options });
			setWaitingBrowser(false);

			const finishRes = await finishMutation.mutateAsync({
				challenge_id: beginRes.challenge_id,
				credential,
			});

			toast.success(t("user_settings.account.mfa_modal.add_key_success"));
			handleOpenChange(false);
			if (finishRes.recovery_codes && finishRes.recovery_codes.length > 0) {
				onRecoveryCodes(finishRes.recovery_codes);
			}
		} catch (err) {
			setWaitingBrowser(false);
			if (err instanceof Error && err.name === "NotAllowedError") {
				setLocalError(t("user_settings.account.mfa_modal.webauthn_cancelled"));
				return;
			}
			if (
				applySettingsFormError(form, err, t, {
					passwordField: "currentPassword",
					paramFieldMap: { name: "name" },
				})
			) {
				return;
			}
			setLocalError(
				translateError(t, err) || t("user_settings.account.mfa_modal.webauthn_cancelled"),
			);
		}
	});

	const busy = beginMutation.isPending || waitingBrowser || finishMutation.isPending;

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
				title={t("user_settings.account.mfa_modal.add_key_title")}
				subtitle={t("user_settings.account.mfa_modal.add_key_subtitle")}
			/>
			<Form {...form}>
				<form onSubmit={(e) => void handleSubmit(e)}>
					<WiChatDialogBody className="space-y-4">
						<TextField
							control={form.control}
							name="name"
							label={t("user_settings.account.mfa_modal.add_key_name_label")}
							placeholder={t("user_settings.account.mfa_modal.add_key_name_placeholder")}
							maxLength={64}
							required
						/>

						<PasswordField
							control={form.control}
							name="currentPassword"
							label={t("user_settings.account.current_password")}
							autoComplete="current-password"
							showLabel={t("login.show_password")}
							hideLabel={t("login.hide_password")}
							required
						/>

						{localError ? (
							<p role="alert" className="text-sm text-destructive">
								{localError}
							</p>
						) : null}
					</WiChatDialogBody>
					<WiChatDialogFooter>
						<WiChatDialogFooterActions
							backLabel={t("user_settings.account.mfa_modal.cancel")}
							onBack={() => handleOpenChange(false)}
							primaryLabel={
								waitingBrowser
									? t("user_settings.account.mfa_modal.add_key_waiting")
									: t("user_settings.account.mfa_modal.add_key_register_cta")
							}
							primaryType="submit"
							primaryDisabled={busy}
							backClassName={cn(authSubmitGhostClass, authFormFooterButtonClass)}
							primaryClassName={cn(authSubmitClass, authFormFooterButtonClass)}
						/>
					</WiChatDialogFooter>
				</form>
			</Form>
		</WiChatDialog>
	);
}
