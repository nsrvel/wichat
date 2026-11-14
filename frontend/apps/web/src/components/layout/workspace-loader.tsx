"use client";

import { useTranslation } from "react-i18next";
import { cn } from "@/lib/utils";

const LOADER_VARIANTS = [
	{
		src: "/loaders/wichat-loading-black.webp",
		visibilityClass: "block dark:hidden",
	},
	{
		src: "/loaders/wichat-loading-white.webp",
		visibilityClass: "hidden dark:block",
	},
] as const;

type WorkspaceLoaderProps = {
	className?: string;
	size?: number;
	showLabel?: boolean;
	labelKey?: "shell.loading" | "group.home.loading";
};

/** Theme-aware animated WiChat logo loader for workspace gating and transitions. */
export function WorkspaceLoader({
	className,
	size = 96,
	showLabel = false,
	labelKey = "shell.loading",
}: WorkspaceLoaderProps) {
	const { t } = useTranslation();

	return (
		<div
			role="status"
			aria-live="polite"
			className={cn(
				"flex flex-1 flex-col items-center justify-center gap-3 bg-background",
				className,
			)}
		>
			{LOADER_VARIANTS.map(({ src, visibilityClass }) => (
				// eslint-disable-next-line @next/next/no-img-element -- Animated local WebPs bypass next/image optimization to preserve frames.
				<img
					key={src}
					src={src}
					alt=""
					aria-hidden="true"
					width={size}
					height={size}
					decoding="async"
					fetchPriority="high"
					className={cn("select-none object-contain", visibilityClass)}
				/>
			))}
			<span
				className={
					showLabel ? "text-xs font-medium text-muted-foreground" : "sr-only"
				}
				suppressHydrationWarning
			>
				{t(labelKey)}
			</span>
		</div>
	);
}

