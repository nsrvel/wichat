"use client";

import type { PresenceStatus } from "@wichat/contracts/user/presence";
import { memo, useCallback, useMemo, type ReactNode } from "react";
import { MessageActionBar } from "@/features/chat/components/message-action-bar";
import { MessageBody } from "@/features/chat/components/message-body";
import {
	MessageBubbleContextMenu,
	type MessageMenuActionProps,
} from "@/features/chat/components/message-context-menu";
import { MessageReactions } from "@/features/chat/components/message-reactions";
import { MessageRow } from "@/features/chat/components/message-row";
import type { ChannelMessage } from "@/features/chat/lib/channel-message";
import { hasInlineReplyBlock, stripInlineReplyBlocks } from "@/features/chat/lib/composer-markdown";

type ChannelMessageItemProps = {
	message: ChannelMessage;
	compact: boolean;
	authorAvatarUrl: string;
	authorNameColor?: string;
	authorIsOwner: boolean;
	authorPresenceStatus?: PresenceStatus;
	ownerLabel: string;
	highlighted: boolean;
	isEditing: boolean;
	canReact: boolean;
	canPin: boolean;
	canDelete: boolean;
	onReply: (message: ChannelMessage) => void;
	onStartEdit: (messageId: string) => void;
	onSaveEdit: (messageId: string, nextBody: string) => void;
	onCancelEdit: () => void;
	onTogglePin: (messageId: string, nextPinned: boolean) => void;
	onDelete: (messageId: string) => void;
	onToggleReaction: (messageId: string, emoji: string) => void;
	onJumpToMessage: (messageId: string) => void;
	onSelectMember?: (memberId: string) => void;
	resolveReactor?: (userId: string) => { displayName: string; avatarUrl?: string } | undefined;
};

export const ChannelMessageItem = memo(function ChannelMessageItem({
	message,
	compact,
	authorAvatarUrl,
	authorNameColor,
	authorIsOwner,
	authorPresenceStatus,
	ownerLabel,
	highlighted,
	isEditing,
	canReact,
	canPin,
	canDelete,
	onReply,
	onStartEdit,
	onSaveEdit,
	onCancelEdit,
	onTogglePin,
	onDelete,
	onToggleReaction,
	onJumpToMessage,
	onSelectMember,
	resolveReactor,
}: ChannelMessageItemProps) {
	const isOpt = message.id.startsWith("opt_");
	const isMine = Boolean(message.mine);
	const hasInlineReply = useMemo(() => hasInlineReplyBlock(message.bodyPlain), [message.bodyPlain]);

	const menuActions = useMemo<MessageMenuActionProps>(
		() => ({
			isMine,
			isPinned: Boolean(message.isPinned),
			canPin,
			canDelete,
			onReply: () => onReply(message),
			onEdit: isMine ? () => onStartEdit(message.id) : undefined,
			onTogglePin: canPin ? () => onTogglePin(message.id, !message.isPinned) : undefined,
			onCopy: () => {
				const plain = stripInlineReplyBlocks(message.bodyPlain);
				if (plain) {
					void navigator.clipboard.writeText(plain);
				}
			},
			onDelete: canDelete ? () => onDelete(message.id) : undefined,
		}),
		[canDelete, canPin, isMine, message, onDelete, onReply, onStartEdit, onTogglePin],
	);

	const renderBubbleWrapper = useCallback(
		(bubble: ReactNode) => (
			<MessageBubbleContextMenu {...menuActions}>{bubble}</MessageBubbleContextMenu>
		),
		[menuActions],
	);

	const handleSaveEdit = useCallback(
		(nextBody: string) => onSaveEdit(message.id, nextBody),
		[message.id, onSaveEdit],
	);

	const handleAuthorClick = useMemo(
		() => (onSelectMember && message.authorId ? () => onSelectMember(message.authorId) : undefined),
		[message.authorId, onSelectMember],
	);

	const handleReact = useCallback(
		(emoji: string) => onToggleReaction(message.id, emoji),
		[message.id, onToggleReaction],
	);

	return (
		<div style={{ contentVisibility: "auto", containIntrinsicSize: "auto 72px" }}>
			<MessageRow
				messageId={message.id}
				authorAvatarUrl={authorAvatarUrl}
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
				renderBubbleWrapper={!isOpt ? renderBubbleWrapper : undefined}
				actionBar={
					!isOpt ? (
						<MessageActionBar {...menuActions} canReact={canReact} onReact={handleReact} />
					) : undefined
				}
				reactions={
					message.reactions && message.reactions.length > 0 ? (
						<MessageReactions
							reactions={message.reactions}
							showAdd={canReact}
							resolveReactor={resolveReactor}
							onToggleReaction={handleReact}
						/>
					) : undefined
				}
			>
				<MessageBody text={message.bodyPlain} onReplyClick={onJumpToMessage} />
			</MessageRow>
		</div>
	);
});
