import { NextResponse, type NextRequest } from "next/server";
import { resolveExpressionMediaRequest } from "@/lib/expressions/expression-server";

export async function GET(request: NextRequest) {
	const { searchParams } = new URL(request.url);
	const mode = searchParams.get("mode") ?? "list";
	const query = (searchParams.get("q") ?? "").trim();
	const rawLimit = Number(searchParams.get("limit") ?? "20");
	const limit = Number.isFinite(rawLimit) ? Math.min(Math.max(rawLimit, 1), 36) : 20;

	const payload = await resolveExpressionMediaRequest({
		kind: "gifs",
		mode,
		query,
		limit,
	});
	return NextResponse.json(payload, {
		headers: {
			"Cache-Control": payload.ok
				? "public, max-age=120, stale-while-revalidate=300"
				: "no-store",
		},
	});
}
