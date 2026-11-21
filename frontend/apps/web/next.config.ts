import { loadEnvConfig } from "@next/env";
import type { NextConfig } from "next";
import type { RemotePattern } from "next/dist/shared/lib/image-config";
import { remotePatternsFromPublicUrls } from "./src/lib/image-remote-patterns";

// next.config runs before Next injects .env.local into process.env — load explicitly for images.remotePatterns.
loadEnvConfig(process.cwd());

const gatewayProxy = process.env.API_GATEWAY_PROXY_TARGET?.trim() || "http://127.0.0.1:3000";

const isDev = process.env.NODE_ENV === "development";

const devMinioImagePatterns: RemotePattern[] = [
	{ protocol: "http", hostname: "localhost", port: "9000", pathname: "/wichat-public/**" },
	{ protocol: "http", hostname: "127.0.0.1", port: "9000", pathname: "/wichat-public/**" },
];

const apiUrl = process.env.NEXT_PUBLIC_API_URL?.trim() || "http://127.0.0.1:3000";
const apiOrigin = new URL(apiUrl).origin;
const wsOrigin = apiOrigin.replace(/^http/, "ws");

const contentSecurityPolicy = [
	"default-src 'self'",
	"base-uri 'self'",
	"object-src 'none'",
	"frame-ancestors 'self'",
	`script-src 'self' 'unsafe-inline'${isDev ? " 'unsafe-eval'" : ""} https://challenges.cloudflare.com`,
	"style-src 'self' 'unsafe-inline'",
	`connect-src 'self' ${apiOrigin} ${wsOrigin} https://challenges.cloudflare.com`,
	"img-src 'self' data: blob: http://127.0.0.1:9000 http://localhost:9000",
	"font-src 'self'",
	"frame-src https://challenges.cloudflare.com",
].join("; ");

const nextConfig: NextConfig = {
	allowedDevOrigins: ["wichat.wistudio.my.id"],
	transpilePackages: ["@wichat/contracts", "@wichat/api-client", "@wichat/i18n"],
	...(isDev
		? {
				onDemandEntries: {
					maxInactiveAge: 5 * 60 * 1000,
					pagesBufferLength: 8,
				},
			}
		: {}),
	images: {
		remotePatterns: [
			...(isDev ? devMinioImagePatterns : []),
			...remotePatternsFromPublicUrls([process.env.NEXT_PUBLIC_OBJECT_PUBLIC_BASE_URL]),
		],
		// Self-hosted MinIO on loopback — required when src is absolute http://localhost:9000/...
		dangerouslyAllowLocalIP: isDev,
	},
	async headers() {
		return [
			{
				source: "/(.*)",
				headers: [
					{ key: "Content-Security-Policy", value: contentSecurityPolicy },
					{ key: "Referrer-Policy", value: "strict-origin-when-cross-origin" },
					{ key: "X-Content-Type-Options", value: "nosniff" },
					{ key: "Permissions-Policy", value: "camera=(), microphone=(), geolocation=()" },
				],
			},
		];
	},
	async rewrites() {
		if (!isDev) {
			return [];
		}
		return [
			// Public object bucket (group icons) — same-origin for next/image + <img>
			{
				source: "/wichat-public/:path*",
				destination: "http://127.0.0.1:9000/wichat-public/:path*",
			},
			// Dev: browser calls same origin (:8080) — no CORS preflight for DELETE etc.
			{ source: "/api/:path*", destination: `${gatewayProxy}/api/:path*` },
		];
	},
};

export default nextConfig;
