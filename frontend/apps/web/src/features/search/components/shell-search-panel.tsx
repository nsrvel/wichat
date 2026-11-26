"use client";

import { useTranslation } from "react-i18next";
import { ShellSearchFilterChips } from "@/features/search/components/shell-search-filter-chips";
import { ShellSearchPeopleStrip } from "@/features/search/components/shell-search-people-strip";
import { ShellSearchRecentQueries } from "@/features/search/components/shell-search-recent-queries";
import { ShellSearchResults } from "@/features/search/components/shell-search-results";
import type { ShellMessageSearchHit } from "@/features/search/lib/shell-message-search-hit";
import type { RecentPersonHit } from "@/features/search/lib/recent-interaction-people";
import type { FileSearchHit } from "@wichat/contracts/chat/search";
import type { ShellChannelSearchHit } from "@/features/search/lib/filter-group-channels";
import type { ShellSearchFilterId } from "@/features/search/lib/shell-search-filters";

export type ShellSearchPanelProps = {
	panelId: string;
	query: string;
	filter: ShellSearchFilterId;
	onFilterChange: (id: ShellSearchFilterId) => void;
	recentPeople: RecentPersonHit[];
	recentQueries: string[];
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
	onRecentSelect: (query: string) => void;
	onClearRecent: () => void;
};

export function ShellSearchPanel({
	panelId,
	query,
	filter,
	onFilterChange,
	recentPeople,
	recentQueries,
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
	onRecentSelect,
	onClearRecent,
}: ShellSearchPanelProps) {
	const { t } = useTranslation();
	const showIdleSections = !query.trim();

	return (
		<div id={panelId} role="dialog" aria-label={t("shell.search.open_label")}>
			<div className="px-5 pb-3 pt-4">
				<ShellSearchFilterChips active={filter} onChange={onFilterChange} />
			</div>

			<div className="overlay-scroll max-h-[min(72vh,30rem)] space-y-6 px-5 py-5">
				<ShellSearchResults
					filter={filter}
					query={query}
					messageHits={messageHits}
					messageLoading={messageLoading}
					channelHits={channelHits}
					fileHits={fileHits}
					fileLoading={fileLoading}
					peopleHits={peopleHits}
					peopleLoading={peopleLoading}
					onSelectMessage={onSelectMessage}
					onSelectChannel={onSelectChannel}
					onSelectFile={onSelectFile}
					onSelectPerson={onSelectPerson}
				/>

				{showIdleSections ? (
					<>
						<ShellSearchPeopleStrip people={recentPeople} onSelectPerson={onSelectPerson} />
						<ShellSearchRecentQueries
							items={recentQueries}
							onSelect={onRecentSelect}
							onClear={onClearRecent}
						/>
					</>
				) : null}
			</div>
		</div>
	);
}
