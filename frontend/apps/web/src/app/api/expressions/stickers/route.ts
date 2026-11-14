import { NextResponse, type NextRequest } from "next/server";
import { resolveExpressionMediaRequest } from "@/lib/expressions/expression-server";

export async function GET(request: NextRequest) {
	const { searchParams } = new URL(request.url);
	const mode = searchParams.get("mode") ?? "list";
	const query = (searchParams.get("q") ?? "").trim();
	const rawLimit = Number(searchParams.get("limit") ?? "24");
	const limit = Number.isFinite(rawLimit) ? Math.min(Math.max(rawLimit, 1), 36) : 24;

	const payload = await resolveExpressionMediaRequest({
		kind: "stickers",
		mode,
		query,
		limit,
	});
	return NextResponse.json(payload);
}

