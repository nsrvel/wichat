export function railOrderMove(
	ids: string[],
	targetId: string,
	where: "first" | "last",
): string[] {
	if (!ids.includes(targetId)) {
		return ids;
	}
	const rest = ids.filter((id) => id !== targetId);
	if (where === "first") {
		return [targetId, ...rest];
	}
	return [...rest, targetId];
}
