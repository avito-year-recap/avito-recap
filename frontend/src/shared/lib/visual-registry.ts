import type { ActionCode } from "../../entities/recap/model";

export type VisualTone = "blue" | "green" | "purple" | "coral" | "neutral";
export type VisualMotif = "orbit" | "stack" | "trail" | "pulse" | "constellation" | "grid";

interface VisualDefinition {
  icon: string;
  secondary?: string;
  tone: VisualTone;
  caption: string;
  motif?: VisualMotif;
}

const categoryVisuals: Record<string, VisualDefinition> = {
  HOME_AND_GARDEN: {
    icon: "🪑",
    secondary: "🪴",
    tone: "green",
    caption: "пространство для новых идей",
    motif: "constellation",
  },
  CARS: {
    icon: "🚙",
    secondary: "🛞",
    tone: "blue",
    caption: "маршрут к следующей находке",
    motif: "trail",
  },
  ELECTRONICS: {
    icon: "📱",
    secondary: "🎧",
    tone: "purple",
    caption: "техника, к которой возвращались чаще",
    motif: "grid",
  },
};

const behaviorVisuals: Record<string, VisualDefinition> = {
  ACTIVE_SELLER: {
    icon: "↗",
    secondary: "▤",
    tone: "green",
    caption: "Продажи в движении",
    motif: "stack",
  },
  STARTING_SELLER: {
    icon: "+",
    secondary: "▤",
    tone: "coral",
    caption: "Старт в продажах",
    motif: "stack",
  },
  DECISIVE_BUYER: {
    icon: "✓",
    secondary: "◎",
    tone: "blue",
    caption: "Решение принято",
    motif: "trail",
  },
  FIND_HUNTER: {
    icon: "♡",
    secondary: "⌕",
    tone: "coral",
    caption: "Охота за находками",
    motif: "constellation",
  },
  RESEARCHER: {
    icon: "⌕",
    secondary: "◎",
    tone: "purple",
    caption: "Глубокое исследование",
    motif: "orbit",
  },
  UNIVERSAL_USER: {
    icon: "✦",
    secondary: "∞",
    tone: "blue",
    caption: "Разные сценарии",
    motif: "pulse",
  },
};

const achievementVisuals: Record<string, VisualDefinition> = {
  SUCCESSFUL_SELLER: { icon: "✓", tone: "green", caption: "Продажи", motif: "pulse" },
  DEAL_CLOSER: { icon: "◆", tone: "blue", caption: "Сделки", motif: "pulse" },
  CONSISTENT_PUBLISHER: { icon: "▤", tone: "green", caption: "Публикации", motif: "stack" },
  BROAD_INTERESTS: { icon: "✦", tone: "purple", caption: "Интересы", motif: "constellation" },
  ALL_ROUNDER: { icon: "∞", tone: "blue", caption: "Всё в одном", motif: "orbit" },
  FIRST_SELLING_STEPS: { icon: "+", tone: "coral", caption: "Первый шаг", motif: "stack" },
  QUICK_DECISION: { icon: "→", tone: "blue", caption: "Решение", motif: "trail" },
  ATTENTIVE_RESEARCHER: { icon: "◎", tone: "purple", caption: "Сравнение", motif: "orbit" },
  MASTER_OF_FAVORITES: { icon: "♡", tone: "coral", caption: "Избранное", motif: "constellation" },
};

const actionVisuals: Record<ActionCode, VisualDefinition> = {
  FINISH_DRAFT: { icon: "▤", secondary: "✓", tone: "coral", caption: "Закончить черновик", motif: "stack" },
  OPEN_FAVORITES: { icon: "♡", secondary: "→", tone: "coral", caption: "Вернуться к избранному", motif: "constellation" },
  IMPROVE_LISTINGS: { icon: "↗", secondary: "▤", tone: "green", caption: "Усилить объявление", motif: "stack" },
  CONTINUE_DIALOGS: { icon: "•••", secondary: "→", tone: "blue", caption: "Продолжить разговор", motif: "trail" },
  OPEN_TOP_CATEGORY: { icon: "⌕", secondary: "→", tone: "blue", caption: "Открыть интерес", motif: "orbit" },
  CREATE_FIRST_LISTING: { icon: "+", secondary: "▤", tone: "coral", caption: "Создать первое объявление", motif: "stack" },
  CREATE_LISTING: { icon: "+", secondary: "▤", tone: "green", caption: "Создать объявление", motif: "stack" },
  SAVE_SEARCH: { icon: "⌕", secondary: "♡", tone: "blue", caption: "Сохранить поиск", motif: "orbit" },
  VIEW_SIMILAR_LISTINGS: { icon: "◎", secondary: "→", tone: "purple", caption: "Посмотреть похожее", motif: "orbit" },
  EXPLORE_RECOMMENDATIONS: { icon: "✦", secondary: "→", tone: "purple", caption: "Открыть рекомендации", motif: "constellation" },
};

const fallback: VisualDefinition = {
  icon: "✦",
  secondary: "◎",
  tone: "blue",
  caption: "Твой год на Авито",
  motif: "pulse",
};

export function getCategoryVisual(code: string) {
  return categoryVisuals[code] ?? fallback;
}

export function getBehaviorVisual(code: string) {
  return behaviorVisuals[code] ?? fallback;
}

export function getAchievementVisual(code: string) {
  return achievementVisuals[code] ?? fallback;
}

export function getActionVisual(code: ActionCode | string) {
  return actionVisuals[code as ActionCode] ?? fallback;
}

export function getMonthVisual(month: number) {
  if ([12, 1, 2].includes(month)) return { season: "winter", icon: "✦", caption: "зимний ритм" } as const;
  if ([3, 4, 5].includes(month)) return { season: "spring", icon: "❋", caption: "время новых идей" } as const;
  if ([6, 7, 8].includes(month)) return { season: "summer", icon: "☼", caption: "летний импульс" } as const;
  return { season: "autumn", icon: "◆", caption: "сезон внимательных находок" } as const;
}
