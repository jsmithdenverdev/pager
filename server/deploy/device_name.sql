-- Deploy pager:device_name to pg

BEGIN;

ALTER TABLE devices
ADD name VARCHAR(64);

COMMIT;
