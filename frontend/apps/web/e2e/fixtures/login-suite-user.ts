import fs from "node:fs";
import path from "node:path";
import type { TestUser } from "./api-helpers";

const CACHE_FILE = path.join(process.cwd(), "e2e/fixtures/login-suite-user.json");

export function readLoginSuiteUser(): TestUser | undefined {
	if (!fs.existsSync(CACHE_FILE)) {
		return undefined;
	}
	try {
		return JSON.parse(fs.readFileSync(CACHE_FILE, "utf8")) as TestUser;
	} catch {
		return undefined;
	}
}

export function writeLoginSuiteUser(user: TestUser): void {
	fs.writeFileSync(CACHE_FILE, JSON.stringify(user, null, 0));
}
