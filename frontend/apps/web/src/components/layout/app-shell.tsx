"use client";

import { useQuery } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { cloneElement, isValidElement, useState, type ReactElement, type ReactNode } from "react";
import { ChannelSidebar } from "@/components/layout/channel-sidebar";
import { IconRail } from "@/components/layout/icon-rail";
import { MemberPanel } from "@/components/layout/member-panel";
import { authQuery } from "@/services/auth";
import { useLogoutMutation } from "@/services/auth/auth.mutation";

type AppShellProps = {
	children: ReactNode;
};

export function AppShell({ children }: AppShellProps) {
	const [sidebarOpen, setSidebarOpen] = useState(false);
	const [memberOpen, setMemberOpen] = useState(false);
	const router = useRouter();
	const { data: user } = useQuery(authQuery.sessionQueryOpts());
	const logout = useLogoutMutation();
	const onLogout = user
		? () => {
				logout.mutate(undefined, { onSettled: () => router.replace("/sign-in") });
			}
		: undefined;

	const main = isValidElement(children)
		? cloneElement(children as ReactElement<ShellMainChildProps>, {
				onOpenSidebar: () => setSidebarOpen(true),
				onOpenMemberPanel: () => setMemberOpen(true),
			})
		: children;

	return (
		<div className="flex h-full min-h-0 w-full overflow-hidden bg-white text-[#18181b]">
			<IconRail onLogout={onLogout} />

			<div className="hidden h-full lg:block">
				<ChannelSidebar />
			</div>

			{sidebarOpen ? (
				<div className="fixed inset-0 z-40 lg:hidden">
					<button
						type="button"
						className="absolute inset-0 bg-black/30"
						aria-label="Close channels"
						onClick={() => setSidebarOpen(false)}
					/>
					<div className="absolute top-0 left-0 h-full shadow-2xl">
						<ChannelSidebar />
					</div>
				</div>
			) : null}

			<div className="flex min-h-0 min-w-0 flex-1 flex-col">{main}</div>

			<div className="hidden h-full lg:block">
				<MemberPanel />
			</div>

			{memberOpen ? (
				<div className="fixed inset-0 z-40 lg:hidden">
					<button
						type="button"
						className="absolute inset-0 bg-black/30"
						aria-label="Close thread details"
						onClick={() => setMemberOpen(false)}
					/>
					<div className="absolute top-0 right-0 h-full shadow-2xl">
						<MemberPanel onClose={() => setMemberOpen(false)} />
					</div>
				</div>
			) : null}
		</div>
	);
}

/** Props injected into the main chat view when it accepts shell callbacks. */
export type ShellMainChildProps = {
	onOpenSidebar?: () => void;
	onOpenMemberPanel?: () => void;
};
