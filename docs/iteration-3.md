# Iteration 3 — product UX

Цель итерации: перестать выглядеть как набор одинаковых dashboard-карточек и сделать recap цельной пользовательской историей, не придумывая данные, которых нет в контракте.

## Изменения

1. **User mode без технической панели** — backend-типы и rules не занимают пользовательский интерфейс.
2. **Story navigation** — progress, tap-зоны, swipe, стрелки клавиатуры и `?slide=N`.
3. **Уникальная режиссура card.type** — отдельная композиция для всех 9 типов.
4. **Visual registry** — `categoryCode`, `behavior.code`, achievement codes и action codes меняют визуальную метафору без изменения API.
5. **Explanation sheet** — один паттерн «Почему?» для поведения, метрик, месяца, ачивок и CTA. Для BEHAVIOR показывается реальный `evidence` и `score`.
6. **Active month без выдуманных данных** — только номер/название месяца и визуальная шкала месяцев, без fake heatmap и количества событий.
7. **CTA реально продолжается** — каждый typed target ведёт на demo destination page.
8. **Share preview** — отдельный публичный сценарий с явной проверкой приватности и Web Share API fallback.
9. **Accessibility** — aria-live, `aria-current=step`, focus trap, возврат focus, reduced motion.
10. **Viewport-fit режим** — recap, progress и управление адаптируются по высоте окна и не требуют прокрутки страницы.

## Контракт остаётся главным

Frontend не определяет:

- какие карточки существуют;
- их порядок;
- behavior;
- score;
- achievement codes;
- следующий action;
- public share payload.

Frontend отвечает за визуальную интерпретацию уже готового server-driven recap.
