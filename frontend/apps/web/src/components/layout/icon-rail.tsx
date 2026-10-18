import { Plus } from "lucide-react";
import Image from "next/image";

function SwirlIcon() {
	return (
		<svg viewBox="0 0 24 24" className="h-6 w-6" fill="currentColor" aria-hidden>
			<path d="M12 2c5.52 0 10 4.48 10 10h-4a6 6 0 0 0-6-6V2z" />
			<path d="M12 22C6.48 22 2 17.52 2 12h4a6 6 0 0 0 6 6v4z" />
		</svg>
	);
}

function OrbitIcon() {
	return (
		<svg viewBox="0 0 24 24" className="h-6 w-6" fill="currentColor" aria-hidden>
			<path d="M12 3a9 9 0 1 0 9 9h-4.2A4.8 4.8 0 1 1 12 7.2V3z" />
		</svg>
	);
}

export function IconRail({ onLogout }: { onLogout?: () => void }) {
	return (
		<div className="hidden w-16 shrink-0 flex-col items-center border-r border-[#E7E8E8] bg-[#F0F1F1] pt-3 pb-4 lg:flex">
			<div className="mb-4 ml-3 flex items-center gap-2 self-start px-1">
				<span className="h-3 w-3 rounded-full border border-[#E0443E]/40 bg-[#FF5F57]" />
				<span className="h-3 w-3 rounded-full border border-[#D89E24]/40 bg-[#FEBC2E]" />
				<span className="h-3 w-3 rounded-full border border-[#1DAD2B]/40 bg-[#28C840]" />
			</div>
			<button
				type="button"
				className="flex h-11 w-11 items-center justify-center rounded-[14px] bg-[#141414] text-white shadow-sm"
			>
				<span className="-mt-0.5 text-[26px] leading-none font-extrabold tracking-tight">C</span>
			</button>
			<div className="mt-5 flex flex-col items-center gap-3">
				<button
					type="button"
					className="flex h-11 w-11 items-center justify-center rounded-[14px] border border-[#ECECEC] bg-white text-[#1c1c1e] shadow-[0_1px_2px_rgba(0,0,0,0.06)]"
				>
					<SwirlIcon />
				</button>
				<button
					type="button"
					className="flex h-11 w-11 items-center justify-center rounded-[14px] border border-[#ECECEC] bg-white text-[#1c1c1e] shadow-[0_1px_2px_rgba(0,0,0,0.06)]"
				>
					<OrbitIcon />
				</button>
				<button
					type="button"
					className="flex h-11 w-11 items-center justify-center rounded-[14px] bg-[#E4E5E6] text-[#3f3f46] transition-colors hover:bg-[#dadbdd]"
				>
					<Plus className="h-5 w-5" strokeWidth={2} />
				</button>
			</div>
			<div className="mt-auto">
				<button type="button" onClick={onLogout} aria-label="Sign out">
					<Image
						src="/avatars/demo/you.jpg"
						alt=""
						width={40}
						height={40}
						className="h-10 w-10 rounded-full object-cover grayscale ring-2 ring-white/70"
					/>
				</button>
			</div>
		</div>
	);
}
