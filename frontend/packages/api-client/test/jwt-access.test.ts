import assert from "node:assert/strict";
import { describe, it } from "node:test";
import {
	accessTokenExpiresAtMs,
	isAccessTokenStale,
} from "../src/jwt-access.ts";

function jwtWithExp(expSec: number): string {
	const header = btoa(JSON.stringify({ alg: "none", typ: "JWT" }));
	const payload = btoa(JSON.stringify({ exp: expSec }));
	return `${header}.${payload}.sig`;
}

describe("jwt-access", () => {
	it("reads exp from payload", () => {
		const expMs = accessTokenExpiresAtMs(jwtWithExp(1_700_000_000));
		assert.equal(expMs, 1_700_000_000_000);
	});

	it("treats near-expiry token as stale", () => {
		const expSec = Math.floor(Date.now() / 1000) + 10;
		assert.equal(isAccessTokenStale(jwtWithExp(expSec), 30_000), true);
	});

	it("keeps fresh token", () => {
		const expSec = Math.floor(Date.now() / 1000) + 3600;
		assert.equal(isAccessTokenStale(jwtWithExp(expSec)), false);
	});
});
