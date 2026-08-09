import { normalizeSonicProfile, sonicProfiles, type SonicProfileCode } from "./soundProfiles";

export type SonicCue =
  | "enable"
  | "navigate"
  | "previous"
  | "metric"
  | "behaviorReveal"
  | "achievement"
  | "secret"
  | "cta"
  | "final";

interface VoiceOptions {
  frequency: number;
  start: number;
  duration: number;
  gain: number;
  type?: OscillatorType;
  detune?: number;
}

export class SonicEngine {
  private ctx: AudioContext | null = null;
  private profileCode: SonicProfileCode = "UNIVERSAL_USER";
  private noiseBuffer: AudioBuffer | null = null;

  setProfile(code: string | undefined) {
    this.profileCode = normalizeSonicProfile(code);
  }

  private context() {
    if (typeof window === "undefined") return null;
    const AudioContextCtor = window.AudioContext ?? (window as typeof window & { webkitAudioContext?: typeof AudioContext }).webkitAudioContext;
    if (!AudioContextCtor) return null;
    this.ctx ??= new AudioContextCtor();
    return this.ctx;
  }

  private async ready() {
    const ctx = this.context();
    if (!ctx) return null;
    if (ctx.state === "suspended") {
      try { await ctx.resume(); } catch { return null; }
    }
    return ctx;
  }

  private voice(ctx: AudioContext, output: AudioNode, options: VoiceOptions) {
    const osc = ctx.createOscillator();
    const gain = ctx.createGain();
    const profile = sonicProfiles[this.profileCode];
    osc.type = options.type ?? "sine";
    osc.frequency.setValueAtTime(options.frequency, options.start);
    if (options.detune) osc.detune.setValueAtTime(options.detune, options.start);
    gain.gain.setValueAtTime(0.0001, options.start);
    gain.gain.exponentialRampToValueAtTime(options.gain, options.start + profile.attack);
    gain.gain.exponentialRampToValueAtTime(0.0001, options.start + options.duration);
    osc.connect(gain);
    gain.connect(output);
    osc.start(options.start);
    osc.stop(options.start + options.duration + 0.03);
  }

  private shimmer(ctx: AudioContext, output: AudioNode, frequency: number, start: number, gainValue = 0.018) {
    this.voice(ctx, output, { frequency, start, duration: 0.38, gain: gainValue, type: "sine" });
    this.voice(ctx, output, { frequency: frequency * 2.01, start: start + 0.012, duration: 0.22, gain: gainValue * 0.36, type: "sine", detune: 4 });
  }

  private noise(ctx: AudioContext, output: AudioNode, start: number, duration: number, gainValue: number, highpass = 1100) {
    if (!this.noiseBuffer || this.noiseBuffer.sampleRate !== ctx.sampleRate) {
      const length = Math.floor(ctx.sampleRate * 0.7);
      const buffer = ctx.createBuffer(1, length, ctx.sampleRate);
      const data = buffer.getChannelData(0);
      for (let i = 0; i < length; i += 1) data[i] = Math.random() * 2 - 1;
      this.noiseBuffer = buffer;
    }
    const source = ctx.createBufferSource();
    const filter = ctx.createBiquadFilter();
    const gain = ctx.createGain();
    source.buffer = this.noiseBuffer;
    filter.type = "highpass";
    filter.frequency.value = highpass;
    gain.gain.setValueAtTime(0.0001, start);
    gain.gain.exponentialRampToValueAtTime(gainValue, start + 0.012);
    gain.gain.exponentialRampToValueAtTime(0.0001, start + duration);
    source.connect(filter);
    filter.connect(gain);
    gain.connect(output);
    source.start(start);
    source.stop(start + duration + 0.02);
  }

  async play(cue: SonicCue) {
    const ctx = await this.ready();
    if (!ctx) return;

    const profile = sonicProfiles[this.profileCode];
    const now = ctx.currentTime + 0.01;
    const master = ctx.createGain();
    const filter = ctx.createBiquadFilter();
    const delay = ctx.createDelay(0.6);
    const feedback = ctx.createGain();
    const wet = ctx.createGain();

    master.gain.value = 0.78;
    filter.type = "lowpass";
    filter.frequency.value = 3400 + profile.brightness * 5200;
    filter.Q.value = 0.55;
    delay.delayTime.value = 0.115;
    feedback.gain.value = 0.16;
    wet.gain.value = 0.22;

    master.connect(filter);
    filter.connect(ctx.destination);
    filter.connect(delay);
    delay.connect(feedback);
    feedback.connect(delay);
    delay.connect(wet);
    wet.connect(ctx.destination);

    const root = profile.root;
    switch (cue) {
      case "enable":
        this.shimmer(ctx, master, root * 1.25, now, 0.015);
        this.shimmer(ctx, master, root * 1.5, now + 0.07, 0.012);
        break;
      case "navigate":
        this.noise(ctx, master, now, 0.09, 0.009, 1500);
        this.voice(ctx, master, { frequency: root, start: now + 0.018, duration: 0.11, gain: 0.012, type: "triangle" });
        break;
      case "previous":
        this.noise(ctx, master, now, 0.075, 0.007, 1700);
        this.voice(ctx, master, { frequency: root * 0.84, start: now, duration: 0.12, gain: 0.01, type: "triangle" });
        break;
      case "metric":
        this.shimmer(ctx, master, root * 1.5, now, 0.013);
        this.voice(ctx, master, { frequency: root * 0.75, start: now, duration: 0.2, gain: 0.007, type: "sine" });
        break;
      case "behaviorReveal":
        this.voice(ctx, master, { frequency: root * 0.5, start: now, duration: 0.55, gain: 0.022, type: "sine" });
        this.voice(ctx, master, { frequency: root, start: now + 0.08, duration: 0.5, gain: 0.014, type: "triangle" });
        this.shimmer(ctx, master, root * 2, now + 0.16, 0.015);
        this.noise(ctx, master, now + 0.12, 0.12, 0.006, 2400);
        break;
      case "achievement": {
        const notes = [1.5, 1.875, 2.25];
        notes.forEach((ratio, index) => this.shimmer(ctx, master, root * ratio, now + index * 0.115, index === 2 ? 0.018 : 0.013));
        this.noise(ctx, master, now + 0.18, 0.18, 0.005, 3000);
        break;
      }
      case "secret":
        [1, 1.26, 1.68, 2].forEach((ratio, index) => this.shimmer(ctx, master, root * ratio, now + index * 0.07, 0.01));
        break;
      case "cta":
        this.voice(ctx, master, { frequency: root * 0.75, start: now, duration: 0.18, gain: 0.018, type: "triangle" });
        this.shimmer(ctx, master, root * 1.5, now + 0.035, 0.012);
        break;
      case "final":
        [0.75, 1, 1.25, 1.5].forEach((ratio, index) => this.voice(ctx, master, { frequency: root * ratio, start: now + index * 0.045, duration: profile.tail + 0.3, gain: index === 0 ? 0.014 : 0.011, type: "sine" }));
        this.shimmer(ctx, master, root * 2, now + 0.18, 0.014);
        this.noise(ctx, master, now + 0.24, 0.24, 0.004, 3400);
        break;
    }
  }
}

export const sonicEngine = new SonicEngine();
