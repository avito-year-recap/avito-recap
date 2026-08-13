import type { Profile } from "../../entities/recap/model";

export type ActiveProfileSnapshot = Pick<
  Profile,
  "profileCode" | "name" | "description" | "avatarUrl"
>;

const STORAGE_KEY = "avito-recap.active-profile";

export const DEFAULT_ACTIVE_PROFILE: ActiveProfileSnapshot = {
  profileCode: "active-buyer",
  name: "Алексей",
  description: "Часто ищет технику, сохраняет интересные варианты и возвращается к ним",
  avatarUrl: "/avatars/active-buyer.png",
};

function isActiveProfileSnapshot(value: unknown): value is ActiveProfileSnapshot {
  if (!value || typeof value !== "object") return false;
  const candidate = value as Partial<ActiveProfileSnapshot>;
  return (
    typeof candidate.profileCode === "string" &&
    typeof candidate.name === "string" &&
    typeof candidate.description === "string" &&
    typeof candidate.avatarUrl === "string"
  );
}

export function getActiveProfile(): ActiveProfileSnapshot {
  if (typeof window === "undefined") return DEFAULT_ACTIVE_PROFILE;

  try {
    const stored = window.localStorage.getItem(STORAGE_KEY);
    if (!stored) return DEFAULT_ACTIVE_PROFILE;

    const parsed: unknown = JSON.parse(stored);
    return isActiveProfileSnapshot(parsed) ? parsed : DEFAULT_ACTIVE_PROFILE;
  } catch {
    return DEFAULT_ACTIVE_PROFILE;
  }
}

export function getActiveProfileCode(): string {
  return getActiveProfile().profileCode;
}

export function setActiveProfile(profile: Profile): void {
  if (typeof window === "undefined") return;

  const snapshot: ActiveProfileSnapshot = {
    profileCode: profile.profileCode,
    name: profile.name,
    description: profile.description,
    avatarUrl: profile.avatarUrl,
  };

  window.localStorage.setItem(STORAGE_KEY, JSON.stringify(snapshot));
}
