"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { useEffect, useRef } from "react";
import { inviteLandingPath, pickInviteToken } from "@/lib/auth-invite";
import { instancePolicyQueryOpts } from "@/services/auth/auth.query";
import { channelPath } from "@/lib/app-routes";
import { Q_SESSION } from "@/services/auth/auth.key";
import { Q_MY_GROUPS } from "@/services/group/group.key";
import { redeemInvite } from "@/services/invite/invite.service";

/** After login in app shell, consumes ?invite= when not on /invite (once per mount). */
export function useRedeemInviteFromQuery() {
	const params = useSearchParams();
	const pathname = usePathname();
	const token = pickInviteToken({ invite: params.get("invite") ?? undefined });
	const queryClient = useQueryClient();
	const router = useRouter();
	const done = useRef(false);
	const { data: policy } = useQuery({
		...instancePolicyQueryOpts(),
		enabled: Boolean(token),
	});

	const redeem = useMutation({
		mutationFn: () => redeemInvite(token!),
		onSuccess: (result) => {
			void queryClient.invalidateQueries({ queryKey: [Q_SESSION] });
			void queryClient.invalidateQueries({ queryKey: [Q_MY_GROUPS] });
			if (result.channel_id) {
				router.replace(channelPath(result.group_id, result.channel_id));
				return;
			}
			router.replace("/");
		},
	});

	useEffect(() => {
		if (pathname?.startsWith("/invite")) {
			return;
		}
		if (!token || done.current || redeem.isPending || redeem.isSuccess) {
			return;
		}
		if (policy?.turnstile_required) {
			done.current = true;
			router.replace(inviteLandingPath(token));
			return;
		}
		done.current = true;
		redeem.mutate();
	}, [token, pathname, redeem, policy?.turnstile_required, router]);
}
