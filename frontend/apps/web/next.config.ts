import type { NextConfig } from "next";

const nextConfig: NextConfig = {
	transpilePackages: ["@wichat/contracts", "@wichat/api-client", "@wichat/i18n"],
};

export default nextConfig;
