-- Enable pg_trgm extension for text search
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Table tags
CREATE TABLE tags (
    tag_id       uuid not null,
    name         varchar(50) not null,
    org_id       uuid not null,
    created_at   timestamp not null default current_timestamp
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