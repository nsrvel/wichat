/** @type {import('lint-staged').Configuration} */
export default {
	"apps/web/**/*.{ts,tsx,mjs,cjs}": [
		"pnpm --filter web-client exec eslint --fix --max-warnings 0",
	],
	"apps/web/**/*.{json,css,md,yml,yaml}": [
		"pnpm --filter web-client exec prettier --write",
	],
	"packages/**/*.{ts,tsx,json}": [
		"pnpm --filter web-client exec prettier --write",
	],
};
