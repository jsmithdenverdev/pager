-- Deploy pager:appschema to pg

BEGIN;

-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS agency (
	id uuid DEFAULT gen_random_uuid() NOT NULL,
	"name" varchar(256) NOT NULL,
	created timestamptz DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
	created_by varchar(256) NOT NULL,
	modified timestamp DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
	modified_by varchar(256) NOT NULL,
	CONSTRAINT agency_pk PRIMARY KEY (id)
);

COMMIT;
