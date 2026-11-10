"use client";

import { Camera, Plus, Trash2 } from "lucide-react";
import { useEffect, useId, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { groupInitials } from "@/features/shell/lib/group-initials";
import { cn } from "@/lib/utils";

// Keep in sync with ms-user/pkg/constants.MaxGroupIconBytes (5 MiB).
const MAX_BYTES = 5 * 1024 * 1024;
const ACCEPT = "image/png,image/jpeg,image/webp";

const OUTER_CLASS = "size-[100px] sm:size-[108px]";
const INNER_CLASS = "size-[84px] sm:size-[92px]";
const CORNER_BADGE_CLASS =
	"absolute top-0.5 right-0 z-[2] flex size-7 translate-x-1 items-center justify-center rounded-full shadow-md ring-2 ring-card sm:top-1 sm:translate-x-1.5";
const CORNER_BADGE_BUTTON_CLASS =
	"transition-[filter,transform] duration-200 ease-out hover:brightness-110 active:scale-95 motion-reduce:transition-none focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-card";
const CAMERA_ICON_CLASS = "relative z-[1] size-8 sm:size-9";

type GroupIconUploadProps = {
	value: File | null;
	onChange: (file: File | null) => void;
	/** Group display name — drives default initials avatar (e.g. Wirang Corp → WC). */
	displayName?: string;
	className?: string;
};

export function GroupIconUpload({
	value,
	onChange,
	displayName = "",
	className,
}: GroupIconUploadProps) {
	const { t } = useTranslation();
	const inputId = useId();
	const inputRef = useRef<HTMLInputElement>(null);
	const [error, setError] = useState<string | null>(null);
	const previewUrl = useMemo(() => (value ? URL.createObjectURL(value) : null), [value]);

	useEffect(() => {
		if (!previewUrl) {
			return;
		}
		return () => URL.revokeObjectURL(previewUrl);
	}, [previewUrl]);

	const clearPhoto = () => {
		setError(null);
		onChange(null);
	};

	const pickFile = (file: File | undefined) => {
		if (!file) {
			return;
		}
		if (!file.type.startsWith("image/")) {
			setError(t("group.home.icon_invalid_type"));
			return;
		}
		if (file.size > MAX_BYTES) {
			setError(t("group.home.icon_too_large"));
			return;
		}
		setError(null);
		onChange(file);
	};

	const hasPreview = Boolean(previewUrl);
	const trimmedName = displayName.trim();
	const initials = trimmedName ? groupInitials(trimmedName) : null;
	const showInitials = !hasPreview && initials !== null;
	const showCameraOnHover = hasPreview || showInitials;

	return (
		<div className={cn("flex flex-col items-center gap-2.5", className)}>
			<div className={cn("relative flex items-center justify-center", OUTER_CLASS)}>
				{!hasPreview && !showInitials ? (
					<svg
						className="pointer-events-none absolute inset-0 text-foreground/15 dark:text-foreground/25"
						viewBox="0 0 108 108"
						aria-hidden
					>
						<circle
							cx="54"
							cy="54"
							r="50"
							fill="none"
							stroke="currentColor"
							strokeWidth="1.25"
							strokeDasharray="4.5 6.5"
							strokeLinecap="round"
						/>
					</svg>
				) : (
					<span
						className="pointer-events-none absolute inset-[3px] rounded-full ring-1 ring-border/60 shadow-[0_2px_12px_rgb(0_0_0/0.08)] dark:shadow-[0_2px_16px_rgb(0_0_0/0.35)]"
						aria-hidden
					/>
				)}

				<button
					type="button"
					className={cn(
						INNER_CLASS,
						"group/avatar relative z-[1] flex items-center justify-center overflow-hidden rounded-full",
						"bg-[radial-gradient(125%_125%_at_50%_0%,color-mix(in_oklch,var(--card-foreground)_6%,var(--field))_0%,var(--field)_50%,color-mix(in_oklch,var(--foreground)_5%,var(--field))_100%)]",
						"shadow-[inset_0_1px_0_0_color-mix(in_oklch,var(--card-foreground)_10%,transparent),0_1px_2px_rgb(0_0_0/0.04)]",
						"transition-[transform,box-shadow,filter] duration-200 ease-[cubic-bezier(0.22,1,0.36,1)] motion-reduce:transition-none",
						"focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-card",
						"active:scale-[0.98]",
						!hasPreview && "hover:shadow-elevated hover:brightness-[1.03]",
						hasPreview && "shadow-elevated cursor-pointer",
					)}
					aria-label={t("group.home.icon_upload")}
					onClick={() => inputRef.current?.click()}
				>
					{hasPreview ? (
						// eslint-disable-next-line @next/next/no-img-element -- local blob preview
						<img src={previewUrl!} alt="" className="h-full w-full object-cover" />
					) : null}
					{showInitials ? (
						<span
							className="absolute inset-0 flex items-center justify-center bg-[color-mix(in_oklch,var(--primary)_12%,var(--field))] text-lg font-semibold tracking-tight text-foreground/80 sm:text-xl"
							aria-hidden
						>
							{initials}
						</span>
					) : null}
					<span
						className="pointer-events-none absolute inset-0 flex items-center justify-center"
						aria-hidden
					>
						{showCameraOnHover ? (
							<span
								className={cn(
									"absolute inset-0 bg-black/0 transition-[background-color] duration-200 ease-[cubic-bezier(0.22,1,0.36,1)]",
									"group-hover/avatar:bg-black/[0.32] motion-reduce:transition-none",
								)}
							/>
						) : null}
						<Camera
							className={cn(
								CAMERA_ICON_CLASS,
								"transition-[opacity,transform,color] duration-200 ease-[cubic-bezier(0.22,1,0.36,1)] motion-reduce:transition-none",
								showCameraOnHover
									? "text-white opacity-0 scale-90 group-hover/avatar:opacity-100 group-hover/avatar:scale-100"
									: "text-muted-foreground/70 group-hover/avatar:text-foreground/60",
							)}
							strokeWidth={1.5}
							aria-hidden
						/>
					</span>
				</button>

				{hasPreview ? (
					<button
						type="button"
						className={cn(
							CORNER_BADGE_CLASS,
							CORNER_BADGE_BUTTON_CLASS,
							"bg-destructive text-destructive-foreground",
						)}
						aria-label={t("group.home.icon_remove")}
						onClick={clearPhoto}
					>
						<Trash2 className="size-3.5" strokeWidth={2.5} aria-hidden />
					</button>
				) : (
					<button
						type="button"
						className={cn(
							CORNER_BADGE_CLASS,
							CORNER_BADGE_BUTTON_CLASS,
							"bg-primary text-primary-foreground",
						)}
						aria-label={t("group.home.icon_upload")}
						onClick={() => inputRef.current?.click()}
					>
						<Plus className="size-3.5" strokeWidth={2.5} aria-hidden />
					</button>
				)}
			</div>

			<input
				ref={inputRef}
				id={inputId}
				tabIndex={-1}
				type="file"
				accept={ACCEPT}
				className="sr-only"
				onChange={(e) => {
					pickFile(e.target.files?.[0]);
					e.target.value = "";
				}}
			/>
			{error ? <p className="text-center text-destructive text-xs">{error}</p> : null}
		</div>
	);
}
