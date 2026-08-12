-- Кэш агрегации над events; event_count служит freshness-маркером
CREATE TABLE IF NOT EXISTS annual_metrics
(
    profile_id  UUID,
    year        UInt16,
    metrics     String,
    event_count UInt64,
    updated_at  DateTime DEFAULT now()
)
ENGINE = ReplacingMergeTree(updated_at)
ORDER BY (profile_id, year)
TTL updated_at + INTERVAL 3 YEAR;
