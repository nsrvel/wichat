"use client";

import {
	mergeAttributes,
	Node,
	NodeViewWrapper,
	ReactNodeViewRenderer,
	type ReactNodeViewProps,
} from "@tiptap/react";
import { ReplyPreviewCard } from "@/features/chat/components/reply-preview-card";

function ReplyQuoteView({ node, deleteNode }: ReactNodeViewProps) {
	const messageId = String(node.attrs.messageId ?? "");
	const authorName = String(node.attrs.authorName ?? "");
	const timeLabel = String(node.attrs.timeLabel ?? "");
	const bodyPreview = String(node.attrs.bodyPreview ?? "");

	return (
		<NodeViewWrapper className="my-1.5 select-none" contentEditable={false}>
			<ReplyPreviewCard
				messageId={messageId}
				authorName={authorName}
				timeLabel={timeLabel}
				bodyPreview={bodyPreview}
				onClick={() => {
					// Click jump is handled by ChatComposer's container listener via data-reply-jump-id
				}}
				onRemove={() => deleteNode()}
			/>
		</NodeViewWrapper>
	);
}

export const ReplyQuoteNode = Node.create({
	name: "replyQuote",
	group: "block",
	atom: true,
	selectable: true,
	draggable: false,

	addAttributes() {
		return {
			messageId: { default: "" },
			authorName: { default: "" },
			timeLabel: { default: "" },
			bodyPreview: { default: "" },
		};
	},

	parseHTML() {
		return [{ tag: "div[data-reply-quote]" }];
	},

	renderHTML({ HTMLAttributes }) {
		return ["div", mergeAttributes(HTMLAttributes, { "data-reply-quote": "" })];
	},

	addNodeView() {
		return ReactNodeViewRenderer(ReplyQuoteView);
	},
});
