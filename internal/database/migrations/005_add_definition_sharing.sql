-- Add sharing functionality for LLM Service Definitions
-- This migration adds a definitions_shared_with table similar to instances_shared_with
-- Note: _system definitions are made accessible to all users via application logic,
-- not through explicit sharing records

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

-- Note: _system definitions are automatically accessible to all users via GetAccessibleDefinitionsByUser query
-- No explicit sharing records needed

---- create above / drop below ----

-- Rollback: Remove definition sharing functionality

-- Step 2 Rollback: Drop indexes
DROP INDEX IF EXISTS definitions_shared_with_definition_idx;
DROP INDEX IF EXISTS definitions_shared_with_user_idx;

-- Step 1 Rollback: Drop the definitions_shared_with table
DROP TABLE IF EXISTS definitions_shared_with;
