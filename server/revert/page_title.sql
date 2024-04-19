-- Revert pager:page_title from pg

BEGIN;

ALTER TABLE pages DROP COLUMN title;

COMMIT;
