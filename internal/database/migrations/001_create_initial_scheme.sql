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
  "metadata_scheme" TEXT,
  "created_at" TIMESTAMP NOT NULL,
  "updated_at" TIMESTAMP NOT NULL,
  UNIQUE ("handle", "owner")
);

CREATE INDEX IF NOT EXISTS projects_handle ON "projects"("handle");

-- This creates the users_projects associations table.

CREATE TABLE IF NOT EXISTS vdb_roles(
  "vdb_role" VARCHAR(20) PRIMARY KEY
);

INSERT INTO "vdb_roles"("vdb_role")
VALUES ('owner'), ('writer'), ('reader');

CREATE TABLE IF NOT EXISTS users_projects(
  "user_handle" VARCHAR(20) REFERENCES "users"("handle") ON DELETE CASCADE,
  "project_id" SERIAL REFERENCES "projects"("project_id") ON DELETE CASCADE,
  "role" VARCHAR(20) NOT NULL REFERENCES "vdb_roles"("vdb_role"),
  "created_at" TIMESTAMP NOT NULL,
  "updated_at" TIMESTAMP NOT NULL,
  PRIMARY KEY ("user_handle", "project_id")
);

-- This creates the LLM Services table.

-- This creates the api_standards table.

CREATE TABLE IF NOT EXISTS key_methods(
  "key_method" VARCHAR(20) PRIMARY KEY
);

INSERT INTO "key_methods"("key_method")
VALUES ('auth_bearer'), ('body_form'), ('query_param'), ('custom_header');

CREATE TABLE IF NOT EXISTS api_standards(
  "handle" VARCHAR(20) PRIMARY KEY,
  "description" TEXT,
  "key_method" VARCHAR(20) NOT NULL REFERENCES "key_methods"("key_method"),
  "key_field" VARCHAR(20),
  "vector_size" INTEGER NOT NULL,
  "created_at" TIMESTAMP NOT NULL,
  "updated_at" TIMESTAMP NOT NULL
);

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
  "role" VARCHAR(20) NOT NULL REFERENCES "vdb_roles"("vdb_role"),
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
  "owner" VARCHAR(20) NOT NULL REFERENCES "users"("handle") ON DELETE CASCADE,
  "project" SERIAL NOT NULL REFERENCES "projects"("project_id") ON DELETE CASCADE,
  "text_id" TEXT,
  "embedding" halfvec NOT NULL,
  "embedding_dim" INTEGER NOT NULL,
  "llmservice" SERIAL NOT NULL REFERENCES "llmservices"("llmservice_id"),
  "text" TEXT,
  -- TODO: add metadata handling
  -- "metadata" jsonb,
  "created_at" TIMESTAMP NOT NULL,
  "updated_at" TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS embeddings_text_id ON "embeddings"("text_id");

-- We will create the index for the vector in a separate schema version
-- CREATE INDEX ON embedding USING hnsw (embedding halfvec_cosine_ops) WITH (m = 16, ef_construction = 128);


---- create above / drop below ----

-- This removes the users table.

DROP TABLE IF EXISTS users;

-- This removes the projects table.

DROP TABLE IF EXISTS projects;

DROP INDEX IF EXISTS projects_handle;

-- This removes the users_projects associations table.

DROP TABLE IF EXISTS users_projects;

DROP TABLE IF EXISTS vdb_roles;

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

DROP TABLE IF EXISTS key_methods;

DROP TABLE IF EXISTS api_standards;

-- This removes the vector extension.
-- Again, as we have disabled it above,
-- we disable it here, too.
-- DROP EXTENSION vector;
