-- Текущее состояние профиля: point-in-time snapshot, не time-decaying лог, поэтому без TTL
CREATE TABLE IF NOT EXISTS actionable_state
(
    profile_id UUID,
    state      String,
    updated_at DateTime DEFAULT now()
)
ENGINE = ReplacingMergeTree(updated_at)
ORDER BY profile_id;
