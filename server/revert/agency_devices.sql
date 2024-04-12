-- Revert pager:agency_devices from pg

BEGIN;

DROP TABLE IF EXISTS agency_devices;

COMMIT;
