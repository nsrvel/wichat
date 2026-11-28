"use client";

import hljs from "highlight.js/lib/common";
import "highlight.js/styles/github-dark.css";
import { useQuery } from "@tanstack/react-query";
import { Check, Copy, Download, Maximize2 } from "lucide-react";
import {
	createContext,
	memo,
	useCallback,
	useContext,
	useEffect,
	useLayoutEffect,
	useMemo,
	useRef,
	useState,
	type ComponentPropsWithoutRef,
} from "react";
import { useTranslation } from "react-i18next";
import ReactMarkdown, { defaultUrlTransform, type Components } from "react-markdown";
import remarkGfm from "remark-gfm";
import { AttachmentPreviewModal } from "@/features/chat/components/attachment-preview-modal";
import { isImageContentType } from "@/features/chat/components/composer-attachment-node";
import { FileAttachmentCard } from "@/features/chat/components/file-attachment-card";
import { ReplyPreviewCard } from "@/features/chat/components/reply-preview-card";
import {
	parseInlineAttachmentPayload,
	parseInlineReplyPayload,
	type InlineAttachmentPayload,
} from "@/features/chat/lib/composer-markdown";
import {
	MentionMemberChip,
	mentionLabelFromChildren,
} from "@/features/chat/components/mention-member-chip";
import { remarkUnderline } from "@/features/chat/lib/remark-underline";
import {
	isMentionLinkHref,
	isRoleMentionLinkHref,
	mentionRoleIdFromHref,
	mentionUserIdFromHref,
} from "@/features/chat/lib/mention-markdown";
import { resolvePublicObjectUrl } from "@/lib/public-object-url";
import { cn } from "@/lib/utils";
import { groupRolesQueryOpts } from "@/services/group-role/group-role.query";

type MessageBodyProps = {
	text: string;
	className?: string;
	onReplyClick?: (messageId: string) => void;
	selfUserId?: string;
	groupId?: string;
	onMentionMemberClick?: (userId: string) => void;
};

type FittedImageDimensions = {
	width: number;
	height: number;
};

const MAX_ATTACHMENT_IMAGE_WIDTH = 320;
const MAX_ATTACHMENT_IMAGE_HEIGHT = 240;
const MAX_IMAGE_DIM_CACHE_ENTRIES = 500;
const imageDimensionCache = new Map<string, FittedImageDimensions>();

function computeFittedImageSize(
	naturalWidth: number,
	naturalHeight: number,
	maxWidth = MAX_ATTACHMENT_IMAGE_WIDTH,
	maxHeight = MAX_ATTACHMENT_IMAGE_HEIGHT,
): FittedImageDimensions | null {
	if (naturalWidth <= 0 || naturalHeight <= 0) {
		return null;
	}
	const scale = Math.min(1, maxWidth / naturalWidth, maxHeight / naturalHeight);
	return {
		width: Math.max(1, Math.round(naturalWidth * scale)),
		height: Math.max(1, Math.round(naturalHeight * scale)),
	};
}

function rememberImageDimensions(url: string, dims: FittedImageDimensions): void {
	if (!url) {
		return;
	}
	if (imageDimensionCache.has(url)) {
		imageDimensionCache.delete(url);
	} else if (imageDimensionCache.size >= MAX_IMAGE_DIM_CACHE_ENTRIES) {
		const oldestKey = imageDimensionCache.keys().next().value;
		if (oldestKey !== undefined) {
			imageDimensionCache.delete(oldestKey);
		}
	}
	imageDimensionCache.set(url, dims);
}

type MessageBodyContextValue = {
	hasReplyClick: boolean;
	onReplyClick: (messageId: string) => void;
	onOpenLightbox: (att: InlineAttachmentPayload) => void;
	selfUserId?: string;
	groupId?: string;
	onMentionMemberClick?: (userId: string) => void;
	roleColorById?: Map<string, string>;
};

const MessageBodyContext = createContext<MessageBodyContextValue | null>(null);

const CodeBlock = memo(function CodeBlock({ language, code }: { language: string; code: string }) {
	const { t } = useTranslation();
	const codeRef = useRef<HTMLElement>(null);
	const [copied, setCopied] = useState(false);

	useEffect(() => {
		const el = codeRef.current;
		if (!el) {
			return;
		}
		el.removeAttribute("data-highlighted");
		try {
			hljs.highlightElement(el);
		} catch {
			// Fallback to plain text if language is unknown
		}
	}, [code, language]);

	const handleCopy = () => {
		void navigator.clipboard.writeText(code);
		setCopied(true);
		window.setTimeout(() => setCopied(false), 1600);
	};

	return (
		<div className="my-2 overflow-hidden rounded-lg border border-border/80 bg-[#0d1117] text-[#e6edf3]">
			<div className="flex items-center justify-between border-b border-white/10 bg-white/[0.04] px-3 py-1 text-[11px] text-zinc-400">
				<span className="font-mono lowercase">{language || "code"}</span>
				<button
					type="button"
					onClick={handleCopy}
					className="inline-flex items-center gap-1 rounded px-1.5 py-0.5 text-[11px] font-medium text-zinc-300 transition hover:bg-white/10 hover:text-white"
				>
					{copied ? (
						<>
							<Check className="size-3 text-emerald-400" />
							{t("chat.copied_code")}
						</>
					) : (
						<>
							<Copy className="size-3" />
							{t("chat.copy_code")}
						</>
					)}
				</button>
			</div>
			<pre className="thin-scroll overflow-x-auto p-3 text-[13px] leading-relaxed">
				<code ref={codeRef} className={language ? `language-${language}` : undefined}>
					{code}
				</code>
			</pre>
		</div>
	);
});

type AttachmentPreviewProps = {
	id: string;
	url: string;
	filename: string;
	contentType: string;
	sizeBytes: number;
	onOpenLightbox: (att: InlineAttachmentPayload) => void;
};

/** GFM Markdown body with syntax-highlighted code blocks, safe for both human and AI Bot output. */
const AttachmentPreview = memo(function AttachmentPreview({
	id,
	url,
	filename,
	contentType,
	sizeBytes,
	onOpenLightbox,
}: AttachmentPreviewProps) {
	const { t } = useTranslation();
	const resolvedUrl = resolvePublicObjectUrl(url);
	const isImage = isImageContentType(contentType);
	const [dims, setDims] = useState<FittedImageDimensions | null>(
		() => imageDimensionCache.get(resolvedUrl) ?? null,
	);
	const [loaded, setLoaded] = useState(() => imageDimensionCache.has(resolvedUrl));
	const [errored, setErrored] = useState(false);

	const openLightbox = useCallback(() => {
		onOpenLightbox({
			id,
			url,
			filename,
			content_type: contentType,
			size_bytes: sizeBytes,
		});
	}, [contentType, filename, id, onOpenLightbox, sizeBytes, url]);

	const handleImgRef = useCallback(
		(node: HTMLImageElement | null) => {
			if (!node || !node.complete || node.naturalWidth <= 0) {
				return;
			}
			const fitted = computeFittedImageSize(node.naturalWidth, node.naturalHeight);
			if (!fitted) {
				return;
			}
			const existing = imageDimensionCache.get(resolvedUrl);
			if (!existing || existing.width !== fitted.width || existing.height !== fitted.height) {
				rememberImageDimensions(resolvedUrl, fitted);
				setDims(fitted);
			}
			setLoaded(true);
		},
		[resolvedUrl],
	);

	if (isImage) {
		const activeDims = dims ?? imageDimensionCache.get(resolvedUrl) ?? null;
		const reservePlaceholder = !activeDims && !loaded && !errored;

		return (
			<div
				className={cn(
					"group/img relative my-2 inline-block w-fit max-w-[min(320px,100%)] overflow-hidden rounded-xl border border-border/60 bg-muted/20",
					reservePlaceholder && "min-h-[140px] min-w-[180px]",
				)}
			>
				<button
					type="button"
					onClick={openLightbox}
					className="block cursor-zoom-in focus:outline-none"
					aria-label={t("chat.attachment_expand")}
				>
					{/* eslint-disable-next-line @next/next/no-img-element */}
					<img
						ref={handleImgRef}
						src={resolvedUrl}
						alt={filename}
						width={activeDims?.width}
						height={activeDims?.height}
						decoding="async"
						draggable={false}
						onLoad={(event) => {
							const img = event.currentTarget;
							const fitted = computeFittedImageSize(img.naturalWidth, img.naturalHeight);
							if (fitted) {
								rememberImageDimensions(resolvedUrl, fitted);
								setDims((prev) =>
									prev && prev.width === fitted.width && prev.height === fitted.height
										? prev
										: fitted,
								);
							}
							setLoaded(true);
						}}
						onError={() => {
							setErrored(true);
							setLoaded(true);
						}}
						style={
							activeDims
								? {
										width: `${activeDims.width}px`,
										aspectRatio: `${activeDims.width} / ${activeDims.height}`,
									}
								: undefined
						}
						className="block max-h-[240px] w-auto max-w-full rounded-xl object-contain"
					/>
				</button>
				<div className="pointer-events-none absolute right-2 top-2 flex items-center gap-1 rounded-lg border border-white/10 bg-black/70 p-1 opacity-0 shadow-md backdrop-blur-sm transition-opacity group-hover/img:pointer-events-auto group-hover/img:opacity-100">
					<button
						type="button"
						onClick={openLightbox}
						title={t("chat.attachment_expand")}
						aria-label={t("chat.attachment_expand")}
						className="inline-flex size-7 items-center justify-center rounded-md text-zinc-200 transition hover:bg-white/15 hover:text-white"
					>
						<Maximize2 className="size-3.5" />
					</button>
					<a
						href={resolvedUrl}
						download={filename}
						target="_blank"
						rel="noopener noreferrer"
						title={t("chat.attachment_download")}
						aria-label={t("chat.attachment_download")}
						className="inline-flex size-7 items-center justify-center rounded-md text-zinc-200 transition hover:bg-white/15 hover:text-white"
					>
						<Download className="size-3.5" />
					</a>
				</div>
			</div>
		);
	}

	return (
		<div className="my-2">
			<FileAttachmentCard
				variant="message"
				id={id}
				filename={filename}
				contentType={contentType}
				sizeBytes={sizeBytes}
				url={url}
				onOpenPreview={openLightbox}
			/>
		</div>
	);
});

function MarkdownLink({ href, children, className, ...props }: ComponentPropsWithoutRef<"a">) {
	const ctx = useContext(MessageBodyContext);
	if (isMentionLinkHref(href)) {
		const userId = mentionUserIdFromHref(href ?? "");
		const label = mentionLabelFromChildren(children).trim() || userId;
		return (
			<MentionMemberChip
				userId={userId}
				label={label}
				selfUserId={ctx?.selfUserId}
				groupId={ctx?.groupId}
				onOpenProfile={ctx?.onMentionMemberClick}
				className={className}
			/>
		);
	}
	if (isRoleMentionLinkHref(href)) {
		const roleId = mentionRoleIdFromHref(href ?? "");
		const cleanLabel = mentionLabelFromChildren(children).trim().replace(/^@+/, "");
		const roleColor = ctx?.roleColorById?.get(roleId) ?? "#99aab5";
		return (
			<span
				className={cn("mention-chip inline rounded-md px-1 font-medium no-underline", className)}
				style={{
					backgroundColor: `color-mix(in oklab, ${roleColor} 22%, transparent)`,
					color: roleColor,
				}}
			>
				@{cleanLabel}
			</span>
		);
	}
	const safeHref = href && /^(https?:\/\/|mailto:|\/)/i.test(href) ? href : undefined;
	return (
		<a
			{...props}
			href={safeHref}
			target="_blank"
			rel="noopener noreferrer"
			className={cn(
				"font-medium text-primary underline decoration-primary/40 underline-offset-2 hover:decoration-primary",
				className,
			)}
		>
			{children}
		</a>
	);
}

function MarkdownPre({ children }: ComponentPropsWithoutRef<"pre">) {
	return <>{children}</>;
}

function MarkdownImg({ src, alt }: ComponentPropsWithoutRef<"img">) {
	const rawSrc = typeof src === "string" ? src.trim() : "";
	const safeSrc = rawSrc && /^(https?:\/\/|\/)/i.test(rawSrc) ? resolvePublicObjectUrl(rawSrc) : "";
	const [dims, setDims] = useState<FittedImageDimensions | null>(() =>
		safeSrc ? (imageDimensionCache.get(safeSrc) ?? null) : null,
	);
	const [loaded, setLoaded] = useState(() => Boolean(safeSrc && imageDimensionCache.has(safeSrc)));
	const [errored, setErrored] = useState(false);

	if (!safeSrc) {
		return null;
	}

	const activeDims = dims ?? imageDimensionCache.get(safeSrc) ?? null;
	const reservePlaceholder = !activeDims && !loaded && !errored;

	return (
		<span
			className={cn(
				"my-2 inline-block w-fit max-w-[min(320px,100%)] overflow-hidden rounded-xl border border-border/60 bg-muted/20",
				reservePlaceholder && "min-h-[140px] min-w-[180px]",
			)}
		>
			{/* eslint-disable-next-line @next/next/no-img-element */}
			<img
				src={safeSrc}
				alt={alt ?? ""}
				width={activeDims?.width}
				height={activeDims?.height}
				decoding="async"
				draggable={false}
				onLoad={(event) => {
					const img = event.currentTarget;
					const fitted = computeFittedImageSize(img.naturalWidth, img.naturalHeight);
					if (fitted) {
						rememberImageDimensions(safeSrc, fitted);
						setDims((prev) =>
							prev && prev.width === fitted.width && prev.height === fitted.height ? prev : fitted,
						);
					}
					setLoaded(true);
				}}
				onError={() => {
					setErrored(true);
					setLoaded(true);
				}}
				style={
					activeDims
						? {
								width: `${activeDims.width}px`,
								aspectRatio: `${activeDims.width} / ${activeDims.height}`,
							}
						: undefined
				}
				className="block max-h-[240px] w-auto max-w-full rounded-xl object-contain"
			/>
		</span>
	);
}

function MarkdownCode({
	className: codeClassName,
	children,
	...props
}: ComponentPropsWithoutRef<"code">) {
	const ctx = useContext(MessageBodyContext);
	const match = /language-([\w-]+)/.exec(codeClassName ?? "");
	const raw = String(children).replace(/\n$/, "");
	if (match?.[1] === "wichat-reply") {
		const reply = parseInlineReplyPayload(raw);
		if (reply) {
			return (
				<div className="my-1.5">
					<ReplyPreviewCard
						messageId={reply.id}
						authorName={reply.author}
						timeLabel={reply.time}
						bodyPreview={reply.preview}
						onClick={reply.id && ctx?.hasReplyClick ? () => ctx.onReplyClick(reply.id) : undefined}
					/>
				</div>
			);
		}
	}
	if (match?.[1] === "wichat-attachment") {
		const att = parseInlineAttachmentPayload(raw);
		if (att && ctx) {
			return (
				<AttachmentPreview
					id={att.id}
					url={att.url}
					filename={att.filename}
					contentType={att.content_type}
					sizeBytes={att.size_bytes}
					onOpenLightbox={ctx.onOpenLightbox}
				/>
			);
		}
	}
	const isBlock = Boolean(match) || raw.includes("\n");
	if (isBlock) {
		return <CodeBlock language={match?.[1] ?? ""} code={raw} />;
	}
	return (
		<code
			{...props}
			className="rounded bg-muted/80 px-1.5 py-0.5 font-mono text-[12.5px] text-foreground"
		>
			{children}
		</code>
	);
}

const REMARK_PLUGINS = [remarkGfm, remarkUnderline];

function messageMarkdownUrlTransform(url: string): string {
	if (isMentionLinkHref(url) || isRoleMentionLinkHref(url)) {
		return url;
	}
	return defaultUrlTransform(url);
}

const MARKDOWN_COMPONENTS: Components = {
	a: MarkdownLink,
	pre: MarkdownPre,
	code: MarkdownCode,
	img: MarkdownImg,
};

export const MessageBody = memo(function MessageBody({
	text,
	className,
	onReplyClick,
	selfUserId,
	groupId,
	onMentionMemberClick,
}: MessageBodyProps) {
	const [lightboxAttachment, setLightboxAttachment] = useState<InlineAttachmentPayload | null>(
		null,
	);
	const onReplyClickRef = useRef(onReplyClick);
	useLayoutEffect(() => {
		onReplyClickRef.current = onReplyClick;
	}, [onReplyClick]);

	const hasReplyClick = Boolean(onReplyClick);
	const rolesQuery = useQuery({
		...groupRolesQueryOpts(groupId ?? ""),
		enabled: Boolean(groupId),
	});
	const roleColorById = useMemo(() => {
		const map = new Map<string, string>();
		for (const role of rolesQuery.data?.roles ?? []) {
			if (role.id && role.color) {
				map.set(role.id, role.color);
			}
		}
		return map;
	}, [rolesQuery.data?.roles]);

	const contextValue = useMemo<MessageBodyContextValue>(
		() => ({
			hasReplyClick,
			onReplyClick: (messageId: string) => onReplyClickRef.current?.(messageId),
			onOpenLightbox: (item: InlineAttachmentPayload) => setLightboxAttachment(item),
			selfUserId,
			groupId,
			onMentionMemberClick,
			roleColorById,
		}),
		[groupId, hasReplyClick, onMentionMemberClick, roleColorById, selfUserId],
	);

	if (!text.trim()) {
		return null;
	}

	return (
		<div
			className={cn(
				"select-text break-words text-[14.5px] leading-[1.55] text-foreground [&_blockquote]:relative [&_blockquote]:my-1.5 [&_blockquote]:rounded-lg [&_blockquote]:border [&_blockquote]:border-border/80 [&_blockquote]:bg-muted/40 [&_blockquote]:py-2 [&_blockquote]:pl-4 [&_blockquote]:pr-3 [&_blockquote]:text-foreground [&_blockquote]:before:absolute [&_blockquote]:before:bottom-2 [&_blockquote]:before:left-1.5 [&_blockquote]:before:top-2 [&_blockquote]:before:w-[3px] [&_blockquote]:before:rounded-full [&_blockquote]:before:bg-foreground/45 [&_blockquote]:before:content-[''] [&_h1]:mt-2 [&_h1]:text-base [&_h1]:font-bold [&_h2]:mt-2 [&_h2]:text-[15px] [&_h2]:font-semibold [&_h3]:mt-1.5 [&_h3]:text-sm [&_h3]:font-semibold [&_li]:my-0.5 [&_ol]:my-1.5 [&_ol]:list-decimal [&_ol]:pl-5 [&_p+p]:mt-2 [&_table]:my-2 [&_table]:w-full [&_table]:border-collapse [&_table]:text-xs [&_td]:border [&_td]:border-border [&_td]:px-2.5 [&_td]:py-1 [&_th]:border [&_th]:border-border [&_th]:bg-muted/60 [&_th]:px-2.5 [&_th]:py-1 [&_th]:text-left [&_th]:font-semibold [&_u]:underline [&_u]:underline-offset-2 [&_ul]:my-1.5 [&_ul]:list-disc [&_ul]:pl-5",
				className,
			)}
			data-wichat-selectable="true"
		>
			<MessageBodyContext.Provider value={contextValue}>
				<ReactMarkdown
					remarkPlugins={REMARK_PLUGINS}
					components={MARKDOWN_COMPONENTS}
					urlTransform={messageMarkdownUrlTransform}
				>
					{text}
				</ReactMarkdown>
			</MessageBodyContext.Provider>
			{lightboxAttachment && (
				<AttachmentPreviewModal
					attachment={lightboxAttachment}
					onClose={() => setLightboxAttachment(null)}
				/>
			)}
		</div>
	);
});
