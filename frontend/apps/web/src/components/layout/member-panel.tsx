"use client";

import { useQuery } from "@tanstack/react-query";
import { useParams } from "next/navigation";
import { useTranslation } from "react-i18next";
import {
	Calendar,
	CheckCircle2,
	ChevronRight,
	Hash,
	Link2,
	ListFilter,
	Plus,
	Sparkles,
	Tag,
	User,
	X,
} from "lucide-react";
import { type ReactNode, useState } from "react";
import { GroupMemberProfileHost } from "@/components/group/group-member-profile-host";
import { OwnerCrownIcon } from "@/components/ui/owner-crown-icon";
import { UserAvatar } from "@/components/ui/user-avatar";
import { memberPresenceDot } from "@/features/presence/presence-dot";
import { groupMembersForPanel, resolveMemberNameColor } from "@/lib/member-role-display";
import type { UserPresenceStatus } from "@/lib/user-status";
import { userInitials } from "@/lib/user-initials";
import { myGroupsQueryOpts } from "@/services/group/group.query";
import { groupMembersQueryOpts } from "@/services/invite/invite.query";

const tabs = ["Info", "Pins", "Files", "Links"];

function InfoRow({
	icon,
	label,
	children,
	divider = true,
}: {
	icon: ReactNode;
	label: string;
	children: ReactNode;
	divider?: boolean;
}) {
	return (
		<div
			className={`flex items-center gap-3 py-[9px] ${divider ? "border-b border-[#EEEEEF]" : ""}`}
		>
			<span className="text-[#71717a]">{icon}</span>
			<span className="text-sm text-[#3f3f46]">{label}</span>
			<span className="ml-auto flex items-center gap-1.5">{children}</span>
		</div>
	);
}

function Badge({ tone, children }: { tone: "green" | "tan" | "blue"; children: ReactNode }) {
	const tones = {
		green: "bg-[#DFE8D3] text-[#4E6B41]",
		tan: "bg-[#E8D7CC] text-[#8A5F4D]",
		blue: "bg-[#D9E8ED] text-[#4A7285]",
	};
	return (
		<span className={`rounded-md px-2 py-[3px] text-xs font-medium ${tones[tone]}`}>
			{children}
		</span>
	);
}

const activity = [
	"#E5EDE2",
	"#DCE9DA",
	"#C7E0C9",
	"#A9D2AC",
	"#D5E6D5",
	"#E0EBDF",
	"#E9F0E7",
	"#CBE2CD",
	"#A9D2AC",
	"#D8E8D9",
	"#E4EEE2",
	"#EDF2EB",
	"#E2ECE1",
	"#C3DDC6",
	"#ABD4AE",
	"#CCE3CE",
	"#D9E9DA",
	"#E7EFE5",
	"#EAF1E8",
	"#E2ECE1",
];

function Member({
	avatar,
	name,
	nameColor,
	role,
	badge,
	dim,
	status,
	isOwner,
	onClick,
}: {
	avatar: string;
	name: string;
	nameColor?: string;
	role: string;
	badge?: ReactNode;
	dim?: boolean;
	status?: UserPresenceStatus | "offline";
	isOwner?: boolean;
	onClick?: () => void;
}) {
	const content = (
		<>
			<UserAvatar
				src={avatar}
				initials={userInitials(name)}
				size="lg"
				status={status ?? (dim ? "invisible" : "online")}
				statusRingClassName="ring-sidebar"
				imgClassName={dim ? "grayscale" : undefined}
			/>
			<div className="min-w-0">
				<div className="flex min-w-0 items-center gap-1.5">
					<div
						className="truncate text-sm font-semibold text-[#18181b]"
						style={nameColor ? { color: nameColor } : undefined}
					>
						{name}
					</div>
					{isOwner ? <OwnerCrownIcon label="Owner" /> : null}
				</div>
				{role ? <div className="truncate text-[13px] text-[#a1a1aa]">{role}</div> : null}
			</div>
			{badge ? <span className="ml-auto shrink-0">{badge}</span> : null}
		</>
	);

	if (onClick) {
		return (
			<button
				type="button"
				onClick={onClick}
				className={`-mx-1.5 flex w-[calc(100%+0.75rem)] cursor-pointer items-center gap-3 rounded-lg px-1.5 py-2 text-left transition-colors hover:bg-black/[0.04] dark:hover:bg-white/[0.05] ${dim ? "opacity-50" : ""}`}
			>
				{content}
			</button>
		);
	}

	return <div className={`flex items-center gap-3 py-2 ${dim ? "opacity-50" : ""}`}>{content}</div>;
}

export function MemberPanel({
	onClose,
	groupId: propGroupId,
}: {
	onClose?: () => void;
	groupId?: string;
}) {
	const { t } = useTranslation();
	const [selectedMemberId, setSelectedMemberId] = useState<string | null>(null);
	const params = useParams<{ groupId?: string }>();
	const groupId = propGroupId ?? (typeof params?.groupId === "string" ? params.groupId : "");
	const { data: myGroups = [] } = useQuery(myGroupsQueryOpts());
	const currentGroup = myGroups.find((g) => g.id === groupId);
	const hideRoles = currentGroup?.hide_roles ?? false;
	const membersQuery = useQuery({
		...groupMembersQueryOpts(groupId, ""),
		enabled: Boolean(groupId),
	});
	const realMembers = membersQuery.data?.members;
	const groupedMembers =
		groupId && realMembers
			? groupMembersForPanel(realMembers, hideRoles, {
					online: t("invite.member_online", "Online"),
					offline: t("invite.member_offline", "Offline"),
				})
			: null;
	const totalCount = membersQuery.data?.total ?? realMembers?.length ?? 9;
	return (
		<div className="flex h-full w-[300px] shrink-0 flex-col border-l border-sidebar-border bg-sidebar text-sidebar-foreground">
			<div className="flex shrink-0 items-center gap-2 px-3.5 pt-2.5 pb-1">
				<div className="flex flex-1 rounded-full bg-[#EFEFEF] p-[3px]">
					{tabs.map((tab) => (
						<button
							key={tab}
							type="button"
							className={`h-7 flex-1 rounded-full text-[13px] font-medium transition-colors ${
								tab === "Info"
									? "bg-white text-[#18181b] shadow-[0_1px_3px_rgba(0,0,0,0.10)]"
									: "text-[#52525b] hover:text-[#18181b]"
							}`}
						>
							{tab}
						</button>
					))}
				</div>
				{onClose ? (
					<button
						type="button"
						onClick={onClose}
						className="flex size-11 items-center justify-center rounded-lg text-[#52525b] hover:bg-[#efefef] lg:hidden"
						aria-label="Close info"
					>
						<X className="size-5" strokeWidth={1.8} />
					</button>
				) : null}
			</div>
			<div className="thin-scroll flex-1 overflow-y-auto px-4 pb-6">
				<h3 className="mt-3 mb-0.5 text-[15px] font-bold text-[#18181b]">Main info</h3>
				<InfoRow icon={<User className="size-4" strokeWidth={1.7} />} label="Creator">
					<UserAvatar src="/avatars/demo/andrew.jpg" initials="AM" size="xs" />
					<span className="text-sm font-medium text-[#18181b]">Andrew M.</span>
				</InfoRow>
				<InfoRow icon={<Calendar className="size-4" strokeWidth={1.7} />} label="Date of creation">
					<span className="text-sm text-[#3f3f46]">28 May</span>
				</InfoRow>
				<InfoRow icon={<Sparkles className="size-4" strokeWidth={1.7} />} label="Status">
					<span className="flex items-center gap-1.5 rounded-md bg-[#E3EBCF] px-2 py-[3px] text-xs font-medium text-[#5B7038]">
						<span className="size-1.5 rounded-full bg-[#6E8B3D]" />
						Active
					</span>
				</InfoRow>
				<InfoRow icon={<Tag className="size-4" strokeWidth={1.7} />} label="Tags">
					<span className="text-sm text-[#3f3f46]">13</span>
					<ChevronRight className="size-4 text-[#a1a1aa]" strokeWidth={2} />
				</InfoRow>
				<InfoRow
					icon={<CheckCircle2 className="size-4" strokeWidth={1.7} />}
					label="Tasks"
					divider={false}
				>
					<span className="text-sm text-[#3f3f46]">4</span>
					<ChevronRight className="size-4 text-[#a1a1aa]" strokeWidth={2} />
				</InfoRow>
				<h3 className="mt-5 mb-1.5 text-[15px] font-bold text-[#18181b]">Linked threads</h3>
				<div className="flex flex-col gap-1">
					<button
						type="button"
						className="flex h-11 items-center gap-2.5 rounded-lg px-1 hover:bg-[#f0f0f1] lg:h-[34px]"
					>
						<span className="flex size-[22px] items-center justify-center rounded-[7px] border border-[#ECECEC] bg-white">
							<Hash className="size-3 text-[#8e8e93]" strokeWidth={2.2} />
						</span>
						<span className="text-sm font-medium text-[#3f3f46]">Front-end</span>
						<span className="ml-auto flex h-[22px] min-w-[22px] items-center justify-center rounded-md bg-[#EBEBED] px-1.5 text-xs font-medium text-[#52525b]">
							4
						</span>
					</button>
					<button
						type="button"
						className="flex h-11 items-center gap-2.5 rounded-lg px-1 hover:bg-[#f0f0f1] lg:h-[34px]"
					>
						<span className="flex size-[22px] items-center justify-center rounded-[7px] border border-[#ECECEC] bg-white">
							<Hash className="size-3 text-[#8e8e93]" strokeWidth={2.2} />
						</span>
						<span className="text-sm font-medium text-[#3f3f46]">UI-kit design standards</span>
					</button>
				</div>
				<h3 className="mt-5 mb-2.5 text-[15px] font-bold text-[#18181b]">Thread activity</h3>
				<div className="flex gap-[3px]">
					{activity.map((color, index) => (
						<span
							key={index}
							className="h-[10px] w-[10px] rounded-[3px]"
							style={{ backgroundColor: color }}
						/>
					))}
				</div>
				<div className="mt-5 mb-0.5 flex items-center">
					<h3 className="text-[15px] font-bold text-[#18181b]">
						Members{" "}
						<span className="ml-1 text-[15px] font-medium text-[#a1a1aa]">{totalCount}</span>
					</h3>
					<div className="ml-auto flex items-center gap-0.5 text-[#52525b]">
						<button
							type="button"
							className="flex size-11 items-center justify-center rounded-lg hover:bg-[#f0f0f1] lg:size-7"
						>
							<Plus className="size-[18px]" strokeWidth={1.8} />
						</button>
						<button
							type="button"
							className="flex size-11 items-center justify-center rounded-lg hover:bg-[#f0f0f1] lg:size-7"
						>
							<Link2 className="size-4" strokeWidth={1.8} />
						</button>
						<button
							type="button"
							className="flex size-11 items-center justify-center rounded-lg hover:bg-[#f0f0f1] lg:size-7"
						>
							<ListFilter className="size-4" strokeWidth={1.8} />
						</button>
					</div>
				</div>
				<div className="flex flex-col">
					{groupedMembers ? (
						groupedMembers.map((section, sectionIndex) => (
							<div key={section.key}>
								{section.kind !== "online" || sectionIndex > 0 ? (
									<div className="mt-3 mb-1 text-[13px] font-medium text-[#a1a1aa]">
										{section.label} — {section.members.length}
									</div>
								) : null}
								{section.members.map((member) => {
									const displayName = member.display_name.trim() || `@${member.username}`;
									const nameColor = resolveMemberNameColor(member.roles, hideRoles);
									const primaryRole = !hideRoles
										? member.roles.find((r) => r.slug !== "everyone")
										: undefined;
									const subtitle = primaryRole
										? primaryRole.name
										: member.username
											? `@${member.username}`
											: "";
									return (
										<Member
											key={member.user_id}
											avatar={member.avatar_url}
											name={displayName}
											nameColor={nameColor}
											role={subtitle}
											isOwner={member.role === "owner"}
											status={memberPresenceDot(member)}
											dim={!member.is_online}
											onClick={() => setSelectedMemberId(member.user_id)}
										/>
									);
								})}
							</div>
						))
					) : (
						<>
							<Member
								avatar="/avatars/demo/daniel.jpg"
								name="Daniel Anderson"
								role="Art director"
								badge={<Badge tone="green">Design</Badge>}
							/>
							<Member
								avatar="/avatars/demo/andrew.jpg"
								name="Andrew Miller"
								role="Product owner"
								badge={<Badge tone="tan">Management</Badge>}
							/>
							<Member
								avatar="/avatars/demo/william.jpg"
								name="William Johnson"
								role="UX/UI designer"
								badge={<Badge tone="green">Design</Badge>}
							/>
							<Member
								avatar="/avatars/demo/emily.jpg"
								name="Emily Davis"
								role="Front-end dev"
								badge={<Badge tone="blue">Development</Badge>}
							/>
							<div className="mt-3 mb-1 text-[13px] text-[#a1a1aa]">Offline</div>
							<Member
								avatar="/avatars/demo/sophia.jpg"
								name="Sophia Wilson"
								role="Product designer"
								dim
							/>
						</>
					)}
				</div>
			</div>
			{groupId ? (
				<GroupMemberProfileHost
					groupId={groupId}
					memberId={selectedMemberId}
					onClose={() => setSelectedMemberId(null)}
				/>
			) : null}
		</div>
	);
}
