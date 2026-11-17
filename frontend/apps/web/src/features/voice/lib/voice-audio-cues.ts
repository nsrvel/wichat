let sharedAudioCtx: AudioContext | null = null;

function getAudioContext(): AudioContext | null {
	if (typeof window === "undefined") return null;
	if (!sharedAudioCtx) {
		const Ctor =
			window.AudioContext ??
			(window as unknown as { webkitAudioContext?: typeof AudioContext }).webkitAudioContext;
		if (!Ctor) return null;
		sharedAudioCtx = new Ctor();
	}
	if (sharedAudioCtx.state === "suspended") {
		void sharedAudioCtx.resume().catch(() => {});
	}
	return sharedAudioCtx;
}

function playToneSequence(freqs: number[], durationSec = 0.08, gainPeak = 0.08) {
	const ctx = getAudioContext();
	if (!ctx) return;

	const now = ctx.currentTime;
	freqs.forEach((freq, idx) => {
		const osc = ctx.createOscillator();
		const gain = ctx.createGain();
		osc.type = "sine";
		osc.frequency.setValueAtTime(freq, now + idx * (durationSec * 0.8));

		gain.gain.setValueAtTime(0.0001, now + idx * (durationSec * 0.8));
		gain.gain.exponentialRampToValueAtTime(gainPeak, now + idx * (durationSec * 0.8) + 0.015);
		gain.gain.exponentialRampToValueAtTime(0.0001, now + idx * (durationSec * 0.8) + durationSec);

		osc.connect(gain);
		gain.connect(ctx.destination);
		osc.start(now + idx * (durationSec * 0.8));
		osc.stop(now + idx * (durationSec * 0.8) + durationSec + 0.02);
	});
}

export function playVoiceJoinCue() {
	playToneSequence([440, 587.33], 0.09, 0.07);
}

export function playVoiceLeaveCue() {
	playToneSequence([523.25, 349.23], 0.09, 0.07);
}

export function playVoiceMuteCue(muted: boolean) {
	playToneSequence(muted ? [392, 311.13] : [311.13, 440], 0.06, 0.05);
}

export function playVoiceDeafenCue(deafened: boolean) {
	playToneSequence(deafened ? [349.23, 261.63] : [261.63, 392], 0.07, 0.06);
}
