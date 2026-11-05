"use client";

import { useTranslation } from "react-i18next";
import type { Group } from "@wichat/contracts";

type Props = {
	groups: readonly Group[];
};

export function HomeMemberIdlePage({ groups }: Props) {
	const { t } = useTranslation();
	const names = groups.map((g) => g.name).filter(Boolean);

	return (
		<div className="flex h-full min-h-0 flex-1 flex-col items-center justify-center px-4 py-12 text-center">
			<div className="max-w-md space-y-3">
				<h1 className="text-xl font-semibold tracking-tight">
					{t("group.home.member_idle_title")}
				</h1>
				<p className="text-sm leading-relaxed text-muted-foreground">
					{t("group.home.member_idle_lead")}
				</p>
				{names.length > 0 ? (
					<p className="text-sm text-muted-foreground">
						{t("group.home.member_idle_groups", { names: names.join(", ") })}
					</p>
				) : null}
			</div>
		</div>
	);
}
