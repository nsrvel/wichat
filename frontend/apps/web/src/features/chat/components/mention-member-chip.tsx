"use client";

import { type ReactNode, useCallback, useRef, useState } from "react";
import { GroupMemberProfileSurface } from "@/components/group/group-member-profile-surface";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { mentionChipClassName } from "@/features/chat/lib/mention-chip-styles";
import { cn } from "@/lib/utils";

const HOVER_OPEN_MS = 320;
const HOVER_CLOSE_MS = 160;

type MentionMemberChipProps = {
	userId: string;
	label: string;
	selfUserId?: string;
	groupId?: string;
	onOpenProfile?: (userId: string) => void;
	className?: string;
};

export function MentionMemberChip({
	userId,
	label,
	selfUserId,
	groupId,
	onOpenProfile,
	className,
}: MentionMemberChipProps) {
	const [previewOpen, setPreviewOpen] = useState(false);
	const [hoverZoneActive, setHoverZoneActive] = useState(false);
	const openTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
	const closeTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
	const canOpenProfile = Boolean(onOpenProfile);
	const showHoverPreview = Boolean(groupId);

	const clearOpenTimer = useCallback(() => {
		if (openTimerRef.current) {
			clearTimeout(openTimerRef.current);
			openTimerRef.current = null;
		}
	}, []);

	const clearCloseTimer = useCallback(() => {
		if (closeTimerRef.current) {
			clearTimeout(closeTimerRef.current);
			closeTimerRef.current = null;
		}
	}, []);

	const handleHoverZoneEnter = useCallback(() => {
		if (!showHoverPreview) {
			return;
		}
		setHoverZoneActive(true);
		clearCloseTimer();
		if (previewOpen) {
			return;
		}
		clearOpenTimer();
		openTimerRef.current = setTimeout(() => setPreviewOpen(true), HOVER_OPEN_MS);
	}, [clearCloseTimer, clearOpenTimer, previewOpen, showHoverPreview]);

	const handleHoverZoneLeave = useCallback(() => {
		setHoverZoneActive(false);
		clearOpenTimer();
		clearCloseTimer();
		closeTimerRef.current = setTimeout(() => setPreviewOpen(false), HOVER_CLOSE_MS);
	}, [clearCloseTimer, clearOpenTimer]);

	const closePreview = useCallback(() => {
		clearOpenTimer();
		clearCloseTimer();
		setPreviewOpen(false);
		setHoverZoneActive(false);
	}, [clearCloseTimer, clearOpenTimer]);

	const onClick = useCallback(
		(e: React.MouseEvent) => {
			if (!canOpenProfile) {
				return;
			}
			e.preventDefault();
			e.stopPropagation();
			closePreview();
			onOpenProfile?.(userId);
		},
		[canOpenProfile, closePreview, onOpenProfile, userId],
	);

	const chipClassName = cn(
		mentionChipClassName(userId, selfUserId),
		canOpenProfile && "cursor-pointer transition-opacity hover:opacity-90",
		className,
	);

	const cleanLabel = label.replace(/^@+/, "");
	const chipLabel = <span className={chipClassName}>@{cleanLabel}</span>;

	const fallbackUser = {
		userId,
		displayName: label,
	};

	if (!showHoverPreview) {
		return (
			<span
				role={canOpenProfile ? "button" : undefined}
				tabIndex={canOpenProfile ? 0 : undefined}
				onClick={onClick}
				onKeyDown={
					canOpenProfile
						? (e) => {
								if (e.key === "Enter" || e.key === " ") {
									e.preventDefault();
									onOpenProfile?.(userId);
								}
							}
						: undefined
				}
				className="inline"
			>
				{chipLabel}
			</span>
		);
	}

	return (
		<Popover open={previewOpen} modal={false}>
			<PopoverTrigger
				nativeButton={false}
				render={
					<span
						className="inline border-0 bg-transparent p-0"
						role={canOpenProfile ? "button" : undefined}
						tabIndex={canOpenProfile ? 0 : undefined}
					/>
				}
				onClick={onClick}
				onPointerEnter={handleHoverZoneEnter}
				onPointerLeave={handleHoverZoneLeave}
			>
				{chipLabel}
			</PopoverTrigger>
			<PopoverContent
				side="top"
				align="start"
				sideOffset={8}
				className="w-auto overflow-visible border-0 bg-transparent p-0 shadow-none"
			>
				<div onPointerEnter={handleHoverZoneEnter} onPointerLeave={handleHoverZoneLeave}>
					<GroupMemberProfileSurface
						groupId={groupId ?? ""}
						memberId={userId}
						fallbackUser={fallbackUser}
						presentation="popover"
						enabled={previewOpen || hoverZoneActive}
						onClose={closePreview}
					/>
				</div>
			</PopoverContent>
		</Popover>
	);
}

export function mentionLabelFromChildren(children: ReactNode): string {
	if (typeof children === "string") {
		return children.replace(/^@/, "");
	}
	if (Array.isArray(children)) {
		return children.map(mentionLabelFromChildren).join("").replace(/^@/, "");
	}
	if (children && typeof children === "object" && "props" in children) {
		const props = (children as { props?: { children?: ReactNode } }).props;
		return mentionLabelFromChildren(props?.children ?? "");
	}
	return "";
}
