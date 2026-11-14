import { useQuery } from "@tanstack/react-query";
import type { MediaExpressionKind } from "@/lib/expressions/expression-types";
import {
	fetchExpressionCategories,
	fetchExpressionMedia,
} from "@/services/expression/expression.service";

export const expressionKeys = {
	all: ["expressions"] as const,
	categories: (kind: MediaExpressionKind) =>
		[...expressionKeys.all, kind, "categories"] as const,
	media: (kind: MediaExpressionKind, query: string, limit: number) =>
		[...expressionKeys.all, kind, query, limit] as const,
};

export function useExpressionCategoriesQuery(kind: MediaExpressionKind, enabled: boolean) {
	return useQuery({
		queryKey: expressionKeys.categories(kind),
		queryFn: ({ signal }) => fetchExpressionCategories({ kind, signal }),
		enabled,
		staleTime: 5 * 60 * 1000,
		retry: false,
	});
}

export function useExpressionMediaQuery(params: {
	kind: MediaExpressionKind;
	query: string;
	limit?: number;
	enabled: boolean;
}) {
	const limit = params.limit ?? 24;
	const trimmed = params.query.trim();
	return useQuery({
		queryKey: expressionKeys.media(params.kind, trimmed, limit),
		queryFn: ({ signal }) =>
			fetchExpressionMedia({ kind: params.kind, query: trimmed, limit, signal }),
		enabled: params.enabled,
		staleTime: 2 * 60 * 1000,
		retry: false,
	});
}
