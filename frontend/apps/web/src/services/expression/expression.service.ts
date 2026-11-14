import type {
	MediaExpressionCategory,
	MediaExpressionEntry,
	MediaExpressionKind,
} from "@/lib/expressions/expression-types";

export type ExpressionMediaResponse = {
	configured: boolean;
	ok: boolean;
	provider: "klipy" | "giphy" | "none";
	items: MediaExpressionEntry[];
	categories: MediaExpressionCategory[];
};

export async function fetchExpressionMedia(params: {
	kind: MediaExpressionKind;
	query?: string;
	limit?: number;
	signal?: AbortSignal;
}): Promise<ExpressionMediaResponse> {
	const search = new URLSearchParams();
	if (params.query?.trim()) {
		search.set("q", params.query.trim());
	}
	if (params.limit) {
		search.set("limit", String(params.limit));
	}
	const res = await fetch(`/api/expressions/${params.kind}?${search.toString()}`, {
		signal: params.signal,
	});
	if (!res.ok) {
		throw new Error(`Failed to load ${params.kind}`);
	}
	return (await res.json()) as ExpressionMediaResponse;
}

export async function fetchExpressionCategories(params: {
	kind: MediaExpressionKind;
	signal?: AbortSignal;
}): Promise<ExpressionMediaResponse> {
	const res = await fetch(`/api/expressions/${params.kind}?mode=categories`, {
		signal: params.signal,
	});
	if (!res.ok) {
		throw new Error(`Failed to load ${params.kind} categories`);
	}
	return (await res.json()) as ExpressionMediaResponse;
}
