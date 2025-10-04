-- Revert activities_agg view to original structure (without tags)
DROP VIEW IF EXISTS activities_agg;

CREATE VIEW activities_agg as
SELECT
  activities.activity_id,
  activities.project_id,
  activities.org_id,
  activities.username,
  activities.start_time,
  activities.end_time,
  activities.description,
  EXTRACT(day from start_time) as day, 
  EXTRACT(week from start_time) as week, 
  EXTRACT(month from start_time) as month, 
  EXTRACT(quarter from start_time) as quarter, 
  EXTRACT(year from start_time) as year, 
  EXTRACT(minute from end_time - start_time) as duration_minutes, 
  EXTRACT(hour from end_time - start_time) as duration_hours,
  EXTRACT(hour from end_time - start_time) * 60 + EXTRACT(minute from end_time - start_time) as duration_minutes_total
FROM 
  activities;

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