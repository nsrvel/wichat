import {
	Bookmark,
	ChevronDown,
	CornerDownRight,
	FileText,
	Hash,
	Inbox,
	Lock,
	MessageSquareMore,
	Plus,
	Search,
	Sparkles,
} from "lucide-react";
import Image from "next/image";
import type { ReactNode } from "react";

function CountPill({ n }: { n: number }) {
	return (
		<span className="ml-auto flex h-5 min-w-[20px] items-center justify-center rounded-[5px] bg-[#EBEBED] px-1 text-[11px] font-medium text-[#52525b] lg:h-[18px] lg:min-w-[18px]">
			{n}
		</span>
	);
}

function NavItem({
	icon,
	label,
	badge,
	count,
	plainCount,
}: {
	icon: ReactNode;
	label: string;
	badge?: boolean;
	count?: number;
	plainCount?: number;
}) {
	return (
		<button
			type="button"
			className="flex h-11 w-full items-center gap-2.5 rounded-md px-2.5 text-[#3f3f46] transition-colors hover:bg-[#f0f0f1] lg:h-[27px]"
		>
			<span className="text-[#6b6b70]">{icon}</span>
			<span className="truncate text-sm font-medium lg:text-[13px]">{label}</span>
			{badge ? (
				<span className="rounded bg-[#F2D9F6] px-1 py-px text-[9px] font-semibold tracking-wide text-[#A855C8]">
					NEW
				</span>
			) : null}
			{count !== undefined ? <CountPill n={count} /> : null}
			{plainCount !== undefined ? (
				<span className="ml-auto text-xs text-[#71717a]">{plainCount}</span>
			) : null}
		</button>
	);
}

function ChannelIcon({ children }: { children: ReactNode }) {
	return (
		<span className="flex h-5 w-5 shrink-0 items-center justify-center rounded-[6px] border border-[#ECECEC] bg-white text-[10px] shadow-[0_1px_1px_rgba(0,0,0,0.03)]">
			{children}
		</span>
	);
}

function ChannelItem({
	icon,
	label,
	count,
	selected,
	indent,
}: {
	icon: ReactNode;
	label: string;
	count?: number;
	selected?: boolean;
	indent?: boolean;
}) {
	return (
		<button
			type="button"
			className={`flex h-11 w-full items-center gap-2.5 rounded-md text-sm font-medium transition-colors lg:h-[30px] lg:text-[13px] ${
				indent ? "pr-2.5 pl-8" : "px-2.5"
			} ${selected ? "bg-[#E9E9EB] text-[#18181b]" : "text-[#3f3f46] hover:bg-[#f0f0f1]"}`}
		>
			{icon}
			<span className="truncate">{label}</span>
			{count !== undefined ? <CountPill n={count} /> : null}
		</button>
	);
}

const hashIcon = <Hash className="h-[11px] w-[11px] text-[#8e8e93]" strokeWidth={2.2} />;
const hashLockIcon = (
	<span className="relative flex items-center justify-center">
		<Hash className="h-[11px] w-[11px] text-[#8e8e93]" strokeWidth={2.2} />
		<Lock
			className="absolute -right-[4px] -bottom-[3px] h-[6px] w-[6px] fill-[#8e8e93] text-[#8e8e93]"
			strokeWidth={0}
		/>
	</span>
);
const nestedIcon = (
	<ChannelIcon>
		<CornerDownRight className="h-[11px] w-[11px] text-[#8e8e93]" strokeWidth={2.2} />
	</ChannelIcon>
);

export function ChannelSidebar() {
	return (
		<div className="flex h-full w-[248px] shrink-0 flex-col border-r border-[#EEEEEF] bg-[#F9F9F9]">
			<div className="flex h-14 shrink-0 items-center justify-between px-4 lg:h-[46px]">
				<button
					type="button"
					className="flex items-center gap-1 text-[16px] font-bold tracking-tight text-[#18181b] lg:text-[15px]"
				>
					Conceptzilla
					<ChevronDown className="h-3.5 w-3.5 text-[#52525b]" strokeWidth={2.4} />
				</button>
				<button
					type="button"
					className="flex h-9 w-9 items-center justify-center rounded-lg text-[#3f3f46] hover:bg-[#f0f0f1]"
				>
					<Search className="h-4 w-4" strokeWidth={2} />
				</button>
			</div>
			<div className="thin-scroll flex-1 overflow-y-auto px-2 pb-4">
				<nav className="flex flex-col gap-px">
					<NavItem
						icon={<Sparkles className="h-[15px] w-[15px]" strokeWidth={1.7} />}
						label="Assistant"
						badge
					/>
					<NavItem
						icon={<FileText className="h-[15px] w-[15px]" strokeWidth={1.7} />}
						label="Drafts"
					/>
					<NavItem
						icon={<Bookmark className="h-[15px] w-[15px]" strokeWidth={1.7} />}
						label="Saved items"
					/>
					<NavItem
						icon={<Inbox className="h-[15px] w-[15px]" strokeWidth={1.7} />}
						label="Inbox"
						plainCount={8}
					/>
					<NavItem
						icon={<MessageSquareMore className="h-[15px] w-[15px]" strokeWidth={1.7} />}
						label="Direct messages"
						count={1}
					/>
				</nav>
				<div className="mt-3.5">
					<button
						type="button"
						className="flex h-6 w-full items-center justify-between px-2.5 text-xs font-semibold text-[#52525b]"
					>
						Favorites
						<ChevronDown className="h-3 w-3" strokeWidth={2.4} />
					</button>
					<div className="mt-0.5 flex flex-col gap-px">
						<button
							type="button"
							className="flex h-11 w-full items-center gap-2.5 rounded-md px-2.5 text-[#3f3f46] transition-colors hover:bg-[#f0f0f1] lg:h-[30px]"
						>
							<Image
								src="/avatars/demo/sophia.jpg"
								alt=""
								width={20}
								height={20}
								className="h-5 w-5 rounded-full object-cover"
							/>
							<span className="truncate text-sm font-medium lg:text-[13px]">Sophia Wilson</span>
							<CountPill n={2} />
						</button>
						<ChannelItem icon={<ChannelIcon>{hashIcon}</ChannelIcon>} label="Front-end" count={4} />
					</div>
				</div>
				<div className="mt-3.5">
					<div className="flex h-6 w-full items-center justify-between px-2.5">
						<span className="text-xs font-semibold text-[#52525b]">Channels</span>
						<span className="flex items-center gap-0.5 text-[#71717a]">
							<button
								type="button"
								className="flex h-6 w-6 items-center justify-center rounded-md hover:bg-[#f0f0f1]"
							>
								<Plus className="h-3.5 w-3.5" strokeWidth={2.2} />
							</button>
							<button
								type="button"
								className="flex h-6 w-6 items-center justify-center rounded-md hover:bg-[#f0f0f1]"
							>
								<ChevronDown className="h-3 w-3" strokeWidth={2.4} />
							</button>
						</span>
					</div>
					<div className="mt-0.5 flex flex-col gap-px">
						<ChannelItem
							icon={
								<ChannelIcon>
									<span>🔥</span>
								</ChannelIcon>
							}
							label="General"
							count={1}
						/>
						<ChannelItem icon={<ChannelIcon>{hashIcon}</ChannelIcon>} label="Front-end" count={4} />
						<ChannelItem icon={<ChannelIcon>{hashLockIcon}</ChannelIcon>} label="Website" />
						<ChannelItem
							icon={
								<ChannelIcon>
									<span>💥</span>
								</ChannelIcon>
							}
							label="v3.0"
						/>
						<ChannelItem icon={nestedIcon} label="Wireframe" indent />
						<ChannelItem icon={nestedIcon} label="Design" indent />
						<ChannelItem icon={nestedIcon} label="UI-kit design" indent selected />
						<ChannelItem
							icon={<ChannelIcon>{hashIcon}</ChannelIcon>}
							label="v2.0 - actual version"
						/>
						<ChannelItem icon={<ChannelIcon>{hashLockIcon}</ChannelIcon>} label="Strategy" />
						<ChannelItem
							icon={
								<ChannelIcon>
									<span>🎈</span>
								</ChannelIcon>
							}
							label="Events"
						/>
						<ChannelItem icon={<ChannelIcon>{hashIcon}</ChannelIcon>} label="Announcements" />
						<ChannelItem icon={<ChannelIcon>{hashIcon}</ChannelIcon>} label="UI/UX" count={2} />
					</div>
				</div>
			</div>
		</div>
	);
}
