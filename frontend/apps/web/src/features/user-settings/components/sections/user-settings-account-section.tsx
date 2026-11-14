"use client";

import type { User } from "@wichat/contracts";
import { CheckCircle2, ChevronRight, Info } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import type { UserSettingsSectionId } from "@/features/user-settings/lib/user-settings-nav";

type UserSettingsAccountSectionProps = {
	user: User | null | undefined;
	onNavigateSection: (section: UserSettingsSectionId) => void;
};

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
}: UserSettingsAccountSectionProps) {
	const { t } = useTranslation();
	const [revealEmail, setRevealEmail] = useState(false);

	const username = user?.username?.trim() || "";
	const email = user?.email?.trim() || "";

	const notifySoon = () => {
		toast.info(t("shell.group_rail_menu.coming_soon"));
	};

	return (
		<div className="space-y-7">
			{/* Account Info */}
			<section id="user-settings-account-info" className="space-y-3">
				<h3 className="text-base font-semibold tracking-tight text-foreground">
					{t("user_settings.account.info_heading")}
				</h3>
				<div className="divide-y divide-border/60 rounded-xl border border-border/70 bg-muted/20 px-4">
					<div className="flex items-center justify-between gap-4 py-3.5">
						<span className="text-sm font-medium text-muted-foreground">
							{t("user_settings.account.username")}
						</span>
						<div className="flex min-w-0 items-center gap-3">
							<span className="truncate text-sm text-foreground">
								{username || t("user_settings.account.not_set")}
							</span>
							<Button type="button" variant="secondary" size="sm" onClick={notifySoon}>
								{t("user_settings.account.edit")}
							</Button>
						</div>
					</div>

					<div className="flex items-center justify-between gap-4 py-3.5">
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
							<Button type="button" variant="secondary" size="sm" onClick={notifySoon}>
								{t("user_settings.account.edit")}
							</Button>
						</div>
					</div>
				</div>
			</section>

			{/* Password & Security */}
			<section id="user-settings-password-security" className="space-y-3">
				<h3 className="text-base font-semibold tracking-tight text-foreground">
					{t("user_settings.account.security_heading")}
				</h3>
				<div className="divide-y divide-border/60 rounded-xl border border-border/70 bg-muted/20 px-4">
					<div className="flex items-center justify-between gap-4 py-3.5">
						<span className="text-sm font-medium text-muted-foreground">
							{t("user_settings.account.password")}
						</span>
						<div className="flex items-center gap-3">
							<span className="text-sm tracking-widest text-muted-foreground">
								{t("user_settings.account.password_masked")}
							</span>
							<Button type="button" variant="secondary" size="sm" onClick={notifySoon}>
								{t("user_settings.account.edit")}
							</Button>
						</div>
					</div>

					<button
						type="button"
						onClick={notifySoon}
						className="flex w-full items-center justify-between gap-4 py-3.5 text-left transition-colors hover:text-foreground"
					>
						<span className="text-sm font-medium text-muted-foreground">
							{t("user_settings.account.mfa")}
						</span>
						<span className="inline-flex items-center gap-1.5 text-sm font-medium text-foreground">
							<span>{t("user_settings.account.mfa_setup")}</span>
							<ChevronRight className="size-4 text-muted-foreground" aria-hidden />
						</span>
					</button>

					<button
						type="button"
						onClick={() => onNavigateSection("devices")}
						className="flex w-full items-center justify-between gap-4 py-3.5 text-left transition-colors hover:text-foreground"
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

			{/* Account Status */}
			<section id="user-settings-account-status" className="space-y-3">
				<h3 className="text-base font-semibold tracking-tight text-foreground">
					{t("user_settings.account.status_heading")}
				</h3>
				<div className="flex items-center gap-2.5 rounded-xl border border-primary/30 bg-primary/10 px-4 py-3 text-xs text-foreground/90">
					<Info className="size-4 shrink-0 text-primary" aria-hidden />
					<span>{t("user_settings.account.status_banner")}</span>
				</div>

				<div className="flex items-center justify-between gap-4 rounded-xl border border-border/70 bg-muted/20 px-4 py-3.5">
					<div className="flex min-w-0 items-center gap-3">
						<span className="flex size-9 shrink-0 items-center justify-center rounded-xl bg-emerald-500/15 text-emerald-500">
							<CheckCircle2 className="size-5" aria-hidden />
						</span>
						<div className="min-w-0">
							<p className="text-sm font-medium text-foreground">
								{t("user_settings.account.standing_title")}
							</p>
							<p className="text-xs text-muted-foreground">
								{t("user_settings.account.standing_subtitle")}
							</p>
						</div>
					</div>
					<span className="shrink-0 text-sm font-medium text-emerald-500">
						{t("user_settings.account.standing_badge")}
					</span>
				</div>
			</section>
		</div>
	);
}
