"use client";

export const MAX_CLIENT_PREVIEW_BYTES = 10 * 1024 * 1024; // 10 MiB

export type AttachmentCategory =
	"excel" | "csv" | "pdf" | "word" | "powerpoint" | "archive" | "text" | "audio" | "video" | "file";

export type AttachmentVisualMeta = {
	category: AttachmentCategory;
	supportsMiniPreview: boolean;
	typeKey: string;
	badgeText: string;
	badgeClass: string;
	languageHint?: string;
};

export type SpreadsheetSheet = {
	name: string;
	rows: string[][];
};

export type SpreadsheetPreviewData = {
	sheets: SpreadsheetSheet[];
};

export type TextPreviewData = {
	text: string;
	lines: string[];
	truncated: boolean;
};

const composerFileRegistry = new Map<string, File>();

export function registerComposerFile(localId: string, file: File) {
	composerFileRegistry.set(localId, file);
}

export function getComposerFile(localId: string): File | undefined {
	return composerFileRegistry.get(localId);
}

export function unregisterComposerFile(localId: string) {
	composerFileRegistry.delete(localId);
}

export function clearComposerFiles() {
	composerFileRegistry.clear();
}

export function columnIndexToLabel(index: number): string {
	let n = index;
	let label = "";
	while (n >= 0) {
		label = String.fromCharCode((n % 26) + 65) + label;
		n = Math.floor(n / 26) - 1;
	}
	return label;
}

function getExtension(filename: string): string {
	const dot = filename.lastIndexOf(".");
	if (dot < 0 || dot === filename.length - 1) {
		return "";
	}
	return filename.slice(dot + 1).toLowerCase();
}

export function getAttachmentVisualMeta(
	filename: string,
	contentType: string,
): AttachmentVisualMeta {
	const ext = getExtension(filename);
	const mime = contentType.toLowerCase();

	if (
		ext === "xlsx" ||
		ext === "xls" ||
		mime.includes("spreadsheetml") ||
		mime === "application/vnd.ms-excel"
	) {
		return {
			category: "excel",
			supportsMiniPreview: ext === "xlsx" || mime.includes("spreadsheetml"),
			typeKey: "chat.attachment_type_excel",
			badgeText: "X",
			badgeClass: "bg-[#107C41] text-white",
		};
	}

	if (ext === "csv" || mime === "text/csv") {
		return {
			category: "csv",
			supportsMiniPreview: true,
			typeKey: "chat.attachment_type_csv",
			badgeText: "CSV",
			badgeClass: "bg-[#107C41] text-white",
		};
	}

	if (ext === "pdf" || mime === "application/pdf") {
		return {
			category: "pdf",
			supportsMiniPreview: true,
			typeKey: "chat.attachment_type_pdf",
			badgeText: "PDF",
			badgeClass: "bg-[#D9383A] text-white",
		};
	}

	if (
		ext === "docx" ||
		ext === "doc" ||
		mime.includes("wordprocessingml") ||
		mime === "application/msword"
	) {
		return {
			category: "word",
			supportsMiniPreview: false,
			typeKey: "chat.attachment_type_word",
			badgeText: "W",
			badgeClass: "bg-[#2B579A] text-white",
		};
	}

	if (
		ext === "pptx" ||
		ext === "ppt" ||
		mime.includes("presentationml") ||
		mime === "application/vnd.ms-powerpoint"
	) {
		return {
			category: "powerpoint",
			supportsMiniPreview: false,
			typeKey: "chat.attachment_type_powerpoint",
			badgeText: "P",
			badgeClass: "bg-[#D24726] text-white",
		};
	}

	if (
		ext === "zip" ||
		ext === "gz" ||
		ext === "tar" ||
		ext === "rar" ||
		ext === "7z" ||
		mime.includes("zip") ||
		mime.includes("gzip")
	) {
		return {
			category: "archive",
			supportsMiniPreview: false,
			typeKey: "chat.attachment_type_archive",
			badgeText: ext ? ext.toUpperCase() : "ZIP",
			badgeClass: "bg-amber-600 text-white",
		};
	}

	if (mime.startsWith("video/") || ext === "mp4" || ext === "webm") {
		return {
			category: "video",
			supportsMiniPreview: true,
			typeKey: "chat.attachment_type_video",
			badgeText: "VID",
			badgeClass: "bg-indigo-600 text-white",
		};
	}

	if (mime.startsWith("audio/") || ext === "mp3" || ext === "wav" || ext === "ogg") {
		return {
			category: "audio",
			supportsMiniPreview: true,
			typeKey: "chat.attachment_type_audio",
			badgeText: "AUD",
			badgeClass: "bg-violet-600 text-white",
		};
	}

	if (
		mime.startsWith("text/") ||
		mime === "application/json" ||
		ext === "txt" ||
		ext === "md" ||
		ext === "json" ||
		ext === "log"
	) {
		return {
			category: "text",
			supportsMiniPreview: true,
			typeKey: "chat.attachment_type_text",
			badgeText: ext ? ext.toUpperCase().slice(0, 4) : "TXT",
			badgeClass: "bg-sky-600 text-white",
			languageHint: ext === "json" ? "json" : ext === "md" ? "markdown" : undefined,
		};
	}

	return {
		category: "file",
		supportsMiniPreview: false,
		typeKey: "chat.attachment_type_file",
		badgeText: ext ? ext.toUpperCase().slice(0, 4) : "FILE",
		badgeClass: "bg-zinc-600 text-white",
	};
}

function parseCsvRows(text: string, maxRows = 100, maxCols = 20): string[][] {
	const rows: string[][] = [];
	let currentRow: string[] = [];
	let cell = "";
	let inQuotes = false;

	for (let i = 0; i < text.length; i += 1) {
		const ch = text[i];
		const next = text[i + 1];

		if (inQuotes) {
			if (ch === '"' && next === '"') {
				cell += '"';
				i += 1;
			} else if (ch === '"') {
				inQuotes = false;
			} else {
				cell += ch;
			}
			continue;
		}

		if (ch === '"') {
			inQuotes = true;
		} else if (ch === ",") {
			if (currentRow.length < maxCols) {
				currentRow.push(cell);
			}
			cell = "";
		} else if (ch === "\r" && next === "\n") {
			if (currentRow.length < maxCols) {
				currentRow.push(cell);
			}
			rows.push(currentRow);
			currentRow = [];
			cell = "";
			i += 1;
			if (rows.length >= maxRows) {
				break;
			}
		} else if (ch === "\n" || ch === "\r") {
			if (currentRow.length < maxCols) {
				currentRow.push(cell);
			}
			rows.push(currentRow);
			currentRow = [];
			cell = "";
			if (rows.length >= maxRows) {
				break;
			}
		} else {
			cell += ch;
		}
	}

	if ((cell.length > 0 || currentRow.length > 0) && rows.length < maxRows) {
		if (currentRow.length < maxCols) {
			currentRow.push(cell);
		}
		rows.push(currentRow);
	}

	return rows;
}

function formatCellValue(value: unknown): string {
	if (value === null || value === undefined) {
		return "";
	}
	if (value instanceof Date) {
		return Number.isNaN(value.getTime()) ? "" : value.toISOString().slice(0, 10);
	}
	return String(value);
}

const MAX_PREVIEW_CACHE_ENTRIES = 30;
const spreadsheetCache = new Map<string, Promise<SpreadsheetPreviewData | null>>();
const resolvedSpreadsheetCache = new Map<string, SpreadsheetPreviewData>();
const textCache = new Map<string, Promise<TextPreviewData | null>>();
const resolvedTextCache = new Map<string, TextPreviewData>();
const pdfBytesCache = new Map<string, Promise<ArrayBuffer | null>>();

export function getCachedSpreadsheetPreview(cacheKey: string): SpreadsheetPreviewData | null {
	return resolvedSpreadsheetCache.get(cacheKey) ?? null;
}

export function getCachedTextPreview(cacheKey: string): TextPreviewData | null {
	return resolvedTextCache.get(cacheKey) ?? null;
}

function setBoundedCache<T>(cache: Map<string, T>, key: string, value: T) {
	if (cache.size >= MAX_PREVIEW_CACHE_ENTRIES) {
		const oldestKey = cache.keys().next().value;
		if (oldestKey !== undefined) {
			cache.delete(oldestKey);
		}
	}
	cache.set(key, value);
}

async function readArrayBuffer(source: File | string): Promise<ArrayBuffer> {
	if (typeof source !== "string") {
		return source.arrayBuffer();
	}
	const res = await fetch(source);
	if (!res.ok) {
		throw new Error(`HTTP ${res.status}`);
	}
	return res.arrayBuffer();
}

export function loadSpreadsheetPreview(
	source: File | string,
	category: "excel" | "csv",
	cacheKey: string,
): Promise<SpreadsheetPreviewData | null> {
	const existing = spreadsheetCache.get(cacheKey);
	if (existing) {
		return existing;
	}

	const task = (async (): Promise<SpreadsheetPreviewData | null> => {
		try {
			const buf = await readArrayBuffer(source);
			if (buf.byteLength > MAX_CLIENT_PREVIEW_BYTES) {
				return null;
			}

			if (category === "csv") {
				const text = new TextDecoder("utf-8").decode(buf.slice(0, 256 * 1024));
				const rows = parseCsvRows(text, 100, 20);
				const data = { sheets: [{ name: "CSV", rows }] };
				setBoundedCache(resolvedSpreadsheetCache, cacheKey, data);
				return data;
			}

			const mod = await import("read-excel-file/browser");
			const blob = new Blob([buf]);
			const rawSheets = await mod.default(blob);
			const sheets: SpreadsheetSheet[] = rawSheets.slice(0, 5).map((rawSheet) => ({
				name: rawSheet.sheet,
				rows: rawSheet.data
					.slice(0, 100)
					.map((row) => row.slice(0, 20).map((cell) => formatCellValue(cell))),
			}));

			const data = { sheets };
			setBoundedCache(resolvedSpreadsheetCache, cacheKey, data);
			return data;
		} catch {
			return null;
		}
	})();

	setBoundedCache(spreadsheetCache, cacheKey, task);
	return task;
}

async function readUpToBytes(res: Response, maxBytes: number): Promise<{ bytes: Uint8Array; truncated: boolean }> {
	if (!res.body) {
		const buf = await res.arrayBuffer();
		return {
			bytes: new Uint8Array(buf.slice(0, maxBytes)),
			truncated: buf.byteLength > maxBytes,
		};
	}

	const reader = res.body.getReader();
	const chunks: Uint8Array[] = [];
	let total = 0;
	let truncated = false;

	while (true) {
		const { done, value } = await reader.read();
		if (done) {
			break;
		}
		if (total + value.byteLength > maxBytes) {
			const remaining = maxBytes - total;
			if (remaining > 0) {
				chunks.push(value.subarray(0, remaining));
				total += remaining;
			}
			truncated = true;
			await reader.cancel();
			break;
		}
		chunks.push(value);
		total += value.byteLength;
	}

	const merged = new Uint8Array(total);
	let offset = 0;
	for (const chunk of chunks) {
		merged.set(chunk, offset);
		offset += chunk.byteLength;
	}
	return { bytes: merged, truncated };
}

export function loadTextPreview(
	source: File | string,
	cacheKey: string,
): Promise<TextPreviewData | null> {
	const existing = textCache.get(cacheKey);
	if (existing) {
		return existing;
	}

	const task = (async (): Promise<TextPreviewData | null> => {
		try {
			const MAX_TEXT_BYTES = 32 * 1024;
			let text = "";
			let truncated = false;

			if (typeof source !== "string") {
				truncated = source.size > MAX_TEXT_BYTES;
				text = await source.slice(0, MAX_TEXT_BYTES).text();
			} else {
				const res = await fetch(source);
				if (!res.ok) {
					return null;
				}
				const result = await readUpToBytes(res, MAX_TEXT_BYTES);
				truncated = result.truncated;
				text = new TextDecoder("utf-8").decode(result.bytes);
			}

			const lines = text.replace(/\r\n/g, "\n").split("\n");
			const data = { text, lines, truncated };
			setBoundedCache(resolvedTextCache, cacheKey, data);
			return data;
		} catch {
			return null;
		}
	})();

	setBoundedCache(textCache, cacheKey, task);
	return task;
}

function loadPdfBytes(source: File | string, cacheKey: string): Promise<ArrayBuffer | null> {
	const existing = pdfBytesCache.get(cacheKey);
	if (existing) {
		return existing;
	}
	const task = readArrayBuffer(source).catch(() => null);
	setBoundedCache(pdfBytesCache, cacheKey, task);
	return task;
}

export async function renderPdfPageToCanvas(
	source: File | string,
	cacheKey: string,
	canvas: HTMLCanvasElement,
	pageNumber = 1,
	maxWidth = 340,
): Promise<{ numPages: number } | null> {
	const bytes = await loadPdfBytes(source, cacheKey);
	if (!bytes || bytes.byteLength > MAX_CLIENT_PREVIEW_BYTES) {
		return null;
	}

	const pdfjsLib = await import("pdfjs-dist");
	if (!pdfjsLib.GlobalWorkerOptions.workerSrc) {
		pdfjsLib.GlobalWorkerOptions.workerSrc = new URL(
			"pdfjs-dist/build/pdf.worker.min.mjs",
			import.meta.url,
		).toString();
	}

	// Clone buffer slice because pdf.js transfers/detaches the ArrayBuffer to the worker
	const loadingTask = pdfjsLib.getDocument({ data: bytes.slice(0) });
	const pdf = await loadingTask.promise;
	const clampedPage = Math.max(1, Math.min(pageNumber, pdf.numPages));
	const page = await pdf.getPage(clampedPage);

	const unscaledViewport = page.getViewport({ scale: 1 });
	const scale = Math.max(0.5, maxWidth / Math.max(1, unscaledViewport.width));
	const dpr = typeof window !== "undefined" ? Math.min(window.devicePixelRatio || 1, 2) : 1;
	const viewport = page.getViewport({ scale: scale * dpr });

	const ctx = canvas.getContext("2d");
	if (!ctx) {
		await loadingTask.destroy();
		return null;
	}

	canvas.width = Math.floor(viewport.width);
	canvas.height = Math.floor(viewport.height);
	canvas.style.width = `${Math.floor(viewport.width / dpr)}px`;
	canvas.style.height = `${Math.floor(viewport.height / dpr)}px`;

	await page.render({ canvas, canvasContext: ctx, viewport }).promise;
	const numPages = pdf.numPages;
	await loadingTask.destroy();
	return { numPages };
}
