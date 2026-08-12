-- Сырые события пользователя: единственный источник истины годовой активности
CREATE TABLE IF NOT EXISTS events
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
TTL occurred_at + INTERVAL 3 YEAR;
