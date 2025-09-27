-- Drop indexes
DROP INDEX IF EXISTS organization_invites_idx_expires_at;
DROP INDEX IF EXISTS organization_invites_idx_org_id;
DROP INDEX IF EXISTS organization_invites_idx_token;

-- Drop table
DROP TABLE IF EXISTS organization_invites;
