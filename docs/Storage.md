# Storage

## 1. Назначение

Storage предоставляет данные `application.Service` и хранит сгенерированные Recap — через набор портов (`internal/recap/application/ports.go`), а не через прямой SQL из бизнес-логики.

Единственная реализация: `internal/storage/clickhouse/` — постоянное аналитическое хранилище. Никакого in-memory demo-режима нет: приложение всегда работает через реальные события в ClickHouse, включая локальную разработку и unit-тесты бизнес-логики (которые используют независимые fake-реализации портов из `internal/recap/testkit/`, а не отдельный storage backend).

---

## 2. Application ports

```go
type ProfileStorage interface {
    ListProfiles(...)
    GetProfile(...)
    GetProfileByCode(...)
}

type AnalyticsStorage interface {
    CalculateMetrics(...) // годовые метрики за RecapPeriod
}

type ActionStateStorage interface {
    GetActionableState(...) // point-in-time snapshot текущего состояния
}

type RecapStorage interface {
    GetRecapByKey(...)      // idempotency lookup
    CreateRecapIfAbsent(...)
    GetRecap(...)           // по internal ID
    GetRecapByShareID(...)  // по public share ID
}
```

Ключевые нюансы:

- Transport обращается к профилю по `profile_code`, UUID — только внутри backend.
- `ActionableState` не часть исторических годовых metrics — это текущее (draft/dialog/favorites/last purchase), а не прошлогоднее.
- `AnalyticsStorage` не раскрывает, были ли метрики предрассчитаны, взяты из кэша или агрегированы на лету — это деталь реализации.

---

## 3. ClickHouse storage

`internal/storage/clickhouse.Repo` — единственный adapter, реализует все четыре порта выше плюс `bootstrap.SeedStorage`. Подключается через `CLICKHOUSE_DSN`, после чего backend вызывает `EnsureSchema()`.

`EnsureSchema()` (в `internal/storage/clickhouse/client.go`) идемпотентна:

- сначала прогоняет `CREATE TABLE IF NOT EXISTS` по всем таблицам — безопасно перезапускать;
- затем — небольшой набор `ALTER TABLE ... MODIFY TTL`, чтобы докатить schema-правки (например, изменённый TTL) и на уже существующую БД, которую `CREATE TABLE IF NOT EXISTS` не трогает.

---

## 4. Таблицы

| Таблица | Engine | Order / Partition | TTL | Назначение |
|---|---|---|---|---|
| `profiles` | `ReplacingMergeTree(updated_at)` | `ORDER BY id` | — | каталог профилей, читается с `FINAL`, безопасно re-seed'ить |
| `events` | `MergeTree` | `PARTITION BY toYYYYMM(occurred_at)`, `ORDER BY (profile_id, occurred_at)` | `occurred_at + 3 года` | источник истины годовой активности |
| `annual_metrics` | `ReplacingMergeTree(updated_at)` | `ORDER BY (profile_id, year)` | `updated_at + 3 года` | кэш агрегации над `events` — не должен переживать данные, кэшем которых является |
| `actionable_state` | `ReplacingMergeTree(updated_at)` | `ORDER BY profile_id` | нет (намеренно) | текущее состояние профиля, а не time-decaying лог — TTL по возрасту сломал бы валидный snapshot неактивного профиля |
| `recaps` | `MergeTree` (не Replacing — после генерации не обновляется) | `ORDER BY (profile_id, year, rules_version, rules_digest)` | нет (намеренно) | immutable Recap, JSON в поле `recap`; share-ссылка должна работать и после того, как исходные `events` истекут |

`annual_metrics.event_count` — freshness-маркер: хранит, сколько событий было в `events` на момент расчёта.

---

## 5. Cache-aside для метрик

`CalculateMetrics`: считает live `count(events)` для `(profile_id, year)` — если 0, `ErrMetricsNotFound`. Иначе читает `annual_metrics`; если сохранённый `event_count` совпадает с live count — отдаёт кэш. Если нет (или строки не было) — агрегирует сырые `events` через `analytics.AggregateEvents` и перезаписывает кэш.

Ограничение: freshness сравнивается только по количеству строк, не по содержимому — если существующий event изменится без изменения count, кэш это не поймает. В текущей append-only модели событий это не проблема; для mutable ingestion потребовался бы более сильный маркер (`max(updated_at)`, offset, checksum).

---

## 6. Recap: idempotency

`RecapKey = (ProfileID, Year, RulesVersion, RulesDigest)`. `CreateRecapIfAbsent` — check-then-insert (`GetRecapByKey` → если не найден → `INSERT`), не атомарный database-level upsert: MergeTree не даёт unique constraint. Для single-writer demo deployment это приемлемо; при нескольких параллельных writers возможна гонка (оба не находят строку и оба вставляют) — для multi-instance production нужен внешний lock, OLTP-хранилище с unique constraint или отдельный idempotency-сервис.

После чтения Recap не считается доверенным сам по себе: application прогоняет `engine.ValidateStored` и сверяет, что сохранённый `RecapKey`/ID/share ID совпадает с запрошенным.

---

## 7. Как события попадают в `events`

**Прямой путь (основной).** При `SEED_DEMO_DATA=true` после `EnsureSchema()` вызывается `bootstrap.LoadDemoData`: грузит `profiles.json`/`scenarios.json`, генерирует сырые события по сценариям и пишет их через `UpsertProfiles`/`InsertEvents`/`PutActionableState`. Recap-метрики руками не пишутся — только события, из которых `CalculateMetrics` сам всё агрегирует, как для реального ingestion.

**Опциональный Kafka-путь.** Не входит в основной билд, не запускается стандартным `docker compose up` — только за профилем `events-gen`. `cmd/eventgen` переиспользует ту же генерацию (`bootstrap.GenerateDemoData`), но публикует события JSON'ом в Kafka-топик `events` вместо прямой записи (retention топика — явные 24 часа, не дефолт брокера). На стороне ClickHouse `clickhouse/kafka/010_events_kafka.sql` (накатывается вручную одноразовым сервисом `clickhouse-kafka-init`, не через `EnsureSchema()`) создаёt `events_queue` (`ENGINE = Kafka`) и `MATERIALIZED VIEW events_mv TO events` — ClickHouse сам поллит топик и льёт в `events`.

```bash
docker compose --profile events-gen up -d kafka clickhouse-kafka-init
docker compose --profile events-gen run --rm eventgen
```

---

## 8. `clickhouse/init`

Монтируется Docker Compose в `/docker-entrypoint-initdb.d` и применяется при первом старте контейнера. Runtime source of truth для схемы приложения — всё же `EnsureSchema()` (раздел 3); `clickhouse/init/001_schema.sql` дублирует создание `events` для случаев ручного/раннего bring-up.

`002_recap_cache.sql` создаёт `recap_cards`/`recap_summary` — их не использует ни один Go-код в проекте, это legacy/prototype-таблицы. Со временем стоит либо удалить эту схему, либо перенести весь SQL в versioned migrations — сейчас она распределена между init-скриптами и `EnsureSchema()`, что создаёт риск рассинхронизации.

---

## 9. Storage invariants

Реализация обязана соблюдать:

1. Profile code однозначно определяет профиль; UUID — внутренний идентификатор.
2. Metrics относятся к конкретному `RecapPeriod`.
3. ActionableState — point-in-time snapshot.
4. Recap после записи не изменяется.
5. Lookup по share ID / internal ID не возвращает чужой Recap.
6. Idempotency lookup — через `RecapKey`.
7. Прочитанный Recap проходит engine validation, а не считается доверенным по факту хранения.
8. Storage не вычисляет Behavior, Achievement или NextAction — это бизнес-логика, а не персистентность.

---

## 10. Где менять storage

| Изменение | Файл/пакет
|---|---|
| Storage interfaces | `internal/recap/application/ports.go` |
| ClickHouse connection/schema | `internal/storage/clickhouse/client.go` |
| Profiles SQL | `internal/storage/clickhouse/profiles.go` |
| Event/metrics analytics | `internal/storage/clickhouse/analytics.go` |
| ActionableState | `internal/storage/clickhouse/current_state.go` |
| Recap persistence | `internal/storage/clickhouse/recaps.go` |
| Demo seeding | `internal/bootstrap/` |
| Docker ClickHouse initialization | `clickhouse/init/` |
| Kafka event generator (опционально) | `cmd/eventgen/` |
| Kafka ingestion в ClickHouse (опционально) | `clickhouse/kafka/` |
