# Iteration 7 — Sonic identity & tactile finale

Эта итерация развивает исключительно пользовательский frontend и не добавляет новых backend-полей.

## Что добавлено

- собственный `SonicEngine` на Web Audio API;
- разные звуковые профили для RESEARCHER / FIND_HUNTER / ACTIVE_SELLER / STARTING_SELLER / DECISIVE_BUYER / UNIVERSAL_USER;
- характер звука: soft glass + paper noise + warm synth tail;
- отдельные sonic cues для навигации, метрик, behavior reveal, ачивок, CTA и финала;
- короткий «момент тишины» перед раскрытием BEHAVIOR;
- progressive color reveal: визуальная среда становится насыщеннее по мере прохождения recap;
- haptic feedback как progressive enhancement;
- физический свет на achievement tokens, следующий за курсором;
- completion easter egg после полного просмотра;
- скрытый SHARE-шаблон `Контур года`, доступный только после полного просмотра на этом устройстве;
- интерактивный личный разбор визуального символа года;
- tilt/rotation символа курсором и объяснение, какие backend-derived данные отвечают за форму, категорию, месяц и ачивки;
- специальные переходы TOP_CATEGORY → ACTIVE_MONTH, BEHAVIOR → ACHIEVEMENT и ACHIEVEMENT/MISSED_OPPORTUNITY → NEXT_ACTION;
- сравнение двух уже просмотренных тестовых историй на странице профилей без сравнения приватных числовых метрик.

## Приватность

Интерактивный разбор символа — личный UI и не становится публичным payload. Публичная карточка по-прежнему использует только `PublicSharePayload`.

## Accessibility

- звук выключен по умолчанию;
- haptics работают только при наличии `navigator.vibrate`;
- `prefers-reduced-motion` отключает reveal-паузу, tilt и декоративную динамику;
- все новые модальные сценарии закрываются по Escape;
- базовая клавиатурная навигация recap сохранена.
