/** Fullscreen overlay — split backdrop is painted in the shell body. */
export const groupSettingsOverlayClass = "fixed inset-0 z-[60] flex flex-col";

export const groupSettingsShellClass = "flex h-full min-h-0 w-full flex-col overflow-hidden";

export const groupSettingsCloseButtonClass =
	"fixed top-3 right-3 z-[70] h-10 gap-3 rounded-md pr-2.5 pl-3.5 text-muted-foreground sm:top-5 sm:right-5";

export const groupSettingsBodyOuterClass =
	"overlay-scroll relative z-[1] flex min-h-0 flex-1 justify-center self-stretch overflow-y-auto overscroll-contain";

/** Full-viewport two-tone split aligned with centered max-w-6xl + 14rem sidebar. */
export const groupSettingsSplitBackdropClass = [
	"pointer-events-none fixed inset-0 z-0 bg-background",
	"lg:bg-transparent",
	"lg:[background:linear-gradient(to_right,var(--sidebar)_0,var(--sidebar)_calc((100vw-min(100vw,72rem))/2+14rem),var(--background)_calc((100vw-min(100vw,72rem))/2+14rem),var(--background)_100%)]",
].join(" ");

/** Sidebar + main share one max width; sidebar bg fills the left slice of this block. */
export const groupSettingsCanvasClass =
	"relative z-[1] grid w-full min-h-full max-w-6xl grid-cols-1 items-stretch lg:min-h-[100dvh] lg:grid-cols-[14rem_minmax(0,1fr)]";

export const groupSettingsSidebarPaneClass =
	"flex min-h-full flex-col bg-sidebar text-sidebar-foreground lg:sticky lg:top-0 lg:z-[2] lg:min-h-[100dvh] lg:max-h-[100dvh] lg:overflow-hidden";

export const groupSettingsMainPaneClass =
	"relative flex min-h-full min-w-0 flex-1 flex-col bg-background lg:min-h-[100dvh]";

export const groupSettingsFormFooterBarClass =
	"flex min-h-12 items-center gap-4 rounded-xl border border-border bg-card px-4 py-2 sm:min-h-[3.25rem] sm:px-5";

export const groupSettingsFormFooterBarNudgeClass =
	"ring-2 ring-destructive/45 ring-offset-2 ring-offset-background";

export const groupSettingsFormFooterFloatShellClass =
	"pointer-events-none fixed inset-x-0 bottom-0 z-[65] flex justify-center px-6 pb-6 sm:px-10 sm:pb-8";

export const groupSettingsFormFooterFloatAlignClass =
	"pointer-events-auto grid w-full max-w-6xl grid-cols-1 lg:grid-cols-[14rem_minmax(0,1fr)]";

export const groupSettingsFormFooterFloatColumnClass = "min-w-0 lg:col-start-2 lg:px-12";

export const groupSettingsMainScrollDirtyPadClass = "pb-28 sm:pb-32";

export const groupSettingsMainScrollClass =
	"overlay-scroll flex min-h-0 flex-1 flex-col overflow-y-auto overscroll-contain px-6 pt-12 pb-6 sm:px-10 sm:pt-14 sm:pb-8 lg:px-12";

/** Members quick view — tighter horizontal inset so the table can use the workspace width. */
export const groupSettingsMembersWorkspaceScrollClass =
	"overlay-scroll flex min-h-0 flex-1 flex-col overflow-y-auto overscroll-contain px-2 py-4 sm:px-3 sm:py-5 lg:px-4 lg:py-6";

export const groupSettingsMembersWorkspaceInnerClass =
	"mx-auto flex w-full max-w-none flex-col gap-6 pb-4";

export const groupSettingsMainInnerClass = "mx-auto w-full max-w-3xl";

export const groupSettingsSectionHeaderClass = "pb-4 sm:pb-5";

export const groupSettingsSectionTitleClass =
	"text-lg font-semibold capitalize tracking-tight text-foreground";

export const groupSettingsSectionLeadClass = "mt-1.5 text-sm leading-relaxed text-muted-foreground";

export const groupSettingsProfileSectionDividerWrapClass = "py-5 sm:py-6";

export const groupSettingsProfileSectionDividerLineClass = "border-t border-border";

export const groupSettingsProfileSectionGapClass = "pt-5 sm:pt-6";

/** Members / roles TanStack tables — borderless card */
export const groupSettingsDataTableShellClass = "overflow-hidden rounded-xl bg-field/40";

export const groupSettingsDataTableHeadRowClass = "";

export const groupSettingsDataTableBodyClass = "";

export const groupSettingsDataTableRowClass = "transition-colors hover:bg-background/40";

export const groupSettingsDataTablePaginationNavClass =
	"flex flex-col gap-2 pt-3 sm:flex-row sm:items-center sm:justify-between";
