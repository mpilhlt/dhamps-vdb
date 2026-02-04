-- Add sharing functionality for LLM Service Definitions
-- This migration adds a definitions_shared_with table similar to instances_shared_with
-- and seeds _system definitions to be shared with all users

-- Step 1: Create table for definition sharing (n:m relationship between definitions and users)
CREATE TABLE IF NOT EXISTS definitions_shared_with(
  "user_handle" VARCHAR(20) NOT NULL REFERENCES "users"("user_handle") ON DELETE CASCADE,
  "definition_id" INTEGER NOT NULL REFERENCES "definitions"("definition_id") ON DELETE CASCADE,
  "role" VARCHAR(20) NOT NULL REFERENCES "vdb_roles"("vdb_role"),
  "created_at" TIMESTAMP NOT NULL,
  "updated_at" TIMESTAMP NOT NULL,
  PRIMARY KEY ("user_handle", "definition_id")
);

-- Step 2: Create index for efficient lookups
CREATE INDEX IF NOT EXISTS definitions_shared_with_user_idx ON "definitions_shared_with"("user_handle");
CREATE INDEX IF NOT EXISTS definitions_shared_with_definition_idx ON "definitions_shared_with"("definition_id");

-- Step 3: Share all _system definitions with all users (represented by '*')
-- This makes global definitions automatically accessible to everyone
INSERT INTO definitions_shared_with ("user_handle", "definition_id", "role", "created_at", "updated_at")
SELECT '*', "definition_id", 'reader', NOW(), NOW()
FROM definitions
WHERE "owner" = '_system'
ON CONFLICT ("user_handle", "definition_id") DO NOTHING;

---- create above / drop below ----

-- Rollback: Remove definition sharing functionality

-- Step 3 Rollback: Remove all sharing entries for _system definitions
DELETE FROM definitions_shared_with
WHERE "definition_id" IN (
  SELECT "definition_id" FROM definitions WHERE "owner" = '_system'
) AND "user_handle" = '*';

-- Step 2 Rollback: Drop indexes
DROP INDEX IF EXISTS definitions_shared_with_definition_idx;
DROP INDEX IF EXISTS definitions_shared_with_user_idx;

-- Step 1 Rollback: Drop the definitions_shared_with table
DROP TABLE IF EXISTS definitions_shared_with;
