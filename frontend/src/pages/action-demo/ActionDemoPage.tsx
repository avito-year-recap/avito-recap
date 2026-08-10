import { Link, Navigate, useNavigate, useParams } from "react-router-dom";
import type { ActionCode } from "../../entities/recap/model";
import { getActionVisual } from "../../shared/lib/visual-registry";
import { Button } from "../../shared/ui/Button";
import { PageShell } from "../../shared/ui/PageShell";
import "./ActionDemoPage.css";

const content: Record<ActionCode, { title: string; description: string }> = {
  FINISH_DRAFT: { title: "Черновик уже ждёт", description: "Здесь пользователь продолжил бы редактировать своё объявление." },
  OPEN_FAVORITES: { title: "Возвращаемся к находкам", description: "Здесь открылся бы раздел с сохранёнными объявлениями." },
  IMPROVE_LISTINGS: { title: "Объявление можно обновить", description: "Здесь открылось бы активное объявление пользователя." },
  CONTINUE_DIALOGS: { title: "Диалог продолжается", description: "Здесь пользователь вернулся бы в нужный чат." },
  OPEN_TOP_CATEGORY: { title: "Главный интерес рядом", description: "Здесь открылась бы категория, которая чаще всего интересовала пользователя." },
  CREATE_FIRST_LISTING: { title: "Первое объявление", description: "Здесь начался бы сценарий создания объявления." },
  CREATE_LISTING: { title: "Новое объявление", description: "Здесь начался бы сценарий создания нового объявления." },
  SAVE_SEARCH: { title: "Поиск сохранён", description: "Здесь пользователь сохранил бы поиск и мог получать новые варианты автоматически." },
  VIEW_SIMILAR_LISTINGS: { title: "Похожие варианты", description: "Здесь открылась бы подборка похожих объявлений." },
  EXPLORE_RECOMMENDATIONS: { title: "Что посмотреть дальше", description: "Здесь открылась бы подборка рекомендаций." },
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
    <PageShell compactHeader fitViewport backToProfiles>
      <section className={`action-demo action-demo--${visual.tone}`}>
        <div className="action-demo__copy">
          <span className="action-demo__eyebrow">Следующий шаг</span>
          <h1>{copy.title}</h1>
          <p>{copy.description}</p>
          <div className="action-demo__actions">
            <Button onClick={() => navigate(-1)}>Вернуться к итогам</Button>
            <Link to="/" className="action-demo__link">Выбрать другой профиль</Link>
          </div>
        </div>
        <div className="action-demo__visual" aria-hidden="true">
          <div className="action-demo__orb"><span>{visual.icon}</span><i>{visual.secondary}</i></div>
        </div>
      </section>
    </PageShell>
  );
}
