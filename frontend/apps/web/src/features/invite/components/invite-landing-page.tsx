"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useRef } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import {
	AuthShell,
	authLeadSwitchClass,
	authLinkClass,
	authPageBackdropClass,
	authSubmitClass,
	groupOnboardingCardShellClass,
} from "@/features/auth/components/auth-shell";
import { translateError } from "@/integrations/i18n/translate-error";
import { authPath } from "@/lib/auth-invite";
import { channelPath } from "@/lib/app-routes";
import { useMounted } from "@/hooks/use-mounted";
import * as authQuery from "@/services/auth/auth.query";
import { Q_SESSION } from "@/services/auth/auth.key";
import { Q_MY_GROUPS } from "@/services/group/group.key";
import { previewInvite, redeemInvite } from "@/services/invite/invite.service";

type InviteLandingPageProps = {
	inviteToken: string;
};

export function InviteLandingPage({ inviteToken }: InviteLandingPageProps) {
	const { t } = useTranslation();
	const router = useRouter();
	const queryClient = useQueryClient();
	const mounted = useMounted();
	const redeemStarted = useRef(false);

	const session = useQuery({ ...authQuery.sessionQueryOpts(), enabled: mounted });
	const preview = useQuery({
		queryKey: ["invite", "preview", inviteToken],
		queryFn: () => previewInvite(inviteToken),
		enabled: Boolean(inviteToken),
		retry: 1,
	});

	const redeem = useMutation({
		mutationFn: () => redeemInvite(inviteToken),
		onSuccess: (result) => {
			void queryClient.invalidateQueries({ queryKey: [Q_SESSION] });
			void queryClient.invalidateQueries({ queryKey: [Q_MY_GROUPS] });
			toast.success(t("invite.landing_joined", { name: preview.data?.group_name ?? "" }));
			if (result.channel_id) {
				router.replace(channelPath(result.group_id, result.channel_id));
				return;
			}
			router.replace("/");
		},
		onError: (err) => {
			toast.error(translateError(t, err));
		},
	});

	useEffect(() => {
		if (
			!session.data ||
			!preview.data?.valid ||
			redeemStarted.current ||
			redeem.isPending ||
			redeem.isSuccess
		) {
			return;
		}
		redeemStarted.current = true;
		redeem.mutate();
	}, [session.data, preview.data?.valid, redeem]);

	if (!inviteToken) {
		return (
			<div className={authPageBackdropClass}>
				<div className={groupOnboardingCardShellClass}>
					<p className="px-6 py-10 text-center text-sm text-muted-foreground">
						{t("invite.landing_missing")}
					</p>
				</div>
			</div>
		);
	}

	const headline = preview.data?.valid
		? t("invite.landing_title", { name: preview.data.group_name })
		: t("invite.landing_title_invalid");

	const subtitle = preview.data?.valid
		? preview.data.channel_name
			? t("invite.landing_lead_channel", { channel: preview.data.channel_name })
			: t("invite.landing_lead")
		: preview.data?.reason === "expired"
			? t("invite.landing_reason_expired")
			: preview.data?.reason === "redeemed"
				? t("invite.landing_reason_redeemed")
				: t("invite.landing_reason_not_found");

	const showAuthCta = !session.data && preview.data?.valid;

	return (
		<AuthShell headline={headline} lead={<p className="text-sm text-muted-foreground">{subtitle}</p>}>
			{preview.isPending ? (
				<p className="text-center text-sm text-muted-foreground">{t("invite.landing_loading")}</p>
			) : null}
			{preview.isError ? (
				<p className="text-center text-sm text-destructive">{translateError(t, preview.error)}</p>
			) : null}
			{session.data && preview.data?.valid ? (
				<p className="text-center text-sm text-muted-foreground">{t("invite.landing_joining")}</p>
			) : null}
			{showAuthCta ? (
				<div className="flex flex-col gap-3">
					<Link
						href={authPath("/sign-up", { invite: inviteToken, flow: "register" })}
						className={authSubmitClass}
					>
						{t("invite.landing_sign_up")}
					</Link>
					<p className={authLeadSwitchClass}>
						<span>{t("auth.have_account")}</span>
						<Link
							href={authPath("/sign-in", { invite: inviteToken, as: "account" })}
							className={authLinkClass}
						>
							{t("auth.to_login")}
						</Link>
					</p>
				</div>
			) : null}
		</AuthShell>
	);
}
