-- Неизменяемый Recap; не Replacing — после генерации строка не обновляется
CREATE TABLE IF NOT EXISTS recaps
(
    id            UUID,
    share_id      UUID,
    profile_id    UUID,
    year          UInt16,
    rules_version String,
    rules_digest  String,
    recap         String,
    created_at    DateTime DEFAULT now()
)
ENGINE = MergeTree
ORDER BY (profile_id, year, rules_version, rules_digest);
