"use client";

import hljs from "highlight.js";
import "highlight.js/styles/github-dark.css";
import { Check, Copy, Download, Maximize2 } from "lucide-react";
import { memo, useEffect, useMemo, useRef, useState, type ComponentPropsWithoutRef } from "react";
import { useTranslation } from "react-i18next";
import ReactMarkdown, { type Components } from "react-markdown";
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
import { resolvePublicObjectUrl } from "@/lib/public-object-url";
import { cn } from "@/lib/utils";

type MessageBodyProps = {
	text: string;
	className?: string;
	onReplyClick?: (messageId: string) => void;
};

function CodeBlock({ language, code }: { language: string; code: string }) {
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
}

/** GFM Markdown body with syntax-highlighted code blocks, safe for both human and AI Bot output. */
function AttachmentPreview({
	attachment,
	onOpenLightbox,
}: {
	attachment: InlineAttachmentPayload;
	onOpenLightbox: (att: InlineAttachmentPayload) => void;
}) {
	const { t } = useTranslation();
	const resolvedUrl = resolvePublicObjectUrl(attachment.url);
	const isImage = isImageContentType(attachment.content_type);

	if (isImage) {
		return (
			<div className="group/img relative my-2 inline-block w-fit max-w-[min(320px,100%)] overflow-hidden rounded-xl border border-border/60 bg-muted/20">
				<button
					type="button"
					onClick={() => onOpenLightbox(attachment)}
					className="block cursor-zoom-in focus:outline-none"
					aria-label={t("chat.attachment_expand")}
				>
					{/* eslint-disable-next-line @next/next/no-img-element */}
					<img
						src={resolvedUrl}
						alt={attachment.filename}
						loading="lazy"
						className="block max-h-[240px] w-auto max-w-full rounded-xl object-contain"
					/>
				</button>
				<div className="pointer-events-none absolute right-2 top-2 flex items-center gap-1 rounded-lg border border-white/10 bg-black/70 p-1 opacity-0 shadow-md backdrop-blur-sm transition-opacity group-hover/img:pointer-events-auto group-hover/img:opacity-100">
					<button
						type="button"
						onClick={() => onOpenLightbox(attachment)}
						title={t("chat.attachment_expand")}
						aria-label={t("chat.attachment_expand")}
						className="inline-flex size-7 items-center justify-center rounded-md text-zinc-200 transition hover:bg-white/15 hover:text-white"
					>
						<Maximize2 className="size-3.5" />
					</button>
					<a
						href={resolvedUrl}
						download={attachment.filename}
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
				id={attachment.id}
				filename={attachment.filename}
				contentType={attachment.content_type}
				sizeBytes={attachment.size_bytes}
				url={attachment.url}
				onOpenPreview={() => onOpenLightbox(attachment)}
			/>
		</div>
	);
}

const REMARK_PLUGINS = [remarkGfm];

export const MessageBody = memo(function MessageBody({
	text,
	className,
	onReplyClick,
}: MessageBodyProps) {
	const [lightboxAttachment, setLightboxAttachment] = useState<InlineAttachmentPayload | null>(
		null,
	);

	const markdownComponents = useMemo<Components>(
		() => ({
			a({ href, children, ...props }: ComponentPropsWithoutRef<"a">) {
				const safeHref = href && /^(https?:\/\/|mailto:|\/)/i.test(href) ? href : undefined;
				return (
					<a
						{...props}
						href={safeHref}
						target="_blank"
						rel="noopener noreferrer"
						className="font-medium text-primary underline decoration-primary/40 underline-offset-2 hover:decoration-primary"
					>
						{children}
					</a>
				);
			},
			pre({ children }) {
				return <>{children}</>;
			},
			code({ className: codeClassName, children, ...props }: ComponentPropsWithoutRef<"code">) {
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
									onClick={reply.id && onReplyClick ? () => onReplyClick(reply.id) : undefined}
								/>
							</div>
						);
					}
				}
				if (match?.[1] === "wichat-attachment") {
					const att = parseInlineAttachmentPayload(raw);
					if (att) {
						return (
							<AttachmentPreview
								attachment={att}
								onOpenLightbox={(item) => setLightboxAttachment(item)}
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
			},
		}),
		[onReplyClick],
	);

	if (!text.trim()) {
		return null;
	}

	return (
		<div
			className={cn(
				"select-text break-words text-[14.5px] leading-[1.55] text-foreground [&_blockquote]:my-1.5 [&_blockquote]:border-l-2 [&_blockquote]:border-primary/60 [&_blockquote]:pl-3 [&_blockquote]:text-muted-foreground [&_h1]:mt-2 [&_h1]:text-base [&_h1]:font-bold [&_h2]:mt-2 [&_h2]:text-[15px] [&_h2]:font-semibold [&_h3]:mt-1.5 [&_h3]:text-sm [&_h3]:font-semibold [&_li]:my-0.5 [&_ol]:my-1.5 [&_ol]:list-decimal [&_ol]:pl-5 [&_p+p]:mt-2 [&_table]:my-2 [&_table]:w-full [&_table]:border-collapse [&_table]:text-xs [&_td]:border [&_td]:border-border [&_td]:px-2.5 [&_td]:py-1 [&_th]:border [&_th]:border-border [&_th]:bg-muted/60 [&_th]:px-2.5 [&_th]:py-1 [&_th]:text-left [&_th]:font-semibold [&_ul]:my-1.5 [&_ul]:list-disc [&_ul]:pl-5",
				className,
			)}
			data-wichat-selectable="true"
		>
			<ReactMarkdown remarkPlugins={REMARK_PLUGINS} components={markdownComponents}>
				{text}
			</ReactMarkdown>
			{lightboxAttachment && (
				<AttachmentPreviewModal
					attachment={lightboxAttachment}
					onClose={() => setLightboxAttachment(null)}
				/>
			)}
		</div>
	);
});
