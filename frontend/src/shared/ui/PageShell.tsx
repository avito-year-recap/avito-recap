import type { PropsWithChildren, ReactNode } from "react";
import { Link } from "react-router-dom";
import { AppLogo } from "./AppLogo";
import "./shared-ui.css";

interface PageShellProps extends PropsWithChildren {
  actions?: ReactNode;
  compactHeader?: boolean;
  fitViewport?: boolean;
  backTo?: string;
  backLabel?: string;
  narrow?: boolean;
}

export function PageShell({
  children,
  actions,
  compactHeader = false,
  fitViewport = false,
  backTo,
  backLabel = "Назад",
  narrow = false,
}: PageShellProps) {
  return (
    <div className={`page-shell${fitViewport ? " page-shell--fit-viewport" : ""}${narrow ? " page-shell--narrow" : ""}`}>
      <header className="page-shell__header">
        <div className="page-shell__leading">
          <Link to="/" className="page-shell__brand" aria-label="На главную">
            <AppLogo compact={compactHeader} />
          </Link>
          {backTo && (
            <Link to={backTo} className="page-shell__back">
              <i className="hgi hgi-stroke hgi-arrow-left-01" aria-hidden="true" />
              <span>{backLabel}</span>
            </Link>
          )}
        </div>
        <div className="page-shell__actions">{actions}</div>
      </header>
      <main className="page-shell__main">{children}</main>
    </div>
  );
}
