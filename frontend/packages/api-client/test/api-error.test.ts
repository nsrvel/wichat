import assert from "node:assert/strict";
import { describe, it } from "node:test";
import { ApiError } from "../src/api-error.ts";
import { isClientErrorCode } from "../src/client-error-code.ts";

describe("isClientErrorCode", () => {
	it("accepts known API keys", () => {
		assert.equal(isClientErrorCode("auth.invalid_credentials"), true);
		assert.equal(isClientErrorCode("error.internal"), true);
		assert.equal(isClientErrorCode("user.not_found"), true);
	});

	it("rejects unknown or leaky codes", () => {
		assert.equal(isClientErrorCode("rpc error: internal"), false);
		assert.equal(isClientErrorCode("validation.login"), false);
		assert.equal(isClientErrorCode(""), false);
	});
});

describe("ApiError.fromEnvelope", () => {
	it("maps valid envelope to ApiError", () => {
		const err = ApiError.fromEnvelope({
			error: { code: "auth.unauthorized", params: null },
		});
		assert.equal(err.code, "auth.unauthorized");
	});

	it("maps invalid code to error.unknown", () => {
		const err = ApiError.fromEnvelope({
			error: { code: "postgres: duplicate key", params: { detail: "secret" } },
		});
		assert.equal(err.code, "error.unknown");
		assert.equal(err.params, undefined);
	});
});
