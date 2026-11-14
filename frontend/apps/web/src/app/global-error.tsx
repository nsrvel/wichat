"use client";

import { useEffect } from "react";

export default function GlobalErrorPage({
	error,
	reset,
}: {
	error: Error & { digest?: string };
	reset: () => void;
}) {
	useEffect(() => {
		console.error(error);
	}, [error]);

	return (
		<html lang="en" className="dark h-full antialiased">
			<body className="flex min-h-dvh w-full flex-col bg-[#0f1014] font-sans text-zinc-100">
				<main className="relative flex min-h-dvh w-full flex-col items-center justify-center overflow-hidden px-6 text-center">
					<div className="relative z-10 flex max-w-sm flex-col items-center">
						<p className="text-sm font-medium text-zinc-400">Something went wrong</p>
						<h1 className="mt-1 text-6xl font-bold leading-none tracking-tight tabular-nums sm:text-7xl">
							500
						</h1>
						<p className="mt-3.5 text-sm leading-relaxed text-zinc-400">
							We ran into an unexpected issue. Please try again in a moment.
						</p>
						<div className="mt-6 flex flex-wrap items-center justify-center gap-2.5">
							<button
								type="button"
								onClick={reset}
								className="inline-flex h-10 items-center justify-center rounded-xl bg-zinc-100 px-5 text-sm font-medium text-zinc-900 transition-colors hover:bg-zinc-200"
							>
								Try again
							</button>
							{/* eslint-disable-next-line @next/next/no-html-link-for-pages */}
							<a
								href="/"
								className="inline-flex h-10 items-center justify-center rounded-xl border border-zinc-800 bg-zinc-900/60 px-5 text-sm font-medium text-zinc-100 transition-colors hover:bg-zinc-800"
							>
								Back to home
							</a>
						</div>
					</div>
				</main>
			</body>
		</html>
	);
}
