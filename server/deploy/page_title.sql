-- Deploy pager:page_title to pg

BEGIN;

ALTER TABLE pages ADD title VARCHAR(256) NOT NULL;

COMMIT;
