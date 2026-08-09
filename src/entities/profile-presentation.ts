import type { Profile } from "./recap/model";

type ProfilePresentation = Pick<Profile, "tags" | "accent">;

const fallback: ProfilePresentation = {
  tags: ["Итоги года"],
  accent: "blue",
};

const presentationByProfileCode: Record<string, ProfilePresentation> = {
  MARINA_RESEARCHER: {
    tags: ["Дом и дача", "Сохранения", "Исследование"],
    accent: "purple",
  },
  ILYA_ACTIVE_SELLER: {
    tags: ["Продажи", "Публикации", "Диалоги"],
    accent: "green",
  },
  ALEXEY_DECISIVE_BUYER: {
    tags: ["Авто", "Диалоги", "Покупки"],
    accent: "blue",
  },
  SONYA_UNIVERSAL: {
    tags: ["Разные категории", "Новый профиль"],
    accent: "coral",
  },
  DASHA_STARTING_SELLER: {
    tags: ["Черновики", "Первые продажи"],
    accent: "coral",
  },
};

/**
 * UI-only metadata. It is intentionally not part of the Protobuf contract:
 * the backend profile contains only name/description/avatarUrl/profileCode.
 */
export function getProfilePresentation(
  profileCode: string,
): ProfilePresentation {
  const value = presentationByProfileCode[profileCode] ?? fallback;
  return { tags: [...value.tags], accent: value.accent };
}
