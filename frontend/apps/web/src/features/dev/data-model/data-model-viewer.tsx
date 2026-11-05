"use client";

import { useState } from "react";
import { MermaidDiagram } from "@/features/dev/data-model/mermaid-diagram";
import { cn } from "@/lib/utils";

type TabId = "implemented" | "planned";

type DataModelViewerProps = {
	implemented: string | null;
	planned: string | null;
	docsPath: string;
};

const TABS: { id: TabId; label: string; hint: string }[] = [
	{
		id: "implemented",
		label: "Implemented",
		hint: "auth.* — ms-user migrations",
	},
	{
		id: "planned",
		label: "Planned",
		hint: "channels, chat, audit — draft",
	},
];

export function DataModelViewer({ implemented, planned, docsPath }: DataModelViewerProps) {
	const [tab, setTab] = useState<TabId>("implemented");
	const chart = tab === "implemented" ? implemented : planned;
	const missing = chart == null;

	return (
		<div className="mx-auto flex w-full max-w-5xl flex-col gap-6 px-4 py-8 sm:px-6">
			<header className="space-y-1">
				<h1 className="text-xl font-semibold tracking-tight">Data model (ERD)</h1>
				<p className="text-sm text-muted-foreground">
					Preview Mermaid dari{" "}
					<code className="rounded bg-muted px-1 py-0.5 text-xs">{docsPath}</code>. DBML untuk
					detail kolom: import di{" "}
					<a
						className="text-foreground underline-offset-2 hover:underline"
						href="https://dbdiagram.io"
						target="_blank"
						rel="noreferrer"
					>
						dbdiagram.io
					</a>
					.
				</p>
			</header>

			<div className="flex flex-wrap gap-2 border-b border-border pb-2">
				{TABS.map(({ id, label, hint }) => (
					<button
						key={id}
						type="button"
						onClick={() => setTab(id)}
						className={cn(
							"rounded-lg px-3 py-1.5 text-left text-sm transition-colors",
							tab === id
								? "bg-muted font-medium text-foreground"
								: "text-muted-foreground hover:bg-muted/60 hover:text-foreground",
						)}
					>
						<span>{label}</span>
						<span className="mt-0.5 block text-[11px] font-normal opacity-80">{hint}</span>
					</button>
				))}
			</div>

			{missing ? (
				<p className="text-sm text-muted-foreground">
					Diagram file tidak ketemu. Jalankan web dari monorepo root dan pastikan folder{" "}
					<code className="text-xs">{docsPath}</code> ada.
				</p>
			) : (
				<MermaidDiagram chart={chart} />
			)}
		</div>
	);
}
