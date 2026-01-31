-- Refactor LLM services architecture into Definitions and Instances
-- This migration separates service templates (definitions) from user-specific instances

-- Step 1: Create the _system user for global definitions
INSERT INTO users ("user_handle", "name", "email", "vdb_api_key", "created_at", "updated_at")
VALUES ('_system', 'System User', 'system@dhamps-vdb.internal', 
        -- Generate a system API key (64 chars of zeros as placeholder)
        '0000000000000000000000000000000000000000000000000000000000000000',
        NOW(), NOW())
ON CONFLICT ("user_handle") DO NOTHING;

-- Step 2: Create LLM Service Definitions table (templates that can be shared)
CREATE TABLE IF NOT EXISTS llm_service_definitions(
  "definition_id" SERIAL PRIMARY KEY,
  "definition_handle" VARCHAR(20) NOT NULL,
  "owner" VARCHAR(20) NOT NULL REFERENCES "users"("user_handle") ON DELETE CASCADE,
  "endpoint" TEXT NOT NULL,
  "description" TEXT,
  "api_standard" VARCHAR(20) NOT NULL REFERENCES "api_standards"("api_standard_handle"),
  "model" TEXT NOT NULL,
  "dimensions" INTEGER NOT NULL,
  "created_at" TIMESTAMP NOT NULL,
  "updated_at" TIMESTAMP NOT NULL,
  UNIQUE ("owner", "definition_handle")
);

CREATE INDEX IF NOT EXISTS llm_service_definitions_handle ON "llm_service_definitions"("definition_handle");

-- Step 3: Rename existing llm_services table to llm_service_instances
ALTER TABLE llm_services RENAME TO llm_service_instances;
ALTER TABLE llm_service_instances RENAME COLUMN llm_service_id TO instance_id;
ALTER TABLE llm_service_instances RENAME COLUMN llm_service_handle TO instance_handle;

-- Step 4: Add definition_id to instances (nullable to allow standalone instances)
ALTER TABLE llm_service_instances ADD COLUMN "definition_id" INTEGER REFERENCES "llm_service_definitions"("definition_id") ON DELETE SET NULL;

-- Step 5: Add encrypted API key column and mark old one for deprecation
ALTER TABLE llm_service_instances ADD COLUMN "api_key_encrypted" BYTEA;
-- Note: We keep the old api_key column temporarily for migration purposes

-- Step 6: Update the index name
DROP INDEX IF EXISTS llm_services_handle;
CREATE INDEX IF NOT EXISTS llm_service_instances_handle ON "llm_service_instances"("instance_handle");

-- Step 7: Create table for instance sharing (n:m relationship between instances and users)
CREATE TABLE IF NOT EXISTS llm_service_instances_shared_with(
  "user_handle" VARCHAR(20) NOT NULL REFERENCES "users"("user_handle") ON DELETE CASCADE,
  "instance_id" INTEGER NOT NULL REFERENCES "llm_service_instances"("instance_id") ON DELETE CASCADE,
  "role" VARCHAR(20) NOT NULL REFERENCES "vdb_roles"("vdb_role"),
  "created_at" TIMESTAMP NOT NULL,
  "updated_at" TIMESTAMP NOT NULL,
  PRIMARY KEY ("user_handle", "instance_id")
);

-- Step 8: Rename users_llm_services table and update to reference instances
ALTER TABLE users_llm_services RENAME TO users_llm_service_instances;
ALTER TABLE users_llm_service_instances RENAME COLUMN llm_service_id TO instance_id;

-- Step 9: Update projects to have a single LLM service instance (1:1 relationship)
-- Add the new column (nullable initially)
ALTER TABLE projects ADD COLUMN "llm_service_instance_id" INTEGER REFERENCES "llm_service_instances"("instance_id") ON DELETE RESTRICT;

-- Step 10: Migrate data - for each project, pick the first linked LLM service instance
-- This is a best-effort migration; admins should verify manually if multiple services were used
UPDATE projects p
SET llm_service_instance_id = (
    SELECT pls.llm_service_id
    FROM projects_llm_services pls
    WHERE pls.project_id = p.project_id
    ORDER BY pls.created_at
    LIMIT 1
)
WHERE EXISTS (
    SELECT 1 FROM projects_llm_services pls WHERE pls.project_id = p.project_id
);

-- Step 11: Update embeddings table to reference instance_id
ALTER TABLE embeddings RENAME COLUMN llm_service_id TO llm_service_instance_id;

-- Step 12: Update the projects_llm_services references
ALTER TABLE projects_llm_services RENAME COLUMN llm_service_id TO llm_service_instance_id;

-- Step 13: Drop the old projects_llm_services table (many-to-many, no longer needed)
-- Projects now have exactly one instance via the llm_service_instance_id column
DROP TABLE IF EXISTS projects_llm_services;

-- Step 14: Ensure required API standards exist before creating definitions
-- These API standards are needed for the default LLM Service Definitions

INSERT INTO api_standards ("api_standard_handle", "description", "key_method", "key_field", "created_at", "updated_at")
VALUES ('openai', 'OpenAI Embeddings API, Version 1', 'auth_bearer', 'Authorization', NOW(), NOW())
ON CONFLICT ("api_standard_handle") DO NOTHING;

INSERT INTO api_standards ("api_standard_handle", "description", "key_method", "key_field", "created_at", "updated_at")
VALUES ('cohere', 'Cohere Embed API, Version 2', 'auth_bearer', 'Authorization', NOW(), NOW())
ON CONFLICT ("api_standard_handle") DO NOTHING;

INSERT INTO api_standards ("api_standard_handle", "description", "key_method", "key_field", "created_at", "updated_at")
VALUES ('gemini', 'Gemini Embeddings API', 'auth_bearer', 'x-goog-api-key', NOW(), NOW())
ON CONFLICT ("api_standard_handle") DO NOTHING;

-- Step 15: Seed default LLM Service Definitions from _system user
-- These serve as templates that all users can reference

-- OpenAI text-embedding-3-large
INSERT INTO llm_service_definitions 
  ("definition_handle", "owner", "endpoint", "description", "api_standard", "model", "dimensions", "created_at", "updated_at")
VALUES 
  ('openai-large', '_system', 'https://api.openai.com/v1/embeddings', 
   'OpenAI text-embedding-3-large service (3072 dimensions)', 
   'openai', 'text-embedding-3-large', 3072, NOW(), NOW())
ON CONFLICT ("owner", "definition_handle") DO NOTHING;

-- OpenAI text-embedding-3-small
INSERT INTO llm_service_definitions 
  ("definition_handle", "owner", "endpoint", "description", "api_standard", "model", "dimensions", "created_at", "updated_at")
VALUES 
  ('openai-small', '_system', 'https://api.openai.com/v1/embeddings', 
   'OpenAI text-embedding-3-small service (1536 dimensions)', 
   'openai', 'text-embedding-3-small', 1536, NOW(), NOW())
ON CONFLICT ("owner", "definition_handle") DO NOTHING;

-- Cohere embed-multilingual-v3.0
INSERT INTO llm_service_definitions 
  ("definition_handle", "owner", "endpoint", "description", "api_standard", "model", "dimensions", "created_at", "updated_at")
VALUES 
  ('cohere-multilingual-v3', '_system', 'https://api.cohere.com/v2/embed', 
   'Cohere embed-multilingual-v3.0 service (1024 dimensions)', 
   'cohere', 'embed-multilingual-v3.0', 1024, NOW(), NOW())
ON CONFLICT ("owner", "definition_handle") DO NOTHING;

-- Cohere embed-v4.0
INSERT INTO llm_service_definitions 
  ("definition_handle", "owner", "endpoint", "description", "api_standard", "model", "dimensions", "created_at", "updated_at")
VALUES 
  ('cohere-v4', '_system', 'https://api.cohere.com/v2/embed', 
   'Cohere embed-v4.0 service (1536 dimensions)', 
   'cohere', 'embed-v4.0', 1536, NOW(), NOW())
ON CONFLICT ("owner", "definition_handle") DO NOTHING;

-- Google Gemini embedding-001
INSERT INTO llm_service_definitions 
  ("definition_handle", "owner", "endpoint", "description", "api_standard", "model", "dimensions", "created_at", "updated_at")
VALUES 
  ('gemini-embedding-001', '_system', 
   'https://generativelanguage.googleapis.com/v1beta/models/gemini-embedding-001:embedContent', 
   'Google Gemini embedding-001 service (768 dimensions)', 
   'gemini', 'gemini-embedding-001', 768, NOW(), NOW())
ON CONFLICT ("owner", "definition_handle") DO NOTHING;


---- create above / drop below ----

-- Rollback instructions (reverse order)

-- Drop seeded definitions
DELETE FROM llm_service_definitions WHERE owner = '_system';

-- Restore projects_llm_services table
CREATE TABLE IF NOT EXISTS projects_llm_services(
  "project_id" SERIAL NOT NULL REFERENCES "projects"("project_id") ON DELETE CASCADE,
  "llm_service_instance_id" SERIAL NOT NULL REFERENCES "llm_service_instances"("instance_id") ON DELETE CASCADE,
  "created_at" TIMESTAMP NOT NULL,
  "updated_at" TIMESTAMP NOT NULL,
  PRIMARY KEY ("project_id", "llm_service_instance_id")
);

-- Rename embeddings column back
ALTER TABLE embeddings RENAME COLUMN llm_service_instance_id TO llm_service_id;

-- Remove the single instance reference from projects
ALTER TABLE projects DROP COLUMN IF EXISTS llm_service_instance_id;

-- Rename users_llm_service_instances table back
ALTER TABLE users_llm_service_instances RENAME COLUMN instance_id TO llm_service_id;
ALTER TABLE users_llm_service_instances RENAME TO users_llm_services;

-- Drop instance sharing table
DROP TABLE IF EXISTS llm_service_instances_shared_with;

-- Restore index name
DROP INDEX IF EXISTS llm_service_instances_handle;
CREATE INDEX IF NOT EXISTS llm_services_handle ON "llm_service_instances"("llm_service_handle");

-- Remove new columns from instances
ALTER TABLE llm_service_instances DROP COLUMN IF EXISTS api_key_encrypted;
ALTER TABLE llm_service_instances DROP COLUMN IF EXISTS definition_id;

-- Rename instances table back to llm_services
ALTER TABLE llm_service_instances RENAME COLUMN instance_handle TO llm_service_handle;
ALTER TABLE llm_service_instances RENAME COLUMN instance_id TO llm_service_id;
ALTER TABLE llm_service_instances RENAME TO llm_services;

-- Drop definitions table
DROP TABLE IF EXISTS llm_service_definitions;

-- Remove _system user
DELETE FROM users WHERE user_handle = '_system';
