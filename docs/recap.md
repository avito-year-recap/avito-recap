# Recap

## 1. recap

Recap — неизменяемый снимок итогов пользователя за **полностью завершённый календарный год UTC**. Он объединяет:

- исторические метрики за год;
- определённый по прозрачным правилам тип поведения;
- до трёх достижений;
- одно исполнимое следующее действие (`NextAction`);
- последовательность story-карточек;
- отдельную публичную `ShareCard`, построенную по privacy allow-list.

Recap не является live-дашбордом. После генерации он не пересчитывается при изменении текущего состояния пользователя. Новое поведение правил означает новый idempotency key и, соответственно, новый снимок.

## 2. Период и условия генерации

Год интерпретируется как полуинтервал:

```text
[YYYY-01-01 00:00:00 UTC, (YYYY+1)-01-01 00:00:00 UTC)
```

Генерировать recap можно только за завершённый год:

- `year == 0` или год в будущем → `ErrInvalidYear`;
- текущий незавершённый год → `ErrYearNotComplete`;
- меньше 5 событий в годовых метриках → `ErrNotEnoughActivity`.

Минимум активности задаётся `Eligibility.MinEvents = 5`.

`TotalEvents` в seed/testkit считается как сумма:

```text
searches
+ listingViews
+ favoritesAdded
+ chatsStarted
+ listingsCreated
+ listingsPublished
+ purchasesCompleted
+ salesCompleted
```

`activeDays`, `categoriesCount` и `chatsWithPurchase` в `TotalEvents` отдельно не добавляются: это производные/дополнительные сигналы, а `chatsWithPurchase` является подмножеством начатых диалогов.

## 3. Что участвует в расчёте

### Исторические метрики

`AnalyticsStorage.CalculateMetrics` возвращает метрики строго за `RecapPeriod`:

- поиски и просмотры;
- уникальные и повторные просмотры;
- избранное;
- диалоги и диалоги, связанные с покупкой;
- создание/публикация объявлений;
- покупки и продажи;
- активные дни;
- число категорий;
- главную категорию и активный месяц;
- активность по отдельным категориям для тематических ачивок.

Две доли рассчитываются самим движком:

```text
repeat_rate   = repeated_views / total_views
purchase_rate = chats_with_purchase / chats_started
```

При нулевом знаменателе значение равно `0`.

### Текущее состояние (`ActionableState`)

Для CTA используется отдельный point-in-time snapshot, полученный на момент генерации:

- текущие черновики и ID конкретного черновика;
- открытые диалоги и ID конкретного диалога;
- активные объявления и ID конкретного объявления;
- текущее число избранных объектов;
- наличие сохранённого поиска по главной категории;
- ID последней покупки;
- факт того, публиковал ли пользователь объявления раньше.

Важно: **годовые счётчики не заменяют текущий snapshot**. Например, разница `listingsCreated - listingsPublished` может быть историческим сигналом для поведения `STARTING_SELLER`, но `FINISH_DRAFT` выдаётся только при наличии актуального черновика с адресуемым `listing_id`.

`ActionableState` хранится во внутреннем доменном recap для воспроизводимости и integrity-проверки, но отсутствует в публичном protobuf `Recap` и тем более в `ShareCard`.

## 4. Pipeline генерации

`application.Service.Generate` выполняет следующую последовательность:

1. Валидирует `profile_id` и завершённость года.
2. Строит `RecapKey`.
3. Пытается найти уже сохранённый recap по этому ключу.
4. Загружает и валидирует профиль.
5. Получает годовые метрики.
6. Получает актуальный `ActionableState` на время генерации.
7. Создаёт два разных UUID: внутренний `recap_id` и публичный `share_id`.
8. Передаёт входы в `engine.Build`.
9. Движок нормализует данные, рассчитывает derived-поля и валидирует результат.
10. `RecapStorage.CreateRecapIfAbsent` атомарно сохраняет объект либо возвращает уже созданный конкурентным запросом.
11. Перед возвратом сохранённый объект повторно проходит integrity validation.

Схематично:

```text
Profile + yearly Metrics + current ActionableState
                    |
                    v
              Recap Engine
      +-------------+-------------+
      |             |             |
   Behavior    Achievements   NextAction
      |             |             |
      +-------------+-------------+
                    |
             Story Cards
                    |
             Share Projection
                    |
              immutable Recap
```

## 5. Версия правил и идемпотентность

Текущие значения:

```text
rules_version = 3.5.0
algorithm     = recap-v3.5-full-next-actions-v1
privacy       = privacy-v2
```

Ключ идемпотентности:

```text
(profile_id, year, rules_version, rules_digest)
```

`rules_digest` — SHA-256 fingerprint нормализованной конфигурации ruleset: порогов, приоритетов, achievement policy и privacy policy. `algorithm` входит в эту конфигурацию и используется как явная версия исполняемой логики.

Следствия:

- одинаковый профиль, год и rules identity → возвращается тот же recap;
- изменение правил должно менять version/algorithm/config и тем самым identity;
- `internal_id` и `share_id` уникальны и не могут совпадать;
- production storage обязан обеспечивать атомарность `CreateRecapIfAbsent` и уникальность idempotency key, `internal_id` и `share_id`.

Хранилище считается недоверенным. При `Get`, `GetShareCard` и idempotent read движок заново выводит `Behavior`, `Achievements`, `NextAction` и `Cards` из сохранённых исходных данных и сравнивает их с сохранённой проекцией.

## 6. Behavior

Behavior определяется не score-моделью, а набором бинарных eligibility rules. Если подходят несколько сценариев, выбирается сценарий с более высоким product priority.

| Код | Условие | Priority |
|---|---|---:|
| `ACTIVE_SELLER` | `listings_published >= 5` и `sales_completed >= 3` | 50 |
| `DECISIVE_BUYER` | `purchases >= 3`, `chats >= 5`, `chats_with_purchase >= 3`, `purchase_rate >= 20%` | 40 |
| `STARTING_SELLER` | `listings_created >= 3`, `listings_published <= 2`, созданий больше публикаций, продаж `0` | 30 |
| `FIND_HUNTER` | `total_views >= 20`, `favorites_added >= 3`, `repeat_rate >= 20%` | 20 |
| `RESEARCHER` | `total_views >= 100`, `categories_count >= 5`, `chats_started <= 4` | 10 |
| `UNIVERSAL_USER` | fallback, если ни одно специализированное правило не выполнено | — |

Например, `seller-buyer-hybrid` одновременно проходит правила `ACTIVE_SELLER` и `DECISIVE_BUYER`, но получает `ACTIVE_SELLER`, потому что его priority выше.

В `BehaviorEvidence` сохраняются реальные метрики, пороги и текстовое объяснение. Полей `score`/`points` в актуальном контракте нет.

## 7. Достижения

Пользователь получает максимум **3** достижения. Сначала движок определяет все заработанные кандидаты, затем собирает продуктовый портфель.

### Базовые достижения

| Код | Название | Условие |
|---|---|---|
| `SUCCESSFUL_SELLER` | Мастер переговоров | `sales_completed >= 5` |
| `CONSISTENT_PUBLISHER` | Маяк стабильности | `listings_published >= 5` и `sales_completed >= 1` |
| `DEAL_CLOSER` | Сделка состоялась | `purchases_completed >= 3` |
| `QUICK_DECISION` | Молния | те же условия, что у `DECISIVE_BUYER` |
| `BROAD_INTERESTS` | Человек-оркестр | `categories_count >= 6` |
| `ATTENTIVE_RESEARCHER` | Стратег | `total_views >= 150` |
| `MASTER_OF_FAVORITES` | Собиратель жемчужин | `favorites_added >= 20` |
| `ALL_ROUNDER` | Человек-швейцарский нож | минимум 5 покупок и 5 продаж; разница не больше 50% от большего числа |
| `FIRST_SELLING_STEPS` | Начинающий бизнесмен | starting-seller signal либо от 1 до 4 завершённых продаж |

### Тематические достижения

Тематическая группа `INTEREST` строится по `Metrics.CategoryActivities`.

Сигнал категории:

```text
signal = views + favorites_added * 4 + purchases_completed * 12
```

Чтобы тематическая ачивка считалась заработанной, одновременно требуется:

1. хотя бы один объёмный сигнал: `30` просмотров **или** `8` добавлений в избранное **или** `3` покупки;
2. доля тематического сигнала не меньше `20%` от всего категорийного сигнала пользователя.

| Категория | Код | Название |
|---|---|---|
| женская мода / красота | `STYLE_ICON` | Икона стиля |
| мужская мода | `FASHIONABLE_MAN` | Модник |
| туризм | `TRAVELER` | Путешественник |
| дача и сад | `FOR_THE_SOUL` | Для души |
| книги | `BOOKWORM` | Книжный червь |
| украшения | `BEAUTY_CONNOISSEUR` | Ценитель прекрасного |
| музыка | `IN_THE_RHYTHM_OF_MUSIC` | В ритме музыки |
| игрушки | `WORLD_OF_PLAY` | Мир игры |
| инструменты | `MASTER_CRAFT` | Дело мастера |
| товары для животных | `CARING_OWNER` | Заботливый хозяин |
| детские товары | `LITTLE_DISCOVERIES` | Для маленьких открытий |

Тематические достижения всегда создаются с `Shareable=false` и не попадают в публичный achievement allow-list.

### Как собирается портфель

- Сбалансированный seller+buyer: `ALL_ROUNDER` + лучшая selling-ачивка + лучшая buying-ачивка.
- Только продажи: одна сильнейшая selling-ачивка.
- Продаж больше, чем покупок: selling-ачивки имеют преимущество, затем добавляются другие категории.
- Только покупки: сначала до трёх тематических personas, затем свободные места заполняются лучшими общими категориями.
- В остальных случаях: лучшая ачивка из каждой логической категории, пока не достигнут лимит.

Порядок детерминирован: выше `Priority`, затем сильнее измеренный сигнал, затем стабильный код.

## 8. NextAction

`NextAction` — ровно одно исполнимое действие. Каждый action содержит структурированный `ActionTarget`, в котором заполнен ровно один destination.

Приоритеты версии 3.5.0:

| Priority | Action | Когда выбирается | Target |
|---:|---|---|---|
| 1000 | `FINISH_DRAFT` | есть текущий черновик и его ID | listing |
| 900 | `CONTINUE_DIALOGS` | есть открытый диалог и его ID | dialog |
| 800 | `IMPROVE_LISTINGS` | минимум 3 активных объявления и есть конкретный listing ID | listing |
| 750 | `VIEW_SIMILAR_LISTINGS` | behavior `DECISIVE_BUYER` и есть ID последней покупки | listing |
| 700 | `SAVE_SEARCH` | behavior `RESEARCHER`, есть top category и поиск по ней не сохранён | search/category code |
| 650 | `OPEN_FAVORITES` | в текущем состоянии есть избранное | route `/favorites` |
| 520 | `CREATE_FIRST_LISTING` | `STARTING_SELLER`, пользователь ещё не публиковал, нет активных/черновиков | route `/listings/new` |
| 500 | `CREATE_LISTING` | seller behavior, публикации раньше были, сейчас нет активных/черновиков | route `/listings/new` |
| 400 | `OPEN_TOP_CATEGORY` | есть безопасный top category code | category |
| 0 | `EXPLORE_RECOMMENDATIONS` | fallback | route `/recommendations` |

Текущее адресуемое состояние специально имеет приоритет над исторической персоной. Например, пользователь может быть `FIND_HUNTER`, но при наличии актуального черновика получит `FINISH_DRAFT`.

## 9. Story cards

Для валидного recap карточки идут в фиксированном порядке:

1. `INTRO` — всегда, без payload.
2. `YEAR_ACTIVITY` — всегда.
3. `TOP_CATEGORY` — только если есть главная категория и просмотры по ней.
4. `ACTIVE_MONTH` — для валидного годового recap.
5. `BEHAVIOR` — всегда.
6. `ACHIEVEMENT` — только если выдана хотя бы одна ачивка.
7. `MISSED_OPPORTUNITY` — только для `SAVE_SEARCH` или `FINISH_DRAFT`.
8. `NEXT_ACTION` — всегда.
9. `SHARE` — всегда последняя.

Позиции непрерывные и начинаются с `1`.

Только финальная `SHARE` имеет `shareable=true`. Все остальные карточки — часть приватного recap.

`CardPayload` — закрытый union, соответствующий protobuf `oneof`. JSON decoder отклоняет неизвестные поля, неизвестные типы и несовместимые пары `type/payload`.

## 10. Public ShareCard и privacy

Публичный share flow использует отдельный случайный `share_id`; внутренний `recap_id` публично не раскрывается.

`ShareCard` содержит только:

```text
share_id
year
behavior_title
achievement_title?  // максимум одно
top_category?       // только если разрешено
privacy_version
```

В ней принципиально отсутствуют:

- profile ID и внутренний recap ID;
- сырые метрики;
- `ActionableState`;
- CTA и action targets;
- список всех ачивок.

Privacy policy `privacy-v2` работает как allow-list:

- behavior title проходит safety-проверку текста;
- выбирается первое выданное достижение, которое одновременно `Shareable=true` и находится в явном achievement allow-list;
- тематические достижения в allow-list не входят;
- top category публикуется только если код известен каталогу, категория разрешена для share и upstream flag тоже разрешает share;
- публичный текст ограничен 80 Unicode code points и не может содержать control/bidi-control symbols.

Категории, которые текущий каталог **не разрешает** публиковать как top category:

- `beauty_cosmetics`;
- `jewelry`;
- `pets`;
- `kids`.

Финальная story-карточка использует тот же DTO `ShareCard`, что и `GetShareCard`, поэтому preview и публичная проекция не должны расходиться.

## 11. API-контракт

В `proto/recap/recap.proto` определены четыре RPC:

| RPC | Назначение |
|---|---|
| `ListProfiles` | вернуть каталог demo/test профилей |
| `GenerateRecap` | создать или вернуть immutable recap для `profile_id + year + rules identity` |
| `GetRecap` | получить приватный recap по `internal_recap_id` |
| `GetShareCard` | получить публичную проекцию по `share_id` |

UUID передаются строками и должны валидироваться transport-слоем до вызова application service.

Рекомендуемая карта статусов описана комментариями в `.proto`: invalid input → `INVALID_ARGUMENT`, незавершённый год/недостаток активности → `FAILED_PRECONDITION`, not found → `NOT_FOUND`, повреждённый сохранённый recap → `DATA_LOSS`.

### Важное ограничение текущего репозитория

В этом snapshot есть доменный/application слой и protobuf-контракт, но **готовый запускаемый API-сервер не собран**: `cmd/api/main.go` содержит только объявление package. Поэтому документация ниже не предлагает несуществующую команду запуска сервера.

Генерация protobuf-кода предусмотрена:

```bash
make generate
```

Также доступны:

```bash
make proto-fmt
make proto-lint
make proto-breaking BASE_BRANCH=master
```

## 12. Где менять правила

| Задача | Файл/пакет |
|---|---|
| версия, пороги, приоритеты, privacy | `internal/recap/ruleset/` |
| годовой период и derived rates | `internal/recap/analytics/` |
| behavior rules | `internal/recap/behavior/` |
| достижения и portfolio | `internal/recap/achievement/` |
| NextAction | `internal/recap/nextaction/` |
| story cards | `internal/recap/presentation/cards/` |
| public share projection | `internal/recap/presentation/share/` |
| orchestration | `internal/recap/application/` и `internal/recap/engine/` |
| structural/domain validation | `internal/recap/validation/` |
| protobuf API | `proto/recap/recap.proto` |
| тестовые сценарии | [`test-profiles.md`](test-profiles.md) |

При изменении продуктового поведения нужно обновлять rules identity и тестовые ожидания. Если меняется внешний wire contract, дополнительно обновляется `.proto` и выполняются proto lint/breaking checks.

## 13. Проверки

Основной набор:

```bash
go test -count=1 ./...
go vet ./...
go test -race -p=1 -count=1 ./...
go test -cover ./internal/recap/...
```

