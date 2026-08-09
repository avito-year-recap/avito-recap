import { Button } from "./Button";
import "./shared-ui.css";

interface ErrorStateProps {
  title: string;
  description: string;
  onRetry?: () => void;
}

export function ErrorState({ title, description, onRetry }: ErrorStateProps) {
  return (
    <section className="async-state" role="alert">
      <div className="async-state__icon" aria-hidden="true">
        !
      </div>
      <h1>{title}</h1>
      <p>{description}</p>
      {onRetry && <Button onClick={onRetry}>Попробовать ещё раз</Button>}
    </section>
  );
}

export function PageLoader({ label = "Загружаем данные" }: { label?: string }) {
  return (
    <div className="page-loader" role="status" aria-live="polite">
      <span className="page-loader__ring" aria-hidden="true" />
      <span>{label}</span>
    </div>
  );
}
