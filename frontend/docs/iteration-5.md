# Iteration 5 — visual story system

Цель: превратить готовый server-driven recap в цельный эмоциональный продукт, не расширяя backend-контракт и не придумывая пользовательские метрики.

## Что добавлено

1. **Персональный визуальный код года** — локально и детерминированно собирается из уже полученных `behavior.code`, `categoryCode`, `month` и achievement codes. Маленькая версия остаётся рядом с progress и развивается по мере прохождения истории.
2. **Связная режиссура переходов** — разные типы карточек входят по-разному: цифры, категория, behavior reveal, достижения и SHARE больше не ощущаются одинаковыми слайдами.
3. **YEAR_ACTIVITY mini-show** — `totalEvents` появляется первым, затем по очереди раскрываются семь переданных счётчиков. Никаких метрик сверх payload.
4. **BEHAVIOR reveal** — сценарий пользователя становится кульминационным экраном; `score`/`evidence` остаются в explanation sheet.
5. **Collectible achievements** — максимум три серверные ачивки визуализируются как отдельные жетоны с unlock-анимацией и hover-реакцией.
6. **Adaptive visual language** — behavior/category/action визуально меняют motif: orbit, trail, stack, constellation, grid или pulse.
7. **Seasonal ACTIVE_MONTH** — номер месяца превращается в атмосферную сезонную сцену. Число событий за месяц не выдумывается.
8. **Interactive MISSED_OPPORTUNITY** — локальная демонстрация `SAVE_SEARCH` и `FINISH_DRAFT` показывает пользу действия до реального CTA.
9. **NEXT_ACTION finale** — один основной action, визуальный акцент и объяснение «Почему рекомендуем?».
10. **SHARE composer** — три шаблона (`Символ года`, `Минимализм`, `Главный интерес`), все строятся исключительно из `PublicSharePayload`.
11. **PNG export** — публичную карточку можно сохранить как PNG через Canvas API без сторонних библиотек.
12. **Replay map** — после первого достижения финала появляется кнопка «Моменты года» для быстрого возврата к любому экрану.
13. **Reduced motion** — ключевые эффекты выключаются через `prefers-reduced-motion` / Framer Motion `useReducedMotion`.
14. **Viewport-fit** — recap и share composer используют `100dvh`-layout и не требуют скролла самой страницы; длинный explanation остаётся единственным внутренне прокручиваемым слоем.

## Privacy

Большой публичный символ на SHARE-странице намеренно строится только из:

- `year`;
- `behaviorTitle`;
- `achievementTitle` (если есть);
- `topCategory` (если разрешена backend-ом).

Приватные `profile`, числовые метрики, `score`, `evidence`, CTA targets и внутренние ID в публичный визуал не попадают.
