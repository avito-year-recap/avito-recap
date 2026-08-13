import { AnimatePresence, motion, useReducedMotion } from "framer-motion";
import { useCallback, useEffect, useMemo, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import type { Recap, RecapCard } from "../../entities/recap/model";
import { RecapCardRenderer } from "../../entities/recap/ui/RecapCardRenderer";
import { buildMockActionUrl } from "../../features/next-action/executeMockAction";
import {
  getPublicPayload,
  readStoredProgress,
  resetStoredProgress,
  writeStoredProgress,
} from "../../shared/lib/experience-utils";
import { getActionVisual, getBehaviorVisual, getCategoryVisual } from "../../shared/lib/visual-registry";
import { playUiSound, setUiSoundProfile } from "../../shared/lib/sound";
import { playHaptic } from "../../shared/lib/haptics";
import { getRecapTotemStage } from "../../shared/ui/year-totem-utils";
import { RecapMomentsDialog } from "./RecapMomentsDialog";
import { YearTrailerDialog } from "./YearTrailerDialog";
import "./RecapPlayer.css";

function clampSlide(raw: string | null, max: number) {
  if (!raw) return 0;
  const parsed = Number(raw);
  if (!Number.isInteger(parsed)) return 0;
  return Math.max(0, Math.min(max, parsed - 1));
}

function transitionForCard(card: RecapCard, direction: number) {
  switch (card.type) {
    case "YEAR_ACTIVITY": return { opacity: 0, y: direction > 0 ? 30 : -20, scale: 0.985 };
    case "TOP_CATEGORY": return { opacity: 0, x: direction > 0 ? 44 : -44, rotate: direction > 0 ? 1.5 : -1.5, scale: 0.985 };
    case "ACTIVE_MONTH": return { opacity: 0, y: 24, scale: 0.965 };
    case "BEHAVIOR": return { opacity: 0, scale: 0.9, filter: "blur(6px)" };
    case "ACHIEVEMENT": return { opacity: 0, y: 28, rotate: direction > 0 ? -1.5 : 1.5, scale: 0.96 };
    case "MISSED_OPPORTUNITY": return { opacity: 0, x: direction > 0 ? 38 : -38, scale: 0.98 };
    case "NEXT_ACTION": return { opacity: 0, y: 24, scale: 0.94 };
    case "SHARE": return { opacity: 0, scale: 0.88, filter: "blur(8px)" };
    case "INTRO":
    default: return { opacity: 0, x: direction > 0 ? 30 : -30, scale: 0.98 };
  }
}

function transitionForPair(from: RecapCard | undefined, to: RecapCard, direction: number) {
  if (!from || direction < 0) return null;
  if (from.type === "TOP_CATEGORY" && to.type === "ACTIVE_MONTH") {
    return { opacity: 0, scale: 0.9, rotate: 5, y: 18, filter: "blur(4px)" };
  }
  if (from.type === "BEHAVIOR" && to.type === "ACHIEVEMENT") {
    return { opacity: 0, scale: 1.08, y: -22, filter: "blur(3px)" };
  }
  if (from.type === "ACHIEVEMENT" && to.type === "MISSED_OPPORTUNITY") {
    return { opacity: 0, x: 46, rotate: -2, scale: .97 };
  }
  if ((from.type === "ACHIEVEMENT" || from.type === "MISSED_OPPORTUNITY") && to.type === "NEXT_ACTION") {
    return { opacity: 0, y: 46, scale: .88, filter: "blur(3px)" };
  }
  return null;
}

function cardDuration(card: RecapCard) {
  switch (card.type) {
    case "YEAR_ACTIVITY": return 4700;
    case "BEHAVIOR": return 5200;
    case "ACHIEVEMENT": return 5000;
    case "MISSED_OPPORTUNITY": return 4700;
    case "NEXT_ACTION": return 5200;
    default: return 3600;
  }
}

function atmosphereFor(card: RecapCard) {
  if (card.type === "TOP_CATEGORY") return getCategoryVisual(card.payload.categoryCode);
  if (card.type === "BEHAVIOR") return getBehaviorVisual(card.payload.code);
  if (card.type === "NEXT_ACTION" || card.type === "MISSED_OPPORTUNITY") return getActionVisual(card.payload.code);
  if (card.type === "ACHIEVEMENT") return { tone: "purple", motif: "constellation" } as const;
  if (card.type === "ACTIVE_MONTH") return { tone: "coral", motif: "pulse" } as const;
  if (card.type === "SHARE") return { tone: "green", motif: "orbit" } as const;
  return { tone: "blue", motif: "pulse" } as const;
}

function isSoundPreferenceEnabled() {
  if (typeof window === "undefined") return false;
  try { return window.localStorage.getItem("avito-recap-sound") === "on"; } catch { return false; }
}

export function RecapPlayer({ recap }: { recap: Recap }) {
  const cards = recap.cards;
  const navigate = useNavigate();
  const reduceMotion = useReducedMotion();
  const [searchParams, setSearchParams] = useSearchParams();
  const activeIndex = clampSlide(searchParams.get("slide"), cards.length - 1);
  const [direction, setDirection] = useState(1);
  const [momentsOpen, setMomentsOpen] = useState(false);
  const [trailerOpen, setTrailerOpen] = useState(false);
  const [cinematic, setCinematic] = useState(false);
  const [soundEnabled, setSoundEnabled] = useState(isSoundPreferenceEnabled);
  const storedAtStart = useMemo(() => readStoredProgress(recap.id), [recap.id]);
  const [resumeIndex, setResumeIndex] = useState(() => {
    if (searchParams.get("slide")) return null;
    if (!storedAtStart || storedAtStart.completed || storedAtStart.index <= 0 || storedAtStart.index >= cards.length - 1) return null;
    return storedAtStart.index;
  });
  const [wasCompleted, setWasCompleted] = useState(activeIndex === cards.length - 1 || Boolean(storedAtStart?.completed));
  const activeCard = cards[activeIndex];
  const publicPayload = useMemo(() => getPublicPayload(recap), [recap]);
  const atmosphere = atmosphereFor(activeCard);
  const behaviorCard = cards.find((card) => card.type === "BEHAVIOR");
  const behaviorCode = behaviorCard?.type === "BEHAVIOR" ? behaviorCard.payload.code : "UNIVERSAL_USER";

  useEffect(() => {
    setUiSoundProfile(behaviorCode);
  }, [behaviorCode]);

  const setSlide = useCallback((index: number, nextDirection: number, source: "manual" | "auto" = "manual") => {
    const clamped = Math.max(0, Math.min(cards.length - 1, index));
    if (clamped === activeIndex) return;
    if (clamped === cards.length - 1) {
      setWasCompleted(true);
      if (source === "auto") setCinematic(false);
    }
    setDirection(nextDirection);
    setSearchParams((current) => {
      const next = new URLSearchParams(current);
      next.set("slide", String(clamped + 1));
      return next;
    }, { replace: true });
    const target = cards[clamped];
    if (soundEnabled) {
      if (target.type === "ACHIEVEMENT") playUiSound("achievement");
      else if (target.type === "SHARE") playUiSound("final");
      else if (target.type === "YEAR_ACTIVITY") playUiSound("metric");
      else if (target.type !== "BEHAVIOR") playUiSound(nextDirection > 0 ? "navigate" : "previous");
    }
    if (source === "manual") {
      if (target.type === "ACHIEVEMENT") playHaptic("achievement");
      else if (target.type === "SHARE") playHaptic("final");
      else playHaptic("tap");
    }
    if (source === "manual") setCinematic(false);
  }, [activeIndex, cards, setSearchParams, soundEnabled]);

  const previous = useCallback(() => setSlide(activeIndex - 1, -1), [activeIndex, setSlide]);
  const next = useCallback(() => setSlide(activeIndex + 1, 1), [activeIndex, setSlide]);

  useEffect(() => {
    writeStoredProgress(recap.id, activeIndex, activeIndex === cards.length - 1);
  }, [activeIndex, cards.length, recap.id]);

  useEffect(() => {
    if (!cinematic || momentsOpen || trailerOpen) return;
    if (activeIndex >= cards.length - 1) return;
    const timeout = window.setTimeout(() => setSlide(activeIndex + 1, 1, "auto"), reduceMotion ? 1800 : cardDuration(activeCard));
    return () => window.clearTimeout(timeout);
  }, [activeCard, activeIndex, cards.length, cinematic, momentsOpen, reduceMotion, setSlide, trailerOpen]);

  useEffect(() => {
    if (!soundEnabled || activeCard.type !== "BEHAVIOR") return;
    const timeout = window.setTimeout(() => playUiSound("behaviorReveal"), reduceMotion ? 0 : 620);
    return () => window.clearTimeout(timeout);
  }, [activeCard.id, activeCard.type, reduceMotion, soundEnabled]);

  useEffect(() => {
    const handler = (event: KeyboardEvent) => {
      if (momentsOpen || trailerOpen) return;
      if (event.key === "ArrowLeft") previous();
      if (event.key === "ArrowRight") next();
      if (event.key.toLowerCase() === "p") setCinematic((value) => !value);
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [momentsOpen, next, previous, trailerOpen]);

  const totemStage = getRecapTotemStage(activeIndex, cards.length);
  const runAction = () => {
    if (soundEnabled) playUiSound("cta");
    playHaptic("cta");
    navigate(buildMockActionUrl(recap.nextAction));
  };
  const adjacentPrevious = cards[activeIndex - direction];
  const pairTransition = transitionForPair(adjacentPrevious, activeCard, direction);
  const initialState = reduceMotion ? false : pairTransition ?? transitionForCard(activeCard, direction);
  const exitState = reduceMotion ? { opacity: 0 } : { ...transitionForCard(activeCard, -direction), filter: "blur(2px)" };

  const toggleSound = () => {
    const nextValue = !soundEnabled;
    setSoundEnabled(nextValue);
    try { window.localStorage.setItem("avito-recap-sound", nextValue ? "on" : "off"); } catch { /* optional preference */ }
    if (nextValue) playUiSound("enable");
  };

  const resume = () => {
    if (resumeIndex === null) return;
    setSlide(resumeIndex, 1);
    setResumeIndex(null);
  };

  const restart = () => {
    resetStoredProgress(recap.id);
    setResumeIndex(null);
    setWasCompleted(false);
    if (activeIndex !== 0) setSlide(0, -1);
  };

  return (
    <div className={`recap-player-layout recap-player-layout--${atmosphere.tone} recap-player-layout--${atmosphere.motif ?? "pulse"} recap-player-layout--stage-${totemStage}`}>
      <section className="recap-stage" aria-label="Персональные итоги года">
        <div className="recap-stage__topline">
          <div className="recap-progress" aria-label={`Экран ${activeIndex + 1} из ${cards.length}`}>
            {cards.map((card, index) => (
              <button
                key={card.id}
                type="button"
                className={`${index < activeIndex ? "is-past" : ""} ${index === activeIndex ? "is-active" : ""} ${cinematic && index === activeIndex ? "is-cinematic" : ""}`}
                onClick={() => setSlide(index, index > activeIndex ? 1 : -1)}
                aria-label={`Экран ${index + 1}: ${card.title}`}
                aria-current={index === activeIndex ? "step" : undefined}
              >
                <span style={cinematic && index === activeIndex ? { animationDuration: `${reduceMotion ? 1.8 : cardDuration(activeCard) / 1000}s` } : undefined} />
              </button>
            ))}
          </div>
          <div className="recap-topline__meta">
            <div className="recap-player-tools" aria-label="Режим просмотра">
              <button type="button" className={cinematic ? "is-active" : ""} onClick={() => setCinematic((value) => !value)} aria-label={cinematic ? "Поставить автоисторию на паузу" : "Запустить автоисторию"} title="Автоистория">
                {cinematic ? "Пауза" : "Авто"}
              </button>
              <button type="button" className={soundEnabled ? "is-active" : ""} onClick={toggleSound} aria-label={soundEnabled ? "Выключить звуки" : "Включить звуки"} title="Звуки">
                {soundEnabled ? "Звук вкл" : "Звук"}
              </button>
              {wasCompleted && publicPayload && (
                <button type="button" onClick={() => { setCinematic(false); setTrailerOpen(true); }} aria-label="Открыть трейлер года" title="Трейлер года">Трейлер</button>
              )}
            </div>
            <div className="recap-counter" aria-hidden="true"><b>{String(activeIndex + 1).padStart(2, "0")}</b><span>/</span><span>{String(cards.length).padStart(2, "0")}</span></div>
          </div>
        </div>

        <p className="sr-only" aria-live="polite">Экран {activeIndex + 1} из {cards.length}. {activeCard.title}</p>

        <div className="recap-story-shell">
          <div className="recap-atmosphere" aria-hidden="true"><i /><i /><i /><i /><span /></div>
          <button type="button" className="story-tap-zone story-tap-zone--previous" onClick={previous} disabled={activeIndex === 0} aria-label="Предыдущий экран" />
          <button type="button" className="story-tap-zone story-tap-zone--next" onClick={next} disabled={activeIndex === cards.length - 1} aria-label="Следующий экран" />

          <AnimatePresence mode="wait" initial={false} custom={direction}>
            <motion.div
              className={`recap-motion ${pairTransition ? `recap-motion--pair-${adjacentPrevious?.type.toLowerCase()}-${activeCard.type.toLowerCase()}` : ""}`}
              key={activeCard.id}
              custom={direction}
              initial={initialState}
              animate={{ opacity: 1, x: 0, y: 0, rotate: 0, scale: 1, filter: "blur(0px)" }}
              exit={exitState}
              transition={{ duration: reduceMotion ? 0 : activeCard.type === "BEHAVIOR" || activeCard.type === "SHARE" ? 0.42 : 0.3, ease: [0.22, 1, 0.36, 1] }}
              drag={reduceMotion ? false : "x"}
              dragConstraints={{ left: 0, right: 0 }}
              dragElastic={0.12}
              onDragEnd={(_, info) => { if (info.offset.x < -70) next(); if (info.offset.x > 70) previous(); }}
            >
              <RecapCardRenderer
                card={activeCard}
                recap={recap}
                onAction={runAction}
                onTrailer={() => { setCinematic(false); setTrailerOpen(true); }}
              />
            </motion.div>
          </AnimatePresence>

          {resumeIndex !== null && (
            <motion.div className="resume-card" initial={{ opacity: 0, y: 14 }} animate={{ opacity: 1, y: 0 }}>
              <span>Продолжить историю?</span>
              <strong>Ты остановился на экране {resumeIndex + 1} из {cards.length}</strong>
              <div><button type="button" onClick={restart}>С начала</button><button type="button" onClick={resume}>Продолжить</button></div>
            </motion.div>
          )}
        </div>

        <div className={`recap-controls${wasCompleted ? " recap-controls--with-moments" : ""}`} aria-label="Навигация по истории">
          <button type="button" onClick={previous} disabled={activeIndex === 0}><i className="hgi hgi-stroke hgi-arrow-left-01" aria-hidden="true" /> Назад</button>
          {wasCompleted && (
            <button className="recap-moments-button" type="button" onClick={() => { setCinematic(false); setMomentsOpen(true); }}>Моменты года</button>
          )}
          <button type="button" onClick={next} disabled={activeIndex === cards.length - 1}>Далее <i className="hgi hgi-stroke hgi-arrow-right-01" aria-hidden="true" /></button>
        </div>
     </section>

      <RecapMomentsDialog open={momentsOpen} cards={cards} activeIndex={activeIndex} onSelect={(index) => setSlide(index, index > activeIndex ? 1 : -1)} onClose={() => setMomentsOpen(false)} />
      <YearTrailerDialog open={trailerOpen} payload={publicPayload} soundEnabled={soundEnabled} onClose={() => setTrailerOpen(false)} />
    </div>
  );
}
