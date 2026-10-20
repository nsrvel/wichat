"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { z } from "zod";
import { TextField } from "@/components/form/text-field";
import { Button } from "@/components/ui/button";
import { Form } from "@/components/ui/form";
import { translateError } from "@/integrations/i18n/translate-error";
import { displayNameSchema } from "@wichat/contracts";
import * as groupQuery from "@/services/group/group.query";
import { useCreateGroupMutation, useJoinGroupMutation } from "@/services/group/group.mutation";

const createSchema = z.object({ name: displayNameSchema });

type Mode = "pick" | "create" | "join";

export function GroupChooser() {
	const { t } = useTranslation();
	const [mode, setMode] = useState<Mode>("pick");
	const joinable = useQuery({ ...groupQuery.joinableGroupsQueryOpts(), enabled: mode === "join" });
	const createGroup = useCreateGroupMutation();
	const joinGroup = useJoinGroupMutation();
	const form = useForm<z.infer<typeof createSchema>>({
		resolver: zodResolver(createSchema),
		defaultValues: { name: "" },
	});

	if (mode === "pick") {
		return (
			<div className="flex flex-1 flex-col items-center justify-center gap-6 p-8 text-center">
				<div className="max-w-md space-y-2">
					<h1 className="text-2xl font-semibold tracking-tight">{t("group.home.title")}</h1>
					<p className="text-muted-foreground text-sm">{t("group.home.lead")}</p>
				</div>
				<div className="flex w-full max-w-sm flex-col gap-3 sm:flex-row">
					<Button type="button" className="flex-1" size="lg" onClick={() => setMode("create")}>
						{t("group.home.create")}
					</Button>
					<Button
						type="button"
						variant="outline"
						className="flex-1"
						size="lg"
						onClick={() => setMode("join")}
					>
						{t("group.home.join")}
					</Button>
				</div>
			</div>
		);
	}

	if (mode === "create") {
		return (
			<div className="flex flex-1 flex-col items-center justify-center p-8">
				<div className="w-full max-w-sm space-y-6">
					<div className="space-y-1 text-center">
						<h2 className="text-xl font-semibold">{t("group.home.create_title")}</h2>
						<p className="text-muted-foreground text-sm">{t("group.home.create_lead")}</p>
					</div>
					<Form {...form}>
						<form
							className="space-y-4"
							onSubmit={form.handleSubmit((values) => {
								createGroup.mutate(values, {
									onError: (err) => {
										form.setError("root", { message: translateError(t, err) });
									},
								});
							})}
						>
							<TextField
								control={form.control}
								name="name"
								label={t("group.home.name")}
								placeholder={t("group.home.name_placeholder")}
								autoComplete="organization"
								required
							/>
							{form.formState.errors.root?.message ? (
								<p className="text-destructive text-sm">{form.formState.errors.root.message}</p>
							) : null}
							<div className="flex gap-2">
								<Button type="button" variant="ghost" onClick={() => setMode("pick")}>
									{t("auth.back")}
								</Button>
								<Button type="submit" className="flex-1" disabled={createGroup.isPending}>
									{createGroup.isPending ? t("group.home.creating") : t("group.home.create_submit")}
								</Button>
							</div>
						</form>
					</Form>
				</div>
			</div>
		);
	}

	const groups = joinable.data ?? [];
	const loading = joinable.isPending;

	return (
		<div className="flex flex-1 flex-col items-center justify-center p-8">
			<div className="w-full max-w-md space-y-6">
				<div className="space-y-1 text-center">
					<h2 className="text-xl font-semibold">{t("group.home.join_title")}</h2>
					<p className="text-muted-foreground text-sm">{t("group.home.join_lead")}</p>
				</div>
				{loading ? (
					<p className="text-muted-foreground text-center text-sm">{t("group.home.loading")}</p>
				) : groups.length === 0 ? (
					<div className="space-y-4 text-center">
						<p className="text-muted-foreground text-sm">{t("group.home.join_empty")}</p>
						<Button type="button" onClick={() => setMode("create")}>
							{t("group.home.create")}
						</Button>
					</div>
				) : (
					<ul className="divide-y rounded-lg border">
						{groups.map((g) => (
							<li key={g.id} className="flex items-center justify-between gap-3 p-4">
								<span className="font-medium">{g.name}</span>
								<Button
									type="button"
									size="sm"
									disabled={joinGroup.isPending}
									onClick={() => joinGroup.mutate(g.id)}
								>
									{t("group.home.join_action")}
								</Button>
							</li>
						))}
					</ul>
				)}
				<div className="flex justify-center">
					<Button type="button" variant="ghost" onClick={() => setMode("pick")}>
						{t("auth.back")}
					</Button>
				</div>
			</div>
		</div>
	);
}
