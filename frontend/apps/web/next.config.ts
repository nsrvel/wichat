import type { NextConfig } from "next";
import { remotePatternsFromPublicUrls } from "./src/lib/image-remote-patterns";

const nextConfig: NextConfig = {
	transpilePackages: ["@wichat/contracts", "@wichat/api-client", "@wichat/i18n"],
	images: {
		remotePatterns: remotePatternsFromPublicUrls([process.env.NEXT_PUBLIC_OBJECT_PUBLIC_BASE_URL]),
	},
};

export default nextConfig;
