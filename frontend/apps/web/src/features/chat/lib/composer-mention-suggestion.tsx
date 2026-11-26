"use client";

import type { QueryClient } from "@tanstack/react-query";
import { Editor, Extension, ReactRenderer } from "@tiptap/react";
import Suggestion, { type SuggestionKeyDownProps, type SuggestionProps } from "@tiptap/suggestion";
import { PluginKey } from "@tiptap/pm/state";
import { forwardRef, useEffect, useImperativeHandle, useState } from "react";
import { UserAvatar } from "@/components/ui/user-avatar";
import { userInitials } from "@/lib/user-initials";
import { positionWithinViewport } from "@/lib/position-within-viewport";
import { cn } from "@/lib/utils";
import { groupMembersQueryOpts } from "@/services/invite/invite.query";

export type MentionSuggestItem = {
	id: string;
	label: string;
	username: string;
	avatarUrl: string;
};

type MentionListRef = {
	onKeyDown: (props: SuggestionKeyDownProps) => boolean;
};

type MentionListProps = {
	items: MentionSuggestItem[];
	command: (item: MentionSuggestItem) => void;
};

const MentionList = forwardRef<MentionListRef, MentionListProps>(function MentionList(
	{ items, command },
	ref,
) {
	const [selectedIndex, setSelectedIndex] = useState(0);

	useEffect(() => {
		setSelectedIndex(0);
	}, [items]);

	useImperativeHandle(ref, () => ({
		onKeyDown: ({ event }) => {
			if (event.key === "ArrowUp") {
				setSelectedIndex((i) => (i + items.length - 1) % items.length);
				return true;
			}
			if (event.key === "ArrowDown") {
				setSelectedIndex((i) => (i + 1) % items.length);
				return true;
			}
			if (event.key === "Enter") {
				const item = items[selectedIndex];
				if (item) {
					command(item);
				}
				return true;
			}
			return false;
		},
	}));

	if (items.length === 0) {
		return null;
	}

	return (
		<div
			className="z-50 w-[min(20rem,calc(100vw-1rem))] rounded-lg border border-border bg-popover p-1 shadow-md"
			role="listbox"
		>
			{items.map((item, index) => (
				<button
					key={item.id}
					type="button"
					role="option"
					aria-selected={index === selectedIndex}
					onClick={() => command(item)}
					className={cn(
						"flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-sm hover:bg-muted",
						index === selectedIndex && "bg-muted",
					)}
				>
					<UserAvatar
						src={item.avatarUrl}
						alt={item.label}
						initials={userInitials(item.label || item.username)}
						size="sm"
					/>
					<span className="min-w-0 truncate font-medium">{item.label}</span>
					<span className="ml-auto truncate text-xs text-muted-foreground">@{item.username}</span>
				</button>
			))}
		</div>
	);
});

function filterMembers(
	members: { user_id: string; display_name: string; username: string; avatar_url?: string | null }[],
	query: string,
): MentionSuggestItem[] {
	const needle = query.trim().toLowerCase();
	const rows = !needle
		? members.slice(0, 12)
		: members
				.filter((m) => {
					const name = m.display_name?.toLowerCase() ?? "";
					const username = m.username?.toLowerCase() ?? "";
					return name.includes(needle) || username.includes(needle);
				})
				.slice(0, 12);
	return rows.map((m) => ({
		id: m.user_id,
		label: m.display_name || m.username,
		username: m.username,
		avatarUrl: m.avatar_url ?? "",
	}));
}

async function loadMentionItems(
	queryClient: QueryClient,
	groupId: string,
	query: string,
): Promise<MentionSuggestItem[]> {
	const data = await queryClient.fetchQuery(groupMembersQueryOpts(groupId, ""));
	return filterMembers(data?.members ?? [], query);
}

export type ComposerMentionSuggestionOptions = {
	groupId: string;
	queryClient: QueryClient;
};

export const ComposerMentionSuggestion = Extension.create<ComposerMentionSuggestionOptions>({
	name: "composerMentionSuggestion",

	addOptions() {
		return {
			groupId: "",
			queryClient: null as unknown as QueryClient,
		};
	},

	addProseMirrorPlugins() {
		const { groupId, queryClient } = this.options;
		if (!groupId || !queryClient) {
			return [];
		}

		return [
			Suggestion<MentionSuggestItem, MentionSuggestItem>({
				editor: this.editor,
				char: "@",
				pluginKey: new PluginKey("composerMentionSuggestion"),
				allow: ({ state, range }) => {
					const $from = state.doc.resolve(range.from);
					return $from.parent.type.name !== "codeBlock";
				},
				items: ({ query }) => loadMentionItems(queryClient, groupId, query),
				command: ({ editor, range, props }) => {
					const id = props.id?.trim() ?? "";
					const label = props.label?.trim() || props.username?.trim() || id;
					editor
						.chain()
						.focus()
						.insertContentAt(range, [
							{
								type: "userMention",
								attrs: { id, label },
							},
							{ type: "text", text: " " },
						])
						.run();
				},
				render: () => {
					let renderer: ReactRenderer<MentionListRef> | null = null;
					let element: HTMLDivElement | null = null;

					const estimateListHeight = (itemCount: number) =>
						Math.min(itemCount, 8) * 44 + 12;

					const position = (_editor: Editor, props: SuggestionProps<MentionSuggestItem>) => {
						const rect = props.clientRect?.();
						if (!element || !rect) {
							return;
						}
						const width = Math.max(element.offsetWidth, 288);
						const measured = element.offsetHeight;
						const height =
							measured > 0 ? measured : estimateListHeight(props.items.length);
						const vh = window.innerHeight;
						// Composer sits low on the screen — prefer opening above the caret when in the lower band.
						const preferTop = rect.bottom > vh * 0.52;
						const { top, left } = positionWithinViewport({
							anchor: rect,
							width,
							height,
							align: "start",
							side: preferTop ? "top" : "bottom",
							gap: 6,
						});
						element.style.width = `${width}px`;
						element.style.left = `${left}px`;
						element.style.top = `${top}px`;
					};

					const schedulePosition = (editor: Editor, props: SuggestionProps<MentionSuggestItem>) => {
						requestAnimationFrame(() => {
							position(editor, props);
							requestAnimationFrame(() => position(editor, props));
						});
					};

					return {
						onStart: (props) => {
							element = document.createElement("div");
							element.style.position = "fixed";
							element.style.zIndex = "50";
							element.style.left = "0";
							element.style.top = "0";
							document.body.appendChild(element);

							renderer = new ReactRenderer(MentionList, {
								props: {
									items: props.items,
									command: props.command,
								},
								editor: props.editor,
							});
							element.appendChild(renderer.element);
							schedulePosition(props.editor, props);
						},
						onUpdate: (props) => {
							renderer?.updateProps({
								items: props.items,
								command: props.command,
							});
							schedulePosition(props.editor, props);
						},
						onKeyDown: (props) => {
							if (props.event.key === "Escape") {
								return true;
							}
							return renderer?.ref?.onKeyDown(props) ?? false;
						},
						onExit: () => {
							renderer?.destroy();
							renderer = null;
							element?.remove();
							element = null;
						},
					};
				},
			}),
		];
	},
});

export function composerMentionSuggestionExtension(
	groupId: string,
	queryClient: QueryClient,
) {
	return ComposerMentionSuggestion.configure({ groupId, queryClient });
}