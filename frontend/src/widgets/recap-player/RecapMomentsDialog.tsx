import { useEffect, useRef } from "react";
import type { RecapCard } from "../../entities/recap/model";

const labels: Record<RecapCard["type"], string> = {
  INTRO: "Начало",
  YEAR_ACTIVITY: "Год в цифрах",
  TOP_CATEGORY: "Главный интерес",
  ACTIVE_MONTH: "Месяц года",
  BEHAVIOR: "Твой сценарий",
  ACHIEVEMENT: "Ачивки",
  MISSED_OPPORTUNITY: "Возможность",
  NEXT_ACTION: "Следующий шаг",
  SHARE: "Финал",
};

const icons: Record<RecapCard["type"], string> = {
  INTRO: "✦",
  YEAR_ACTIVITY: "#",
  TOP_CATEGORY: "⌕",
  ACTIVE_MONTH: "◷",
  BEHAVIOR: "◎",
  ACHIEVEMENT: "◇",
  MISSED_OPPORTUNITY: "→",
  NEXT_ACTION: "↗",
  SHARE: "♡",
};

export function RecapMomentsDialog({
  open,
  cards,
  activeIndex,
  onSelect,
  onClose,
}: {
  open: boolean;
  cards: RecapCard[];
  activeIndex: number;
  onSelect: (index: number) => void;
  onClose: () => void;
}) {
  const closeRef = useRef<HTMLButtonElement>(null);

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

  return (
    <div className="moments-backdrop" role="presentation" onMouseDown={(event) => event.target === event.currentTarget && onClose()}>
      <section className="moments-dialog" role="dialog" aria-modal="true" aria-labelledby="moments-title">
        <header>
          <div>
            <span>Твоя история</span>
            <h2 id="moments-title">Вернуться к моменту</h2>
          </div>
          <button ref={closeRef} type="button" onClick={onClose} aria-label="Закрыть">×</button>
        </header>
        <div className="moments-grid">
          {cards.map((card, index) => (
            <button
              key={card.id}
              type="button"
              className={index === activeIndex ? "is-active" : ""}
              onClick={() => {
                onSelect(index);
                onClose();
              }}
            >
              <i aria-hidden="true">{icons[card.type]}</i>
              <span>{String(index + 1).padStart(2, "0")}</span>
              <strong>{labels[card.type]}</strong>
            </button>
          ))}
        </div>
      </section>
    </div>
  );
}
