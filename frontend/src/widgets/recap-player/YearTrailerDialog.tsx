import { AnimatePresence, motion, useReducedMotion } from "framer-motion";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import type { PublicSharePayload } from "../../entities/recap/model";
import { PublicYearTotem } from "../../shared/ui/YearTotem";
import { playUiSound } from "../../shared/lib/sound";
import "./YearTrailerDialog.css";

interface TrailerScene {
  kicker: string;
  title: string;
  detail?: string;
  glyph: string;
  kind: "totem" | "behavior" | "achievement" | "interest" | "final";
}

function buildScenes(payload: PublicSharePayload): TrailerScene[] {
  const scenes: TrailerScene[] = [
    { kicker: `Avito · ${payload.year}`, title: "Один год. Один визуальный код.", detail: "Собран из безопасной публичной выжимки.", glyph: "✦", kind: "totem" },
    { kicker: "Сценарий года", title: payload.behaviorTitle, detail: "То, что лучше всего описывает твой год на площадке.", glyph: "◎", kind: "behavior" },
  ];
  if (payload.achievementTitle) scenes.push({ kicker: "Ачивка года", title: payload.achievementTitle, detail: "Один публичный знак из личной коллекции.", glyph: "◇", kind: "achievement" });
  if (payload.topCategory) scenes.push({ kicker: "Главный интерес", title: payload.topCategory, detail: "Категория показана только потому, что она разрешена для SHARE.", glyph: "⌕", kind: "interest" });
  scenes.push({ kicker: "Финал", title: "#МойГодНаАвито", detail: "Без имени, числовой статистики и внутренних идентификаторов.", glyph: "↗", kind: "final" });
  return scenes;
}

export function YearTrailerDialog({
  open,
  payload,
  soundEnabled,
  onClose,
}: {
  open: boolean;
  payload: PublicSharePayload | null;
  soundEnabled: boolean;
  onClose: () => void;
}) {
  const reduceMotion = useReducedMotion();
  const closeRef = useRef<HTMLButtonElement>(null);
  const scenes = useMemo(() => (payload ? buildScenes(payload) : []), [payload]);
  const [scene, setScene] = useState(0);
  const [playing, setPlaying] = useState(true);

  useEffect(() => {
    if (!open) return;
    closeRef.current?.focus();
  }, [open]);

  const handleClose = useCallback(() => {
    setScene(0);
    setPlaying(true);
    onClose();
  }, [onClose]);

  useEffect(() => {
    if (!open) return;
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") handleClose();
      if (event.key === " ") {
        event.preventDefault();
        setPlaying((value) => !value);
      }
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [handleClose, open]);

  useEffect(() => {
    if (!open || !playing || scenes.length === 0) return;
    if (scene >= scenes.length - 1) return;
    const timeout = window.setTimeout(() => {
      setScene((value) => value + 1);
      if (soundEnabled) playUiSound(scene === scenes.length - 2 ? "final" : "navigate");
    }, reduceMotion ? 1400 : 2300);
    return () => window.clearTimeout(timeout);
  }, [open, playing, reduceMotion, scene, scenes.length, soundEnabled]);

  if (!open || !payload || scenes.length === 0) return null;
  const current = scenes[scene];

  return (
    <div className="trailer-backdrop" role="presentation" onMouseDown={(event) => event.target === event.currentTarget && handleClose()}>
      <section className="trailer-dialog" role="dialog" aria-modal="true" aria-labelledby="trailer-title">
        <header className="trailer-dialog__topline">
          <div className="trailer-progress" aria-label={`Сцена ${scene + 1} из ${scenes.length}`}>
            {scenes.map((item, index) => (
              <span key={`${item.kind}-${index}`} className={index <= scene ? "is-complete" : ""} />
            ))}
          </div>
          <button ref={closeRef} type="button" onClick={handleClose} aria-label="Закрыть трейлер">×</button>
        </header>

        <div className={`trailer-stage trailer-stage--${current.kind}`}>
          <AnimatePresence mode="wait">
            <motion.div
              key={`${current.kind}-${scene}`}
              className="trailer-scene"
              initial={reduceMotion ? false : { opacity: 0, scale: 0.9, y: 18 }}
              animate={{ opacity: 1, scale: 1, y: 0 }}
              exit={reduceMotion ? { opacity: 0 } : { opacity: 0, scale: 1.05, y: -12 }}
              transition={{ duration: reduceMotion ? 0 : 0.45, ease: [0.22, 1, 0.36, 1] }}
            >
              {current.kind === "totem" ? (
                <div className="trailer-scene__totem"><PublicYearTotem payload={payload} /></div>
              ) : (
                <motion.div className="trailer-scene__glyph" aria-hidden="true" animate={reduceMotion ? undefined : { rotate: [0, -4, 4, 0], scale: [1, 1.04, 1] }} transition={{ duration: 2.2, repeat: Infinity }}>{current.glyph}</motion.div>
              )}
              <p>{current.kicker}</p>
              <h2 id="trailer-title">{current.title}</h2>
              {current.detail && <span>{current.detail}</span>}
            </motion.div>
          </AnimatePresence>
        </div>

        <footer className="trailer-controls">
          <button type="button" onClick={() => setScene((value) => Math.max(0, value - 1))} disabled={scene === 0}>←</button>
          <button type="button" onClick={() => setPlaying((value) => !value)}>{playing && scene < scenes.length - 1 ? "Пауза" : "Продолжить"}</button>
          {scene < scenes.length - 1 ? (
            <button type="button" onClick={() => setScene((value) => Math.min(scenes.length - 1, value + 1))}>→</button>
          ) : (
            <button type="button" onClick={() => { setScene(0); setPlaying(true); }}>↻</button>
          )}
        </footer>
      </section>
    </div>
  );
}
