"use client";

import type { User } from "@wichat/contracts";
import { X } from "lucide-react";
import { useCallback, useEffect, useId, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { useTranslation } from "react-i18next";
import { UserSettingsAccountSection } from "@/features/user-settings/components/sections/user-settings-account-section";
import { UserSettingsLanguageSection } from "@/features/user-settings/components/sections/user-settings-language-section";
import { UserSettingsPlaceholderSection } from "@/features/user-settings/components/sections/user-settings-placeholder-section";
import { UserSettingsProfilesSection } from "@/features/user-settings/components/sections/user-settings-profiles-section";
import { UserSettingsSidebar } from "@/features/user-settings/components/user-settings-sidebar";
import {
	USER_SETTINGS_ITEM_BY_ID,
	USER_SETTINGS_NAV_GROUPS,
	type UserSettingsSectionId,
} from "@/features/user-settings/lib/user-settings-nav";
import type { UserPresenceStatus } from "@/lib/user-status";
import { cn } from "@/lib/utils";
import { useLogoutMutation } from "@/services/auth/auth.mutation";

type UserSettingsDialogProps = {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	initialSection?: UserSettingsSectionId;
	user: User | null | undefined;
	displayName: string;
	secondaryText: string;
	avatarUrl: string;
	initials: string;
	status: UserPresenceStatus;
};

export function UserSettingsDialog({
	open,
	onOpenChange,
	initialSection = "account",
	user,
	displayName,
	secondaryText,
	avatarUrl,
	initials,
	status,
}: UserSettingsDialogProps) {
	const { t } = useTranslation();
	const titleId = useId();
	const contentScrollRef = useRef<HTMLDivElement>(null);
	const logout = useLogoutMutation();
	const [activeSection, setActiveSection] = useState<UserSettingsSectionId>(initialSection);
	const [activeSubAnchor, setActiveSubAnchor] = useState<string | null>(null);
	const [searchQuery, setSearchQuery] = useState("");
	const [prevOpen, setPrevOpen] = useState(open);

	if (open && !prevOpen) {
		setPrevOpen(true);
		setActiveSection(initialSection);
		setActiveSubAnchor(null);
		setSearchQuery("");
	} else if (!open && prevOpen) {
		setPrevOpen(false);
	}

	useEffect(() => {
		if (!open) {
			return;
		}
		const onKeyDown = (e: KeyboardEvent) => {
			if (e.key === "Escape") {
				e.preventDefault();
				onOpenChange(false);
			}
		};
		window.addEventListener("keydown", onKeyDown);
		return () => window.removeEventListener("keydown", onKeyDown);
	}, [onOpenChange, open]);

	const handleSelectSection = useCallback((section: UserSettingsSectionId, anchorId?: string) => {
		setActiveSection(section);
		setActiveSubAnchor(anchorId ?? null);
		if (anchorId) {
			requestAnimationFrame(() => {
				const targetId = anchorId.startsWith("user-settings-")
					? anchorId
					: `user-settings-${anchorId}`;
				const el = document.getElementById(targetId);
				if (!el) {
					return;
				}
				el.scrollIntoView({ behavior: "smooth", block: "center" });
				if (
					el instanceof HTMLInputElement ||
					el instanceof HTMLTextAreaElement ||
					el instanceof HTMLButtonElement
				) {
					el.focus({ preventScroll: true });
				}
				el.classList.add("ring-2", "ring-ring/70", "ring-offset-2", "ring-offset-background", "rounded-lg");
				window.setTimeout(() => {
					el.classList.remove("ring-2", "ring-ring/70", "ring-offset-2", "ring-offset-background", "rounded-lg");
				}, 1400);
			});
			return;
		}
		contentScrollRef.current?.scrollTo({ top: 0, behavior: "auto" });
	}, []);

	const handleSignOut = () => {
		logout.mutate(undefined, {
			onSettled: () => onOpenChange(false),
		});
	};

	if (!open || typeof document === "undefined") {
		return null;
	}

	const activeItem = USER_SETTINGS_ITEM_BY_ID[activeSection];
	const username = user?.username?.trim() || "";

	return createPortal(
		<div className="fixed inset-0 z-[70] grid place-items-center p-3 sm:p-6">
			<button
				type="button"
				tabIndex={-1}
				aria-label={t("user_settings.close")}
				onClick={() => onOpenChange(false)}
				className="fixed inset-0 bg-black/65 backdrop-blur-[2px]"
			/>

			<div
				role="dialog"
				aria-modal="true"
				aria-labelledby={titleId}
				className="relative z-10 flex h-[min(86dvh,720px)] w-[min(960px,calc(100vw-1.5rem))] overflow-hidden rounded-2xl border border-border bg-background text-foreground shadow-elevated"
			>
				{/* Desktop left sidebar */}
				<div className="hidden md:flex">
					<UserSettingsSidebar
						displayName={displayName}
						secondaryText={secondaryText}
						avatarUrl={avatarUrl}
						initials={initials}
						status={status}
						activeSection={activeSection}
						activeSubAnchor={activeSubAnchor}
						searchQuery={searchQuery}
						onSearchChange={setSearchQuery}
						onSelectSection={handleSelectSection}
						onSignOut={handleSignOut}
						signingOut={logout.isPending}
					/>
				</div>

				{/* Right content pane */}
				<div className="relative flex min-w-0 flex-1 flex-col">
					<header className="flex shrink-0 items-start justify-between gap-3 px-5 pt-4 pb-2 sm:px-7">
						<div className="min-w-0">
							<h2 id={titleId} className="truncate text-base font-semibold text-foreground">
								{t(activeItem.labelKey)}
							</h2>
							{activeItem.subtitleKey ? (
								<p className="mt-0.5 text-xs text-muted-foreground">
									{t(activeItem.subtitleKey)}
								</p>
							) : null}
						</div>
						<button
							type="button"
							onClick={() => onOpenChange(false)}
							aria-label={t("user_settings.close")}
							className="flex size-8 shrink-0 items-center justify-center rounded-lg text-muted-foreground transition-colors hover:bg-foreground/10 hover:text-foreground"
						>
							<X className="size-4" aria-hidden />
						</button>
					</header>

					{/* Mobile section strip */}
					<div className="flex shrink-0 items-center gap-1 overflow-x-auto border-b border-border/60 px-4 py-2 md:hidden">
						{USER_SETTINGS_NAV_GROUPS.flatMap((group) => group.items).map((item) => {
							const active = item.id === activeSection;
							return (
								<button
									key={item.id}
									type="button"
									onClick={() => handleSelectSection(item.id)}
									className={cn(
										"shrink-0 rounded-lg px-2.5 py-1.5 text-xs font-medium transition-colors",
										active
											? "bg-sidebar-accent text-sidebar-accent-foreground"
											: "text-muted-foreground hover:bg-foreground/5 hover:text-foreground",
									)}
								>
									{t(item.labelKey)}
								</button>
							);
						})}
					</div>

					<div
						ref={contentScrollRef}
						className="overlay-scroll flex-1 overflow-y-auto px-5 py-6 sm:px-7"
					>
						{activeSection === "account" ? (
							<UserSettingsAccountSection
								user={user}
								onNavigateSection={(sec) => handleSelectSection(sec)}
							/>
						) : activeSection === "profiles" ? (
							<UserSettingsProfilesSection
								displayName={displayName}
								username={username}
								avatarUrl={avatarUrl}
								initials={initials}
								status={status}
							/>
						) : activeSection === "language" ? (
							<UserSettingsLanguageSection />
						) : (
							<UserSettingsPlaceholderSection sectionId={activeSection} />
						)}
					</div>
				</div>
			</div>
		</div>,
		document.body,
	);
}
