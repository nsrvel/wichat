"use client";

import { ApiError } from "@wichat/api-client/api-error";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import type { MessageReplyPreview } from "@wichat/contracts/chat/message";
import type { GroupMember } from "@wichat/contracts/group/member";
import { ChevronDown, EyeOff, ShieldAlert } from "lucide-react";
import { memo, useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { GroupMemberProfileHost } from "@/components/group/group-member-profile-host";
import { toChannelShellNavTarget } from "@/features/channel/lib/channel-shell-nav-target";
import { ChannelEmptyState } from "@/features/chat/components/channel-empty-state";
import { ChannelMessageSkeleton } from "@/features/chat/components/channel-message-skeleton";
import { ChatComposer } from "@/features/chat/components/chat-composer";
import { ConversationHeader } from "@/features/chat/components/conversation-header";
import { MessageActionBar } from "@/features/chat/components/message-action-bar";
import { MessageBody } from "@/features/chat/components/message-body";
import {
	MessageBubbleContextMenu,
	type MessageMenuActionProps,
} from "@/features/chat/components/message-context-menu";
import { MessageList } from "@/features/chat/components/message-list";
import {
	MessageReactions,
	type ReactionUserInfo,
} from "@/features/chat/components/message-reactions";
import { MessageRow } from "@/features/chat/components/message-row";
import { PinnedMessagesPopover } from "@/features/chat/components/pinned-messages-popover";
import type { ChannelMessage } from "@/features/chat/lib/channel-message";
import { groupTimelineItems } from "@/features/chat/lib/group-timeline-items";
import { ChatCacheTrimHintProvider } from "@/features/chat/lib/chat-cache-trim-hint";
import { type ChannelMessagesInfiniteData } from "@/features/chat/lib/message-cache-window";
import { resolveJumpSeq } from "@/features/chat/lib/resolve-message-jump";
import { useChannelMessages } from "@/features/chat/lib/use-channel-messages";
import { Q_CHANNEL_MESSAGES } from "@/services/message/message.key";
import { hasInlineReplyBlock, stripInlineReplyBlocks } from "@/features/chat/lib/composer-markdown";
import { isPendingMessageId } from "@/features/chat/lib/message-seq";
import { useChannelRealtime } from "@/features/chat/lib/use-channel-realtime";
import { useChatScroller } from "@/features/chat/lib/use-chat-scroller";
import { translateError } from "@/integrations/i18n/translate-error";
import { resolveMemberNameColor } from "@/lib/member-role-display";
import { sessionQueryOpts } from "@/services/auth/auth.query";
import * as channelQuery from "@/services/channel/channel.query";
import { myGroupsQueryOpts } from "@/services/group/group.query";
import { uploadChannelAttachment } from "@/services/message/message.service";
import { memberPresenceDot, type AvatarPresenceDot } from "@/features/presence/presence-dot";
import { groupMembersQueryOpts } from "@/services/invite/invite.query";
import {
	useCreateMessageMutation,
	useDeleteMessageMutation,
	usePinMessageMutation,
	useToggleReactionMutation,
	useUpdateMessageMutation,
} from "@/services/message/message.mutation";

type ChannelChatViewProps = {
	groupId: string;
	channelId: string;
	onOpenSidebar?: () => void;
};

function formatDateDividerLabel(isoDate: string, todayLabel: string, yesterdayLabel: string) {
	const d = new Date(isoDate);
	if (Number.isNaN(d.getTime())) {
		return "";
	}
	const now = new Date();
	const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
	const target = new Date(d.getFullYear(), d.getMonth(), d.getDate());
	const diffDays = Math.round((today.getTime() - target.getTime()) / 86_400_000);
	if (diffDays === 0) {
		return todayLabel;
	}
	if (diffDays === 1) {
		return yesterdayLabel;
	}
	return d.toLocaleDateString(undefined, {
		weekday: "long",
		year: "numeric",
		month: "short",
		day: "numeric",
	});
}

type TimelineMessageItemProps = {
	message: ChannelMessage;
	compact: boolean;
	authorAvatarFallbackUrl?: string;
	authorNameColor?: string;
	authorIsOwner: boolean;
	authorPresenceStatus?: AvatarPresenceDot;
	ownerLabel: string;
	canReact: boolean;
	canPin: boolean;
	canManageMessages: boolean;
	highlighted: boolean;
	isEditing: boolean;
	hasAuthorProfile: boolean;
	onJumpToMessage: (messageId: string) => void;
	onSelectAuthor: (authorId: string) => void;
	onReplyMessage: (message: ChannelMessage) => void;
	onStartEdit: (messageId: string) => void;
	onCancelEdit: () => void;
	onSaveEdit: (messageId: string, nextBody: string) => void;
	onTogglePin: (messageId: string, nextPinned: boolean) => void;
	onDeleteMessage: (messageId: string) => void;
	onToggleReaction: (messageId: string, emoji: string) => void;
	resolveReactor: (userId: string) => ReactionUserInfo | undefined;
};

const TimelineMessageItem = memo(function TimelineMessageItem({
	message,
	compact,
	authorAvatarFallbackUrl,
	authorNameColor,
	authorIsOwner,
	authorPresenceStatus,
	ownerLabel,
	canReact,
	canPin,
	canManageMessages,
	highlighted,
	isEditing,
	hasAuthorProfile,
	onJumpToMessage,
	onSelectAuthor,
	onReplyMessage,
	onStartEdit,
	onCancelEdit,
	onSaveEdit,
	onTogglePin,
	onDeleteMessage,
	onToggleReaction,
	resolveReactor,
}: TimelineMessageItemProps) {
	const isOpt = isPendingMessageId(message.id);
	const isMine = Boolean(message.mine);
	const canDelete = isMine || canManageMessages;
	const hasInlineReply = useMemo(() => hasInlineReplyBlock(message.bodyPlain), [message.bodyPlain]);

	const menuActions = useMemo<MessageMenuActionProps>(
		() => ({
			isMine,
			isPinned: Boolean(message.isPinned),
			canPin,
			canDelete,
			onReply: () => onReplyMessage(message),
			onEdit: isMine ? () => onStartEdit(message.id) : undefined,
			onTogglePin: canPin ? () => onTogglePin(message.id, !message.isPinned) : undefined,
			onCopy: () => {
				const plain = stripInlineReplyBlocks(message.bodyPlain);
				if (plain) {
					void navigator.clipboard.writeText(plain);
				}
			},
			onDelete: canDelete ? () => onDeleteMessage(message.id) : undefined,
		}),
		[canDelete, canPin, isMine, message, onDeleteMessage, onReplyMessage, onStartEdit, onTogglePin],
	);

	const handleSaveEdit = useCallback(
		(nextBody: string) => {
			onSaveEdit(message.id, nextBody);
		},
		[message.id, onSaveEdit],
	);

	const handleAuthorClick = useMemo(
		() => (hasAuthorProfile ? () => onSelectAuthor(message.authorId) : undefined),
		[hasAuthorProfile, message.authorId, onSelectAuthor],
	);

	const renderBubbleWrapper = useMemo(
		() =>
			!isOpt
				? (bubble: React.ReactNode) => (
						<MessageBubbleContextMenu {...menuActions}>{bubble}</MessageBubbleContextMenu>
					)
				: undefined,
		[isOpt, menuActions],
	);

	const handleReact = useCallback(
		(emoji: string) => {
			onToggleReaction(message.id, emoji);
		},
		[message.id, onToggleReaction],
	);

	const actionBar = useMemo(
		() =>
			!isOpt ? (
				<MessageActionBar {...menuActions} canReact={canReact} onReact={handleReact} />
			) : undefined,
		[canReact, handleReact, isOpt, menuActions],
	);

	const reactionsNode = useMemo(
		() =>
			message.reactions && message.reactions.length > 0 ? (
				<MessageReactions
					reactions={message.reactions}
					showAdd={canReact}
					resolveReactor={resolveReactor}
					onToggleReaction={handleReact}
				/>
			) : undefined,
		[canReact, handleReact, message.reactions, resolveReactor],
	);

	return (
		<MessageRow
			messageId={message.id}
			authorAvatarUrl={message.authorAvatarUrl || authorAvatarFallbackUrl || ""}
			authorDisplayName={message.authorDisplayName}
			authorNameColor={authorNameColor}
			authorIsOwner={authorIsOwner}
			authorPresenceStatus={authorPresenceStatus}
			ownerLabel={ownerLabel}
			timeLabel={message.timeLabel}
			mine={isMine}
			edited={Boolean(message.editedAt)}
			isPinned={Boolean(message.isPinned)}
			highlighted={highlighted}
			compact={compact}
			replyTo={message.replyTo}
			hasInlineReply={hasInlineReply}
			onReplyClick={onJumpToMessage}
			isEditing={isEditing}
			initialEditText={message.bodyPlain}
			onSaveEdit={handleSaveEdit}
			onCancelEdit={onCancelEdit}
			onAuthorClick={handleAuthorClick}
			renderBubbleWrapper={renderBubbleWrapper}
			actionBar={actionBar}
			reactions={reactionsNode}
		>
			<MessageBody text={message.bodyPlain} onReplyClick={onJumpToMessage} />
		</MessageRow>
	);
});

type ChannelSlowmodeComposerProps = {
	groupId: string;
	channelId: string;
	composerDisabled: boolean;
	disabledReason?: string;
	slowmodeSeconds: number;
	slowmodeUntilMs: number;
	canBypassSlowmode: boolean;
	replyTo: MessageReplyPreview | null;
	replyTimeLabel?: string;
	onSubmit: (body: string) => void;
	onCancelReply: () => void;
	onReplyClick: (messageId: string) => void;
};

const ChannelSlowmodeComposer = memo(function ChannelSlowmodeComposer({
	groupId,
	channelId,
	composerDisabled,
	disabledReason,
	slowmodeSeconds,
	slowmodeUntilMs,
	canBypassSlowmode,
	replyTo,
	replyTimeLabel,
	onSubmit,
	onCancelReply,
	onReplyClick,
}: ChannelSlowmodeComposerProps) {
	const { t } = useTranslation();
	const [nowMs, setNowMs] = useState(() => Date.now());

	useEffect(() => {
		if (canBypassSlowmode || slowmodeUntilMs <= 0) {
			return;
		}
		const tick = () => setNowMs(Date.now());
		tick();
		if (slowmodeUntilMs <= Date.now()) {
			return;
		}
		const id = window.setInterval(() => {
			const current = Date.now();
			setNowMs(current);
			if (current >= slowmodeUntilMs) {
				window.clearInterval(id);
			}
		}, 1000);
		return () => window.clearInterval(id);
	}, [canBypassSlowmode, slowmodeUntilMs]);

	const slowmodeRemaining =
		!canBypassSlowmode && slowmodeUntilMs > nowMs ? Math.ceil((slowmodeUntilMs - nowMs) / 1000) : 0;

	const handleUploadFile = useCallback(
		(file: File) => uploadChannelAttachment(groupId, channelId, file),
		[channelId, groupId],
	);

	return (
		<ChatComposer
			onSubmit={onSubmit}
			onUploadFile={handleUploadFile}
			disabled={composerDisabled}
			disabledReason={disabledReason}
			submitDisabled={slowmodeRemaining > 0}
			replyTo={replyTo}
			replyTimeLabel={replyTimeLabel}
			onCancelReply={onCancelReply}
			onReplyClick={onReplyClick}
			statusNotice={
				slowmodeRemaining > 0
					? t("chat.slowmode_wait", { seconds: slowmodeRemaining })
					: slowmodeSeconds > 0 && !canBypassSlowmode && !composerDisabled
						? t("chat.slowmode_active", { seconds: slowmodeSeconds })
						: undefined
			}
		/>
	);
});

export function ChannelChatView(props: ChannelChatViewProps) {
	return (
		<ChatCacheTrimHintProvider>
			<ChannelChatViewBody {...props} />
		</ChatCacheTrimHintProvider>
	);
}

function ChannelChatViewBody({ groupId, channelId, onOpenSidebar }: ChannelChatViewProps) {
	const { t } = useTranslation();
	const queryClient = useQueryClient();
	const [selectedMemberId, setSelectedMemberId] = useState<string | null>(null);
	const [replyState, setReplyState] = useState<{
		channelId: string;
		preview: MessageReplyPreview;
		timeLabel?: string;
	} | null>(null);
	const [editingState, setEditingState] = useState<{
		channelId: string;
		messageId: string;
	} | null>(null);
	const replyTo = replyState?.channelId === channelId ? replyState.preview : null;
	const editingMessageId = editingState?.channelId === channelId ? editingState.messageId : null;
	const [acknowledgedChannels, setAcknowledgedChannels] = useState<Set<string>>(() => new Set());
	const [slowmodeUntilByChannel, setSlowmodeUntilByChannel] = useState<Record<string, number>>({});

	const channels = useQuery(channelQuery.groupChannelsQueryOpts(groupId));
	const activeChannel = useMemo(
		() => channels.data?.channels.find((ch) => ch.id === channelId),
		[channels.data?.channels, channelId],
	);
	const active = toChannelShellNavTarget(channelId, channels.data?.channels);
	const { data: sessionUser } = useQuery(sessionQueryOpts());
	const { data: myGroups = [] } = useQuery(myGroupsQueryOpts());
	const hideRoles = myGroups.find((g) => g.id === groupId)?.hide_roles ?? false;
	const membersQuery = useQuery({
		...groupMembersQueryOpts(groupId, ""),
		enabled: Boolean(groupId),
	});

	const memberById = useMemo(() => {
		const map = new Map<string, GroupMember>();
		for (const member of membersQuery.data?.members ?? []) {
			map.set(member.user_id, member);
		}
		return map;
	}, [membersQuery.data?.members]);

	const perms = useMemo(
		() => activeChannel?.caller_permissions ?? [],
		[activeChannel?.caller_permissions],
	);
	const isAdmin = perms.includes("administrator");
	const hasSendMessages = perms.length === 0 || isAdmin || perms.includes("send_messages");
	const canPostBroadcast =
		activeChannel?.type !== "broadcast" ||
		isAdmin ||
		perms.includes("manage_channels") ||
		perms.includes("manage_messages");
	const canBypassSlowmode =
		isAdmin ||
		perms.includes("bypass_slowmode") ||
		perms.includes("manage_channels") ||
		perms.includes("manage_messages");
	const canReact = perms.length === 0 || isAdmin || perms.includes("add_reactions");
	const canPin = isAdmin || perms.includes("pin_messages") || perms.includes("manage_messages");
	const canManageMessages = isAdmin || perms.includes("manage_messages");

	const composerDisabled = !hasSendMessages || !canPostBroadcast;
	const disabledReason = !hasSendMessages
		? t("chat.composer_no_permission")
		: !canPostBroadcast
			? t("chat.composer_broadcast_readonly")
			: undefined;

	const slowmodeSeconds = activeChannel?.slowmode_seconds ?? 0;
	const slowmodeUntilMs = slowmodeUntilByChannel[channelId] ?? 0;

	const contentVisibility = activeChannel?.content_visibility ?? "default";
	const needsVisibilityAck =
		(contentVisibility === "spoiler" || contentVisibility === "age_restricted") &&
		!acknowledgedChannels.has(channelId);

	const {
		messages,
		isPending,
		isFetching,
		isError,
		hasOlderMessages,
		isFetchingOlderMessages,
		isHistoryLoadPaused,
		fetchOlderMessages,
		loadContextAroundSeq,
		refetch,
	} = useChannelMessages(groupId, channelId);

	const timelineItems = useMemo(() => groupTimelineItems(messages), [messages]);
	const showLoadingSkeleton = isPending && messages.length === 0 && !isError;
	const showEmpty = !isPending && !isError && messages.length === 0;
	const showBackgroundRefresh = isFetching && !isPending && messages.length > 0;

	const {
		scrollRef,
		contentRef,
		virtualizer,
		showJumpToLatest,
		unreadNewCount,
		highlightedMessageId,
		handleScroll,
		scrollToBottom,
		scrollToMessage,
		notifyIncomingMessage,
		prepareForSend,
	} = useChatScroller({
		channelId,
		items: timelineItems,
		hasOlderMessages,
		isFetchingOlderMessages,
		fetchOlderMessages,
	});

	const pendingJumpRef = useRef<{ targetId: string; attempts: number } | null>(null);
	const jumpDepsRef = useRef({
		hasOlderMessages,
		isFetchingOlderMessages,
		fetchOlderMessages,
		loadContextAroundSeq,
		scrollToMessage,
		t,
		queryClient,
		groupId,
		channelId,
	});
	useLayoutEffect(() => {
		jumpDepsRef.current = {
			hasOlderMessages,
			isFetchingOlderMessages,
			fetchOlderMessages,
			loadContextAroundSeq,
			scrollToMessage,
			t,
			queryClient,
			groupId,
			channelId,
		};
	});

	useEffect(() => {
		pendingJumpRef.current = null;
	}, [channelId]);

	const handleJumpToMessage = useCallback((targetId: string) => {
		if (!targetId) {
			return;
		}
		const {
			loadContextAroundSeq: loadAround,
			scrollToMessage: jumpNow,
			queryClient: qc,
			groupId: gid,
			channelId: cid,
			t: translate,
		} = jumpDepsRef.current;
		if (jumpNow(targetId)) {
			pendingJumpRef.current = null;
			return;
		}

		pendingJumpRef.current = { targetId, attempts: 0 };
		void (async () => {
			try {
				const seq = await resolveJumpSeq({
					groupId: gid,
					channelId: cid,
					messageId: targetId,
					cache: qc.getQueryData<ChannelMessagesInfiniteData>([Q_CHANNEL_MESSAGES, gid, cid]),
				});
				if (seq <= 0) {
					pendingJumpRef.current = null;
					toast.info(translate("chat.reply_target_unavailable"));
					return;
				}
				const loaded = await loadAround(seq);
				if (!loaded) {
					pendingJumpRef.current = null;
					toast.info(translate("chat.reply_target_unavailable"));
					return;
				}
				requestAnimationFrame(() => {
					if (jumpNow(targetId)) {
						pendingJumpRef.current = null;
					}
				});
			} catch {
				pendingJumpRef.current = null;
				toast.info(translate("chat.reply_target_unavailable"));
			}
		})();
	}, []);

	useEffect(() => {
		const pending = pendingJumpRef.current;
		if (!pending) {
			return;
		}
		if (scrollToMessage(pending.targetId)) {
			pendingJumpRef.current = null;
		}
	}, [messages, scrollToMessage]);

	const handleRealtimeMessage = useCallback(
		(authorId: string) => {
			notifyIncomingMessage(Boolean(sessionUser?.id) && authorId === sessionUser?.id);
		},
		[notifyIncomingMessage, sessionUser?.id],
	);

	useChannelRealtime(groupId, channelId, handleRealtimeMessage);

	const createMessage = useCreateMessageMutation();
	const updateMessage = useUpdateMessageMutation();
	const deleteMessage = useDeleteMessageMutation();
	const toggleReaction = useToggleReactionMutation();
	const pinMessage = usePinMessageMutation();

	const reactorLookupRef = useRef({ memberById, sessionUser });
	useLayoutEffect(() => {
		reactorLookupRef.current = { memberById, sessionUser };
	});

	const resolveReactor = useCallback((userId: string): ReactionUserInfo | undefined => {
		const { memberById: members, sessionUser: me } = reactorLookupRef.current;
		const m = members.get(userId);
		if (m) {
			return {
				displayName: m.display_name || m.username,
				avatarUrl: m.avatar_url,
			};
		}
		if (me?.id === userId) {
			return {
				displayName: me.display_name || me.username,
				avatarUrl: me.avatar_url,
			};
		}
		return undefined;
	}, []);

	const rowActionCtxRef = useRef({
		groupId,
		channelId,
		updateMessage,
		deleteMessage,
		toggleReaction,
		pinMessage,
		t,
	});
	useLayoutEffect(() => {
		rowActionCtxRef.current = {
			groupId,
			channelId,
			updateMessage,
			deleteMessage,
			toggleReaction,
			pinMessage,
			t,
		};
	});

	const handleSelectAuthor = useCallback((authorId: string) => {
		setSelectedMemberId(authorId);
	}, []);

	const handleCloseMemberProfile = useCallback(() => {
		setSelectedMemberId(null);
	}, []);

	const handleReplyMessage = useCallback((message: ChannelMessage) => {
		const { channelId: activeChanId } = rowActionCtxRef.current;
		setReplyState({
			channelId: activeChanId,
			preview: {
				id: message.id,
				author_id: message.authorId,
				author_display_name: message.authorDisplayName,
				body_preview: stripInlineReplyBlocks(message.bodyPlain).slice(0, 120),
			},
			timeLabel: message.timeLabel,
		});
	}, []);

	const handleCancelReply = useCallback(() => {
		setReplyState(null);
	}, []);

	const handleStartEdit = useCallback((messageId: string) => {
		const { channelId: activeChanId } = rowActionCtxRef.current;
		setEditingState({ channelId: activeChanId, messageId });
	}, []);

	const handleCancelEdit = useCallback(() => {
		setEditingState(null);
	}, []);

	const handleSaveEdit = useCallback((messageId: string, nextBody: string) => {
		const {
			groupId: gid,
			channelId: cid,
			updateMessage: mut,
			t: translate,
		} = rowActionCtxRef.current;
		setEditingState(null);
		mut.mutate(
			{ groupId: gid, channelId: cid, messageId, body: nextBody },
			{ onError: (err) => toast.error(translateError(translate, err)) },
		);
	}, []);

	const handleTogglePin = useCallback((messageId: string, nextPinned: boolean) => {
		const { groupId: gid, channelId: cid, pinMessage: mut, t: translate } = rowActionCtxRef.current;
		mut.mutate(
			{ groupId: gid, channelId: cid, messageId, pinned: nextPinned },
			{ onError: (err) => toast.error(translateError(translate, err)) },
		);
	}, []);

	const handleDeleteMessage = useCallback((messageId: string) => {
		const {
			groupId: gid,
			channelId: cid,
			deleteMessage: mut,
			t: translate,
		} = rowActionCtxRef.current;
		mut.mutate(
			{ groupId: gid, channelId: cid, messageId },
			{ onError: (err) => toast.error(translateError(translate, err)) },
		);
	}, []);

	const handleToggleReaction = useCallback((messageId: string, emoji: string) => {
		const {
			groupId: gid,
			channelId: cid,
			toggleReaction: mut,
			t: translate,
		} = rowActionCtxRef.current;
		mut.mutate(
			{ groupId: gid, channelId: cid, messageId, emoji },
			{ onError: (err) => toast.error(translateError(translate, err)) },
		);
	}, []);

	const handleSendMessage = useCallback(
		(body: string) => {
			prepareForSend();
			const activeReply = replyTo;
			setReplyState(null);

			createMessage.mutate(
				{
					groupId,
					channelId,
					body,
					reply_to_id: activeReply?.id,
					replyPreview: activeReply ?? undefined,
					authorDisplayName: sessionUser?.display_name,
					authorAvatarUrl: sessionUser?.avatar_url ?? "",
					authorId: sessionUser?.id,
				},
				{
					onSuccess: () => {
						if (slowmodeSeconds > 0 && !canBypassSlowmode) {
							const now = Date.now();
							setSlowmodeUntilByChannel((prev) => ({
								...prev,
								[channelId]: now + slowmodeSeconds * 1000,
							}));
						}
					},
					onError: (err) => {
						if (err instanceof ApiError && err.code === "error.rate_limited") {
							const retry = Number(err.params?.retry_after_seconds);
							if (Number.isFinite(retry) && retry > 0) {
								const now = Date.now();
								setSlowmodeUntilByChannel((prev) => ({
									...prev,
									[channelId]: now + retry * 1000,
								}));
								toast.error(t("chat.slowmode_wait", { seconds: Math.ceil(retry) }));
								return;
							}
						}
						toast.error(translateError(t, err));
					},
				},
			);
		},
		[
			prepareForSend,
			replyTo,
			createMessage,
			groupId,
			channelId,
			sessionUser,
			slowmodeSeconds,
			canBypassSlowmode,
			t,
		],
	);

	const ownerLabel = t("invite.role_owner");

	return (
		<div className="relative flex h-full min-w-0 flex-1 flex-col bg-chat-surface text-foreground">
			<div className="lg:hidden">
				<ConversationHeader active={active} onOpenSidebar={onOpenSidebar} />
			</div>

			{groupId && channelId ? (
				<div className="pointer-events-none absolute right-4 top-2.5 z-20 hidden lg:block">
					<div className="pointer-events-auto">
						<PinnedMessagesPopover groupId={groupId} channelId={channelId} canPin={canPin} />
					</div>
				</div>
			) : null}

			<div className="relative flex min-h-0 flex-1 flex-col">
				{needsVisibilityAck ? (
					<div className="flex flex-1 items-center justify-center p-6">
						<div className="flex max-w-md flex-col items-center gap-3 rounded-xl border border-border bg-card p-6 text-center shadow-sm">
							<div className="flex size-11 items-center justify-center rounded-full bg-amber-500/15 text-amber-600 dark:text-amber-400">
								{contentVisibility === "age_restricted" ? (
									<ShieldAlert className="size-5" />
								) : (
									<EyeOff className="size-5" />
								)}
							</div>
							<h3 className="text-base font-semibold text-foreground">
								{contentVisibility === "age_restricted"
									? t("chat.visibility_age_title")
									: t("chat.visibility_spoiler_title")}
							</h3>
							<p className="text-xs leading-relaxed text-muted-foreground">
								{contentVisibility === "age_restricted"
									? t("chat.visibility_age_desc")
									: t("chat.visibility_spoiler_desc")}
							</p>
							<button
								type="button"
								onClick={() =>
									setAcknowledgedChannels((prev) => {
										const next = new Set(prev);
										next.add(channelId);
										return next;
									})
								}
								className="mt-1 rounded-lg bg-primary px-4 py-2 text-xs font-semibold text-primary-foreground transition hover:brightness-110"
							>
								{t("chat.visibility_reveal_cta")}
							</button>
						</div>
					</div>
				) : (
					<>
						<div className="relative flex min-h-0 flex-1 flex-col">
							{showLoadingSkeleton ? (
								<div className="overlay-scroll min-h-0 flex-1 overflow-y-auto">
									<ChannelMessageSkeleton />
								</div>
							) : null}
							{showBackgroundRefresh ? (
								<div className="pointer-events-none absolute top-3 right-4 z-10">
									<div className="flex items-center gap-2 rounded-full border border-border/80 bg-card/95 px-2.5 py-1 text-[11px] text-muted-foreground shadow-sm backdrop-blur-xs">
										<span className="inline-block size-2.5 animate-spin rounded-full border-2 border-primary border-t-transparent" />
										{t("chat.syncing")}
									</div>
								</div>
							) : null}
							{isFetchingOlderMessages ? (
								<div className="pointer-events-none absolute top-3 left-1/2 z-10 -translate-x-1/2">
									<div className="flex items-center gap-2 rounded-full border border-border/80 bg-card/95 px-3 py-1 text-xs text-muted-foreground shadow-sm backdrop-blur-xs">
										<span className="inline-block size-3 animate-spin rounded-full border-2 border-primary border-t-transparent" />
										{t("chat.loading_older")}
									</div>
								</div>
							) : null}
							{isHistoryLoadPaused ? (
								<div className="pointer-events-none absolute top-3 left-1/2 z-10 -translate-x-1/2">
									<div className="max-w-sm rounded-full border border-border/80 bg-card/95 px-3 py-1 text-center text-xs text-muted-foreground shadow-sm backdrop-blur-xs">
										{t("chat.history_rate_limited")}
									</div>
								</div>
							) : null}

							<MessageList
								scrollRef={scrollRef}
								contentRef={contentRef}
								onScroll={handleScroll}
								className={showLoadingSkeleton ? "hidden" : undefined}
								innerClassName={showEmpty || isError ? "min-h-full justify-center py-0" : undefined}
								virtualizer={!isError && !showEmpty && !showLoadingSkeleton ? virtualizer : undefined}
								header={
									!hasOlderMessages && messages.length > 0 && !isError ? (
										<div className="flex flex-col items-center justify-center gap-1.5 border-b border-border/40 pb-6 pt-8 text-center">
											<h3 className="text-base font-semibold text-foreground">
												{t("chat.history_start_title", {
													name: active?.channelName || active?.label || "channel",
												})}
											</h3>
											<p className="max-w-md text-xs text-muted-foreground">
												{t("chat.history_start_description", {
													name: active?.channelName || active?.label || "channel",
												})}
											</p>
										</div>
									) : undefined
								}
								renderVirtualItem={(index) => {
									const item = timelineItems[index];
									if (!item) {
										return null;
									}
									if (item.kind === "date-divider") {
										const label = formatDateDividerLabel(
											item.isoDate,
											t("chat.today"),
											t("chat.yesterday"),
										);
										return (
											<div key={item.key} className="my-4 flex items-center justify-center">
												<span className="text-[11px] font-semibold text-muted-foreground">
													{label}
												</span>
											</div>
										);
									}

									const message = item.message;
									const member = memberById.get(message.authorId);
									const authorNameColor = member
										? resolveMemberNameColor(member.roles, hideRoles)
										: undefined;
									const authorIsOwner = member?.role === "owner";

									return (
										<TimelineMessageItem
											key={item.key}
											message={message}
											compact={item.compact}
											authorAvatarFallbackUrl={member?.avatar_url}
											authorNameColor={authorNameColor}
											authorIsOwner={authorIsOwner}
											authorPresenceStatus={member ? memberPresenceDot(member) : undefined}
											ownerLabel={ownerLabel}
											canReact={canReact}
											canPin={canPin}
											canManageMessages={canManageMessages}
											highlighted={highlightedMessageId === message.id}
											isEditing={editingMessageId === message.id}
											hasAuthorProfile={Boolean(groupId && message.authorId)}
											onJumpToMessage={handleJumpToMessage}
											onSelectAuthor={handleSelectAuthor}
											onReplyMessage={handleReplyMessage}
											onStartEdit={handleStartEdit}
											onCancelEdit={handleCancelEdit}
											onSaveEdit={handleSaveEdit}
											onTogglePin={handleTogglePin}
											onDeleteMessage={handleDeleteMessage}
											onToggleReaction={handleToggleReaction}
											resolveReactor={resolveReactor}
										/>
									);
								}}
							>
								{isError ? (
									<div className="flex flex-col items-center justify-center gap-3 py-8 text-center text-sm">
										<p className="text-muted-foreground">{t("chat.load_error")}</p>
										<button
											type="button"
											onClick={refetch}
											className="rounded-md border border-border bg-card px-3 py-1.5 text-xs font-medium text-foreground shadow-xs transition hover:bg-muted"
										>
											{t("chat.retry")}
										</button>
									</div>
								) : null}
								{showEmpty ? <ChannelEmptyState /> : null}
							</MessageList>

							{showJumpToLatest ? (
								<div className="absolute bottom-3 left-1/2 z-10 -translate-x-1/2 animate-in fade-in slide-in-from-bottom-2 duration-150">
									<button
										type="button"
										onClick={() => scrollToBottom("smooth")}
										className="flex items-center gap-1.5 rounded-full border border-border bg-primary px-3.5 py-1.5 text-xs font-semibold text-primary-foreground shadow-lg transition hover:brightness-110 active:scale-95"
									>
										<ChevronDown className="size-3.5" />
										{unreadNewCount > 0
											? t("chat.unread_new_messages", { count: unreadNewCount })
											: t("chat.jump_to_latest")}
									</button>
								</div>
							) : null}
						</div>

						<ChannelSlowmodeComposer
							groupId={groupId}
							channelId={channelId}
							composerDisabled={composerDisabled}
							disabledReason={disabledReason}
							slowmodeSeconds={slowmodeSeconds}
							slowmodeUntilMs={slowmodeUntilMs}
							canBypassSlowmode={canBypassSlowmode}
							replyTo={replyTo}
							replyTimeLabel={
								replyState?.channelId === channelId ? replyState.timeLabel : undefined
							}
							onSubmit={handleSendMessage}
							onCancelReply={handleCancelReply}
							onReplyClick={handleJumpToMessage}
						/>
					</>
				)}
			</div>
			{groupId && selectedMemberId ? (
				<GroupMemberProfileHost
					groupId={groupId}
					memberId={selectedMemberId}
					onClose={handleCloseMemberProfile}
				/>
			) : null}
		</div>
	);
}
