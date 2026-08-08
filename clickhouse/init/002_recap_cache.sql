--В теории можем и не иметь эти таблицы - логика создания карточек переносится на сервер.
-- Это позволяет нам сохранять состояние + мы все равно выигрываем по времени, когда сервер их создает
-- Остается возможность создать админку и переодически менять типы карточек
-- Таблица пользователь + карточка (уже созданная)
CREATE TABLE IF NOT EXISTS recap.recap_cards
(
    user_id      UInt64,
    year         UInt16,
    card_id      LowCardinality(String),
    type         LowCardinality(String),
    position     UInt8,
    visibility   LowCardinality(String),
    title        String,
    subtitle     String,
    emoji        String,
    metric_value Nullable(Float64),
    metric_unit  String,
    metric_label String,
    reason       String,
    cta_type     String,
    cta_label    String,
    cta_url      String,
    generated_at DateTime
)
ENGINE = ReplacingMergeTree(generated_at)
ORDER BY (user_id, year, card_id);

-- Общая информация о пользователе
CREATE TABLE IF NOT EXISTS recap.recap_summary
(
    user_id             UInt64,
    year                UInt16,
    persona_code        LowCardinality(String),
    persona_title       String,
    persona_description String,
    persona_reason      String,
    next_action_type    LowCardinality(String),
    next_action_label   String,
    next_action_url     String,
    generated_at        DateTime
)
ENGINE = ReplacingMergeTree(generated_at)
ORDER BY (user_id, year);
