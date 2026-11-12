import type { Column } from "@tanstack/react-table";
import type { CSSProperties } from "react";

export const ACTIONS_COLUMN_FIXED_WIDTH = 52;
export const SELECT_COLUMN_FIXED_WIDTH = 44;

/** Members table column specs using native TanStack Table sizing properties. */
export const MEMBERS_COLUMN_SIZE = {
	select: {
		size: SELECT_COLUMN_FIXED_WIDTH,
		minSize: SELECT_COLUMN_FIXED_WIDTH,
		maxSize: SELECT_COLUMN_FIXED_WIDTH,
		enableResizing: false,
	},
	name: {
		size: 260,
		minSize: 180,
	},
	member_since: {
		size: 140,
		minSize: 104,
	},
	joined_wichat: {
		size: 140,
		minSize: 104,
	},
	join_method: {
		size: 120,
		minSize: 96,
	},
	roles: {
		size: 160,
		minSize: 120,
	},
	last_updated: {
		size: 150,
		minSize: 104,
	},
	actions: {
		size: ACTIONS_COLUMN_FIXED_WIDTH,
		minSize: ACTIONS_COLUMN_FIXED_WIDTH,
		maxSize: ACTIONS_COLUMN_FIXED_WIDTH,
		enableResizing: false,
	},
} as const;

/** Roles table column specs using native TanStack Table sizing properties. */
export const ROLES_COLUMN_SIZE = {
	role: {
		size: 320,
		minSize: 180,
	},
	members: {
		size: 120,
		minSize: 100,
	},
	actions: {
		size: ACTIONS_COLUMN_FIXED_WIDTH,
		minSize: ACTIONS_COLUMN_FIXED_WIDTH,
		maxSize: ACTIONS_COLUMN_FIXED_WIDTH,
		enableResizing: false,
	},
} as const;

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

function isFixedColumn<TData>(column: Column<TData, unknown>): boolean {
	return (
		column.columnDef.enableResizing === false &&
		column.columnDef.maxSize === column.columnDef.minSize
	);
}

/**
 * Calculates CSS properties for a column header/cell.
 * For fixed columns (checkbox, actions), returns fixed width in pixels.
 * For flexible columns, allocates available percentage based on TanStack Table `column.getSize()`.
 */
export function dataTableColumnPercentStyle<TData>(
	column: Column<TData, unknown>,
	leafColumns: Column<TData, unknown>[],
): CSSProperties {
	const minSize = column.columnDef.minSize;
	const maxSize = column.columnDef.maxSize;
	const isFixed = isFixedColumn(column);

	if (isFixed) {
		const fixedWidth = column.getSize();
		return {
			width: `${fixedWidth}px`,
			minWidth: `${fixedWidth}px`,
			maxWidth: `${fixedWidth}px`,
		};
	}

	const fixedColumnsWidth = leafColumns.reduce((sum, col) => {
		return isFixedColumn(col) ? sum + col.getSize() : sum;
	}, 0);

	const flexColumnsTotalSize = leafColumns.reduce((sum, col) => {
		return isFixedColumn(col) ? sum : sum + col.getSize();
	}, 0);

	const currentSize = column.getSize();
	const style: CSSProperties =
		flexColumnsTotalSize > 0
			? fixedColumnsWidth > 0
				? { width: `calc((100% - ${fixedColumnsWidth}px) * ${currentSize / flexColumnsTotalSize})` }
				: { width: `${(currentSize / flexColumnsTotalSize) * 100}%` }
			: { width: `${(1 / Math.max(leafColumns.length, 1)) * 100}%` };

	if (minSize != null) {
		style.minWidth = `${minSize}px`;
	}
	if (maxSize != null && maxSize < Number.MAX_SAFE_INTEGER) {
		style.maxWidth = `${maxSize}px`;
	}

	return style;
}

/**
 * Computes the minimum table width based on the sum of `minSize` (or `size`) across all visible leaf columns.
 */
export function dataTableMinTableStyle<TData>(columns: Column<TData, unknown>[]): CSSProperties {
	let minWidth = 0;
	for (const col of columns) {
		const px = col.columnDef.minSize ?? col.getSize();
		if (px != null) {
			minWidth += px;
		}
	}
	return minWidth > 0 ? { minWidth: `${minWidth}px` } : {};
}
