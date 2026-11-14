"use client";

import type { InvitePreview } from "@wichat/contracts";
import Image from "next/image";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { authProductHeadlineClass, authSubmitClass } from "@/lib/form-field-styles";
import { groupInitials } from "@/lib/group-initials";
import { resolveGroupLogoUrl } from "@/lib/group-logo-url";
import { cn } from "@/lib/utils";

type InviteAcceptCardProps = {
	preview: InvitePreview;
	acceptPending?: boolean;
	onAccept: () => void;
	className?: string;
};

export function InviteAcceptCard({
	preview,
	acceptPending = false,
	onAccept,
	className,
}: InviteAcceptCardProps) {
	const { t } = useTranslation();
	const [logoBroken, setLogoBroken] = useState(false);
	const logoSrc = resolveGroupLogoUrl(preview.group_logo_url ?? "");
	const showLogo = Boolean(logoSrc) && !logoBroken;
	const memberCount = preview.member_count ?? 0;
	const onlineCount = preview.online_count ?? 0;

	return (
		<div className={cn("flex w-full flex-col items-center text-center", className)}>
			<div className="mb-5 flex size-[5.5rem] shrink-0 items-center justify-center overflow-hidden rounded-2xl border border-border bg-card shadow-sm sm:mb-6 sm:size-24">
				{showLogo ? (
					<Image
						src={logoSrc}
						alt=""
						width={96}
						height={96}
						unoptimized
						decoding="async"
						className="size-full object-cover"
						onError={() => setLogoBroken(true)}
					/>
				) : (
					<span className="text-xl font-semibold text-foreground sm:text-2xl">
						{groupInitials(preview.group_name ?? "")}
					</span>
				)}
			</div>

			<p className="mb-1.5 text-sm font-medium text-muted-foreground sm:text-base">
				{t("invite.landing_eyebrow")}
			</p>
			<h2 className={cn(authProductHeadlineClass, "mb-5 sm:mb-6")}>{preview.group_name}</h2>

			<div className="mb-1 flex flex-wrap items-center justify-center gap-x-3 gap-y-1 text-[0.9375rem] text-muted-foreground sm:mb-2 sm:text-base">
				<span className="inline-flex items-center gap-1.5">
					<span className="size-2 shrink-0 rounded-full bg-presence-online" aria-hidden />
					{t("invite.landing_stats_online", { count: onlineCount })}
				</span>
				<span className="text-border" aria-hidden>
					·
				</span>
				<span>{t("invite.landing_stats_members", { count: memberCount })}</span>
			</div>

			{preview.channel_name ? (
				<p className="mt-2 mb-5 text-xs text-muted-foreground sm:mb-6 sm:text-sm">
					{t("invite.landing_lead_channel", { channel: preview.channel_name })}
				</p>
			) : (
				<div className="mb-5 sm:mb-6" />
			)}

			<Button
				type="button"
				disabled={acceptPending}
				className={cn("w-full", authSubmitClass)}
				onClick={onAccept}
			>
				{acceptPending ? t("invite.landing_joining") : t("invite.landing_accept_invite")}
			</Button>
		</div>
	);
}
