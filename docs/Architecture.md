# Architecture

## 1. Назначение документа

Этот документ описывает архитектуру приложения Year Recap: основные слои, направления зависимостей, поток данных от HTTP-запроса до построения recap и границы между бизнес-логикой, transport-слоем и хранилищем.

Подробные продуктовые правила Recap описаны отдельно в `recap.md`.

Основная архитектурная идея проекта:

```text
Frontend
   |
   v
ConnectRPC / HTTP
   |
   v
Application Service
   |
   v
Recap Engine
   |
   +--> Analytics
   +--> Behavior
   +--> Achievements
   +--> NextAction
   +--> Presentation
   |
   v
Storage ports
   |
   +--> Memory
   +--> ClickHouse
```

Backend является владельцем всей бизнес-логики Recap.

Frontend не пересчитывает behavior, achievements или NextAction самостоятельно, а только отображает полученную от backend последовательность карточек.

---

## 2. Основные компоненты

Проект условно разделён на следующие части:

```text
cmd/api/
internal/
    bootstrap/
    config/
    recap/
    server/
    storage/
    transport/
proto/
frontend/
seeds/
testdata/
clickhouse/
```

### `cmd/api`

Точка входа backend-приложения.

`cmd/api/main.go` отвечает за:

- загрузку конфигурации;
- выбор storage backend;
- подключение к ClickHouse или загрузку memory storage;
- создание `application.Service`;
- создание HTTP handler;
- запуск HTTP-сервера;
- graceful shutdown по `SIGINT` и `SIGTERM`.

Бизнес-логика в `cmd/api` отсутствует.

---

## 3. Конфигурация

Конфигурация находится в:

```text
internal/config/
```

Основные параметры задаются через environment variables:

```text
API_ADDRESS
HTTP_ADDR
PORT

STORAGE_BACKEND
CLICKHOUSE_DSN

SEED_DEMO_DATA
PROFILES_PATH
SCENARIOS_PATH

STATIC_DIR
FRONTEND_DIR

CORS_ALLOWED_ORIGINS
SHUTDOWN_TIMEOUT
```

Приоритет адреса сервера:

```text
API_ADDRESS
    ↓
HTTP_ADDR
    ↓
PORT
    ↓
:8080
```

Поддерживаются два storage backend:

```text
memory
clickhouse
```

---

## 4. HTTP и transport layer

HTTP routing находится в:

```text
internal/server/
```

ConnectRPC transport находится в:

```text
internal/transport/connect/
```

HTTP server выполняет несколько независимых задач:

```text
/
├── ConnectRPC API
├── /api/...         ConnectRPC API для production frontend
├── /health          health check
├── /avatars/...     статические avatar-файлы
└── React SPA        если задан FRONTEND_DIR
```

### Connect endpoints

Generated Connect handler доступен напрямую:

```text
/recap.v1.RecapService/...
```

и через production-prefix:

```text
/api/recap.v1.RecapService/...
```

Prefix `/api` используется frontend'ом в single-service production deployment.

---

## 5. Transport boundary

Transport layer отвечает только за задачи wire-протокола:

- чтение protobuf request;
- проверку обязательных параметров;
- преобразование `profile_code` во внутренний профиль;
- проверку UUID;
- вызов application service;
- преобразование domain model в protobuf;
- преобразование application errors в ConnectRPC status codes.

Transport не принимает решений о:

- Behavior;
- Achievements;
- NextAction;
- privacy allow-list;
- порядке карточек;
- idempotency rules.

Эти решения принадлежат application/domain слоям.

---

## 6. Application layer

Основной orchestration layer находится в:

```text
internal/recap/application/
```

Главный тип:

```go
application.Service
```

Он зависит от четырёх портов:

```text
ProfileStorage
AnalyticsStorage
ActionStateStorage
RecapStorage
```

Application layer не знает, используется ли ClickHouse, memory storage или другой adapter.

Схема зависимости:

```text
                 +------------------+
                 | application      |
                 | Service          |
                 +--------+---------+
                          |
             +------------+------------+
             |            |            |
             v            v            v
         Profile      Analytics    ActionState
         Storage      Storage      Storage
                          |
                          v
                      RecapStorage
```

Реализации портов передаются в `NewService`.

---

## 7. Generation flow

Основной сценарий — генерация recap.

```text
GenerateRecap
     |
     v
resolve profile_code
     |
     v
application.Service.Generate
     |
     +--> validate year
     |
     +--> build RecapKey
     |
     +--> check existing recap
     |
     +--> load Profile
     |
     +--> CalculateMetrics
     |
     +--> GetActionableState
     |
     +--> generate recap_id
     |
     +--> generate share_id
     |
     v
Recap Engine
     |
     +--> Behavior
     +--> Achievements
     +--> NextAction
     +--> Story Cards
     +--> Share Projection
     |
     v
CreateRecapIfAbsent
     |
     v
ValidateStored
     |
     v
protobuf response
```

Ключевой принцип:

> Business derivation завершается до persistence.

Хранилище не решает, какой behavior или NextAction должен получить пользователь.

---

## 8. Domain и Recap Engine

Бизнес-логика находится внутри:

```text
internal/recap/
```

Основные области:

```text
analytics/
behavior/
achievement/
nextaction/
presentation/
ruleset/
validation/
model/
engine/
application/
```

### `model`

Определяет основные domain entities:

- Profile;
- Metrics;
- Event;
- ActionableState;
- Recap;
- RecapCard;
- Achievement;
- NextAction;
- ShareCard.

Этот пакет не зависит от storage или HTTP.

### `analytics`

Отвечает за:

- построение периода;
- агрегацию событий;
- derived metrics;
- annual metrics.

### `behavior`

Определяет пользовательскую persona/behavior.

### `achievement`

Определяет заработанные достижения и итоговый portfolio.

### `nextaction`

Выбирает одно исполнимое действие.

### `presentation`

Строит story cards и публичную share projection.

### `ruleset`

Содержит конфигурацию продуктовых правил:

- thresholds;
- priorities;
- algorithm identity;
- privacy version.

### `validation`

Проверяет structural/domain invariants.

### `engine`

Координирует domain-компоненты и строит итоговый immutable `Recap`.

---

## 9. Правила зависимостей

Domain packages не должны зависеть от инфраструктуры.

Запрещённое направление:

```text
recap
  |
  X
  v
storage
```

Правильное направление:

```text
storage
  |
  v
application ports / model
```

Также domain packages не должны зависеть от application orchestration:

```text
behavior ------X-----> application
achievement ---X-----> application
analytics -----X-----> application
```

`presentation` является downstream-слоем и не должен становиться dependency для расчётных пакетов.

Архитектурные ограничения защищены тестами в:

```text
internal/architecture/
```

---

## 10. Storage adapters

Инфраструктурные адаптеры находятся в:

```text
internal/storage/
```

Сейчас есть:

```text
memory/
clickhouse/
```

Оба реализуют application ports.

Это позволяет запускать один и тот же application service поверх разных хранилищ:

```text
application.Service
        |
        +-------- Memory Store
        |
        +-------- ClickHouse Repo
```

Подробнее устройство persistence описано в `storage.md`.

---

## 11. Seed и bootstrap

Demo-сценарии находятся в:

```text
seeds/profiles.json
seeds/scenarios.json
```

Для `memory` storage они загружаются непосредственно при старте.

Для ClickHouse используется:

```text
internal/bootstrap/
```

Bootstrap:

1. читает profiles;
2. читает scenarios;
3. создаёт raw events;
4. записывает profiles;
5. записывает events;
6. записывает ActionableState.

Важно:

```text
scenario
   ↓
raw events
   ↓
AnalyticsStorage.CalculateMetrics
   ↓
Metrics
```

Bootstrap намеренно не записывает готовые annual metrics как источник истины.

---

## 12. Frontend architecture boundary

Frontend находится в:

```text
frontend/
```

Он работает с generated protobuf contract через ConnectRPC client.

Типовой flow:

```text
React page
   |
   v
shared/api/recap-api.ts
   |
   v
ConnectRPC generated client
   |
   v
Go backend
```

После ответа backend:

```text
protobuf response
      |
      v
proto mapper
      |
      v
frontend Recap model
      |
      v
Recap Player
      |
      v
card renderer
```

Frontend не должен повторять backend rules.

Например, он не должен самостоятельно определять:

```text
if purchases >= 3 -> DECISIVE_BUYER
```

Он должен использовать уже готовый `BehaviorCode`, полученный от сервера.

---

## 13. Static files

Avatar-файлы находятся в:

```text
frontend/public/avatars/
```

Публичный URL имеет вид:

```text
/avatars/<profile-code>.png
```

Локально backend использует:

```text
STATIC_DIR=frontend/public
```

и раздаёт:

```text
frontend/public/avatars/
        ↓
     /avatars/
```

При frontend build Vite переносит содержимое `public` в `dist`:

```text
frontend/public/avatars
        ↓
npm run build
        ↓
frontend/dist/avatars
```

В production Docker:

```text
frontend/dist
     ↓
/app/web
```

и:

```text
STATIC_DIR=/app/web
FRONTEND_DIR=/app/web
```

Поэтому URL остаётся одинаковым:

```text
/avatars/<code>.png
```

---

## 14. Deployment modes

### Single-service production

Используется основной `Dockerfile`.

Build:

```text
Node builder
    ↓
frontend/dist

Go builder
    ↓
api binary

        ↓

runtime image
├── /usr/local/bin/api
├── /app/seeds
└── /app/web
```

Go process одновременно обслуживает:

- API;
- avatars;
- React SPA.

Эта схема используется Render deployment.

---

### Docker Compose

В `docker-compose.yml` используется другая topology:

```text
Browser
   |
   v
Nginx
   |
   +--> frontend
   |
   +--> Go API
             |
             v
         ClickHouse
```

Backend работает с:

```text
STORAGE_BACKEND=clickhouse
```

а frontend обслуживается отдельным nginx container.

---

## 15. Privacy boundary

Приватные данные Recap и public share рассматриваются как разные projections.

Внутренний recap может содержать:

- metrics;
- evidence;
- achievements;
- NextAction;
- ActionTarget;
- internal recap ID.

Public share формируется отдельно.

Публичный endpoint работает только через `share_id`.

Нельзя отдавать через public contract:

- profile UUID;
- actionable state;
- listing/dialog identifiers;
- raw metrics;
- полный список achievements;
- внутренний recap ID.

Privacy boundary дополнительно защищён architecture tests.

---

## 16. Architectural invariants

Основные invariants проекта:

1. Backend является владельцем Recap business rules.
2. Frontend только отображает результат.
3. Domain не зависит от storage.
4. Domain не зависит от HTTP/ConnectRPC.
5. Storage реализует application ports.
6. Recap после генерации считается immutable.
7. Public share строится отдельной projection.
8. `share_id` отличается от внутреннего `recap_id`.
9. ActionableState не входит в public protobuf contract.
10. Изменение rules должно приводить к изменению rules identity.
11. Persistence не должна самостоятельно выполнять business derivation.
12. Seed и production storage должны предоставлять одинаковый application contract.

---

## 17. Куда добавлять новую функциональность

| Задача | Место |
|---|---|
| Новый RPC | `proto/recap/v1/` + `internal/transport/connect/` |
| Новый environment parameter | `internal/config/` |
| Новый behavior | `internal/recap/behavior/` |
| Новое achievement | `internal/recap/achievement/` |
| Новый NextAction | `internal/recap/nextaction/` |
| Новая story-card | `internal/recap/presentation/` |
| Новое product rule | `internal/recap/ruleset/` |
| Новый structural invariant | `internal/recap/validation/` |
| Новый storage backend | `internal/storage/` |
| Изменение generation flow | `internal/recap/application/` или `engine/` |
| HTTP/static/CORS | `internal/server/` |
| Demo data | `seeds/` |
| Frontend API integration | `frontend/src/shared/api/` |

Главный принцип при расширении:

> Сначала определить слой, которому принадлежит решение, и не переносить бизнес-правила в transport, storage или frontend.