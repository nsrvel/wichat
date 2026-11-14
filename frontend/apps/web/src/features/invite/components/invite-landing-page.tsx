"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useEffect } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { AuthShell } from "@/features/auth/components/auth-shell";
import { authPageBackdropClass, groupOnboardingCardShellClass } from "@/lib/form-field-styles";
import { InviteAcceptCard } from "@/features/invite/components/invite-accept-card";
import { InviteAcceptCardSkeleton } from "@/features/invite/components/invite-accept-card-skeleton";
import { translateError } from "@/integrations/i18n/translate-error";
import { authPathWithReturn, inviteLandingPath } from "@/lib/auth-invite";
import { channelPath } from "@/lib/app-routes";
import { useMounted } from "@/hooks/use-mounted";
import * as authQuery from "@/services/auth/auth.query";
import { Q_SESSION } from "@/services/auth/auth.key";
import { Q_MY_GROUPS } from "@/services/group/group.key";
import { Q_INVITE_PREVIEW, invitePreviewQueryOpts } from "@/services/invite/invite.query";
import { readInvitePreviewCache } from "@/services/invite/invite-preview-cache";
import { redeemInvite } from "@/services/invite/invite.service";

type InviteLandingPageProps = {
	inviteToken: string;
};

export function InviteLandingPage({ inviteToken }: InviteLandingPageProps) {
	const { t } = useTranslation();
	const router = useRouter();
	const queryClient = useQueryClient();
	const mounted = useMounted();

	const token = inviteToken.trim();
	const session = useQuery({ ...authQuery.sessionQueryOpts(), enabled: mounted });
	const preview = useQuery({
		...invitePreviewQueryOpts(inviteToken),
		enabled: mounted && Boolean(token),
	});

	useEffect(() => {
		if (!mounted || !token) {
			return;
		}
		const cached = readInvitePreviewCache(token);
		if (cached) {
			queryClient.setQueryData([Q_INVITE_PREVIEW, token], cached);
		}
	}, [mounted, token, queryClient]);

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

	const returnTo = inviteLandingPath(inviteToken);
	const authReady = mounted && !session.isPending;
	const previewData = preview.data;
	const previewValid = previewData?.valid === true;
	const previewInvalid = preview.isSuccess && previewData?.valid === false;
	const needsSignIn = authReady && !session.data && previewValid;

	const sessionKnown = Boolean(session.data) || session.isPending || !authReady;
	const showAcceptCard =
		mounted && previewValid && Boolean(previewData) && sessionKnown && !needsSignIn;
	const showAcceptSkeleton = !previewData && preview.isPending && !preview.isError;

	useEffect(() => {
		if (!needsSignIn) {
			return;
		}
		router.replace(authPathWithReturn("/sign-in", returnTo));
	}, [needsSignIn, returnTo, router]);

	const invalidLead =
		previewData?.reason === "expired"
			? t("invite.landing_reason_expired")
			: previewData?.reason === "redeemed"
				? t("invite.landing_reason_redeemed")
				: previewInvalid
					? t("invite.landing_reason_not_found")
					: null;

	const headline = !mounted ? (
		<span className="sr-only">{t("invite.landing_loading")}</span>
	) : previewValid ? (
		<span className="sr-only">
			{t("invite.landing_title", { name: previewData?.group_name ?? "" })}
		</span>
	) : previewData ? (
		t("invite.landing_title_invalid")
	) : (
		<span className="sr-only">{t("invite.landing_loading")}</span>
	);

	const lead = invalidLead ? <p className="text-sm text-muted-foreground">{invalidLead}</p> : null;

	const handleAccept = () => {
		if (!session.data) {
			router.push(authPathWithReturn("/sign-in", returnTo));
			return;
		}
		redeem.mutate();
	};

	if (needsSignIn) {
		return <div className={authPageBackdropClass} aria-hidden suppressHydrationWarning />;
	}

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

	const showLoggedOutPreviewWait = authReady && !session.data && showAcceptSkeleton;

	return (
		<AuthShell headline={headline} lead={lead}>
			{preview.isError && !previewData ? (
				<p className="text-center text-sm text-destructive">{translateError(t, preview.error)}</p>
			) : null}
			{showAcceptCard ? (
				<InviteAcceptCard
					preview={previewData}
					acceptPending={redeem.isPending}
					onAccept={handleAccept}
				/>
			) : null}
			{!showAcceptCard && (showAcceptSkeleton || showLoggedOutPreviewWait) ? (
				<InviteAcceptCardSkeleton />
			) : null}
		</AuthShell>
	);
}
