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
INTO llmservices (
  "llmservice_handle", "owner", "description", "endpoint", "api_key", "api_standard", "created_at", "updated_at"
) VALUES (
  $1, $2, $3, $4, $5, $6, NOW(), NOW()
)
ON CONFLICT ("owner", "llmservice_handle") DO UPDATE SET
  "description" = $3,
  "endpoint" = $4,
  "api_key" = $5,
  "api_standard" = $6,
  "updated_at" = NOW()
RETURNING "llmservice_id", "llmservice_handle", "owner";

-- name: DeleteLLM :exec
DELETE
FROM llmservices
WHERE "owner" = $1
AND "llmservice_handle" = $2;

-- name: RetrieveLLM :one
SELECT *
FROM llmservices
WHERE "owner" = $1
AND "llmservice_handle" = $2
LIMIT 1;

-- name: LinkUserToLLM :exec
INSERT
INTO users_llmservices (
  "user_handle", "llmservice_id", "role", "created_at", "updated_at"
) VALUES (
  $1, $2, $3, NOW(), NOW()
)
ON CONFLICT ("user_handle", "llmservice_id") DO UPDATE SET
  "role" = $3,
  "updated_at" = NOW()
RETURNING *;

-- name: LinkProjectToLLM :exec
INSERT
INTO projects_llmservices (
  "project_id", "llmservice_id", "created_at", "updated_at"
) VALUES (
  $1, $2, NOW(), NOW()
)
ON CONFLICT ("project_id", "llmservice_id") DO NOTHING
RETURNING *;

-- name: GetLLMsByProject :many
SELECT llmservices.*
FROM llmservices
JOIN (
  projects_llmservices JOIN projects
  ON projects_llmservices."project_id" = projects."project_id"
)
ON llmservices."llmservice_id" = projects_llmservices."llmservice_id"
WHERE projects."owner" = $1
  AND projects."project_handle" = $2
ORDER BY llmservices."llmservice_handle" ASC LIMIT $3 OFFSET $4;

-- name: GetLLMsByUser :many
SELECT llmservices.*
FROM llmservices
JOIN (
  projects_llmservices JOIN users_projects
  ON projects_llmservices."project_id" = users_projects."project_id"
)
ON llmservices."llmservice_id" = projects_llmservices."llmservice_id"
WHERE users_projects."user_handle" = $1
ORDER BY llmservices."llmservice_handle" ASC LIMIT $2 OFFSET $3;


-- name: UpsertEmbeddings :one
INSERT
INTO embeddings (
  "id", "owner", "project_id", "text_id", "embedding", "embedding_dim", "llmservice_id", "text", "created_at", "updated_at"
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $9, NOW(), NOW()
)
ON CONFLICT ("id") DO UPDATE SET
  "text_id" = $2,
  "owner" = $3,
  "project_id" = $4,
  "embedding" = $5,
  "embedding_dim" = $6,
  "llmservice_id" = $7,
  "text" = $8,
  "updated_at" = NOW()
RETURNING "id", "text_id";

-- name: DeleteEmbeddingsByID :exec
DELETE
FROM embeddings
WHERE "id" = $1;

-- name: DeleteEmbeddingsByProject :exec
DELETE
FROM embeddings
USING embeddings AS e
JOIN projects AS p
ON e."project_id" = p."project_id"
WHERE embeddings."owner" = $1
AND embeddings."project_id" = e."project_id"
AND p."project_handle" = $2;

-- DELETE FROM tv_episodes
-- USING tv_episodes AS ed
-- LEFT OUTER JOIN data AS nd ON
--    ed.file_name = nd.file_name AND 
--    ed.path = nd.path
-- WHERE
--    tv_episodes.id = ed.id AND
--    ed.cd_name = 'MediaLibraryDrive' AND nd.cd_name IS NULL;

-- name: DeleteDocEmbeddings :exec
DELETE
FROM embeddings
USING embeddings as e
JOIN projects as p
ON e."project_id" = p."project_id"
WHERE embeddings."owner" = $1
AND embeddings."project_id" = e."project_id"
AND p."project_handle" = $2
AND embeddings."text_id" = $3;

-- name: RetrieveEmbeddings :one
SELECT embeddings.*, projects."project_handle", llmservices."llmservice_handle"
FROM embeddings
JOIN llmservices
ON embeddings."llmservice_id" = llmservices."llmservice_id"
JOIN projects
ON embeddings."project_id" = projects."project_id"
WHERE embeddings."owner" = $1
AND projects."project_handle" = $2
AND embeddings."text_id" = $3
LIMIT 1;

-- name: GetEmbeddingsByProject :many
SELECT embeddings.*, projects."project_handle", llmservices."llmservice_handle"
FROM embeddings
JOIN llmservices
ON llmservices."llmservice_id" = embeddings."llmservice_id"
JOIN projects
ON projects."project_id" = embeddings."project_id"
WHERE embeddings."owner" = $1
AND projects."project_handle" = $2
ORDER BY embeddings."text_id" ASC LIMIT $3 OFFSET $4;


-- name: UpsertAPI :one
INSERT
INTO api_standards (
  "api_standard_handle", "description", "key_method", "key_field", "vector_size", "created_at", "updated_at"
) VALUES (
  $1, $2, $3, $4, $5, NOW(), NOW()
)
ON CONFLICT ("api_standard_handle") DO UPDATE SET
  "description" = $2,
  "key_method" = $3,
  "key_field" = $4,
  "vector_size" = $5,
  "updated_at" = NOW()
RETURNING "api_standard_handle";

-- name: DeleteAPI :exec
DELETE
FROM api_standards
WHERE "api_standard_handle" = $1;

-- name: RetrieveAPI :one
SELECT *
FROM api_standards
WHERE "api_standard_handle" = $1 LIMIT 1;

-- name: GetAPIs :many
SELECT *
FROM api_standards
ORDER BY "api_standard_handle" ASC LIMIT $1 OFFSET $2;



-- name: GetSimilarsByVector :many
SELECT embeddings."id", embeddings."text_id", llmservices."owner", llmservices."llmservice_handle"
FROM embeddings
JOIN llmservices
ON embeddings."llmservice_id" = llmservices."llmservice_id"
ORDER BY "embedding" <=> $1
LIMIT $2 OFFSET $3;

-- name: GetSimilarsByID :many
SELECT e2."id", e2."text_id", 1 - (e1.embedding <=> e2.embedding) AS cosine_similarity
FROM embeddings e1
CROSS JOIN embeddings e2
WHERE e1."text_id" = $1
  AND e2."id" != e1."id"
ORDER BY e1.embedding <=> e2.embedding
LIMIT $2 OFFSET $3;
