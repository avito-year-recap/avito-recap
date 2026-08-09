import type { PublicSharePayload, Recap } from "../../entities/recap/model";
import {
  getAchievementVisual,
  getBehaviorVisual,
  getCategoryVisual,
} from "../lib/visual-registry";
import "./year-totem.css";

function findCard<T extends Recap["cards"][number]["type"]>(recap: Recap, type: T) {
  return recap.cards.find((card) => card.type === type) as Extract<Recap["cards"][number], { type: T }> | undefined;
}

function clampStage(stage: number) {
  return Math.max(0, Math.min(4, stage));
}

export function YearTotem({
  recap,
  stage = 4,
  compact = false,
  className = "",
}: {
  recap: Recap;
  stage?: number;
  compact?: boolean;
  className?: string;
}) {
  const behavior = findCard(recap, "BEHAVIOR");
  const category = findCard(recap, "TOP_CATEGORY");
  const month = findCard(recap, "ACTIVE_MONTH");
  const achievement = findCard(recap, "ACHIEVEMENT");
  const behaviorVisual = getBehaviorVisual(behavior?.payload.code ?? "UNIVERSAL_USER");
  const categoryVisual = getCategoryVisual(category?.payload.categoryCode ?? "UNKNOWN");
  const achievementCodes = achievement?.payload.codes.slice(0, 3) ?? [];
  const safeStage = clampStage(stage);

  return (
    <div
      className={`year-symbol year-symbol--${behaviorVisual.tone} year-symbol--stage-${safeStage}${compact ? " year-symbol--compact" : ""} ${className}`.trim()}
      aria-hidden="true"
    >
      <div className="year-symbol__halo" />
      <div className="year-symbol__orbit year-symbol__orbit--outer" />
      <div className="year-symbol__orbit year-symbol__orbit--inner" />
      <div className="year-symbol__core">
        <span>{behaviorVisual.icon}</span>
        {!compact && <small>{recap.year}</small>}
      </div>
      <div className="year-symbol__category" title={category?.payload.category}>
        {categoryVisual.icon}
      </div>
      <div className="year-symbol__month">{String(month?.payload.month ?? ((recap.year % 12) || 1)).padStart(2, "0")}</div>
      <div className="year-symbol__achievement year-symbol__achievement--one">
        {getAchievementVisual(achievementCodes[0] ?? "BROAD_INTERESTS").icon}
      </div>
      <div className="year-symbol__achievement year-symbol__achievement--two">
        {getAchievementVisual(achievementCodes[1] ?? "MASTER_OF_FAVORITES").icon}
      </div>
      <div className="year-symbol__achievement year-symbol__achievement--three">
        {getAchievementVisual(achievementCodes[2] ?? "ATTENTIVE_RESEARCHER").icon}
      </div>
      {!compact && <div className="year-symbol__caption">твой визуальный код года</div>}
    </div>
  );
}

function hashText(text: string) {
  let hash = 2166136261;
  for (let index = 0; index < text.length; index += 1) {
    hash ^= text.charCodeAt(index);
    hash = Math.imul(hash, 16777619);
  }
  return Math.abs(hash >>> 0);
}

const publicGlyphs = ["✦", "◎", "◇", "↗", "♡", "⌕"];

export function PublicYearTotem({
  payload,
  compact = false,
}: {
  payload: PublicSharePayload;
  compact?: boolean;
}) {
  const source = `${payload.behaviorTitle}|${payload.achievementTitle ?? ""}|${payload.topCategory ?? ""}|${payload.year}`;
  const hash = hashText(source);
  const tone = (["blue", "green", "purple", "coral"] as const)[hash % 4];
  const glyph = publicGlyphs[hash % publicGlyphs.length];
  const second = publicGlyphs[(hash >> 3) % publicGlyphs.length];
  const third = publicGlyphs[(hash >> 6) % publicGlyphs.length];

  return (
    <div className={`year-symbol year-symbol--public year-symbol--${tone} year-symbol--stage-4${compact ? " year-symbol--compact" : ""}`} aria-hidden="true">
      <div className="year-symbol__halo" />
      <div className="year-symbol__orbit year-symbol__orbit--outer" />
      <div className="year-symbol__orbit year-symbol__orbit--inner" />
      <div className="year-symbol__core"><span>{glyph}</span>{!compact && <small>{payload.year}</small>}</div>
      <div className="year-symbol__category">{second}</div>
      <div className="year-symbol__month">{String((hash % 12) + 1).padStart(2, "0")}</div>
      <div className="year-symbol__achievement year-symbol__achievement--one">{third}</div>
      <div className="year-symbol__achievement year-symbol__achievement--two">✦</div>
      <div className="year-symbol__achievement year-symbol__achievement--three">◇</div>
    </div>
  );
}
