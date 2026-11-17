"use client";

import {
	mergeAttributes,
	Node,
	NodeViewWrapper,
	ReactNodeViewRenderer,
	type ReactNodeViewProps,
} from "@tiptap/react";
import { AlertCircle, Loader2, RotateCw, X } from "lucide-react";
import { useTranslation } from "react-i18next";
import { FileAttachmentCard } from "@/features/chat/components/file-attachment-card";
import { getComposerFile, unregisterComposerFile } from "@/features/chat/lib/attachment-preview";
import { resolvePublicObjectUrl } from "@/lib/public-object-url";

export function isImageContentType(contentType: string): boolean {
	return contentType.toLowerCase().startsWith("image/");
}

function AttachmentBlockView({ node, deleteNode }: ReactNodeViewProps) {
	const { t } = useTranslation();
	const localId = String(node.attrs.localId ?? "");
	const url = String(node.attrs.url ?? "");
	const previewUrl = String(node.attrs.previewUrl ?? "");
	const filename = String(node.attrs.filename ?? "file");
	const contentType = String(node.attrs.contentType ?? "application/octet-stream");
	const sizeBytes = Number(node.attrs.sizeBytes ?? 0);
	const status = String(node.attrs.status ?? "uploaded");
	const errorMessage = String(node.attrs.errorMessage ?? "");

	const isImage = isImageContentType(contentType);
	const displayUrl = previewUrl || (url ? resolvePublicObjectUrl(url) : "");
	const dispatchRetry = () => {
		window.dispatchEvent(new CustomEvent("wichat:retry-attachment", { detail: { localId } }));
	};

	const handleRemove = () => {
		if (previewUrl.startsWith("blob:")) {
			URL.revokeObjectURL(previewUrl);
		}
		if (localId) {
			unregisterComposerFile(localId);
		}
		deleteNode();
	};

	if (isImage && displayUrl) {
		return (
			<NodeViewWrapper className="my-2 inline-block select-none" contentEditable={false}>
				<div className="group/att relative inline-block overflow-hidden rounded-xl border border-border/70 bg-muted/30">
					{/* eslint-disable-next-line @next/next/no-img-element */}
					<img
						src={displayUrl}
						alt={filename}
						className="block max-h-[140px] max-w-[200px] object-cover"
					/>

					{status === "uploading" && (
						<div className="absolute inset-0 flex flex-col items-center justify-center gap-1.5 bg-black/55 px-3 text-white backdrop-blur-[1px]">
							<Loader2 className="size-5 animate-spin" />
							<span className="text-[11px] font-medium">{t("chat.attachment_uploading")}</span>
						</div>
					)}

					{status === "error" && (
						<div className="absolute inset-0 flex flex-col items-center justify-center gap-1.5 bg-destructive/85 px-3 text-center text-white">
							<AlertCircle className="size-5" />
							<span className="line-clamp-2 text-[11px] font-medium">
								{errorMessage || t("chat.attachment_upload_failed")}
							</span>
							<button
								type="button"
								onClick={dispatchRetry}
								className="inline-flex items-center gap-1 rounded-md bg-black/40 px-2 py-1 text-[11px] font-medium text-white transition hover:bg-black/60"
							>
								<RotateCw className="size-3" />
								{t("chat.attachment_retry")}
							</button>
						</div>
					)}

					<button
						type="button"
						onClick={handleRemove}
						title={t("chat.attachment_remove")}
						aria-label={t("chat.attachment_remove")}
						className="absolute right-1.5 top-1.5 inline-flex size-6 items-center justify-center rounded-md bg-black/70 text-white transition hover:bg-black/90"
					>
						<X className="size-3.5" />
					</button>
				</div>
			</NodeViewWrapper>
		);
	}

	const localFile = localId ? getComposerFile(localId) : undefined;
	const cardStatus: "uploading" | "uploaded" | "error" =
		status === "uploading" || status === "error" ? status : "uploaded";

	return (
		<NodeViewWrapper className="my-2 block select-none" contentEditable={false}>
			<FileAttachmentCard
				variant="composer"
				id={localId || String(node.attrs.attachmentId ?? filename)}
				filename={filename}
				contentType={contentType}
				sizeBytes={sizeBytes}
				url={url}
				localFile={localFile}
				status={cardStatus}
				errorMessage={errorMessage}
				onRemove={handleRemove}
				onRetry={dispatchRetry}
			/>
		</NodeViewWrapper>
	);
}

export const AttachmentBlockNode = Node.create({
	name: "attachmentBlock",
	group: "block",
	atom: true,
	selectable: true,
	draggable: false,

	addAttributes() {
		return {
			localId: { default: "" },
			attachmentId: { default: "" },
			url: { default: "" },
			previewUrl: { default: "" },
			filename: { default: "file" },
			contentType: { default: "application/octet-stream" },
			sizeBytes: { default: 0 },
			status: { default: "uploaded" },
			errorMessage: { default: "" },
		};
	},

	parseHTML() {
		return [{ tag: "div[data-attachment-block]" }];
	},

	renderHTML({ HTMLAttributes }) {
		return ["div", mergeAttributes(HTMLAttributes, { "data-attachment-block": "" })];
	},

	addNodeView() {
		return ReactNodeViewRenderer(AttachmentBlockView);
	},
});
