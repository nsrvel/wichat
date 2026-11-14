"use client";

import { useEffect } from "react";
import { StatusScreen } from "@/components/status/status-screen";

export default function AppErrorPage({
	error,
	reset,
}: {
	error: Error & { digest?: string };
	reset: () => void;
}) {
	useEffect(() => {
		console.error(error);
	}, [error]);

	return <StatusScreen variant="error" onRetry={reset} />;
}
