"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import { MentionsListRow } from "@/features/mentions/components/mentions-list-row";
import {
	mentionsFilterToKind,
	type MentionsFilterId,
} from "@/features/mentions/lib/mentions-filters";
import { sessionQueryOpts } from "@/services/auth/auth.query";
import type { GroupMentionHit } from "@wichat/contracts/chat/mentions";
import {
	Q_GROUP_MENTIONS,
	Q_GROUP_MENTIONS_UNREAD,
	groupMentionsQueryOpts,
} from "@/services/mentions/mentions.query";
import { markGroupMentionsRead } from "@/services/mentions/mentions.service";
import { cn } from "@/lib/utils";

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
		mutationFn: (messageIds: string[]) => markGroupMentionsRead(groupId, { message_ids: messageIds }),
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
	const { data, isPending, isError } = useQuery(
		groupMentionsQueryOpts(groupId, { kind: listKind, limit: 30 }),
	);

	const hits = data?.hits ?? [];

	return (
		<div className="flex h-full min-h-0 w-full max-w-sm flex-col border-r border-border bg-sidebar lg:w-80">
			<div className="flex flex-wrap gap-1.5 border-b border-border px-3 py-2.5">
				{FILTER_PILLS.map((id) => {
					return (
						<button
							key={id}
							type="button"
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
			<div className="overlay-scroll min-h-0 flex-1 overflow-y-auto">
				{isPending ? (
					<p className="px-3 py-6 text-center text-xs text-muted-foreground">{t("shell.loading")}</p>
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
							channelName={hit.channel_name}
							channelType={hit.channel_type}
							authorDisplayName={hit.author_display_name}
							bodyPreview={hit.body_preview}
							createdAtLabel={new Date(hit.created_at).toLocaleDateString(undefined, {
								month: "short",
								day: "numeric",
							})}
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