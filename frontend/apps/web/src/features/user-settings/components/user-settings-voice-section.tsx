"use client";

import { useEffect, useRef, useState, useSyncExternalStore } from "react";
import { Mic, Volume2 } from "lucide-react";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import {
	DEFAULT_VOICE_SETTINGS,
	readVoiceSettings,
	subscribeVoiceSettings,
	writeVoiceSettings,
} from "@/features/voice/lib/voice-settings-storage";

export function UserSettingsVoiceSection() {
	const { t } = useTranslation();
	const settings = useSyncExternalStore(
		subscribeVoiceSettings,
		readVoiceSettings,
		() => DEFAULT_VOICE_SETTINGS,
	);

	const [devices, setDevices] = useState<MediaDeviceInfo[]>([]);
	const [testingMic, setTestingMic] = useState(false);
	const [micLevel, setMicLevel] = useState(0);
	const [capturingPtt, setCapturingPtt] = useState(false);
	const testStreamRef = useRef<MediaStream | null>(null);
	const rafRef = useRef<number | null>(null);

	useEffect(() => {
		if (typeof navigator === "undefined" || !navigator.mediaDevices?.enumerateDevices) {
			return;
		}
		const load = () => {
			void navigator.mediaDevices
				.enumerateDevices()
				.then(setDevices)
				.catch(() => {});
		};
		load();
		navigator.mediaDevices.addEventListener?.("devicechange", load);
		return () => {
			navigator.mediaDevices.removeEventListener?.("devicechange", load);
		};
	}, []);

	const stopMicTest = () => {
		if (rafRef.current !== null) {
			cancelAnimationFrame(rafRef.current);
			rafRef.current = null;
		}
		testStreamRef.current?.getTracks().forEach((tr) => tr.stop());
		testStreamRef.current = null;
		setTestingMic(false);
		setMicLevel(0);
	};

	useEffect(() => stopMicTest, []);

	const startMicTest = async () => {
		if (testingMic) {
			stopMicTest();
			return;
		}
		try {
			const stream = await navigator.mediaDevices.getUserMedia({
				audio: {
					deviceId: settings.inputDeviceId ? { exact: settings.inputDeviceId } : undefined,
					echoCancellation: settings.echoCancellation,
					noiseSuppression: settings.noiseSuppression,
					autoGainControl: settings.autoGainControl,
				},
			});
			testStreamRef.current = stream;
			const ctx = new AudioContext();
			const src = ctx.createMediaStreamSource(stream);
			const analyser = ctx.createAnalyser();
			analyser.fftSize = 256;
			src.connect(analyser);
			const data = new Uint8Array(analyser.frequencyBinCount);
			setTestingMic(true);

			const tick = () => {
				if (!testStreamRef.current) {
					void ctx.close();
					return;
				}
				analyser.getByteFrequencyData(data);
				let sum = 0;
				for (let i = 0; i < data.length; i++) sum += data[i];
				const avg = sum / Math.max(1, data.length);
				setMicLevel(Math.min(100, Math.round((avg / 128) * 100)));
				rafRef.current = requestAnimationFrame(tick);
			};
			rafRef.current = requestAnimationFrame(tick);
		} catch {
			stopMicTest();
		}
	};

	useEffect(() => {
		if (!capturingPtt) return;
		const onKeyDown = (e: KeyboardEvent) => {
			e.preventDefault();
			writeVoiceSettings({ pttKey: e.code });
			setCapturingPtt(false);
		};
		window.addEventListener("keydown", onKeyDown, { once: true });
		return () => window.removeEventListener("keydown", onKeyDown);
	}, [capturingPtt]);

	const audioInputs = devices.filter((d) => d.kind === "audioinput");
	const audioOutputs = devices.filter((d) => d.kind === "audiooutput");
	const videoInputs = devices.filter((d) => d.kind === "videoinput");

	return (
		<div className="flex flex-col gap-5">
			<h3 className="text-base font-semibold text-text-primary">{t("voice.settings.title")}</h3>

			<div className="grid gap-4 sm:grid-cols-2">
				<div className="flex flex-col gap-1.5">
					<Label>{t("voice.settings.input_device")}</Label>
					<select
						className="h-9 rounded-md border border-border bg-chat-surface px-2.5 text-sm text-text-primary"
						value={settings.inputDeviceId}
						onChange={(e) => writeVoiceSettings({ inputDeviceId: e.target.value })}
					>
						<option value="">{t("voice.settings.default_device")}</option>
						{audioInputs.map((d) => (
							<option key={d.deviceId} value={d.deviceId}>
								{d.label || d.deviceId}
							</option>
						))}
					</select>
				</div>

				<div className="flex flex-col gap-1.5">
					<Label>{t("voice.settings.output_device")}</Label>
					<select
						className="h-9 rounded-md border border-border bg-chat-surface px-2.5 text-sm text-text-primary"
						value={settings.outputDeviceId}
						onChange={(e) => writeVoiceSettings({ outputDeviceId: e.target.value })}
					>
						<option value="">{t("voice.settings.default_device")}</option>
						{audioOutputs.map((d) => (
							<option key={d.deviceId} value={d.deviceId}>
								{d.label || d.deviceId}
							</option>
						))}
					</select>
				</div>

				<div className="flex flex-col gap-1.5 sm:col-span-2">
					<Label>{t("voice.settings.camera_device")}</Label>
					<select
						className="h-9 rounded-md border border-border bg-chat-surface px-2.5 text-sm text-text-primary"
						value={settings.cameraDeviceId}
						onChange={(e) => writeVoiceSettings({ cameraDeviceId: e.target.value })}
					>
						<option value="">{t("voice.settings.default_device")}</option>
						{videoInputs.map((d) => (
							<option key={d.deviceId} value={d.deviceId}>
								{d.label || d.deviceId}
							</option>
						))}
					</select>
				</div>
			</div>

			<div className="flex items-center gap-3 rounded-lg border border-border bg-chat-surface p-3">
				<Button type="button" size="sm" variant="outline" onClick={() => void startMicTest()}>
					<Mic className="mr-1.5 size-4" />
					{testingMic ? t("voice.settings.mic_test_stop") : t("voice.settings.mic_test")}
				</Button>
				<div className="h-2.5 flex-1 overflow-hidden rounded-full bg-chat-surface-elevated">
					<div
						className="h-full bg-emerald-400 transition-[width] duration-75"
						style={{ width: `${micLevel}%` }}
					/>
				</div>
				<Volume2 className="size-4 text-text-muted" />
			</div>

			<div className="flex flex-col gap-3 rounded-lg border border-border bg-chat-surface p-3">
				<Label>{t("voice.settings.input_mode")}</Label>
				<div className="flex gap-2">
					<Button
						type="button"
						size="sm"
						variant={settings.inputMode === "vad" ? "default" : "outline"}
						onClick={() => writeVoiceSettings({ inputMode: "vad" })}
					>
						{t("voice.settings.vad")}
					</Button>
					<Button
						type="button"
						size="sm"
						variant={settings.inputMode === "ptt" ? "default" : "outline"}
						onClick={() => writeVoiceSettings({ inputMode: "ptt" })}
					>
						{t("voice.settings.ptt")}
					</Button>
				</div>

				{settings.inputMode === "vad" ? (
					<div className="flex flex-col gap-1.5">
						<div className="flex justify-between text-xs text-text-muted">
							<span>{t("voice.settings.vad_sensitivity")}</span>
							<span>{settings.vadThreshold}%</span>
						</div>
						<input
							type="range"
							min={0}
							max={100}
							value={settings.vadThreshold}
							onChange={(e) => writeVoiceSettings({ vadThreshold: Number(e.target.value) })}
						/>
					</div>
				) : (
					<div className="flex items-center justify-between">
						<span className="text-xs text-text-muted">{t("voice.settings.ptt_key")}</span>
						<Button type="button" size="sm" variant="outline" onClick={() => setCapturingPtt(true)}>
							{capturingPtt ? t("voice.settings.press_any_key") : settings.pttKey}
						</Button>
					</div>
				)}
			</div>

			<div className="flex flex-col gap-2.5 text-sm">
				{(
					[
						["echoCancellation", t("voice.settings.echo_cancellation")],
						["noiseSuppression", t("voice.settings.noise_suppression")],
						["autoGainControl", t("voice.settings.auto_gain_control")],
					] as const
				).map(([key, label]) => (
					<label key={key} className="flex items-center justify-between gap-3">
						<span className="text-text-secondary">{label}</span>
						<input
							type="checkbox"
							checked={settings[key]}
							onChange={(e) => writeVoiceSettings({ [key]: e.target.checked })}
							className="size-4 accent-primary"
						/>
					</label>
				))}
			</div>
		</div>
	);
}
