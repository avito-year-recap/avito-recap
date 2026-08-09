import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { Navigate, useNavigate, useParams } from "react-router-dom";
import type { PublicSharePayload } from "../../entities/recap/model";
import { getPublicShare } from "../../shared/api/recap-api";
import { ErrorState, PageLoader } from "../../shared/ui/AsyncState";
import { Button } from "../../shared/ui/Button";
import { PageShell } from "../../shared/ui/PageShell";
import { PublicYearTotem } from "../../shared/ui/YearTotem";
import { YearTrailerDialog } from "../../widgets/recap-player/YearTrailerDialog";
import { hasCompletionBonus } from "../../shared/lib/experience-utils";
import "./SharePage.css";

type ShareTemplate = "symbol" | "minimal" | "interest" | "bonus";
type ShareFormat = "portrait" | "square";
type ShareTone = "soft" | "color";

const baseTemplateLabels: Array<{ id: ShareTemplate; label: string; hint: string }> = [
  { id: "symbol", label: "Символ года", hint: "Главный визуальный образ" },
  { id: "minimal", label: "Минимализм", hint: "Типографика и ачивка" },
  { id: "interest", label: "Главный интерес", hint: "Категория в центре" },
];
const bonusTemplate = { id: "bonus" as const, label: "Контур года ✦", hint: "Скрытый стиль за полный просмотр" };

function SharePreview({
  payload,
  template,
  format,
  tone,
  showAchievement,
  showCategory,
}: {
  payload: PublicSharePayload;
  template: ShareTemplate;
  format: ShareFormat;
  tone: ShareTone;
  showAchievement: boolean;
  showCategory: boolean;
}) {
  const category = showCategory ? payload.topCategory : undefined;
  const achievement = showAchievement ? payload.achievementTitle : undefined;
  return (
    <article className={`share-preview share-preview--${template} share-preview--${format} share-preview--tone-${tone}`} aria-label={`Предпросмотр публичной карточки: ${template}`}>
      <div className="share-preview__topline"><span>Avito · Итоги {payload.year}</span><i aria-hidden="true">✦</i></div>

      {(template === "symbol" || template === "bonus") && (
        <>
          <div className="share-preview__totem">
            {template === "bonus" && <div className="share-preview__bonus-rings" aria-hidden="true"><i /><i /><i /></div>}
            <PublicYearTotem payload={{ ...payload, achievementTitle: achievement, topCategory: category }} />
          </div>
          <div className="share-preview__story-copy"><p>{template === "bonus" ? "История просмотрена целиком" : "Мой сценарий года"}</p><h2>{payload.behaviorTitle}</h2></div>
        </>
      )}

      {template === "minimal" && (
        <div className="share-preview__minimal-copy">
          <span>{payload.year}</span><p>В этом году я —</p><h2>{payload.behaviorTitle}</h2>
          {achievement && <b>◇ {achievement}</b>}
        </div>
      )}

      {template === "interest" && (
        <div className="share-preview__interest-copy">
          <i aria-hidden="true">⌕</i><p>{category ? "Главный интерес года" : "Сценарий года"}</p><h2>{category ?? payload.behaviorTitle}</h2>{category && <span>{payload.behaviorTitle}</span>}
        </div>
      )}

      <div className="share-preview__facts">
        {achievement && template !== "minimal" && <section><small>Ачивка</small><strong>{achievement}</strong></section>}
        {category && template !== "interest" && <section><small>Главный интерес</small><strong>{category}</strong></section>}
      </div>
      <b className="share-preview__hashtag">#МойГодНаАвито</b>
    </article>
  );
}

function wrapCanvasText(ctx: CanvasRenderingContext2D, text: string, maxWidth: number) {
  const words = text.split(" ");
  const lines: string[] = [];
  let line = "";
  for (const word of words) {
    const candidate = line ? `${line} ${word}` : word;
    if (ctx.measureText(candidate).width > maxWidth && line) { lines.push(line); line = word; } else { line = candidate; }
  }
  if (line) lines.push(line);
  return lines;
}

function downloadShareImage(
  payload: PublicSharePayload,
  template: ShareTemplate,
  format: ShareFormat,
  tone: ShareTone,
  showAchievement: boolean,
  showCategory: boolean,
) {
  const canvas = document.createElement("canvas");
  canvas.width = 1080;
  canvas.height = format === "square" ? 1080 : 1350;
  const ctx = canvas.getContext("2d");
  if (!ctx) return;

  const gradient = ctx.createLinearGradient(0, 0, canvas.width, canvas.height);
  if (tone === "soft") {
    gradient.addColorStop(0, "#ffffff"); gradient.addColorStop(1, "#f1f5fb");
  } else if (template === "interest") {
    gradient.addColorStop(0, "#eaf8f1"); gradient.addColorStop(0.58, "#f7f0ff"); gradient.addColorStop(1, "#e7f3ff");
  } else {
    gradient.addColorStop(0, "#dbeeff"); gradient.addColorStop(0.5, "#eadfff"); gradient.addColorStop(1, "#ddf8e9");
  }
  ctx.fillStyle = gradient; ctx.fillRect(0, 0, canvas.width, canvas.height);
  ctx.fillStyle = "rgba(255,255,255,.76)"; ctx.beginPath(); ctx.roundRect(70, 70, 260, 70, 35); ctx.fill();
  ctx.fillStyle = "#0f0f0f"; ctx.font = "700 28px Arial"; ctx.fillText(`Avito · ${payload.year}`, 105, 115);

  const compact = format === "square";
  if (template === "symbol" || template === "bonus") {
    ctx.save(); ctx.translate(540, compact ? 360 : 480);
    ctx.fillStyle = "rgba(255,255,255,.56)"; ctx.beginPath(); ctx.ellipse(0, 0, compact ? 240 : 295, compact ? 180 : 220, -0.2, 0, Math.PI * 2); ctx.fill();
    ctx.fillStyle = "#8c4dff"; ctx.beginPath(); ctx.roundRect(-100, -100, 200, 200, 62); ctx.fill();
    ctx.fillStyle = "#ffffff"; ctx.font = "800 96px Arial"; ctx.textAlign = "center"; ctx.fillText("✦", 0, 35); ctx.restore();
  }

  ctx.textAlign = "left";
  const titleY = template === "symbol" || template === "bonus" ? (compact ? 610 : 820) : template === "interest" ? (compact ? 400 : 610) : (compact ? 320 : 430);
  ctx.fillStyle = "#676b73"; ctx.font = "600 34px Arial";
  ctx.fillText(template === "interest" && showCategory && payload.topCategory ? "Главный интерес года" : "Мой сценарий года", 90, titleY);
  ctx.fillStyle = "#0f0f0f"; ctx.font = `800 ${compact ? 74 : 88}px Arial`;
  const mainText = template === "interest" && showCategory ? (payload.topCategory ?? payload.behaviorTitle) : payload.behaviorTitle;
  const lines = wrapCanvasText(ctx, mainText, 900).slice(0, 3);
  lines.forEach((line, index) => ctx.fillText(line, 90, titleY + 90 + index * (compact ? 78 : 92)));

  let detailY = titleY + 120 + lines.length * (compact ? 78 : 92);
  ctx.font = "700 32px Arial";
  if (showAchievement && payload.achievementTitle) {
    ctx.fillStyle = "rgba(255,255,255,.76)"; ctx.beginPath(); ctx.roundRect(90, detailY, 900, 86, 26); ctx.fill();
    ctx.fillStyle = "#0f0f0f"; ctx.fillText(`◇ ${payload.achievementTitle}`, 125, detailY + 55); detailY += 104;
  }
  if (showCategory && payload.topCategory && template !== "interest") {
    ctx.fillStyle = "rgba(255,255,255,.76)"; ctx.beginPath(); ctx.roundRect(90, detailY, 900, 86, 26); ctx.fill();
    ctx.fillStyle = "#0f0f0f"; ctx.fillText(`⌕ ${payload.topCategory}`, 125, detailY + 55);
  }
  ctx.fillStyle = "#0b78ff"; ctx.font = "800 30px Arial"; ctx.fillText("#МойГодНаАвито", 90, canvas.height - 70);

  const link = document.createElement("a");
  link.download = `avito-recap-${payload.year}-${template}-${format}.png`;
  link.href = canvas.toDataURL("image/png"); link.click();
}

export function SharePage() {
  const { shareId } = useParams();
  const navigate = useNavigate();
  const [copied, setCopied] = useState(false);
  const [template, setTemplate] = useState<ShareTemplate>("symbol");
  const [format, setFormat] = useState<ShareFormat>("portrait");
  const [tone, setTone] = useState<ShareTone>("color");
  const [showAchievement, setShowAchievement] = useState(true);
  const [showCategory, setShowCategory] = useState(true);
  const [trailerOpen, setTrailerOpen] = useState(false);
  const query = useQuery({ queryKey: ["share", shareId], queryFn: () => getPublicShare(shareId ?? ""), enabled: Boolean(shareId) });

  if (!shareId) return <Navigate to="/" replace />;
  if (query.isPending) return <PageShell compactHeader fitViewport><PageLoader label="Готовим публичную карточку" /></PageShell>;
  if (query.isError) return <PageShell compactHeader fitViewport><ErrorState title="Публичная карточка не найдена" description="Проверьте ссылку или вернитесь к recap." onRetry={() => query.refetch()} /></PageShell>;

  const payload = query.data;
  const bonusUnlocked = hasCompletionBonus(payload.shareId);
  const templateLabels = bonusUnlocked ? [...baseTemplateLabels, bonusTemplate] : baseTemplateLabels;
  const copy = async () => {
    try { await navigator.clipboard.writeText(window.location.href); } catch { /* demo fallback */ }
    setCopied(true); window.setTimeout(() => setCopied(false), 2000);
  };
  const share = async () => {
    if (navigator.share) {
      try { await navigator.share({ title: `Мои итоги ${payload.year} на Авито`, text: `Мой сценарий года — «${payload.behaviorTitle}»`, url: window.location.href }); return; } catch { return; }
    }
    await copy();
  };

  return (
    <PageShell compactHeader fitViewport actions={<span className="public-chip">Публичная версия</span>}>
      <section className="share-page share-page--composer">
        <div className="share-page__copy">
          <span>Перед публикацией</span><h1>Собери свою публичную карточку</h1>
          <p>Меняется только представление. Данные остаются в безопасном SHARE payload — без имени, личной статистики и внутренних идентификаторов.</p>

          <div className="share-template-picker" aria-label="Стиль публичной карточки">
            {templateLabels.map((item) => <button key={item.id} type="button" className={template === item.id ? "is-active" : ""} onClick={() => setTemplate(item.id)}><strong>{item.label}</strong><span>{item.hint}</span></button>)}
          </div>
          {bonusUnlocked && <div className="share-bonus-note"><span>✦</span><p><b>Скрытый стиль открыт</b>Ты посмотрел recap целиком — поэтому появился дополнительный вариант оформления.</p></div>}

          <div className="share-editor-row" aria-label="Формат карточки">
            <span>Формат</span><button type="button" className={format === "portrait" ? "is-active" : ""} onClick={() => setFormat("portrait")}>4:5</button><button type="button" className={format === "square" ? "is-active" : ""} onClick={() => setFormat("square")}>1:1</button>
          </div>
          <div className="share-editor-row" aria-label="Оформление карточки">
            <span>Фон</span><button type="button" className={tone === "color" ? "is-active" : ""} onClick={() => setTone("color")}>Цветной</button><button type="button" className={tone === "soft" ? "is-active" : ""} onClick={() => setTone("soft")}>Светлый</button>
          </div>
          <div className="share-editor-row share-editor-row--toggles" aria-label="Публичные поля">
            <span>Показывать</span>
            {payload.achievementTitle && <button type="button" className={showAchievement ? "is-active" : ""} onClick={() => setShowAchievement((value) => !value)}>◇ Ачивку</button>}
            {payload.topCategory && <button type="button" className={showCategory ? "is-active" : ""} onClick={() => setShowCategory((value) => !value)}>⌕ Категорию</button>}
          </div>

          <div className="privacy-checklist privacy-checklist--compact" aria-label="Проверка приватности"><span>✓ имя скрыто</span><span>✓ числовые метрики скрыты</span><span>✓ только разрешённые итоги</span></div>
        </div>

        <div className="share-wrap">
          <SharePreview payload={payload} template={template} format={format} tone={tone} showAchievement={showAchievement} showCategory={showCategory} />
          <div className="share-actions share-actions--four">
            <Button fullWidth onClick={share}>Поделиться</Button>
            <Button variant="secondary" fullWidth onClick={() => downloadShareImage(payload, template, format, tone, showAchievement, showCategory)}>PNG</Button>
            <Button variant="secondary" fullWidth onClick={() => setTrailerOpen(true)}>Трейлер</Button>
            <Button variant="secondary" fullWidth onClick={copy}>{copied ? "Скопировано ✓" : "Ссылка"}</Button>
          </div>
          <button className="share-back" type="button" onClick={() => navigate(-1)}>← Вернуться к итогам</button>
        </div>
      </section>
      <YearTrailerDialog open={trailerOpen} payload={payload} soundEnabled={false} onClose={() => setTrailerOpen(false)} />
    </PageShell>
  );
}
