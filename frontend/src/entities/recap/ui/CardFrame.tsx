import { useReducedMotion } from "framer-motion";
import { useRef, type PointerEvent, type PropsWithChildren, type ReactNode } from "react";
import type { VisualTone } from "../../../shared/lib/visual-registry";
import "./recap-cards.css";

interface Props extends PropsWithChildren {
  eyebrow: string;
  title: string;
  description: string;
  tone?: VisualTone;
  footer?: ReactNode;
  className?: string;
}

export function CardFrame({
  eyebrow,
  title,
  description,
  tone = "neutral",
  children,
  footer,
  className = "",
}: Props) {
  const cardRef = useRef<HTMLElement>(null);
  const reduceMotion = useReducedMotion();

  const onPointerMove = (event: PointerEvent<HTMLElement>) => {
    if (reduceMotion || event.pointerType !== "mouse") return;
    const node = cardRef.current;
    if (!node || !window.matchMedia("(hover: hover)").matches) return;
    const rect = node.getBoundingClientRect();
    const x = (event.clientX - rect.left) / rect.width - 0.5;
    const y = (event.clientY - rect.top) / rect.height - 0.5;
    node.style.setProperty("--card-tilt-x", `${(-y * 2.2).toFixed(2)}deg`);
    node.style.setProperty("--card-tilt-y", `${(x * 2.6).toFixed(2)}deg`);
    node.style.setProperty("--card-light-x", `${((x + 0.5) * 100).toFixed(1)}%`);
    node.style.setProperty("--card-light-y", `${((y + 0.5) * 100).toFixed(1)}%`);
  };

  const resetTilt = () => {
    const node = cardRef.current;
    if (!node) return;
    node.style.setProperty("--card-tilt-x", "0deg");
    node.style.setProperty("--card-tilt-y", "0deg");
    node.style.setProperty("--card-light-x", "50%");
    node.style.setProperty("--card-light-y", "50%");
  };

  return (
    <article
      ref={cardRef}
      className={`recap-card recap-card--${tone} ${className}`}
      onPointerMove={onPointerMove}
      onPointerLeave={resetTilt}
    >
      <div className="recap-card__cursor-light" aria-hidden="true" />
      <header className="recap-card__header">
        <span>{eyebrow}</span>
        <h1>{title}</h1>
        <p>{description}</p>
      </header>
      <div className="recap-card__content">{children}</div>
      {footer && <footer className="recap-card__footer">{footer}</footer>}
    </article>
  );
}
