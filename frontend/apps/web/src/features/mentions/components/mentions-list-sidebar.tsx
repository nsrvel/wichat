"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import { MentionsListRow } from "@/features/mentions/components/mentions-list-row";
import {
	mentionsFilterToKind,
	type MentionsFilterId,
} from "@/features/mentions/lib/mentions-filters";
import { sessionQueryOpts } from "@/services/auth/auth.query";
import type { GroupMentionHit } from "@wichat/contracts/chat/mentions";
import { groupChannelsQueryOpts } from "@/services/channel/channel.query";
import { groupMembersQueryOpts } from "@/services/invite/invite.query";
import {
	Q_GROUP_MENTIONS,
	Q_GROUP_MENTIONS_UNREAD,
	groupMentionsQueryOpts,
} from "@/services/mentions/mentions.query";
import { markGroupMentionsRead } from "@/services/mentions/mentions.service";
import { cn } from "@/lib/utils";

function formatMentionDate(iso: string): string {
	const d = new Date(iso);
	if (Number.isNaN(d.getTime())) {
		return "";
	}
	const day = String(d.getDate()).padStart(2, "0");
	const month = String(d.getMonth() + 1).padStart(2, "0");
	return `${day}/${month}`;
}

type MentionsListSidebarProps = {
	groupId: string;
	filter: MentionsFilterId;
	selectedChannelId?: string;
	selectedMessageId?: string;
	onSelect: (channelId: string, messageId: string) => void;
	onFilterChange: (filter: MentionsFilterId) => void;
};

const FILTER_PILLS: MentionsFilterId[] = ["unread", "channels", "chats"];

export function MentionsListSidebar({
	groupId,
	filter,
	selectedChannelId,
	selectedMessageId,
	onSelect,
	onFilterChange,
}: MentionsListSidebarProps) {
	const { t } = useTranslation();
	const queryClient = useQueryClient();
	const markRead = useMutation({
		mutationFn: (messageIds: string[]) =>
			markGroupMentionsRead(groupId, { message_ids: messageIds }),
		onSuccess: async () => {
			await Promise.all([
				queryClient.invalidateQueries({ queryKey: [Q_GROUP_MENTIONS, groupId] }),
				queryClient.invalidateQueries({ queryKey: [Q_GROUP_MENTIONS_UNREAD, groupId] }),
			]);
		},
	});
	const kind = mentionsFilterToKind(filter);
	const listKind =
		kind === "unread" ? "unread" : kind === "channels" || kind === "chats" ? kind : undefined;
	const { data: sessionUser } = useQuery(sessionQueryOpts());
	const { data: channelsData } = useQuery(groupChannelsQueryOpts(groupId));
	const { data: membersData } = useQuery(groupMembersQueryOpts(groupId, ""));
	const { data, isPending, isError } = useQuery(
		groupMentionsQueryOpts(groupId, { kind: listKind, limit: 30 }),
	);

	const channelImageById = useMemo(() => {
		const map = new Map<string, string>();
		for (const ch of channelsData?.channels ?? []) {
			if (ch.id && ch.image_url) {
				map.set(ch.id, ch.image_url);
			}
		}
		return map;
	}, [channelsData?.channels]);

	const memberAvatarByName = useMemo(() => {
		const map = new Map<string, string>();
		for (const m of membersData?.members ?? []) {
			if (m.avatar_url) {
				if (m.display_name) {
					map.set(m.display_name.toLowerCase(), m.avatar_url);
				}
				if (m.username) {
					map.set(m.username.toLowerCase(), m.avatar_url);
				}
			}
		}
		return map;
	}, [membersData?.members]);

	const hits = data?.hits ?? [];

	return (
		<div className="flex h-full min-h-0 w-full max-w-sm flex-col border-r border-border bg-sidebar lg:w-80">
			<div className="border-b border-border px-3 pt-3 pb-2.5">
				<h1 className="mb-2.5 text-base font-semibold tracking-tight text-foreground">
					{t("shell.sidebar.mentions")}
				</h1>
				<div className="flex flex-wrap gap-1.5">
					{FILTER_PILLS.map((id) => {
						return (
							<button
								key={id}
								type="button"
								aria-pressed={filter === id}
								onClick={() => onFilterChange(id)}
								className={cn(
									"rounded-full px-2.5 py-1 text-xs font-medium transition",
									filter === id
										? "bg-primary text-primary-foreground"
										: "bg-muted text-muted-foreground hover:bg-muted/80",
								)}
							>
								{t(`shell.mentions.filter_${id}`)}
							</button>
						);
					})}
				</div>
			</div>
			<div className="overlay-scroll min-h-0 flex-1 space-y-0.5 overflow-y-auto p-1.5">
				{isPending ? (
					<p className="px-3 py-6 text-center text-xs text-muted-foreground">
						{t("shell.loading")}
					</p>
				) : isError ? (
					<p className="px-3 py-6 text-center text-xs text-muted-foreground">
						{t("shell.mentions.load_error")}
					</p>
				) : hits.length === 0 ? (
					<p className="px-3 py-6 text-center text-xs text-muted-foreground">
						{t("shell.mentions.empty")}
					</p>
				) : (
					hits.map((hit: GroupMentionHit) => (
						<MentionsListRow
							key={`${hit.channel_id}-${hit.message_id}`}
							groupId={groupId}
							channelName={hit.channel_name}
							channelType={hit.channel_type}
							channelImageUrl={channelImageById.get(hit.channel_id)}
							authorDisplayName={hit.author_display_name}
							authorAvatarUrl={memberAvatarByName.get(hit.author_display_name.toLowerCase())}
							bodyPreview={hit.body_preview}
							createdAtLabel={formatMentionDate(hit.created_at)}
							isUnread={hit.is_unread}
							selected={
								selectedChannelId === hit.channel_id && selectedMessageId === hit.message_id
							}
							selfUserId={sessionUser?.id}
							onClick={() => {
								if (hit.is_unread) {
									markRead.mutate([hit.message_id]);
								}
								onSelect(hit.channel_id, hit.message_id);
							}}
						/>
					))
				)}
			</div>
		</div>
	);
}
