export type SonicProfileCode =
  | "ACTIVE_SELLER"
  | "STARTING_SELLER"
  | "DECISIVE_BUYER"
  | "FIND_HUNTER"
  | "RESEARCHER"
  | "UNIVERSAL_USER";

export interface SonicProfile {
  root: number;
  brightness: number;
  attack: number;
  tail: number;
  movement: "glass" | "paper" | "warm" | "precise" | "spring";
}

export const sonicProfiles: Record<SonicProfileCode, SonicProfile> = {
  RESEARCHER: { root: 392, brightness: 0.86, attack: 0.012, tail: 0.58, movement: "glass" },
  FIND_HUNTER: { root: 440, brightness: 0.78, attack: 0.008, tail: 0.34, movement: "spring" },
  ACTIVE_SELLER: { root: 330, brightness: 0.62, attack: 0.007, tail: 0.28, movement: "precise" },
  STARTING_SELLER: { root: 349.23, brightness: 0.72, attack: 0.014, tail: 0.46, movement: "warm" },
  DECISIVE_BUYER: { root: 415.3, brightness: 0.7, attack: 0.006, tail: 0.22, movement: "precise" },
  UNIVERSAL_USER: { root: 369.99, brightness: 0.68, attack: 0.012, tail: 0.4, movement: "paper" },
};

export function normalizeSonicProfile(code: string | undefined): SonicProfileCode {
  return code && code in sonicProfiles ? code as SonicProfileCode : "UNIVERSAL_USER";
}
