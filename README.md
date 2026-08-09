# Avito Recap Frontend

Frontend MVP персональных «Итогов года» в светлом визуальном стиле. Данные мокируются, но структура и ограничения повторяют присланную продуктовую спецификацию: динамическая история из 5–9 карточек, server-driven порядок, `evidence`, типизированный CTA и отдельный безопасный SHARE payload.

## Что реализовано

- 5 тестовых профилей с разным количеством карточек, поведением, ачивками и CTA;
- обязательные `INTRO`, `YEAR_ACTIVITY`, `BEHAVIOR`, `NEXT_ACTION`, `SHARE`;
- условные `TOP_CATEGORY`, `ACTIVE_MONTH`, `ACHIEVEMENT`, `MISSED_OPPORTUNITY`;
- восемь метрик `YEAR_ACTIVITY` без придуманных frontend-данных;
- `ACTIVE_MONTH` использует только номер месяца — без фейкового heatmap и числовой активности;
- `BEHAVIOR` с `score` и `evidence` в отдельном explanation sheet;
- ачивки связываются по `codes` с верхнеуровневым `recap.achievements`;
- CTA использует `recap.nextAction.buttonText` и типизированный target;
- отдельный публичный mock endpoint и маршрут с минимальным SHARE payload;
- полноценные mock-destination страницы для всех CTA;
- персональный «визуальный код года», который по мере просмотра собирается из поведения, категории, месяца и ачивок;
- разные переходы между типами сцен + постоянный мини-тотем в progress-зоне;
- staged reveal для YEAR_ACTIVITY и отдельный reveal-момент BEHAVIOR;
- интерактивная демонстрация пользы MISSED_OPPORTUNITY без новых backend-данных;
- replay-карта «Моменты года» после первого полного просмотра;
- SHARE composer с тремя безопасными шаблонами, Web Share API, копированием ссылки и экспортом PNG через Canvas API;
- свайп, tap-зоны, стрелки клавиатуры и progress navigation;
- текущий экран хранится в `?slide=N`, поэтому recap можно открыть с конкретного шага;
- recap-экран фиксируется в пределах текущего viewport: progress, карточка и навигация помещаются без прокрутки страницы;
- `prefers-reduced-motion`, aria-live, focus trap и возврат focus после explanation sheet;
- visual registry для category / behavior / achievement / action codes;
- mapper между backend-подобным DTO и frontend-моделью;
- unit-тесты mapper, privacy payload и action target URL.

## UX-направление

Интерфейс больше не выглядит как девять одинаковых карточек. Каждый тип данных получает собственную композицию:

- `INTRO` — старт персонального визуального кода года;
- `YEAR_ACTIVITY` — крупное общее число + иерархия семи метрик;
- `TOP_CATEGORY` — визуальная сцена категории;
- `ACTIVE_MONTH` — типографический постер месяца;
- `BEHAVIOR` — главный персональный постер и прозрачное объяснение результата;
- `ACHIEVEMENT` — коллекционные жетоны с unlock-анимацией;
- `MISSED_OPPORTUNITY` — интерактивное «до → после» для SAVE_SEARCH / FINISH_DRAFT;
- `NEXT_ACTION` — один сильный CTA;
- `SHARE` — отдельная безопасная композиция и выбор одного из трёх публичных шаблонов.

## Запуск

```bash
npm install
npm run dev
```

Проверки:

```bash
npm run check
npm run build
```

> В sandbox TypeScript typecheck и ESLint проходят. Запуск Vite/Vitest в этом окружении упирается в отсутствующий Linux native binding Rolldown из импортированного `node_modules`; в обычном окружении после чистого `npm install` используйте команды выше.

## Полезные URL

```text
/                                  выбор профиля
/generate/:profileCode             генерация
/recap/:profileCode                пользовательский recap
/recap/:profileCode?slide=5        deep-link на конкретный экран
/share/:shareId                    безопасная публичная карточка
/demo/action/:actionCode           mock следующего действия
```

## Архитектура

```text
src/
├── app/                    router + providers
├── pages/                  profiles, generation, recap, share, action demo
├── widgets/                recap player + explanation sheet + replay map
├── entities/recap/         domain model, mapper, visual card renderer
├── features/next-action/   typed target → demo destination
└── shared/
    ├── api/                 backend contract + mock API + ConnectRPC transport
    ├── lib/                 visual registry
    ├── styles/              global tokens
    └── ui/                  reusable UI primitives + YearTotem
```

Поток данных:

```text
mock backend DTO
      ↓
mapper
      ↓
frontend discriminated union
      ↓
RecapCardRenderer
      ↓
уникальная композиция по card.type
```

Для подключения ConnectRPC используйте `src/shared/api/connect-transport.ts`, сгенерированный service descriptor и замените реализацию `src/shared/api/recap-api.ts`, сохранив интерфейс frontend-модели `Recap`.

## Дизайн-референс

Исходное светлое направление сохранено в `docs/design-reference-white.png`. Текущая реализация развивает его в сторону story experience: больше белого пространства, меньше dashboard-плиток, сильнее типографика, персональные visual codes и отдельный безопасный share-flow.

## UX iteration 6

Поверх server-driven recap добавлен расширенный пользовательский слой без изменения backend-контракта: cinematic autoplay, опциональные звуки, resume, интерактивный портрет поведения, visual easter eggs, pointer physics, динамическая атмосфера, before/after CTA, public-safe trailer и расширенный SHARE editor. Подробности — `docs/iteration-6.md`.

## Iteration 7 · Sonic identity

Последняя UX-итерация добавляет frontend-only sonic identity, behavior-dependent Web Audio sound profiles, haptic feedback, момент тишины перед behavior reveal, интерактивный разбор визуального символа, completion easter egg со скрытым SHARE-шаблоном, физический свет achievement tokens и comparison уже просмотренных историй. Новых backend-полей для этих функций не требуется. Подробнее: `docs/iteration-7.md`.

## Protobuf / ConnectRPC

Frontend RPC contract is defined in `proto/avito/recap/v1/recap.proto`.
It contains the complete typed recap card model, behavior evidence, achievements,
typed CTA targets and the privacy-safe public SHARE payload.

```bash
npm run proto:lint       # validate the schema
npm run proto:format     # format .proto files
npm run proto:generate   # generate TypeScript into src/shared/api/generated
```

`RecapCard.payload` and `ActionTarget.target` use Protobuf `oneof`, so the future
ConnectRPC client can narrow payloads safely in TypeScript. The current demo
continues to use mocks until a real RPC endpoint is connected.

See `docs/protobuf-contract.md` for the mapping and integration notes.
