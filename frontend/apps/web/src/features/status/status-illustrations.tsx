import Image from "next/image";

export type StatusScreenVariant =
	| "not-found"
	| "forbidden"
	| "unauthorized"
	| "error"
	| "maintenance";

const ILLUSTRATION_ASSETS: Record<
	StatusScreenVariant,
	{ src: string; width: number; height: number }
> = {
	"not-found": { src: "/errors/error-404.png", width: 1648, height: 1648 },
	forbidden: { src: "/errors/error-403.png", width: 1681, height: 1301 },
	unauthorized: { src: "/errors/error-401.png", width: 1680, height: 1328 },
	error: { src: "/errors/error-500.png", width: 1651, height: 1228 },
	maintenance: { src: "/errors/error-503.png", width: 1681, height: 1328 },
};

export function StatusIllustration({ variant }: { variant: StatusScreenVariant }) {
	const asset = ILLUSTRATION_ASSETS[variant];

	return (
		<div aria-hidden className="flex items-center justify-center">
			<Image
				src={asset.src}
				alt=""
				width={asset.width}
				height={asset.height}
				sizes="(max-width: 768px) 224px, 288px"
				priority
				draggable={false}
				className="h-auto w-56 select-none object-contain drop-shadow-[0_18px_36px_rgba(0,0,0,0.45)] md:w-72"
			/>
		</div>
	);
}
