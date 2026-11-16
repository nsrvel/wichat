"use client";

import {
	DndContext,
	type DragEndEvent,
	KeyboardSensor,
	PointerSensor,
	closestCenter,
	useSensor,
	useSensors,
} from "@dnd-kit/core";
import {
	SortableContext,
	arrayMove,
	sortableKeyboardCoordinates,
	useSortable,
	verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import type { Channel } from "@wichat/contracts/chat/channel";
import { ChannelListIcon } from "@/components/layout/channel-list-icon";
import { SidebarChannelRow } from "@/components/layout/channel-sidebar/sidebar-channel-row";
import { sidebarNestedListClass } from "@/components/layout/channel-sidebar/sidebar-nav-styles";
import { formatChannelDisplayName } from "@/features/channel/lib/channel-display";

export type SortableChannelRowProps = {
	channel: Channel;
	categoryId?: string;
	selected: boolean;
	disabled: boolean;
	activeMenuId: string | null;
	onActiveMenuIdChange: (id: string | null) => void;
	onSelect: () => void;
};

export function SortableChannelRow({
	channel,
	categoryId,
	selected,
	disabled,
	activeMenuId,
	onActiveMenuIdChange,
	onSelect,
}: SortableChannelRowProps) {
	const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
		id: channel.id,
		disabled,
		data: { type: "channel", channelId: channel.id, categoryId: categoryId ?? "" },
	});

	const style = {
		transform: CSS.Translate.toString(transform),
		transition,
		opacity: isDragging ? 0.35 : undefined,
	};

	return (
		<div
			ref={setNodeRef}
			style={style}
			className={disabled ? undefined : "touch-none"}
			{...(disabled ? {} : attributes)}
			{...(disabled ? {} : listeners)}
		>
			<SidebarChannelRow
				menuId={`ch-${channel.id}`}
				activeMenuId={activeMenuId}
				onActiveMenuIdChange={onActiveMenuIdChange}
				icon={
					<ChannelListIcon
						type={channel.type}
						channelName={channel.name}
						imageUrl={channel.image_url}
					/>
				}
				label={formatChannelDisplayName(channel.name)}
				selected={selected}
				onSelect={onSelect}
			/>
		</div>
	);
}

type SortableChannelListProps = {
	channels: Channel[];
	activeChannelId: string;
	hasActiveQuickView: boolean;
	canReorder: boolean;
	activeMenuId: string | null;
	onActiveMenuIdChange: (id: string | null) => void;
	onSelectChannel: (channelId: string) => void;
	onReorder: (channelIds: string[]) => void;
};

export function SortableChannelList({
	channels,
	activeChannelId,
	hasActiveQuickView,
	canReorder,
	activeMenuId,
	onActiveMenuIdChange,
	onSelectChannel,
	onReorder,
}: SortableChannelListProps) {
	const sensors = useSensors(
		useSensor(PointerSensor, { activationConstraint: { distance: 8 } }),
		useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates }),
	);

	const onDragEnd = (event: DragEndEvent) => {
		const { active, over } = event;
		if (!over || active.id === over.id) {
			return;
		}
		const oldIndex = channels.findIndex((c) => c.id === active.id);
		const newIndex = channels.findIndex((c) => c.id === over.id);
		if (oldIndex < 0 || newIndex < 0) {
			return;
		}
		const reordered = arrayMove(channels, oldIndex, newIndex);
		onReorder(reordered.map((c) => c.id));
	};

	const list = (
		<div className={sidebarNestedListClass}>
			{channels.map((ch) => (
				<SortableChannelRow
					key={ch.id}
					channel={ch}
					selected={!hasActiveQuickView && ch.id === activeChannelId}
					disabled={!canReorder}
					activeMenuId={activeMenuId}
					onActiveMenuIdChange={onActiveMenuIdChange}
					onSelect={() => onSelectChannel(ch.id)}
				/>
			))}
		</div>
	);

	if (!canReorder) {
		return list;
	}

	return (
		<DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={onDragEnd}>
			<SortableContext items={channels.map((c) => c.id)} strategy={verticalListSortingStrategy}>
				{list}
			</SortableContext>
		</DndContext>
	);
}
