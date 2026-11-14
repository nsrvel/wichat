"use client";

import { useEffect, useId, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { FieldLabelContent } from "@/components/form/field-label";
import { Label } from "@/components/ui/label";
import { authFieldLabelClass } from "@/lib/form-field-styles";
import { GroupBannerImageSwatch } from "@/components/group/group-banner-image-swatch";
import { GroupBannerPaletteSwatch } from "@/components/group/group-banner-palette-swatch";
import {
	GROUP_BANNER_ACCEPT,
	GROUP_BANNER_COLOR_PRESETS,
	GROUP_BANNER_IMAGE_ID,
	MAX_GROUP_BANNER_BYTES,
	isGroupBannerImageId,
	type GroupBannerColorId,
} from "@/lib/group-banner-presets";
import { cn } from "@/lib/utils";

type GroupBannerFieldProps = {
	selection: GroupBannerColorId;
	onSelectionChange: (id: GroupBannerColorId) => void;
	imageFile: File | null;
	onImageFileChange: (file: File | null) => void;
	existingImageUrl?: string;
	onClearExistingImage?: () => void;
	className?: string;
};

export function GroupBannerField({
	selection,
	onSelectionChange,
	imageFile,
	onImageFileChange,
	existingImageUrl = "",
	onClearExistingImage,
	className,
}: GroupBannerFieldProps) {
	const { t } = useTranslation();
	const inputId = useId();
	const fileRef = useRef<HTMLInputElement>(null);
	const [error, setError] = useState<string | null>(null);

	const previewUrl = useMemo(
		() => (imageFile ? URL.createObjectURL(imageFile) : null),
		[imageFile],
	);

	useEffect(() => {
		if (!previewUrl) {
			return;
		}
		return () => URL.revokeObjectURL(previewUrl);
	}, [previewUrl]);

	const remoteUrl = existingImageUrl.trim();
	const swatchPreviewSrc = previewUrl ?? (remoteUrl !== "" ? remoteUrl : "");

	const pickFile = (file: File | undefined) => {
		if (!file) {
			return;
		}
		if (!file.type.startsWith("image/")) {
			setError(t("group_settings.profile.banner_invalid_type"));
			return;
		}
		if (file.size > MAX_GROUP_BANNER_BYTES) {
			setError(t("group_settings.profile.banner_too_large"));
			return;
		}
		setError(null);
		onSelectionChange(GROUP_BANNER_IMAGE_ID);
		onImageFileChange(file);
	};

	const activateImageSwatch = () => {
		onSelectionChange(GROUP_BANNER_IMAGE_ID);
		fileRef.current?.click();
	};

	const selectColor = (id: GroupBannerColorId) => {
		if (isGroupBannerImageId(id)) {
			return;
		}
		onSelectionChange(id);
		onImageFileChange(null);
		if (remoteUrl !== "") {
			onClearExistingImage?.();
		}
	};

	return (
		<div className={cn("flex flex-col gap-3", className)}>
			<Label className={authFieldLabelClass}>
				<FieldLabelContent>{t("group_settings.profile.banner_label")}</FieldLabelContent>
			</Label>

			<div className="grid grid-cols-5 gap-2.5">
				{GROUP_BANNER_COLOR_PRESETS.map((preset) =>
					isGroupBannerImageId(preset) ? (
						<GroupBannerImageSwatch
							key={preset}
							selected={selection === GROUP_BANNER_IMAGE_ID}
							previewSrc={swatchPreviewSrc}
							onActivate={activateImageSwatch}
						/>
					) : (
						<GroupBannerPaletteSwatch
							key={preset}
							id={preset}
							selected={selection === preset}
							onSelect={selectColor}
						/>
					),
				)}
			</div>

			<input
				ref={fileRef}
				id={inputId}
				tabIndex={-1}
				type="file"
				accept={GROUP_BANNER_ACCEPT}
				className="sr-only"
				onChange={(e) => {
					pickFile(e.target.files?.[0]);
					e.target.value = "";
				}}
			/>
			{error ? <p className="text-xs text-destructive">{error}</p> : null}
		</div>
	);
}
