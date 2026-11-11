"use client";

import { Send, X } from "lucide-react";
import { useEffect, useId, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { authInputClass } from "@/features/auth/components/auth-shell";
import {
	type RecipientChip,
	type UserRecipientChip,
	parseEmailChip,
} from "@/features/invite/lib/recipient-chip";
import { userInitials } from "@/features/invite/lib/user-initials";
import { UserSearchHitSchema } from "@wichat/contracts";
import { cn } from "@/lib/utils";
import type { z } from "zod";

type SearchHit = z.infer<typeof UserSearchHitSchema>;

type InviteRecipientFieldProps = {
	chips: RecipientChip[];
	onChangeChips: (next: RecipientChip[]) => void;
	searchHits: SearchHit[];
	searchLoading: boolean;
	query: string;
	onQueryChange: (value: string) => void;
	onPickUser: (hit: SearchHit) => void;
	disabled?: boolean;
};

export function InviteRecipientField({
	chips,
	onChangeChips,
	searchHits,
	searchLoading,
	query,
	onQueryChange,
	onPickUser,
	disabled,
}: InviteRecipientFieldProps) {
	const { t } = useTranslation();
	const inputId = useId();
	const listId = useId();
	const [open, setOpen] = useState(false);
	const wrapRef = useRef<HTMLDivElement>(null);

	useEffect(() => {
		const onDoc = (e: MouseEvent) => {
			if (!wrapRef.current?.contains(e.target as Node)) {
				setOpen(false);
			}
		};
		document.addEventListener("mousedown", onDoc);
		return () => document.removeEventListener("mousedown", onDoc);
	}, []);

	const removeChip = (chip: RecipientChip) => {
		onChangeChips(chips.filter((c) => c !== chip));
	};

	const tryCommitEmail = () => {
		const emailChip = parseEmailChip(query);
		if (!emailChip) {
			return false;
		}
		if (chips.some((c) => c.kind === "email" && c.email === emailChip.email)) {
			onQueryChange("");
			return true;
		}
		onChangeChips([...chips, emailChip]);
		onQueryChange("");
		setOpen(false);
		return true;
	};

	const onKeyDown: React.KeyboardEventHandler<HTMLInputElement> = (e) => {
		if (e.key === "Enter" || e.key === ",") {
			e.preventDefault();
			if (searchHits[0] && open) {
				onPickUser(searchHits[0]);
				return;
			}
			tryCommitEmail();
		}
		if (e.key === "Backspace" && query === "" && chips.length > 0) {
			onChangeChips(chips.slice(0, -1));
		}
	};

	const pick = (hit: SearchHit) => {
		const chip: UserRecipientChip = {
			kind: "user",
			id: hit.user_id,
			label: hit.display_name || hit.email,
			email: hit.email,
		};
		if (chips.some((c) => c.kind === "user" && c.id === chip.id)) {
			onQueryChange("");
			setOpen(false);
			return;
		}
		onChangeChips([...chips, chip]);
		onQueryChange("");
		setOpen(false);
	};

	const trimmedQuery = query.trim();
	const showDropdown = open && trimmedQuery.length > 0;
	const pendingEmail = parseEmailChip(query);
	const showEmailHint =
		showDropdown && !searchLoading && searchHits.length === 0 && pendingEmail !== null;
	const emailOnlySuggest = showEmailHint;
	const hasChips = chips.length > 0;

	const fieldShellClass = cn(
		authInputClass,
		"flex flex-wrap items-center gap-2",
		hasChips
			? [
					"h-auto min-h-10 py-2 sm:h-auto sm:min-h-11 sm:py-2.5",
					"max-h-[880px]:h-auto max-h-[880px]:min-h-9 max-h-[880px]:sm:h-auto max-h-[880px]:sm:min-h-10",
				]
			: "h-10 min-h-10 py-0 sm:h-11 sm:min-h-11 max-h-[880px]:h-9 max-h-[880px]:sm:h-10",
	);

	const chipClass =
		"inline-flex h-8 max-w-[min(100%,18rem)] shrink-0 items-center gap-1 rounded-lg bg-background pl-2.5 pr-1 text-xs font-medium text-foreground ring-1 ring-border/55";

	return (
		<div ref={wrapRef} className="relative z-[80] min-w-0 flex-1">
			<div
				className={fieldShellClass}
				onMouseDown={(e) => {
					if (disabled) {
						return;
					}
					const target = e.target as HTMLElement;
					if (target.closest("button")) {
						return;
					}
					const el = document.getElementById(inputId);
					if (el instanceof HTMLInputElement && target !== el) {
						e.preventDefault();
						el.focus();
					}
				}}
			>
				{chips.map((chip) => (
					<span key={chip.kind === "user" ? chip.id : chip.email} className={chipClass}>
						<span className="min-w-0 truncate leading-none">{chip.label}</span>
						<button
							type="button"
							className="flex h-6 w-6 shrink-0 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-muted/90 hover:text-foreground"
							aria-label={t("invite.remove_recipient")}
							disabled={disabled}
							onClick={() => removeChip(chip)}
						>
							<X className="h-3.5 w-3.5" aria-hidden />
						</button>
					</span>
				))}
				<Input
					id={inputId}
					value={query}
					disabled={disabled}
					placeholder={hasChips ? "" : t("invite.search_placeholder")}
					className={cn(
						"min-w-[7rem] flex-1 border-0 bg-transparent px-0 py-0 text-sm leading-normal shadow-none focus-visible:ring-0",
						hasChips ? "h-8 min-h-8 self-center sm:h-8 sm:min-h-8" : "h-full min-h-0 self-center",
					)}
					aria-expanded={showDropdown}
					aria-controls={listId}
					autoComplete="off"
					onFocus={() => setOpen(true)}
					onChange={(e) => {
						onQueryChange(e.target.value);
						setOpen(true);
					}}
					onKeyDown={onKeyDown}
				/>
			</div>
			{showDropdown ? (
				<ul
					id={listId}
					role="listbox"
					className={cn(
						"thin-scroll absolute top-[calc(100%+6px)] z-[90] max-h-52 w-full overflow-y-auto rounded-xl border border-border bg-card shadow-elevated",
						emailOnlySuggest ? "p-1.5" : "py-1",
					)}
				>
					{searchLoading ? (
						<li className="px-3 py-2.5 text-sm text-muted-foreground">
							{t("invite.search_loading")}
						</li>
					) : null}
					{searchHits.map((hit) => (
						<li key={hit.user_id}>
							<button
								type="button"
								role="option"
								aria-selected={false}
								className="flex w-full items-center gap-2.5 px-3 py-2 text-left text-sm transition-colors hover:bg-muted/80"
								onClick={() => pick(hit)}
							>
								<span
									className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-field text-xs font-semibold ring-1 ring-border/50"
									aria-hidden
								>
									{userInitials(hit.display_name, hit.email)}
								</span>
								<span className="min-w-0 truncate font-medium text-foreground">
									{hit.display_name.trim() || hit.email}
								</span>
							</button>
						</li>
					))}
					{showEmailHint ? (
						<li className={searchHits.length > 0 ? "border-t border-border/80" : undefined}>
							<button
								type="button"
								role="option"
								aria-selected={false}
								className={cn(
									"flex w-full items-center gap-3 text-left transition-colors",
									emailOnlySuggest
										? "rounded-lg bg-field/80 px-3 py-3 hover:bg-field focus-visible:bg-field"
										: "px-3 py-2.5 hover:bg-muted/80",
								)}
								disabled={disabled}
								onClick={() => tryCommitEmail()}
							>
								<span className="min-w-0 flex-1">
									<span className="block truncate text-sm font-semibold text-foreground">
										{trimmedQuery}
									</span>
									<span className="mt-0.5 block text-xs leading-snug text-muted-foreground">
										{t("invite.invite_email_caption")}
									</span>
								</span>
								<span className="shrink-0 rounded-md bg-primary px-2.5 py-1 text-[11px] font-semibold tracking-wide text-primary-foreground uppercase">
									{t("invite.invite_email_add")}
								</span>
							</button>
						</li>
					) : null}
					{!searchLoading && searchHits.length === 0 && !showEmailHint ? (
						<li className="px-3 py-2.5 text-sm text-muted-foreground">
							{t("invite.search_empty")}
						</li>
					) : null}
				</ul>
			) : null}
		</div>
	);
}

export function InviteRecipientActions({
	onShare,
	shareDisabled,
	pending,
}: {
	onShare: () => void;
	shareDisabled: boolean;
	pending: boolean;
}) {
	const { t } = useTranslation();
	return (
		<Button
			type="button"
			variant="default"
			size="default"
			disabled={shareDisabled || pending}
			className="h-10 shrink-0 gap-1.5 self-center px-3.5 sm:h-11 sm:min-h-11 sm:rounded-xl sm:px-4 max-h-[880px]:h-9 max-h-[880px]:sm:h-10"
			onClick={onShare}
		>
			<Send className="h-4 w-4 shrink-0" aria-hidden />
			{t("invite.send")}
		</Button>
	);
}
