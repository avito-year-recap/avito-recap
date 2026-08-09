export type HapticCue = "tap" | "achievement" | "secret" | "cta" | "final";

const patterns: Record<HapticCue, number | number[]> = {
  tap: 8,
  achievement: [12, 26, 18],
  secret: [8, 18, 8, 18, 20],
  cta: 14,
  final: [14, 32, 22],
};

export function playHaptic(cue: HapticCue) {
  if (typeof navigator === "undefined" || typeof navigator.vibrate !== "function") return;
  try { navigator.vibrate(patterns[cue]); } catch { /* Progressive enhancement only. */ }
}
