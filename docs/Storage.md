# Storage

## 1. Назначение

Storage layer предоставляет данные, необходимые `application.Service`, и сохраняет сгенерированные Recap.

Application layer взаимодействует с хранилищем не напрямую через SQL или конкретную реализацию, а через набор портов.

```text
application.Service
       |
       +--> ProfileStorage
       +--> AnalyticsStorage
       +--> ActionStateStorage
       +--> RecapStorage
```

Поддерживаются две реализации:

```text
internal/storage/memory/
internal/storage/clickhouse/
```

---

## 2. Application ports

Порты определены в:

```text
internal/recap/application/ports.go
```

### ProfileStorage

```go
type ProfileStorage interface {
    ListProfiles(...)
    GetProfile(...)
    GetProfileByCode(...)
}
```

Отвечает за каталог профилей.

Transport обращается к пользователю по `profile_code`, тогда как UUID используется внутри backend.

---

### AnalyticsStorage

```go
type AnalyticsStorage interface {
    CalculateMetrics(...)
}
```

Возвращает годовые агрегированные метрики пользователя за `RecapPeriod`.

Application layer не знает, были ли метрики:

- заранее рассчитаны;
- получены из cache;
- агрегированы из raw events.

---

### ActionStateStorage

```go
type ActionStateStorage interface {
    GetActionableState(...)
}
```

Возвращает point-in-time snapshot текущего адресуемого состояния.

Примеры:

- текущий draft;
- открытый dialog;
- активное listing;
- favorites;
- last purchased listing.

`ActionableState` не является частью исторических годовых metrics.

---

### RecapStorage

```go
type RecapStorage interface {
    GetRecapByKey(...)
    CreateRecapIfAbsent(...)
    GetRecap(...)
    GetRecapByShareID(...)
}
```

Отвечает за persistence immutable recap.

Storage должен позволять искать Recap по:

- idempotency key;
- internal recap ID;
- public share ID.

---

## 3. Memory storage

Реализация:

```text
internal/storage/memory/
```

Используется для:

- локальной разработки;
- demo deployment;
- Render demo;
- unit/integration scenarios без ClickHouse.

Memory storage загружает:

```text
seeds/profiles.json
seeds/scenarios.json
```

при запуске процесса.

---

## 4. Структура memory store

`memory.Store` содержит несколько индексов:

```text
profiles
profilesByID
profilesByCode

metrics

actionStates

recapsByKey
recapsByID
recapsByShare
```

Схематично:

```text
Profile code -----> Profile
Profile UUID -----> Profile

(Profile UUID, year)
        |
        v
      Metrics

Profile UUID
        |
        v
 ActionableState

RecapKey -------> Recap
Recap ID -------> Recap
Share ID ------> Recap
```

Доступ синхронизирован через `sync.RWMutex`.

---

## 5. Lifetime memory storage

Memory storage не является persistent.

```text
process started
      |
      v
load seeds
      |
      v
generate recaps
      |
      v
memory only
      |
process stopped
      |
      v
generated recaps lost
```

После restart:

- profiles и scenarios снова загружаются из seed;
- ранее сгенерированные recaps исчезают;
- новые Recap ID и Share ID могут быть созданы заново.

Поэтому memory backend подходит для demo, но не для production persistence.

---

## 6. ClickHouse storage

Реализация:

```text
internal/storage/clickhouse/
```

Главный adapter:

```go
clickhouse.Repo
```

Он одновременно реализует:

```text
ProfileStorage
AnalyticsStorage
ActionStateStorage
RecapStorage
SeedStorage
```

Подключение выполняется через:

```text
CLICKHOUSE_DSN
```

После подключения backend вызывает:

```text
EnsureSchema()
```

---

## 7. Runtime schema

Основной runtime schema contract находится в:

```text
internal/storage/clickhouse/client.go
```

`EnsureSchema()` создаёт необходимые таблицы через:

```sql
CREATE TABLE IF NOT EXISTS
```

Поэтому повторный запуск приложения безопасен.

Используются таблицы:

```text
profiles
events
annual_metrics
actionable_state
recaps
```

---

## 8. `profiles`

Хранит demo/user profiles.

Основные поля:

```text
id
code
display_name
description
avatar_url
updated_at
```

Engine:

```text
ReplacingMergeTree(updated_at)
```

Чтение выполняется с `FINAL`.

Это позволяет повторно seed'ить один и тот же profile catalogue.

---

## 9. `events`

`events` — источник истины для годовой активности пользователя.

Пример данных:

```text
profile_id
event_type
occurred_at
category
ad_id
dialog_id
```

Engine:

```text
MergeTree
```

Partition:

```text
toYYYYMM(occurred_at)
```

Order:

```text
(profile_id, occurred_at)
```

TTL:

```text
occurred_at + 3 years
```

Готовый `Metrics` не является основным источником истины.

Flow:

```text
events
   |
   v
AnalyticsStorage.CalculateMetrics
   |
   v
analytics.AggregateEvents
   |
   v
Metrics
```

---

## 10. Annual metrics cache

Чтобы каждый запрос не агрегировал все события заново, используется:

```text
annual_metrics
```

Она содержит:

```text
profile_id
year
metrics
event_count
updated_at
```

`metrics` хранится сериализованным JSON.

`event_count` является freshness marker.

---

## 11. Cache-aside algorithm

Алгоритм `CalculateMetrics`:

```text
count current events
       |
       v
liveCount == 0 ?
       |
       +-- yes --> metrics not found
       |
       v
read annual_metrics
       |
       v
cached.event_count == liveCount ?
       |
       +-- yes --> return cached Metrics
       |
       v
read raw events
       |
       v
AggregateEvents
       |
       v
write annual_metrics cache
       |
       v
return Metrics
```

Таким образом наличие cache row само по себе не означает, что cache актуален.

Если после первого расчёта появились новые events:

```text
cached event_count != live count
```

и metrics рассчитываются заново.

---

## 12. Ограничение cache invalidation

Текущий freshness mechanism сравнивает количество событий.

Он хорошо обнаруживает:

```text
+ новый event
- удалённый event
```

если меняется итоговый count.

Но он не является универсальной системой versioning содержимого событий.

Если данные существующего event будут изменены без изменения количества строк, одного `event_count` недостаточно для обнаружения изменения.

В текущем demo/event model предполагается append-oriented поток событий.

Если появится полноценный mutable event ingestion, freshness marker стоит заменить на более сильный механизм, например:

```text
max(updated_at)
event stream version
source offset
checksum
```

---

## 13. `actionable_state`

Хранит текущий point-in-time snapshot.

Поля domain model сериализуются в JSON:

```text
profile_id
state
updated_at
```

Engine:

```text
ReplacingMergeTree(updated_at)
```

При чтении используется:

```sql
FINAL
```

`CapturedAt` не читается как историческое время записи.

Storage устанавливает:

```text
CapturedAt = asOf
```

То есть это время, на которое application запросил snapshot.

---

## 14. `recaps`

Хранит полностью построенный immutable Recap.

Поля:

```text
id
share_id
profile_id
year
rules_version
rules_digest
recap
created_at
```

Сам domain `Recap` хранится сериализованным JSON в поле:

```text
recap
```

Engine:

```text
MergeTree
```

`ReplacingMergeTree` здесь намеренно не используется.

После генерации Recap не должен обновляться.

---

## 15. Idempotency key

Recap ищется по:

```text
profile_id
year
rules_version
rules_digest
```

То есть:

```text
RecapKey =
(
    ProfileID,
    Year,
    RulesVersion,
    RulesDigest
)
```

При одинаковом ключе application ожидает один логический Recap.

---

## 16. CreateRecapIfAbsent

Memory storage обеспечивает эту операцию атомарно внутри процесса:

```text
mutex
  |
  v
check key
  |
  +--> exists -> return existing
  |
  v
insert
```

ClickHouse implementation использует:

```text
GetRecapByKey
        |
        v
if not found
        |
        v
INSERT
```

То есть это:

```text
check-then-insert
```

а не database-level atomic upsert.

---

## 17. Ограничение ClickHouse idempotency

MergeTree не предоставляет обычный OLTP unique constraint на idempotency key.

Поэтому при нескольких параллельных backend writers возможна race:

```text
Writer A: check -> not found
Writer B: check -> not found

Writer A: insert
Writer B: insert
```

Для текущего single-writer demo deployment это принято как допустимое ограничение.

Для multi-instance production deployment потребовался бы отдельный concurrency boundary:

```text
distributed lock

или

OLTP database with UNIQUE constraint

или

dedicated idempotency service
```

Это важно учитывать перед горизонтальным масштабированием backend.

---

## 18. Stored Recap считается недоверенным

Storage не считается источником бизнес-истины для derived projections.

После чтения Recap application вызывает:

```text
engine.ValidateStored
```

И проверяет соответствие сохранённого объекта текущим structural/integrity invariants.

Для idempotent lookup также проверяется:

```text
stored RecapKey == requested RecapKey
```

Для lookup по IDs:

```text
stored recap ID == requested recap ID
stored share ID == requested share ID
```

---

## 19. Demo bootstrap в ClickHouse

Если:

```text
SEED_DEMO_DATA=true
```

после `EnsureSchema()` выполняется:

```text
bootstrap.LoadDemoData
```

Pipeline:

```text
profiles.json
     |
     v
UpsertProfiles

scenarios.json
     |
     v
generate raw events
     |
     v
InsertEvents

actionableState
     |
     v
PutActionableState
```

Важно:

> Bootstrap не записывает вручную подготовленный `Metrics`.

Он создаёт события, из которых production-like analytics path самостоятельно рассчитывает metrics.

---

## 20. `clickhouse/init`

Docker Compose монтирует:

```text
clickhouse/init/
```

в:

```text
/docker-entrypoint-initdb.d
```

В текущем проекте runtime source of truth для таблиц приложения — `EnsureSchema()`.

Файл:

```text
001_schema.sql
```

создаёт raw `events`.

Файл:

```text
002_recap_cache.sql
```

содержит более ранние таблицы:

```text
recap_cards
recap_summary
```

Текущий application/storage layer их не использует.

Их следует считать legacy/prototype schema, пока код не начнёт обращаться к ним снова.

В дальнейшем желательно либо:

1. удалить legacy schema;
2. либо перенести всю актуальную schema в versioned SQL migrations.

Сейчас schema распределена между SQL init и `EnsureSchema()`, что увеличивает риск рассинхронизации.

---

## 21. Выбор storage backend

### Memory

```text
STORAGE_BACKEND=memory
```

Подходит для:

- demo;
- Render;
- локального UI;
- тестирования.

Плюсы:

- не требуется внешняя БД;
- быстрый startup;
- seed catalogue загружается сразу.

Минусы:

- нет persistence;
- данные теряются после restart;
- нельзя делить состояние между несколькими instances.

---

### ClickHouse

```text
STORAGE_BACKEND=clickhouse
```

Подходит для:

- event analytics;
- Docker Compose;
- persistence demo;
- больших объёмов событий.

Плюсы:

- эффективная аналитика событий;
- persistent recaps;
- cache годовых metrics.

Ограничения:

- idempotency insert не является строго атомарным между несколькими writers;
- JSON domain blobs хуже подходят для ad-hoc SQL analytics;
- schema management пока не оформлен как полноценные migrations.

---

## 22. Принцип выбора технологии

ClickHouse хорошо подходит для:

```text
events
analytics
annual aggregates
```

Но immutable recap с сильным уникальным idempotency key по своей природе ближе к OLTP entity.

Если проект станет multi-instance production service, возможна гибридная схема:

```text
ClickHouse
   |
   +--> events
   +--> analytics

PostgreSQL / OLTP store
   |
   +--> profiles
   +--> actionable state
   +--> recaps
   +--> idempotency
```

Для текущего масштаба единый ClickHouse adapter сохраняет проект проще.

---

## 23. Storage invariants

Storage implementations должны обеспечивать одинаковый application-level contract:

1. Profile code однозначно определяет профиль.
2. Profile UUID остаётся внутренним идентификатором.
3. Metrics относятся к конкретному `RecapPeriod`.
4. ActionableState является point-in-time snapshot.
5. Recap после записи не изменяется.
6. Lookup по share ID не возвращает другой Recap.
7. Lookup по internal ID не возвращает другой Recap.
8. Idempotency lookup проверяется через `RecapKey`.
9. Stored Recap после чтения проходит engine validation.
10. Storage не вычисляет Behavior, Achievement или NextAction.

---

## 24. Где менять storage

| Изменение | Файл/пакет |
|---|---|
| Storage interfaces | `internal/recap/application/ports.go` |
| Memory implementation | `internal/storage/memory/` |
| ClickHouse connection/schema | `internal/storage/clickhouse/client.go` |
| Profiles SQL | `internal/storage/clickhouse/profiles.go` |
| Event/metrics analytics | `internal/storage/clickhouse/analytics.go` |
| ActionableState | `internal/storage/clickhouse/current_state.go` |
| Recap persistence | `internal/storage/clickhouse/recaps.go` |
| Demo seeding | `internal/bootstrap/` |
| Docker ClickHouse initialization | `clickhouse/init/` |