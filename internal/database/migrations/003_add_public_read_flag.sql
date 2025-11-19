-- Add public_read flag to projects table to support unauthenticated access

ALTER TABLE projects ADD COLUMN IF NOT EXISTS "public_read" BOOLEAN DEFAULT FALSE;

---- create above / drop below ----

ALTER TABLE projects DROP COLUMN IF EXISTS "public_read";
