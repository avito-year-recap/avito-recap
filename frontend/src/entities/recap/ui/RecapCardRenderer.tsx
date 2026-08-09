import { motion, useReducedMotion } from "framer-motion";
import { useEffect, useMemo, useState, type PointerEvent } from "react";
import { useNavigate } from "react-router-dom";
import type { Achievement, Recap, RecapCard } from "../model";
import {
  getAchievementVisual,
  getActionVisual,
  getBehaviorVisual,
  getCategoryVisual,
  getMonthVisual,
} from "../../../shared/lib/visual-registry";
import {
  deriveBehaviorTraits,
  getActionBeforeAfter,
  getDominantActivity,
  getPersonalizedFinalLine,
  getSecretVisualBonus,
} from "../../../shared/lib/experience-utils";
import { Button } from "../../../shared/ui/Button";
import { playHaptic } from "../../../shared/lib/haptics";
import { playUiSound } from "../../../shared/lib/sound";
import { PublicYearTotem, YearTotem } from "../../../shared/ui/YearTotem";
import { CardFrame } from "./CardFrame";

const metrics = [
  ["searches", "⌕", "Поиски"],
  ["totalViews", "◉", "Просмотры"],
  ["favoritesAdded", "♡", "Избранное"],
  ["chatsStarted", "•••", "Диалоги"],
  ["listingsPublished", "▤", "Публикации"],
  ["purchasesCompleted", "◆", "Покупки"],
  ["salesCompleted", "✓", "Продажи"],
] as const;

const months = [
  "Январь", "Февраль", "Март", "Апрель", "Май", "Июнь",
  "Июль", "Август", "Сентябрь", "Октябрь", "Ноябрь", "Декабрь",
];

function ExplainLink({ onClick, label = "Почему я это вижу?" }: { onClick: () => void; label?: string }) {
  return (
    <button className="explain-link" type="button" onClick={onClick}>
      <span aria-hidden="true">?</span>
      {label}
    </button>
  );
}

function Intro({ card, recap }: { card: Extract<RecapCard, { type: "INTRO" }>; recap: Recap }) {
  const reduceMotion = useReducedMotion();
  return (
    <CardFrame
      eyebrow="Твой год на Авито"
      title={card.title}
      description={card.description}
      tone={recap.profile.accent}
      className="recap-card--intro"
    >
      <motion.div
        className="intro-symbol-wrap"
        initial={reduceMotion ? false : { opacity: 0, scale: 0.84, rotate: -6 }}
        animate={{ opacity: 1, scale: 1, rotate: 0 }}
        transition={{ duration: reduceMotion ? 0 : 0.7, ease: [0.22, 1, 0.36, 1] }}
      >
        <YearTotem recap={recap} stage={0} />
        <img className="intro-avatar" src={recap.profile.avatarUrl} alt="" />
      </motion.div>
      <div className="intro-profile intro-profile--minimal">
        <p>{recap.profile.description}</p>
        <div>
          {recap.profile.tags.map((tag) => <span key={tag}>{tag}</span>)}
        </div>
      </div>
    </CardFrame>
  );
}

function Activity({ card, onExplain }: { card: Extract<RecapCard, { type: "YEAR_ACTIVITY" }>; onExplain: () => void }) {
  const reduceMotion = useReducedMotion();
  const dominant = useMemo(() => getDominantActivity(card), [card]);
  return (
    <CardFrame eyebrow="Масштаб года" title={card.title} description={card.description} tone="blue" className="recap-card--activity">
      <div className="activity-show">
        <motion.div
          className="activity-hero"
          initial={reduceMotion ? false : { opacity: 0, y: 18, scale: 0.94 }}
          animate={{ opacity: 1, y: 0, scale: 1 }}
          transition={{ duration: reduceMotion ? 0 : 0.55, ease: [0.22, 1, 0.36, 1] }}
        >
          <strong>{card.payload.totalEvents.toLocaleString("ru-RU")}</strong>
          <span>действий сложились в одну историю</span>
        </motion.div>
        <div className="activity-metrics">
          {metrics.map(([key, icon, label], index) => (
            <motion.article
              key={key}
              className={index < 3 ? "is-featured" : ""}
              initial={reduceMotion ? false : { opacity: 0, y: 14, scale: 0.94 }}
              animate={{ opacity: 1, y: 0, scale: 1 }}
              transition={{ delay: reduceMotion ? 0 : 0.18 + index * 0.055, duration: reduceMotion ? 0 : 0.35 }}
            >
              <span aria-hidden="true">{icon}</span>
              <strong>{card.payload[key].toLocaleString("ru-RU")}</strong>
              <small>{label}</small>
            </motion.article>
          ))}
        </div>
        <motion.div
          className="activity-self-insight"
          initial={reduceMotion ? false : { opacity: 0, y: 8 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: reduceMotion ? 0 : 0.68 }}
        >
          <span aria-hidden="true">↗</span>
          <p><b>{dominant.label}</b> — самая заметная часть среди показанных событий: {dominant.value.toLocaleString("ru-RU")}.</p>
        </motion.div>
      </div>
      <ExplainLink onClick={onExplain} label="Что мы считали?" />
    </CardFrame>
  );
}

function TopCategory({ card, onExplain }: { card: Extract<RecapCard, { type: "TOP_CATEGORY" }>; onExplain: () => void }) {
  const visual = getCategoryVisual(card.payload.categoryCode);
  return (
    <CardFrame eyebrow="Главный интерес" title={card.title} description={card.description} tone={visual.tone} className={`recap-card--category recap-card--motif-${visual.motif ?? "pulse"}`}>
      <div className={`category-world category-world--${visual.tone} category-world--${visual.motif ?? "pulse"}`} aria-hidden="true">
        <motion.span className="category-world__primary" initial={{ scale: 0.75, rotate: -10 }} animate={{ scale: 1, rotate: 0 }} transition={{ type: "spring", stiffness: 170, damping: 18 }}>{visual.icon}</motion.span>
        <motion.span className="category-world__secondary" initial={{ opacity: 0, x: 20, rotate: 12 }} animate={{ opacity: 1, x: 0, rotate: -4 }} transition={{ delay: 0.2 }}>{visual.secondary}</motion.span>
        <i /><i /><i /><div className="category-world__trail" />
      </div>
      <div className="category-stat">
        <strong>{card.payload.categoryViews.toLocaleString("ru-RU")}</strong>
        <div><span>просмотров</span><small>{visual.caption}</small></div>
      </div>
      <ExplainLink onClick={onExplain} />
    </CardFrame>
  );
}

function ActiveMonth({ card, onExplain }: { card: Extract<RecapCard, { type: "ACTIVE_MONTH" }>; onExplain: () => void }) {
  const index = card.payload.month - 1;
  const month = months[index] ?? "Месяц";
  const monthVisual = getMonthVisual(card.payload.month);
  return (
    <CardFrame eyebrow="Момент года" title={card.title} description={card.description} tone="coral" className={`recap-card--month recap-card--season-${monthVisual.season}`}>
      <div className={`month-poster month-poster--${monthVisual.season}`}>
        <span>{String(card.payload.month).padStart(2, "0")}</span>
        <i className="month-poster__season" aria-hidden="true">{monthVisual.icon}</i>
        <div className="month-poster__decor" aria-hidden="true"><i /><i /><i /></div>
        <strong>{month}</strong>
        <p>{monthVisual.caption}</p>
      </div>
      <div className="months-timeline" aria-label={`Самый активный месяц: ${month}`}>
        {months.map((item, i) => <span key={item} className={i === index ? "is-active" : ""} title={item}>{String(i + 1).padStart(2, "0")}</span>)}
      </div>
      <ExplainLink onClick={onExplain} label="Как выбран месяц?" />
    </CardFrame>
  );
}

function Behavior({ card, onExplain }: { card: Extract<RecapCard, { type: "BEHAVIOR" }>; onExplain: () => void }) {
  const visual = getBehaviorVisual(card.payload.code);
  const reduceMotion = useReducedMotion();
  const traits = useMemo(() => deriveBehaviorTraits(card), [card]);
  const [activeTrait, setActiveTrait] = useState(0);
  const [revealed, setRevealed] = useState(Boolean(reduceMotion));
  const trait = traits[activeTrait];

  useEffect(() => {
    if (reduceMotion) return;
    const timeout = window.setTimeout(() => setRevealed(true), 620);
    return () => window.clearTimeout(timeout);
  }, [card.id, reduceMotion]);

  return (
    <CardFrame
      eyebrow="Твой сценарий года"
      title={card.title}
      description={card.description}
      tone={visual.tone}
      className={`recap-card--behavior recap-card--behavior-${visual.motif ?? "pulse"} ${revealed ? "is-revealed" : "is-silent"}`}
      footer={<Button variant="secondary" fullWidth onClick={onExplain}>Разобрать сценарий по данным</Button>}
    >
      {!revealed && (
        <motion.div className="behavior-silence" initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }}>
          <span aria-hidden="true">·</span><p>Кажется, мы тебя поняли.</p>
        </motion.div>
      )}
      <div className="behavior-reveal">
        <motion.p className="behavior-reveal__prelude" initial={reduceMotion ? false : { opacity: 0, y: 8 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: reduceMotion ? 0 : 0.32 }}>
          Кажется, мы поняли, как прошёл твой год.
        </motion.p>
        <motion.div
          className={`behavior-poster behavior-poster--${visual.tone} behavior-poster--${visual.motif ?? "pulse"}`}
          initial={reduceMotion ? false : { opacity: 0, scale: 0.78, rotate: -12 }}
          animate={{ opacity: 1, scale: 1, rotate: 0 }}
          transition={{ delay: reduceMotion ? 0 : 0.22, duration: reduceMotion ? 0 : 0.62, ease: [0.22, 1, 0.36, 1] }}
        >
          <div className="behavior-poster__motif" aria-hidden="true"><i /><i /><i /></div>
          <div className="behavior-poster__mark" aria-hidden="true"><span>{visual.icon}</span><i>{visual.secondary}</i></div>
          <p>{visual.caption}</p>
        </motion.div>
        {traits.length > 0 && (
          <div className="behavior-portrait" aria-label="Портрет поведения">
            <div className="behavior-portrait__tabs">
              {traits.map((item, index) => (
                <button key={item.id} type="button" className={index === activeTrait ? "is-active" : ""} onClick={() => setActiveTrait(index)}>
                  <span aria-hidden="true">{item.icon}</span>{item.title}
                </button>
              ))}
            </div>
            {trait && <p><b>{trait.value}</b><span>{trait.explanation}</span></p>}
          </div>
        )}
      </div>
    </CardFrame>
  );
}

function AchievementCard({ card, achievements, onExplain, soundEnabled }: { card: Extract<RecapCard, { type: "ACHIEVEMENT" }>; achievements: Achievement[]; onExplain: () => void; soundEnabled: boolean }) {
  const reduceMotion = useReducedMotion();
  const [bonusOpen, setBonusOpen] = useState(false);
  const items = card.payload.codes.map((code) => achievements.find((item) => item.code === code)).filter((item): item is Achievement => Boolean(item)).slice(0, 3);
  const bonus = getSecretVisualBonus(card.payload.codes);
  const moveLight = (event: PointerEvent<HTMLElement>) => {
    if (event.pointerType !== "mouse") return;
    const rect = event.currentTarget.getBoundingClientRect();
    event.currentTarget.style.setProperty("--token-light-x", `${((event.clientX - rect.left) / rect.width * 100).toFixed(1)}%`);
    event.currentTarget.style.setProperty("--token-light-y", `${((event.clientY - rect.top) / rect.height * 100).toFixed(1)}%`);
  };

  return (
    <CardFrame eyebrow="Коллекция года" title={card.title} description={card.description} tone="purple" className="recap-card--achievements">
      <div className="achievement-gallery" aria-label="Полученные ачивки">
        {items.map((item, index) => {
          const visual = getAchievementVisual(item.code);
          return (
            <motion.article
              key={item.code}
              className={`achievement-token achievement-token--${visual.tone} achievement-token--${index + 1} achievement-token--${visual.motif ?? "pulse"}`}
              initial={reduceMotion ? false : { opacity: 0, scale: 0.7, y: 24, rotate: index === 0 ? -16 : index === 1 ? 16 : 0 }}
              animate={{ opacity: 1, scale: 1, y: 0, rotate: index === 0 ? -7 : index === 1 ? 8 : -1 }}
              transition={{ delay: reduceMotion ? 0 : 0.12 + index * 0.16, type: "spring", stiffness: 150, damping: 16 }}
              whileHover={reduceMotion ? undefined : { y: -8, rotate: 0, scale: 1.035 }}
              onPointerMove={moveLight}
              onPointerEnter={() => playHaptic("tap")}
            >
              <span className="achievement-token__unlocked">получено</span>
              <div aria-hidden="true">{visual.icon}</div><small>{visual.caption}</small><strong>{item.title}</strong>
            </motion.article>
          );
        })}
        {bonus && (
          <motion.button
            type="button"
            className={`achievement-secret ${bonusOpen ? "is-open" : ""}`}
            onClick={() => { setBonusOpen((value) => !value); playHaptic("secret"); if (soundEnabled) playUiSound("secret"); }}
            initial={reduceMotion ? false : { opacity: 0, scale: 0.7 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ delay: reduceMotion ? 0 : 0.65, type: "spring" }}
            aria-label={bonusOpen ? `Визуальный бонус: ${bonus.title}` : "Открыть секретный визуальный бонус"}
          >
            <span aria-hidden="true">{bonusOpen ? bonus.glyph : "?"}</span>
            <b>{bonusOpen ? bonus.title : "секрет"}</b>
            {bonusOpen && <small>{bonus.caption}. Не backend-ачивка.</small>}
          </motion.button>
        )}
      </div>
      <ExplainLink onClick={onExplain} label="За что получены ачивки?" />
    </CardFrame>
  );
}

function Missed({ card, onExplain }: { card: Extract<RecapCard, { type: "MISSED_OPPORTUNITY" }>; onExplain: () => void }) {
  const visual = getActionVisual(card.payload.code);
  const [demoActive, setDemoActive] = useState(false);
  const isSearch = card.payload.code === "SAVE_SEARCH";
  return (
    <CardFrame eyebrow="Можно сделать проще" title={card.title} description={card.description} tone={visual.tone} className={`recap-card--opportunity ${demoActive ? "is-demo-active" : ""}`}>
      <div className={`opportunity-demo opportunity-demo--${visual.tone} opportunity-demo--${isSearch ? "search" : "draft"}`}>
        <div className="opportunity-demo__canvas" aria-hidden="true">
          {isSearch ? (
            <><span className="opportunity-listing opportunity-listing--one">▤</span><span className="opportunity-listing opportunity-listing--two">▤</span><span className="opportunity-listing opportunity-listing--three">▤</span><div className="opportunity-search-box"><b>⌕</b><i>новые объявления</i></div><div className="opportunity-bell">{demoActive ? "✓" : "♡"}</div></>
          ) : (
            <><div className="opportunity-draft-card"><span>▧</span><i /><i /><i /><b>{demoActive ? "Готово" : "Черновик"}</b></div><div className="opportunity-draft-progress"><span /></div><div className="opportunity-bell">{demoActive ? "✓" : "→"}</div></>
          )}
        </div>
        <div className="opportunity-demo__caption">
          <strong>{demoActive ? (isSearch ? "Теперь новые варианты собраны в одном месте" : "Остался один шаг до публикации") : visual.caption}</strong>
          <button type="button" onClick={() => setDemoActive((value) => !value)}>{demoActive ? "Вернуть" : "Показать, как это работает"}</button>
        </div>
      </div>
      <ExplainLink onClick={onExplain} />
    </CardFrame>
  );
}

function NextActionCard({ card, recap, onAction, onExplain }: { card: Extract<RecapCard, { type: "NEXT_ACTION" }>; recap: Recap; onAction: () => void; onExplain: () => void }) {
  const visual = getActionVisual(card.payload.code);
  const reduceMotion = useReducedMotion();
  const [after, setAfter] = useState(false);
  const beforeAfter = getActionBeforeAfter(card.payload.code);
  const finalLine = getPersonalizedFinalLine(recap);
  return (
    <CardFrame
      eyebrow="Что дальше"
      title={card.title}
      description={card.description}
      tone={visual.tone}
      className={`recap-card--next ${after ? "is-after" : ""}`}
      footer={
        <div className="next-action-footer">
          <motion.div whileHover={reduceMotion ? undefined : { scale: 1.015 }} whileTap={reduceMotion ? undefined : { scale: 0.99 }}>
            <Button fullWidth onClick={onAction}>{recap.nextAction.buttonText}</Button>
          </motion.div>
          <ExplainLink onClick={onExplain} label="Почему рекомендуем это?" />
        </div>
      }
    >
      <div className={`next-action-hero next-action-hero--${visual.tone} next-action-hero--${visual.motif ?? "pulse"}`}>
        <div className="next-action-hero__particles" aria-hidden="true"><i /><i /><i /><i /></div>
        <motion.div
          aria-hidden="true"
          animate={{ scale: after ? 1.06 : 1, rotate: after ? 2 : -5 }}
          initial={reduceMotion ? false : { scale: 0.8, rotate: -12, opacity: 0 }}
          transition={{ type: "spring", stiffness: 150, damping: 17 }}
        >
          <span>{after ? "✓" : visual.icon}</span><i>{after ? visual.icon : visual.secondary}</i>
        </motion.div>
        <p>{finalLine}</p>
        <strong>{after ? beforeAfter.after : beforeAfter.before}</strong>
        <div className="next-action-toggle" role="group" aria-label="Показать эффект следующего шага">
          <button type="button" className={!after ? "is-active" : ""} onClick={() => setAfter(false)}>Сейчас</button>
          <button type="button" className={after ? "is-active" : ""} onClick={() => setAfter(true)}>После шага</button>
        </div>
      </div>
    </CardFrame>
  );
}

function Share({ card, onTrailer, onExploreTotem }: { card: Extract<RecapCard, { type: "SHARE" }>; onTrailer: () => void; onExploreTotem: () => void }) {
  const navigate = useNavigate();
  return (
    <CardFrame eyebrow="Финал" title={card.title} description={card.description} tone="green" className="recap-card--share">
      <div className="share-poster share-poster--symbol">
        <div className="share-poster__topline"><span>Avito · {card.payload.year}</span><i aria-hidden="true">✦</i></div>
        <div className="share-poster__symbol"><PublicYearTotem payload={card.payload} /></div>
        <div className="share-poster__copy">
          <p>Мой сценарий года</p><strong>{card.payload.behaviorTitle}</strong>
          <div className="share-poster__facts">
            {card.payload.achievementTitle && <span><small>Ачивка</small><b>{card.payload.achievementTitle}</b></span>}
            {card.payload.topCategory && <span><small>Главный интерес</small><b>{card.payload.topCategory}</b></span>}
          </div>
        </div>
        <em>#МойГодНаАвито</em>
      </div>
      <div className="share-privacy-row"><span>✓ без имени</span><span>✓ без личной статистики</span><span>✓ только публичные итоги</span></div>
      <div className="share-card-actions share-card-actions--triple">
        <Button variant="secondary" fullWidth onClick={onExploreTotem}>Как собрался символ?</Button>
        <Button variant="secondary" fullWidth onClick={onTrailer}>Смотреть трейлер</Button>
        <Button fullWidth onClick={() => navigate(`/share/${card.payload.shareId}`)}>Настроить и поделиться</Button>
      </div>
    </CardFrame>
  );
}

export function RecapCardRenderer({
  card,
  recap,
  onExplain,
  onAction,
  onTrailer,
  onExploreTotem,
  soundEnabled,
}: {
  card: RecapCard;
  recap: Recap;
  onExplain: (card: RecapCard) => void;
  onAction: () => void;
  onTrailer: () => void;
  onExploreTotem: () => void;
  soundEnabled: boolean;
}) {
  switch (card.type) {
    case "INTRO": return <Intro card={card} recap={recap} />;
    case "YEAR_ACTIVITY": return <Activity card={card} onExplain={() => onExplain(card)} />;
    case "TOP_CATEGORY": return <TopCategory card={card} onExplain={() => onExplain(card)} />;
    case "ACTIVE_MONTH": return <ActiveMonth card={card} onExplain={() => onExplain(card)} />;
    case "BEHAVIOR": return <Behavior card={card} onExplain={() => onExplain(card)} />;
    case "ACHIEVEMENT": return <AchievementCard card={card} achievements={recap.achievements} onExplain={() => onExplain(card)} soundEnabled={soundEnabled} />;
    case "MISSED_OPPORTUNITY": return <Missed card={card} onExplain={() => onExplain(card)} />;
    case "NEXT_ACTION": return <NextActionCard card={card} recap={recap} onAction={onAction} onExplain={() => onExplain(card)} />;
    case "SHARE": return <Share card={card} onTrailer={onTrailer} onExploreTotem={onExploreTotem} />;
  }
}
