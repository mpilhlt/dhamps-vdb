-- Generate go code with: sqlc generate

-- name: UpsertUser :one
INSERT INTO users (
  "handle", "name", "email", "vdb_api_key", "created_at", "updated_at"
) VALUES (
  $1, $2, $3, $4, NOW(), NOW()
)
ON CONFLICT ("handle") DO UPDATE SET
  "name" = $2,
  "email" = $3,
  "vdb_api_key" = $4,
  "updated_at" = NOW()
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE "handle" = $1;

-- name: RetrieveUser :one
SELECT * FROM users
WHERE "handle" = $1 LIMIT 1;

-- name: GetUsers :many
SELECT "handle" FROM users ORDER BY "handle" ASC LIMIT $1 OFFSET $2;

-- name: GetUsersByProject :many
SELECT users."handle", users_projects."role"
FROM users JOIN users_projects
ON users."handle" = users_projects."user_handle"
JOIN projects ON users_projects."project_id" = projects."project_id"
WHERE projects."owner" = $1 AND projects."handle" = $2
ORDER BY users."handle" ASC LIMIT $3 OFFSET $4;


-- name: UpsertProject :one
INSERT INTO projects (
  "handle", "owner", "description", "metadata_scheme", "created_at", "updated_at"
) VALUES (
  $1, $2, $3, $4, NOW(), NOW()
)
ON CONFLICT ("handle", "owner") DO UPDATE SET
  "description" = $3,
  "metadata_scheme" = $4,
  "updated_at" = NOW()
RETURNING "project_id", "handle", "owner";

-- name: DeleteProject :exec
DELETE FROM projects
WHERE "owner" = $1
  AND "handle" = $2;

-- name: GetProjectsByUser :many
SELECT projects.*, users_projects."role"
FROM projects JOIN users_projects
ON projects."project_id" = users_projects."project_id"
WHERE users_projects."user_handle" = $1
ORDER BY projects."handle" ASC LIMIT $2 OFFSET $3;

-- name: RetrieveProject :one
SELECT * FROM projects
WHERE "owner" = $1
  AND "handle" = $2
LIMIT 1;

-- name: LinkProjectToUser :one
INSERT INTO users_projects (
  "user_handle", "project_id", "role", "created_at", "updated_at"
) VALUES (
  $1, $2, $3, NOW(), NOW()
)
ON CONFLICT ("user_handle", "project_id") DO UPDATE SET
  "role" = $3,
  "updated_at" = NOW()
RETURNING *;

-- TODO: name: TransferProject :one



-- name: UpsertLLM :one
INSERT INTO llmservices (
  "handle", "owner", "description", "endpoint", "api_key", "api_standard", "created_at", "updated_at"
) VALUES (
  $1, $2, $3, $4, $5, $6, NOW(), NOW()
)
ON CONFLICT ("handle", "owner") DO UPDATE SET
  "description" = $3,
  "endpoint" = $4,
  "api_key" = $5,
  "api_standard" = $6,
  "updated_at" = NOW()
RETURNING "llmservice_id", "handle", "owner";

-- name: DeleteLLM :exec
DELETE FROM llmservices
WHERE "owner" = $1
  AND "handle" = $2;

-- name: RetrieveLLM :one
SELECT * FROM llmservices
WHERE "owner" = $1
  AND "handle" = $2
LIMIT 1;

-- name: LinkUserToLLM :exec
INSERT INTO users_llmservices (
  "user", "llmservice", "role", "created_at", "updated_at"
) VALUES (
  $1, $2, $3, NOW(), NOW()
)
ON CONFLICT ("user", "llmservice") DO UPDATE SET
  "role" = $3,
  "updated_at" = NOW()
RETURNING *;

-- name: LinkProjectToLLM :exec
INSERT INTO projects_llmservices (
  "project", "llmservice", "created_at", "updated_at"
) VALUES (
  $1, $2, NOW(), NOW()
)
ON CONFLICT ("project", "llmservice") DO NOTHING
RETURNING *;

-- name: GetLLMsByProject :many
SELECT llmservices.* FROM llmservices
JOIN (
  projects_llmservices JOIN projects
  ON projects_llmservices."project" = projects."project_id"
)
ON llmservices."llmservice_id" = projects_llmservices."llmservice"
WHERE projects."owner" = $1
  AND projects."handle" = $2
ORDER BY llmservices."handle" ASC LIMIT $3 OFFSET $4;

-- name: GetLLMsByUser :many
SELECT llmservices.* FROM llmservices
JOIN (
  projects_llmservices JOIN users_projects
  ON projects_llmservices."project" = users_projects."project_id"
)
ON llmservices."llmservice_id" = projects_llmservices."llmservice"
WHERE users_projects."user_handle" = $1
ORDER BY llmservices."handle" ASC LIMIT $2 OFFSET $3;


-- TODO: Add metadata field
-- name: UpsertEmbeddings :one
INSERT INTO embeddings (
  "id", "owner", "project", "text_id", "embedding", "embedding_dim", "llmservice", "text", "created_at", "updated_at"
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $9, NOW(), NOW()
)
ON CONFLICT ("id") DO UPDATE SET
  "text_id" = $2,
  "owner" = $3,
  "project" = $4,
  "embedding" = $5,
  "embedding_dim" = $6,
  "llmservice" = $7,
  "text" = $8,
  "updated_at" = NOW()
RETURNING "id", "text_id";

-- name: DeleteEmbeddingsByID :exec
DELETE FROM embeddings
WHERE "id" = $1;

-- name: DeleteEmbeddingsByProject :exec
DELETE FROM embeddings
WHERE "owner" = $1
  AND "project" = $2;

-- name: DeleteDocEmbeddings :exec
DELETE FROM embeddings
WHERE "owner" = $1
  AND "project" = $2
  AND "text_id" = $3;

-- name: RetrieveEmbeddings :one
SELECT embeddings.*, projects."handle" AS "project", llmservices."handle"
FROM embeddings
JOIN llmservices
  ON embeddings."llmservice" = llmservices."llmservice_id"
JOIN projects
  ON projects."project_id" = embeddings."project"
WHERE embeddings."owner" = $1
  AND "project" = $2
  AND embeddings."text_id" = $3
LIMIT 1;

-- name: GetEmbeddingsByProject :many
SELECT embeddings.*, projects."handle" AS "project", llmservices."handle" AS "llmservice"
FROM embeddings
JOIN llmservices
  ON llmservices."llmservice_id" = embeddings."llmservice"
JOIN projects
  ON projects."project_id" = embeddings."project"
WHERE embeddings."owner" = $1
  AND "project" = $2
ORDER BY embeddings."text_id" ASC LIMIT $3 OFFSET $4;


-- name: UpsertAPI :one
INSERT INTO api_standards (
  "handle", "description", "key_method", "key_field", "vector_size", "created_at", "updated_at"
) VALUES (
  $1, $2, $3, $4, $5, NOW(), NOW()
)
ON CONFLICT ("handle") DO UPDATE SET
  "description" = $2,
  "key_method" = $3,
  "key_field" = $4,
  "vector_size" = $5,
  "updated_at" = NOW()
RETURNING "handle";

-- name: DeleteAPI :exec
DELETE FROM api_standards
WHERE "handle" = $1;

-- name: RetrieveAPI :one
SELECT * FROM api_standards
WHERE "handle" = $1 LIMIT 1;

-- name: GetAPIs :many
SELECT * FROM api_standards
ORDER BY "handle" ASC LIMIT $1 OFFSET $2;



-- name: GetSimilarsByVector :many
SELECT embeddings."id", embeddings."text_id", llmservices."owner", llmservices."handle"
FROM embeddings JOIN llmservices
ON embeddings."llmservice" = llmservices."llmservice_id"
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
