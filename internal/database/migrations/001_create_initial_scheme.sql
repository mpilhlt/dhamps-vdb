-- This enables the vector extension.

-- Currently disabled, because it requires
-- Superuser privileges in PostgreSQL
-- CREATE EXTENSION IF NOT EXISTS vector;

-- This creates the users table.

CREATE TABLE IF NOT EXISTS users(
  "handle" VARCHAR(20) PRIMARY KEY,
  "name" TEXT,
  "email" TEXT UNIQUE NOT NULL,
  "vdb_api_key" CHAR(32) UNIQUE NOT NULL,
  "created_at" TIMESTAMP NOT NULL,
  "updated_at" TIMESTAMP NOT NULL
);

-- This creates the projects table.

CREATE TABLE IF NOT EXISTS projects(
  "project_id" SERIAL PRIMARY KEY,
  "handle" VARCHAR(20) NOT NULL,
  "owner" VARCHAR(20) NOT NULL REFERENCES "users"("handle") ON DELETE CASCADE,
  "description" TEXT,
  "created_at" TIMESTAMP NOT NULL,
  "updated_at" TIMESTAMP NOT NULL,
  UNIQUE ("handle", "owner")
);

CREATE INDEX IF NOT EXISTS projects_handle ON "projects"("handle");

-- This creates the users_projects associations table.

DO $$ BEGIN
    IF to_regtype('vdb_role') IS NULL THEN
        CREATE TYPE vdb_role AS ENUM ('owner', 'writer', 'reader');
    END IF;
END $$;


CREATE TABLE IF NOT EXISTS users_projects(
  "user_handle" VARCHAR(20) REFERENCES "users"("handle") ON DELETE CASCADE,
  "project_id" SERIAL REFERENCES "projects"("project_id") ON DELETE CASCADE,
  "role" vdb_role NOT NULL,
  "created_at" TIMESTAMP NOT NULL,
  "updated_at" TIMESTAMP NOT NULL,
  PRIMARY KEY ("user_handle", "project_id")
);

-- This creates the LLM Services table.

CREATE TABLE IF NOT EXISTS llmservices(
  "llmservice_id" SERIAL PRIMARY KEY,
  "handle" VARCHAR(20) NOT NULL,
  "owner" VARCHAR(20) NOT NULL REFERENCES "users"("handle") ON DELETE CASCADE,
  "description" TEXT,
  "endpoint" TEXT NOT NULL,
  "api_key" TEXT,
  "api_standard" VARCHAR(20) NOT NULL REFERENCES "api_standards"("handle"),
  "created_at" TIMESTAMP NOT NULL,
  "updated_at" TIMESTAMP NOT NULL,
  UNIQUE ("handle", "owner")
);

CREATE INDEX IF NOT EXISTS llmservices_handle ON "llmservices"("handle");

-- This creates the users_llmservices associations table.

CREATE TABLE IF NOT EXISTS users_llmservices(
  "user" VARCHAR(20) NOT NULL REFERENCES "users"("handle") ON DELETE CASCADE,
  "llmservice" SERIAL NOT NULL REFERENCES "llmservices"("llmservice_id") ON DELETE CASCADE,
  "role" vdb_role NOT NULL,
  "created_at" TIMESTAMP NOT NULL,
  "updated_at" TIMESTAMP NOT NULL,
  PRIMARY KEY ("user", "llmservice")
);

-- This creates the projects_llmservices associations table.

CREATE TABLE IF NOT EXISTS projects_llmservices(
  "project" SERIAL NOT NULL REFERENCES "projects"("project_id") ON DELETE CASCADE,
  "llmservice" SERIAL NOT NULL REFERENCES "llmservices"("llmservice_id") ON DELETE CASCADE,
  "created_at" TIMESTAMP NOT NULL,
  "updated_at" TIMESTAMP NOT NULL,
  PRIMARY KEY ("project", "llmservice")
);

-- This creates the embeddings table.

CREATE TABLE IF NOT EXISTS embeddings(
  "id" SERIAL PRIMARY KEY,
  "text_id" TEXT,
  "embedding" HALFVEC,
  "embedding_dim" INTEGER NOT NULL,
  "llmservice" SERIAL NOT NULL REFERENCES "llmservices"("llmservice_id"),
  "text" TEXT,
  "metadata" JSONB,
  "created_at" TIMESTAMP NOT NULL,
  "updated_at" TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS embeddings_text_id ON "embeddings"("text_id");

-- We will create the index for the vector in a separate schema version
-- CREATE INDEX ON embedding USING hnsw (embedding halfvec_cosine_ops) WITH (m = 16, ef_construction = 128);

-- This creates the api_standards table.

DO $$ BEGIN
    IF to_regtype('key_method') IS NULL THEN
        CREATE TYPE key_method AS ENUM ('auth_bearer', 'body_form', 'query_param', 'custom_header');
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS api_standards(
  "handle" VARCHAR(20) PRIMARY KEY,
  "description" TEXT,
  "key_method" key_method NOT NULL,
  "key_field" VARCHAR(20),
  "vector_size" INTEGER NOT NULL,
  "created_at" TIMESTAMP NOT NULL,
  "updated_at" TIMESTAMP NOT NULL
);

---- create above / drop below ----

-- This removes the users table.

DROP TABLE IF EXISTS users;

-- This removes the projects table.

DROP TABLE IF EXISTS projects;

DROP INDEX IF EXISTS projects_handle;

-- This removes the users_projects associations table.

DROP TABLE IF EXISTS users_projects;

-- This removes the LLM Services table.

DROP TABLE IF EXISTS llmservices;

DROP INDEX IF EXISTS llmservices_handle;

-- This removes the users_llmservices associations table.

DROP TABLE IF EXISTS users_llmservices;

-- This removes the projects_llmservices associations table.

DROP TABLE IF EXISTS projects_llmservices;

-- This removes the embeddings table.

DROP TABLE IF EXISTS embeddings;

DROP INDEX IF EXISTS embeddings_text_id;

-- This removes the api_standards table.

DROP TABLE IF EXISTS api_standards;

-- This removes the key_method enum type.

DROP TYPE IF EXISTS "key_method";

-- This removes the vdb_role enum type.

DROP TYPE IF EXISTS "vdb_role";

-- This removes the vector extension.
-- Again, as we have disabled it above,
-- we disable it here, too.
-- DROP EXTENSION vector;
