CREATE OR REPLACE VIEW activities_agg as
SELECT
  activities.activity_id,
  activities.project_id,
  activities.org_id,
  activities.username,
  activities.start_time,
  activities.end_time,
  EXTRACT(day from start_time) as day, 
  EXTRACT(week from start_time) as week, 
  EXTRACT(month from start_time) as month, 
  EXTRACT(quarter from start_time) as quarter, 
  EXTRACT(year from start_time) as year, 
  EXTRACT(minute from end_time - start_time) as duration_minutes, 
  EXTRACT(hour from end_time - start_time) as duration_hours,
  EXTRACT(hour from end_time - start_time) * 60 + EXTRACT(minute from end_time - start_time) as duration_minutes_total
FROM 
  activities