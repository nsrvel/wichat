"use client";

import type { MessageAttachment, MessageReplyPreview } from "@wichat/contracts/chat/message";
import Placeholder from "@tiptap/extension-placeholder";
import { EditorContent, useEditor } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import {
	Bold,
	Check,
	Code,
	Code2,
	Italic,
	Link2,
	List,
	ListOrdered,
	Paperclip,
	Quote,
	Send,
	Smile,
	Trash2,
	Underline,
	Upload,
	X,
} from "lucide-react";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { ExpressionPickerPopover } from "@/components/expression-picker/expression-picker-popover";
import { AttachmentBlockNode } from "@/features/chat/components/composer-attachment-node";
import { ReplyQuoteNode } from "@/features/chat/components/composer-reply-node";
import { clearComposerFiles, registerComposerFile } from "@/features/chat/lib/attachment-preview";
import {
	hasInlineReplyBlock,
	markdownToTipTapContent,
	serializeTipTapToMarkdown,
	stripInlineReplyBlocks,
} from "@/features/chat/lib/composer-markdown";
import type { ExpressionSelection } from "@/lib/expressions/expression-types";
import { cn } from "@/lib/utils";

const MAX_MULTILINE_EDITOR_PX = 200;
const FORMAT_TOOLBAR_HEIGHT_PX = 40;
const MAX_FORMAT_EDITOR_PX = MAX_MULTILINE_EDITOR_PX - FORMAT_TOOLBAR_HEIGHT_PX;
const MIN_FORMAT_EDITOR_PX = 76;

function FormatComposeIcon({ className }: { className?: string }) {
	return (
		<svg
			viewBox="0 0 24 24"
			fill="none"
			stroke="currentColor"
			strokeWidth={1.75}
			strokeLinecap="round"
			strokeLinejoin="round"
			className={className}
			aria-hidden
		>
			{/* Capital A on the left */}
			<path d="M3.5 16.5 7.5 5.5l4 11" />
			<path d="M4.8 13h5.4" />
			{/* Diagonal pencil on the bottom-right */}
			<path d="m14.2 19.2 5.6-5.6a1.4 1.4 0 0 1 2 2l-5.6 5.6-2.8.8.8-2.8Z" />
		</svg>
	);
}

type ChatComposerProps = {
	mode?: "send" | "edit";
	initialValue?: string;
	onCancel?: () => void;
	className?: string;
	placeholder?: string;
	onChange?: (value: string) => void;
	onSubmit?: (value: string) => void;
	onUploadFile?: (file: File) => Promise<MessageAttachment>;
	onSelectExpression?: (selection: ExpressionSelection) => void;
	disabled?: boolean;
	disabledReason?: string;
	submitDisabled?: boolean;
	statusNotice?: string;
	replyTo?: MessageReplyPreview | null;
	replyTimeLabel?: string;
	onCancelReply?: () => void;
	onReplyClick?: (messageId: string) => void;
};

export function ChatComposer({
	mode = "send",
	initialValue,
	onCancel,
	className,
	placeholder,
	onChange,
	onSubmit,
	onUploadFile,
	onSelectExpression,
	disabled = false,
	disabledReason,
	submitDisabled,
	statusNotice,
	replyTo,
	replyTimeLabel,
	onCancelReply,
	onReplyClick,
}: ChatComposerProps) {
	const { t } = useTranslation();
	const isEditMode = mode === "edit";
	const initialDoc = useMemo(
		() => (initialValue ? markdownToTipTapContent(initialValue) : undefined),
		[initialValue],
	);
	const editorContainerRef = useRef<HTMLDivElement>(null);
	const wrapCharThresholdRef = useRef<number>(0);
	const [showFormatting, setShowFormatting] = useState(false);
	const [isMultiline, setIsMultiline] = useState(() => Boolean(initialValue?.includes("\n")));
	const [hasContent, setHasContent] = useState(() =>
		Boolean(initialValue && stripInlineReplyBlocks(initialValue).length > 0),
	);
	const [hasReplyNode, setHasReplyNode] = useState(() =>
		Boolean(initialValue && hasInlineReplyBlock(initialValue)),
	);
	const [hasAttachmentNode, setHasAttachmentNode] = useState(() =>
		Boolean(initialValue?.includes("```wichat-attachment\n")),
	);
	const [hasUploadedAttachment, setHasUploadedAttachment] = useState(() =>
		Boolean(initialValue?.includes("```wichat-attachment\n")),
	);
	const [hasPendingOrErroredAttachment, setHasPendingOrErroredAttachment] = useState(false);
	const [isDragOver, setIsDragOver] = useState(false);
	const [attachmentNotice, setAttachmentNotice] = useState("");
	const fileInputRef = useRef<HTMLInputElement>(null);
	const pendingFilesRef = useRef<Map<string, File>>(new Map());
	const onUploadFileRef = useRef(onUploadFile);
	const enqueueFilesRef = useRef<(files: FileList | File[]) => void>(() => {});
	const showFormattingRef = useRef(showFormatting);
	const submitRef = useRef<() => void>(() => {});
	const onCancelRef = useRef(onCancel);
	const onCancelReplyRef = useRef(onCancelReply);
	const onChangeRef = useRef(onChange);
	const lastSyncedReplyIdRef = useRef<string | null>(null);

	useEffect(() => {
		onUploadFileRef.current = onUploadFile;
	}, [onUploadFile]);

	useEffect(() => {
		showFormattingRef.current = showFormatting;
	}, [showFormatting]);

	useEffect(() => {
		onCancelRef.current = onCancel;
	}, [onCancel]);

	useEffect(() => {
		onCancelReplyRef.current = onCancelReply;
	}, [onCancelReply]);

	useEffect(() => {
		onChangeRef.current = onChange;
	}, [onChange]);

	const resolvedPlaceholder =
		disabled && disabledReason ? disabledReason : (placeholder ?? t("chat.composer_placeholder"));
	const isStacked =
		showFormatting || isMultiline || Boolean(replyTo) || hasReplyNode || hasAttachmentNode;
	const canSend =
		(hasContent || hasUploadedAttachment) &&
		!hasPendingOrErroredAttachment &&
		!disabled &&
		submitDisabled !== true;

	const syncMultilineState = useCallback(
		(text: string, blockCount: number, replyPresent: boolean) => {
			if (replyPresent || blockCount > 1 || text.includes("\n")) {
				wrapCharThresholdRef.current = 0;
				setIsMultiline(blockCount > 1 || text.includes("\n"));
				return;
			}
			if (!text) {
				wrapCharThresholdRef.current = 0;
				setIsMultiline(false);
				return;
			}
			const container = editorContainerRef.current;
			const dom = container?.querySelector(".tiptap") as HTMLElement | null;
			if (!dom) {
				return;
			}
			const wraps = dom.scrollHeight > 36;
			setIsMultiline((prev) => {
				if (!prev && wraps) {
					wrapCharThresholdRef.current = text.length;
					return true;
				}
				if (
					prev &&
					!wraps &&
					(wrapCharThresholdRef.current === 0 || text.length <= wrapCharThresholdRef.current - 8)
				) {
					wrapCharThresholdRef.current = 0;
					return false;
				}
				return prev;
			});
		},
		[],
	);

	const editor = useEditor({
		immediatelyRender: false,
		editable: !disabled,
		content: initialDoc,
		autofocus: isEditMode ? "end" : false,
		extensions: [
			StarterKit.configure({
				heading: false,
				link: {
					openOnClick: false,
					autolink: true,
				},
			}),
			Placeholder.configure({
				placeholder: resolvedPlaceholder,
			}),
			ReplyQuoteNode,
			AttachmentBlockNode,
		],
		editorProps: {
			attributes: {
				class:
					"tiptap min-h-8 w-full text-[15px] leading-6 text-foreground focus:outline-none py-1",
				"aria-label": resolvedPlaceholder,
			},
			handlePaste: (_view, event) => {
				const files = event.clipboardData?.files;
				if (files && files.length > 0 && onUploadFileRef.current) {
					event.preventDefault();
					enqueueFilesRef.current(files);
					return true;
				}
				return false;
			},
			handleDrop: (_view, event, _slice, moved) => {
				const files = event.dataTransfer?.files;
				if (!moved && files && files.length > 0 && onUploadFileRef.current) {
					event.preventDefault();
					setIsDragOver(false);
					enqueueFilesRef.current(files);
					return true;
				}
				return false;
			},
			handleKeyDown: (view, event) => {
				if (isEditMode && event.key === "Escape") {
					event.preventDefault();
					onCancelRef.current?.();
					return true;
				}
				const mod = event.metaKey || event.ctrlKey;
				if (mod && event.shiftKey && event.key.toLowerCase() === "b") {
					event.preventDefault();
					editor?.chain().focus().toggleCodeBlock().run();
					return true;
				}
				if (event.key === "Enter" && !event.shiftKey && !showFormattingRef.current) {
					const { $from } = view.state.selection;
					const inSpecialBlock =
						$from.parent.type.name === "codeBlock" ||
						Boolean(
							$from.node(-1)?.type.name === "listItem" || $from.node(-2)?.type.name === "listItem",
						);
					if (!inSpecialBlock) {
						event.preventDefault();
						submitRef.current();
						return true;
					}
				}
				if (event.key === "Enter" && mod && showFormattingRef.current) {
					event.preventDefault();
					submitRef.current();
					return true;
				}
				return false;
			},
		},
		onUpdate: ({ editor: ed }) => {
			const json = ed.getJSON();
			const { markdown } = serializeTipTapToMarkdown(json);
			const blocks = json.content ?? [];
			const replyExists = blocks.some((b) => b.type === "replyQuote");
			if (!replyExists && lastSyncedReplyIdRef.current !== null) {
				lastSyncedReplyIdRef.current = null;
				onCancelReplyRef.current?.();
			}
			const text = ed.getText();
			let anyAttachment = false;
			let uploadedAttachment = false;
			let pendingOrErrored = false;
			ed.state.doc.descendants((node) => {
				if (node.type.name === "attachmentBlock") {
					anyAttachment = true;
					const status = String(node.attrs.status ?? "");
					if (status === "uploaded" && node.attrs.url) {
						uploadedAttachment = true;
					} else {
						pendingOrErrored = true;
					}
				}
			});
			setHasReplyNode(replyExists);
			setHasAttachmentNode(anyAttachment);
			setHasUploadedAttachment(uploadedAttachment);
			setHasPendingOrErroredAttachment(pendingOrErrored);
			setHasContent(stripInlineReplyBlocks(markdown).length > 0);
			onChangeRef.current?.(markdown);
			syncMultilineState(text, blocks.length, replyExists || anyAttachment);
		},
	});

	useEffect(() => {
		if (!editor) {
			return;
		}
		editor.setEditable(!disabled);
	}, [disabled, editor]);

	const updateAttachmentNodeAttrs = useCallback(
		(localId: string, patch: Record<string, unknown>) => {
			if (!editor) {
				return;
			}
			let found: { pos: number; attrs: Record<string, unknown> } | undefined;
			editor.state.doc.descendants((node, pos) => {
				if (node.type.name === "attachmentBlock" && node.attrs.localId === localId) {
					found = { pos, attrs: { ...node.attrs } };
					return false;
				}
				return true;
			});
			if (!found) {
				return;
			}
			const { pos, attrs } = found;
			const tr = editor.state.tr.setNodeMarkup(pos, undefined, {
				...attrs,
				...patch,
			});
			editor.view.dispatch(tr);
		},
		[editor],
	);

	const startUploadForNode = useCallback(
		async (localId: string, file: File) => {
			const uploader = onUploadFileRef.current;
			if (!uploader) {
				return;
			}
			try {
				const uploaded = await uploader(file);
				pendingFilesRef.current.delete(localId);
				updateAttachmentNodeAttrs(localId, {
					attachmentId: uploaded.id,
					url: uploaded.url,
					filename: uploaded.filename,
					contentType: uploaded.content_type,
					sizeBytes: uploaded.size_bytes,
					status: "uploaded",
					errorMessage: "",
				});
			} catch (err) {
				updateAttachmentNodeAttrs(localId, {
					status: "error",
					errorMessage: err instanceof Error ? err.message : t("chat.attachment_upload_failed"),
				});
			}
		},
		[t, updateAttachmentNodeAttrs],
	);

	const enqueueFiles = useCallback(
		(fileList: FileList | File[]) => {
			if (!editor || disabled || !onUploadFileRef.current) {
				return;
			}
			setAttachmentNotice("");
			const files = Array.from(fileList);
			for (const file of files) {
				if (file.size > 25 * 1024 * 1024) {
					setAttachmentNotice(t("chat.attachment_file_too_large"));
					continue;
				}
				const localId = crypto.randomUUID();
				const previewUrl = file.type.startsWith("image/") ? URL.createObjectURL(file) : "";
				pendingFilesRef.current.set(localId, file);
				registerComposerFile(localId, file);
				editor
					.chain()
					.focus()
					.insertContent([
						{
							type: "attachmentBlock",
							attrs: {
								localId,
								attachmentId: "",
								url: "",
								previewUrl,
								filename: file.name || "file",
								contentType: file.type || "application/octet-stream",
								sizeBytes: file.size,
								status: "uploading",
								errorMessage: "",
							},
						},
						{ type: "paragraph" },
					])
					.focus()
					.run();
				void startUploadForNode(localId, file);
			}
		},
		[disabled, editor, startUploadForNode, t],
	);

	const retryAttachment = useCallback(
		(localId: string) => {
			const file = pendingFilesRef.current.get(localId);
			if (!file) {
				return;
			}
			updateAttachmentNodeAttrs(localId, {
				status: "uploading",
				errorMessage: "",
			});
			void startUploadForNode(localId, file);
		},
		[startUploadForNode, updateAttachmentNodeAttrs],
	);

	useEffect(() => {
		enqueueFilesRef.current = enqueueFiles;
	}, [enqueueFiles]);

	useEffect(() => {
		const handler = (event: Event) => {
			const detail = (event as CustomEvent<{ localId?: string }>).detail;
			if (detail?.localId) {
				retryAttachment(detail.localId);
			}
		};
		window.addEventListener("wichat:retry-attachment", handler);
		return () => window.removeEventListener("wichat:retry-attachment", handler);
	}, [retryAttachment]);

	const submit = useCallback(() => {
		if (!editor || disabled || submitDisabled === true || hasPendingOrErroredAttachment) {
			return;
		}
		const { markdown } = serializeTipTapToMarkdown(editor.getJSON());
		if (!stripInlineReplyBlocks(markdown) && !hasUploadedAttachment) {
			return;
		}
		onSubmit?.(markdown);
		if (isEditMode) {
			return;
		}
		lastSyncedReplyIdRef.current = null;
		pendingFilesRef.current.clear();
		clearComposerFiles();
		setAttachmentNotice("");
		editor.commands.clearContent(true);
		setHasContent(false);
		setHasReplyNode(false);
		setHasAttachmentNode(false);
		setHasUploadedAttachment(false);
		setHasPendingOrErroredAttachment(false);
		setIsMultiline(false);
		wrapCharThresholdRef.current = 0;
	}, [
		disabled,
		editor,
		hasPendingOrErroredAttachment,
		hasUploadedAttachment,
		isEditMode,
		onSubmit,
		submitDisabled,
	]);

	useEffect(() => {
		submitRef.current = submit;
	}, [submit]);

	// Sync external replyTo prop into inline TipTap replyQuote node
	useEffect(() => {
		if (!editor) {
			return;
		}
		const nextReplyId = replyTo?.id ?? null;
		if (nextReplyId === lastSyncedReplyIdRef.current) {
			return;
		}

		// Remove any existing replyQuote nodes first
		const tr = editor.state.tr;
		const positions: Array<{ from: number; to: number }> = [];
		editor.state.doc.descendants((node, pos) => {
			if (node.type.name === "replyQuote") {
				positions.push({ from: pos, to: pos + node.nodeSize });
			}
		});
		for (let i = positions.length - 1; i >= 0; i -= 1) {
			const range = positions[i];
			if (range) {
				tr.delete(range.from, range.to);
			}
		}
		if (positions.length > 0) {
			editor.view.dispatch(tr);
		}

		lastSyncedReplyIdRef.current = nextReplyId;

		if (!replyTo) {
			return;
		}

		const replyNodeContent = {
			type: "replyQuote",
			attrs: {
				messageId: replyTo.id,
				authorName: replyTo.author_display_name,
				timeLabel: replyTimeLabel ?? "",
				bodyPreview: replyTo.body_preview,
			},
		};

		const isEmptyDoc = editor.isEmpty;
		if (isEmptyDoc) {
			editor
				.chain()
				.focus("end")
				.insertContent([replyNodeContent, { type: "paragraph" }])
				.focus("end")
				.run();
		} else {
			editor
				.chain()
				.focus()
				.insertContent([replyNodeContent, { type: "paragraph" }])
				.focus()
				.run();
		}
	}, [editor, replyTimeLabel, replyTo]);

	const insertLink = useCallback(() => {
		if (!editor) {
			return;
		}
		const previousUrl = editor.getAttributes("link").href as string | undefined;
		const url = window.prompt(t("chat.link_prompt"), previousUrl || "https://");
		if (url === null) {
			return;
		}
		if (url === "") {
			editor.chain().focus().extendMarkRange("link").unsetLink().run();
			return;
		}
		if (editor.state.selection.empty) {
			editor.chain().focus().insertContent(`<a href="${url}">${url}</a>`).run();
			return;
		}
		editor.chain().focus().extendMarkRange("link").setLink({ href: url }).run();
	}, [editor, t]);

	const handleSelectExpression = useCallback(
		(selection: ExpressionSelection) => {
			if (selection.kind === "emoji" && editor) {
				editor.chain().focus().insertContent(selection.emoji).run();
			}
			onSelectExpression?.(selection);
		},
		[editor, onSelectExpression],
	);

	const toolbarBtnClass = (active: boolean) =>
		cn(
			"flex size-7 items-center justify-center rounded-md transition",
			active
				? "bg-primary/15 text-primary"
				: "text-muted-foreground hover:bg-sidebar-accent hover:text-foreground",
		);

	return (
		<div
			className={cn(
				isEditMode
					? "w-full"
					: "w-full shrink-0 bg-chat-surface pb-2.5 pr-[var(--scrollbar-size)] pt-2 lg:pb-3 lg:pt-2.5",
				className,
			)}
		>
			<div
				className={cn(
					isEditMode ? "w-full" : "mx-auto w-full max-w-4xl px-4 sm:px-5 lg:max-w-5xl lg:px-6",
				)}
			>
				<div className={cn(!isEditMode && "pl-[42px]")}>
					<div
						className={cn(
							"relative flex flex-col overflow-hidden rounded-xl border bg-chat-bubble transition-colors",
							isDragOver ? "border-primary ring-2 ring-primary/25" : "border-chat-bubble-border",
						)}
						onDragOver={(e) => {
							if (
								!disabled &&
								onUploadFileRef.current &&
								e.dataTransfer?.types?.includes("Files")
							) {
								e.preventDefault();
								setIsDragOver(true);
							}
						}}
						onDragLeave={(e) => {
							if (!e.currentTarget.contains(e.relatedTarget as Node | null)) {
								setIsDragOver(false);
							}
						}}
						onDrop={(e) => {
							if (!disabled && onUploadFileRef.current && e.dataTransfer?.files?.length) {
								e.preventDefault();
								setIsDragOver(false);
								enqueueFiles(e.dataTransfer.files);
							}
						}}
					>
						{isDragOver && (
							<div className="pointer-events-none absolute inset-0 z-20 flex items-center justify-center gap-2 bg-chat-bubble/90 text-xs font-medium text-primary backdrop-blur-[1px]">
								<Upload className="size-4" />
								{t("chat.drop_files_here")}
							</div>
						)}
						<div
							className={cn(
								"grid transition-[grid-template-rows,opacity] duration-200 ease-out",
								showFormatting && !disabled
									? "grid-rows-[1fr] opacity-100"
									: "pointer-events-none grid-rows-[0fr] opacity-0",
							)}
							aria-hidden={!showFormatting || disabled}
						>
							<div className="overflow-hidden">
								<div className="flex flex-wrap items-center justify-between gap-1 border-b border-border/60 px-2.5 py-1.5">
									<div className="flex flex-wrap items-center gap-0.5">
										<button
											type="button"
											onClick={() => editor?.chain().focus().toggleBold().run()}
											className={toolbarBtnClass(Boolean(editor?.isActive("bold")))}
											title={`${t("chat.toolbar_bold")} (⌘B)`}
											aria-label={t("chat.toolbar_bold")}
										>
											<Bold className="size-3.5" />
										</button>
										<button
											type="button"
											onClick={() => editor?.chain().focus().toggleItalic().run()}
											className={toolbarBtnClass(Boolean(editor?.isActive("italic")))}
											title={`${t("chat.toolbar_italic")} (⌘I)`}
											aria-label={t("chat.toolbar_italic")}
										>
											<Italic className="size-3.5" />
										</button>
										<button
											type="button"
											onClick={() => editor?.chain().focus().toggleUnderline().run()}
											className={toolbarBtnClass(Boolean(editor?.isActive("underline")))}
											title={`${t("chat.toolbar_underline")} (⌘U)`}
											aria-label={t("chat.toolbar_underline")}
										>
											<Underline className="size-3.5" />
										</button>
										<span className="mx-1 h-4 w-px bg-border/70" aria-hidden />
										<button
											type="button"
											onClick={() => editor?.chain().focus().toggleBulletList().run()}
											className={toolbarBtnClass(Boolean(editor?.isActive("bulletList")))}
											title={t("chat.toolbar_bullet_list")}
											aria-label={t("chat.toolbar_bullet_list")}
										>
											<List className="size-3.5" />
										</button>
										<button
											type="button"
											onClick={() => editor?.chain().focus().toggleOrderedList().run()}
											className={toolbarBtnClass(Boolean(editor?.isActive("orderedList")))}
											title={t("chat.toolbar_ordered_list")}
											aria-label={t("chat.toolbar_ordered_list")}
										>
											<ListOrdered className="size-3.5" />
										</button>
										<span className="mx-1 h-4 w-px bg-border/70" aria-hidden />
										<button
											type="button"
											onClick={() => editor?.chain().focus().toggleBlockquote().run()}
											className={toolbarBtnClass(Boolean(editor?.isActive("blockquote")))}
											title={t("chat.toolbar_quote")}
											aria-label={t("chat.toolbar_quote")}
										>
											<Quote className="size-3.5" />
										</button>
										<button
											type="button"
											onClick={insertLink}
											className={toolbarBtnClass(Boolean(editor?.isActive("link")))}
											title={t("chat.toolbar_link")}
											aria-label={t("chat.toolbar_link")}
										>
											<Link2 className="size-3.5" />
										</button>
										<button
											type="button"
											onClick={() => editor?.chain().focus().toggleCode().run()}
											className={toolbarBtnClass(Boolean(editor?.isActive("code")))}
											title={t("chat.toolbar_inline_code")}
											aria-label={t("chat.toolbar_inline_code")}
										>
											<Code className="size-3.5" />
										</button>
										<button
											type="button"
											onClick={() => editor?.chain().focus().toggleCodeBlock().run()}
											className={toolbarBtnClass(Boolean(editor?.isActive("codeBlock")))}
											title={`${t("chat.toolbar_code_block")} (⇧⌘B)`}
											aria-label={t("chat.toolbar_code_block")}
										>
											<Code2 className="size-3.5" />
										</button>
									</div>

									<button
										type="button"
										onClick={() => {
											setShowFormatting(false);
											lastSyncedReplyIdRef.current = null;
											editor?.commands.clearContent(true);
											setHasContent(false);
											setHasReplyNode(false);
											setIsMultiline(false);
											onCancelReplyRef.current?.();
										}}
										className="flex size-7 items-center justify-center rounded-md text-muted-foreground transition hover:bg-destructive/15 hover:text-destructive"
										title={t("chat.toolbar_clear_draft")}
										aria-label={t("chat.toolbar_clear_draft")}
									>
										<Trash2 className="size-3.5" />
									</button>
								</div>
							</div>
						</div>

						<div
							ref={editorContainerRef}
							onClick={(event) => {
								const target = event.target as HTMLElement | null;
								const jumpEl = target?.closest("[data-reply-jump-id]") as HTMLElement | null;
								const jumpId = jumpEl?.dataset.replyJumpId;
								if (jumpId && !target?.closest("button")) {
									onReplyClick?.(jumpId);
								}
							}}
							className={cn(
								"relative px-3 pt-1.5 transition-[padding] duration-150 ease-out",
								isStacked ? "pb-10" : isEditMode ? "pb-1.5 pr-[196px]" : "pb-1.5 pr-[164px]",
							)}
						>
							<div
								style={{
									minHeight: showFormatting ? `${MIN_FORMAT_EDITOR_PX}px` : undefined,
									maxHeight: showFormatting
										? `${MAX_FORMAT_EDITOR_PX}px`
										: `${MAX_MULTILINE_EDITOR_PX}px`,
								}}
								className="thin-scroll overflow-y-auto transition-[min-height,max-height] duration-150 ease-out"
							>
								<EditorContent editor={editor} />
							</div>

							<div className="absolute bottom-1.5 right-3 flex items-center justify-end">
								<button
									type="button"
									disabled={disabled}
									onClick={() => setShowFormatting((prev) => !prev)}
									className={cn(
										"flex size-8 shrink-0 items-center justify-center rounded-md transition-colors disabled:opacity-50",
										showFormatting
											? "bg-sidebar-accent text-sidebar-foreground"
											: "text-muted-foreground hover:bg-sidebar-accent hover:text-sidebar-foreground",
									)}
									aria-label={t("chat.formatting")}
									aria-pressed={showFormatting}
								>
									<FormatComposeIcon className="size-[18px]" />
								</button>

								{!disabled ? (
									<ExpressionPickerPopover
										trigger={
											<button
												type="button"
												className="flex size-8 shrink-0 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-sidebar-accent hover:text-sidebar-foreground data-[popup-open]:bg-sidebar-accent data-[popup-open]:text-sidebar-foreground"
												aria-label={t("chat.emoji")}
											>
												<Smile className="size-[18px]" strokeWidth={1.75} />
											</button>
										}
										onSelect={handleSelectExpression}
									/>
								) : (
									<button
										type="button"
										disabled
										className="flex size-8 shrink-0 items-center justify-center rounded-md text-muted-foreground opacity-50"
										aria-label={t("chat.emoji")}
									>
										<Smile className="size-[18px]" strokeWidth={1.75} />
									</button>
								)}

								<input
									ref={fileInputRef}
									type="file"
									multiple
									className="hidden"
									onChange={(e) => {
										if (e.target.files && e.target.files.length > 0) {
											enqueueFiles(e.target.files);
										}
										e.target.value = "";
									}}
								/>
								<button
									type="button"
									disabled={disabled || !onUploadFile}
									onClick={() => fileInputRef.current?.click()}
									className="flex size-8 shrink-0 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-sidebar-accent hover:text-sidebar-foreground disabled:opacity-50"
									aria-label={t("chat.attach")}
									title={t("chat.attach")}
								>
									<Paperclip className="size-[18px]" strokeWidth={1.75} />
								</button>

								<span className="mx-1.5 h-6 w-px shrink-0 bg-chat-bubble-border" aria-hidden />
								{isEditMode ? (
									<>
										<button
											type="button"
											onClick={onCancel}
											className="flex size-8 shrink-0 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-sidebar-accent hover:text-sidebar-foreground"
											aria-label={t("chat.cancel_edit")}
											title={t("chat.cancel_edit")}
										>
											<X className="size-[18px]" strokeWidth={1.75} />
										</button>
										<button
											type="button"
											disabled={!canSend}
											onClick={submit}
											className="flex size-8 shrink-0 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-sidebar-accent hover:text-sidebar-foreground disabled:opacity-40"
											aria-label={t("chat.save_edit")}
											title={t("chat.save_edit")}
										>
											<Check className="size-[18px]" strokeWidth={1.75} />
										</button>
									</>
								) : (
									<button
										type="button"
										disabled={!canSend}
										onClick={submit}
										className="flex size-8 shrink-0 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-sidebar-accent hover:text-sidebar-foreground"
										aria-label={t("chat.send")}
										title={submitDisabled ? t("chat.send_soon") : undefined}
									>
										<Send className="size-[18px]" strokeWidth={1.75} />
									</button>
								)}
							</div>
						</div>
					</div>

					<div className="mt-1 flex items-center justify-between gap-2 px-1 text-[11px] text-muted-foreground">
						{attachmentNotice ? (
							<p className="font-medium text-destructive">{attachmentNotice}</p>
						) : hasPendingOrErroredAttachment ? (
							<p className="font-medium">{t("chat.attachment_wait_upload")}</p>
						) : statusNotice ? (
							<p className="font-medium">{statusNotice}</p>
						) : (
							<span />
						)}
						{!disabled ? (
							<p className="select-none">
								<span className="font-medium text-primary">
									{showFormatting ? "Enter" : "Shift+Enter"}
								</span>{" "}
								{t("chat.hint_starts_new_line")}
							</p>
						) : null}
					</div>
				</div>
			</div>
		</div>
	);
}
