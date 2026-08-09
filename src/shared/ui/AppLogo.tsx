import "./shared-ui.css";

interface AppLogoProps {
  compact?: boolean;
}

export function AppLogo({ compact = false }: AppLogoProps) {
  return (
    <div className="app-logo" aria-label="Avito Итоги года">
      <span className="app-logo__mark" aria-hidden="true">
        <i className="app-logo__dot app-logo__dot--blue" />
        <i className="app-logo__dot app-logo__dot--purple" />
        <i className="app-logo__dot app-logo__dot--green" />
        <i className="app-logo__dot app-logo__dot--red" />
      </span>
      <strong>Avito</strong>
      {!compact && <span className="app-logo__product">Итоги</span>}
    </div>
  );
}
