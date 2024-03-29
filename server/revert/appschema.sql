-- Revert pager:appschema from pg

BEGIN;

DROP TABLE IF EXISTS page_deliveries;
DROP TABLE IF EXISTS pages;
DROP TABLE IF EXISTS page_delivery_status;

DROP TABLE IF EXISTS devices;
DROP TABLE IF EXISTS device_status;

DROP TABLE IF EXISTS user_agencies;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS user_status;

DROP TABLE IF EXISTS agencies;
DROP TABLE IF EXISTS agency_status;

DROP EXTENSION IF EXISTS "uuid-ossp";;

COMMIT;
