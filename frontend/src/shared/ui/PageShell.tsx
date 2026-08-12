import type { PropsWithChildren, ReactNode } from "react";
import { Link } from "react-router-dom";
import { AppLogo } from "./AppLogo";
import "./shared-ui.css";

interface PageShellProps extends PropsWithChildren {
  actions?: ReactNode;
  compactHeader?: boolean;
  fitViewport?: boolean;
}

export function PageShell({
  children,
  actions,
  compactHeader = false,
  fitViewport = false,
}: PageShellProps) {
  return (
    <div className={`page-shell${fitViewport ? " page-shell--fit-viewport" : ""}`}>
      <header className="page-shell__header">
        <Link to="/" className="page-shell__brand" aria-label="На главную">
          <AppLogo compact={compactHeader} />
        </Link>
        <div className="page-shell__actions">{actions}</div>
      </header>
      <main className="page-shell__main">{children}</main>
    </div>
  );
}
