-- Revert pager:pubsub from pg

BEGIN;

DROP TRIGGER IF EXISTS after_messages_insert_or_update ON messages;
DROP FUNCTION IF EXISTS after_messages_insert_or_update;
DROP TABLE IF EXISTS unprocessable_messages;
DROP TABLE IF EXISTS messages_dl;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS topics;

COMMIT;
