-- Deploy pager:user_roles to pg

BEGIN;

---------------------------
-- USER ROLE
---------------------------
CREATE TABLE IF NOT EXISTS user_role (
  "role" varchar(256) NOT NULL,
  CONSTRAINT user_role_pk PRIMARY KEY ("role")
);

CREATE TABLE IF NOT EXISTS user_roles (
  "role" varchar(256) NOT NULL REFERENCES user_role("role"),
  user_id UUID NOT NULL REFERENCES users(id),
  agency_id UUID NULL REFERENCES agencies(id),
  CONSTRAINT user_roles_unique_role_user_agency UNIQUE ("role", user_id, agency_id)
);

INSERT INTO user_role ("role")
VALUES ('READER'), ('WRITER'), ('PLATFORM_ADMIN');

COMMIT;
