"use client";

import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import type { ReactNode } from "react";
import { SidebarCategoryContextMenu } from "@/components/layout/channel-sidebar/sidebar-category-context-menu";
import { SidebarCategoryRow } from "@/components/layout/channel-sidebar/sidebar-category-row";

type SortableCategoryItemProps = {
	categoryId: string;
	categoryName: string;
	open: boolean;
	canManage: boolean;
	isFirst: boolean;
	isLast: boolean;
	addLabel: string;
	onToggle: () => void;
	onCollapseAll: () => void;
	onAddClick?: () => void;
	onMoveToTop: () => void;
	onMoveToBottom: () => void;
	children: ReactNode;
};

export function SortableCategoryItem({
	categoryId,
	categoryName,
	open,
	canManage,
	isFirst,
	isLast,
	addLabel,
	onToggle,
	onCollapseAll,
	onAddClick,
	onMoveToTop,
	onMoveToBottom,
	children,
}: SortableCategoryItemProps) {
	const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
		id: categoryId,
		disabled: !canManage,
		data: { type: "category", categoryId },
	});

	const style = {
		transform: CSS.Translate.toString(transform),
		transition,
		opacity: isDragging ? 0.35 : undefined,
	};

	return (
		<div ref={setNodeRef} style={style} className="group/cat mt-2">
			<div
				className={canManage ? "touch-none" : undefined}
				{...(canManage ? attributes : {})}
				{...(canManage ? listeners : {})}
			>
				<SidebarCategoryContextMenu
					categoryName={categoryName}
					open={open}
					onToggle={onToggle}
					onCollapseAll={onCollapseAll}
					canManage={canManage}
					isFirst={isFirst}
					isLast={isLast}
					onMoveToTop={onMoveToTop}
					onMoveToBottom={onMoveToBottom}
				>
					<SidebarCategoryRow
						title={categoryName}
						open={open}
						onToggle={onToggle}
						onAddClick={onAddClick}
						addLabel={addLabel}
					/>
				</SidebarCategoryContextMenu>
			</div>
			{children}
		</div>
	);
}
