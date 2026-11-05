import { Inter } from "next/font/google";

/** Default WiChat UI font (Inter) — wired from root `layout.tsx`. */
export const appSans = Inter({
	subsets: ["latin"],
	weight: ["400", "500", "600", "700", "800"],
	display: "swap",
	variable: "--font-inter",
});
