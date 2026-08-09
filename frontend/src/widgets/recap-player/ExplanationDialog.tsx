import { useEffect, useMemo, useRef } from "react";
import type { Achievement, Recap, RecapCard } from "../../entities/recap/model";
import { Button } from "../../shared/ui/Button";
import "./RecapPlayer.css";

function getSheetTitle(card: RecapCard) {
  switch (card.type) {
    case "BEHAVIOR":
      return "Почему именно такой сценарий?";
    case "ACHIEVEMENT":
      return "За что получены ачивки?";
    case "NEXT_ACTION":
      return "Почему именно этот шаг?";
    case "MISSED_OPPORTUNITY":
      return "Почему это полезно сейчас?";
    default:
      return "Почему я это вижу?";
  }
}

export function ExplanationDialog({
  card,
  recap,
  open,
  onClose,
}: {
  card: RecapCard | null;
  recap: Recap;
  open: boolean;
  onClose: () => void;
}) {
  const closeRef = useRef<HTMLButtonElement>(null);
  const previousFocus = useRef<HTMLElement | null>(null);

  useEffect(() => {
    if (!open) return;
    previousFocus.current = document.activeElement as HTMLElement | null;
    closeRef.current?.focus();

    const handler = (event: KeyboardEvent) => {
      if (event.key === "Escape") onClose();
      if (event.key !== "Tab") return;
      const dialog = closeRef.current?.closest<HTMLElement>("[role='dialog']");
      if (!dialog) return;
      const focusables = (Array.from(
        dialog.querySelectorAll<HTMLElement>(
          "button, a[href], input, select, textarea, [tabindex]:not([tabindex='-1'])",
        ),
      ) as HTMLElement[]).filter((element) => !element.hasAttribute("disabled"));
      if (!focusables.length) return;
      const first = focusables[0];
      const last = focusables[focusables.length - 1];
      if (event.shiftKey && document.activeElement === first) {
        event.preventDefault();
        last.focus();
      } else if (!event.shiftKey && document.activeElement === last) {
        event.preventDefault();
        first.focus();
      }
    };

    window.addEventListener("keydown", handler);
    return () => {
      window.removeEventListener("keydown", handler);
      previousFocus.current?.focus();
    };
  }, [onClose, open]);

  const achievementItems = useMemo(() => {
    if (!card || card.type !== "ACHIEVEMENT") return [];
    return card.payload.codes
      .map((code) => recap.achievements.find((item) => item.code === code))
      .filter((item): item is Achievement => Boolean(item));
  }, [card, recap.achievements]);

  if (!open || !card) return null;

  return (
    <div className="dialog-backdrop" role="presentation" onMouseDown={onClose}>
      <section
        className="explanation-dialog"
        role="dialog"
        aria-modal="true"
        aria-labelledby="explanation-title"
        onMouseDown={(event) => event.stopPropagation()}
      >
        <div className="explanation-dialog__header">
          <div>
            <span>Прозрачный recap</span>
            <h2 id="explanation-title">{getSheetTitle(card)}</h2>
          </div>
          <button ref={closeRef} type="button" onClick={onClose} aria-label="Закрыть">
            ×
          </button>
        </div>

        {card.explanation && <p className="explanation-dialog__lead">{card.explanation}</p>}

        {card.type === "BEHAVIOR" && (
          <>
            {card.payload.evidence.length ? (
              <div className="evidence-list">
                {card.payload.evidence.map((item) => (
                  <article key={item.metric}>
                    <div>
                      <strong>{item.label}</strong>
                      <span>+{item.points} баллов</span>
                    </div>
                    <dl>
                      <div>
                        <dt>Факт</dt>
                        <dd>{item.actualValue.toLocaleString("ru-RU")}</dd>
                      </div>
                      <div>
                        <dt>Порог правила</dt>
                        <dd>
                          {item.comparison === "GTE" ? "≥" : "≤"} {item.threshold.toLocaleString("ru-RU")}
                        </dd>
                      </div>
                    </dl>
                    <p>{item.explanation}</p>
                  </article>
                ))}
              </div>
            ) : (
              <p className="evidence-empty">
                Для универсального сценария evidence отсутствует: специализированные правила не сработали.
              </p>
            )}
            <div className="score-row">
              <span>Итоговый score</span>
              <strong>{card.payload.score}</strong>
            </div>
          </>
        )}

        {card.type === "ACHIEVEMENT" && (
          <div className="achievement-reasons">
            {achievementItems.map((item) => (
              <article key={item.code}>
                <strong>{item.title}</strong>
                <p>{item.reason}</p>
              </article>
            ))}
          </div>
        )}

        <Button variant="secondary" fullWidth onClick={onClose}>
          Понятно
        </Button>
      </section>
    </div>
  );
}
