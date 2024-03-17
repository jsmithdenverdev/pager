-- Revert pager:appschema from pg

BEGIN;

DROP TABLE IF EXISTS agency;
DROP EXTENSION IF EXISTS "uuid-ossp";;

COMMIT;
