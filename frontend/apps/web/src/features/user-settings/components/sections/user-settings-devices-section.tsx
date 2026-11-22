"use client";

import { ApiError } from "@wichat/api-client/api-error";
import { useQuery } from "@tanstack/react-query";
import type { DeviceSession } from "@wichat/contracts/auth/device";
import { Monitor, MonitorSmartphone, Smartphone, X } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { isQueryFailureWithoutData } from "@/lib/query-display";
import {
	DeviceLogoutDialog,
	type DeviceLogoutTarget,
} from "@/features/user-settings/components/dialogs/device-logout-dialog";
import type { UserSettingsSectionId } from "@/features/user-settings/lib/user-settings-nav";
import { translateError } from "@/integrations/i18n/translate-error";
import {
	settingsSectionDividerLineClass,
	settingsSectionDividerWrapClass,
} from "@/lib/settings-chrome";
import { useRevokeMyDeviceMutation } from "@/services/user/user.mutation";
import { myDevicesQueryOptions } from "@/services/user/user.query";

function DeviceIcon({ platform }: { platform: string }) {
	const Icon = platform === "mobile" ? Smartphone : Monitor;
	return (
		<div className="flex size-10 shrink-0 items-center justify-center rounded-xl border border-border/60 bg-muted/60 text-foreground">
			<Icon className="size-5" aria-hidden />
		</div>
	);
}

function splitDeviceLabel(
	device: DeviceSession,
	fallback: string,
): { primary: string; secondary?: string } {
	const raw = device.device_label.trim();
	if (!raw) {
		return { primary: device.platform ? device.platform.toUpperCase() : fallback };
	}
	const parts = raw
		.split("·")
		.map((s) => s.trim())
		.filter(Boolean);
	if (parts.length >= 2) {
		return { primary: parts[0], secondary: parts.slice(1).join(" · ") };
	}
	return { primary: raw };
}

function formatRelativeTime(iso: string, locale: string): string {
	const ts = Date.parse(iso);
	if (Number.isNaN(ts)) {
		return "";
	}
	const diffSec = Math.round((ts - Date.now()) / 1000);
	const absSec = Math.abs(diffSec);
	const rtf = new Intl.RelativeTimeFormat(locale || "en", { numeric: "auto" });

	if (absSec < 60) {
		return rtf.format(diffSec, "second");
	}
	const diffMin = Math.round(diffSec / 60);
	if (Math.abs(diffMin) < 60) {
		return rtf.format(diffMin, "minute");
	}
	const diffHour = Math.round(diffMin / 60);
	if (Math.abs(diffHour) < 24) {
		return rtf.format(diffHour, "hour");
	}
	const diffDay = Math.round(diffHour / 24);
	if (Math.abs(diffDay) < 30) {
		return rtf.format(diffDay, "day");
	}
	const diffMonth = Math.round(diffDay / 30);
	if (Math.abs(diffMonth) < 12) {
		return rtf.format(diffMonth, "month");
	}
	return rtf.format(Math.round(diffDay / 365), "year");
}

function DeviceRowSkeleton() {
	return (
		<div className="flex items-center gap-4 py-1">
			<div className="size-10 shrink-0 animate-pulse rounded-xl bg-muted/70" />
			<div className="min-w-0 flex-1 space-y-2">
				<div className="h-3.5 w-40 animate-pulse rounded bg-muted/70" />
				<div className="h-3 w-56 animate-pulse rounded bg-muted/50" />
			</div>
		</div>
	);
}

type UserSettingsDevicesSectionProps = {
	hasPassword: boolean;
	onNavigateSection: (section: UserSettingsSectionId) => void;
};

export function UserSettingsDevicesSection({
	hasPassword,
	onNavigateSection,
}: UserSettingsDevicesSectionProps) {
	const { t, i18n } = useTranslation();
	const [logoutTarget, setLogoutTarget] = useState<DeviceLogoutTarget | null>(null);
	const [verifiedPassword, setVerifiedPassword] = useState<string | null>(null);
	const { data, isPending, isError, refetch } = useQuery(myDevicesQueryOptions());
	const showLoadError = isQueryFailureWithoutData(isError, data);
	const revokeDevice = useRevokeMyDeviceMutation();

	const handleLogoutSingleClick = (deviceId: string, label: string) => {
		if (hasPassword && verifiedPassword !== null) {
			revokeDevice.mutate(
				{ id: deviceId, current_password: verifiedPassword },
				{
					onSuccess: () => {
						toast.success(t("user_settings.devices.revoke_success"));
					},
					onError: (err) => {
						if (err instanceof ApiError && err.code === "auth.invalid_credentials") {
							setVerifiedPassword(null);
							setLogoutTarget({ kind: "single", deviceId, deviceLabel: label });
							return;
						}
						toast.error(translateError(t, err));
					},
				},
			);
			return;
		}
		setLogoutTarget({ kind: "single", deviceId, deviceLabel: label });
	};

	const devices = data?.devices ?? [];
	const currentDevice = devices.find((d) => d.is_current) ?? null;
	const otherDevices = devices.filter((d) => !d.is_current);
	const locale = i18n.resolvedLanguage ?? i18n.language ?? "en";

	return (
		<div className="flex flex-col">
			{/* Current Device */}
			<section className="space-y-3">
				<h3 className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
					{t("user_settings.devices.current_device")}
				</h3>
				{isPending ? (
					<DeviceRowSkeleton />
				) : showLoadError ? (
					<div className="flex items-center justify-between gap-4 py-1">
						<p className="text-xs text-muted-foreground">{t("user_settings.devices.load_error")}</p>
						<Button type="button" variant="secondary" size="sm" onClick={() => void refetch()}>
							{t("common.retry")}
						</Button>
					</div>
				) : currentDevice ? (
					(() => {
						const title = splitDeviceLabel(
							currentDevice,
							t("user_settings.devices.unknown_device"),
						);
						return (
							<div className="flex items-center gap-4 py-1">
								<DeviceIcon platform={currentDevice.platform} />
								<div className="min-w-0 flex-1">
									<p className="truncate text-sm font-medium text-foreground">
										<span>{title.primary}</span>
										{title.secondary ? (
											<>
												<span className="mx-1.5 text-muted-foreground">•</span>
												<span className="font-normal text-foreground/90">{title.secondary}</span>
											</>
										) : null}
									</p>
									<p className="truncate text-xs text-muted-foreground">
										{currentDevice.location_label || t("user_settings.devices.unknown_location")}
									</p>
								</div>
							</div>
						);
					})()
				) : null}
			</section>

			<div className={settingsSectionDividerWrapClass} role="presentation">
				<div className={settingsSectionDividerLineClass} />
			</div>

			{/* Other Devices */}
			<section className="space-y-3">
				<h3 className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
					{t("user_settings.devices.other_devices")}
				</h3>

				{isPending ? (
					<div className="space-y-3">
						<DeviceRowSkeleton />
						<DeviceRowSkeleton />
					</div>
				) : otherDevices.length > 0 ? (
					<div className="space-y-3">
						{otherDevices.map((device) => {
							const title = splitDeviceLabel(device, t("user_settings.devices.unknown_device"));
							const relative = formatRelativeTime(device.last_used_at, locale);
							const location = device.location_label || t("user_settings.devices.unknown_location");
							const label = title.secondary
								? `${title.primary} • ${title.secondary}`
								: title.primary;

							return (
								<div key={device.id} className="flex items-center gap-4 py-1">
									<DeviceIcon platform={device.platform} />
									<div className="min-w-0 flex-1">
										<p className="truncate text-sm font-medium text-foreground">
											<span>{title.primary}</span>
											{title.secondary ? (
												<>
													<span className="mx-1.5 text-muted-foreground">•</span>
													<span className="font-normal text-foreground/90">{title.secondary}</span>
												</>
											) : null}
										</p>
										<p className="truncate text-xs text-muted-foreground">
											<span>{location}</span>
											{relative ? (
												<>
													<span className="mx-1.5">•</span>
													<span>{relative}</span>
												</>
											) : null}
										</p>
									</div>
									<Button
										type="button"
										variant="ghost"
										size="icon"
										disabled={revokeDevice.isPending}
										aria-label={t("user_settings.devices.log_out_device")}
										onClick={() => handleLogoutSingleClick(device.id, label)}
										className="size-8 shrink-0 text-muted-foreground hover:text-foreground"
									>
										<X className="size-4" aria-hidden />
									</Button>
								</div>
							);
						})}
					</div>
				) : (
					<p className="py-1 text-xs text-muted-foreground">
						{t("user_settings.devices.no_other_devices")}
					</p>
				)}

				<div className="flex items-center gap-4 pt-2">
					<div className="flex size-10 shrink-0 items-center justify-center rounded-xl border border-border/60 bg-muted/60 text-muted-foreground">
						<MonitorSmartphone className="size-5" aria-hidden />
					</div>
					<div className="min-w-0 flex-1">
						<p className="text-sm font-medium text-foreground">
							{t("user_settings.devices.older_devices_title")}
						</p>
						<p className="text-xs text-muted-foreground">
							{t("user_settings.devices.older_devices_prefix")}{" "}
							<button
								type="button"
								onClick={() => onNavigateSection("account")}
								className="font-medium text-primary underline-offset-2 hover:underline"
							>
								{t("user_settings.devices.older_devices_action")}
							</button>
						</p>
					</div>
				</div>
			</section>

			<div className={settingsSectionDividerWrapClass} role="presentation">
				<div className={settingsSectionDividerLineClass} />
			</div>

			{/* Log out of all known devices */}
			<section className="space-y-3">
				<div className="space-y-1">
					<h3 className="text-sm font-semibold text-foreground">
						{t("user_settings.devices.logout_all_title")}
					</h3>
					<p className="text-xs text-muted-foreground">
						{t("user_settings.devices.logout_all_desc")}
					</p>
				</div>
				<Button
					type="button"
					variant="destructive"
					size="lg"
					disabled={isPending || otherDevices.length === 0}
					onClick={() => setLogoutTarget({ kind: "all" })}
					className="min-w-[140px] shrink-0 px-4 text-[13px]"
				>
					{t("user_settings.devices.logout_all_button")}
				</Button>
			</section>

			<DeviceLogoutDialog
				open={logoutTarget !== null}
				onOpenChange={(nextOpen) => {
					if (!nextOpen) {
						setLogoutTarget(null);
					}
				}}
				target={logoutTarget}
				hasPassword={hasPassword}
				onSingleVerified={setVerifiedPassword}
			/>
		</div>
	);
}
