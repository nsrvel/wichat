"use client";

import { useTranslation } from "react-i18next";
import { UserAvatar } from "@/components/ui/user-avatar";
import {
	recentPersonPresenceDot,
	type RecentPersonHit,
} from "@/features/search/lib/recent-interaction-people";
import { userInitials } from "@/lib/user-initials";

type ShellSearchPeopleStripProps = {
	people: RecentPersonHit[];
	onSelectPerson: (person: RecentPersonHit) => void;
};

export function ShellSearchPeopleStrip({ people, onSelectPerson }: ShellSearchPeopleStripProps) {
	const { t } = useTranslation();

	if (people.length === 0) {
		return null;
	}

	return (
		<div className="space-y-3.5">
			<p className="text-[11px] font-semibold uppercase tracking-wide text-muted-foreground/90">
				{t("shell.search.people_section")}
			</p>
			<div className="overlay-scroll -mx-1 flex gap-4 overflow-x-auto px-1 pb-1.5 pt-0.5">
				{people.map((person) => {
					const initials = userInitials(person.displayName, person.username);
					return (
						<button
							key={person.userId}
							type="button"
							onClick={() => onSelectPerson(person)}
							className="flex w-[5.25rem] shrink-0 flex-col items-center gap-2 rounded-xl px-1.5 py-2 text-center transition-colors hover:bg-muted/60"
						>
							<UserAvatar
								size="md"
								initials={initials}
								src={person.avatarUrl || undefined}
								status={recentPersonPresenceDot(person)}
								statusRingClassName="ring-card"
								alt=""
							/>
							<span className="line-clamp-2 w-full text-[11px] font-medium leading-tight text-foreground">
								{person.displayName}
							</span>
						</button>
					);
				})}
			</div>
		</div>
	);
}
