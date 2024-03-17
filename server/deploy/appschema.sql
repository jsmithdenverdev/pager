-- Deploy pager:appschema to pg

BEGIN;

-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS agency (
	id uuid DEFAULT gen_random_uuid() NOT NULL,
	"name" varchar NOT NULL,
	created_by_id varchar NOT NULL,
	created timestamptz DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
	modified_by_id varchar NOT NULL,
	modified timestamp DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
	CONSTRAINT agency_pk PRIMARY KEY (id)
);

COMMIT;
