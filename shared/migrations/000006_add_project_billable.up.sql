-- Add billable column to projects table
ALTER TABLE projects ADD COLUMN billable boolean NOT NULL DEFAULT false;

-- Update activities_agg view to include billable from projects
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
  projects.billable,
  EXTRACT(day from start_time) as day, 
  EXTRACT(week from start_time) as week, 
  EXTRACT(month from start_time) as month, 
  EXTRACT(quarter from start_time) as quarter, 
  EXTRACT(year from start_time) as year, 
  EXTRACT(minute from end_time - start_time) as duration_minutes, 
  EXTRACT(hour from end_time - start_time) as duration_hours,
  EXTRACT(hour from end_time - start_time) * 60 + EXTRACT(minute from end_time - start_time) as duration_minutes_total,
  -- Tag information as JSON array
  COALESCE(
    JSON_AGG(
      JSON_BUILD_OBJECT(
        'name', t.name,
        'color', t.color
      ) ORDER BY t.name
    ) FILTER (WHERE t.tag_id IS NOT NULL),
    '[]'::json
  ) as tags_info
FROM 
  activities
LEFT JOIN projects ON activities.project_id = projects.project_id
LEFT JOIN activity_tags at ON activities.activity_id = at.activity_id
LEFT JOIN tags t ON at.tag_id = t.tag_id
GROUP BY 
  activities.activity_id,
  activities.project_id,
  activities.org_id,
  activities.username,
  activities.start_time,
  activities.end_time,
  activities.description,
  projects.billable;

