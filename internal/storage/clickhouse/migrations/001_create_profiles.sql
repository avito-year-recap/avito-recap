-- Каталог тестовых профилей
CREATE TABLE IF NOT EXISTS profiles
(
    id           UUID,
    code         String,
    display_name String,
    description  String,
    avatar_url   String,
    updated_at   DateTime DEFAULT now()
)
ENGINE = ReplacingMergeTree(updated_at)
ORDER BY id;
