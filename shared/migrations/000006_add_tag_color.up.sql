-- Add color column to tags table
ALTER TABLE tags
ADD COLUMN color varchar(7) NOT NULL DEFAULT '#6c757d';

-- Update existing tags with default color
UPDATE tags SET color = '#6c757d' WHERE color IS NULL;