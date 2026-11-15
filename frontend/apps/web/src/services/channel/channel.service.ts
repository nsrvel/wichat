import { ChannelsListSchema, type ChannelsList } from "@wichat/contracts/chat/channel";
import { api } from "@/integrations/api-client";

export async function listGroupChannels(groupId: string): Promise<ChannelsList> {
	return api.get(`/groups/${encodeURIComponent(groupId)}/channels`, ChannelsListSchema);
}
