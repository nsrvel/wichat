"use client";

import { ChevronDown, ChevronRight } from "lucide-react";
import { useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import type { GroupActivityLog } from "@wichat/contracts/group/audit-log";
import { UserAvatar } from "@/components/ui/user-avatar";
import { userInitials } from "@/lib/user-initials";
import { cn } from "@/lib/utils";

function formatAuditTimestamp(
	iso: string,
	t: (key: string, options?: Record<string, string>) => string,
): string {
	const date = new Date(iso);
	if (Number.isNaN(date.getTime())) {
		return iso;
	}

	const time = new Intl.DateTimeFormat(undefined, { timeStyle: "short" }).format(date);
	const now = new Date();
	const startOfToday = new Date(now.getFullYear(), now.getMonth(), now.getDate());
	const startOfTarget = new Date(date.getFullYear(), date.getMonth(), date.getDate());
	const diffDays = Math.round((startOfToday.getTime() - startOfTarget.getTime()) / 86_400_000);

	if (diffDays === 0) {
		return t("group_settings.audit_logs.time_today", { time });
	}
	if (diffDays === 1) {
		return t("group_settings.audit_logs.time_yesterday", { time });
	}

	const dateLabel = new Intl.DateTimeFormat(undefined, {
		month: "short",
		day: "numeric",
		year: date.getFullYear() !== now.getFullYear() ? "numeric" : undefined,
	}).format(date);
	return `${dateLabel}, ${time}`;
}

function formatChangeValue(val: unknown): string {
	if (val === null || val === undefined || val === "") {
		return "—";
	}
	if (typeof val === "boolean") {
		return val ? "true" : "false";
	}
	return String(val);
}

export function GroupSettingsAuditLogRow({
	log,
	onActorClick,
}: {
	log: GroupActivityLog;
	onActorClick?: (log: GroupActivityLog) => void;
}) {
	const { t } = useTranslation();
	const [expanded, setExpanded] = useState(false);

	const actorName =
		log.actor_display_name || log.actor_username || t("group_settings.audit_logs.unknown_actor");
	const actionVerb = t(`group_settings.audit_logs.actions.${log.action}`, {
		defaultValue: log.summary || log.action,
	});
	const targetLabel = log.target_name || (log.target_type === "invite" ? log.target_id : "");

	const detailLines = useMemo(() => {
		const lines: string[] = [];

		for (const [field, diff] of Object.entries(log.changes ?? {})) {
			if (diff.old !== undefined && diff.old !== "") {
				lines.push(
					t("group_settings.audit_logs.field_changed", {
						field,
						old: formatChangeValue(diff.old),
						new: formatChangeValue(diff.new),
					}),
				);
			} else {
				lines.push(
					t("group_settings.audit_logs.field_set", {
						field,
						new: formatChangeValue(diff.new),
					}),
				);
			}
		}

		const meta = log.metadata ?? {};
		if (typeof meta.reason === "string" && meta.reason.trim()) {
			lines.push(`${t("group_settings.audit_logs.reason_label")}: ${meta.reason.trim()}`);
		}
		if (typeof meta.duration_seconds === "number") {
			lines.push(
				`${t("group_settings.audit_logs.duration_label")}: ${
					meta.duration_seconds === 0
						? t("group_settings.audit_logs.duration_permanent")
						: `${meta.duration_seconds}s`
				}`,
			);
		}
		if (typeof meta.code === "string" && meta.code.trim()) {
			lines.push(`${t("group_settings.audit_logs.invite_code_label")}: ${meta.code.trim()}`);
		}
		if (typeof meta.via === "string" && meta.via.trim()) {
			lines.push(`${t("group_settings.audit_logs.via_label")}: ${meta.via.trim()}`);
		}
		if (typeof meta.role_name === "string" && meta.role_name.trim() && !targetLabel) {
			lines.push(`${t("group_settings.audit_logs.role_label")}: ${meta.role_name.trim()}`);
		}

		return lines;
	}, [log.changes, log.metadata, t, targetLabel]);

	const hasDetails = detailLines.length > 0;

	return (
		<div
			data-testid={`audit-log-row-${log.id}`}
			className="overflow-hidden rounded-xl border border-border/50 bg-field/40 transition-colors hover:bg-field/60"
		>
			<div className="flex w-full items-center gap-3 px-4 py-3">
				{onActorClick && log.actor_id ? (
					<button
						type="button"
						onClick={() => onActorClick(log)}
						aria-label={actorName}
						className="shrink-0 cursor-pointer rounded-full transition-opacity hover:opacity-85"
					>
						<UserAvatar
							src={log.actor_avatar_url || null}
							alt={actorName}
							initials={userInitials(actorName, log.actor_username)}
							size="sm"
						/>
					</button>
				) : (
					<UserAvatar
						src={log.actor_avatar_url || null}
						alt={actorName}
						initials={userInitials(actorName, log.actor_username)}
						size="sm"
					/>
				)}

				<button
					type="button"
					data-testid="audit-log-expand-trigger"
					disabled={!hasDetails}
					aria-expanded={hasDetails ? expanded : undefined}
					onClick={() => {
						if (hasDetails) {
							setExpanded((prev) => !prev);
						}
					}}
					className={cn(
						"flex min-w-0 flex-1 items-center gap-3 text-left",
						hasDetails ? "cursor-pointer" : "cursor-default",
					)}
				>
					<div className="min-w-0 flex-1">
						<p className="truncate text-sm font-medium text-foreground">
							{actorName} {actionVerb}
							{targetLabel ? ` ${targetLabel}` : ""}
						</p>
						<p className="mt-0.5 text-xs text-muted-foreground">
							{formatAuditTimestamp(log.created_at, t)}
						</p>
					</div>

					{hasDetails ? (
						expanded ? (
							<ChevronDown className="size-4 shrink-0 text-muted-foreground" aria-hidden />
						) : (
							<ChevronRight className="size-4 shrink-0 text-muted-foreground" aria-hidden />
						)
					) : null}
				</button>
			</div>

			{hasDetails && expanded ? (
				<div className="border-t border-border/40 bg-background/45 px-4 py-3 pl-[3.75rem]">
					<ul className="flex flex-col gap-2">
						{detailLines.map((line, index) => (
							<li
								key={`${index}-${line}`}
								className="flex items-baseline gap-2.5 text-[13px] leading-relaxed"
							>
								<span className="font-mono text-xs font-semibold text-amber-400">
									{String(index + 1).padStart(2, "0")}
								</span>
								<span className="text-muted-foreground">—</span>
								<span className="text-foreground">{line}</span>
							</li>
						))}
					</ul>
				</div>
			) : null}
		</div>
	);
}
