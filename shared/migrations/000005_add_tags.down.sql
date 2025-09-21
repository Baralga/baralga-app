-- Drop indexes
DROP INDEX IF EXISTS idx_activity_tags_org;
DROP INDEX IF EXISTS idx_activity_tags_tag;
DROP INDEX IF EXISTS idx_activity_tags_activity;
DROP INDEX IF EXISTS idx_tags_org_id;
DROP INDEX IF EXISTS idx_tags_name_text;
DROP INDEX IF EXISTS idx_tags_org_name;

-- Drop tables
DROP TABLE IF EXISTS activity_tags;
DROP TABLE IF EXISTS tags;

-- Note: We don't drop the pg_trgm extension as it might be used by other parts of the system