import type { BackendRecapResponse } from "./backend-contract";
import type {
  Achievement,
  BehaviorEvidence,
  RecapCard,
} from "../../entities/recap/model";

const explanation =
  "Учитывались поиски, просмотры, избранное, диалоги, создание и публикация объявлений, покупки и продажи.";

function intro(
  id: string,
  name: string,
  year: number,
  position = 1,
): RecapCard {
  return {
    id: `${id}-intro`,
    type: "INTRO",
    position,
    title: `${name}, вот твои итоги за ${year} год`,
    description: "Это финальный recap за завершённый календарный год.",
    shareable: false,
  };
}

function activity(
  id: string,
  position: number,
  totalEvents: number,
  values: Omit<
    Extract<RecapCard, { type: "YEAR_ACTIVITY" }>["payload"],
    "totalEvents"
  >,
): RecapCard {
  return {
    id: `${id}-activity`,
    type: "YEAR_ACTIVITY",
    position,
    title: "Год в цифрах",
    description: `За завершённый год система учла ${totalEvents.toLocaleString("ru-RU")} действий.`,
    explanation,
    shareable: false,
    payload: { totalEvents, ...values },
  };
}

function behavior(
  id: string,
  position: number,
  title: string,
  description: string,
  reason: string,
  code: string,
  score: number,
  evidence: BehaviorEvidence[],
): RecapCard {
  return {
    id: `${id}-behavior`,
    type: "BEHAVIOR",
    position,
    title,
    description,
    explanation: reason,
    shareable: false,
    payload: { code, score, evidence },
  };
}

function next(
  id: string,
  position: number,
  title: string,
  description: string,
  reason: string,
  code: Extract<RecapCard, { type: "NEXT_ACTION" }>["payload"]["code"],
  target: Extract<RecapCard, { type: "NEXT_ACTION" }>["payload"]["target"],
): RecapCard {
  return {
    id: `${id}-action`,
    type: "NEXT_ACTION",
    position,
    title,
    description,
    explanation: reason,
    shareable: false,
    payload: { code, target },
  };
}

function share(
  id: string,
  position: number,
  year: number,
  behaviorTitle: string,
  achievementTitle?: string,
  topCategory?: string,
): RecapCard {
  return {
    id: `${id}-share`,
    type: "SHARE",
    position,
    title: "Итоги готовы — можно делиться",
    description: `В ${year} году ты — «${behaviorTitle}». Карточка содержит только данные, разрешённые для публичного показа.`,
    explanation:
      "Годовой recap сохранён как неизменяемый снимок состояния на момент генерации.",
    shareable: true,
    payload: {
      shareId: `share-${id}-${year}`,
      year,
      behaviorTitle,
      achievementTitle,
      topCategory,
    },
  };
}

const marinaAchievements: Achievement[] = [
  {
    code: "ATTENTIVE_RESEARCHER",
    title: "Внимательное сравнение",
    reason: "Просмотрено не меньше 150 объявлений.",
    shareable: true,
  },
  {
    code: "MASTER_OF_FAVORITES",
    title: "Коллекция находок",
    reason: "В избранное добавлено не меньше 20 объявлений.",
    shareable: true,
  },
  {
    code: "BROAD_INTERESTS",
    title: "Широкий круг интересов",
    reason: "Активность была минимум в 6 категориях.",
    shareable: true,
  },
];

const marina: BackendRecapResponse = {
  profile: {
    name: "Марина",
    description: "Любит уют, сравнивает варианты и сохраняет идеи для дома",
    avatarUrl: "/avatars/marina.svg",
    profileCode: "MARINA_RESEARCHER",
  },
  recap: {
    id: "recap-marina-2025-v3-1-0",
    year: 2025,
    ruleVersion: "3.1.0",
    cards: [
      intro("marina", "Марина", 2025),
      activity("marina", 2, 1785, {
        searches: 356,
        totalViews: 1284,
        favoritesAdded: 96,
        chatsStarted: 18,
        listingsPublished: 24,
        purchasesCompleted: 4,
        salesCompleted: 3,
      }),
      {
        id: "marina-category",
        type: "TOP_CATEGORY",
        position: 3,
        title: "Главный интерес года",
        description: "Чаще всего внимание привлекала категория «Дом и дача».",
        explanation: "Просмотров в этой категории: 538.",
        shareable: false,
        payload: {
          categoryCode: "HOME_AND_GARDEN",
          category: "Дом и дача",
          categoryViews: 538,
        },
      },
      {
        id: "marina-month",
        type: "ACTIVE_MONTH",
        position: 4,
        title: "Самый активный месяц",
        description: "Больше всего действий пришлось на октябрь.",
        explanation: "Месяц выбран по максимальному числу событий профиля.",
        shareable: false,
        payload: { month: 10 },
      },
      behavior(
        "marina",
        5,
        "Глубокое исследование",
        "Ты подробно изучаешь варианты, сравниваешь и только потом принимаешь решение.",
        "Этот сценарий набрал больше всего баллов по просмотрам, широте интересов и небольшому числу диалогов.",
        "RESEARCHER",
        80,
        [
          {
            metric: "totalViews",
            label: "Просмотры",
            actualValue: 1284,
            threshold: 100,
            comparison: "GTE",
            points: 35,
            explanation: "Просмотров было значительно больше порога.",
          },
          {
            metric: "categoriesCount",
            label: "Категории",
            actualValue: 7,
            threshold: 5,
            comparison: "GTE",
            points: 25,
            explanation: "Интерес распределился минимум по пяти категориям.",
          },
          {
            metric: "chatsStarted",
            label: "Диалоги",
            actualValue: 3,
            threshold: 4,
            comparison: "LTE",
            points: 20,
            explanation: "Диалогов было не больше четырёх.",
          },
        ],
      ),
      {
        id: "marina-achievement",
        type: "ACHIEVEMENT",
        position: 6,
        title: "Ачивки года",
        description:
          "Внимательное сравнение • Коллекция находок • Широкий круг интересов",
        explanation:
          "Три ачивки с максимальным приоритетом собраны в одном экране.",
        shareable: false,
        payload: { codes: marinaAchievements.map((item) => item.code) },
      },
      {
        id: "marina-missed",
        type: "MISSED_OPPORTUNITY",
        position: 7,
        title: "Возможность, которая сэкономит время",
        description:
          "Сохранённый поиск может сам показывать новые объявления по главному интересу.",
        explanation: "Просмотров было 1 284, а сохранённого поиска сейчас нет.",
        shareable: false,
        payload: {
          code: "SAVE_SEARCH",
          target: { type: "search", categoryCode: "HOME_AND_GARDEN" },
        },
      },
      next(
        "marina",
        8,
        "Продолжим исследование?",
        "Сохрани поиск — новые объявления по категории будут ждать тебя в одном месте.",
        "Рекомендация выбрана, потому что ты много смотрела объявления, но ещё не сохранила поиск.",
        "SAVE_SEARCH",
        { type: "search", categoryCode: "HOME_AND_GARDEN" },
      ),
      share(
        "marina",
        9,
        2025,
        "Глубокое исследование",
        "Внимательное сравнение",
        "Дом и дача",
      ),
    ],
    achievements: marinaAchievements,
    nextAction: {
      code: "SAVE_SEARCH",
      title: "Продолжим исследование?",
      description:
        "Новые объявления по главному интересу будут появляться автоматически.",
      explanation:
        "Главная категория определена, сохранённого поиска пока нет.",
      buttonText: "Сохранить поиск",
      target: { type: "search", categoryCode: "HOME_AND_GARDEN" },
    },
  },
};

const ilyaAchievements: Achievement[] = [
  {
    code: "SUCCESSFUL_SELLER",
    title: "Успешные продажи",
    reason: "Завершено не меньше пяти продаж.",
    shareable: true,
  },
  {
    code: "CONSISTENT_PUBLISHER",
    title: "Стабильные публикации",
    reason:
      "Опубликовано не меньше пяти объявлений и была хотя бы одна продажа.",
    shareable: true,
  },
  {
    code: "ALL_ROUNDER",
    title: "Всё в одном году",
    reason: "Были покупка, продажа, публикация и не меньше трёх диалогов.",
    shareable: true,
  },
];

const ilya: BackendRecapResponse = {
  profile: {
    name: "Илья",
    description: "Активно публикует объявления и доводит продажи до результата",
    avatarUrl: "/avatars/ilya.svg",
    profileCode: "ILYA_ACTIVE_SELLER",
  },
  recap: {
    id: "recap-ilya-2025-v3-1-0",
    year: 2025,
    ruleVersion: "3.1.0",
    cards: [
      intro("ilya", "Илья", 2025),
      activity("ilya", 2, 684, {
        searches: 42,
        totalViews: 410,
        favoritesAdded: 12,
        chatsStarted: 96,
        listingsPublished: 18,
        purchasesCompleted: 1,
        salesCompleted: 9,
      }),
      {
        id: "ilya-category",
        type: "TOP_CATEGORY",
        position: 3,
        title: "Главный интерес года",
        description: "Чаще всего внимание привлекала категория «Электроника».",
        explanation: "Просмотров в этой категории: 206.",
        shareable: false,
        payload: {
          categoryCode: "ELECTRONICS",
          category: "Электроника",
          categoryViews: 206,
        },
      },
      behavior(
        "ilya",
        4,
        "Продажи в движении",
        "Ты регулярно публикуешь объявления и доводишь общение до завершённых продаж.",
        "Сценарий выбран по числу публикаций и успешных продаж.",
        "ACTIVE_SELLER",
        120,
        [
          {
            metric: "listingsPublished",
            label: "Публикации",
            actualValue: 18,
            threshold: 5,
            comparison: "GTE",
            points: 60,
            explanation: "Опубликовано не меньше пяти объявлений.",
          },
          {
            metric: "salesCompleted",
            label: "Продажи",
            actualValue: 9,
            threshold: 3,
            comparison: "GTE",
            points: 60,
            explanation: "Завершено не меньше трёх продаж.",
          },
        ],
      ),
      {
        id: "ilya-achievement",
        type: "ACHIEVEMENT",
        position: 5,
        title: "Ачивки года",
        description:
          "Успешные продажи • Стабильные публикации • Всё в одном году",
        explanation: "Показываются три ачивки с максимальным приоритетом.",
        shareable: false,
        payload: { codes: ilyaAchievements.map((item) => item.code) },
      },
      next(
        "ilya",
        6,
        "Есть место для следующей продажи",
        "Создай новое объявление и продолжи успешный сценарий.",
        "Для активного продавца следующий логичный шаг — новая публикация.",
        "CREATE_LISTING",
        { type: "route", path: "/listings/new" },
      ),
      share(
        "ilya",
        7,
        2025,
        "Продажи в движении",
        "Успешные продажи",
        "Электроника",
      ),
    ],
    achievements: ilyaAchievements,
    nextAction: {
      code: "CREATE_LISTING",
      title: "Есть место для следующей продажи",
      description: "Продолжи успешный сценарий новым объявлением.",
      explanation: "Профиль соответствует сценарию активного продавца.",
      buttonText: "Создать объявление",
      target: { type: "route", path: "/listings/new" },
    },
  },
};

const alexeyAchievements: Achievement[] = [
  {
    code: "DEAL_CLOSER",
    title: "Сделка состоялась",
    reason: "Завершено не меньше трёх покупок.",
    shareable: true,
  },
  {
    code: "QUICK_DECISION",
    title: "Быстрое решение",
    reason: "Покупок не меньше трёх, purchaseRate не меньше 20%.",
    shareable: true,
  },
  {
    code: "ATTENTIVE_RESEARCHER",
    title: "Внимательное сравнение",
    reason: "Просмотрено не меньше 150 объявлений.",
    shareable: true,
  },
];

const alexey: BackendRecapResponse = {
  profile: {
    name: "Алексей",
    description:
      "Сравнивает автомобили, общается с продавцами и принимает решения",
    avatarUrl: "/avatars/alexey.svg",
    profileCode: "ALEXEY_DECISIVE_BUYER",
  },
  recap: {
    id: "recap-alexey-2025-v3-1-0",
    year: 2025,
    ruleVersion: "3.1.0",
    cards: [
      intro("alexey", "Алексей", 2025),
      activity("alexey", 2, 1036, {
        searches: 118,
        totalViews: 822,
        favoritesAdded: 41,
        chatsStarted: 37,
        listingsPublished: 3,
        purchasesCompleted: 7,
        salesCompleted: 1,
      }),
      {
        id: "alexey-category",
        type: "TOP_CATEGORY",
        position: 3,
        title: "Главный интерес года",
        description: "Чаще всего внимание привлекала категория «Автомобили».",
        explanation: "Просмотров в этой категории: 614.",
        shareable: false,
        payload: {
          categoryCode: "CARS",
          category: "Автомобили",
          categoryViews: 614,
        },
      },
      {
        id: "alexey-month",
        type: "ACTIVE_MONTH",
        position: 4,
        title: "Самый активный месяц",
        description: "Больше всего действий пришлось на август.",
        explanation: "Месяц выбран по максимальному числу событий профиля.",
        shareable: false,
        payload: { month: 8 },
      },
      behavior(
        "alexey",
        5,
        "Решительный покупатель",
        "Ты не останавливаешься на просмотрах: начинаешь диалоги и доводишь выбор до покупки.",
        "Сценарий выбран по числу покупок, диалогов и доле диалогов с покупкой.",
        "DECISIVE_BUYER",
        110,
        [
          {
            metric: "purchasesCompleted",
            label: "Покупки",
            actualValue: 7,
            threshold: 3,
            comparison: "GTE",
            points: 40,
            explanation: "Завершено не меньше трёх покупок.",
          },
          {
            metric: "chatsStarted",
            label: "Диалоги",
            actualValue: 37,
            threshold: 5,
            comparison: "GTE",
            points: 35,
            explanation: "Начато не меньше пяти диалогов.",
          },
          {
            metric: "purchaseRate",
            label: "Диалоги с покупкой, %",
            actualValue: 24,
            threshold: 20,
            comparison: "GTE",
            points: 35,
            explanation: "Доля диалогов с покупкой не меньше 20%.",
          },
        ],
      ),
      {
        id: "alexey-achievement",
        type: "ACHIEVEMENT",
        position: 6,
        title: "Ачивки года",
        description:
          "Сделка состоялась • Быстрое решение • Внимательное сравнение",
        explanation: "Показываются три ачивки с максимальным приоритетом.",
        shareable: false,
        payload: { codes: alexeyAchievements.map((item) => item.code) },
      },
      next(
        "alexey",
        7,
        "Посмотрим похожие варианты?",
        "Открой объявления, похожие на последнюю завершённую покупку.",
        "Для решительного покупателя подготовлены похожие предложения.",
        "VIEW_SIMILAR_LISTINGS",
        { type: "listing", listingId: "listing-auto-4821" },
      ),
      share(
        "alexey",
        8,
        2025,
        "Решительный покупатель",
        "Сделка состоялась",
        "Автомобили",
      ),
    ],
    achievements: alexeyAchievements,
    nextAction: {
      code: "VIEW_SIMILAR_LISTINGS",
      title: "Посмотрим похожие варианты?",
      description: "Открой предложения, похожие на последнюю покупку.",
      explanation:
        "Для решительного покупателя выбран сценарий похожих объявлений.",
      buttonText: "Смотреть похожее",
      target: { type: "listing", listingId: "listing-auto-4821" },
    },
  },
};

const sonya: BackendRecapResponse = {
  profile: {
    name: "Соня",
    description:
      "Пробует разные категории — доминирующий сценарий пока не определился",
    avatarUrl: "/avatars/sonya.svg",
    profileCode: "SONYA_UNIVERSAL",
  },
  recap: {
    id: "recap-sonya-2025-v3-1-0",
    year: 2025,
    ruleVersion: "3.1.0",
    achievements: [],
    cards: [
      intro("sonya", "Соня", 2025),
      activity("sonya", 2, 73, {
        searches: 21,
        totalViews: 43,
        favoritesAdded: 4,
        chatsStarted: 2,
        listingsPublished: 1,
        purchasesCompleted: 1,
        salesCompleted: 1,
      }),
      behavior(
        "sonya",
        3,
        "Разные сценарии",
        "В этом году не было одного доминирующего сценария — ты пробовала продукт по-разному.",
        "Специализированные правила не сработали, поэтому выбран универсальный профиль.",
        "UNIVERSAL_USER",
        0,
        [],
      ),
      next(
        "sonya",
        4,
        "Продолжим знакомство?",
        "В рекомендациях собраны объявления из разных категорий.",
        "Для универсального сценария используется безопасный fallback.",
        "EXPLORE_RECOMMENDATIONS",
        { type: "route", path: "/recommendations" },
      ),
      share("sonya", 5, 2025, "Разные сценарии"),
    ],
    nextAction: {
      code: "EXPLORE_RECOMMENDATIONS",
      title: "Продолжим знакомство?",
      description: "Посмотри рекомендации, собранные из разных категорий.",
      explanation: "Специализированный CTA не выбран, используется fallback.",
      buttonText: "Открыть рекомендации",
      target: { type: "route", path: "/recommendations" },
    },
  },
};

const dashaAchievements: Achievement[] = [
  {
    code: "FIRST_SELLING_STEPS",
    title: "Первые шаги в продажах",
    reason:
      "Создано не меньше трёх объявлений, опубликовано меньше, продаж пока нет.",
    shareable: true,
  },
];

const dasha: BackendRecapResponse = {
  profile: {
    name: "Даша",
    description:
      "Начала готовить объявления, но один актуальный черновик ещё ждёт публикации",
    avatarUrl: "/avatars/dasha.svg",
    profileCode: "DASHA_STARTING_SELLER",
  },
  recap: {
    id: "recap-dasha-2025-v3-1-0",
    year: 2025,
    ruleVersion: "3.1.0",
    achievements: dashaAchievements,
    cards: [
      intro("dasha", "Даша", 2025),
      activity("dasha", 2, 184, {
        searches: 19,
        totalViews: 132,
        favoritesAdded: 8,
        chatsStarted: 4,
        listingsPublished: 2,
        purchasesCompleted: 1,
        salesCompleted: 0,
      }),
      behavior(
        "dasha",
        3,
        "Старт в продажах",
        "Ты уже подготовила несколько объявлений, но часть пути до публикации ещё впереди.",
        "Сценарий выбран по числу созданных и опубликованных объявлений при отсутствии завершённых продаж.",
        "STARTING_SELLER",
        96,
        [
          {
            metric: "listingsCreated",
            label: "Созданные объявления",
            actualValue: 5,
            threshold: 3,
            comparison: "GTE",
            points: 48,
            explanation: "Создано не меньше трёх объявлений.",
          },
          {
            metric: "listingsPublished",
            label: "Опубликованные объявления",
            actualValue: 2,
            threshold: 2,
            comparison: "LTE",
            points: 48,
            explanation: "Опубликовано не больше двух объявлений.",
          },
        ],
      ),
      {
        id: "dasha-achievement",
        type: "ACHIEVEMENT",
        position: 4,
        title: "Ачивки года",
        description: "Первые шаги в продажах",
        explanation:
          "Полные данные ачивки связаны с кодом из recap.achievements.",
        shareable: false,
        payload: { codes: ["FIRST_SELLING_STEPS"] },
      },
      {
        id: "dasha-missed",
        type: "MISSED_OPPORTUNITY",
        position: 5,
        title: "Шанс довести актуальный черновик до публикации",
        description: "Текущий черновик можно открыть напрямую.",
        explanation:
          "Рекомендация опирается на текущее состояние профиля и известный ID черновика.",
        shareable: false,
        payload: {
          code: "FINISH_DRAFT",
          target: { type: "listing", listingId: "draft-listing-1904" },
        },
      },
      next(
        "dasha",
        6,
        "Объявление почти готово",
        "Вернись к актуальному черновику и закончи публикацию.",
        "Актуальный черновик с известным ID имеет приоритет над годовыми рекомендациями.",
        "FINISH_DRAFT",
        { type: "listing", listingId: "draft-listing-1904" },
      ),
      share("dasha", 7, 2025, "Старт в продажах", "Первые шаги в продажах"),
    ],
    nextAction: {
      code: "FINISH_DRAFT",
      title: "Объявление почти готово",
      description: "Открой актуальный черновик и закончи публикацию.",
      explanation:
        "Текущее состояние профиля содержит известный ID актуального черновика.",
      buttonText: "Открыть черновик",
      target: { type: "listing", listingId: "draft-listing-1904" },
    },
  },
};

export const mockRecaps: BackendRecapResponse[] = [
  marina,
  ilya,
  alexey,
  sonya,
  dasha,
];
