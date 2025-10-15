-- Enable pg_trgm extension for text search
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Table tags
CREATE TABLE tags (
    tag_id       uuid not null,
    name         varchar(50) not null,
    org_id       uuid not null,
    created_at   timestamp not null default current_timestamp,
    color        varchar(7) not null default '#6c757d'
);

ALTER TABLE tags
ADD CONSTRAINT pk_tags PRIMARY KEY (tag_id);

ALTER TABLE tags
ADD CONSTRAINT fk_tags_orgs
FOREIGN KEY (org_id) REFERENCES organizations (org_id) ON DELETE CASCADE;

-- Unique constraint on (name, org_id) for organization-level uniqueness
ALTER TABLE tags
ADD CONSTRAINT uk_tags_name_org UNIQUE (name, org_id);

-- Table activity_tags (junction table for many-to-many relationship)
CREATE TABLE activity_tags (
    activity_id  uuid not null,
    tag_id       uuid not null,
    org_id       uuid not null
);

ALTER TABLE activity_tags
ADD CONSTRAINT pk_activity_tags PRIMARY KEY (activity_id, tag_id);

ALTER TABLE activity_tags
ADD CONSTRAINT fk_activity_tags_activities
FOREIGN KEY (activity_id) REFERENCES activities (activity_id) ON DELETE CASCADE;

ALTER TABLE activity_tags
ADD CONSTRAINT fk_activity_tags_tags
FOREIGN KEY (tag_id) REFERENCES tags (tag_id) ON DELETE CASCADE;

ALTER TABLE activity_tags
ADD CONSTRAINT fk_activity_tags_orgs
FOREIGN KEY (org_id) REFERENCES organizations (org_id);

-- Indexes for efficient tag queries and autocomplete functionality

-- Index for exact lookups within organization
CREATE INDEX idx_tags_org_name
ON tags (org_id, name);

-- GIN index for autocomplete text search using trigrams
CREATE INDEX idx_tags_name_text
ON tags USING gin (name gin_trgm_ops);

-- Index for organization-wide tag queries
CREATE INDEX idx_tags_org_id
ON tags (org_id);

-- Index for activity queries
CREATE INDEX idx_activity_tags_activity
ON activity_tags (activity_id);

-- Index for tag-based filtering
CREATE INDEX idx_activity_tags_tag
ON activity_tags (tag_id);

-- Index for organization-specific filtering
CREATE INDEX idx_activity_tags_org
ON activity_tags (org_id, tag_id);

-- Extend activities_agg view to include tag information
-- Drop and recreate the view to include tags
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
LEFT JOIN activity_tags at ON activities.activity_id = at.activity_id
LEFT JOIN tags t ON at.tag_id = t.tag_id
GROUP BY 
  activities.activity_id,
  activities.project_id,
  activities.org_id,
  activities.username,
  activities.start_time,
  activities.end_time,
  activities.description