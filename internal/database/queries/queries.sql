-- name: UploadUser :one
INSERT INTO users (
  "handle", "name", "email", "vdb_api_key", "created_at", "updated_at"
) VALUES (
  $1, $2, $3, $4, NOW(), NOW()
)
RETURNING *;

-- name: UpdateUser :one
UPDATE users
  SET "name" = $2,
  "email" = $3,
  "vdb_api_key" = $4,
  "created_at" = $5,
  "updated_at" = NOW()
WHERE "handle" = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE "handle" = $1;

-- name: RetrieveUser :one
SELECT * FROM users
WHERE "handle" = $1 LIMIT 1;

-- name: GetUsers :many
SELECT "handle" FROM users;


-- name: UploadProject :one
INSERT INTO projects (
  "handle", "owner", "description", "created_at", "updated_at"
) VALUES (
  $1, $2, $3, NOW(), NOW()
)
RETURNING "project_id", "handle", "owner";

-- name: UpdateProject :one
UPDATE projects
SET "description" = $3,
    "created_at" = $4,
    "updated_at" = NOW()
WHERE "owner" = $1
  AND "handle" = $2
RETURNING "project_id", "handle", "owner";

-- name: DeleteProject :exec
DELETE FROM projects
WHERE "owner" = $1
  AND "handle" = $2;

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
RETURNING *;

-- TODO: name: TransferProject :one

-- name: GetProjectsByUser :many
SELECT projects.*, users_projects."role"
FROM projects JOIN users_projects
ON projects."project_id" = users_projects."project_id"
WHERE users_projects."user_handle" = $1;


-- name: UploadLLM :exec
INSERT INTO llmservices (
  "llmservice_id", "handle", "owner", "description", "endpoint", "api_key", "api_standard", "created_at", "updated_at"
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, NOW(), NOW()
)
RETURNING "llmservice_id", "handle", "owner";

-- name: UpdateLLM :one
UPDATE llmservices
  SET "handle" = $2,
  "description" = $3,
  "endpoint" = $4,
  "api_key" = $5,
  "api_standard" = $6,
  "created_at" = $7,
  "updated_at" = NOW()
WHERE "owner" = $1
  AND "handle" = $2
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
); 

-- name: LinkProjectToLLM :exec
INSERT INTO projects_llmservices (
  "project", "llmservice", "created_at", "updated_at"
) VALUES (
  $1, $2, NOW(), NOW()
);

-- name: GetLLMsByProject :many
SELECT llmservices.* FROM llmservices
JOIN (
  projects_llmservices JOIN projects
  ON projects_llmservices."project" = projects."project_id"
)
ON llmservices."llmservice_id" = projects_llmservices."llmservice"
WHERE projects."owner" = $1
  AND projects."handle" = $2;

-- name: GetLLMsByUser :many
SELECT llmservices.* FROM llmservices
JOIN (
  projects_llmservices JOIN users_projects
  ON projects_llmservices."project" = users_projects."project_id"
)
ON llmservices."llmservice_id" = projects_llmservices."llmservice"
WHERE users_projects."user_handle" = $1;


-- name: UploadEmbeddings :one
INSERT INTO embeddings (
  "text_id", "embedding", "embedding_dim", "llmservice", "text", "metadata", "created_at", "updated_at"
) VALUES (
  $1, $2, $3, $4, $5, $6, NOW(), NOW()
)
RETURNING "id";

-- name: UpdateEmbeddings :one
UPDATE embeddings
  SET "text_id" = $2,
  "embedding" = $3,
  "embedding_dim" = $4,
  "llmservice" = $5,
  "text" = $6,
  "metadata" = $7,
  "created_at" = $8,
  "updated_at" = NOW()
WHERE "id" = $1
RETURNING "id";

-- name: DeleteEmbeddings :exec
DELETE FROM embeddings
WHERE "id" = $1;

-- name: RetrieveEmbeddings :one
SELECT embeddings.*, llmservices."owner", llmservices."handle"
FROM embeddings JOIN llmservices
ON embeddings."llmservice" = llmservices."llmservice_id"
WHERE "id" = $1 LIMIT 1;

-- name: GetEmbeddingsByProject :many
SELECT embeddings.*, llmservices."owner", llmservices."handle"
FROM embeddings JOIN llmservices
ON embeddings."llmservice" = llmservices."llmservice_id"
JOIN projects_llmservices
ON embeddings."llmservice" = projects_llmservices."llmservice"
JOIN projects
ON projects_llmservices."project" = projects."project_id"
WHERE projects."owner" = $1
  AND projects."handle" = $2;


-- name: UploadAPI :one
INSERT INTO api_standards (
  "handle", "description", "key_method", "key_field", "vector_size", "created_at", "updated_at"
) VALUES (
  $1, $2, $3, $4, $5, NOW(), NOW()
)
RETURNING "handle";

-- name: UpdateAPI :one
UPDATE api_standards
  SET "description" = $2,
  "key_method" = $3,
  "key_field" = $4,
  "vector_size" = $5,
  "created_at" = $6,
  "updated_at" = NOW()
WHERE "handle" = $1
RETURNING "handle";

-- name: DeleteAPI :exec
DELETE FROM api_standards
WHERE "handle" = $1;

-- name: RetrieveAPI :one
SELECT * FROM api_standards
WHERE "handle" = $1 LIMIT 1;

-- name: GetAPIs :many
SELECT * FROM api_standards;



-- name: GetSimilarsByVector :many
SELECT embeddings."id", embeddings."text_id", llmservices."owner", llmservices."handle"
FROM embeddings JOIN llmservices
ON embeddings."llmservice" = llmservices."llmservice_id"
ORDER BY "embedding" <#> $1 LIMIT 5;

