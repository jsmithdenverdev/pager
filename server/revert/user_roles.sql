-- Revert pager:user_roles from pg

BEGIN;

DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS user_role;

COMMIT;
