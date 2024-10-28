// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package database

import (
	"github.com/jackc/pgx/v5/pgtype"
	pgvector_go "github.com/pgvector/pgvector-go"
)

type ApiStandard struct {
	Handle      string           `db:"handle" json:"handle"`
	Description pgtype.Text      `db:"description" json:"description"`
	KeyMethod   string           `db:"key_method" json:"key_method"`
	KeyField    pgtype.Text      `db:"key_field" json:"key_field"`
	VectorSize  int32            `db:"vector_size" json:"vector_size"`
	CreatedAt   pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt   pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type Embedding struct {
	ID           int32                  `db:"id" json:"id"`
	Owner        string                 `db:"owner" json:"owner"`
	Project      int32                  `db:"project" json:"project"`
	TextID       pgtype.Text            `db:"text_id" json:"text_id"`
	Embedding    pgvector_go.HalfVector `db:"embedding" json:"embedding"`
	EmbeddingDim int32                  `db:"embedding_dim" json:"embedding_dim"`
	Llmservice   int32                  `db:"llmservice" json:"llmservice"`
	Text         pgtype.Text            `db:"text" json:"text"`
	CreatedAt    pgtype.Timestamp       `db:"created_at" json:"created_at"`
	UpdatedAt    pgtype.Timestamp       `db:"updated_at" json:"updated_at"`
}

type KeyMethod struct {
	KeyMethod string `db:"key_method" json:"key_method"`
}

type Llmservice struct {
	LlmserviceID int32            `db:"llmservice_id" json:"llmservice_id"`
	Handle       string           `db:"handle" json:"handle"`
	Owner        string           `db:"owner" json:"owner"`
	Description  pgtype.Text      `db:"description" json:"description"`
	Endpoint     string           `db:"endpoint" json:"endpoint"`
	ApiKey       pgtype.Text      `db:"api_key" json:"api_key"`
	ApiStandard  string           `db:"api_standard" json:"api_standard"`
	CreatedAt    pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt    pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type Project struct {
	ProjectID      int32            `db:"project_id" json:"project_id"`
	Handle         string           `db:"handle" json:"handle"`
	Owner          string           `db:"owner" json:"owner"`
	Description    pgtype.Text      `db:"description" json:"description"`
	MetadataScheme pgtype.Text      `db:"metadata_scheme" json:"metadata_scheme"`
	CreatedAt      pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt      pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type ProjectsLlmservice struct {
	Project    int32            `db:"project" json:"project"`
	Llmservice int32            `db:"llmservice" json:"llmservice"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type User struct {
	Handle    string           `db:"handle" json:"handle"`
	Name      pgtype.Text      `db:"name" json:"name"`
	Email     string           `db:"email" json:"email"`
	VdbApiKey string           `db:"vdb_api_key" json:"vdb_api_key"`
	CreatedAt pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type UsersLlmservice struct {
	User       string           `db:"user" json:"user"`
	Llmservice int32            `db:"llmservice" json:"llmservice"`
	Role       string           `db:"role" json:"role"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type UsersProject struct {
	UserHandle string           `db:"user_handle" json:"user_handle"`
	ProjectID  int32            `db:"project_id" json:"project_id"`
	Role       string           `db:"role" json:"role"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type VdbRole struct {
	VdbRole string `db:"vdb_role" json:"vdb_role"`
}