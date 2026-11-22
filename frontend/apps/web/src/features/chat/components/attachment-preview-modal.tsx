"use client";

import hljs from "highlight.js/lib/common";
import { Check, ChevronLeft, ChevronRight, Copy, Download, Loader2, X } from "lucide-react";
import { useEffect, useMemo, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { useTranslation } from "react-i18next";
import { formatAttachmentBytes } from "@/features/chat/components/file-attachment-card";
import { isImageContentType } from "@/features/chat/components/composer-attachment-node";
import {
	columnIndexToLabel,
	getAttachmentVisualMeta,
	loadSpreadsheetPreview,
	loadTextPreview,
	renderPdfPageToCanvas,
	type SpreadsheetPreviewData,
	type TextPreviewData,
} from "@/features/chat/lib/attachment-preview";
import type { InlineAttachmentPayload } from "@/features/chat/lib/composer-markdown";
import { resolvePublicObjectUrl } from "@/lib/public-object-url";
import { cn } from "@/lib/utils";

type AttachmentPreviewModalProps = {
	attachment: InlineAttachmentPayload;
	onClose: () => void;
};

export function AttachmentPreviewModal({ attachment, onClose }: AttachmentPreviewModalProps) {
	const { t } = useTranslation();
	const resolvedUrl = resolvePublicObjectUrl(attachment.url);
	const isImage = isImageContentType(attachment.content_type);
	const meta = useMemo(
		() => getAttachmentVisualMeta(attachment.filename, attachment.content_type),
		[attachment.content_type, attachment.filename],
	);

	const [sheetData, setSheetData] = useState<SpreadsheetPreviewData | null>(null);
	const [activeSheetIdx, setActiveSheetIdx] = useState(0);
	const [textData, setTextData] = useState<TextPreviewData | null>(null);
	const [pdfPage, setPdfPage] = useState(1);
	const [pdfRenderedPage, setPdfRenderedPage] = useState<number | null>(null);
	const [pdfTotalPages, setPdfTotalPages] = useState(1);
	const [loadFailed, setLoadFailed] = useState(false);
	const [copied, setCopied] = useState(false);

	const pdfCanvasRef = useRef<HTMLCanvasElement>(null);
	const codeRef = useRef<HTMLElement>(null);

	useEffect(() => {
		const onKeyDown = (e: KeyboardEvent) => {
			if (e.key === "Escape") {
				onClose();
			}
		};
		window.addEventListener("keydown", onKeyDown);
		return () => window.removeEventListener("keydown", onKeyDown);
	}, [onClose]);

	useEffect(() => {
		if (isImage) {
			return;
		}
		let cancelled = false;
		const cacheKey = `${attachment.id}:${resolvedUrl}`;

		if (meta.category === "excel" || meta.category === "csv") {
			void loadSpreadsheetPreview(resolvedUrl, meta.category, cacheKey).then((res) => {
				if (!cancelled) {
					setSheetData(res);
					setActiveSheetIdx(0);
					setLoadFailed(!res);
				}
			});
		} else if (meta.category === "text") {
			void loadTextPreview(resolvedUrl, cacheKey).then((res) => {
				if (!cancelled) {
					setTextData(res);
					setLoadFailed(!res);
				}
			});
		} else if (meta.category === "pdf") {
			const canvas = pdfCanvasRef.current;
			if (!canvas) {
				return;
			}
			void renderPdfPageToCanvas(resolvedUrl, cacheKey, canvas, pdfPage, 760)
				.then((res) => {
					if (!cancelled) {
						if (res) {
							setPdfTotalPages(res.numPages);
							setPdfRenderedPage(pdfPage);
							setLoadFailed(false);
						} else {
							setLoadFailed(true);
						}
					}
				})
				.catch(() => {
					if (!cancelled) {
						setLoadFailed(true);
					}
				});
		}

		return () => {
			cancelled = true;
		};
	}, [attachment.id, isImage, meta.category, pdfPage, resolvedUrl]);

	const loading =
		!isImage &&
		!loadFailed &&
		(meta.category === "excel" || meta.category === "csv"
			? !sheetData
			: meta.category === "text"
				? !textData
				: meta.category === "pdf"
					? pdfRenderedPage !== pdfPage
					: false);

	useEffect(() => {
		const el = codeRef.current;
		if (!el || !textData) {
			return;
		}
		el.removeAttribute("data-highlighted");
		try {
			hljs.highlightElement(el);
		} catch {
			// Fallback to plain monospace
		}
	}, [textData]);

	if (typeof document === "undefined") {
		return null;
	}

	const activeSheet = sheetData?.sheets[activeSheetIdx];
	const activeRows = activeSheet?.rows ?? [];
	const maxCols = Math.max(1, ...activeRows.map((r) => r.length));

	return createPortal(
		<div
			className="fixed inset-0 z-50 flex flex-col items-center justify-center bg-black/85 p-4 backdrop-blur-sm"
			onClick={onClose}
			role="dialog"
			aria-modal="true"
			aria-label={attachment.filename}
		>
			{/* Top bar */}
			<div
				className="flex w-full max-w-5xl items-center justify-between gap-4 pb-3 text-white"
				onClick={(e) => e.stopPropagation()}
			>
				<div className="min-w-0">
					<p className="truncate text-sm font-medium">{attachment.filename}</p>
					<p className="text-xs text-zinc-400">
						{isImage
							? formatAttachmentBytes(attachment.size_bytes)
							: `${t(meta.typeKey)} • ${formatAttachmentBytes(attachment.size_bytes)}`}
					</p>
				</div>
				<div className="flex shrink-0 items-center gap-2">
					{meta.category === "text" && textData ? (
						<button
							type="button"
							onClick={() => {
								void navigator.clipboard.writeText(textData.text);
								setCopied(true);
								window.setTimeout(() => setCopied(false), 1600);
							}}
							className="inline-flex items-center gap-1.5 rounded-lg bg-white/10 px-3 py-1.5 text-xs font-medium text-white transition hover:bg-white/20"
						>
							{copied ? (
								<Check className="size-3.5 text-emerald-400" />
							) : (
								<Copy className="size-3.5" />
							)}
							{copied ? t("chat.copied_code") : t("chat.copy_code")}
						</button>
					) : null}
					<a
						href={resolvedUrl}
						download={attachment.filename}
						target="_blank"
						rel="noopener noreferrer"
						className="inline-flex items-center gap-1.5 rounded-lg bg-white/10 px-3 py-1.5 text-xs font-medium text-white transition hover:bg-white/20"
					>
						<Download className="size-3.5" />
						{t("chat.attachment_download")}
					</a>
					<button
						type="button"
						onClick={onClose}
						aria-label={t("chat.attachment_close")}
						className="inline-flex size-8 items-center justify-center rounded-lg bg-white/10 text-white transition hover:bg-white/20"
					>
						<X className="size-4" />
					</button>
				</div>
			</div>

			{/* Content */}
			{isImage ? (
				/* eslint-disable-next-line @next/next/no-img-element */
				<img
					src={resolvedUrl}
					alt={attachment.filename}
					onClick={(e) => e.stopPropagation()}
					className="max-h-[82vh] max-w-full rounded-xl object-contain shadow-2xl"
				/>
			) : (
				<div
					onClick={(e) => e.stopPropagation()}
					className="flex max-h-[82vh] w-full max-w-5xl flex-col overflow-hidden rounded-xl border border-white/15 bg-card text-foreground shadow-2xl"
				>
					{loading ? (
						<div className="flex h-64 items-center justify-center gap-2 text-sm text-muted-foreground">
							<Loader2 className="size-4 animate-spin" />
							{t("chat.attachment_preview_loading")}
						</div>
					) : null}

					{(meta.category === "excel" || meta.category === "csv") && !loading ? (
						activeSheet ? (
							<>
								{sheetData && sheetData.sheets.length > 1 ? (
									<div className="flex items-center gap-1 border-b border-border bg-muted/40 px-3 py-1.5">
										{sheetData.sheets.map((s, idx) => (
											<button
												key={s.name}
												type="button"
												onClick={() => setActiveSheetIdx(idx)}
												className={cn(
													"rounded-md px-2.5 py-1 text-xs font-medium transition",
													idx === activeSheetIdx
														? "bg-primary text-primary-foreground"
														: "text-muted-foreground hover:bg-muted hover:text-foreground",
												)}
											>
												{s.name}
											</button>
										))}
									</div>
								) : null}
								<div className="thin-scroll flex-1 overflow-auto bg-white text-zinc-900">
									<table className="w-full border-collapse text-xs">
										<thead className="sticky top-0 z-10 bg-zinc-100 text-zinc-600">
											<tr>
												<th className="w-10 border-b border-r border-zinc-200 px-2 py-1.5 text-center font-normal" />
												{Array.from({ length: maxCols }, (_, colIdx) => (
													<th
														key={colIdx}
														className="min-w-[100px] border-b border-r border-zinc-200 px-2.5 py-1.5 text-center font-semibold"
													>
														{columnIndexToLabel(colIdx)}
													</th>
												))}
											</tr>
										</thead>
										<tbody>
											{activeRows.map((row, rowIdx) => (
												<tr key={rowIdx} className="odd:bg-white even:bg-zinc-50/60">
													<td className="border-b border-r border-zinc-200 bg-zinc-100 px-2 py-1 text-center text-zinc-500">
														{rowIdx + 1}
													</td>
													{Array.from({ length: maxCols }, (_, colIdx) => (
														<td
															key={colIdx}
															className="max-w-[260px] truncate border-b border-r border-zinc-200 px-2.5 py-1 text-zinc-800"
															title={row[colIdx] ?? ""}
														>
															{row[colIdx] ?? ""}
														</td>
													))}
												</tr>
											))}
										</tbody>
									</table>
								</div>
							</>
						) : (
							<div className="flex h-56 items-center justify-center text-sm text-muted-foreground">
								{t("chat.attachment_preview_unavailable")}
							</div>
						)
					) : null}

					{meta.category === "pdf" ? (
						<div className="flex flex-1 flex-col overflow-hidden">
							{pdfTotalPages > 1 ? (
								<div className="flex items-center justify-center gap-3 border-b border-border bg-muted/40 px-3 py-1.5 text-xs">
									<button
										type="button"
										disabled={pdfPage <= 1}
										onClick={() => setPdfPage((p) => Math.max(1, p - 1))}
										className="inline-flex size-7 items-center justify-center rounded-md hover:bg-muted disabled:opacity-40"
									>
										<ChevronLeft className="size-4" />
									</button>
									<span className="font-medium">
										{t("chat.attachment_page_of", { page: pdfPage, total: pdfTotalPages })}
									</span>
									<button
										type="button"
										disabled={pdfPage >= pdfTotalPages}
										onClick={() => setPdfPage((p) => Math.min(pdfTotalPages, p + 1))}
										className="inline-flex size-7 items-center justify-center rounded-md hover:bg-muted disabled:opacity-40"
									>
										<ChevronRight className="size-4" />
									</button>
								</div>
							) : null}
							<div className="thin-scroll flex flex-1 justify-center overflow-auto bg-zinc-100 p-4">
								<canvas ref={pdfCanvasRef} className="block rounded shadow-md" />
							</div>
						</div>
					) : null}

					{meta.category === "text" && !loading ? (
						textData ? (
							<pre className="thin-scroll flex-1 overflow-auto bg-[#0d1117] p-4 font-mono text-[13px] leading-relaxed text-[#e6edf3]">
								<code
									ref={codeRef}
									className={meta.languageHint ? `language-${meta.languageHint}` : undefined}
								>
									{textData.text}
								</code>
							</pre>
						) : (
							<div className="flex h-56 items-center justify-center text-sm text-muted-foreground">
								{t("chat.attachment_preview_unavailable")}
							</div>
						)
					) : null}
				</div>
			)}
		</div>,
		document.body,
	);
}
