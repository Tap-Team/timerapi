BEGIN;

ALTER TABLE timers ADD COLUMN created_at timestamp(0) default now();

COMMIT;