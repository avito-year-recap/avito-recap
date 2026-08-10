-- Optional on-demand ingestion path for recap.events via Kafka. Not part of
-- clickhouse/init: docker-entrypoint-initdb.d always runs everything in that
-- directory on first container start, but this is only meant to exist when
-- the "events-gen" compose profile is in use, so it's applied separately by
-- the clickhouse-kafka-init one-off service (see docker-compose.yml).
--
-- events_queue is a Kafka engine table: it does not store rows, it's a view
-- over the topic. events_mv is what actually moves messages into
-- recap.events, triggered as ClickHouse polls the topic. String columns +
-- an explicit cast in the view (rather than typing events_queue itself as
-- UUID/DateTime) mean a single malformed message fails that row's insert
-- into events_mv, not the whole poll batch.
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
