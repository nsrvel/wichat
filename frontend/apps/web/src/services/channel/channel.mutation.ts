import { useMutation, useQueryClient } from "@tanstack/react-query";
import type {
	Channel,
	ChannelsList,
	CreateGroupChannelRequest,
	UpdateGroupChannelRequest,
} from "@wichat/contracts/chat/channel";
import type {
	CategoryPermissionTargetType,
	UpsertCategoryPermissionOverrideRequest,
} from "@wichat/contracts/chat/channel-category";
import { Q_CHANNEL_PERMISSIONS, Q_GROUP_CHANNELS } from "@/services/channel/channel.key";
import {
	createGroupChannel,
	deleteChannelPermission,
	deleteGroupChannel,
	reorderGroupChannels,
	syncChannelPermissions,
	updateGroupChannel,
	upsertChannelPermission,
} from "@/services/channel/channel.service";

export function useCreateGroupChannelMutation(groupId: string) {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: (payload: CreateGroupChannelRequest) => createGroupChannel(groupId, payload),
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: [Q_GROUP_CHANNELS, groupId] });
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

export function useUpdateGroupChannelMutation(groupId: string) {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: ({
			channelId,
			payload,
		}: {
			channelId: string;
			payload: UpdateGroupChannelRequest;
		}) => updateGroupChannel(groupId, channelId, payload),
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: [Q_GROUP_CHANNELS, groupId] });
		},
	});
}

export function useDeleteGroupChannelMutation(groupId: string) {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: (channelId: string) => deleteGroupChannel(groupId, channelId),
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: [Q_GROUP_CHANNELS, groupId] });
		},
	});
}

export function useUpsertChannelPermissionMutation(groupId: string, channelId: string) {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: (payload: UpsertCategoryPermissionOverrideRequest) =>
			upsertChannelPermission(groupId, channelId, payload),
		onSuccess: () => {
			void queryClient.invalidateQueries({
				queryKey: [Q_CHANNEL_PERMISSIONS, groupId, channelId],
			});
		},
	});
}

export function useDeleteChannelPermissionMutation(groupId: string, channelId: string) {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: ({
			targetType,
			targetId,
		}: {
			targetType: CategoryPermissionTargetType;
			targetId: string;
		}) => deleteChannelPermission(groupId, channelId, targetType, targetId),
		onSuccess: () => {
			void queryClient.invalidateQueries({
				queryKey: [Q_CHANNEL_PERMISSIONS, groupId, channelId],
			});
		},
	});
}

export function useSyncChannelPermissionsMutation(groupId: string, channelId: string) {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: () => syncChannelPermissions(groupId, channelId),
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: [Q_GROUP_CHANNELS, groupId] });
			void queryClient.invalidateQueries({
				queryKey: [Q_CHANNEL_PERMISSIONS, groupId, channelId],
			});
		},
	});
}
