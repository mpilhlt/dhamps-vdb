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
--  "vdb_api_key" = (decode(sqlc.arg(vdb_api_key)::bytea, 'hex')),
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
  "project_handle", "owner", "description", "metadata_scheme", "created_at", "updated_at"
) VALUES (
  $1, $2, $3, $4, NOW(), NOW()
)
ON CONFLICT ("owner", "project_handle") DO UPDATE SET
  "description" = $3,
  "metadata_scheme" = $4,
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


-- name: UpsertLLM :one
INSERT
INTO llm_services (
  "llm_service_handle", "owner", "endpoint", "description", "api_key", "api_standard", "model", "dimensions", "created_at", "updated_at"
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW()
)
ON CONFLICT ("owner", "llm_service_handle") DO UPDATE SET
  "endpoint" = $3,
  "description" = $4,
  "api_key" = $5,
  "api_standard" = $6,
  "model" = $7,
  "dimensions" = $8,
  "updated_at" = NOW()
RETURNING "llm_service_id", "llm_service_handle", "owner";

-- name: DeleteLLM :exec
DELETE
FROM llm_services
WHERE "owner" = $1
AND "llm_service_handle" = $2;

-- name: RetrieveLLM :one
SELECT *
FROM llm_services
WHERE "owner" = $1
AND "llm_service_handle" = $2
LIMIT 1;

-- name: LinkUserToLLM :exec
INSERT
INTO users_llm_services (
  "user_handle", "llm_service_id", "role", "created_at", "updated_at"
) VALUES (
  $1, $2, $3, NOW(), NOW()
)
ON CONFLICT ("user_handle", "llm_service_id") DO UPDATE SET
  "role" = $3,
  "updated_at" = NOW()
RETURNING *;

-- name: LinkProjectToLLM :exec
INSERT
INTO projects_llm_services (
  "project_id", "llm_service_id", "created_at", "updated_at"
) VALUES (
  $1, $2, NOW(), NOW()
)
ON CONFLICT ("project_id", "llm_service_id") DO NOTHING
RETURNING *;

-- name: GetLLMsByProject :many
SELECT llm_services.*
FROM llm_services
JOIN (
  projects_llm_services JOIN projects
  ON projects_llm_services."project_id" = projects."project_id"
)
ON llm_services."llm_service_id" = projects_llm_services."llm_service_id"
WHERE projects."owner" = $1
  AND projects."project_handle" = $2
ORDER BY llm_services."llm_service_handle" ASC LIMIT $3 OFFSET $4;

-- name: GetLLMsByUser :many
SELECT llm_services.*
FROM llm_services
WHERE llm_services."owner" = $1
ORDER BY llm_services."llm_service_handle" ASC LIMIT $2 OFFSET $3;


-- name: UpsertEmbeddings :one
INSERT
INTO embeddings (
  "text_id", "owner", "project_id", "llm_service_id", "text", "vector", "vector_dim", "created_at", "updated_at"
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, NOW(), NOW()
)
ON CONFLICT ("text_id", "owner", "project_id", "llm_service_id") DO UPDATE SET
  "text" = $5,
  "vector" = $6,
  "vector_dim" = $7,
  "updated_at" = NOW()
RETURNING "embeddings_id", "text_id", "owner", "project_id", "llm_service_id";

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
SELECT embeddings.*, projects."project_handle", llm_services."llm_service_handle"
FROM embeddings
JOIN llm_services
ON embeddings."llm_service_id" = llm_services."llm_service_id"
JOIN projects
ON embeddings."project_id" = projects."project_id"
WHERE embeddings."owner" = $1
AND projects."project_handle" = $2
AND embeddings."text_id" = $3
LIMIT 1;

-- name: GetEmbeddingsByProject :many
SELECT embeddings.*, projects."project_handle", llm_services."llm_service_handle"
FROM embeddings
JOIN llm_services
ON llm_services."llm_service_id" = embeddings."llm_service_id"
JOIN projects
ON projects."project_id" = embeddings."project_id"
WHERE embeddings."owner" = $1
AND projects."project_handle" = $2
ORDER BY embeddings."text_id" ASC LIMIT $3 OFFSET $4;


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
SELECT embeddings."embeddings_id", embeddings."text_id", llm_services."owner", llm_services."llm_service_handle"
FROM embeddings
JOIN llm_services
ON embeddings."llm_service_id" = llm_services."llm_service_id"
ORDER BY "vector" <=> $1
LIMIT $2 OFFSET $3;

-- name: GetSimilarsByID :many
SELECT e2."embeddings_id", e2."text_id", 1 - (e1.vector <=> e2.vector) AS cosine_similarity
FROM embeddings e1
CROSS JOIN embeddings e2
WHERE e1."text_id" = $1
  AND e2."embeddings_id" != e1."embeddings_id"
ORDER BY e1.vector <=> e2.vector
LIMIT $2 OFFSET $3;

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
