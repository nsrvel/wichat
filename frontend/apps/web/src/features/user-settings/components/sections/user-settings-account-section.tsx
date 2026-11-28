"use client";

import { useQuery } from "@tanstack/react-query";
import type { User } from "@wichat/contracts/user";
import { ChevronRight } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import {
	AccountStatusDialog,
	type AccountStatusDialogMode,
} from "@/features/user-settings/components/dialogs/account-status-dialog";
import { ChangeEmailDialog } from "@/features/user-settings/components/dialogs/change-email-dialog";
import { ChangePasswordDialog } from "@/features/user-settings/components/dialogs/change-password-dialog";
import { ChangeUsernameDialog } from "@/features/user-settings/components/dialogs/change-username-dialog";
import { MfaSettingsDialog } from "@/features/user-settings/components/dialogs/mfa/mfa-settings-dialog";
import type { UserSettingsSectionId } from "@/features/user-settings/lib/user-settings-nav";
import {
	settingsSectionDividerLineClass,
	settingsSectionDividerWrapClass,
} from "@/lib/settings-chrome";
import { isMfaFeatureEnabled } from "@/features/auth/lib/mfa-policy";
import { instancePolicyQueryOpts } from "@/services/auth/auth.query";
import { myMfaStatusQueryOptions } from "@/services/user/user.query";

type UserSettingsAccountSectionProps = {
	user: User | null | undefined;
	onNavigateSection: (section: UserSettingsSectionId) => void;
	onAccountClosed?: () => void;
};

const editButtonClass = "px-3.5 text-[13px]";
const statusActionButtonClass = "min-w-[140px] px-4 text-[13px]";

function maskEmail(email: string): string {
	const trimmed = email.trim();
	const atIdx = trimmed.indexOf("@");
	if (atIdx <= 1) {
		return trimmed;
	}
	return `${"*".repeat(Math.min(12, atIdx))}${trimmed.slice(atIdx)}`;
}

export function UserSettingsAccountSection({
	user,
	onNavigateSection,
	onAccountClosed,
}: UserSettingsAccountSectionProps) {
	const { t } = useTranslation();
	const [revealEmail, setRevealEmail] = useState(false);
	const [changeUsernameOpen, setChangeUsernameOpen] = useState(false);
	const [changeEmailOpen, setChangeEmailOpen] = useState(false);
	const [changePasswordOpen, setChangePasswordOpen] = useState(false);
	const [mfaOpen, setMfaOpen] = useState(false);
	const [accountStatusMode, setAccountStatusMode] = useState<AccountStatusDialogMode | null>(null);

	const { data: policy } = useQuery(instancePolicyQueryOpts());
	const mfaFeatureEnabled = isMfaFeatureEnabled(policy);
	const { data: mfaStatus } = useQuery(myMfaStatusQueryOptions(mfaFeatureEnabled));
	const userMfaActive = Boolean(
		mfaStatus?.totp_enabled || (mfaStatus?.security_keys?.length ?? 0) > 0,
	);

	const username = user?.username?.trim() || "";
	const email = user?.email?.trim() || "";
	const hasPassword = user?.has_password ?? true;

	return (
		<div className="flex flex-col">
			{/* Account Info */}
			<section id="user-settings-account-info" className="scroll-mt-4 space-y-2">
				<h3 className="text-base font-semibold tracking-tight text-foreground">
					{t("user_settings.account.info_heading")}
				</h3>
				<div className="space-y-1">
					<div className="flex items-center justify-between gap-4 py-2.5">
						<span className="text-sm font-medium text-muted-foreground">
							{t("user_settings.account.username")}
						</span>
						<div className="flex min-w-0 items-center gap-3">
							<span className="truncate text-sm text-foreground">
								{username || t("user_settings.account.not_set")}
							</span>
							<Button
								type="button"
								variant="secondary"
								className={editButtonClass}
								onClick={() => setChangeUsernameOpen(true)}
							>
								{t("user_settings.account.edit")}
							</Button>
						</div>
					</div>

					<div className="flex items-center justify-between gap-4 py-2.5">
						<span className="text-sm font-medium text-muted-foreground">
							{t("user_settings.account.email")}
						</span>
						<div className="flex min-w-0 items-center gap-3">
							<div className="flex min-w-0 items-center gap-2 text-sm text-foreground">
								<span className="truncate">
									{email
										? revealEmail
											? email
											: maskEmail(email)
										: t("user_settings.account.not_set")}
								</span>
								{email ? (
									<button
										type="button"
										onClick={() => setRevealEmail((prev) => !prev)}
										className="shrink-0 text-xs font-medium text-primary hover:underline"
									>
										{revealEmail
											? t("user_settings.account.hide")
											: t("user_settings.account.reveal")}
									</button>
								) : null}
							</div>
							<Button
								type="button"
								variant="secondary"
								className={editButtonClass}
								onClick={() => setChangeEmailOpen(true)}
							>
								{t("user_settings.account.edit")}
							</Button>
						</div>
					</div>
				</div>
			</section>

			<div className={settingsSectionDividerWrapClass} role="presentation">
				<div className={settingsSectionDividerLineClass} />
			</div>

			{/* Password & Security */}
			<section id="user-settings-password-security" className="scroll-mt-4 space-y-2">
				<h3 className="text-base font-semibold tracking-tight text-foreground">
					{t("user_settings.account.security_heading")}
				</h3>
				<div className="space-y-1">
					<div className="flex items-center justify-between gap-4 py-2.5">
						<span className="text-sm font-medium text-muted-foreground">
							{t("user_settings.account.password")}
						</span>
						<div className="flex items-center gap-3">
							<span
								className={
									hasPassword
										? "text-sm tracking-widest text-muted-foreground"
										: "text-sm text-muted-foreground"
								}
							>
								{hasPassword
									? t("user_settings.account.password_masked")
									: t("user_settings.account.not_set")}
							</span>
							<Button
								type="button"
								variant="secondary"
								className={editButtonClass}
								onClick={() => setChangePasswordOpen(true)}
							>
								{hasPassword
									? t("user_settings.account.edit")
									: t("user_settings.account.password_setup")}
							</Button>
						</div>
					</div>

					{mfaFeatureEnabled ? (
						<button
							type="button"
							onClick={() => setMfaOpen(true)}
							className="flex w-full items-center justify-between gap-4 py-2.5 text-left transition-colors hover:text-foreground"
						>
							<span className="inline-flex items-center gap-2 text-sm font-medium text-muted-foreground">
								<span>{t("user_settings.account.mfa")}</span>
								{userMfaActive ? (
									<span className="rounded bg-emerald-500/15 px-1.5 py-0.5 text-[11px] font-medium text-emerald-500">
										{t("user_settings.account.mfa_enabled")}
									</span>
								) : null}
							</span>
							<span className="inline-flex items-center gap-1.5 text-sm font-medium text-foreground">
								<span>
									{userMfaActive
										? t("user_settings.account.mfa_manage")
										: t("user_settings.account.mfa_setup")}
								</span>
								<ChevronRight className="size-4 text-muted-foreground" aria-hidden />
							</span>
						</button>
					) : null}

					<button
						type="button"
						onClick={() => onNavigateSection("devices")}
						className="flex w-full items-center justify-between gap-4 py-2.5 text-left transition-colors hover:text-foreground"
					>
						<span className="text-sm font-medium text-muted-foreground">
							{t("user_settings.account.devices")}
						</span>
						<span className="inline-flex items-center gap-1.5 text-sm font-medium text-foreground">
							<span>{t("user_settings.account.devices_count")}</span>
							<ChevronRight className="size-4 text-muted-foreground" aria-hidden />
						</span>
					</button>
				</div>
			</section>

			<div className={settingsSectionDividerWrapClass} role="presentation">
				<div className={settingsSectionDividerLineClass} />
			</div>

			{/* Account Status */}
			<section id="user-settings-account-status" className="scroll-mt-4 space-y-2">
				<h3 className="text-base font-semibold tracking-tight text-foreground">
					{t("user_settings.account.status_heading")}
				</h3>
				<div className="space-y-1">
					<div className="flex items-center justify-between gap-4 py-2.5">
						<div className="min-w-0">
							<p className="text-sm font-medium text-foreground">
								{t("user_settings.account.disable_title")}
							</p>
							<p className="mt-0.5 text-xs text-muted-foreground">
								{t("user_settings.account.disable_subtitle")}
							</p>
						</div>
						<Button
							type="button"
							variant="secondary"
							size="lg"
							onClick={() => setAccountStatusMode("disable")}
							className={`shrink-0 text-destructive hover:text-destructive ${statusActionButtonClass}`}
						>
							{t("user_settings.account.disable_action")}
						</Button>
					</div>

					<div className="flex items-center justify-between gap-4 py-2.5">
						<div className="min-w-0">
							<p className="text-sm font-medium text-foreground">
								{t("user_settings.account.delete_title")}
							</p>
							<p className="mt-0.5 text-xs text-muted-foreground">
								{t("user_settings.account.delete_subtitle")}
							</p>
						</div>
						<Button
							type="button"
							variant="destructive"
							size="lg"
							onClick={() => setAccountStatusMode("delete")}
							className={`shrink-0 ${statusActionButtonClass}`}
						>
							{t("user_settings.account.delete_action")}
						</Button>
					</div>
				</div>
			</section>

			<ChangeUsernameDialog
				open={changeUsernameOpen}
				onOpenChange={setChangeUsernameOpen}
				currentUsername={username}
			/>
			<ChangeEmailDialog
				open={changeEmailOpen}
				onOpenChange={setChangeEmailOpen}
				currentEmail={email}
			/>
			<ChangePasswordDialog
				open={changePasswordOpen}
				onOpenChange={setChangePasswordOpen}
				hasPassword={hasPassword}
			/>
			{mfaFeatureEnabled ? <MfaSettingsDialog open={mfaOpen} onOpenChange={setMfaOpen} /> : null}
			<AccountStatusDialog
				open={accountStatusMode !== null}
				onOpenChange={(nextOpen) => {
					if (!nextOpen) {
						setAccountStatusMode(null);
					}
				}}
				mode={accountStatusMode ?? "disable"}
				hasPassword={hasPassword}
				onAccountClosed={onAccountClosed}
			/>
		</div>
	);
}
