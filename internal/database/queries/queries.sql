-- Generate go code with: sqlc generate

-- name: UpsertUser :one
INSERT
INTO users (
  "user_handle", "name", "email", "vdb_api_key", "created_at", "updated_at"
) VALUES (
--  $1, $2, $3, (decode(sqlc.arg(vdb_api_key)::bytea, 'hex')), NOW(), NOW()
  $1, $2, $3, $4, NOW(), NOW()
)
ON CONFLICT ("user_handle") DO UPDATE SET
  "name" = $2,
  "email" = $3,
  "vdb_api_key" = $4,
  "updated_at" = NOW()
RETURNING *;

-- name: DeleteUser :exec
DELETE
FROM users
WHERE "user_handle" = $1;

-- name: RetrieveUser :one
SELECT *
FROM users
WHERE "user_handle" = $1 LIMIT 1;

-- name: GetUsers :many
SELECT "user_handle"
FROM users
ORDER BY "user_handle" ASC LIMIT $1 OFFSET $2;

-- name: GetUsersByProject :many
SELECT users."user_handle", users_projects."role"
FROM users JOIN users_projects
ON users."user_handle" = users_projects."user_handle"
JOIN projects ON users_projects."project_id" = projects."project_id"
WHERE projects."owner" = $1 AND projects."project_handle" = $2
ORDER BY users."user_handle" ASC LIMIT $3 OFFSET $4;

-- name: GetKeyByUser :one
SELECT "vdb_api_key"
FROM users
-- SELECT encode("vdb_api_key", 'hex') AS "vdb_api_key" FROM users
WHERE "user_handle" = $1 LIMIT 1;

-- name: GetKeysByLinkedUsers :many
SELECT users."user_handle", users_projects."role", users."vdb_api_key"
-- SELECT users."user_handle", users_projects."role", encode(users."vdb_api_key", 'hex') AS "vdb_api_key"
FROM users
JOIN users_projects
ON users."user_handle" = users_projects."user_handle"
JOIN projects
ON users_projects."project_id" = projects."project_id"
WHERE projects."owner" = $1
AND projects."project_handle" = $2
ORDER BY users."user_handle" ASC LIMIT $3 OFFSET $4;

-- name: UpsertProject :one
INSERT
INTO projects (
  "project_handle", "owner", "description", "metadata_scheme", "public_read", "llm_service_instance_id", "created_at", "updated_at"
) VALUES (
  $1, $2, $3, $4, $5, $6, NOW(), NOW()
)
ON CONFLICT ("owner", "project_handle") DO UPDATE SET
  "description" = $3,
  "metadata_scheme" = $4,
  "public_read" = $5,
  "llm_service_instance_id" = $6,
  "updated_at" = NOW()
RETURNING "project_id", "owner", "project_handle";

-- name: DeleteProject :exec
DELETE
FROM projects
WHERE "owner" = $1
AND "project_handle" = $2;

-- name: GetProjectsByUser :many
SELECT projects.*, users_projects."role"
FROM projects
JOIN users_projects
ON projects."project_id" = users_projects."project_id"
WHERE users_projects."user_handle" = $1
ORDER BY projects."project_handle" ASC LIMIT $2 OFFSET $3;

-- name: RetrieveProject :one
SELECT *
FROM projects
WHERE "owner" = $1
AND "project_handle" = $2
LIMIT 1;

-- name: IsProjectPubliclyReadable :one
SELECT "public_read"
FROM projects
WHERE "owner" = $1
AND "project_handle" = $2
LIMIT 1;

-- name: GetAllProjects :many
SELECT *
FROM projects
ORDER BY "owner" ASC, "project_handle" ASC;

-- name: LinkProjectToUser :one
INSERT
INTO users_projects (
  "user_handle", "project_id", "role", "created_at", "updated_at"
) VALUES (
  $1, $2, $3, NOW(), NOW()
)
ON CONFLICT ("user_handle", "project_id") DO UPDATE SET
  "role" = $3,
  "updated_at" = NOW()
RETURNING *;


-- LLM Service Definitions (templates that can be shared)

-- name: UpsertLLMDefinition :one
INSERT
INTO llm_service_definitions (
  "owner", "definition_handle", "endpoint", "description", "api_standard", "model", "dimensions", "created_at", "updated_at"
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, NOW(), NOW()
)
ON CONFLICT ("owner", "definition_handle") DO UPDATE SET
  "endpoint" = $3,
  "description" = $4,
  "api_standard" = $5,
  "model" = $6,
  "dimensions" = $7,
  "updated_at" = NOW()
RETURNING "owner", "definition_handle", "definition_id";

-- name: DeleteLLMDefinition :exec
DELETE
FROM llm_service_definitions
WHERE "owner" = $1
AND "definition_handle" = $2;

-- name: RetrieveLLMDefinition :one
SELECT *
FROM llm_service_definitions
WHERE "owner" = $1
AND "definition_handle" = $2
LIMIT 1;

-- name: GetLLMDefinitionsByUser :many
SELECT *
FROM llm_service_definitions
WHERE "owner" = $1
ORDER BY "definition_handle" ASC LIMIT $2 OFFSET $3;

-- name: GetAllLLMDefinitions :many
SELECT *
FROM llm_service_definitions
ORDER BY "owner" ASC, "definition_handle" ASC LIMIT $1 OFFSET $2;

-- name: GetSystemLLMDefinitions :many
SELECT *
FROM llm_service_definitions
WHERE "owner" = '_system'
ORDER BY "definition_handle" ASC LIMIT $1 OFFSET $2;

-- LLM Service Instances (user-specific instances with optional API keys)

-- name: UpsertLLMInstance :one
INSERT
INTO llm_service_instances (
  "owner", "instance_handle", "definition_id", "endpoint", "description", "api_key", "api_key_encrypted", "api_standard", "model", "dimensions", "created_at", "updated_at"
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW()
)
ON CONFLICT ("owner", "instance_handle") DO UPDATE SET
  "definition_id" = $3,
  "endpoint" = $4,
  "description" = $5,
  "api_key" = $6,
  "api_key_encrypted" = $7,
  "api_standard" = $8,
  "model" = $9,
  "dimensions" = $10,
  "updated_at" = NOW()
RETURNING "owner", "instance_handle", "instance_id";

-- name: CreateLLMInstanceFromDefinition :one
-- Create an instance based on a definition (copies definition fields, allows user to specify API key)
INSERT
INTO llm_service_instances (
  "owner", "instance_handle", "definition_id", "endpoint", "description", "api_key", "api_key_encrypted", "api_standard", "model", "dimensions", "created_at", "updated_at"
)
SELECT 
  $1 as "owner",
  $2 as "instance_handle", 
  def."definition_id",
  COALESCE($3, def."endpoint") as "endpoint",
  COALESCE($4, def."description") as "description",
  $5 as "api_key",
  $6 as "api_key_encrypted",
  COALESCE($7, def."api_standard") as "api_standard",
  COALESCE($8, def."model") as "model",
  COALESCE($9::INTEGER, def."dimensions") as "dimensions",
  NOW() as "created_at",
  NOW() as "updated_at"
FROM llm_service_definitions def
WHERE def."owner" = $10 AND def."definition_handle" = $11
ON CONFLICT ("owner", "instance_handle") DO UPDATE SET
  "definition_id" = EXCLUDED."definition_id",
  "endpoint" = EXCLUDED."endpoint",
  "description" = EXCLUDED."description",
  "api_key" = EXCLUDED."api_key",
  "api_key_encrypted" = EXCLUDED."api_key_encrypted",
  "api_standard" = EXCLUDED."api_standard",
  "model" = EXCLUDED."model",
  "dimensions" = EXCLUDED."dimensions",
  "updated_at" = NOW()
RETURNING "owner", "instance_handle", "instance_id";

-- name: DeleteLLMInstance :exec
DELETE
FROM llm_service_instances
WHERE "owner" = $1
AND "instance_handle" = $2;

-- name: RetrieveLLMInstance :one
SELECT *
FROM llm_service_instances
WHERE "owner" = $1
AND "instance_handle" = $2
LIMIT 1;

-- name: RetrieveLLMInstanceByOwnerHandle :one
-- Get instance by owner/handle format for shared instances
SELECT *
FROM llm_service_instances
WHERE ("owner" = $1 AND "instance_handle" = $2)
LIMIT 1;

-- name: RetrieveLLMInstanceByID :one
SELECT *
FROM llm_service_instances
WHERE "instance_id" = $1
LIMIT 1;

-- name: LinkUserToLLMInstance :exec
INSERT
INTO users_llm_service_instances (
  "user_handle", "instance_id", "role", "created_at", "updated_at"
) VALUES (
  $1, $2, $3, NOW(), NOW()
)
ON CONFLICT ("user_handle", "instance_id") DO UPDATE SET
  "role" = $3,
  "updated_at" = NOW()
RETURNING *;

-- name: ShareLLMInstance :exec
INSERT
INTO llm_service_instances_shared_with (
  "user_handle", "instance_id", "role", "created_at", "updated_at"
) VALUES (
  $1, $2, $3, NOW(), NOW()
)
ON CONFLICT ("user_handle", "instance_id") DO UPDATE SET
  "role" = $3,
  "updated_at" = NOW();

-- name: UnshareLLMInstance :exec
DELETE
FROM llm_service_instances_shared_with
WHERE "user_handle" = $1
AND "instance_id" = $2;

-- name: GetSharedUsersForInstance :many
SELECT "user_handle", "role"
FROM llm_service_instances_shared_with
WHERE "instance_id" = $1
ORDER BY "user_handle" ASC;

-- name: GetLLMInstanceByProject :one
SELECT llm_service_instances.*
FROM llm_service_instances
JOIN projects
ON projects."llm_service_instance_id" = llm_service_instances."instance_id"
WHERE projects."owner" = $1
  AND projects."project_handle" = $2
LIMIT 1;

-- name: GetLLMInstancesByUser :many
SELECT llm_service_instances.*, users_llm_service_instances."role"
FROM llm_service_instances
JOIN users_llm_service_instances
ON llm_service_instances."instance_id" = users_llm_service_instances."instance_id"
WHERE users_llm_service_instances."user_handle" = $1
ORDER BY llm_service_instances."instance_handle" ASC LIMIT $2 OFFSET $3;

-- name: GetSharedLLMInstances :many
SELECT llm_service_instances.*, llm_service_instances_shared_with."role"
FROM llm_service_instances
JOIN llm_service_instances_shared_with
ON llm_service_instances."instance_id" = llm_service_instances_shared_with."instance_id"
WHERE llm_service_instances_shared_with."user_handle" = $1
ORDER BY llm_service_instances_shared_with."role" ASC, llm_service_instances."owner" ASC, llm_service_instances."instance_handle" ASC 
LIMIT $2 OFFSET $3;

-- name: GetAllAccessibleLLMInstances :many
-- Get all instances accessible to a user (owned + shared)
-- Returns instances with metadata indicating ownership
SELECT 
  llm_service_instances.*,
  CASE 
    WHEN llm_service_instances."owner" = $1 THEN 'owner'
    ELSE COALESCE(llm_service_instances_shared_with."role", users_llm_service_instances."role")
  END as "role",
  llm_service_instances."owner" = $1 as "is_owner"
FROM llm_service_instances
LEFT JOIN users_llm_service_instances
  ON llm_service_instances."instance_id" = users_llm_service_instances."instance_id"
  AND users_llm_service_instances."user_handle" = $1
LEFT JOIN llm_service_instances_shared_with
  ON llm_service_instances."instance_id" = llm_service_instances_shared_with."instance_id"
  AND llm_service_instances_shared_with."user_handle" = $1
WHERE llm_service_instances."owner" = $1
   OR users_llm_service_instances."user_handle" = $1
   OR llm_service_instances_shared_with."user_handle" = $1
ORDER BY llm_service_instances."owner" ASC, llm_service_instances."instance_handle" ASC 
LIMIT $2 OFFSET $3;

-- name: UpsertEmbeddings :one
INSERT
INTO embeddings (
  "text_id", "owner", "project_id", "llm_service_instance_id", "text", "vector", "vector_dim", "metadata", "created_at", "updated_at"
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW()
)
ON CONFLICT ("text_id", "owner", "project_id", "llm_service_instance_id") DO UPDATE SET
  "text" = $5,
  "vector" = $6,
  "vector_dim" = $7,
  "metadata" = $8,
  "updated_at" = NOW()
RETURNING "embeddings_id", "text_id", "owner", "project_id", "llm_service_instance_id";

-- name: DeleteEmbeddingsByID :exec
DELETE
FROM embeddings
WHERE "embeddings_id" = $1;

-- name: DeleteEmbeddingsByProject :exec
DELETE
FROM embeddings
USING embeddings AS e
JOIN projects AS p
ON e."project_id" = p."project_id"
WHERE embeddings."owner" = $1
AND embeddings."project_id" = e."project_id"
AND p."project_handle" = $2;

-- name: DeleteDocEmbeddings :exec
DELETE FROM embeddings e
USING projects p
WHERE e."owner" = $1
  AND e."project_id" = p."project_id"
  AND p."project_handle" = $2
  AND e."text_id" = $3;

-- name: RetrieveEmbeddings :one
SELECT embeddings.*, projects."project_handle", llm_service_instances."instance_handle"
FROM embeddings
JOIN llm_service_instances
ON embeddings."llm_service_instance_id" = llm_service_instances."instance_id"
JOIN projects
ON embeddings."project_id" = projects."project_id"
WHERE embeddings."owner" = $1
AND projects."project_handle" = $2
AND embeddings."text_id" = $3
LIMIT 1;

-- name: GetEmbeddingsByProject :many
SELECT embeddings.*, projects."project_handle", llm_service_instances."instance_handle"
FROM embeddings
JOIN llm_service_instances
ON llm_service_instances."instance_id" = embeddings."llm_service_instance_id"
JOIN projects
ON projects."project_id" = embeddings."project_id"
WHERE embeddings."owner" = $1
AND projects."project_handle" = $2
ORDER BY embeddings."text_id" ASC LIMIT $3 OFFSET $4;

-- name: GetNumberOfEmbeddingsByProject :one
SELECT COUNT(*)
FROM embeddings
JOIN projects
ON embeddings."project_id" = projects."project_id"
WHERE embeddings."owner" = $1
AND projects."project_handle" = $2;


-- name: UpsertAPIStandard :one
INSERT
INTO api_standards (
  "api_standard_handle", "description", "key_method", "key_field", "created_at", "updated_at"
) VALUES (
  $1, $2, $3, $4, NOW(), NOW()
)
ON CONFLICT ("api_standard_handle") DO UPDATE SET
  "description" = $2,
  "key_method" = $3,
  "key_field" = $4,
  "updated_at" = NOW()
RETURNING "api_standard_handle";

-- name: DeleteAPIStandard :exec
DELETE
FROM api_standards
WHERE "api_standard_handle" = $1;

-- name: RetrieveAPIStandard :one
SELECT *
FROM api_standards
WHERE "api_standard_handle" = $1 LIMIT 1;

-- name: GetAPIStandards :many
SELECT *
FROM api_standards
ORDER BY "api_standard_handle" ASC LIMIT $1 OFFSET $2;



-- name: GetSimilarsByVector :many
SELECT embeddings."embeddings_id", embeddings."text_id", llm_service_instances."owner", llm_service_instances."instance_handle"
FROM embeddings
JOIN llm_service_instances
ON embeddings."llm_service_instance_id" = llm_service_instances."instance_id"
ORDER BY "vector" <=> $1
LIMIT $2 OFFSET $3;

-- name: GetSimilarsByID :many
SELECT e2."text_id"
FROM embeddings e1
CROSS JOIN embeddings e2
JOIN projects
ON e1."project_id" = projects."project_id"
WHERE e2."embeddings_id" != e1."embeddings_id"
  AND e1."text_id" = $1
  AND e1."owner" = $2
  AND projects."project_handle" = $3
  AND e1."vector_dim" = e2."vector_dim"
  AND e1."project_id" = e2."project_id"
  AND 1 - (e1.vector <=> e2.vector) >= $4::double precision
ORDER BY e1.vector <=> e2.vector
LIMIT $5 OFFSET $6;

-- name: GetSimilarsByIDWithFilter :many
SELECT e2."text_id"
FROM embeddings e1
CROSS JOIN embeddings e2
JOIN projects
ON e1."project_id" = projects."project_id"
WHERE e2."embeddings_id" != e1."embeddings_id"
  AND e1."text_id" = $1
  AND e1."owner" = $2
  AND projects."project_handle" = $3
  AND e1."vector_dim" = e2."vector_dim"
  AND e1."project_id" = e2."project_id"
  AND 1 - (e1.vector <=> e2.vector) >= $4::double precision
  AND (e2."metadata" ->> $5::text IS NULL OR trim(e2."metadata" ->> $5::text) <> trim($6::text))
ORDER BY e1.vector <=> e2.vector
LIMIT $7 OFFSET $8;


-- name: ResetAllSerials :exec
DO $$
DECLARE
    seq_name text;
BEGIN
    FOR seq_name IN
      SELECT sequence_name
      FROM information_schema.sequences
      WHERE sequence_schema = 'public' AND sequence_name LIKE '%_seq'
    LOOP
        EXECUTE format('ALTER SEQUENCE public.%I RESTART WITH 1', seq_name);
    END LOOP;
END $$;

-- name: DeleteAllRecords :exec
DO $$
DECLARE
    r RECORD;
BEGIN
    FOR r IN
        SELECT table_name 
        FROM information_schema.tables 
        WHERE table_schema = 'public'
          AND table_name NOT IN ('key_methods', 'vdb_roles')
    LOOP
        EXECUTE format('DELETE FROM %I;', r.table_name);
    END LOOP;
END $$;
