# Тестовые профили Recap

В репозитории есть 17 воспроизводимых demo-профилей. Они нужны не как случайные фикстуры, а как **каталог продуктовых сценариев**: через обычный `Service.Generate` покрываются все 6 behavior-кодов и все 10 `NextAction`-кодов версии `3.5.0`, а также базовые и тематические достижения, privacy-кейсы и граничные условия.

## 1. Из каких файлов состоит профиль

Один логический тестовый профиль описан сразу в нескольких местах:

| Файл | Что хранит |
|---|---|
| `seeds/profiles.json` | ID, machine code, имя, описание, avatar URL |
| `seeds/scenarios.json` | годовая активность, категории, месяцы и текущее `ActionableState` |
| `testdata/expected/recaps.json` | ожидаемые behavior, achievements, NextAction, top category, active month и последовательность карточек |
| `testdata/golden/<profile-code>.json` | полный ожидаемый внутренний `model.Recap` |
| `static/avatars/<profile-code>.png` | локальная avatar-заглушка |

Связующий ключ — `profileCode` / `code`. Для каждого кода должна существовать ровно одна запись в `profiles`, одна в `scenarios`, одна в `expected` и один golden-файл.

## 2. Какой профиль брать

| Профиль | Behavior | Achievements | NextAction | Для чего использовать |
|---|---|---|---|---|
| `active-buyer` | `FIND_HUNTER` | `ATTENTIVE_RESEARCHER`, `MASTER_OF_FAVORITES` | `OPEN_FAVORITES` | активный поиск, повторные просмотры и избранное |
| `active-seller` | `ACTIVE_SELLER` | `SUCCESSFUL_SELLER` | `IMPROVE_LISTINGS` | сильный seller flow с адресуемым активным объявлением |
| `researcher` | `RESEARCHER` | `BROAD_INTERESTS` | `SAVE_SEARCH` | researcher thresholds + `MISSED_OPPORTUNITY` для сохранённого поиска |
| `universal-user` | `UNIVERSAL_USER` | `FIRST_SELLING_STEPS` | `CONTINUE_DIALOGS` | fallback behavior при смешанной активности + приоритет текущего диалога |
| `draft-seller` | `STARTING_SELLER` | `FIRST_SELLING_STEPS` | `FINISH_DRAFT` | starting seller + актуальный черновик + `MISSED_OPPORTUNITY` |
| `decisive-buyer` | `DECISIVE_BUYER` | `QUICK_DECISION` | `VIEW_SIMILAR_LISTINGS` | buyer conversion и адресуемая последняя покупка |
| `seller-buyer-hybrid` | `ACTIVE_SELLER` | `ALL_ROUNDER`, `SUCCESSFUL_SELLER`, `QUICK_DECISION` | `IMPROVE_LISTINGS` | пересечение seller/buyer rules, проверка behavior priority и balanced portfolio |
| `returning-publisher` | `ACTIVE_SELLER` | `CONSISTENT_PUBLISHER` | `CREATE_LISTING` | продавец с историей публикаций, но без текущего listing/draft |
| `category-browser` | `UNIVERSAL_USER` | — | `OPEN_TOP_CATEGORY` | нейтральный behavior с безопасной главной категорией |
| `recommendation-newcomer` | `UNIVERSAL_USER` | — | `EXPLORE_RECOMMENDATIONS` | минимальный допустимый recap без category signal, чистый fallback CTA |
| `steady-buyer` | `UNIVERSAL_USER` | `DEAL_CLOSER` | `OPEN_TOP_CATEGORY` | покупки есть, но conversion evidence недостаточно для `DECISIVE_BUYER`/`QUICK_DECISION` |
| `book-collector` | `DECISIVE_BUYER` | `QUICK_DECISION`, `BOOKWORM` | `OPEN_FAVORITES` | buyer-only + тематическая ачивка по книгам |
| `private-style-hunter` | `FIND_HUNTER` | `MASTER_OF_FAVORITES`, `STYLE_ICON` | `OPEN_FAVORITES` | privacy: личная тематическая ачивка и приватная top category не должны утечь в share |
| `maker-with-draft` | `STARTING_SELLER` | `FIRST_SELLING_STEPS`, `MASTER_CRAFT` | `FINISH_DRAFT` | тематическая ачивка + более приоритетный адресуемый текущий черновик |
| `pet-threshold-buyer` | `DECISIVE_BUYER` | `QUICK_DECISION`, `CARING_OWNER` | `OPEN_TOP_CATEGORY` | точная проверка decisive-buyer thresholds + приватная категория `pets` |
| `music-traveler` | `FIND_HUNTER` | `DEAL_CLOSER`, `IN_THE_RHYTHM_OF_MUSIC`, `TRAVELER` | `OPEN_TOP_CATEGORY` | несколько тематических интересов и лимит портфеля из трёх ачивок |
| `listing-restart` | `STARTING_SELLER` | `FIRST_SELLING_STEPS` | `CREATE_FIRST_LISTING` | были попытки создать объявления, но ещё не было ни одной публикации |

Все seed-сценарии сейчас рассчитаны на `2025` год и тестируются с фиксированным clock `2026-08-04T12:00:00Z`, то есть год гарантированно завершён.

## 3. Покрытие Behavior

Для проверки конкретного behavior удобно использовать:

| Behavior | Основной профиль |
|---|---|
| `ACTIVE_SELLER` | `active-seller` |
| `STARTING_SELLER` | `draft-seller` |
| `DECISIVE_BUYER` | `decisive-buyer` |
| `FIND_HUNTER` | `active-buyer` |
| `RESEARCHER` | `researcher` |
| `UNIVERSAL_USER` | `category-browser` или `recommendation-newcomer` |

Отдельный `seller-buyer-hybrid` нужен для ситуации, когда подходит больше одного специализированного behavior. Он одновременно проходит `ACTIVE_SELLER` и `DECISIVE_BUYER`, но ожидаемый результат — `ACTIVE_SELLER` из-за большего priority.

Для низкоуровневых boundary cases behavior также есть `testdata/metrics/behavior_cases.json`.

## 4. Покрытие NextAction

| NextAction | Профиль |
|---|---|
| `FINISH_DRAFT` | `draft-seller` / `maker-with-draft` |
| `CONTINUE_DIALOGS` | `universal-user` |
| `IMPROVE_LISTINGS` | `active-seller` / `seller-buyer-hybrid` |
| `VIEW_SIMILAR_LISTINGS` | `decisive-buyer` |
| `SAVE_SEARCH` | `researcher` |
| `OPEN_FAVORITES` | `active-buyer` / `book-collector` / `private-style-hunter` |
| `CREATE_FIRST_LISTING` | `listing-restart` |
| `CREATE_LISTING` | `returning-publisher` |
| `OPEN_TOP_CATEGORY` | `category-browser`, `steady-buyer`, `pet-threshold-buyer`, `music-traveler` |
| `EXPLORE_RECOMMENDATIONS` | `recommendation-newcomer` |

Это покрытие проверяется автоматически в `TestSeedCatalogueCoversAllBehaviorsActionsAndCoreAchievements`.

## 5. Формат `seeds/profiles.json`

Пример:

```json
{
  "id": "26a3f4e0-1ae7-5b46-b2b6-2ae9fc180ba2",
  "code": "active-buyer",
  "displayName": "Алексей",
  "description": "Часто ищет технику, сохраняет интересные варианты и возвращается к ним",
  "avatarUrl": "/avatars/active-buyer.png"
}
```

Требования, которые уже проверяются тестами:

- `id` — валидный ненулевой UUID;
- `code` уникален;
- UUID не переиспользуются между профилями и адресуемыми объектами сценариев;
- avatar URL имеет вид `/avatars/<name>.png`;
- соответствующий PNG существует в `static/avatars/`.

`code` — стабильный machine-readable идентификатор. Имя пользователя и описание можно менять как presentation data, но код нельзя менять без синхронного обновления всех связанных fixture-файлов.

## 6. Формат `seeds/scenarios.json`

Сценарий состоит из четырёх частей:

```json
{
  "profileCode": "active-buyer",
  "year": 2025,
  "activity": { "...": "..." },
  "categories": [ "..." ],
  "months": [ "..." ],
  "actionableState": { "...": "..." }
}
```

### `activity`

Доступные поля:

```text
searches
listingViews
uniqueListings
favoritesAdded
chatsStarted
chatsWithPurchase
listingsCreated
listingsPublished
purchasesCompleted
salesCompleted
activeDays
```

Ограничения:

- `uniqueListings <= listingViews`;
- `chatsWithPurchase <= chatsStarted`;
- итоговые метрики должны проходить доменную validation;
- recap должен иметь минимум 5 `TotalEvents`.

`RepeatedViews` не задаётся напрямую:

```text
repeatedViews = listingViews - uniqueListings
```

### `categories`

Пример:

```json
{
  "code": "books",
  "title": "Книги",
  "weight": 70,
  "shareable": true,
  "views": 55,
  "favoritesAdded": 9,
  "purchasesCompleted": 3
}
```

`weight` и явная category activity решают **разные задачи**:

- `weight` участвует в выборе `TopCategory` и приблизительном `TopCategoryViews`;
- `views` / `favoritesAdded` / `purchasesCompleted` формируют `CategoryActivities` и используются для тематических ачивок.

Если у категории заданы только `weight/title/code`, она может стать top category, но сама по себе не даст тематическую ачивку.

Правила для `weight`:

- у каждой категории weight > 0;
- сумма weights в сценарии должна быть ровно `100`;
- при равном weight top category выбирается по `code` в лексикографическом порядке.

Встроенный каталог находится в `internal/recap/analytics/categories.go` и синхронизирован с `seeds/categories.json`.

### `months`

Пример:

```json
[
  { "month": 9, "weight": 25 },
  { "month": 10, "weight": 40 },
  { "month": 11, "weight": 35 }
]
```

Правила:

- `month` от `1` до `12`;
- month не повторяется;
- weight > 0;
- сумма weights ровно `100`;
- максимальный weight определяет `MostActiveMonth`;
- при равенстве выбирается меньший номер месяца.

### `actionableState`

Пример для черновика:

```json
{
  "currentDrafts": 2,
  "draftListingId": "981a4c04-8057-537f-afac-794c49192178",
  "hasEverPublishedListing": true
}
```

Поддерживаемые поля:

```text
currentDrafts + draftListingId
openDialogs + openDialogId
activeListings + activeListingId
favoritesCount
hasSavedSearchForTopCategory
lastPurchasedListingId
hasEverPublishedListing
```

Если счётчик означает адресуемую работу (`currentDrafts`, `openDialogs`, `activeListings`), для CTA должен присутствовать валидный representative UUID. Тестовый `ActionStateStorage` автоматически подставляет `capturedAt = asOf`, если в JSON время не указано.

## 7. `expected/recaps.json` против golden

`testdata/expected/recaps.json` — короткая семантическая спецификация. Она проверяет то, что удобно читать человеку:

- rules version;
- behavior code;
- список achievement codes в порядке выдачи;
- next action code;
- top category;
- active month;
- точную последовательность card types.

Golden-файл — полная сериализация внутреннего recap. Он ловит более тонкие изменения:

- тексты и reasons;
- evidence;
- payload карточек;
- action targets;
- public share payload;
- rules digest;
- нормализованные metrics.

Поэтому при намеренном изменении логики обычно нужно обновить **оба уровня**: сначала человекочитаемый expected contract, затем golden.

## 8. Privacy-профили

### `private-style-hunter`

Top category — `beauty_cosmetics`, а тематическая ачивка — `STYLE_ICON`.

Этот профиль проверяет сразу два ограничения:

- `STYLE_ICON` остаётся в приватном recap, но не может быть выбран как public achievement;
- `beauty_cosmetics` помечена в trusted category catalogue как `Shareable=false`, поэтому не должна попасть в `ShareCard.TopCategory`, даже если входные данные пытаются разрешить share.

### `pet-threshold-buyer`

Профиль ровно достигает decisive-buyer thresholds и получает `CARING_OWNER`. Категория `pets` также privacy-sensitive и не должна публиковаться как top category.

Эти сценарии особенно полезны при изменениях в `presentation/share`, каталоге категорий и privacy policy.

## 9. Граничные и комбинированные профили

- `pet-threshold-buyer` — значения ровно на минимальных порогах `DECISIVE_BUYER`.
- `steady-buyer` — `purchases_completed >= 3`, поэтому есть `DEAL_CLOSER`, но слишком мало `chats_with_purchase` для `QUICK_DECISION` и decisive behavior.
- `seller-buyer-hybrid` — конфликт нескольких behavior rules + balanced seller/buyer portfolio.
- `recommendation-newcomer` — минимально допустимое число событий, нет top category и ни одного достижения.
- `listing-restart` — исторический starting-seller signal без существующего черновика и без прошлых публикаций, поэтому CTA именно `CREATE_FIRST_LISTING`, а не `FINISH_DRAFT`/`CREATE_LISTING`.
- `maker-with-draft` — показывает, что текущий адресуемый draft имеет больший приоритет, чем favorites/category fallback.

## 10. Как добавить новый тестовый профиль

1. Добавить профиль в `seeds/profiles.json` с новым уникальным UUID и уникальным `code`.
2. Добавить avatar-заглушку в `static/avatars/<code>.png`.
3. Добавить сценарий с тем же `profileCode` в `seeds/scenarios.json`.
4. Убедиться, что category/month weights суммируются до `100` и все UUID объектов уникальны.
5. Добавить ожидаемый результат в `testdata/expected/recaps.json`.
6. Создать/обновить `testdata/golden/<code>.json`.
7. Запустить seed test и общий test suite.

Для первого создания golden можно использовать:

```bash
UPDATE_GOLDEN=1 go test -run TestSeedProfilesGenerateExpectedRecaps ./internal/recap/application
```

После этого обязательно повторно запустить тест **без** `UPDATE_GOLDEN`, чтобы убедиться, что golden стабилен.

## 11. Как изменить существующий профиль

Если вы меняете метрики или actionable state, заранее определите, какое продуктово ожидаемое поведение должно измениться. Затем синхронно обновите:

```text
seeds/scenarios.json
      ↓
testdata/expected/recaps.json
      ↓
testdata/golden/<profile>.json
```

Если меняется только имя/описание/avatar, обычно достаточно `seeds/profiles.json` + avatar + golden, потому что профиль целиком хранится внутри recap.

Если изменение ruleset меняет многие профили, сначала обновите [`recap.md`](recap.md) и `expected/recaps.json`, а golden обновляйте последним. Так проще отличить намеренное изменение бизнес-контракта от случайного массового diff.

## 12. Тесты, которые защищают каталог

Главный интеграционный тест:

```text
internal/recap/application/seed_profiles_test.go
```

Он проверяет:

- валидность и уникальность профилей;
- отсутствие повторного использования UUID;
- наличие локальных avatar PNG;
- соответствие количества profiles/scenarios/expected;
- генерацию через реальный `application.Service` + `engine`;
- behavior, achievements, action, top category, active month и card sequence;
- полное совпадение с golden;
- сохранение результата;
- покрытие всех behavior и NextAction codes.

Дополнительно тематические и privacy-инварианты защищены тестами в `internal/recap/achievement`, `internal/recap/presentation/share`, `internal/recap/engine` и `internal/architecture`.

## 13. Быстрая проверка

```bash
go test -run TestSeedProfilesGenerateExpectedRecaps ./internal/recap/application
go test -run TestSeedCatalogueCoversAllBehaviorsActionsAndCoreAchievements ./internal/recap/application
go test -count=1 ./...
```

Для полной информации о самой бизнес-логике recap см. [`recap.md`](recap.md).
