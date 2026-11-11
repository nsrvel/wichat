import type { Column } from "@tanstack/react-table";
import type { CSSProperties } from "react";

/** Column sizing metadata for group settings data tables (`table-fixed` + `w-full`). */
export type GroupSettingsColumnMeta = {
	/** Relative flex weight. Omit or set to 0 if column is static/fixed width. */
	columnWeight?: number;
	/** Minimum width in px (for horizontal scroll on narrow viewports). */
	columnMinWidth?: number;
	/** Static/fixed width in px (e.g. action buttons, select checkbox). */
	columnFixedWidth?: number;
};

declare module "@tanstack/react-table" {
	// eslint-disable-next-line @typescript-eslint/no-unused-vars
	interface ColumnMeta<TData, TValue> {
		groupSettings?: GroupSettingsColumnMeta;
	}
}

type GroupSettingsColumnLayout = {
	size: number;
	meta: { groupSettings: GroupSettingsColumnMeta };
};

export const ACTIONS_COLUMN_FIXED_WIDTH = 40;
export const SELECT_COLUMN_FIXED_WIDTH = 40;

/** Members table column specs: actions/select fixed width; others percentage/flex with minWidth. */
export const MEMBERS_COLUMN_SIZE = {
	select: {
		size: SELECT_COLUMN_FIXED_WIDTH,
		meta: { groupSettings: { columnFixedWidth: SELECT_COLUMN_FIXED_WIDTH } },
	},
	name: {
		size: 28,
		meta: { groupSettings: { columnWeight: 28, columnMinWidth: 180 } },
	},
	member_since: {
		size: 13,
		meta: { groupSettings: { columnWeight: 13, columnMinWidth: 104 } },
	},
	joined_wichat: {
		size: 13,
		meta: { groupSettings: { columnWeight: 13, columnMinWidth: 104 } },
	},
	join_method: {
		size: 12,
		meta: { groupSettings: { columnWeight: 12, columnMinWidth: 96 } },
	},
	roles: {
		size: 17,
		meta: { groupSettings: { columnWeight: 17, columnMinWidth: 120 } },
	},
	last_updated: {
		size: 17,
		meta: { groupSettings: { columnWeight: 17, columnMinWidth: 104 } },
	},
	actions: {
		size: ACTIONS_COLUMN_FIXED_WIDTH,
		meta: { groupSettings: { columnFixedWidth: ACTIONS_COLUMN_FIXED_WIDTH } },
	},
} as const satisfies Record<string, GroupSettingsColumnLayout>;

/** Roles table column specs: actions fixed width; role & members flex with minWidth. */
export const ROLES_COLUMN_SIZE = {
	role: {
		size: 70,
		meta: { groupSettings: { columnWeight: 70, columnMinWidth: 180 } },
	},
	members: {
		size: 30,
		meta: { groupSettings: { columnWeight: 30, columnMinWidth: 100 } },
	},
	actions: {
		size: ACTIONS_COLUMN_FIXED_WIDTH,
		meta: { groupSettings: { columnFixedWidth: ACTIONS_COLUMN_FIXED_WIDTH } },
	},
} as const satisfies Record<string, GroupSettingsColumnLayout>;

export const groupSettingsDataTableThClass =
	"px-3 py-2.5 text-left text-[11px] font-medium tracking-normal text-muted-foreground sm:text-xs";

export const groupSettingsDataTableTdClass = "px-3 py-2.5 align-middle text-sm text-foreground";

export function dataTableEdgeCellClass(columnId: string): string | undefined {
	if (columnId === "select" || columnId === "role" || columnId === "name") {
		return "pl-4";
	}
	if (columnId === "actions") {
		return "pr-4";
	}
	return undefined;
}

function columnMeta<TData>(column: Column<TData, unknown>): GroupSettingsColumnMeta | undefined {
	return column.columnDef.meta?.groupSettings;
}

function columnWeight<TData>(column: Column<TData, unknown>): number {
	const meta = columnMeta(column);
	if (meta?.columnFixedWidth != null) {
		return 0;
	}
	return meta?.columnWeight ?? column.columnDef.size ?? 1;
}

function isFlexColumn<TData>(column: Column<TData, unknown>): boolean {
	return columnMeta(column)?.columnFixedWidth == null;
}

function fixedWidthTotal<TData>(columns: Column<TData, unknown>[]): number {
	return columns.reduce((sum, col) => {
		const fixed = columnMeta(col)?.columnFixedWidth;
		return fixed != null ? sum + fixed : sum;
	}, 0);
}

export function dataTableTotalWeight<TData>(columns: Column<TData, unknown>[]): number {
	return columns.filter(isFlexColumn).reduce((sum, col) => sum + columnWeight(col), 0);
}

export function dataTableColumnPercentStyle<TData>(
	column: Column<TData, unknown>,
	leafColumns: Column<TData, unknown>[],
): CSSProperties {
	const meta = columnMeta(column);
	const fixed = meta?.columnFixedWidth;
	if (fixed != null) {
		return {
			width: `${fixed}px`,
			minWidth: `${fixed}px`,
			maxWidth: `${fixed}px`,
		};
	}

	const flexTotal = dataTableTotalWeight(leafColumns);
	const weight = columnWeight(column);
	const fixedTotal = fixedWidthTotal(leafColumns);
	const minWidth = meta?.columnMinWidth;

	const style: CSSProperties =
		flexTotal > 0
			? fixedTotal > 0
				? { width: `calc((100% - ${fixedTotal}px) * ${weight / flexTotal})` }
				: { width: `${(weight / flexTotal) * 100}%` }
			: { width: `${(weight / Math.max(leafColumns.length, 1)) * 100}%` };

	if (minWidth != null) {
		style.minWidth = `${minWidth}px`;
	}
	return style;
}

export function dataTableMinTableStyle<TData>(columns: Column<TData, unknown>[]): CSSProperties {
	let minWidth = 0;
	for (const col of columns) {
		const meta = columnMeta(col);
		const px = meta?.columnFixedWidth ?? meta?.columnMinWidth;
		if (px != null) {
			minWidth += px;
		}
	}
	return minWidth > 0 ? { minWidth: `${minWidth}px` } : {};
}
