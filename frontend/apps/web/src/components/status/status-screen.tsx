"use client";

import Link from "next/link";
import { useTranslation } from "react-i18next";
import { WiStudioLogoMark } from "@/components/brand/wistudio-logo-mark";
import { Button, buttonVariants } from "@/components/ui/button";
import {
	StatusIllustration,
	type StatusScreenVariant,
} from "@/components/status/status-illustrations";
import { cn } from "@/lib/utils";

type StatusScreenProps = {
	variant: StatusScreenVariant;
	onRetry?: () => void;
	homeHref?: string;
	signInHref?: string;
	className?: string;
};

const VARIANT_KEY: Record<
	StatusScreenVariant,
	"not_found" | "forbidden" | "unauthorized" | "server_error" | "maintenance"
> = {
	"not-found": "not_found",
	forbidden: "forbidden",
	unauthorized: "unauthorized",
	error: "server_error",
	maintenance: "maintenance",
};

export function StatusScreen({
	variant,
	onRetry,
	homeHref = "/",
	signInHref = "/sign-in",
	className,
}: StatusScreenProps) {
	const { t } = useTranslation();
	const key = VARIANT_KEY[variant];
	const hasPrimaryOverride = (variant === "error" && Boolean(onRetry)) || variant === "unauthorized";

	return (
		<main
			className={cn(
				"relative flex min-h-dvh w-full flex-col overflow-hidden bg-background text-foreground",
				className,
			)}
		>
			<header className="relative z-10 flex items-center px-6 pt-6 sm:px-10 sm:pt-8">
				<Link
					href={homeHref}
					className="inline-flex items-center gap-2.5 rounded-xl text-sm font-semibold tracking-tight text-foreground/90 transition-opacity hover:opacity-80"
				>
					<WiStudioLogoMark size={28} />
					<span>WiChat</span>
				</Link>
			</header>

			<section className="relative z-10 flex flex-1 items-center justify-center px-6 pb-16 pt-8 sm:px-10">
				<div className="flex w-full max-w-2xl flex-col items-center gap-8 md:flex-row md:items-center md:justify-center md:gap-12">
					<div className="flex shrink-0 items-center justify-center">
						<StatusIllustration variant={variant} />
					</div>

					<div className="flex max-w-sm flex-col items-center text-center md:items-start md:text-left">
						<p className="text-sm font-medium text-muted-foreground sm:text-base">
							{t(`status_pages.${key}.eyebrow`)}
						</p>
						<h1 className="mt-1 font-sans text-6xl font-bold leading-none tracking-tight tabular-nums text-foreground sm:text-7xl">
							{t(`status_pages.${key}.title`)}
						</h1>
						<p className="mt-3.5 text-sm leading-relaxed text-muted-foreground">
							{t(`status_pages.${key}.description`)}
						</p>

						<div key={variant} className="mt-6 flex flex-wrap items-center justify-center gap-2.5 md:justify-start">
							{variant === "error" && onRetry ? (
								<Button
									type="button"
									onClick={onRetry}
									className="h-10 rounded-xl px-5 font-medium transition-colors"
								>
									{t("status_pages.actions.try_again")}
								</Button>
							) : null}

							{variant === "unauthorized" ? (
								<Link
									href={signInHref}
									className={cn(
										buttonVariants({ variant: "default" }),
										"h-10 rounded-xl px-5 font-medium transition-colors",
									)}
								>
									{t("status_pages.actions.sign_in")}
								</Link>
							) : null}

							<Link
								href={homeHref}
								className={cn(
									buttonVariants({ variant: hasPrimaryOverride ? "outline" : "default" }),
									"h-10 rounded-xl px-5 font-medium transition-colors",
								)}
							>
								{t("status_pages.actions.back_home")}
							</Link>
						</div>
					</div>
				</div>
			</section>
		</main>
	);
}
