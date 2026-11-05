"use client";

import { ChevronLeft } from "lucide-react";
import Link from "next/link";
import { useTranslation } from "react-i18next";
import { authFieldLinkClass } from "@/features/auth/components/auth-shell";

export function ForgotPasswordBackToLogin() {
	const { t } = useTranslation();

	return (
		<div className="flex justify-center pt-1 sm:pt-1.5">
			<Link
				href="/sign-in"
				className={`${authFieldLinkClass} inline-flex items-center gap-0.5 text-sm`}
			>
				<ChevronLeft className="size-4 shrink-0" strokeWidth={1.8} aria-hidden />
				{t("forgot_password.back_to_login")}
			</Link>
		</div>
	);
}
