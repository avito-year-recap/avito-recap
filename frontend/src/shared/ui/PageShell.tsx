import type { PropsWithChildren, ReactNode } from "react";
import { Link } from "react-router-dom";
import { AppLogo } from "./AppLogo";
import "./shared-ui.css";

interface PageShellProps extends PropsWithChildren {
  actions?: ReactNode;
  compactHeader?: boolean;
  fitViewport?: boolean;
  backToProfiles?: boolean;
}

export function PageShell({
  children,
  actions,
  compactHeader = false,
  fitViewport = false,
  backToProfiles = false,
}: PageShellProps) {
  return (
    <div className={`page-shell${fitViewport ? " page-shell--fit-viewport" : ""}`}>
      <header className="page-shell__header">
        <div className="page-shell__leading">
          <Link to="/" className="page-shell__brand" aria-label="На главную">
            <AppLogo compact={compactHeader} />
          </Link>
          {backToProfiles && (
            <Link to="/" className="page-shell__back">
              <span aria-hidden="true">←</span>
              <span>К профилям</span>
            </Link>
          )}
        </div>
        <div className="page-shell__actions">{actions}</div>
      </header>
      <main className="page-shell__main">{children}</main>
    </div>
  );
}
