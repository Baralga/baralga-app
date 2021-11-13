-- Table organizations
CREATE TABLE organizations (
     org_id       uuid not null,
     title        varchar(255),
     description  varchar(4000)
);

ALTER TABLE organizations
    ADD CONSTRAINT pk_organizations PRIMARY KEY (org_id);


-- Table projects
CREATE TABLE projects (
     project_id   uuid not null,
     org_id       uuid not null,
     title        varchar(255),
     description  varchar(4000),
     active       boolean
);

ALTER TABLE projects
ADD CONSTRAINT pk_projects PRIMARY KEY (project_id);

ALTER TABLE projects
ADD CONSTRAINT fk_projects_orgs
FOREIGN KEY (org_id) REFERENCES organizations (org_id);

CREATE INDEX projects_idx_project_id_org_id
ON projects (project_id, org_id);

-- Table activities
CREATE TABLE activities (
     activity_id  uuid not null,
     description  varchar(4000),
     username     varchar(36) not null,
     start_time   timestamp,
     end_time     timestamp,
     project_id   uuid not null,
     org_id       uuid not null
);

ALTER TABLE activities
ADD CONSTRAINT pk_activities PRIMARY KEY (activity_id);

ALTER TABLE activities
ADD CONSTRAINT fk_activities_orgs
FOREIGN KEY (org_id) REFERENCES organizations (org_id);

ALTER TABLE activities
ADD CONSTRAINT fk_activities_project
FOREIGN KEY (project_id) REFERENCES projects (project_id);

CREATE INDEX activities_idx_user
ON activities (org_id, username, start_time);

CREATE INDEX activities_idx_project_id_org_id
ON activities (activity_id, org_id);


-- Table users
CREATE TABLE users (
  user_id  UUID NOT NULL,
  username VARCHAR(50) NOT NULL,
  password VARCHAR(100) NOT NULL,
  enabled  INTEGER NOT NULL DEFAULT 1,
  org_id   uuid not null
);

ALTER TABLE users
ADD CONSTRAINT pk_users PRIMARY KEY (user_id);

CREATE UNIQUE INDEX users_idx_username
ON users (username);

ALTER TABLE users
ADD CONSTRAINT fk_users_orgs
FOREIGN KEY (org_id) REFERENCES organizations (org_id);


-- Table roles
CREATE TABLE roles (
  user_id   UUID NOT NULL,
  role VARCHAR(50) NOT NULL,
  org_id    uuid not null
);

ALTER TABLE roles
ADD CONSTRAINT fk_roles_users
FOREIGN KEY (user_id) REFERENCES users (user_id);

ALTER TABLE roles
ADD CONSTRAINT fk_roles_orgs
FOREIGN KEY (org_id) REFERENCES organizations (org_id);

CREATE UNIQUE INDEX roles_idx_user_id
  on roles (user_id, role, org_id);


-- Insert initial data
INSERT INTO organizations (org_id, title, description)
    values ('4ed0c11d-3d6a-41c1-9873-558e86084591', 'main', null);

INSERT INTO projects (project_id, title, description, active, org_id)
    values ('f4b1087c-8fbb-4c8d-bbb7-ab4d46da16ea', 
           'My Project',
           null, 
           true,
           '4ed0c11d-3d6a-41c1-9873-558e86084591'
    );

INSERT INTO users (user_id, username, password, enabled, org_id)
  values (
        'eeeeeb80-33f3-4d3f-befe-58694d2ac841',
        'admin',
        '$2a$10$NuzYobDOSTCx/EKBClGwGe0A9c8/yC7D4IP75hwz1jn.RCBfdEtb2', -- adm1n
        1,
        '4ed0c11d-3d6a-41c1-9873-558e86084591'
  );

INSERT INTO roles (user_id, role, org_id)
  values ('eeeeeb80-33f3-4d3f-befe-58694d2ac841',
         'ROLE_ADMIN',
        '4ed0c11d-3d6a-41c1-9873-558e86084591'
  );

INSERT INTO users (user_id, username, password, enabled, org_id)
  values (
        '04b4adc8-2b7f-4ec0-aeb8-407ce164484e',
        'user1',
        '$2a$10$IhFsXJYqYG56/b1JgzZzv.kPcPsJnXeQzD9evMOUHg2LT/.Oz9uEu', -- us3r
        1,
        '4ed0c11d-3d6a-41c1-9873-558e86084591'
);

INSERT INTO roles (user_id, role, org_id)
  values ('04b4adc8-2b7f-4ec0-aeb8-407ce164484e', 
         'ROLE_USER',
        '4ed0c11d-3d6a-41c1-9873-558e86084591'
  );

INSERT INTO users (user_id, username, password, enabled, org_id)
  values (
        '6504aa30-94e1-4337-a0bb-d57b6c0fe62f',
        'user2',
        '$2a$10$IhFsXJYqYG56/b1JgzZzv.kPcPsJnXeQzD9evMOUHg2LT/.Oz9uEu', -- us3r
        1,
        '4ed0c11d-3d6a-41c1-9873-558e86084591'
);

INSERT INTO roles (user_id, role, org_id)
  values ('6504aa30-94e1-4337-a0bb-d57b6c0fe62f', 
        'ROLE_USER',
        '4ed0c11d-3d6a-41c1-9873-558e86084591'
  );
