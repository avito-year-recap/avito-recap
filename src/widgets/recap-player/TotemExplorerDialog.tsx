import { motion, useReducedMotion } from "framer-motion";
import { useEffect, useMemo, useRef, useState, type PointerEvent } from "react";
import type { Recap } from "../../entities/recap/model";
import { getTotemExplanation } from "../../shared/lib/experience-utils";
import { playHaptic } from "../../shared/lib/haptics";
import { playUiSound } from "../../shared/lib/sound";
import { YearTotem } from "../../shared/ui/YearTotem";

export function TotemExplorerDialog({
  open,
  recap,
  soundEnabled,
  onClose,
}: {
  open: boolean;
  recap: Recap;
  soundEnabled: boolean;
  onClose: () => void;
}) {
  const reduceMotion = useReducedMotion();
  const closeRef = useRef<HTMLButtonElement>(null);
  const stageRef = useRef<HTMLDivElement>(null);
  const items = useMemo(() => getTotemExplanation(recap), [recap]);
  const [active, setActive] = useState(0);

  useEffect(() => {
    if (!open) return;
    closeRef.current?.focus();
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") onClose();
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [onClose, open]);

  if (!open) return null;

  const select = (index: number) => {
    setActive(index);
    playHaptic("tap");
    if (soundEnabled) playUiSound("metric");
  };

  const onPointerMove = (event: PointerEvent<HTMLDivElement>) => {
    if (reduceMotion || event.pointerType !== "mouse") return;
    const node = stageRef.current;
    if (!node) return;
    const rect = node.getBoundingClientRect();
    const x = (event.clientX - rect.left) / rect.width - 0.5;
    const y = (event.clientY - rect.top) / rect.height - 0.5;
    node.style.setProperty("--totem-explorer-x", `${(x * 8).toFixed(2)}deg`);
    node.style.setProperty("--totem-explorer-y", `${(-y * 7).toFixed(2)}deg`);
  };

  const reset = () => {
    stageRef.current?.style.setProperty("--totem-explorer-x", "0deg");
    stageRef.current?.style.setProperty("--totem-explorer-y", "0deg");
  };

  return (
    <div className="totem-explorer-backdrop" role="presentation" onMouseDown={(event) => event.target === event.currentTarget && onClose()}>
      <motion.section
        className="totem-explorer-dialog"
        role="dialog"
        aria-modal="true"
        aria-labelledby="totem-explorer-title"
        initial={reduceMotion ? false : { opacity: 0, y: 18, scale: .97 }}
        animate={{ opacity: 1, y: 0, scale: 1 }}
      >
        <header>
          <div><span>Твой символ года</span><h2 id="totem-explorer-title">Из чего он сложился?</h2><p>Каждая деталь связана с тем, как прошёл твой год на Авито.</p></div>
          <button ref={closeRef} type="button" onClick={onClose} aria-label="Закрыть разбор символа">×</button>
        </header>
        <div className="totem-explorer-layout">
          <div
            ref={stageRef}
            className={`totem-explorer-stage totem-explorer-stage--part-${active}`}
            onPointerMove={onPointerMove}
            onPointerLeave={reset}
          >
            <YearTotem recap={recap} stage={4} className="totem-explorer-symbol" />
          </div>
          <div className="totem-explorer-parts" role="list" aria-label="Детали символа года">
            {items.map((item, index) => (
              <button key={item.part} type="button" className={active === index ? "is-active" : ""} onClick={() => select(index)} role="listitem">
                <span>{String(index + 1).padStart(2, "0")}</span><div><small>{item.part}</small><strong>{item.value}</strong><p>{item.detail}</p></div>
              </button>
            ))}
          </div>
        </div>
      </motion.section>
    </div>
  );
}
