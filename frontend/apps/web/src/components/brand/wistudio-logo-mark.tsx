import Image from "next/image";
import { cn } from "@/lib/utils";

type WiStudioLogoMarkProps = {
	size?: number;
	className?: string;
};

export function WiStudioLogoMark({ size = 32, className }: WiStudioLogoMarkProps) {
	return (
		<span
			className={cn("relative block shrink-0 overflow-hidden rounded-[13px]", className)}
			style={{ width: size, height: size }}
		>
			<Image
				src="/logo_dark.jpg"
				alt=""
				fill
				sizes={`${size}px`}
				className="object-cover dark:hidden"
			/>
			<Image
				src="/logo_light.jpg"
				alt=""
				fill
				sizes={`${size}px`}
				aria-hidden
				className="hidden object-cover dark:block"
			/>
		</span>
	);
}
