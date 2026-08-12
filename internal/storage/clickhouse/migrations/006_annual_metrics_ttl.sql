-- Докатывает TTL на annual_metrics для БД, созданных до 003_create_annual_metrics.sql
ALTER TABLE annual_metrics MODIFY TTL updated_at + INTERVAL 3 YEAR;
