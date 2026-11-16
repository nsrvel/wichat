import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { ChannelsList } from "@wichat/contracts/chat/channel";
import type {
	CategoryPermissionTargetType,
	ChannelCategory,
	CreateChannelCategoryRequest,
	UpdateChannelCategoryRequest,
	UpsertCategoryPermissionOverrideRequest,
} from "@wichat/contracts/chat/channel-category";
import { Q_GROUP_CHANNELS } from "@/services/channel/channel.key";
import { Q_CATEGORY_PERMISSIONS } from "@/services/channel-category/channel-category.key";
import {
	createGroupCategory,
	deleteCategoryPermission,
	deleteGroupCategory,
	reorderGroupCategories,
	updateGroupCategory,
	upsertCategoryPermission,
} from "@/services/channel-category/channel-category.service";

export function useCreateGroupCategoryMutation(groupId: string) {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: (payload: CreateChannelCategoryRequest) => createGroupCategory(groupId, payload),
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: [Q_GROUP_CHANNELS, groupId] });
		},
	});
}

export function useUpdateGroupCategoryMutation(groupId: string) {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: ({
			categoryId,
			payload,
		}: {
			categoryId: string;
			payload: UpdateChannelCategoryRequest;
		}) => updateGroupCategory(groupId, categoryId, payload),
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: [Q_GROUP_CHANNELS, groupId] });
		},
	});
}

export function useDeleteGroupCategoryMutation(groupId: string) {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: (categoryId: string) => deleteGroupCategory(groupId, categoryId),
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: [Q_GROUP_CHANNELS, groupId] });
		},
	});
}

export function useUpsertCategoryPermissionMutation(groupId: string, categoryId: string) {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: (payload: UpsertCategoryPermissionOverrideRequest) =>
			upsertCategoryPermission(groupId, categoryId, payload),
		onSuccess: () => {
			void queryClient.invalidateQueries({
				queryKey: [Q_CATEGORY_PERMISSIONS, groupId, categoryId],
			});
		},
	});
}

export function useDeleteCategoryPermissionMutation(groupId: string, categoryId: string) {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: ({
			targetType,
			targetId,
		}: {
			targetType: CategoryPermissionTargetType;
			targetId: string;
		}) => deleteCategoryPermission(groupId, categoryId, targetType, targetId),
		onSuccess: () => {
			void queryClient.invalidateQueries({
				queryKey: [Q_CATEGORY_PERMISSIONS, groupId, categoryId],
			});
		},
	});
}

export function useReorderGroupCategoriesMutation(groupId: string) {
	const queryClient = useQueryClient();
	const queryKey = [Q_GROUP_CHANNELS, groupId] as const;

	return useMutation({
		mutationFn: (categoryIds: string[]) => reorderGroupCategories(groupId, categoryIds),
		onMutate: async (categoryIds) => {
			await queryClient.cancelQueries({ queryKey });
			const previous = queryClient.getQueryData<ChannelsList>(queryKey);
			if (previous) {
				const byId = new Map(previous.categories.map((c) => [c.id, c]));
				const reordered = categoryIds
					.map((id, index) => {
						const cat = byId.get(id);
						return cat ? { ...cat, position: index } : null;
					})
					.filter((c): c is ChannelCategory => c !== null);
				if (reordered.length === previous.categories.length) {
					queryClient.setQueryData<ChannelsList>(queryKey, {
						...previous,
						categories: reordered,
					});
				}
			}
			return { previous };
		},
		onError: (_err, _ids, ctx) => {
			if (ctx?.previous) {
				queryClient.setQueryData(queryKey, ctx.previous);
			}
		},
		onSettled: () => {
			void queryClient.invalidateQueries({ queryKey });
		},
	});
}
