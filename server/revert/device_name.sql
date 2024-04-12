-- Revert pager:device_name from pg

BEGIN;

ALTER TABLE devices
DROP COLUMN name;

COMMIT;
