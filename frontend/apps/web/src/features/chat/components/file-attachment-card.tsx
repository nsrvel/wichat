"use client";

import {
	AlertCircle,
	Download,
	FileArchive,
	FileAudio,
	FileCode2,
	FileSpreadsheet,
	FileText,
	FileVideo,
	Loader2,
	Maximize2,
	RotateCw,
	X,
} from "lucide-react";
import { useEffect, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import {
	columnIndexToLabel,
	getAttachmentVisualMeta,
	getCachedSpreadsheetPreview,
	getCachedTextPreview,
	loadSpreadsheetPreview,
	loadTextPreview,
	MAX_CLIENT_PREVIEW_BYTES,
	renderPdfPageToCanvas,
	type SpreadsheetPreviewData,
	type TextPreviewData,
} from "@/features/chat/lib/attachment-preview";
import { resolvePublicObjectUrl } from "@/lib/public-object-url";
import { cn } from "@/lib/utils";

export function formatAttachmentBytes(bytes: number): string {
	if (!Number.isFinite(bytes) || bytes <= 0) {
		return "0 B";
	}
	const units = ["B", "KB", "MB", "GB"];
	let value = bytes;
	let unitIndex = 0;
	while (value >= 1024 && unitIndex < units.length - 1) {
		value /= 1024;
		unitIndex += 1;
	}
	return `${value >= 10 || unitIndex === 0 ? Math.round(value) : value.toFixed(1)} ${units[unitIndex]}`;
}

export type FileAttachmentCardProps = {
	variant: "composer" | "message";
	id: string;
	filename: string;
	contentType: string;
	sizeBytes: number;
	url?: string;
	localFile?: File;
	status?: "uploading" | "uploaded" | "error";
	errorMessage?: string;
	onRemove?: () => void;
	onRetry?: () => void;
	onOpenPreview?: () => void;
};

function CategoryIcon({ category }: { category: string }) {
	switch (category) {
		case "excel":
		case "csv":
			return <FileSpreadsheet className="size-4" />;
		case "archive":
			return <FileArchive className="size-4" />;
		case "audio":
			return <FileAudio className="size-4" />;
		case "video":
			return <FileVideo className="size-4" />;
		case "text":
			return <FileCode2 className="size-4" />;
		default:
			return <FileText className="size-4" />;
	}
}

export function FileAttachmentCard({
	variant,
	id,
	filename,
	contentType,
	sizeBytes,
	url,
	localFile,
	status = "uploaded",
	errorMessage,
	onRemove,
	onRetry,
	onOpenPreview,
}: FileAttachmentCardProps) {
	const { t } = useTranslation();
	const meta = useMemo(
		() => getAttachmentVisualMeta(filename, contentType),
		[contentType, filename],
	);

	const resolvedUrl = url ? resolvePublicObjectUrl(url) : "";
	const source: File | string | null = localFile ?? (resolvedUrl || null);
	const cacheKey = `${id}:${localFile ? `${localFile.name}:${localFile.size}` : resolvedUrl}`;
	const withinSizeLimit = sizeBytes <= MAX_CLIENT_PREVIEW_BYTES;

	const [sheetData, setSheetData] = useState<SpreadsheetPreviewData | null>(() =>
		getCachedSpreadsheetPreview(cacheKey),
	);
	const [textData, setTextData] = useState<TextPreviewData | null>(() =>
		getCachedTextPreview(cacheKey),
	);
	const [pdfReady, setPdfReady] = useState(false);
	const [previewFailed, setPreviewFailed] = useState(false);
	const [localMediaUrl, setLocalMediaUrl] = useState<string>("");
	const [isNearViewport, setIsNearViewport] = useState(variant === "composer");
	const cardRef = useRef<HTMLDivElement>(null);
	const pdfCanvasRef = useRef<HTMLCanvasElement>(null);

	useEffect(() => {
		if (isNearViewport) {
			return;
		}
		const el = cardRef.current;
		if (!el || typeof IntersectionObserver === "undefined") {
			const timer = window.setTimeout(() => setIsNearViewport(true), 0);
			return () => window.clearTimeout(timer);
		}
		const observer = new IntersectionObserver(
			(entries) => {
				if (entries.some((entry) => entry.isIntersecting)) {
					setIsNearViewport(true);
					observer.disconnect();
				}
			},
			{ rootMargin: "200px" },
		);
		observer.observe(el);
		return () => observer.disconnect();
	}, [isNearViewport]);

	useEffect(() => {
		if (resolvedUrl || !localFile || (meta.category !== "audio" && meta.category !== "video")) {
			return;
		}
		const objUrl = URL.createObjectURL(localFile);
		const timer = window.setTimeout(() => setLocalMediaUrl(objUrl), 0);
		return () => {
			window.clearTimeout(timer);
			URL.revokeObjectURL(objUrl);
		};
	}, [localFile, meta.category, resolvedUrl]);

	useEffect(() => {
		if (!isNearViewport || !source || !meta.supportsMiniPreview || !withinSizeLimit) {
			return;
		}
		if (meta.category === "audio" || meta.category === "video") {
			return;
		}

		let cancelled = false;

		if (meta.category === "excel" || meta.category === "csv") {
			void loadSpreadsheetPreview(source, meta.category, cacheKey).then((res) => {
				if (cancelled) {
					return;
				}
				if (!res || res.sheets.length === 0) {
					setPreviewFailed(true);
					return;
				}
				setPreviewFailed(false);
				setSheetData(res);
			});
		} else if (meta.category === "text") {
			void loadTextPreview(source, cacheKey).then((res) => {
				if (cancelled) {
					return;
				}
				if (!res) {
					setPreviewFailed(true);
					return;
				}
				setPreviewFailed(false);
				setTextData(res);
			});
		} else if (meta.category === "pdf") {
			const canvas = pdfCanvasRef.current;
			if (!canvas) {
				return;
			}
			void renderPdfPageToCanvas(source, cacheKey, canvas, 1, 340)
				.then((res) => {
					if (cancelled) {
						return;
					}
					if (!res) {
						setPreviewFailed(true);
						return;
					}
					setPreviewFailed(false);
					setPdfReady(true);
				})
				.catch(() => {
					if (!cancelled) {
						setPreviewFailed(true);
					}
				});
		}

		return () => {
			cancelled = true;
		};
	}, [cacheKey, isNearViewport, meta.category, meta.supportsMiniPreview, source, withinSizeLimit]);

	const isLoadingPreview =
		meta.supportsMiniPreview &&
		withinSizeLimit &&
		Boolean(source) &&
		!previewFailed &&
		(meta.category === "excel" || meta.category === "csv"
			? !sheetData
			: meta.category === "text"
				? !textData
				: meta.category === "pdf"
					? !pdfReady
					: false);

	const mediaSrc = resolvedUrl || localMediaUrl;

	const canShowMiniPreview =
		meta.supportsMiniPreview && withinSizeLimit && Boolean(source) && !previewFailed;

	const firstSheetRows = sheetData?.sheets[0]?.rows ?? [];
	const miniRows = firstSheetRows.slice(0, 8);
	const miniColCount = Math.min(6, Math.max(3, ...miniRows.map((row) => Math.min(6, row.length))));

	const isModalPreviewable =
		variant === "message" &&
		Boolean(onOpenPreview) &&
		(meta.category === "excel" ||
			meta.category === "csv" ||
			meta.category === "pdf" ||
			meta.category === "text");

	return (
		<div
			ref={cardRef}
			className="w-[320px] max-w-full overflow-hidden rounded-xl border border-border/80 bg-card/90 text-foreground shadow-xs sm:w-[356px]"
		>
			{/* Header */}
			<div className="flex items-start gap-2.5 px-3 py-2.5">
				<div
					className={cn(
						"mt-0.5 flex size-9 shrink-0 flex-col items-center justify-center rounded-lg text-[10px] font-bold tracking-tight shadow-xs",
						meta.badgeClass,
					)}
				>
					{status === "uploading" ? (
						<Loader2 className="size-4 animate-spin" />
					) : status === "error" ? (
						<AlertCircle className="size-4" />
					) : (
						<>
							<CategoryIcon category={meta.category} />
							<span className="mt-0.5 text-[8px] leading-none opacity-90">{meta.badgeText}</span>
						</>
					)}
				</div>

				<div className="min-w-0 flex-1">
					<TooltipProvider delay={250}>
						<Tooltip>
							<TooltipTrigger
								render={
									<div className="truncate text-[13px] font-medium leading-snug text-foreground">
										{filename}
									</div>
								}
							/>
							<TooltipContent side="top" className="break-all">
								{filename}
							</TooltipContent>
						</Tooltip>
					</TooltipProvider>
					<div className="mt-0.5 flex flex-wrap items-center gap-1.5 text-[11px] text-muted-foreground">
						{status === "uploading" ? (
							<span>{t("chat.attachment_uploading")}</span>
						) : status === "error" ? (
							<span className="text-destructive">
								{errorMessage || t("chat.attachment_upload_failed")}
							</span>
						) : (
							<>
								<span>{t(meta.typeKey)}</span>
								<span aria-hidden>•</span>
								<span>{formatAttachmentBytes(sizeBytes)}</span>
							</>
						)}
					</div>
				</div>

				<div className="flex shrink-0 items-center gap-0.5">
					{variant === "composer" && status === "error" && onRetry ? (
						<button
							type="button"
							onClick={onRetry}
							title={t("chat.attachment_retry")}
							aria-label={t("chat.attachment_retry")}
							className="inline-flex size-7 items-center justify-center rounded-md text-muted-foreground transition hover:bg-muted hover:text-foreground"
						>
							<RotateCw className="size-3.5" />
						</button>
					) : null}

					{variant === "composer" && onRemove ? (
						<button
							type="button"
							onClick={onRemove}
							title={t("chat.attachment_remove")}
							aria-label={t("chat.attachment_remove")}
							className="inline-flex size-7 items-center justify-center rounded-md text-muted-foreground transition hover:bg-muted hover:text-foreground"
						>
							<X className="size-4" />
						</button>
					) : null}

					{variant === "message" && isModalPreviewable ? (
						<button
							type="button"
							onClick={onOpenPreview}
							title={t("chat.attachment_preview")}
							aria-label={t("chat.attachment_preview")}
							className="inline-flex size-7 items-center justify-center rounded-md text-muted-foreground transition hover:bg-muted hover:text-foreground"
						>
							<Maximize2 className="size-3.5" />
						</button>
					) : null}

					{variant === "message" && resolvedUrl ? (
						<a
							href={resolvedUrl}
							download={filename}
							target="_blank"
							rel="noopener noreferrer"
							title={t("chat.attachment_download")}
							aria-label={t("chat.attachment_download")}
							className="inline-flex size-7 items-center justify-center rounded-md text-muted-foreground transition hover:bg-muted hover:text-foreground"
						>
							<Download className="size-3.5" />
						</a>
					) : null}
				</div>
			</div>

			{/* Mini Preview */}
			{canShowMiniPreview ? (
				<div
					className={cn(
						"relative border-t border-border/70",
						isModalPreviewable && "cursor-pointer",
					)}
					onClick={isModalPreviewable ? onOpenPreview : undefined}
				>
					{(meta.category === "excel" || meta.category === "csv") && (
						<div className="relative h-[136px] overflow-hidden bg-white text-zinc-900">
							{isLoadingPreview && !sheetData ? (
								<div className="flex h-full items-center justify-center gap-2 text-xs text-zinc-500">
									<Loader2 className="size-3.5 animate-spin" />
									{t("chat.attachment_preview_loading")}
								</div>
							) : (
								<table className="w-full border-collapse text-[10px] leading-tight">
									<thead>
										<tr className="bg-zinc-100 text-zinc-500">
											<th className="w-6 border-b border-r border-zinc-200 px-1 py-0.5 text-center font-normal" />
											{Array.from({ length: miniColCount }, (_, colIdx) => (
												<th
													key={colIdx}
													className="border-b border-r border-zinc-200 px-1.5 py-0.5 text-center font-medium"
												>
													{columnIndexToLabel(colIdx)}
												</th>
											))}
										</tr>
									</thead>
									<tbody>
										{miniRows.map((row, rowIdx) => (
											<tr key={rowIdx} className="odd:bg-white even:bg-zinc-50/50">
												<td className="border-b border-r border-zinc-200 bg-zinc-100 px-1 py-0.5 text-center text-zinc-500">
													{rowIdx + 1}
												</td>
												{Array.from({ length: miniColCount }, (_, colIdx) => (
													<td
														key={colIdx}
														className="max-w-[90px] truncate border-b border-r border-zinc-200 px-1.5 py-0.5 text-zinc-800"
													>
														{row[colIdx] ?? ""}
													</td>
												))}
											</tr>
										))}
									</tbody>
								</table>
							)}
							<div className="pointer-events-none absolute inset-x-0 bottom-0 h-5 bg-gradient-to-t from-white/95 to-transparent" />
						</div>
					)}

					{meta.category === "pdf" && (
						<div className="relative flex max-h-[164px] items-center justify-center overflow-hidden bg-white">
							{isLoadingPreview && !pdfReady ? (
								<div className="flex h-[84px] items-center justify-center gap-2 text-xs text-zinc-500">
									<Loader2 className="size-3.5 animate-spin" />
									{t("chat.attachment_preview_loading")}
								</div>
							) : null}
							<canvas
								ref={pdfCanvasRef}
								className={cn("block max-w-full", !pdfReady && "hidden")}
							/>
							{pdfReady ? (
								<div className="pointer-events-none absolute inset-x-0 bottom-0 h-5 bg-gradient-to-t from-white/90 to-transparent" />
							) : null}
						</div>
					)}

					{meta.category === "text" && (
						<div className="relative max-h-[116px] overflow-hidden bg-muted/40 px-3 py-2 font-mono text-[11px] leading-[1.45] text-foreground/90">
							{isLoadingPreview && !textData ? (
								<div className="flex h-[56px] items-center justify-center gap-2 font-sans text-xs text-muted-foreground">
									<Loader2 className="size-3.5 animate-spin" />
									{t("chat.attachment_preview_loading")}
								</div>
							) : (
								<pre className="whitespace-pre-wrap break-words">
									{(textData?.lines ?? []).slice(0, 8).join("\n")}
								</pre>
							)}
							<div className="pointer-events-none absolute inset-x-0 bottom-0 h-5 bg-gradient-to-t from-card/90 to-transparent" />
						</div>
					)}

					{meta.category === "audio" && mediaSrc && (
						<div className="bg-muted/30 px-3 py-2" onClick={(e) => e.stopPropagation()}>
							<audio controls preload="metadata" src={mediaSrc} className="h-8 w-full" />
						</div>
					)}

					{meta.category === "video" && mediaSrc && (
						<div className="bg-black" onClick={(e) => e.stopPropagation()}>
							<video controls preload="metadata" src={mediaSrc} className="max-h-[180px] w-full" />
						</div>
					)}
				</div>
			) : null}
		</div>
	);
}
