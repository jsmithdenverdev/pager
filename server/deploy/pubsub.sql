-- Deploy pager:pubsub to pg

BEGIN;

CREATE TABLE IF NOT EXISTS topics (
  topic VARCHAR(256) NOT NULL,
  retries_enabled BOOLEAN NOT NULL DEFAULT false,
  retries INTEGER NOT NULL DEFAULT 0,
  CONSTRAINT topics_unique_topic UNIQUE(topic)
);

CREATE TABLE IF NOT EXISTS "messages" (
  id uuid DEFAULT gen_random_uuid() NOT NULL,
  topic VARCHAR(256) NOT NULL REFERENCES topics(topic),
  "payload" JSONB NOT NULL,
  retries INTEGER NOT NULL DEFAULT 0,
  created timestamptz DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
  created_by varchar(256) NOT NULL,
  modified timestamptz DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
  modified_by varchar(256) NOT NULL,
  CONSTRAINT messages_pk PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS "messages_dl" (
  id uuid DEFAULT gen_random_uuid() NOT NULL,
  message_id uuid NULL REFERENCES messages(id),
  created timestamptz DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
  created_by varchar(256) NOT NULL,
  modified timestamptz DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
  modified_by varchar(256) NOT NULL,
  CONSTRAINT messages_dl_pk PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS unprocessable_messages (
  id uuid DEFAULT gen_random_uuid() NOT NULL,
  topic VARCHAR(256) NOT NULL REFERENCES topics(topic),
  "payload" VARCHAR(4096) NOT NULL,
  created timestamptz DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
  created_by varchar(256) NOT NULL,
  modified timestamptz DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
  modified_by varchar(256) NOT NULL,
  CONSTRAINT unprocessable_messages_pk  PRIMARY KEY (id)
);

-- Create or replace the trigger function to handle insertions and updates
CREATE OR REPLACE FUNCTION after_messages_insert_or_update()
RETURNS TRIGGER AS $$
BEGIN
    -- Check if the operation is an insertion or update
    IF TG_OP = 'INSERT' THEN
      PERFORM pg_notify(NEW.topic, row_to_json(NEW)::text);
    ELSIF TG_OP = 'UPDATE' THEN
      PERFORM pg_notify(NEW.topic, row_to_json(NEW)::text);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create or replace the trigger to execute the trigger function after insertions or updates
CREATE TRIGGER after_messages_insert_or_update
AFTER INSERT OR UPDATE ON messages
FOR EACH ROW
EXECUTE FUNCTION after_messages_insert_or_update();

INSERT INTO topics (topic, retries_enabled, retries)
VALUES 
  ('PROVISION_USER', true, 5),
  ('SEND_PAGE', true, 5);

COMMIT;
