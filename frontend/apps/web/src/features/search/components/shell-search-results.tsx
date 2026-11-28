"use client";

import type { FileSearchHit } from "@wichat/contracts/chat/search";
import { FileText, Hash, MessageSquare } from "lucide-react";
import { useTranslation } from "react-i18next";
import type { ShellChannelSearchHit } from "@/features/search/lib/filter-group-channels";
import type { ShellMessageSearchHit } from "@/features/search/lib/shell-message-search-hit";
import {
	recentPersonPresenceDot,
	type RecentPersonHit,
} from "@/features/search/lib/recent-interaction-people";
import type { ShellSearchFilterId } from "@/features/search/lib/shell-search-filters";
import { UserAvatar } from "@/components/ui/user-avatar";
import { userInitials } from "@/lib/user-initials";

type ShellSearchResultsProps = {
	filter: ShellSearchFilterId;
	query: string;
	messageHits: ShellMessageSearchHit[];
	messageLoading?: boolean;
	channelHits: ShellChannelSearchHit[];
	fileHits: FileSearchHit[];
	fileLoading?: boolean;
	peopleHits: RecentPersonHit[];
	peopleLoading: boolean;
	onSelectMessage: (hit: ShellMessageSearchHit) => void;
	onSelectChannel: (hit: ShellChannelSearchHit) => void;
	onSelectFile: (hit: FileSearchHit) => void;
	onSelectPerson: (person: RecentPersonHit) => void;
};

export function ShellSearchResults({
	filter,
	query,
	messageHits,
	messageLoading,
	channelHits,
	fileHits,
	fileLoading,
	peopleHits,
	peopleLoading,
	onSelectMessage,
	onSelectChannel,
	onSelectFile,
	onSelectPerson,
}: ShellSearchResultsProps) {
	const { t } = useTranslation();
	const trimmed = query.trim();

	if (!trimmed) {
		return null;
	}

	if (filter === "files" || filter === "images") {
		if (fileLoading && fileHits.length === 0) {
			return <p className="py-3 text-xs text-muted-foreground">{t("shell.search.loading")}</p>;
		}
		if (fileHits.length === 0) {
			return <p className="py-3 text-sm text-muted-foreground">{t("shell.search.no_results")}</p>;
		}
		return (
			<div className="space-y-3 pt-0.5">
				<p className="text-[11px] font-semibold uppercase tracking-wide text-muted-foreground/90">
					{filter === "images" ? t("shell.search.filter_images") : t("shell.search.filter_files")}
				</p>
				<ul className="space-y-1">
					{fileHits.map((hit) => (
						<li key={hit.attachment_id}>
							<button
								type="button"
								onClick={() => onSelectFile(hit)}
								className="flex w-full items-center gap-3 rounded-lg px-2.5 py-2.5 text-left transition-colors hover:bg-muted/70"
							>
								<FileText className="size-4 shrink-0 text-muted-foreground" aria-hidden />
								<span className="truncate text-sm font-medium text-foreground">{hit.filename}</span>
							</button>
						</li>
					))}
				</ul>
			</div>
		);
	}

	if (filter === "channels") {
		if (channelHits.length === 0) {
			return <p className="py-3 text-sm text-muted-foreground">{t("shell.search.no_results")}</p>;
		}
		return (
			<div className="space-y-3 pt-0.5">
				<p className="text-[11px] font-semibold uppercase tracking-wide text-muted-foreground/90">
					{t("shell.search.filter_channels")}
				</p>
				<ul className="space-y-1">
					{channelHits.map((hit) => (
						<li key={hit.channelId}>
							<button
								type="button"
								onClick={() => onSelectChannel(hit)}
								className="flex w-full items-center gap-3 rounded-lg px-2.5 py-2.5 text-left transition-colors hover:bg-muted/70"
							>
								<Hash className="size-4 shrink-0 text-muted-foreground" aria-hidden />
								<span className="truncate text-sm font-medium text-foreground">{hit.name}</span>
							</button>
						</li>
					))}
				</ul>
			</div>
		);
	}

	const showPeople = filter === "messages" && (peopleHits.length > 0 || peopleLoading);
	const showMessages = filter === "messages" && (messageHits.length > 0 || messageLoading);

	if (!showPeople && !showMessages) {
		return <p className="py-3 text-sm text-muted-foreground">{t("shell.search.no_results")}</p>;
	}

	return (
		<div className="space-y-6 pt-0.5">
			{showPeople ? (
				<div className="space-y-3">
					<p className="text-[11px] font-semibold uppercase tracking-wide text-muted-foreground/90">
						{t("shell.search.people_results")}
					</p>
					{peopleLoading && peopleHits.length === 0 ? (
						<p className="px-0.5 text-xs text-muted-foreground/80">{t("shell.search.loading")}</p>
					) : (
						<ul className="space-y-1">
							{peopleHits.map((person) => (
								<li key={person.userId}>
									<button
										type="button"
										onClick={() => onSelectPerson(person)}
										className="flex w-full items-center gap-3 rounded-lg px-2.5 py-2.5 text-left transition-colors hover:bg-muted/70"
									>
										<UserAvatar
											size="sm"
											initials={userInitials(person.displayName, person.username)}
											src={person.avatarUrl || undefined}
											status={recentPersonPresenceDot(person)}
											statusRingClassName="ring-card"
											alt=""
										/>
										<div className="min-w-0 flex-1">
											<p className="truncate text-sm font-medium text-foreground">
												{person.displayName}
											</p>
											{person.username ? (
												<p className="truncate text-xs text-muted-foreground">@{person.username}</p>
											) : null}
										</div>
									</button>
								</li>
							))}
						</ul>
					)}
				</div>
			) : null}

			{showMessages ? (
				<div className="space-y-3">
					<p className="text-[11px] font-semibold uppercase tracking-wide text-muted-foreground/90">
						{t("shell.search.messages_in_channel")}
					</p>
					{messageLoading && messageHits.length === 0 ? (
						<p className="text-xs text-muted-foreground/80">{t("shell.search.loading")}</p>
					) : null}
					<ul className="space-y-1">
						{messageHits.map((hit) => (
							<li key={hit.messageId}>
								<button
									type="button"
									onClick={() => onSelectMessage(hit)}
									className="flex w-full gap-3 rounded-lg px-2.5 py-2.5 text-left transition-colors hover:bg-muted/70"
								>
									<MessageSquare
										className="mt-0.5 size-4 shrink-0 text-muted-foreground"
										aria-hidden
									/>
									<div className="min-w-0 flex-1">
										<p className="truncate text-xs font-medium text-foreground">
											{hit.authorDisplayName}
										</p>
										<p className="line-clamp-2 text-sm text-muted-foreground">{hit.bodyPreview}</p>
									</div>
								</button>
							</li>
						))}
					</ul>
				</div>
			) : null}
		</div>
	);
}
