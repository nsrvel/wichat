"use client";

import { useTheme } from "next-themes";
import { useEffect, useId, useRef, useState } from "react";
import { cn } from "@/lib/utils";

type MermaidDiagramProps = {
	chart: string;
	className?: string;
};

export function MermaidDiagram({ chart, className }: MermaidDiagramProps) {
	const containerRef = useRef<HTMLDivElement>(null);
	const reactId = useId().replace(/:/g, "");
	const { resolvedTheme } = useTheme();
	const [error, setError] = useState<string | null>(null);

	useEffect(() => {
		let cancelled = false;

		async function render() {
			setError(null);
			const el = containerRef.current;
			if (!el) return;

			try {
				const mermaid = (await import("mermaid")).default;
				mermaid.initialize({
					startOnLoad: false,
					securityLevel: "strict",
					theme: resolvedTheme === "dark" ? "dark" : "default",
					er: { useMaxWidth: true },
				});

				const renderId = `wichat-erd-${reactId}`;
				const { svg, bindFunctions } = await mermaid.render(renderId, chart);
				if (cancelled) return;
				el.innerHTML = svg;
				bindFunctions?.(el);
			} catch (e) {
				if (!cancelled) {
					setError(e instanceof Error ? e.message : "Failed to render diagram");
				}
			}
		}

		void render();
		return () => {
			cancelled = true;
		};
	}, [chart, reactId, resolvedTheme]);

	if (error) {
		return (
			<pre className="rounded-lg border border-destructive/40 bg-destructive/10 p-4 text-sm text-destructive">
				{error}
			</pre>
		);
	}

	return (
		<div
			ref={containerRef}
			className={cn(
				"thin-scroll min-h-[12rem] overflow-auto rounded-lg border border-border bg-card p-4 [&_svg]:mx-auto [&_svg]:max-w-full",
				className,
			)}
		/>
	);
}
