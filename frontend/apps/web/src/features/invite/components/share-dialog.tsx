"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { WiChatDialog, WiChatDialogBody } from "@/components/ui/wichat-dialog";
import { GroupAccessList } from "@/features/invite/components/group-access-list";
import { ShareDialogFooter } from "@/features/invite/components/share-dialog-footer";
import { ShareDialogHeader } from "@/features/invite/components/share-dialog-header";
import {
	InviteRecipientActions,
	InviteRecipientField,
} from "@/features/invite/components/invite-recipient-field";
import type { InviteTarget } from "@/features/invite/lib/invite-target";
import type { RecipientChip } from "@/features/invite/lib/recipient-chip";
import { parseEmailChip } from "@/features/invite/lib/recipient-chip";
import { shareDialogShellClass } from "@/features/invite/lib/share-dialog-chrome";
import { ShareDialogSection } from "@/features/invite/lib/share-dialog-section";
import { instancePolicyQueryOpts } from "@/services/auth/auth.query";
import { Q_MY_GROUPS } from "@/services/group/group.key";
import { copyTextToClipboard } from "@/lib/copy-to-clipboard";
import { searchUsers, sendGroupInvites } from "@/services/invite/invite.service";
import {
	Q_INVITE_MEMBERS,
	groupInviteLinkQueryOpts,
	groupMembersQueryOpts,
} from "@/services/invite/invite.query";

const SHARE_DIALOG_TITLE_ID = "wichat-share-dialog-title";

type ShareDialogProps = {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	target: InviteTarget | null;
};

export function ShareDialog({ open, onOpenChange, target }: ShareDialogProps) {
	const { t } = useTranslation();
	const queryClient = useQueryClient();
	const [chips, setChips] = useState<RecipientChip[]>([]);
	const [query, setQuery] = useState("");
	const [debouncedQuery, setDebouncedQuery] = useState("");
	const [cursorStack, setCursorStack] = useState<string[]>([""]);
	const [linkCopied, setLinkCopied] = useState(false);

	const groupId = target?.scope === "group" ? target.groupId : target?.groupId;
	const channelId = target?.scope === "channel" ? target.channelId : undefined;

	useEffect(() => {
		const id = window.setTimeout(() => setDebouncedQuery(query.trim()), 300);
		return () => window.clearTimeout(id);
	}, [query]);

	const cursor = cursorStack[cursorStack.length - 1] ?? "";

	const { data: policy } = useQuery({ ...instancePolicyQueryOpts(), enabled: open });
	// ponytail: wire from group invite settings when they exist
	const canInviteAndCopy = true;

	const membersQuery = useQuery({
		...groupMembersQueryOpts(groupId!, cursor),
		enabled: open && Boolean(groupId),
	});

	const searchQuery = useQuery({
		queryKey: ["invite", "search", groupId, debouncedQuery],
		queryFn: () =>
			searchUsers(debouncedQuery, {
				excludeGroupId: groupId,
				limit: 20,
			}),
		enabled:
			open &&
			Boolean(groupId) &&
			debouncedQuery.length > 0 &&
			parseEmailChip(debouncedQuery) === null,
	});

	const inviteLinkQuery = useQuery({
		...groupInviteLinkQueryOpts(groupId!, channelId),
		enabled: open && Boolean(groupId),
	});

	const sendMutation = useMutation({
		mutationFn: () => {
			const userIds = chips.filter((c) => c.kind === "user").map((c) => c.id);
			const emails = chips.filter((c) => c.kind === "email").map((c) => c.email);
			const pendingEmail = parseEmailChip(query);
			if (pendingEmail && !emails.includes(pendingEmail.email)) {
				emails.push(pendingEmail.email);
			}
			return sendGroupInvites(groupId!, {
				user_ids: userIds,
				emails,
				channel_id: channelId,
			});
		},
		onSuccess: (result) => {
			setChips([]);
			setQuery("");
			void queryClient.invalidateQueries({ queryKey: [Q_INVITE_MEMBERS, groupId] });
			void queryClient.invalidateQueries({ queryKey: [Q_MY_GROUPS] });
			toast.success(
				t("invite.share_success", {
					added: result.added,
					emailed: result.emailed,
				}),
			);
		},
		onError: () => {
			toast.error(t("invite.share_error"));
		},
	});

	const finishCopy = (ok: boolean) => {
		if (ok) {
			setLinkCopied(true);
			toast.success(t("invite.copy_success"));
			return;
		}
		toast.error(t("invite.copy_error"));
	};

	const handleCopyLink = () => {
		const url = inviteLinkQuery.data?.url;
		if (url) {
			void copyTextToClipboard(url).then(finishCopy);
			return;
		}
		void inviteLinkQuery.refetch().then((result) => {
			const next = result.data?.url;
			if (!next) {
				toast.error(t("invite.copy_error"));
				return;
			}
			void copyTextToClipboard(next).then(finishCopy);
		});
	};

	const membersFilterQuery = useMemo(() => {
		if (parseEmailChip(query) || parseEmailChip(debouncedQuery)) {
			return "";
		}
		return debouncedQuery;
	}, [debouncedQuery, query]);

	const pendingEmailInvite = parseEmailChip(query);
	const shareDisabled = chips.length === 0 && !pendingEmailInvite;

	if (!target || !groupId) {
		return null;
	}

	const membersPage = membersQuery.data;
	const members = membersPage?.members ?? [];
	const membersPending = membersQuery.isPending && !membersPage;
	const membersFailed = membersQuery.isError;
	const hasNext = Boolean(membersPage?.next_cursor);
	const hasPrev = cursorStack.length > 1;

	const handleOpenChange = (next: boolean) => {
		if (!next) {
			setChips([]);
			setQuery("");
			setDebouncedQuery("");
			setCursorStack([""]);
			setLinkCopied(false);
		}
		onOpenChange(next);
	};

	return (
		<WiChatDialog
			open={open}
			onOpenChange={handleOpenChange}
			titleId={SHARE_DIALOG_TITLE_ID}
			className={shareDialogShellClass}
		>
			<ShareDialogHeader
				target={target}
				titleId={SHARE_DIALOG_TITLE_ID}
				canInviteAndCopy={canInviteAndCopy}
			/>
			<WiChatDialogBody className="space-y-6 pt-1">
				{canInviteAndCopy ? (
					<div className="flex flex-col gap-2 sm:flex-row sm:items-center">
						<InviteRecipientField
							chips={chips}
							onChangeChips={setChips}
							searchHits={searchQuery.data?.users ?? []}
							searchLoading={searchQuery.isFetching}
							query={query}
							onQueryChange={setQuery}
							onPickUser={(hit) => {
								const chip = {
									kind: "user" as const,
									id: hit.user_id,
									label: hit.display_name || hit.email,
									email: hit.email,
								};
								if (!chips.some((c) => c.kind === "user" && c.id === chip.id)) {
									setChips([...chips, chip]);
								}
								setQuery("");
							}}
							disabled={sendMutation.isPending}
						/>
						<InviteRecipientActions
							onShare={() => sendMutation.mutate()}
							shareDisabled={shareDisabled}
							pending={sendMutation.isPending}
						/>
					</div>
				) : null}

				<ShareDialogSection label={t("invite.section_members")}>
					<GroupAccessList
						members={members}
						loading={membersPending || membersQuery.isFetching}
						failed={membersFailed}
						filterQuery={membersFilterQuery}
						hasPrev={hasPrev}
						hasNext={hasNext}
						onPrev={() => setCursorStack((s) => (s.length > 1 ? s.slice(0, -1) : s))}
						onNext={() => {
							const next = membersQuery.data?.next_cursor;
							if (next) {
								setCursorStack((s) => [...s, next]);
							}
						}}
					/>
				</ShareDialogSection>
			</WiChatDialogBody>
			<ShareDialogFooter
				currentMode={policy?.registration_mode}
				showCopy={canInviteAndCopy}
				copyDisabled={linkCopied || inviteLinkQuery.isLoading}
				linkCopied={linkCopied}
				onCopyLink={handleCopyLink}
			/>
		</WiChatDialog>
	);
}
