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
import Image from "next/image";
import { useTranslation } from "react-i18next";
import { IconRailTileButton } from "@/components/layout/shell-rail/icon-rail-tile-button";
import { IconRailWorkspaceRow } from "@/components/layout/shell-rail/icon-rail-row";
import { SHELL_RAIL_TILE_PX } from "@/components/layout/shell-rail/shell-rail-metrics";
import { GroupRailTileContextMenu } from "@/features/shell/components/group-rail-tile-context-menu";
import { groupInitials } from "@/lib/group-initials";
import { resolveGroupLogoUrl } from "@/lib/group-logo-url";
import { useReorderMyGroupsMutation } from "@/services/group/group.mutation";
import type { MyGroup } from "@/services/group/group.service";

const railLogoTileClass =
	"relative overflow-hidden border-0 bg-transparent p-0 shadow-none hover:opacity-90 select-none";

type ShellGroupRailListProps = {
	groups: MyGroup[];
	activeGroupId?: string;
	activeChannelId?: string;
	onNavigate: (group: MyGroup) => void;
	onPrefetchGroup?: (groupId: string) => void;
};

type SortableGroupTileProps = {
	group: MyGroup;
	active: boolean;
	activeGroupId?: string;
	orderedGroupIds: string[];
	reorderLabel: string;
	onNavigate: (group: MyGroup) => void;
	onPrefetchGroup?: (groupId: string) => void;
};

function SortableGroupTile({
	group,
	active,
	activeGroupId,
	orderedGroupIds,
	reorderLabel,
	onNavigate,
	onPrefetchGroup,
}: SortableGroupTileProps) {
	const logoSrc = resolveGroupLogoUrl(group.logo_url);
	const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
		id: group.id,
	});

	const style = {
		transform: CSS.Transform.toString(transform),
		transition,
		opacity: isDragging ? 0.85 : undefined,
	};

	return (
		<div
			ref={setNodeRef}
			style={style}
			className="touch-none"
			{...attributes}
			{...listeners}
			aria-roledescription={reorderLabel}
			onMouseEnter={() => onPrefetchGroup?.(group.id)}
		>
			<IconRailWorkspaceRow selected={active}>
				<GroupRailTileContextMenu
					group={group}
					activeGroupId={activeGroupId}
					orderedGroupIds={orderedGroupIds}
				>
					<IconRailTileButton
						variant="tile"
						className={logoSrc ? railLogoTileClass : undefined}
						aria-label={group.name}
						aria-current={active ? "true" : undefined}
						onClick={() => onNavigate(group)}
					>
						{logoSrc ? (
							<Image
								src={logoSrc}
								alt=""
								width={SHELL_RAIL_TILE_PX}
								height={SHELL_RAIL_TILE_PX}
								unoptimized
								draggable={false}
								className="pointer-events-none h-[var(--shell-rail-tile-size)] w-[var(--shell-rail-tile-size)] rounded-[var(--shell-rail-tile-radius)] object-cover"
							/>
						) : (
							<span className="text-sm font-semibold">{groupInitials(group.name)}</span>
						)}
					</IconRailTileButton>
				</GroupRailTileContextMenu>
			</IconRailWorkspaceRow>
		</div>
	);
}

export function ShellGroupRailList({
	groups,
	activeGroupId,
	onNavigate,
	onPrefetchGroup,
}: ShellGroupRailListProps) {
	const { t } = useTranslation();
	const reorderMutation = useReorderMyGroupsMutation();

	const sensors = useSensors(
		useSensor(PointerSensor, { activationConstraint: { distance: 8 } }),
		useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates }),
	);

	const onDragEnd = (event: DragEndEvent) => {
		const { active, over } = event;
		if (!over || active.id === over.id) {
			return;
		}
		const oldIndex = groups.findIndex((g) => g.id === active.id);
		const newIndex = groups.findIndex((g) => g.id === over.id);
		if (oldIndex < 0 || newIndex < 0) {
			return;
		}
		const reordered = arrayMove(groups, oldIndex, newIndex);
		reorderMutation.mutate(reordered.map((g) => g.id));
	};

	const reorderLabel = t("shell.group_rail_reorder");
	const orderedGroupIds = groups.map((g) => g.id);

	return (
		<DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={onDragEnd}>
			<SortableContext items={groups.map((g) => g.id)} strategy={verticalListSortingStrategy}>
				{groups.map((group) => (
					<SortableGroupTile
						key={group.id}
						group={group}
						active={group.id === activeGroupId}
						activeGroupId={activeGroupId}
						orderedGroupIds={orderedGroupIds}
						reorderLabel={reorderLabel}
						onNavigate={onNavigate}
						onPrefetchGroup={onPrefetchGroup}
					/>
				))}
			</SortableContext>
		</DndContext>
	);
}
