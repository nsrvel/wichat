"use client";

import {
	type CollisionDetection,
	DndContext,
	type DragEndEvent,
	type DragOverEvent,
	DragOverlay,
	type DragStartEvent,
	KeyboardSensor,
	PointerSensor,
	closestCenter,
	pointerWithin,
	useDroppable,
	useSensor,
	useSensors,
} from "@dnd-kit/core";
import {
	SortableContext,
	arrayMove,
	sortableKeyboardCoordinates,
	verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import type { Channel } from "@wichat/contracts/chat/channel";
import { useQuery } from "@tanstack/react-query";
import { AtSign, Compass, FileText, Users } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { ChannelListIcon } from "@/components/layout/channel-list-icon";
import { SidebarCategoryContextMenu } from "@/components/layout/channel-sidebar/sidebar-category-context-menu";
import { SidebarCategoryEmptyHint } from "@/components/layout/channel-sidebar/sidebar-category-empty-hint";
import { SidebarCategoryRow } from "@/components/layout/channel-sidebar/sidebar-category-row";
import { SidebarChannelRow } from "@/components/layout/channel-sidebar/sidebar-channel-row";
import { SidebarNavItem } from "@/components/layout/channel-sidebar/sidebar-nav-item";
import {
	SIDEBAR_INSET_X,
	sidebarNestedListClass,
} from "@/components/layout/channel-sidebar/sidebar-nav-styles";
import { sidebarNavGlyphIconClass } from "@/components/layout/sidebar-list-icon";
import { SortableCategoryItem } from "@/features/channel/components/sortable-category-item";
import { SortableChannelRow } from "@/features/channel/components/sortable-channel-list";
import { formatChannelDisplayName } from "@/features/channel/lib/channel-display";
import { useCreateChannelDialog } from "@/features/channel/create-channel-dialog-provider";
import { railOrderMove } from "@/features/shell/lib/rail-order";
import { hasGroupPermission } from "@/lib/group-permissions";
import {
	CHANNEL_SIDEBAR_MIN_WIDTH,
	useResizableSidebarWidth,
} from "@/hooks/use-resizable-sidebar-width";
import { channelPath, quickViewPath, type QuickViewId } from "@/lib/app-routes";
import { cn } from "@/lib/utils";
import { partitionChannelsBySidebarSection } from "@/features/channel/lib/partition-group-channels";
import {
	useReorderGroupCategoriesMutation,
	useReorderGroupChannelsMutation,
} from "@/services/channel/channel.mutation";
import * as channelQuery from "@/services/channel/channel.query";
import { myGroupsQueryOpts } from "@/services/group/group.query";

type ProductionChannelSidebarProps = {
	groupId: string;
	activeChannelId: string;
	activeQuickView?: QuickViewId;
	width?: number;
	panelResizeHandle?: boolean;
};

export function ProductionChannelSidebar({
	groupId,
	activeChannelId,
	activeQuickView,
	width: controlledWidth,
	panelResizeHandle = false,
}: ProductionChannelSidebarProps) {
	const { t } = useTranslation();
	const router = useRouter();
	const { openCreateChannel } = useCreateChannelDialog();
	const channels = useQuery(channelQuery.groupChannelsQueryOpts(groupId));
	const { data: myGroups = [] } = useQuery(myGroupsQueryOpts());
	const currentGroup = myGroups.find((g) => g.id === groupId);
	const isGroupOwner = currentGroup?.role === "owner";
	const canManageChannels = isGroupOwner || hasGroupPermission(currentGroup, "manage_channels");

	const embeddedInPanel = controlledWidth !== undefined;
	const { width: internalWidth } = useResizableSidebarWidth(!embeddedInPanel);
	const width = controlledWidth ?? internalWidth;

	const [quickViewsOpen, setQuickViewsOpen] = useState(true);
	const [favoritesOpen, setFavoritesOpen] = useState(true);
	const [collapsedCategories, setCollapsedCategories] = useState<Record<string, boolean>>({});
	const [chatsOpen, setChatsOpen] = useState(true);
	const [activeMenuId, setActiveMenuId] = useState<string | null>(null);

	const FAVORITES_BUCKET_ID = "__favorites__";
	const { setNodeRef: setFavoritesDroppableRef } = useDroppable({
		id: FAVORITES_BUCKET_ID,
		disabled: !canManageChannels,
		data: { type: "favorites", categoryId: FAVORITES_BUCKET_ID },
	});

	const channelList = channels.data?.channels ?? [];
	const categories = channels.data?.categories ?? [];

	type ActiveDragItem =
		| { type: "category"; categoryId: string; name: string; open: boolean }
		| { type: "channel"; channel: Channel };

	const [activeDrag, setActiveDrag] = useState<ActiveDragItem | null>(null);
	const [dragChannels, setDragChannels] = useState<Channel[] | null>(null);
	const { favorites: favoriteChannels, listChannels } = partitionChannelsBySidebarSection(
		dragChannels ?? channelList,
	);

	const reorderCategories = useReorderGroupCategoriesMutation(groupId);
	const reorderChannels = useReorderGroupChannelsMutation(groupId);
	const sensors = useSensors(
		useSensor(PointerSensor, { activationConstraint: { distance: 8 } }),
		useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates }),
	);
	const orderedCategoryIds = categories.map((cat) => cat.id);

	const onDragStart = (event: DragStartEvent) => {
		const data = event.active.data.current as
			| { type?: "category"; categoryId?: string }
			| { type?: "channel"; channelId?: string }
			| undefined;
		if (data?.type === "category") {
			const cat = categories.find((c) => c.id === data.categoryId);
			if (cat) {
				setActiveDrag({
					type: "category",
					categoryId: cat.id,
					name: cat.name,
					open: !collapsedCategories[cat.id],
				});
			}
			return;
		}
		if (data?.type === "channel") {
			const ch = channelList.find((c) => c.id === data.channelId);
			if (ch) {
				setActiveDrag({ type: "channel", channel: ch });
				setDragChannels(channelList);
			}
		}
	};

	const onDragOver = (event: DragOverEvent) => {
		if (activeDrag?.type !== "channel") {
			return;
		}
		const { active, over } = event;
		if (!over || active.id === over.id) {
			return;
		}
		const overData = over.data.current as
			{ type?: "category" | "channel" | "favorites"; categoryId?: string } | undefined;
		const targetCategoryId = overData?.categoryId;
		if (!targetCategoryId) {
			return;
		}

		const activeId = String(active.id);
		const overId = String(over.id);

		setDragChannels((prev) => {
			if (!prev) {
				return prev;
			}
			const activeIndex = prev.findIndex((c) => c.id === activeId);
			if (activeIndex < 0) {
				return prev;
			}
			const activeChannel = prev[activeIndex];
			const currentBucket =
				activeChannel.sidebar_section === "favorites"
					? FAVORITES_BUCKET_ID
					: (activeChannel.category_id ?? "");
			if (currentBucket === targetCategoryId) {
				return prev;
			}

			const updated: Channel =
				targetCategoryId === FAVORITES_BUCKET_ID
					? { ...activeChannel, sidebar_section: "favorites", category_id: "" }
					: { ...activeChannel, sidebar_section: "channels", category_id: targetCategoryId };
			const withoutActive = prev.filter((c) => c.id !== activeId);

			if (overData?.type === "channel") {
				const overIndex = withoutActive.findIndex((c) => c.id === overId);
				if (overIndex >= 0) {
					const next = [...withoutActive];
					next.splice(overIndex, 0, updated);
					return next;
				}
			}

			return [...withoutActive, updated];
		});
	};

	const resetDragState = () => {
		setActiveDrag(null);
		setDragChannels(null);
	};

	const collisionDetection: CollisionDetection = (args) => {
		const activeType = args.active.data.current?.type as "category" | "channel" | undefined;

		if (activeType === "category") {
			const categoryContainers = args.droppableContainers.filter(
				(container) =>
					container.data.current?.type === "category" ||
					container.data.current?.type === "favorites",
			);
			const pointerCollisions = pointerWithin({
				...args,
				droppableContainers: categoryContainers,
			});
			if (pointerCollisions.length > 0) {
				return pointerCollisions;
			}
			return closestCenter({
				...args,
				droppableContainers: categoryContainers,
			});
		}

		if (activeType === "channel") {
			const channelContainers = args.droppableContainers.filter(
				(container) => container.data.current?.type === "channel",
			);
			const pointerChannelCollisions = pointerWithin({
				...args,
				droppableContainers: channelContainers,
			});
			if (pointerChannelCollisions.length > 0) {
				return pointerChannelCollisions;
			}

			const categoryContainers = args.droppableContainers.filter(
				(container) =>
					container.data.current?.type === "category" ||
					container.data.current?.type === "favorites",
			);
			const pointerCategoryCollisions = pointerWithin({
				...args,
				droppableContainers: categoryContainers,
			});
			if (pointerCategoryCollisions.length > 0) {
				return pointerCategoryCollisions;
			}

			return closestCenter({
				...args,
				droppableContainers: channelContainers,
			});
		}

		return closestCenter(args);
	};

	const onDragEnd = (event: DragEndEvent) => {
		const { active, over } = event;
		const currentDrag = activeDrag;
		const snapshotChannels = dragChannels ?? channelList;
		resetDragState();

		if (!over) {
			return;
		}

		if (currentDrag?.type === "category") {
			const overData = over.data.current as { categoryId?: string } | undefined;
			const overCategoryId = overData?.categoryId ?? String(over.id);
			if (active.id === overCategoryId) {
				return;
			}
			const oldIndex = categories.findIndex((c) => c.id === active.id);
			const newIndex = categories.findIndex((c) => c.id === overCategoryId);
			if (oldIndex < 0 || newIndex < 0) {
				return;
			}
			const reordered = arrayMove(categories, oldIndex, newIndex);
			reorderCategories.mutate(reordered.map((c) => c.id));
			return;
		}

		if (currentDrag?.type === "channel") {
			const activeId = String(active.id);
			const movedChannel = snapshotChannels.find((c) => c.id === activeId);
			if (!movedChannel) {
				return;
			}
			const isFavorites = movedChannel.sidebar_section === "favorites";
			const targetCategoryId = movedChannel.category_id ?? "";
			const targetChannels = isFavorites
				? snapshotChannels.filter((c) => c.sidebar_section === "favorites")
				: snapshotChannels.filter(
						(c) => c.sidebar_section !== "favorites" && c.category_id === targetCategoryId,
					);
			const oldIndex = targetChannels.findIndex((c) => c.id === activeId);
			const overIndex = targetChannels.findIndex((c) => c.id === String(over.id));
			const reordered =
				oldIndex >= 0 && overIndex >= 0 && oldIndex !== overIndex
					? arrayMove(targetChannels, oldIndex, overIndex)
					: targetChannels;

			reorderChannels.mutate(
				isFavorites
					? { channelIds: reordered.map((c) => c.id), sidebarSection: "favorites" }
					: {
							channelIds: reordered.map((c) => c.id),
							categoryId: targetCategoryId || undefined,
							sidebarSection: "channels",
						},
			);
		}
	};

	const moveCategory = (categoryId: string, where: "first" | "last") => {
		const next = railOrderMove(orderedCategoryIds, categoryId, where);
		if (next.join() === orderedCategoryIds.join()) {
			return;
		}
		reorderCategories.mutate(next);
	};

	return (
		<div
			className={cn(
				"relative flex h-full flex-col border-r border-sidebar-border bg-sidebar text-sidebar-foreground",
				embeddedInPanel ? "min-w-0 w-full" : "shrink-0",
			)}
			style={embeddedInPanel ? undefined : { width }}
		>
			<div className="flex min-h-0 flex-1 flex-col pt-1">
				<div className="thin-scroll flex-1 overflow-y-auto pb-20">
					<div className={SIDEBAR_INSET_X}>
						<div className="group/cat">
							<SidebarCategoryRow
								title={t("shell.sidebar.quick_views")}
								open={quickViewsOpen}
								onToggle={() => setQuickViewsOpen((o) => !o)}
								menuId="cat-quick-views"
								activeMenuId={activeMenuId}
								onActiveMenuIdChange={setActiveMenuId}
							/>
							{quickViewsOpen ? (
								<nav className={sidebarNestedListClass}>
									<SidebarNavItem
										menuId="qv-mentions"
										activeMenuId={activeMenuId}
										onActiveMenuIdChange={setActiveMenuId}
										selected={activeQuickView === "mentions"}
										onSelect={() => router.push(quickViewPath(groupId, "mentions"))}
										icon={<AtSign className={sidebarNavGlyphIconClass} strokeWidth={2} />}
										label={t("shell.sidebar.mentions")}
									/>
									<SidebarNavItem
										menuId="qv-discover"
										activeMenuId={activeMenuId}
										onActiveMenuIdChange={setActiveMenuId}
										selected={activeQuickView === "discover"}
										onSelect={() => router.push(quickViewPath(groupId, "discover"))}
										icon={<Compass className={sidebarNavGlyphIconClass} strokeWidth={2} />}
										label={t("shell.sidebar.discover")}
									/>
									<SidebarNavItem
										menuId="qv-drafts"
										activeMenuId={activeMenuId}
										onActiveMenuIdChange={setActiveMenuId}
										selected={activeQuickView === "drafts"}
										onSelect={() => router.push(quickViewPath(groupId, "drafts"))}
										icon={<FileText className={sidebarNavGlyphIconClass} strokeWidth={2} />}
										label={t("shell.sidebar.drafts")}
									/>
									{isGroupOwner ? (
										<SidebarNavItem
											menuId="qv-members"
											activeMenuId={activeMenuId}
											onActiveMenuIdChange={setActiveMenuId}
											selected={activeQuickView === "members"}
											onSelect={() => router.push(quickViewPath(groupId, "members"))}
											icon={<Users className={sidebarNavGlyphIconClass} strokeWidth={2} />}
											label={t("shell.sidebar.members")}
										/>
									) : null}
								</nav>
							) : null}
						</div>
						<DndContext
							sensors={sensors}
							collisionDetection={collisionDetection}
							onDragStart={onDragStart}
							onDragOver={onDragOver}
							onDragEnd={onDragEnd}
							onDragCancel={resetDragState}
						>
							<div ref={setFavoritesDroppableRef} className="group/cat mt-2">
								<SidebarCategoryRow
									title={t("shell.sidebar.favorites")}
									open={favoritesOpen}
									onToggle={() => setFavoritesOpen((o) => !o)}
									menuId="cat-favorites"
									activeMenuId={activeMenuId}
									onActiveMenuIdChange={setActiveMenuId}
								/>
								{favoritesOpen ? (
									favoriteChannels.length > 0 ? (
										<SortableContext
											items={favoriteChannels.map((ch) => ch.id)}
											strategy={verticalListSortingStrategy}
										>
											<div className={sidebarNestedListClass}>
												{favoriteChannels.map((ch) => (
													<SortableChannelRow
														key={ch.id}
														channel={ch}
														categoryId={FAVORITES_BUCKET_ID}
														selected={!activeQuickView && ch.id === activeChannelId}
														disabled={!canManageChannels}
														activeMenuId={activeMenuId}
														onActiveMenuIdChange={setActiveMenuId}
														onSelect={() => router.push(channelPath(groupId, ch.id))}
													/>
												))}
											</div>
										</SortableContext>
									) : (
										<SidebarCategoryEmptyHint>
											{t("shell.sidebar.empty_favorites")}
										</SidebarCategoryEmptyHint>
									)
								) : null}
							</div>
							<SortableContext items={orderedCategoryIds} strategy={verticalListSortingStrategy}>
								{categories.map((cat, index) => {
									const isOpen = !collapsedCategories[cat.id];
									const catChannels = listChannels.filter((ch) => ch.category_id === cat.id);
									const toggleCategory = () =>
										setCollapsedCategories((prev) => ({
											...prev,
											[cat.id]: isOpen,
										}));
									const collapseAllCategories = () =>
										setCollapsedCategories(
											Object.fromEntries(categories.map((entry) => [entry.id, true])),
										);
									return (
										<SortableCategoryItem
											key={cat.id}
											categoryId={cat.id}
											categoryName={cat.name}
											open={isOpen}
											canManage={canManageChannels}
											isFirst={index === 0}
											isLast={index === categories.length - 1}
											addLabel={t("channel.category_context_menu.create_channel")}
											onToggle={toggleCategory}
											onCollapseAll={collapseAllCategories}
											onAddClick={
												canManageChannels
													? () =>
															openCreateChannel({
																groupId,
																groupName: currentGroup?.name,
																defaultCategoryId: cat.id,
															})
													: undefined
											}
											onMoveToTop={() => moveCategory(cat.id, "first")}
											onMoveToBottom={() => moveCategory(cat.id, "last")}
										>
											{isOpen ? (
												catChannels.length > 0 ? (
													<SortableContext
														items={catChannels.map((ch) => ch.id)}
														strategy={verticalListSortingStrategy}
													>
														<div className={sidebarNestedListClass}>
															{catChannels.map((ch) => (
																<SortableChannelRow
																	key={ch.id}
																	channel={ch}
																	categoryId={cat.id}
																	selected={!activeQuickView && ch.id === activeChannelId}
																	disabled={!canManageChannels}
																	activeMenuId={activeMenuId}
																	onActiveMenuIdChange={setActiveMenuId}
																	onSelect={() => router.push(channelPath(groupId, ch.id))}
																/>
															))}
														</div>
													</SortableContext>
												) : channels.isPending ? null : (
													<SidebarCategoryEmptyHint>
														{t("shell.sidebar.empty_channels")}
													</SidebarCategoryEmptyHint>
												)
											) : null}
										</SortableCategoryItem>
									);
								})}
							</SortableContext>
							<DragOverlay dropAnimation={null}>
								{activeDrag?.type === "category" ? (
									<div className="rounded-md bg-sidebar/95 shadow-md ring-1 ring-border">
										<SidebarCategoryRow
											title={activeDrag.name}
											open={activeDrag.open}
											onToggle={() => {}}
										/>
									</div>
								) : activeDrag?.type === "channel" ? (
									<div className="rounded-md bg-sidebar/95 shadow-md ring-1 ring-border">
										<SidebarChannelRow
											menuId={`drag-${activeDrag.channel.id}`}
											activeMenuId={null}
											onActiveMenuIdChange={() => {}}
											icon={
												<ChannelListIcon
													type={activeDrag.channel.type}
													channelName={activeDrag.channel.name}
													imageUrl={activeDrag.channel.image_url}
												/>
											}
											label={formatChannelDisplayName(activeDrag.channel.name)}
											selected={!activeQuickView && activeDrag.channel.id === activeChannelId}
											onSelect={() => {}}
										/>
									</div>
								) : null}
							</DragOverlay>
						</DndContext>
						<div className="group/cat mt-2">
							<SidebarCategoryContextMenu
								categoryName={t("shell.sidebar.chats")}
								open={chatsOpen}
								onToggle={() => setChatsOpen((o) => !o)}
								onCollapseAll={() => {
									setChatsOpen(false);
									setCollapsedCategories(
										Object.fromEntries(categories.map((entry) => [entry.id, true])),
									);
								}}
							>
								<SidebarCategoryRow
									title={t("shell.sidebar.chats")}
									open={chatsOpen}
									onToggle={() => setChatsOpen((o) => !o)}
									onAddClick={() => toast.info(t("shell.sidebar.chats"))}
									addLabel={t("shell.sidebar.chats")}
								/>
							</SidebarCategoryContextMenu>
							{chatsOpen ? (
								<SidebarCategoryEmptyHint>
									{t("shell.sidebar.empty_chats")}
								</SidebarCategoryEmptyHint>
							) : null}
						</div>
					</div>
				</div>
			</div>
			{!panelResizeHandle ? null : (
				<div className="pointer-events-none absolute inset-y-0 right-0 w-px" aria-hidden />
			)}
		</div>
	);
}

export { CHANNEL_SIDEBAR_MIN_WIDTH };
