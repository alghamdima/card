-- 000003_bilingual_titles.up.sql
ALTER TABLE campaigns ADD COLUMN title_ar TEXT;
ALTER TABLE campaigns ADD COLUMN title_en TEXT;

-- Backfill from existing title
UPDATE campaigns SET title_ar = title WHERE title_ar IS NULL OR title_ar = '';
UPDATE campaigns SET title_en = title WHERE title_en IS NULL OR title_en = '';
