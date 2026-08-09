import { sonicEngine, type SonicCue } from "../sound/SonicEngine";

export type UiSoundCue = SonicCue;

export function setUiSoundProfile(behaviorCode: string | undefined) {
  sonicEngine.setProfile(behaviorCode);
}

export function playUiSound(cue: UiSoundCue) {
  void sonicEngine.play(cue);
}
