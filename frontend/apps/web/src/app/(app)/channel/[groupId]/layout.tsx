import { GroupWorkspaceShellLayout } from "@/features/shell/group-workspace-shell-layout";

type LayoutProps = {
	children: React.ReactNode;
	params: Promise<{ groupId: string }>;
};

export default async function GroupWorkspaceLayout({ children, params }: LayoutProps) {
	const { groupId } = await params;

	return <GroupWorkspaceShellLayout groupId={groupId}>{children}</GroupWorkspaceShellLayout>;
}
