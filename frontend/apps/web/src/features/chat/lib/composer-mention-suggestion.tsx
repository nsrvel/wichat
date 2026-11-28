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
import {
	filterMentionSuggestMembers,
	type MentionSuggestMemberItem,
} from "@/features/chat/lib/mention-suggest-members";
import { mentionPickerRoles } from "@/features/chat/lib/mentionable-roles";
import { groupMembersQueryOpts } from "@/services/invite/invite.query";
import { groupRolesQueryOpts } from "@/services/group-role/group-role.query";

export type MentionSuggestRoleItem = {
	kind: "role";
	id: string;
	label: string;
	color: string;
};

export type MentionSuggestItem = MentionSuggestMemberItem | MentionSuggestRoleItem;

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

	const members = items.filter((i): i is MentionSuggestMemberItem => i.kind === "user");
	const roles = items.filter((i): i is MentionSuggestRoleItem => i.kind === "role");

	const renderRow = (item: MentionSuggestItem, index: number) => (
		<button
			key={`${item.kind}-${item.id}`}
			type="button"
			role="option"
			aria-selected={index === selectedIndex}
			data-testid={
				item.kind === "user" ? `mention-suggest-user-${item.id}` : `mention-suggest-role-${item.id}`
			}
			onClick={() => command(item)}
			className={cn(
				"flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-sm hover:bg-muted",
				index === selectedIndex && "bg-muted",
			)}
		>
			{item.kind === "user" ? (
				<>
					<UserAvatar
						src={item.avatarUrl}
						alt={item.label}
						initials={userInitials(item.label || item.username)}
						size="sm"
					/>
					<span className="min-w-0 truncate font-medium">{item.label}</span>
					<span className="ml-auto truncate text-xs text-muted-foreground">@{item.username}</span>
				</>
			) : (
				<>
					<span
						className="min-w-0 truncate font-medium"
						style={
							item.color && item.color.toLowerCase() !== "#99aab5"
								? { color: item.color }
								: undefined
						}
					>
						@{item.label.replace(/^@+/, "")}
					</span>
				</>
			)}
		</button>
	);

	let rowIndex = 0;
	return (
		<div
			className="z-50 max-h-72 w-full overflow-y-auto rounded-lg border border-border bg-popover p-1 shadow-md"
			role="listbox"
		>
			{members.length > 0 ? (
				<div className="px-2 py-1 text-xs font-semibold text-muted-foreground">Members</div>
			) : null}
			{members.map((item) => renderRow(item, rowIndex++))}
			{members.length > 0 && roles.length > 0 ? (
				<div className="my-1 border-t border-border/60" role="separator" />
			) : null}
			{roles.map((item) => renderRow(item, rowIndex++))}
		</div>
	);
});

async function loadMentionItems(
	queryClient: QueryClient,
	groupId: string,
	query: string,
	selfUserId: string | undefined,
	callerPermissions: string[] | undefined,
): Promise<MentionSuggestItem[]> {
	const [membersData, rolesData] = await Promise.all([
		queryClient.fetchQuery(groupMembersQueryOpts(groupId, "")),
		queryClient.fetchQuery(groupRolesQueryOpts(groupId)),
	]);
	const members = filterMentionSuggestMembers(membersData?.members ?? [], query, selfUserId);
	const roles = mentionPickerRoles(rolesData?.roles ?? [], callerPermissions, query).map((r) => ({
		kind: "role" as const,
		id: r.id,
		label: r.name.replace(/^@+/, ""),
		color: r.color,
	}));
	return [...members, ...roles];
}

export type ComposerMentionSuggestionOptions = {
	groupId: string;
	queryClient: QueryClient;
	selfUserId?: string;
	callerPermissions?: string[];
};

export const ComposerMentionSuggestion = Extension.create<ComposerMentionSuggestionOptions>({
	name: "composerMentionSuggestion",

	addOptions() {
		return {
			groupId: "",
			queryClient: null as unknown as QueryClient,
			selfUserId: undefined as string | undefined,
			callerPermissions: [] as string[] | undefined,
		};
	},

	addProseMirrorPlugins() {
		const { groupId, queryClient, selfUserId, callerPermissions } = this.options;
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
				items: ({ query }) =>
					loadMentionItems(queryClient, groupId, query, selfUserId, callerPermissions),
				command: ({ editor, range, props }) => {
					if (props.kind === "user") {
						const id = props.id?.trim() ?? "";
						if (selfUserId?.trim() && id === selfUserId.trim()) {
							return;
						}
						const label = (props.label?.trim() || props.username?.trim() || id).replace(/^@+/, "");
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
						return;
					}
					const id = props.id?.trim() ?? "";
					const label = (props.label?.trim() || id).replace(/^@+/, "");
					const color = props.color?.trim() || "#99aab5";
					editor
						.chain()
						.focus()
						.insertContentAt(range, [
							{
								type: "roleMention",
								attrs: { id, label, color },
							},
							{ type: "text", text: " " },
						])
						.run();
				},
				render: () => {
					let renderer: ReactRenderer<MentionListRef> | null = null;
					let element: HTMLDivElement | null = null;

					const estimateListHeight = (itemCount: number) => Math.min(itemCount, 8) * 44 + 12;

					const position = (editor: Editor, props: SuggestionProps<MentionSuggestItem>) => {
						const rect = props.clientRect?.();
						if (!element || !rect) {
							return;
						}
						const composerBox = editor.view.dom.closest("[data-composer-box]");
						const composerRect = composerBox?.getBoundingClientRect();
						const anchorRect = composerRect ?? rect;
						const width = composerRect ? composerRect.width : Math.max(element.offsetWidth, 288);
						const measured = element.offsetHeight;
						const height = measured > 0 ? measured : estimateListHeight(props.items.length);
						const vh = window.innerHeight;
						const preferTop = anchorRect.bottom > vh * 0.52;
						const { top, left } = positionWithinViewport({
							anchor: anchorRect,
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
	selfUserId?: string,
	callerPermissions?: string[],
) {
	return ComposerMentionSuggestion.configure({
		groupId,
		queryClient,
		selfUserId,
		callerPermissions,
	});
}
