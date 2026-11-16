import { expect, type Locator, type Page } from "@playwright/test";

export class MembersPage {
	constructor(private readonly page: Page) {}

	async gotoMembers(groupId: string): Promise<void> {
		await this.page.goto(`/channel/${encodeURIComponent(groupId)}/view/members`);
		await expect(this.page.getByRole("table")).toBeVisible({ timeout: 15_000 });
	}

	rowByName(displayName: string): Locator {
		return this.page.locator("tbody tr", { hasText: displayName }).first();
	}

	async search(query: string): Promise<void> {
		const input = this.page.getByPlaceholder(/search/i);
		await input.fill(query);
	}

	async sortBy(label: string): Promise<void> {
		await this.page.getByRole("button", { name: /^sort$/i }).click();
		await this.page.getByRole("menuitemradio", { name: label }).click();
	}

	async openRowActions(displayName: string): Promise<void> {
		const row = this.rowByName(displayName);
		await row.getByRole("button", { name: new RegExp(`actions for ${displayName}`, "i") }).click();
	}

	async openProfile(displayName: string): Promise<Locator> {
		await this.openRowActions(displayName);
		await this.page.getByRole("menuitem", { name: /^profile$/i }).click();
		const dialog = this.page.getByRole("dialog");
		await expect(dialog).toBeVisible();
		return dialog;
	}

	async toggleRoleOnRow(displayName: string, roleName: string): Promise<void> {
		const row = this.rowByName(displayName);
		const addBtn = row.getByRole("button", { name: /add role/i });
		if (await addBtn.isVisible()) {
			await addBtn.click();
		} else {
			await row.getByRole("button", { name: /more roles/i }).click();
		}
		await this.page.getByRole("button", { name: roleName, exact: false }).click();
		await this.page.keyboard.press("Escape");
	}

	async timeoutMember(displayName: string, durationLabel: string, reason: string): Promise<void> {
		await this.openRowActions(displayName);
		await this.page
			.getByRole("menuitem", { name: new RegExp(`^timeout ${displayName}`, "i") })
			.click();

		const dialog = this.page.getByRole("dialog");
		await expect(dialog).toBeVisible();
		await dialog.locator('button[aria-haspopup="listbox"]').click();
		await this.page.getByRole("option", { name: durationLabel }).click();
		await dialog.locator("textarea").fill(reason);
		await dialog.getByRole("button", { name: /^timeout$/i }).click();
		await expect(dialog).toBeHidden();
	}

	async removeTimeout(displayName: string): Promise<void> {
		await this.openRowActions(displayName);
		await this.page
			.getByRole("menuitem", { name: new RegExp(`^remove timeout ${displayName}`, "i") })
			.click();
	}

	async kickMember(displayName: string, reason: string): Promise<void> {
		await this.openRowActions(displayName);
		await this.page
			.getByRole("menuitem", { name: new RegExp(`^kick ${displayName}`, "i") })
			.click();

		const dialog = this.page.getByRole("dialog");
		await expect(dialog).toBeVisible();
		await dialog.locator("textarea").fill(reason);
		await dialog.getByRole("button", { name: /^kick$/i }).click();
		await expect(dialog).toBeHidden();
	}

	async banMember(displayName: string, durationLabel: string, reason: string): Promise<void> {
		await this.openRowActions(displayName);
		await this.page.getByRole("menuitem", { name: new RegExp(`^ban ${displayName}`, "i") }).click();

		const dialog = this.page.getByRole("dialog");
		await expect(dialog).toBeVisible();
		await dialog.locator('button[aria-haspopup="listbox"]').click();
		await this.page.getByRole("option", { name: durationLabel }).click();
		await dialog.locator("textarea").fill(reason);
		await dialog.getByRole("button", { name: /^ban$/i }).click();
		await expect(dialog).toBeHidden();
	}

	async selectRow(displayName: string): Promise<void> {
		await this.page
			.getByRole("checkbox", { name: new RegExp(`select ${displayName}`, "i") })
			.check();
	}

	async bulkBanSelected(durationLabel: string, reason: string): Promise<void> {
		await this.page.getByRole("button", { name: /^ban members$/i }).click();
		const dialog = this.page.getByRole("dialog");
		await expect(dialog).toBeVisible();
		await dialog.locator('button[aria-haspopup="listbox"]').click();
		await this.page.getByRole("option", { name: durationLabel }).click();
		await dialog.locator("textarea").fill(reason);
		await dialog.getByRole("button", { name: /^ban$/i }).click();
		await expect(dialog).toBeHidden();
	}
}
