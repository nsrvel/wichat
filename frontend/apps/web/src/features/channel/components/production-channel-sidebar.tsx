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
	useSensor,
	useSensors,
} from "@dnd-kit/core";
import {
	SortableContext,
	arrayMove,
	sortableKeyboardCoordinates,
	verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { type Channel, creatableChannelKindSchema } from "@wichat/contracts/chat/channel";
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
import { SortableCategoryItem } from "@/features/channel-category/components/sortable-category-item";
import { SortableChannelRow } from "@/features/channel/components/sortable-channel-list";
import { formatChannelDisplayName } from "@/features/channel/lib/channel-display";
import { useCategorySettings } from "@/features/channel-category/category-settings/category-settings-provider";
import { useChannelSettings } from "@/features/channel/channel-settings/channel-settings-provider";
import { DeleteChannelConfirmDialog } from "@/features/channel/channel-settings/delete-channel-confirm-dialog";
import { DeleteCategoryConfirmDialog } from "@/features/channel-category/category-settings/delete-category-confirm-dialog";
import { useCreateChannelDialog } from "@/features/channel/create-channel-dialog-provider";
import { useInviteDialog } from "@/features/invite/invite-dialog-provider";
import { translateError } from "@/integrations/i18n/translate-error";
import { railOrderMove } from "@/features/shell/lib/rail-order";
import { hasGroupPermission } from "@/lib/group-permissions";
import {
	CHANNEL_SIDEBAR_MIN_WIDTH,
	useResizableSidebarWidth,
} from "@/hooks/use-resizable-sidebar-width";
import { channelPath, quickViewPath, type QuickViewId } from "@/lib/app-routes";
import { cn } from "@/lib/utils";
import {
	useCreateGroupChannelMutation,
	useDeleteGroupChannelMutation,
	useReorderGroupChannelsMutation,
} from "@/services/channel/channel.mutation";
import {
	useDeleteGroupCategoryMutation,
	useReorderGroupCategoriesMutation,
} from "@/services/channel-category/channel-category.mutation";
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
	const { openCategorySettings } = useCategorySettings();
	const { openChannelSettings } = useChannelSettings();
	const { openInvite } = useInviteDialog();
	const createChannel = useCreateGroupChannelMutation(groupId);
	const deleteChannelMutation = useDeleteGroupChannelMutation(groupId);
	const deleteCategoryMutation = useDeleteGroupCategoryMutation(groupId);
	const [pendingDeleteChannel, setPendingDeleteChannel] = useState<{
		id: string;
		name: string;
	} | null>(null);
	const [pendingDeleteCategory, setPendingDeleteCategory] = useState<{
		id: string;
		name: string;
	} | null>(null);
	const channels = useQuery(channelQuery.groupChannelsQueryOpts(groupId));
	const { data: myGroups = [] } = useQuery(myGroupsQueryOpts());
	const currentGroup = myGroups.find((g) => g.id === groupId);
	const isGroupOwner = currentGroup?.role === "owner";
	const canManageChannels = isGroupOwner || hasGroupPermission(currentGroup, "manage_channels");

	const embeddedInPanel = controlledWidth !== undefined;
	const { width: internalWidth } = useResizableSidebarWidth(!embeddedInPanel);
	const width = controlledWidth ?? internalWidth;

	const [quickViewsOpen, setQuickViewsOpen] = useState(true);
	const [collapsedCategories, setCollapsedCategories] = useState<Record<string, boolean>>({});
	const [chatsOpen, setChatsOpen] = useState(true);
	const [activeMenuId, setActiveMenuId] = useState<string | null>(null);

	const channelList = channels.data?.channels ?? [];
	const categories = channels.data?.categories ?? [];

	type ActiveDragItem =
		| { type: "category"; categoryId: string; name: string; open: boolean }
		| { type: "channel"; channel: Channel };

	const [activeDrag, setActiveDrag] = useState<ActiveDragItem | null>(null);
	const [dragChannels, setDragChannels] = useState<Channel[] | null>(null);
	const displayedChannels = dragChannels ?? channelList;
	const uncategorizedChannels = displayedChannels.filter((c) => !c.category_id);

	const reorderCategories = useReorderGroupCategoriesMutation(groupId);
	const reorderChannels = useReorderGroupChannelsMutation(groupId);
	const collapseAllCategories = () => {
		setQuickViewsOpen(false);
		setChatsOpen(false);
		setCollapsedCategories(Object.fromEntries(categories.map((entry) => [entry.id, true])));
	};
	const moveChannelToCategory = (ch: Channel, targetCategoryId: string | null) => {
		const targetIds = channelList
			.filter((entry) => (entry.category_id || null) === targetCategoryId && entry.id !== ch.id)
			.map((entry) => entry.id);
		targetIds.push(ch.id);
		reorderChannels.mutate(
			{
				channelIds: targetIds,
				categoryId: targetCategoryId ?? undefined,
				sidebarSection: "channels",
			},
			{
				onSuccess: () => toast.success(t("channel.channel_context_menu.move_success")),
				onError: (err) => toast.error(translateError(t, err)),
			},
		);
	};
	const duplicateChannel = (ch: Channel) => {
		const parsedKind = creatableChannelKindSchema.safeParse(ch.type);
		createChannel.mutate(
			{
				name: `${ch.name}-copy`.slice(0, 64),
				type: parsedKind.success ? parsedKind.data : "text_only",
				category_id: ch.category_id ?? undefined,
				is_private: ch.is_private,
			},
			{
				onSuccess: () => toast.success(t("channel.channel_context_menu.duplicate_success")),
				onError: (err) => toast.error(translateError(t, err)),
			},
		);
	};
	const moveChannelInCategory = (ch: Channel, edge: "first" | "last") => {
		const catId = ch.category_id || null;
		const siblingIds = channelList
			.filter((entry) => (entry.category_id || null) === catId)
			.map((entry) => entry.id);
		const nextIds = railOrderMove(siblingIds, ch.id, edge);
		if (nextIds.every((id, i) => id === siblingIds[i])) {
			return;
		}
		reorderChannels.mutate(
			{
				channelIds: nextIds,
				categoryId: catId ?? undefined,
				sidebarSection: "channels",
			},
			{
				onError: (err) => toast.error(translateError(t, err)),
			},
		);
	};
	const renderChannelRow = (
		ch: Channel,
		index: number,
		siblingCount: number,
		categoryId?: string,
	) => (
		<SortableChannelRow
			key={ch.id}
			channel={ch}
			categories={categories}
			categoryId={categoryId}
			selected={!activeQuickView && ch.id === activeChannelId}
			disabled={!canManageChannels}
			canManage={canManageChannels}
			isFirstInCategory={index === 0}
			isLastInCategory={index >= siblingCount - 1}
			activeMenuId={activeMenuId}
			onActiveMenuIdChange={setActiveMenuId}
			onSelect={() => router.push(channelPath(groupId, ch.id))}
			onInvite={() =>
				openInvite({
					scope: "channel",
					groupId,
					groupName: currentGroup?.name ?? "",
					channelId: ch.id,
					channelName: ch.name,
				})
			}
			onMoveToCategory={(targetCategoryId) => moveChannelToCategory(ch, targetCategoryId)}
			onMoveToFirstInCategory={() => moveChannelInCategory(ch, "first")}
			onMoveToLastInCategory={() => moveChannelInCategory(ch, "last")}
			onEditChannel={() =>
				openChannelSettings({
					groupId,
					channelId: ch.id,
					channelName: ch.name,
					channelType: ch.type,
				})
			}
			onDuplicateChannel={() => duplicateChannel(ch)}
			onCreateChannel={() =>
				openCreateChannel({
					groupId,
					groupName: currentGroup?.name,
					defaultCategoryId: ch.category_id ?? undefined,
				})
			}
			onDeleteChannel={() => setPendingDeleteChannel({ id: ch.id, name: ch.name })}
		/>
	);
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
			{ type?: "category" | "channel"; categoryId?: string } | undefined;
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
			const currentBucket = activeChannel.category_id ?? "";
			if (currentBucket === targetCategoryId) {
				return prev;
			}

			const updated: Channel = {
				...activeChannel,
				sidebar_section: "channels",
				category_id: targetCategoryId,
			};
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
			const targetCategoryId = movedChannel.category_id ?? "";
			const targetChannels = snapshotChannels.filter(
				(c) => (c.category_id ?? "") === targetCategoryId,
			);
			const oldIndex = targetChannels.findIndex((c) => c.id === activeId);
			const overIndex = targetChannels.findIndex((c) => c.id === String(over.id));
			const reordered =
				oldIndex >= 0 && overIndex >= 0 && oldIndex !== overIndex
					? arrayMove(targetChannels, oldIndex, overIndex)
					: targetChannels;

			reorderChannels.mutate({
				channelIds: reordered.map((c) => c.id),
				categoryId: targetCategoryId || undefined,
				sidebarSection: "channels",
			});
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
							<SidebarCategoryContextMenu
								categoryName={t("shell.sidebar.quick_views")}
								open={quickViewsOpen}
								onToggle={() => setQuickViewsOpen((o) => !o)}
								onCollapseAll={collapseAllCategories}
							>
								<SidebarCategoryRow
									title={t("shell.sidebar.quick_views")}
									open={quickViewsOpen}
									onToggle={() => setQuickViewsOpen((o) => !o)}
									menuId="cat-quick-views"
									activeMenuId={activeMenuId}
									onActiveMenuIdChange={setActiveMenuId}
								/>
							</SidebarCategoryContextMenu>
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
							{uncategorizedChannels.length > 0 ? (
								<SortableContext
									items={uncategorizedChannels.map((ch) => ch.id)}
									strategy={verticalListSortingStrategy}
								>
									<div className={cn(sidebarNestedListClass, "mt-2")}>
										{uncategorizedChannels.map((ch, idx) =>
											renderChannelRow(ch, idx, uncategorizedChannels.length, ""),
										)}
									</div>
								</SortableContext>
							) : null}
							<SortableContext items={orderedCategoryIds} strategy={verticalListSortingStrategy}>
								{categories.map((cat, index) => {
									const isOpen = !collapsedCategories[cat.id];
									const catChannels = displayedChannels.filter((ch) => ch.category_id === cat.id);
									const toggleCategory = () =>
										setCollapsedCategories((prev) => ({
											...prev,
											[cat.id]: isOpen,
										}));
									return (
										<SortableCategoryItem
											key={cat.id}
											categoryId={cat.id}
											categoryName={cat.name}
											open={isOpen}
											canManage={canManageChannels}
											isFirst={index === 0}
											isLast={index === categories.length - 1}
											addLabel={t("channel_category.context_menu.create_channel")}
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
											onEditCategory={
												canManageChannels
													? () =>
															openCategorySettings({
																groupId,
																categoryId: cat.id,
																categoryName: cat.name,
																isPrivate: cat.is_private,
															})
													: undefined
											}
											onDeleteCategory={
												canManageChannels
													? () => setPendingDeleteCategory({ id: cat.id, name: cat.name })
													: undefined
											}
										>
											{isOpen ? (
												catChannels.length > 0 ? (
													<SortableContext
														items={catChannels.map((ch) => ch.id)}
														strategy={verticalListSortingStrategy}
													>
														<div className={sidebarNestedListClass}>
															{catChannels.map((ch, idx) =>
																renderChannelRow(ch, idx, catChannels.length, cat.id),
															)}
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
								onCollapseAll={collapseAllCategories}
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
			<DeleteChannelConfirmDialog
				channelName={pendingDeleteChannel?.name ?? ""}
				open={pendingDeleteChannel !== null}
				onOpenChange={(open) => {
					if (!open) {
						setPendingDeleteChannel(null);
					}
				}}
				confirmDisabled={deleteChannelMutation.isPending}
				onConfirm={() => {
					if (!pendingDeleteChannel) {
						return;
					}
					const targetChannel = pendingDeleteChannel;
					deleteChannelMutation.mutate(targetChannel.id, {
						onSuccess: () => {
							setPendingDeleteChannel(null);
							toast.success(
								t("channel.channel_settings.deleted_toast", { name: targetChannel.name }),
							);
							if (targetChannel.id === activeChannelId) {
								const nextChannel = channels.data?.channels.find((c) => c.id !== targetChannel.id);
								router.replace(
									nextChannel
										? channelPath(groupId, nextChannel.id)
										: `/channel/${encodeURIComponent(groupId)}`,
								);
							}
						},
						onError: (err) => {
							toast.error(translateError(t, err));
						},
					});
				}}
			/>
			<DeleteCategoryConfirmDialog
				categoryName={pendingDeleteCategory?.name ?? ""}
				open={pendingDeleteCategory !== null}
				onOpenChange={(open) => {
					if (!open) {
						setPendingDeleteCategory(null);
					}
				}}
				confirmDisabled={deleteCategoryMutation.isPending}
				onConfirm={() => {
					if (!pendingDeleteCategory) {
						return;
					}
					const targetCat = pendingDeleteCategory;
					deleteCategoryMutation.mutate(targetCat.id, {
						onSuccess: () => {
							setPendingDeleteCategory(null);
							toast.success(t("channel_category.settings.deleted_toast", { name: targetCat.name }));
						},
						onError: (err) => {
							toast.error(translateError(t, err));
						},
					});
				}}
			/>
		</div>
	);
}

export { CHANNEL_SIDEBAR_MIN_WIDTH };
