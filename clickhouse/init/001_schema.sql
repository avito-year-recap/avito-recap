-- Сырые события пользователя: единственный источник истины для годовой
-- активности. AnalyticsStorage.CalculateMetrics не хранит готовые метрики
-- напрямую — они лениво считаются из этой таблицы через
-- analytics.AggregateEvents и кэшируются в recap.annual_metrics
-- (см. internal/storage/clickhouse/client.go, schemaStatements).
CREATE TABLE IF NOT EXISTS recap.events
(
    id          UUID,
    profile_id  UUID,
    event_type  LowCardinality(String),
    occurred_at DateTime,
    category    LowCardinality(String),
    ad_id       Nullable(UInt64),
    dialog_id   Nullable(UInt64)
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(occurred_at)
ORDER BY (profile_id, occurred_at)
TTL occurred_at + INTERVAL 3 YEAR
SETTINGS index_granularity = 8192;
