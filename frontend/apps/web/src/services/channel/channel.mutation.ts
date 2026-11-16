import { useMutation, useQueryClient } from "@tanstack/react-query";
import type {
	Channel,
	ChannelCategory,
	ChannelsList,
	CreateChannelCategoryRequest,
	CreateGroupChannelRequest,
} from "@wichat/contracts/chat/channel";
import { Q_GROUP_CHANNELS } from "@/services/channel/channel.key";
import {
	createGroupCategory,
	createGroupChannel,
	reorderGroupCategories,
	reorderGroupChannels,
} from "@/services/channel/channel.service";

export function useCreateGroupCategoryMutation(groupId: string) {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: (payload: CreateChannelCategoryRequest) => createGroupCategory(groupId, payload),
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: [Q_GROUP_CHANNELS, groupId] });
		},
	});
}

export function useCreateGroupChannelMutation(groupId: string) {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: (payload: CreateGroupChannelRequest) => createGroupChannel(groupId, payload),
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: [Q_GROUP_CHANNELS, groupId] });
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

export function useReorderGroupChannelsMutation(groupId: string) {
	const queryClient = useQueryClient();
	const queryKey = [Q_GROUP_CHANNELS, groupId] as const;

	return useMutation({
		mutationFn: (payload: {
			channelIds: string[];
			categoryId?: string;
			sidebarSection?: "favorites" | "channels";
		}) => reorderGroupChannels(groupId, payload),
		onMutate: async ({ channelIds, categoryId, sidebarSection }) => {
			await queryClient.cancelQueries({ queryKey });
			const previous = queryClient.getQueryData<ChannelsList>(queryKey);
			if (previous) {
				const posById = new Map(channelIds.map((id, index) => [id, index]));
				const nextChannels: Channel[] = previous.channels.map((ch) => {
					const nextPos = posById.get(ch.id);
					if (nextPos === undefined) {
						return ch;
					}
					const nextSection = sidebarSection ?? ch.sidebar_section;
					return {
						...ch,
						position: nextPos,
						sidebar_section: nextSection,
						category_id: nextSection === "favorites" ? "" : (categoryId ?? ch.category_id),
					};
				});
				nextChannels.sort((a, b) => a.position - b.position);
				queryClient.setQueryData<ChannelsList>(queryKey, {
					...previous,
					channels: nextChannels,
				});
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
