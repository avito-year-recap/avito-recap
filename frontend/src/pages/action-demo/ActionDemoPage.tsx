import { Link, Navigate, useNavigate, useParams } from "react-router-dom";
import type { ActionCode } from "../../entities/recap/model";
import { getActionVisual } from "../../shared/lib/visual-registry";
import { Button } from "../../shared/ui/Button";
import { PageShell } from "../../shared/ui/PageShell";
import "./ActionDemoPage.css";

const content: Record<ActionCode, { title: string; description: string }> = {
  FINISH_DRAFT: { title: "Черновик уже ждёт", description: "Продолжи редактирование с того места, где остановился." },
  OPEN_FAVORITES: { title: "Возвращаемся к находкам", description: "Все сохранённые объявления снова под рукой." },
  IMPROVE_LISTINGS: { title: "Объявление можно обновить", description: "Проверь детали и сделай объявление заметнее." },
  CONTINUE_DIALOGS: { title: "Диалог продолжается", description: "Вернись к разговору с того же места." },
  OPEN_TOP_CATEGORY: { title: "Главный интерес рядом", description: "Посмотри новые объявления в категории, к которой возвращался чаще всего." },
  CREATE_FIRST_LISTING: { title: "Первое объявление", description: "Самое время опубликовать вещь, которую давно хотел продать." },
  CREATE_LISTING: { title: "Новое объявление", description: "Продолжи год с новой публикации." },
  SAVE_SEARCH: { title: "Поиск сохранён", description: "Новые подходящие варианты будет проще не пропустить." },
  VIEW_SIMILAR_LISTINGS: { title: "Похожие варианты", description: "Посмотри объявления, похожие на твои находки года." },
  EXPLORE_RECOMMENDATIONS: { title: "Что посмотреть дальше", description: "Продолжи с подборкой новых идей и находок." },
};

const actionCodes = new Set(Object.keys(content));

export function ActionDemoPage() {
  const { actionCode } = useParams();
  const navigate = useNavigate();
  if (!actionCode || !actionCodes.has(actionCode)) return <Navigate to="/404" replace />;

  const code = actionCode as ActionCode;
  const visual = getActionVisual(code);
  const copy = content[code];

  return (
    <PageShell compactHeader fitViewport backTo="/account" backLabel="В кабинет">
      <section className={`action-demo action-demo--${visual.tone}`}>
        <div className="action-demo__copy">
          <span className="action-demo__eyebrow">Следующий шаг</span>
          <h1>{copy.title}</h1>
          <p>{copy.description}</p>
          <div className="action-demo__actions">
            <Button onClick={() => navigate(-1)}>Вернуться к итогам</Button>
            <Link to="/profiles?return=/account" className="action-demo__link">Сменить профиль</Link>
          </div>
        </div>
        <div className="action-demo__visual" aria-hidden="true">
          <div className="action-demo__orb"><span>{visual.icon}</span><i>{visual.secondary}</i></div>
        </div>
      </section>
    </PageShell>
  );
}
