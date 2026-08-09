import type {
  AchievementCard,
  BehaviorCard,
  BehaviorEvidence,
  PublicSharePayload,
  Recap,
  RecapCard,
  YearActivityCard,
} from "../../entities/recap/model";

export interface BehaviorTrait {
  id: string;
  title: string;
  value: string;
  explanation: string;
  icon: string;
}

const traitDictionary: Record<string, { title: string; icon: string }> = {
  totalViews: { title: "Сравниваешь", icon: "◎" },
  categoriesCount: { title: "Исследуешь широко", icon: "✦" },
  chatsStarted: { title: "Не спешишь писать", icon: "•••" },
  favoritesAdded: { title: "Сохраняешь лучшее", icon: "♡" },
  purchasesCompleted: { title: "Решаешься", icon: "✓" },
  salesCompleted: { title: "Доводишь до результата", icon: "↗" },
  listingsPublished: { title: "Публикуешь", icon: "▤" },
  searches: { title: "Ищешь варианты", icon: "⌕" },
};

function evidenceValue(evidence: BehaviorEvidence) {
  const sign = evidence.comparison === "LTE" ? "≤" : "≥";
  return `${evidence.actualValue.toLocaleString("ru-RU")} · порог ${sign} ${evidence.threshold.toLocaleString("ru-RU")}`;
}

export function deriveBehaviorTraits(card: BehaviorCard): BehaviorTrait[] {
  return card.payload.evidence.slice(0, 4).map((evidence) => {
    const preset = traitDictionary[evidence.metric] ?? { title: evidence.label, icon: "◇" };
    return {
      id: evidence.metric,
      title: preset.title,
      icon: preset.icon,
      value: evidenceValue(evidence),
      explanation: evidence.explanation,
    };
  });
}

export function getSecretVisualBonus(codes: string[]) {
  const has = (code: string) => codes.includes(code);
  if (has("MASTER_OF_FAVORITES") && has("BROAD_INTERESTS")) {
    return { title: "Куратор находок", glyph: "✦♡", caption: "визуальный бонус за сочетание интересов" };
  }
  if (has("SUCCESSFUL_SELLER") && has("CONSISTENT_PUBLISHER")) {
    return { title: "Витрина в движении", glyph: "▤↗", caption: "визуальный бонус за связку публикаций и продаж" };
  }
  if (has("DEAL_CLOSER") && has("QUICK_DECISION")) {
    return { title: "Точный выбор", glyph: "◆✓", caption: "визуальный бонус за решительный сценарий" };
  }
  if (codes.length >= 3) {
    return { title: "Тройной знак", glyph: "◇✦◇", caption: "декоративный бонус за полный сет экрана" };
  }
  return null;
}

const finalLines: Record<string, string> = {
  ACTIVE_SELLER: "Продолжай двигать вещи дальше",
  STARTING_SELLER: "Самое время довести идею до публикации",
  DECISIVE_BUYER: "Похоже, ты знаешь, чего хочешь",
  FIND_HUNTER: "Следующая находка уже где-то рядом",
  RESEARCHER: "Продолжай находить лучшее",
  UNIVERSAL_USER: "У тебя ещё много сценариев впереди",
};

export function getPersonalizedFinalLine(recap: Recap) {
  const behavior = recap.cards.find((card): card is BehaviorCard => card.type === "BEHAVIOR");
  return finalLines[behavior?.payload.code ?? "UNIVERSAL_USER"] ?? finalLines.UNIVERSAL_USER;
}

const metricLabels: Record<keyof YearActivityCard["payload"], string> = {
  totalEvents: "Все события",
  searches: "Поиски",
  totalViews: "Просмотры",
  favoritesAdded: "Избранное",
  chatsStarted: "Диалоги",
  listingsPublished: "Публикации",
  purchasesCompleted: "Покупки",
  salesCompleted: "Продажи",
};

export function getDominantActivity(card: YearActivityCard) {
  const entries = (Object.entries(card.payload) as Array<[keyof YearActivityCard["payload"], number]>)
    .filter(([key]) => key !== "totalEvents");
  const [key, value] = entries.reduce((best, current) => (current[1] > best[1] ? current : best));
  return {
    key,
    value,
    label: metricLabels[key],
    text: `Среди показанных событий сильнее всего выделяются ${metricLabels[key].toLowerCase()}.`,
  };
}

export function getPublicPayload(recap: Recap): PublicSharePayload | null {
  const share = recap.cards.find((card): card is Extract<RecapCard, { type: "SHARE" }> => card.type === "SHARE");
  return share?.payload ?? null;
}

export function getActionBeforeAfter(code: string) {
  const map: Record<string, { before: string; after: string }> = {
    SAVE_SEARCH: { before: "Каждый раз искать заново", after: "Новые варианты собраны автоматически" },
    FINISH_DRAFT: { before: "Черновик ждёт последнего шага", after: "Объявление готово к публикации" },
    OPEN_FAVORITES: { before: "Вспоминать, что понравилось", after: "Вернуться к сохранённым находкам" },
    CONTINUE_DIALOGS: { before: "Оставить разговор на паузе", after: "Продолжить с того же места" },
    CREATE_FIRST_LISTING: { before: "Первое объявление ещё не опубликовано", after: "Первое объявление уже создаётся" },
    CREATE_LISTING: { before: "Идея остаётся идеей", after: "Новое объявление уже в работе" },
    IMPROVE_LISTINGS: { before: "Объявление остаётся как есть", after: "Можно усилить его прямо сейчас" },
    OPEN_TOP_CATEGORY: { before: "Начинать поиск с нуля", after: "Сразу открыть главный интерес" },
    VIEW_SIMILAR_LISTINGS: { before: "Искать похожее вручную", after: "Открыть похожие варианты одним шагом" },
    EXPLORE_RECOMMENDATIONS: { before: "Не знать, куда идти дальше", after: "Получить подборку следующих идей" },
  };
  return map[code] ?? { before: "Остановиться на итогах", after: "Сделать следующий полезный шаг" };
}

export function getProfileTeaser(profile: Recap["profile"]) {
  const first = profile.tags[0] ?? "интересы";
  const second = profile.tags[1] ?? "активность";
  return `В этой истории особенно заметны ${first.toLowerCase()} и ${second.toLowerCase()}. Финальный сценарий откроется только внутри recap.`;
}

export function getRecapStorageKey(recapId: string) {
  return `avito-recap-progress:${recapId}`;
}

export function readStoredProgress(recapId: string): { index: number; completed: boolean } | null {
  if (typeof window === "undefined") return null;
  try {
    const raw = window.localStorage.getItem(getRecapStorageKey(recapId));
    if (!raw) return null;
    const parsed = JSON.parse(raw) as { index?: number; completed?: boolean };
    if (!Number.isInteger(parsed.index)) return null;
    return { index: parsed.index ?? 0, completed: Boolean(parsed.completed) };
  } catch {
    return null;
  }
}

export function writeStoredProgress(recapId: string, index: number, completed: boolean) {
  if (typeof window === "undefined") return;
  try {
    window.localStorage.setItem(getRecapStorageKey(recapId), JSON.stringify({ index, completed }));
  } catch {
    // Local persistence is an enhancement; the story still works without it.
  }
}

export function resetStoredProgress(recapId: string) {
  if (typeof window === "undefined") return;
  try {
    window.localStorage.removeItem(getRecapStorageKey(recapId));
  } catch {
    // Ignore storage restrictions.
  }
}

export function getAchievementCodes(recap: Recap) {
  const card = recap.cards.find((item): item is AchievementCard => item.type === "ACHIEVEMENT");
  return card?.payload.codes ?? [];
}

const COMPLETED_PROFILES_KEY = "avito-recap-completed-profiles";
const BONUS_PREFIX = "avito-recap-finale-bonus:";

export function markProfileCompleted(profileCode: string) {
  if (typeof window === "undefined") return;
  try {
    const current = readCompletedProfiles();
    const next = Array.from(new Set([...current, profileCode]));
    window.localStorage.setItem(COMPLETED_PROFILES_KEY, JSON.stringify(next));
  } catch {
    // Demo enhancement only.
  }
}

export function readCompletedProfiles(): string[] {
  if (typeof window === "undefined") return [];
  try {
    const raw = window.localStorage.getItem(COMPLETED_PROFILES_KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw);
    return Array.isArray(parsed) ? parsed.filter((item): item is string => typeof item === "string") : [];
  } catch {
    return [];
  }
}

export function unlockCompletionBonus(shareId: string) {
  if (typeof window === "undefined") return;
  try { window.localStorage.setItem(`${BONUS_PREFIX}${shareId}`, "1"); } catch { /* optional */ }
}

export function hasCompletionBonus(shareId: string) {
  if (typeof window === "undefined") return false;
  try { return window.localStorage.getItem(`${BONUS_PREFIX}${shareId}`) === "1"; } catch { return false; }
}

export function getTotemExplanation(recap: Recap) {
  const behavior = recap.cards.find((card): card is BehaviorCard => card.type === "BEHAVIOR");
  const category = recap.cards.find((card): card is Extract<RecapCard, { type: "TOP_CATEGORY" }> => card.type === "TOP_CATEGORY");
  const month = recap.cards.find((card): card is Extract<RecapCard, { type: "ACTIVE_MONTH" }> => card.type === "ACTIVE_MONTH");
  const achievement = recap.cards.find((card): card is AchievementCard => card.type === "ACHIEVEMENT");
  return [
    { part: "Форма", value: behavior?.title ?? "Разные сценарии", detail: behavior ? `Опирается на ${behavior.payload.code}` : "Универсальный сценарий" },
    { part: "Цвет и объект", value: category?.payload.category ?? "Общий интерес", detail: category ? `Категория ${category.payload.categoryCode}` : "Категория не вошла в recap" },
    { part: "Кольцо", value: month ? `Месяц ${String(month.payload.month).padStart(2, "0")}` : "Ритм года", detail: month ? "Самый активный месяц" : "Активный месяц не определён" },
    { part: "Знаки", value: achievement?.payload.codes.length ? `${achievement.payload.codes.length} ачивки` : "Без ачивок", detail: achievement?.payload.codes.join(" · ") || "Ачивки не сформированы" },
  ];
}
