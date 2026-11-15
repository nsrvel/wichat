"use client";

import Image from "next/image";
import { useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { groupInitials } from "@/lib/group-initials";
import { resolveGroupLogoUrl } from "@/lib/group-logo-url";
import { cn } from "@/lib/utils";

type GroupProfileCardPreviewProps = {
	groupName: string;
	description?: string;
	bannerColor: string;
	logoUrl?: string;
	iconFile?: File | null;
	className?: string;
};

export function GroupProfileCardPreview({
	groupName,
	description = "",
	bannerColor,
	logoUrl = "",
	iconFile = null,
	className,
}: GroupProfileCardPreviewProps) {
	const { t } = useTranslation();
	const [brokenLogoKey, setBrokenLogoKey] = useState<string | null>(null);

	const iconBlobUrl = useMemo(() => (iconFile ? URL.createObjectURL(iconFile) : null), [iconFile]);

	useEffect(() => {
		if (!iconBlobUrl) {
			return;
		}
		return () => URL.revokeObjectURL(iconBlobUrl);
	}, [iconBlobUrl]);

	const currentLogoKey = `${logoUrl}::${iconBlobUrl ?? ""}`;
	const logoBroken = brokenLogoKey === currentLogoKey;
	const resolvedLogo = resolveGroupLogoUrl(logoUrl);
	const iconSrc = iconBlobUrl ?? (resolvedLogo && !logoBroken ? resolvedLogo : "");
	const showIconImage = Boolean(iconSrc);
	const displayName = groupName.trim() || t("shell.unknown_group");
	const initials = groupInitials(displayName);
	const trimmedDescription = description.trim();

	return (
		<div
			className={cn(
				"overflow-hidden rounded-2xl border border-border bg-card shadow-sm",
				className,
			)}
		>
			<div
				className="h-[76px] transition-colors duration-150"
				style={{ backgroundColor: bannerColor }}
			/>
			<div className="px-5 pb-5">
				<div className="-mt-[40px] flex flex-col items-center text-center">
					<div className="size-[72px] shrink-0 overflow-hidden rounded-full bg-card ring-[5px] ring-card">
						{showIconImage ? (
							<Image
								src={iconSrc}
								alt=""
								width={72}
								height={72}
								unoptimized
								className="size-full object-cover"
								onError={() => setBrokenLogoKey(currentLogoKey)}
							/>
						) : (
							<span className="flex size-full items-center justify-center bg-[color-mix(in_oklch,var(--primary)_12%,var(--field))] text-lg font-semibold text-foreground/80">
								{initials || "?"}
							</span>
						)}
					</div>
					<div className="mt-3 w-full min-w-0">
						<p className="truncate text-base font-bold leading-tight text-foreground">
							{displayName}
						</p>
						{trimmedDescription ? (
							<p className="mt-2 whitespace-pre-wrap break-words text-xs leading-relaxed text-muted-foreground">
								{trimmedDescription}
							</p>
						) : null}
					</div>
				</div>
			</div>
		</div>
	);
}
