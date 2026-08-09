import { Link, Navigate, useNavigate, useParams, useSearchParams } from "react-router-dom";
import type { ActionCode } from "../../entities/recap/model";
import { getActionVisual } from "../../shared/lib/visual-registry";
import { Button } from "../../shared/ui/Button";
import { PageShell } from "../../shared/ui/PageShell";
import "./ActionDemoPage.css";

const content: Record<ActionCode, { title: string; description: string; result: string }> = {
  FINISH_DRAFT: {
    title: "Черновик уже ждёт",
    description: "В реальном продукте здесь открылось бы конкретное объявление в редакторе.",
    result: "Можно продолжить с того места, где остановились.",
  },
  OPEN_FAVORITES: {
    title: "Возвращаемся к находкам",
    description: "В реальном продукте здесь открылся бы раздел «Избранное».",
    result: "Все сохранённые варианты снова под рукой.",
  },
  IMPROVE_LISTINGS: {
    title: "Объявление можно усилить",
    description: "В реальном продукте здесь открылось бы активное объявление.",
    result: "Следующий шаг — обновить карточку и вернуть к ней внимание.",
  },
  CONTINUE_DIALOGS: {
    title: "Диалог продолжается",
    description: "В реальном продукте здесь открылся бы нужный чат.",
    result: "Контекст не теряется — пользователь продолжает сценарий сразу после recap.",
  },
  OPEN_TOP_CATEGORY: {
    title: "Главный интерес рядом",
    description: "В реальном продукте здесь открылась бы категория года.",
    result: "Можно продолжить исследование без повторного поиска.",
  },
  CREATE_FIRST_LISTING: {
    title: "Первое объявление начинается здесь",
    description: "В реальном продукте здесь открылся бы сценарий создания первого объявления.",
    result: "Можно перейти от первых шагов продавца к первой публикации.",
  },
  CREATE_LISTING: {
    title: "Готово к новой продаже",
    description: "В реальном продукте здесь открылся бы сценарий создания объявления.",
    result: "Recap превращается в начало следующего полезного действия.",
  },
  SAVE_SEARCH: {
    title: "Поиск сохранён",
    description: "В реальном продукте новые объявления по выбранной категории могли бы приходить автоматически.",
    result: "Теперь не нужно повторять один и тот же поиск вручную.",
  },
  VIEW_SIMILAR_LISTINGS: {
    title: "Нашли похожие варианты",
    description: "В реальном продукте здесь появилась бы подборка по последней покупке или объявлению.",
    result: "Следующий выбор начинается с уже понятного контекста.",
  },
  EXPLORE_RECOMMENDATIONS: {
    title: "Есть идеи, куда дальше",
    description: "В реальном продукте здесь открылась бы персональная лента рекомендаций.",
    result: "Даже универсальный сценарий заканчивается полезным продолжением.",
  },
};

const actionCodes = new Set(Object.keys(content));

export function ActionDemoPage() {
  const { actionCode } = useParams();
  const [params] = useSearchParams();
  const navigate = useNavigate();

  if (!actionCode || !actionCodes.has(actionCode)) return <Navigate to="/404" replace />;

  const code = actionCode as ActionCode;
  const visual = getActionVisual(code);
  const copy = content[code];
  const entries: Array<[string, string]> = [];
  params.forEach((value, key) => entries.push([key, value]));
  const targetLabel = entries.length
    ? entries.map(([key, value]) => `${key}: ${value}`).join(" · ")
    : "Без дополнительных параметров";

  return (
    <PageShell compactHeader actions={<span className="demo-chip">Демо следующего шага</span>}>
      <section className={`action-demo action-demo--${visual.tone}`}>
        <div className="action-demo__copy">
          <span className="action-demo__eyebrow">Recap завершён — сценарий продолжается</span>
          <h1>{copy.title}</h1>
          <p>{copy.description}</p>
          <div className="action-demo__result">
            <span aria-hidden="true">✓</span>
            <strong>{copy.result}</strong>
          </div>
          <details className="action-demo__target">
            <summary>Технический target демо-перехода</summary>
            <code>{targetLabel}</code>
          </details>
          <div className="action-demo__actions">
            <Button onClick={() => navigate(-1)}>Вернуться к итогам</Button>
            <Link to="/" className="action-demo__link">Выбрать другой профиль</Link>
          </div>
        </div>
        <div className="action-demo__visual" aria-hidden="true">
          <div className="action-demo__orb">
            <span>{visual.icon}</span>
            <i>{visual.secondary}</i>
          </div>
          <p>{visual.caption}</p>
        </div>
      </section>
    </PageShell>
  );
}
