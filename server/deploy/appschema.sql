-- Deploy pager:appschema to pg
BEGIN;

---------------------------
-- EXTENSIONS
---------------------------
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

---------------------------
-- AGENCIES
---------------------------
CREATE TABLE IF NOT EXISTS agency_status (
  "status" varchar(256) NOT NULL,
  CONSTRAINT agency_status_pk PRIMARY KEY ("status")
);

INSERT INTO agency_status ("status")
VALUES ('PENDING'),
  ('ACTIVE'),
  ('INACTIVE');

CREATE TABLE IF NOT EXISTS agencies (
  id uuid DEFAULT gen_random_uuid() NOT NULL,
  "name" varchar(256) NOT NULL,
  "status" varchar(256) NOT NULL REFERENCES agency_status("status"),
  created timestamptz DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
  created_by varchar(256) NOT NULL,
  modified timestamp DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
  modified_by varchar(256) NOT NULL,
  CONSTRAINT agencies_pk PRIMARY KEY (id)
);

---------------------------
-- USERS
---------------------------
CREATE TABLE IF NOT EXISTS user_status (
  "status" varchar(256) NOT NULL,
  CONSTRAINT user_status_pk PRIMARY KEY ("status")
);

INSERT INTO user_status ("status")
VALUES ('PENDING'),
  ('ACTIVE'),
  ('INACTIVE');

CREATE TABLE IF NOT EXISTS users (
  id uuid DEFAULT gen_random_uuid() NOT NULL,
  email varchar(256) NOT NULL UNIQUE,
  "idp_id" varchar(256) NOT NULL UNIQUE,
  "status" varchar(256) NOT NULL REFERENCES user_status("status"),
  created timestamptz DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
  created_by varchar(256) NOT NULL,
  modified timestamp DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
  modified_by varchar(256) NOT NULL,
  CONSTRAINT users_pk PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS user_agencies(
  user_id uuid NOT NULL REFERENCES users(id),
  agency_id uuid NOT NULL REFERENCES agencies(id),
  CONSTRAINT user_agencies_unique_user_id_agency_id UNIQUE (user_id, agency_id)
);

---------------------------
-- DEVICES
---------------------------
CREATE TABLE IF NOT EXISTS device_status (
  "status" varchar(256) NOT NULL,
  CONSTRAINT device_status_pk PRIMARY KEY ("status")
);

INSERT INTO device_status ("status")
VALUES ('PENDING'),
  ('ACTIVE'),
  ('INACTIVE');

CREATE TABLE IF NOT EXISTS devices (
  id uuid DEFAULT gen_random_uuid() NOT NULL,
  "status" varchar(256) NOT NULL REFERENCES device_status("status"),
  "endpoint" varchar(256) NULL,
  user_id uuid NOT NULL REFERENCES users(id),
  code varchar(256) NOT NULL,
  created timestamptz DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
  created_by varchar(256) NOT NULL,
  modified timestamp DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
  modified_by varchar(256) NOT NULL,
  CONSTRAINT devices_pk PRIMARY KEY (id)
);

---------------------------
-- PAGES
---------------------------
CREATE TABLE IF NOT EXISTS page_delivery_status (
  "status" varchar(256) NOT NULL,
  CONSTRAINT page_delivery_status_pk PRIMARY KEY ("status")
);

INSERT INTO page_delivery_status ("status")
VALUES ('PENDING'),
  ('DELIVERING'),
  ('DELIVERED'),
  ('DELIVERY_FAILED');

CREATE TABLE IF NOT EXISTS pages (
  id uuid DEFAULT gen_random_uuid() NOT NULL,
  agency_id uuid NOT NULL REFERENCES agencies(id),
  -- Page content contains CAD notes and thus should be expected to be very
  -- large
  "content" varchar(65536) NULL,
  created timestamptz DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
  created_by varchar(256) NOT NULL,
  modified timestamp DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
  modified_by varchar(256) NOT NULL,
  CONSTRAINT pages_pk PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS page_deliveries(
  id uuid DEFAULT gen_random_uuid() NOT NULL,
  page_id uuid NOT NULL REFERENCES pages(id),
  device_id uuid NOT NULL REFERENCES devices(id),
  "status" varchar(256) NOT NULL references page_delivery_status("status"),
  created timestamptz DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
  -- Not sure that created_by and modified_by make sense, this will always be
  -- 'system'. But it doesn't hurt to have the data.
  created_by varchar(256) NOT NULL,
  modified timestamp DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
  modified_by varchar(256) NOT NULL,
  CONSTRAINT page_deliveries_pk PRIMARY KEY (id),
  CONSTRAINT page_deliveries_unique_page_id_device_id UNIQUE (page_id, device_id)
);

COMMIT;