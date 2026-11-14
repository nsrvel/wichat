"use client";

import { useState } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import type { StatusScreenVariant } from "@/components/status/status-illustrations";
import { StatusScreen } from "@/components/status/status-screen";
import { cn } from "@/lib/utils";

const VARIANTS: { id: StatusScreenVariant; label: string }[] = [
	{ id: "not-found", label: "404 Not Found" },
	{ id: "forbidden", label: "403 Forbidden" },
	{ id: "unauthorized", label: "401 Unauthorized" },
	{ id: "error", label: "500 Error" },
	{ id: "maintenance", label: "503 Maintenance" },
];

export function StatusPagesPreview() {
	const { t } = useTranslation();
	const [variant, setVariant] = useState<StatusScreenVariant>("not-found");

	return (
		<div className="fixed inset-0 z-[100] overflow-y-auto bg-background">
			<div className="fixed left-1/2 top-4 z-20 flex max-w-[calc(100vw-2rem)] -translate-x-1/2 items-center gap-1 overflow-x-auto rounded-xl border border-border bg-card/90 p-1 shadow-elevated backdrop-blur-md">
				<span className="hidden px-2.5 text-[11px] font-medium text-muted-foreground sm:inline">
					{t("status_pages.preview.label")}
				</span>
				{VARIANTS.map((item) => {
					const active = item.id === variant;
					return (
						<button
							key={item.id}
							type="button"
							onClick={() => setVariant(item.id)}
							className={cn(
								"shrink-0 rounded-lg px-3 py-1.5 text-xs font-medium transition-colors",
								active
									? "bg-primary text-primary-foreground"
									: "text-muted-foreground hover:bg-muted hover:text-foreground",
							)}
						>
							{item.label}
						</button>
					);
				})}
			</div>

			<StatusScreen
				variant={variant}
				onRetry={() => toast.info("Retry action triggered")}
			/>
		</div>
	);
}
