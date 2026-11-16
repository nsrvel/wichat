"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useQuery } from "@tanstack/react-query";
import type { SecurityKey } from "@wichat/contracts/auth/mfa";
import { KeyRound, Pencil, Plus, ShieldCheck, Trash2 } from "lucide-react";
import Image from "next/image";
import { useId, useState } from "react";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { z } from "zod";
import { TextField } from "@/components/form/text-field";
import { Button } from "@/components/ui/button";
import { Form } from "@/components/ui/form";
import {
	WiChatDialog,
	WiChatDialogBody,
	WiChatDialogFooter,
	WiChatDialogFooterActions,
	WiChatDialogHeader,
} from "@/components/ui/wichat-dialog";
import { AddSecurityKeyDialog } from "@/features/user-settings/components/dialogs/mfa/add-security-key-dialog";
import { PasswordConfirmDialog } from "@/features/user-settings/components/dialogs/mfa/disable-mfa-dialog";
import { RecoveryCodesDialog } from "@/features/user-settings/components/dialogs/mfa/recovery-codes-dialog";
import { TotpSetupDialog } from "@/features/user-settings/components/dialogs/mfa/totp-setup-dialog";
import { applySettingsFormError } from "@/features/user-settings/lib/settings-form-error";
import {
	authFormFooterButtonClass,
	authSubmitClass,
	authSubmitGhostClass,
} from "@/lib/form-field-styles";
import { cn } from "@/lib/utils";
import {
	useDeleteSecurityKeyMutation,
	useDisableTotpMutation,
	useRegenerateRecoveryCodesMutation,
	useRenameSecurityKeyMutation,
} from "@/services/user/user.mutation";
import { myMfaStatusQueryOptions } from "@/services/user/user.query";

const RenameSecurityKeySchema = z.object({
	name: z.string().trim().min(1, "validation.security_key_name").max(64),
});

type MfaSettingsDialogProps = {
	open: boolean;
	onOpenChange: (open: boolean) => void;
};

function formatDate(iso?: string): string {
	if (!iso) {
		return "";
	}
	const d = new Date(iso);
	if (Number.isNaN(d.getTime())) {
		return "";
	}
	return d.toLocaleDateString();
}

export function MfaSettingsDialog({ open, onOpenChange }: MfaSettingsDialogProps) {
	const { t } = useTranslation();
	const titleId = useId();
	const renameTitleId = useId();
	const { data: status } = useQuery({
		...myMfaStatusQueryOptions(),
		enabled: open,
	});

	const [totpOpen, setTotpOpen] = useState(false);
	const [addKeyOpen, setAddKeyOpen] = useState(false);
	const [recoveryCodes, setRecoveryCodes] = useState<string[] | null>(null);
	const [disableTotpOpen, setDisableTotpOpen] = useState(false);
	const [regenCodesOpen, setRegenCodesOpen] = useState(false);
	const [deleteKeyTarget, setDeleteKeyTarget] = useState<SecurityKey | null>(null);
	const [renameKeyTarget, setRenameKeyTarget] = useState<SecurityKey | null>(null);
	const renameForm = useForm<z.infer<typeof RenameSecurityKeySchema>>({
		resolver: zodResolver(RenameSecurityKeySchema),
		defaultValues: { name: "" },
	});

	const disableTotp = useDisableTotpMutation();
	const deleteKey = useDeleteSecurityKeyMutation();
	const renameKey = useRenameSecurityKeyMutation();
	const regenCodes = useRegenerateRecoveryCodesMutation();

	const hasAnyMfa = Boolean(status?.totp_enabled) || (status?.security_keys.length ?? 0) > 0;

	return (
		<>
			<WiChatDialog
				elevated
				open={open}
				onOpenChange={onOpenChange}
				titleId={titleId}
				className="max-w-xl"
			>
				<WiChatDialogHeader
					titleId={titleId}
					title={t("user_settings.account.mfa_modal.title")}
					subtitle={t("user_settings.account.mfa_modal.subtitle")}
					className="pb-2 sm:pb-3"
				/>
				<WiChatDialogBody className="space-y-5 pb-2">
					<div className="flex justify-center">
						<Image
							src="/mfa.png"
							alt=""
							width={160}
							height={120}
							className="h-24 w-auto object-contain select-none sm:h-28"
							priority
						/>
					</div>

					{/* Security Keys section */}
					<div className="space-y-3">
						<div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
							<div className="min-w-0 space-y-1">
								<h3 className="text-sm font-semibold text-foreground sm:text-base">
									{t("user_settings.account.mfa_modal.security_keys_title")}
								</h3>
								<p className="text-xs leading-relaxed text-muted-foreground sm:text-sm">
									{t("user_settings.account.mfa_modal.security_keys_desc")}
								</p>
							</div>
							<Button
								type="button"
								variant="secondary"
								size="lg"
								className="min-w-[140px] shrink-0 px-4 text-[13px]"
								onClick={() => setAddKeyOpen(true)}
							>
								<Plus className="mr-1.5 size-4" />
								{t("user_settings.account.mfa_modal.add_security_key")}
							</Button>
						</div>

						{status && status.security_keys.length > 0 ? (
							<div className="divide-y divide-border rounded-lg border border-border bg-muted/20">
								{status.security_keys.map((key) => (
									<div
										key={key.id}
										className="flex items-center justify-between gap-3 px-3.5 py-2.5"
									>
										<div className="flex min-w-0 items-center gap-3">
											<KeyRound className="h-4 w-4 shrink-0 text-muted-foreground" />
											<div className="min-w-0">
												<p className="truncate text-sm font-medium text-foreground">{key.name}</p>
												<p className="text-xs text-muted-foreground">
													{t("user_settings.account.mfa_modal.key_added_at", {
														date: formatDate(key.created_at),
													})}
													{" · "}
													{key.last_used_at
														? t("user_settings.account.mfa_modal.key_last_used", {
																date: formatDate(key.last_used_at),
															})
														: t("user_settings.account.mfa_modal.key_never_used")}
												</p>
											</div>
										</div>
										<div className="flex shrink-0 items-center gap-1">
											<Button
												type="button"
												variant="ghost"
												size="icon"
												aria-label={t("user_settings.account.mfa_modal.rename")}
												onClick={() => {
													renameKey.reset();
													setRenameKeyTarget(key);
													renameForm.reset({ name: key.name });
												}}
											>
												<Pencil className="h-4 w-4" />
											</Button>
											<Button
												type="button"
												variant="ghost"
												size="icon"
												aria-label={t("user_settings.account.mfa_modal.remove")}
												className="text-destructive hover:text-destructive"
												onClick={() => {
													deleteKey.reset();
													setDeleteKeyTarget(key);
												}}
											>
												<Trash2 className="h-4 w-4" />
											</Button>
										</div>
									</div>
								))}
							</div>
						) : null}
					</div>

					<hr className="border-border" />

					{/* Authenticator App section */}
					<div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
						<div className="min-w-0 space-y-1">
							<div className="flex flex-wrap items-center gap-2">
								<h3 className="text-sm font-semibold text-foreground sm:text-base">
									{t("user_settings.account.mfa_modal.authenticator_title")}
								</h3>
								{status?.totp_enabled ? (
									<span className="inline-flex items-center gap-1 rounded-full bg-emerald-500/15 px-2 py-0.5 text-xs font-medium text-emerald-500">
										<ShieldCheck className="h-3 w-3" />
										{t("user_settings.account.mfa_enabled")}
									</span>
								) : null}
							</div>
							<p className="text-xs leading-relaxed text-muted-foreground sm:text-sm">
								{t("user_settings.account.mfa_modal.authenticator_desc")}
							</p>
						</div>
						{status?.totp_enabled ? (
							<Button
								type="button"
								variant="secondary"
								size="lg"
								className="min-w-[140px] shrink-0 px-4 text-[13px] text-destructive hover:text-destructive"
								onClick={() => {
									disableTotp.reset();
									setDisableTotpOpen(true);
								}}
							>
								{t("user_settings.account.mfa_modal.disable_authenticator")}
							</Button>
						) : (
							<Button
								type="button"
								variant="secondary"
								size="lg"
								className="min-w-[140px] shrink-0 px-4 text-[13px]"
								onClick={() => setTotpOpen(true)}
							>
								{t("user_settings.account.mfa_modal.enable_authenticator")}
							</Button>
						)}
					</div>

					{/* Backup Recovery Codes section */}
					{hasAnyMfa ? (
						<>
							<hr className="border-border" />
							<div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
								<div className="min-w-0 space-y-0.5">
									<h3 className="text-sm font-semibold text-foreground">
										{t("user_settings.account.mfa_modal.recovery_codes_title")}
									</h3>
									<p className="text-xs text-muted-foreground">
										{t("user_settings.account.mfa_modal.recovery_codes_remaining", {
											count: status?.recovery_codes_remaining ?? 0,
										})}
									</p>
								</div>
								<Button
									type="button"
									variant="secondary"
									size="lg"
									className="min-w-[140px] shrink-0 px-4 text-[13px]"
									onClick={() => {
										regenCodes.reset();
										setRegenCodesOpen(true);
									}}
								>
									{t("user_settings.account.mfa_modal.view_regenerate_codes")}
								</Button>
							</div>
						</>
					) : null}
				</WiChatDialogBody>
			</WiChatDialog>

			<TotpSetupDialog
				open={totpOpen}
				onOpenChange={setTotpOpen}
				onRecoveryCodes={(codes) => setRecoveryCodes(codes)}
			/>

			<AddSecurityKeyDialog
				open={addKeyOpen}
				onOpenChange={setAddKeyOpen}
				onRecoveryCodes={(codes) => setRecoveryCodes(codes)}
			/>

			<RecoveryCodesDialog
				open={recoveryCodes !== null}
				codes={recoveryCodes ?? []}
				onClose={() => setRecoveryCodes(null)}
			/>

			<PasswordConfirmDialog
				open={disableTotpOpen}
				onOpenChange={setDisableTotpOpen}
				title={t("user_settings.account.mfa_modal.disable_totp_title")}
				subtitle={t("user_settings.account.mfa_modal.disable_totp_subtitle")}
				confirmLabel={t("user_settings.account.mfa_modal.disable_authenticator")}
				destructive
				isPending={disableTotp.isPending}
				error={disableTotp.error}
				onConfirm={(currentPassword) => {
					disableTotp.mutate(currentPassword, {
						onSuccess: () => {
							toast.success(t("user_settings.account.mfa_modal.totp_disabled_success"));
							setDisableTotpOpen(false);
						},
					});
				}}
			/>

			<PasswordConfirmDialog
				open={deleteKeyTarget !== null}
				onOpenChange={(next) => {
					if (!next) {
						setDeleteKeyTarget(null);
					}
				}}
				title={t("user_settings.account.mfa_modal.remove_key_title")}
				subtitle={t("user_settings.account.mfa_modal.remove_key_subtitle", {
					name: deleteKeyTarget?.name ?? "",
				})}
				confirmLabel={t("user_settings.account.mfa_modal.remove_key_confirm")}
				destructive
				isPending={deleteKey.isPending}
				error={deleteKey.error}
				onConfirm={(currentPassword) => {
					if (!deleteKeyTarget) {
						return;
					}
					deleteKey.mutate(
						{ id: deleteKeyTarget.id, current_password: currentPassword },
						{
							onSuccess: () => {
								toast.success(t("user_settings.account.mfa_modal.remove_key_success"));
								setDeleteKeyTarget(null);
							},
						},
					);
				}}
			/>

			<PasswordConfirmDialog
				open={regenCodesOpen}
				onOpenChange={setRegenCodesOpen}
				title={t("user_settings.account.mfa_modal.regen_codes_title")}
				subtitle={t("user_settings.account.mfa_modal.regen_codes_subtitle")}
				confirmLabel={t("user_settings.account.mfa_modal.regen_codes_cta")}
				isPending={regenCodes.isPending}
				error={regenCodes.error}
				onConfirm={(currentPassword) => {
					regenCodes.mutate(currentPassword, {
						onSuccess: (res) => {
							setRegenCodesOpen(false);
							setRecoveryCodes(res.recovery_codes);
						},
					});
				}}
			/>

			<WiChatDialog
				elevated
				open={renameKeyTarget !== null}
				titleId={renameTitleId}
				onOpenChange={(next) => {
					if (!next) {
						setRenameKeyTarget(null);
					}
				}}
				className="max-w-md"
			>
				<WiChatDialogHeader
					titleId={renameTitleId}
					title={t("user_settings.account.mfa_modal.rename_key_title")}
				/>
				<Form {...renameForm}>
					<form
						onSubmit={renameForm.handleSubmit(({ name }) => {
							if (!renameKeyTarget) {
								return;
							}
							renameKey.mutate(
								{ id: renameKeyTarget.id, name },
								{
									onSuccess: () => {
										toast.success(t("user_settings.account.mfa_modal.rename_key_success"));
										setRenameKeyTarget(null);
									},
									onError: (err) => {
										applySettingsFormError(renameForm, err, t, {
											paramFieldMap: { name: "name" },
										});
									},
								},
							);
						})}
					>
						<WiChatDialogBody className="space-y-4">
							<TextField
								control={renameForm.control}
								name="name"
								label={t("user_settings.account.mfa_modal.add_key_name_label")}
								placeholder={t("user_settings.account.mfa_modal.add_key_name_placeholder")}
								maxLength={64}
								required
							/>
						</WiChatDialogBody>
						<WiChatDialogFooter>
							<WiChatDialogFooterActions
								backLabel={t("user_settings.account.mfa_modal.cancel")}
								onBack={() => setRenameKeyTarget(null)}
								primaryLabel={t("user_settings.account.mfa_modal.save")}
								primaryType="submit"
								primaryDisabled={renameKey.isPending}
								backClassName={cn(authSubmitGhostClass, authFormFooterButtonClass)}
								primaryClassName={cn(authSubmitClass, authFormFooterButtonClass)}
							/>
						</WiChatDialogFooter>
					</form>
				</Form>
			</WiChatDialog>
		</>
	);
}
