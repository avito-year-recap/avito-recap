--Необработанная таблица событий: каждое отслеживаемое действие пользователя на платформе
CREATE TABLE IF NOT EXISTS recap.events
(
    event_id     UUID,
    user_id      UInt64,
    event_type   LowCardinality(String),
    event_time   DateTime,
    category     LowCardinality(String),
    subcategory  LowCardinality(String),
    city         LowCardinality(String),
    price        Nullable(Float64),
    -- id объявления
    ad_id        Nullable(UInt64),
    dialog_id    Nullable(UInt64),
    search_query Nullable(String),
    metadata     Map(String, String)
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(event_time)
ORDER BY (user_id, event_time)
TTL event_time + INTERVAL 3 YEAR
SETTINGS index_granularity = 8192;

-- Предварительно агрегированная таблица
CREATE TABLE IF NOT EXISTS recap.recap_category_month_agg
(
    user_id    UInt64,
    year       UInt16,
    month      UInt8,
    category   LowCardinality(String),
    event_type LowCardinality(String),
    cnt        AggregateFunction(count)
)
ENGINE = AggregatingMergeTree
ORDER BY (user_id, year, month, category, event_type);

-- Создаем VIEW для заполнения таблицы выше
-- (в отличие от постгреса работает как триггер - обновлять не нужно)
CREATE MATERIALIZED VIEW IF NOT EXISTS recap.recap_category_month_agg_mv
TO recap.recap_category_month_agg
AS
SELECT
    user_id,
    toYear(event_time)  AS year,
    toMonth(event_time) AS month,
    category,
    event_type,
    countState() AS cnt
FROM recap.events
GROUP BY user_id, year, month, category, event_type;
