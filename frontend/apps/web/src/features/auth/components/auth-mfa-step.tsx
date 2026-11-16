"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import {
	startAuthentication,
	type PublicKeyCredentialRequestOptionsJSON,
} from "@simplewebauthn/browser";
import type { AuthChallenge } from "@wichat/contracts/auth";
import type { MfaMethod } from "@wichat/contracts/auth/mfa";
import { KeyRound, ShieldCheck, Smartphone } from "lucide-react";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { z } from "zod";
import { FieldLabelContent } from "@/components/form/field-label";
import { OtpField } from "@/components/form/otp-field";
import { Button } from "@/components/ui/button";
import {
	Form,
	FormControl,
	FormField,
	FormItem,
	FormLabel,
	FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { translateError } from "@/integrations/i18n/translate-error";
import { authFormClass, authSubmitClass } from "@/lib/form-field-styles";
import { cn } from "@/lib/utils";
import {
	useBeginLoginWebAuthnMutation,
	useVerifyLoginMfaMutation,
} from "@/services/auth/auth.mutation";

const TotpSchema = z.object({
	code: z.string().regex(/^\d{6}$/, "validation.otp"),
});

const RecoverySchema = z.object({
	code: z.string().trim().min(4, "validation.required").max(32),
});

type AuthMfaStepProps = {
	challenge: AuthChallenge;
	onSuccess: () => void;
	onBack: () => void;
};

function unwrapPublicKeyOptions(raw: unknown): PublicKeyCredentialRequestOptionsJSON | null {
	if (!raw || typeof raw !== "object") {
		return null;
	}
	const record = raw as Record<string, unknown>;
	if (record.publicKey && typeof record.publicKey === "object") {
		return record.publicKey as PublicKeyCredentialRequestOptionsJSON;
	}
	return record as unknown as PublicKeyCredentialRequestOptionsJSON;
}

export function AuthMfaStep({ challenge, onSuccess, onBack }: AuthMfaStepProps) {
	const { t } = useTranslation();
	const methods = useMemo<MfaMethod[]>(() => {
		const list = challenge.mfa_methods ?? [];
		return list.length > 0 ? list : ["totp", "recovery"];
	}, [challenge.mfa_methods]);

	const [activeMethod, setActiveMethod] = useState<MfaMethod>(
		methods.includes("webauthn") ? "webauthn" : methods.includes("totp") ? "totp" : "recovery",
	);
	const [webauthnError, setWebauthnError] = useState<string | null>(null);
	const [webauthnBusy, setWebauthnBusy] = useState(false);
	const autoTriggeredRef = useRef(false);

	const beginWebAuthn = useBeginLoginWebAuthnMutation();
	const verifyMfa = useVerifyLoginMfaMutation();

	const totpForm = useForm<z.infer<typeof TotpSchema>>({
		resolver: zodResolver(TotpSchema),
		defaultValues: { code: "" },
	});

	const recoveryForm = useForm<z.infer<typeof RecoverySchema>>({
		resolver: zodResolver(RecoverySchema),
		defaultValues: { code: "" },
	});

	const runWebAuthn = useCallback(async () => {
		setWebauthnError(null);
		setWebauthnBusy(true);
		try {
			let options = unwrapPublicKeyOptions(challenge.webauthn_options);
			if (!options) {
				const res = await beginWebAuthn.mutateAsync(challenge.challenge_id);
				options = unwrapPublicKeyOptions(res.webauthn_options);
			}
			if (!options) {
				setWebauthnError(t("login.mfa.webauthn_failed"));
				return;
			}
			const assertion = await startAuthentication({ optionsJSON: options });
			await verifyMfa.mutateAsync({
				challenge_id: challenge.challenge_id,
				method: "webauthn",
				assertion,
			});
			onSuccess();
		} catch (err) {
			if (err instanceof Error && err.name === "NotAllowedError") {
				setWebauthnError(t("login.mfa.webauthn_failed"));
			} else {
				setWebauthnError(translateError(t, err) || t("login.mfa.webauthn_failed"));
			}
		} finally {
			setWebauthnBusy(false);
		}
	}, [beginWebAuthn, challenge.challenge_id, challenge.webauthn_options, onSuccess, t, verifyMfa]);

	useEffect(() => {
		if (activeMethod === "webauthn" && !autoTriggeredRef.current && challenge.webauthn_options) {
			autoTriggeredRef.current = true;
			void runWebAuthn();
		}
	}, [activeMethod, challenge.webauthn_options, runWebAuthn]);

	const handleTotpSubmit = totpForm.handleSubmit(async ({ code }) => {
		verifyMfa.reset();
		try {
			await verifyMfa.mutateAsync({
				challenge_id: challenge.challenge_id,
				method: "totp",
				code,
			});
			onSuccess();
		} catch {
			// handled via verifyMfa.error
		}
	});

	const handleRecoverySubmit = recoveryForm.handleSubmit(async ({ code }) => {
		verifyMfa.reset();
		try {
			await verifyMfa.mutateAsync({
				challenge_id: challenge.challenge_id,
				method: "recovery",
				code,
			});
			onSuccess();
		} catch {
			// handled via verifyMfa.error
		}
	});

	const serverError =
		activeMethod === "webauthn"
			? webauthnError
			: verifyMfa.isError
				? translateError(t, verifyMfa.error)
				: null;

	return (
		<div className={authFormClass}>
			<div className="space-y-1 text-center">
				<h2 className="text-lg font-semibold text-foreground">{t("login.mfa.title")}</h2>
				<p className="text-sm text-muted-foreground">{t("login.mfa.subtitle")}</p>
			</div>

			{methods.length > 1 ? (
				<div className="grid grid-cols-3 gap-1.5 rounded-lg bg-muted/60 p-1">
					{methods.includes("webauthn") ? (
						<button
							type="button"
							onClick={() => {
								verifyMfa.reset();
								setWebauthnError(null);
								setActiveMethod("webauthn");
							}}
							className={cn(
								"flex items-center justify-center gap-1.5 rounded-md px-2.5 py-1.5 text-xs font-medium transition-colors",
								activeMethod === "webauthn"
									? "bg-background text-foreground shadow-sm"
									: "text-muted-foreground hover:text-foreground",
							)}
						>
							<KeyRound className="h-3.5 w-3.5" />
							{t("login.mfa.tab_webauthn")}
						</button>
					) : null}
					{methods.includes("totp") ? (
						<button
							type="button"
							onClick={() => {
								verifyMfa.reset();
								setWebauthnError(null);
								setActiveMethod("totp");
							}}
							className={cn(
								"flex items-center justify-center gap-1.5 rounded-md px-2.5 py-1.5 text-xs font-medium transition-colors",
								activeMethod === "totp"
									? "bg-background text-foreground shadow-sm"
									: "text-muted-foreground hover:text-foreground",
							)}
						>
							<Smartphone className="h-3.5 w-3.5" />
							{t("login.mfa.tab_totp")}
						</button>
					) : null}
					{methods.includes("recovery") ? (
						<button
							type="button"
							onClick={() => {
								verifyMfa.reset();
								setWebauthnError(null);
								setActiveMethod("recovery");
							}}
							className={cn(
								"flex items-center justify-center gap-1.5 rounded-md px-2.5 py-1.5 text-xs font-medium transition-colors",
								activeMethod === "recovery"
									? "bg-background text-foreground shadow-sm"
									: "text-muted-foreground hover:text-foreground",
							)}
						>
							<ShieldCheck className="h-3.5 w-3.5" />
							{t("login.mfa.tab_recovery")}
						</button>
					) : null}
				</div>
			) : null}

			{serverError ? (
				<p role="alert" className="text-center text-sm text-destructive">
					{serverError}
				</p>
			) : null}

			{activeMethod === "webauthn" ? (
				<div className="space-y-4">
					<div className="rounded-lg border border-border bg-muted/30 p-4 text-center text-sm text-muted-foreground">
						{t("login.mfa.webauthn_prompt")}
					</div>
					<Button
						type="button"
						className={authSubmitClass}
						disabled={webauthnBusy || verifyMfa.isPending}
						onClick={() => void runWebAuthn()}
					>
						{webauthnBusy || verifyMfa.isPending
							? t("login.mfa.webauthn_waiting")
							: t("login.mfa.webauthn_cta")}
					</Button>
				</div>
			) : null}

			{activeMethod === "totp" ? (
				<Form {...totpForm}>
					<form onSubmit={handleTotpSubmit} className="space-y-4">
						<OtpField
							control={totpForm.control}
							name="code"
							label={t("login.mfa.totp_label")}
							length={6}
							required
						/>
						<Button type="submit" className={authSubmitClass} disabled={verifyMfa.isPending}>
							{verifyMfa.isPending ? t("login.mfa.verifying") : t("login.mfa.verify_cta")}
						</Button>
					</form>
				</Form>
			) : null}

			{activeMethod === "recovery" ? (
				<Form {...recoveryForm}>
					<form onSubmit={handleRecoverySubmit} className="space-y-4">
						<FormField
							control={recoveryForm.control}
							name="code"
							render={({ field }) => (
								<FormItem>
									<FormLabel>
										<FieldLabelContent required>{t("login.mfa.recovery_label")}</FieldLabelContent>
									</FormLabel>
									<FormControl>
										<Input
											{...field}
											aria-required="true"
											autoComplete="one-time-code"
											placeholder={t("login.mfa.recovery_placeholder")}
											className="font-mono uppercase tracking-wider"
										/>
									</FormControl>
									<FormMessage />
								</FormItem>
							)}
						/>
						<Button type="submit" className={authSubmitClass} disabled={verifyMfa.isPending}>
							{verifyMfa.isPending ? t("login.mfa.verifying") : t("login.mfa.verify_cta")}
						</Button>
					</form>
				</Form>
			) : null}

			<Button
				type="button"
				variant="ghost"
				className="w-full text-sm text-muted-foreground"
				onClick={onBack}
			>
				{t("login.mfa.back_to_login")}
			</Button>
		</div>
	);
}
