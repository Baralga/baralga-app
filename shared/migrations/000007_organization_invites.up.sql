-- Create organization_invites table
CREATE TABLE organization_invites (
    invite_id        uuid not null,
    org_id           uuid not null,
    token            varchar(255) not null,
    created_by       uuid not null,
    created_at       timestamp not null default now(),
    expires_at       timestamp not null,
    used_at          timestamp null,
    used_by          uuid null,
    active           boolean not null default true
);

-- Add constraints and indexes
ALTER TABLE organization_invites
    ADD CONSTRAINT pk_organization_invites PRIMARY KEY (invite_id);

ALTER TABLE organization_invites
    ADD CONSTRAINT fk_organization_invites_orgs
    FOREIGN KEY (org_id) REFERENCES organizations (org_id);

ALTER TABLE organization_invites
    ADD CONSTRAINT fk_organization_invites_created_by
    FOREIGN KEY (created_by) REFERENCES users (user_id);

ALTER TABLE organization_invites
    ADD CONSTRAINT fk_organization_invites_used_by
    FOREIGN KEY (used_by) REFERENCES users (user_id);

CREATE UNIQUE INDEX organization_invites_idx_token
    ON organization_invites (token);

CREATE INDEX organization_invites_idx_org_id
    ON organization_invites (org_id, active, expires_at);

CREATE INDEX organization_invites_idx_expires_at
    ON organization_invites (expires_at, active);
