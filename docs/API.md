# API

## 1. Назначение

Backend предоставляет frontend-facing API для Year Recap через ConnectRPC.

Source of truth wire-контракта:

```text
proto/recap/v1/recap.proto
```

Generated Go-код находится в:

```text
gen/go/recap/v1/
```

Generated frontend TypeScript contract:

```text
frontend/src/gen/recap/v1/
```

Frontend должен использовать generated contract, а не вручную поддерживаемые копии типов.

---

## 2. Transport

API использует:

```text
Protocol Buffers
+
ConnectRPC
```

Transport implementation:

```text
internal/transport/connect/
```

HTTP server:

```text
internal/server/
```

Frontend client:

```text
frontend/src/shared/api/recap-api.ts
frontend/src/shared/api/connect-transport.ts
```

---

## 3. Service

В protobuf объявлен:

```protobuf
service RecapService
```

Доступны четыре RPC:

| RPC | Назначение |
|---|---|
| `ListProfiles` | Получить demo-каталог профилей |
| `GenerateRecap` | Создать или вернуть существующий immutable recap |
| `GetRecap` | Получить recap по `profile_code + year` |
| `GetPublicShare` | Получить безопасную публичную share projection |

---

## 4. HTTP paths

Generated Connect handler использует service path:

```text
/recap.v1.RecapService/
```

Например:

```text
/recap.v1.RecapService/ListProfiles
/recap.v1.RecapService/GenerateRecap
/recap.v1.RecapService/GetRecap
/recap.v1.RecapService/GetPublicShare
```

Backend также публикует те же endpoints через `/api`:

```text
/api/recap.v1.RecapService/ListProfiles
/api/recap.v1.RecapService/GenerateRecap
...
```

---

## 5. Frontend base URL

В development frontend обращается напрямую к Go backend:

```text
http://localhost:8080
```

В production single-service build:

```text
/api
```

Таким образом frontend configuration выглядит концептуально так:

```text
DEV
http://localhost:8080
        |
        v
/recap.v1.RecapService/...

PROD
same origin
        |
        v
/api/recap.v1.RecapService/...
```

---

## 6. ListProfiles

Request:

```protobuf
message ListProfilesRequest {}
```

Response:

```protobuf
message ListProfilesResponse {
  repeated Profile profiles = 1;
}
```

Profile:

```text
name
description
avatar_url
profile_code
```

В публичный wire-profile не входит internal Profile UUID.

Пример логического response:

```json
{
  "profiles": [
    {
      "name": "Алексей",
      "description": "Часто ищет технику...",
      "avatarUrl": "/avatars/active-buyer.png",
      "profileCode": "active-buyer"
    }
  ]
}
```

---

## 7. GenerateRecap

Request:

```protobuf
message GenerateRecapRequest {
  string profile_code = 1;
  uint32 year = 2;
}
```

Пример:

```json
{
  "profileCode": "active-buyer",
  "year": 2025
}
```

Transport:

1. проверяет request;
2. проверяет `profile_code`;
3. получает профиль через `GetProfileByCode`;
4. передаёт internal UUID в application service;
5. вызывает `Service.Generate`.

---

## 8. Generate semantics

`GenerateRecap` является idempotent относительно rules identity.

Conceptually:

```text
(profile, year, rules identity)
             |
             v
        existing?
        /       \
      yes       no
       |         |
       v         v
 return old    build
                |
                v
              store
```

Повторный вызов для одинакового RecapKey должен вернуть существующий Recap, а не создавать независимую новую версию.

---

## 9. Условия генерации

Backend валидирует requested year.

Recap можно строить только за полностью завершённый календарный год UTC.

Основные ошибки:

```text
year == 0
future year
current unfinished year
insufficient activity
```

Конкретные продуктовые условия описаны в `recap.md`.

---

## 10. RecapResponse

Оба RPC:

```text
GenerateRecap
GetRecap
```

возвращают:

```protobuf
message RecapResponse {
  Profile profile = 1;
  Recap recap = 2;
}
```

`Recap` содержит:

```text
id
year
rule_version
cards
achievements
next_action
```

---

## 11. Recap ID

Поле:

```text
Recap.id
```

является ID приватного recap.

Оно может использоваться внутри приватного приложения, но не является public share identifier.

Для публичного доступа существует отдельный:

```text
share_id
```

Эти идентификаторы выполняют разные функции.

---

## 12. GetRecap

Request:

```protobuf
message GetRecapRequest {
  string profile_code = 1;
  uint32 year = 2;
}
```

На wire-уровне Recap адресуется через:

```text
profile_code + year
```

а не через internal profile UUID.

### Текущее поведение

В текущей реализации `GetRecap` вызывает тот же:

```text
application.Service.Generate
```

что и `GenerateRecap`.

То есть сейчас `GetRecap` фактически имеет semantics:

```text
get existing
OR
generate if absent
```

Это возможно благодаря idempotent generation.

В коде transport layer явно отмечено, что в будущем может быть добавлен отдельный read-only service method по `(profile, year)`.

После этого `GetRecap` сможет иметь строгую семантику:

```text
GET existing only
```

без создания отсутствующего recap.

---

## 13. Recap cards

Recap содержит ordered collection:

```text
Recap.cards
```

Каждая карточка имеет:

```text
id
type
position
title
description
explanation
shareable
payload
```

Payload является protobuf `oneof`.

Это означает, что карточка определённого типа получает только соответствующий payload.

Например:

```text
YEAR_ACTIVITY
     |
     v
YearActivityPayload
```

а:

```text
NEXT_ACTION
     |
     v
NextActionPayload
```

Frontend не должен читать payload, принадлежащий другому card type.

---

## 14. Card types

Текущий contract содержит:

```text
INTRO
YEAR_ACTIVITY
TOP_CATEGORY
ACTIVE_MONTH
BEHAVIOR
ACHIEVEMENT
MISSED_OPPORTUNITY
NEXT_ACTION
SHARE
```

Порядок карточек задаёт backend.

Frontend должен использовать:

```text
position
```

и не пытаться самостоятельно вычислять story order.

---

## 15. Behavior

Wire contract передаёт:

```text
BehaviorCode
BehaviorEvidence
```

Поддерживаемые Behavior codes:

```text
ACTIVE_SELLER
STARTING_SELLER
DECISIVE_BUYER
FIND_HUNTER
RESEARCHER
UNIVERSAL_USER
```

Frontend использует code как presentation key.

Правила выбора behavior принадлежат backend.

---

## 16. Achievements

Achievement содержит:

```text
code
title
reason
shareable
```

Wire enum может расширяться.

Frontend должен корректно обрабатывать неизвестный будущий код или обновляться вместе с protobuf schema.

Бизнес-условия получения achievement описываются не в API documentation, а в `recap.md`.

---

## 17. NextAction

`NextAction` содержит:

```text
code
title
description
explanation
button_text
target
```

Target является typed `oneof`.

Возможные destinations:

```text
listing
dialog
category
search
route
```

Пример conceptual target:

```json
{
  "code": "ACTION_CODE_OPEN_FAVORITES",
  "target": {
    "route": {
      "path": "/favorites"
    }
  }
}
```

Frontend должен выполнять именно переданный backend target.

---

## 18. ActionTarget invariant

В корректном NextAction должен быть заполнен ровно один destination.

Нельзя одновременно передать:

```text
listing + route
```

или:

```text
dialog + category
```

Transport mapper считает подобную domain projection некорректной.

---

## 19. CREATE_FIRST_LISTING

В protobuf enum присутствует:

```text
ACTION_CODE_CREATE_FIRST_LISTING
```

Он помечен:

```text
deprecated = true
```

для wire compatibility.

При этом текущая domain logic всё ещё может формировать `CREATE_FIRST_LISTING` для соответствующего starting-seller сценария.

Поэтому frontend пока должен поддерживать этот код.

Удалять enum можно только после синхронного изменения:

- domain rules;
- transport mapper;
- frontend;
- golden/test scenarios;
- protobuf compatibility policy.

---

## 20. GetPublicShare

Request:

```protobuf
message GetPublicShareRequest {
  string share_id = 1;
}
```

`share_id` должен быть canonical lowercase UUID.

Transport отклоняет:

- invalid UUID;
- nil UUID;
- UUID не в canonical lowercase representation.

---

## 21. PublicShare

Response:

```protobuf
message GetPublicShareResponse {
  PublicShare share = 1;
}
```

Public payload намеренно минимален:

```text
share_id
year
behavior_title
achievement_title?
top_category?
```

В public response отсутствуют:

```text
profile UUID
internal recap ID
raw metrics
ActionableState
NextAction
ActionTarget
полный achievement list
```

Public API является отдельной privacy projection, а не урезанным сериализованным приватным Recap.

---

## 22. Privacy flow

```text
Stored Recap
     |
     v
ValidateStored
     |
     v
PublicProjection
     |
     v
PublicShare protobuf
```

Transport не выбирает вручную, какие поля безопасны.

Privacy logic принадлежит recap engine/presentation layer.

---

## 23. Error mapping

Application errors преобразуются в ConnectRPC codes.

### `INVALID_ARGUMENT`

Используется для:

```text
invalid profile ID
invalid recap ID
invalid share ID
invalid year
invalid canonical UUID
missing request parameters
```

---

### `FAILED_PRECONDITION`

Используется для:

```text
year not complete
not enough activity
```

---

### `NOT_FOUND`

Используется для:

```text
profile code not found
recap not found
share not found
```

---

### `DATA_LOSS`

Используется, когда storage/domain data нарушают contract:

```text
invalid stored Profile
invalid Metrics
invalid ActionableState
invalid Recap

profile ID mismatch
recap ID mismatch
share ID mismatch
RecapKey mismatch

invalid transport projection
```

Это означает:

> данные были получены, но backend не может безопасно им доверять.

---

### `INTERNAL`

Все остальные неожиданные ошибки преобразуются в:

```text
INTERNAL
```

---

### Context errors

```text
context.Canceled
    ↓
CANCELED

context.DeadlineExceeded
    ↓
DEADLINE_EXCEEDED
```

---

## 24. CORS

HTTP server поддерживает CORS для configured origins.

Переменная:

```text
CORS_ALLOWED_ORIGINS
```

Default:

```text
http://localhost:3000
http://localhost:5173
```

Same-origin request разрешается без отдельного allow-list lookup.

Для разрешённого cross-origin request backend добавляет:

```text
Access-Control-Allow-Origin
Access-Control-Allow-Credentials
```

Также разрешены Connect/gRPC-web headers.

---

## 25. Health endpoint

Backend предоставляет:

```text
GET /health
HEAD /health
```

Response GET:

```json
{
  "status": "ok"
}
```

Другие HTTP methods получают:

```text
405 Method Not Allowed
```

Endpoint используется, например, Render health check.

---

## 26. Avatar endpoint

Backend раздаёт:

```text
/avatars/*
```

из:

```text
STATIC_DIR/avatars
```

Для профиля:

```json
{
  "avatarUrl": "/avatars/active-buyer.png"
}
```

backend ожидает файл:

```text
<STATIC_DIR>/avatars/active-buyer.png
```

Локально:

```text
frontend/public/avatars/active-buyer.png
```

В production:

```text
/app/web/avatars/active-buyer.png
```

---

## 27. SPA routing

Если установлен:

```text
FRONTEND_DIR
```

Go server также обслуживает React build.

Для реального файла:

```text
/app/web/assets/index.js
```

возвращается файл.

Для client-side route:

```text
/recap/active-buyer
```

если физического файла нет, backend возвращает:

```text
index.html
```

после чего route обрабатывает React Router.

Неизвестные URLs под:

```text
/api/
```

не fallback'ятся в React SPA.

Они возвращают HTTP 404.

---

## 28. Frontend API client

Frontend создаёт generated Connect client:

```text
createClient(
    RecapService,
    createRecapTransport(API_BASE_URL)
)
```

Основные wrapper functions:

```text
listProfiles()
generateRecap(profileCode)
getRecap(profileCode)
getPublicShare(shareId)
```

Для текущего demo frontend фиксированно использует:

```text
RECAP_YEAR = 2025
```

Это UI/demo decision, а не ограничение protobuf contract.

Сам API принимает любой `uint32 year`, после чего backend проверяет допустимость года.

---

## 29. Изменение API contract

При изменении `.proto` необходимо:

1. изменить `proto/recap/v1/recap.proto`;
2. перегенерировать Go code;
3. перегенерировать frontend protobuf code;
4. обновить transport mapper;
5. обновить frontend mapper;
6. обновить tests;
7. проверить backward compatibility.

Нельзя вручную изменять generated files как source of truth.

---

## 30. API invariants

1. Frontend адресует профиль через `profile_code`.
2. Internal Profile UUID не публикуется.
3. Backend является владельцем business rules.
4. Cards возвращаются в готовом порядке.
5. ActionTarget содержит ровно один destination.
6. Public share доступен только через `share_id`.
7. Public share не содержит ActionableState.
8. Public share не содержит внутренних IDs.
9. Transport errors имеют стабильную ConnectRPC classification.
10. Stored domain objects валидируются перед возвратом.
11. Unknown `/api/*` не превращается в React page.
12. Wire contract изменяется через protobuf, а не ad-hoc JSON DTO.

---

## 31. Текущие технические ограничения

### `GetRecap` не является строго read-only

Сейчас он использует idempotent `Generate`.

В будущем желательно добавить application method вида:

```text
GetByProfileAndYear
```

и сделать:

```text
GenerateRecap -> get-or-create
GetRecap      -> get-only
```

---

### Deprecated action всё ещё используется domain logic

`CREATE_FIRST_LISTING` помечен deprecated в protobuf, но пока остаётся активным продуктовым сценарием.

До миграции frontend обязан поддерживать его.

---

### Versioning API

Сейчас version namespace задаётся protobuf package:

```text
recap.v1
```

Breaking wire changes должны либо сохранять protobuf compatibility, либо приводить к новому API version/package.

---

## 32. Связанные документы

```text
recap.md
```

Подробно описывает:

- Recap semantics;
- Behavior;
- Achievements;
- NextAction;
- cards;
- privacy;
- rules identity.

```text
test_profiles.md
```

Описывает:

- demo profiles;
- scenario coverage;
- expected fixtures;
- golden tests.

```text
architecture.md
```

Описывает слои приложения и направления зависимостей.

```text
storage.md
```

Описывает memory/ClickHouse persistence и storage contracts.