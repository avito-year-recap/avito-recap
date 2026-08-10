import { motion, useReducedMotion } from "framer-motion";
import { useNavigate } from "react-router-dom";
import type { Achievement, Recap, RecapCard } from "../model";
import {
  getAchievementVisual,
  getActionVisual,
  getBehaviorVisual,
  getCategoryVisual,
  getMonthVisual,
} from "../../../shared/lib/visual-registry";
import { Button } from "../../../shared/ui/Button";
import { PublicYearTotem } from "../../../shared/ui/YearTotem";
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
        className="intro-avatar-stage"
        initial={reduceMotion ? false : { opacity: 0, scale: 0.92 }}
        animate={{ opacity: 1, scale: 1 }}
        transition={{ duration: reduceMotion ? 0 : 0.45, ease: [0.22, 1, 0.36, 1] }}
      >
        <div className="intro-avatar-stage__halo" aria-hidden="true" />
        <img src={recap.profile.avatarUrl} alt="" className="intro-avatar-stage__avatar" />
        <span className="intro-avatar-stage__year">{recap.year}</span>
      </motion.div>
      <p className="intro-profile-description">{recap.profile.description}</p>
    </CardFrame>
  );
}

function Activity({ card }: { card: Extract<RecapCard, { type: "YEAR_ACTIVITY" }> }) {
  const reduceMotion = useReducedMotion();
  return (
    <CardFrame eyebrow="Год в цифрах" title={card.title} description={card.description} tone="blue" className="recap-card--activity">
      <div className="activity-show">
        <motion.div
          className="activity-hero"
          initial={reduceMotion ? false : { opacity: 0, y: 12 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: reduceMotion ? 0 : 0.42 }}
        >
          <strong>{card.payload.totalEvents.toLocaleString("ru-RU")}</strong>
          <span>действий за год</span>
        </motion.div>
        <div className="activity-metrics">
          {metrics.map(([key, icon, label], index) => (
            <motion.article
              key={key}
              initial={reduceMotion ? false : { opacity: 0, y: 8 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: reduceMotion ? 0 : 0.08 + index * 0.035 }}
            >
              <span aria-hidden="true">{icon}</span>
              <strong>{card.payload[key].toLocaleString("ru-RU")}</strong>
              <small>{label}</small>
            </motion.article>
          ))}
        </div>
      </div>
    </CardFrame>
  );
}

function TopCategory({ card }: { card: Extract<RecapCard, { type: "TOP_CATEGORY" }> }) {
  const visual = getCategoryVisual(card.payload.categoryCode);
  return (
    <CardFrame eyebrow="Главный интерес" title={card.title} description={card.description} tone={visual.tone} className="recap-card--category">
      <div className={`category-world category-world--${visual.tone}`} aria-hidden="true">
        <motion.span className="category-world__primary" initial={{ scale: 0.82 }} animate={{ scale: 1 }} transition={{ type: "spring", stiffness: 170, damping: 18 }}>{visual.icon}</motion.span>
        <motion.span className="category-world__secondary" initial={{ opacity: 0, x: 12 }} animate={{ opacity: 1, x: 0 }} transition={{ delay: 0.15 }}>{visual.secondary}</motion.span>
        <i /><i /><i />
      </div>
      <div className="category-stat">
        <strong>{card.payload.categoryViews.toLocaleString("ru-RU")}</strong>
        <div><span>просмотров</span><small>{card.payload.category}</small></div>
      </div>
    </CardFrame>
  );
}

function ActiveMonth({ card }: { card: Extract<RecapCard, { type: "ACTIVE_MONTH" }> }) {
  const index = card.payload.month - 1;
  const month = months[index] ?? "Месяц";
  const monthVisual = getMonthVisual(card.payload.month);
  return (
    <CardFrame eyebrow="Момент года" title={card.title} description={card.description} tone="coral" className={`recap-card--month recap-card--season-${monthVisual.season}`}>
      <div className={`month-poster month-poster--${monthVisual.season}`}>
        <span>{String(card.payload.month).padStart(2, "0")}</span>
        <i className="month-poster__season" aria-hidden="true">{monthVisual.icon}</i>
        <strong>{month}</strong>
      </div>
      <div className="months-timeline" aria-label={`Самый активный месяц: ${month}`}>
        {months.map((item, i) => <span key={item} className={i === index ? "is-active" : ""} title={item}>{String(i + 1).padStart(2, "0")}</span>)}
      </div>
    </CardFrame>
  );
}

function Behavior({ card }: { card: Extract<RecapCard, { type: "BEHAVIOR" }> }) {
  const visual = getBehaviorVisual(card.payload.code);
  const reduceMotion = useReducedMotion();
  return (
    <CardFrame
      eyebrow="Твой сценарий года"
      title={card.title}
      description={card.description}
      tone={visual.tone}
      className={`recap-card--behavior recap-card--behavior-${visual.motif ?? "pulse"}`}
    >
      <motion.div
        className={`behavior-poster behavior-poster--${visual.tone} behavior-poster--${visual.motif ?? "pulse"}`}
        initial={reduceMotion ? false : { opacity: 0, scale: 0.9 }}
        animate={{ opacity: 1, scale: 1 }}
        transition={{ duration: reduceMotion ? 0 : 0.45, ease: [0.22, 1, 0.36, 1] }}
      >
        <div className="behavior-poster__motif" aria-hidden="true"><i /><i /><i /></div>
        <div className="behavior-poster__mark" aria-hidden="true"><span>{visual.icon}</span><i>{visual.secondary}</i></div>
      </motion.div>
    </CardFrame>
  );
}

function AchievementCard({ card, achievements }: { card: Extract<RecapCard, { type: "ACHIEVEMENT" }>; achievements: Achievement[] }) {
  const reduceMotion = useReducedMotion();
  const items = card.payload.codes
    .map((code) => achievements.find((item) => item.code === code))
    .filter((item): item is Achievement => Boolean(item))
    .slice(0, 3);

  return (
    <CardFrame eyebrow="Коллекция года" title={card.title} description={card.description} tone="purple" className="recap-card--achievements">
      <div className={`achievement-gallery achievement-gallery--count-${items.length}`} aria-label="Полученные ачивки">
        {items.map((item, index) => {
          const visual = getAchievementVisual(item.code);
          return (
            <motion.article
              key={item.code}
              className={`achievement-token achievement-token--${visual.tone}`}
              initial={reduceMotion ? false : { opacity: 0, y: 16, scale: 0.96 }}
              animate={{ opacity: 1, y: 0, scale: 1 }}
              transition={{ delay: reduceMotion ? 0 : 0.08 + index * 0.09, type: "spring", stiffness: 170, damping: 18 }}
            >
              <span className="achievement-token__bar" aria-hidden="true" />
              <div aria-hidden="true">{visual.icon}</div>
              <strong>{item.title}</strong>
            </motion.article>
          );
        })}
      </div>
    </CardFrame>
  );
}

function Missed({ card }: { card: Extract<RecapCard, { type: "MISSED_OPPORTUNITY" }> }) {
  const visual = getActionVisual(card.payload.code);
  return (
    <CardFrame eyebrow="Можно сделать проще" title={card.title} description={card.description} tone={visual.tone} className="recap-card--opportunity">
      <div className={`opportunity-visual opportunity-visual--${visual.tone}`} aria-hidden="true">
        <div className="opportunity-visual__main"><span>{visual.icon}</span><i>{visual.secondary}</i></div>
        <div className="opportunity-visual__cards"><i /><i /><i /></div>
      </div>
    </CardFrame>
  );
}

function NextActionCard({ card, recap, onAction }: { card: Extract<RecapCard, { type: "NEXT_ACTION" }>; recap: Recap; onAction: () => void }) {
  const visual = getActionVisual(card.payload.code);
  const reduceMotion = useReducedMotion();
  return (
    <CardFrame
      eyebrow="Что дальше"
      title={card.title}
      description={card.description}
      tone={visual.tone}
      className="recap-card--next"
      footer={
        <motion.div whileHover={reduceMotion ? undefined : { scale: 1.01 }} whileTap={reduceMotion ? undefined : { scale: 0.99 }}>
          <Button fullWidth onClick={onAction}>{recap.nextAction.buttonText}</Button>
        </motion.div>
      }
    >
      <div className={`next-action-hero next-action-hero--${visual.tone}`}>
        <motion.div
          aria-hidden="true"
          initial={reduceMotion ? false : { scale: 0.86, opacity: 0 }}
          animate={{ scale: 1, opacity: 1 }}
          transition={{ type: "spring", stiffness: 160, damping: 18 }}
        >
          <span>{visual.icon}</span><i>{visual.secondary}</i>
        </motion.div>
      </div>
    </CardFrame>
  );
}

function Share({ card, onTrailer }: { card: Extract<RecapCard, { type: "SHARE" }>; onTrailer: () => void }) {
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
      <div className="share-privacy-row"><span>✓ без имени</span><span>✓ без личной статистики</span></div>
      <div className="share-card-actions">
        <Button variant="secondary" fullWidth onClick={onTrailer}>Смотреть трейлер</Button>
        <Button fullWidth onClick={() => navigate(`/share/${card.payload.shareId}`)}>Поделиться</Button>
      </div>
    </CardFrame>
  );
}

export function RecapCardRenderer({
  card,
  recap,
  onAction,
  onTrailer,
}: {
  card: RecapCard;
  recap: Recap;
  onAction: () => void;
  onTrailer: () => void;
}) {
  switch (card.type) {
    case "INTRO": return <Intro card={card} recap={recap} />;
    case "YEAR_ACTIVITY": return <Activity card={card} />;
    case "TOP_CATEGORY": return <TopCategory card={card} />;
    case "ACTIVE_MONTH": return <ActiveMonth card={card} />;
    case "BEHAVIOR": return <Behavior card={card} />;
    case "ACHIEVEMENT": return <AchievementCard card={card} achievements={recap.achievements} />;
    case "MISSED_OPPORTUNITY": return <Missed card={card} />;
    case "NEXT_ACTION": return <NextActionCard card={card} recap={recap} onAction={onAction} />;
    case "SHARE": return <Share card={card} onTrailer={onTrailer} />;
  }
}
