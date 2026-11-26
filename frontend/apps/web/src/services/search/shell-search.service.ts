import { FileSearchHitsSchema, MessageSearchHitsSchema } from "@wichat/contracts/chat/search";
import { api } from "@/integrations/api-client";

export async function searchGroupMessages(groupId: string, query: string, limit = 8) {
	const params = new URLSearchParams({ q: query, limit: String(limit) });
	return api.get(
		`/groups/${encodeURIComponent(groupId)}/search/messages?${params.toString()}`,
		MessageSearchHitsSchema,
	);
}

export async function searchGroupFiles(
	groupId: string,
	query: string,
	opts?: { limit?: number; imagesOnly?: boolean },
) {
	const params = new URLSearchParams({ q: query, limit: String(opts?.limit ?? 8) });
	if (opts?.imagesOnly) {
		params.set("images", "1");
	}
	return api.get(
		`/groups/${encodeURIComponent(groupId)}/search/files?${params.toString()}`,
		FileSearchHitsSchema,
	);
}