"use client";

import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { authLinkClass } from "@/features/auth/components/auth-shell";

const DEFAULT_RESEND_COOLDOWN_SEC = 60;

type ForgotPasswordResendCodeProps = {
	onResend?: () => void;
	/** Seconds before resend is allowed (from API after each send). */
	cooldownSec?: number;
	isResending?: boolean;
};

export function ForgotPasswordResendCode({
	onResend,
	cooldownSec = DEFAULT_RESEND_COOLDOWN_SEC,
	isResending = false,
}: ForgotPasswordResendCodeProps) {
	const effectiveCooldown = cooldownSec > 0 ? cooldownSec : DEFAULT_RESEND_COOLDOWN_SEC;

	return (
		<ResendCooldownTimer
			key={effectiveCooldown}
			cooldownSec={effectiveCooldown}
			onResend={onResend}
			isResending={isResending}
		/>
	);
}

function ResendCooldownTimer({
	onResend,
	cooldownSec,
	isResending,
}: {
	onResend?: () => void;
	cooldownSec: number;
	isResending: boolean;
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
		if (secondsLeft > 0 || isResending) {
			return;
		}
		setSecondsLeft(cooldownSec);
		onResend?.();
	};

	const canResend = secondsLeft <= 0 && !isResending;

	return (
		<div className="relative z-10 flex justify-center pt-1 text-center sm:pt-1.5">
			{secondsLeft > 0 ? (
				<p className="text-xs font-light text-muted-foreground sm:text-sm">
					{t("forgot_password.resend_wait", { seconds: secondsLeft })}
				</p>
			) : (
				<p className="text-xs font-light text-muted-foreground sm:text-sm">
					{t("forgot_password.resend_prompt")}{" "}
					<button
						type="button"
						disabled={!canResend}
						className={`${authLinkClass} cursor-pointer border-0 bg-transparent p-0 font-inherit text-xs sm:text-sm disabled:cursor-not-allowed disabled:opacity-50`}
						onClick={handleResend}
					>
						{isResending ? t("forgot_password.sending_code") : t("forgot_password.resend_action")}
					</button>
				</p>
			)}
		</div>
	);
}
