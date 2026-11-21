"use client";

import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { authSubmitBlocked } from "@/features/auth/lib/captcha-policy";
import { instancePolicyQueryOpts } from "@/services/auth/auth.query";

/** Turnstile state for invite create/preview/redeem calls when instance policy requires it. */
export function useInviteTurnstile(active: boolean) {
	const { data: policy, isSuccess: policyReady } = useQuery({
		...instancePolicyQueryOpts(),
		enabled: active,
	});
	const [token, setToken] = useState("");
	const [resetKey, setResetKey] = useState(0);

	const siteKey = policy?.turnstile_site_key ?? "";
	const blocked = active && authSubmitBlocked(policy, policyReady, token);
	const turnstileField = token.trim() !== "" ? token.trim() : undefined;

	const bumpReset = () => setResetKey((key) => key + 1);

	return {
		siteKey,
		token,
		setToken,
		resetKey,
		bumpReset,
		captchaReady: active && policyReady && !blocked,
		turnstileField,
	};
}