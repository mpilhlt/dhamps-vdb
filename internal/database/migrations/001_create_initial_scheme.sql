-- This enables the vector extension.

-- Currently disabled, because it requires
-- Superuser privileges in PostgreSQL
-- CREATE EXTENSION IF NOT EXISTS vector;

-- This creates the users table.

CREATE TABLE IF NOT EXISTS users(
  "user_handle" VARCHAR(20) PRIMARY KEY,
  "name" TEXT,
  "email" TEXT UNIQUE NOT NULL,
  "vdb_key" CHAR(64) UNIQUE NOT NULL,
  "created_at" TIMESTAMP NOT NULL,
  "updated_at" TIMESTAMP NOT NULL
);

-- This creates the projects table.

CREATE TABLE IF NOT EXISTS projects(
  "project_id" SERIAL PRIMARY KEY,
  "project_handle" VARCHAR(20) NOT NULL,
  "owner" VARCHAR(20) NOT NULL REFERENCES "users"("user_handle") ON DELETE CASCADE,
  "description" TEXT,
  "metadata_scheme" TEXT,
  "created_at" TIMESTAMP NOT NULL,
  "updated_at" TIMESTAMP NOT NULL,
  UNIQUE ("owner", "project_handle")
);

-- This creates the users_projects associations table.

CREATE TABLE IF NOT EXISTS vdb_roles(
  "vdb_role" VARCHAR(20) PRIMARY KEY
);

INSERT INTO "vdb_roles"("vdb_role")
VALUES ('owner'), ('editor'), ('reader');

CREATE TABLE IF NOT EXISTS users_projects(
  "user_handle" VARCHAR(20) REFERENCES "users"("user_handle") ON DELETE CASCADE,
  "project_id" SERIAL REFERENCES "projects"("project_id") ON DELETE CASCADE,
  "role" VARCHAR(20) NOT NULL REFERENCES "vdb_roles"("vdb_role"),
  "created_at" TIMESTAMP NOT NULL,
  "updated_at" TIMESTAMP NOT NULL,
  PRIMARY KEY ("user_handle", "project_id")
);

-- This creates the api_standards table (and the key_methods table it presupposes).

CREATE TABLE IF NOT EXISTS key_methods(
  "key_method" VARCHAR(20) PRIMARY KEY
);

INSERT INTO "key_methods"("key_method")
VALUES ('auth_bearer'), ('body_form'), ('query_param'), ('custom_header');

CREATE TABLE IF NOT EXISTS api_standards(
  "api_standard_handle" VARCHAR(20) PRIMARY KEY,
  "description" TEXT,
  "key_method" VARCHAR(20) NOT NULL REFERENCES "key_methods"("key_method"),
  "key_field" VARCHAR(20),
  "created_at" TIMESTAMP NOT NULL,
  "updated_at" TIMESTAMP NOT NULL
);

-- This creates the LLM Services table.

CREATE TABLE IF NOT EXISTS llm_services(
  "llm_service_id" SERIAL PRIMARY KEY,
  "llm_service_handle" VARCHAR(20) NOT NULL,
  "owner" VARCHAR(20) NOT NULL REFERENCES "users"("user_handle") ON DELETE CASCADE,
  "endpoint" TEXT NOT NULL,
  "description" TEXT,
  "api_standard" VARCHAR(20) NOT NULL REFERENCES "api_standards"("api_standard_handle"),
  "model" TEXT NOT NULL,
  "dimensions" INTEGER NOT NULL,
  "created_at" TIMESTAMP NOT NULL,
  "updated_at" TIMESTAMP NOT NULL,
  UNIQUE ("owner", "llm_service_handle")
);

-- This creates the users_llm_services associations table.

CREATE TABLE IF NOT EXISTS users_llm_services(
  "user_handle" VARCHAR(20) NOT NULL REFERENCES "users"("user_handle") ON DELETE CASCADE,
  "llm_service_id" SERIAL NOT NULL REFERENCES "llm_services"("llm_service_id") ON DELETE CASCADE,
  "role" VARCHAR(20) NOT NULL REFERENCES "vdb_roles"("vdb_role"),
  "created_at" TIMESTAMP NOT NULL,
  "updated_at" TIMESTAMP NOT NULL,
  PRIMARY KEY ("user_handle", "llm_service_id")
);

-- This creates the projects_llm_services associations table.

CREATE TABLE IF NOT EXISTS projects_llm_services(
  "project_id" SERIAL NOT NULL REFERENCES "projects"("project_id") ON DELETE CASCADE,
  "llm_service_id" SERIAL NOT NULL REFERENCES "llm_services"("llm_service_id") ON DELETE CASCADE,
  "created_at" TIMESTAMP NOT NULL,
  "updated_at" TIMESTAMP NOT NULL,
  PRIMARY KEY ("project_id", "llm_service_id")
);

-- This creates the embeddings table.

CREATE TABLE IF NOT EXISTS embeddings(
  "embeddings_id" SERIAL PRIMARY KEY,
  "text_id" TEXT,
  "owner" VARCHAR(20) NOT NULL REFERENCES "users"("user_handle") ON DELETE CASCADE,
  "project_id" SERIAL NOT NULL REFERENCES "projects"("project_id") ON DELETE CASCADE,
  "llm_service_id" SERIAL NOT NULL REFERENCES "llm_services"("llm_service_id"),
  "text" TEXT,
  "vector" halfvec NOT NULL,
  "vector_dim" INTEGER NOT NULL,
  "metadata" jsonb,
  "created_at" TIMESTAMP NOT NULL,
  "updated_at" TIMESTAMP NOT NULL,
  UNIQUE ("text_id", "owner", "project_id", "llm_service_id")
);

CREATE INDEX IF NOT EXISTS embeddings_text_id ON "embeddings"("text_id");

-- We will create the index for the vector in a separate schema version
-- CREATE INDEX ON embeddings USING hnsw (vector halfvec_cosine_ops) WITH (m = 16, ef_construction = 128);

---- create above / drop below ----

-- This removes the users table.

DROP TABLE IF EXISTS users;

-- This removes the projects table.

DROP TABLE IF EXISTS projects;

-- This removes the users_projects associations table.

DROP TABLE IF EXISTS users_projects;

DROP TABLE IF EXISTS vdb_roles;

-- This removes the LLM Services table.

DROP TABLE IF EXISTS llm_services;

-- This removes the users_llm_services associations table.

DROP TABLE IF EXISTS users_llm_services;

-- This removes the projects_llm_services associations table.

DROP TABLE IF EXISTS projects_llm_services;

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
