import { useNavigate } from "react-router-dom";
import { Button } from "../../shared/ui/Button";
import { PageShell } from "../../shared/ui/PageShell";

export function NotFoundPage() {
  const navigate = useNavigate();
  return (
    <PageShell compactHeader>
      <section className="async-state">
        <div className="async-state__icon">404</div>
        <h1>Такой страницы нет</h1>
        <p>Вернитесь на главную или откройте итоги года.</p>
        <Button onClick={() => navigate("/")}>На главную</Button>
      </section>
    </PageShell>
  );
}
