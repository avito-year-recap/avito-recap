
CREATE TABLE IF NOT EXISTS recap.events_queue
(
    id          String,
    profile_id  String,
    event_type  String,
    occurred_at String,
    category    String,
    ad_id       Nullable(UInt64),
    dialog_id   Nullable(UInt64)
)
ENGINE = Kafka
SETTINGS
    kafka_broker_list = 'kafka:9092',
    kafka_topic_list = 'events',
    kafka_group_name = 'clickhouse-events-consumer',
    kafka_format = 'JSONEachRow',
    kafka_num_consumers = 1,
    kafka_skip_broken_messages = 5;

CREATE MATERIALIZED VIEW IF NOT EXISTS recap.events_mv
TO recap.events
AS
SELECT
    toUUID(id)                           AS id,
    toUUID(profile_id)                   AS profile_id,
    event_type,
    parseDateTimeBestEffort(occurred_at) AS occurred_at,
    category,
    ad_id,
    dialog_id
FROM recap.events_queue;
