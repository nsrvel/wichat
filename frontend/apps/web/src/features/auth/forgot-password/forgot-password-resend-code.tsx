"use client";

import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { authLinkClass } from "@/features/auth/components/auth-shell";

const DEFAULT_RESEND_COOLDOWN_SEC = 60;

type ForgotPasswordResendCodeProps = {
	onResend?: () => void;
	/** Seconds before resend is allowed (from API after each send). */
	cooldownSec?: number;
};

export function ForgotPasswordResendCode({
	onResend,
	cooldownSec = DEFAULT_RESEND_COOLDOWN_SEC,
}: ForgotPasswordResendCodeProps) {
	return (
		<ResendCooldownTimer key={cooldownSec} cooldownSec={cooldownSec} onResend={onResend} />
	);
}

function ResendCooldownTimer({
	onResend,
	cooldownSec,
}: {
	onResend?: () => void;
	cooldownSec: number;
}) {
	const { t } = useTranslation();
	const [secondsLeft, setSecondsLeft] = useState(cooldownSec);

	useEffect(() => {
		if (secondsLeft <= 0) {
			return;
		}
		const id = window.setInterval(() => {
			setSecondsLeft((current) => (current <= 1 ? 0 : current - 1));
		}, 1000);
		return () => window.clearInterval(id);
	}, [secondsLeft]);

	const handleResend = () => {
		if (secondsLeft > 0) {
			return;
		}
		onResend?.();
	};

	return (
		<div className="flex justify-center pt-1 text-center sm:pt-1.5">
			{secondsLeft > 0 ? (
				<p className="text-xs font-light text-[#6d6d6d] sm:text-sm dark:text-[#a3a3a3]">
					{t("forgot_password.resend_wait", { seconds: secondsLeft })}
				</p>
			) : (
				<p className="text-xs font-light text-[#6d6d6d] sm:text-sm dark:text-[#a3a3a3]">
					{t("forgot_password.resend_prompt")}{" "}
					<button
						type="button"
						className={`${authLinkClass} cursor-pointer border-0 bg-transparent p-0 font-inherit text-xs sm:text-sm`}
						onClick={handleResend}
					>
						{t("forgot_password.resend_action")}
					</button>
				</p>
			)}
		</div>
	);
}
