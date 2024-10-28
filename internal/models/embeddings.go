package models

import (
	"net/http"

	"github.com/pgvector/pgvector-go"
)

// Embeddings contains a single document's embeddings record with id, embeddings and possibly more information.
type Embeddings struct {
  TextID              string `json:"id" doc:"Identifier for the document"`
  Vector              pgvector.HalfVector `json:"vector" doc:"Half-precision embeddings vector for the document"`
  VectorDim           int32 `json:"vector_dim" doc:"Dimensionality of the embeddings vector"`
  Llmservice          int32 `json:"llmservice" doc:"ID of the language model service used to generate the embeddings"`
  Text                string `json:"text,omitempty" doc:"Text content of the document"`
  ProjectId           int    `json:"project_id" doc:"Unique project identifier"`
  User                string `json:"user" path:"user" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
  Project             string `json:"project" path:"project" maxLength:"20" minLength:"3" example:"my-gpt-4" doc:"Project handle"`
  // TODO: add metadata handling
  // Metadata map[string]interface{} `json:"metadata,omitempty" doc:"Metadata for the document. E.g. creation year, author name or text genre."`
}

type Embeddingss []Embeddings

func (es Embeddingss) GetIDs() []string {
  var ids []string
  for _, e := range es {
    ids = append(ids, e.TextID)
  }
  return ids
}

// Request and Response structs for the user administration API
// The request structs must be structs with fields for the request path/query/header/cookie parameters and/or body.
// The response structs must be structs with fields for the output headers and body of the operation, if any.

// Put/post project embeddings
// PUT Path: "/embeddings/{user}/{project}/{id}"

type PutProjEmbeddingsRequest struct {
  User       string `json:"user" path:"user" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
  Project    string `json:"project" path:"project" maxLength:"20" minLength:"3" example:"my-gpt-4" doc:"Project handle"`
  ID         string `json:"id" path:"id" maxLength:"200" minLength:"3" example:"https%3A%2F%2Fid.salamanca.school%2Ftexts%2FW0017%3Afrontmatter.1.1%0A" doc:"Document identifier"`
  Body       struct {
    Embeddings Embeddings `json:"embeddings" doc:"Single set of document embeddings"`
  }
}

// POST Path: "/embeddings/{user}/{project}"

type PostProjEmbeddingsRequest struct {
  User       string `json:"user" path:"user" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
  Project    string `json:"project" path:"project" maxLength:"20" minLength:"3" example:"my-gpt-4" doc:"Project handle"`
  Body       struct {
    Embeddings Embeddingss `json:"embeddings" doc:"List of document embeddings"`
  }
}

type UploadProjEmbeddingsResponse struct {
  Header []http.Header `json:"header,omitempty" doc:"Response headers"`
  Body struct {
    IDs       []string `json:"ids" doc:"List of document identifiers"`
  }
}

// Get project embeddings
// Path: "/embeddings/{user}/{project}"

type GetProjEmbeddingsRequest struct {
  User       string `json:"user" path:"user" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
  Project    string `json:"project" path:"project" maxLength:"20" minLength:"3" example:"my-gpt-4" doc:"Project handle"`
  Limit      int `json:"limit,omitempty" query:"limit" minimum:"1" maximum:"200" example:"10" default:"20" doc:"Maximum number of embeddings to return"`
  Offset     int `json:"offset,omitempty" query:"offset" minimum:"0" example:"0" default:"0" doc:"Offset into the list of embeddings"`
}

type GetProjEmbeddingsResponse struct {
  Header []http.Header `json:"header,omitempty" doc:"Response headers"`
  Body struct {
    Embeddings Embeddingss `json:"embeddings" doc:"List of document embeddings"`
  }
}

// Delete project embeddings
// Path: "/embeddings/{user}/{project}"

type DeleteProjEmbeddingsRequest struct {
  User       string `json:"user" path:"user" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
  Project    string `json:"project" path:"project" maxLength:"20" minLength:"3" example:"my-gpt-4" doc:"Project handle"`
}

type DeleteProjEmbeddingsResponse struct {
  Header []http.Header `json:"header,omitempty" doc:"Response headers"`
  Body string `json:"body" doc:"Success message"`
}

// Get document embeddings
// Path: "/embeddings/{user}/{project}/{id}"

type GetDocEmbeddingsRequest struct {
  User       string `json:"user" path:"user" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
  Project    string `json:"project" path:"project" maxLength:"20" minLength:"3" example:"my-gpt-4" doc:"Project handle"`
  ID         string `json:"id" path:"id" maxLength:"200" minLength:"3" example:"https%3A%2F%2Fid.salamanca.school%2Ftexts%2FW0017%3Afrontmatter.1.1%0A" doc:"Document identifier"`
}

type GetDocEmbeddingsResponse struct {
  Header []http.Header `json:"header,omitempty" doc:"Response headers"`
  Body struct {
    Embeddings `json:"body" doc:"Document embeddings"`
  }
}

// Delete document embeddings
// Path: "/embeddings/{user}/{project}/{id}"

type DeleteDocEmbeddingsRequest struct {
  User       string `json:"user" path:"user" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
  Project    string `json:"project" path:"project" maxLength:"20" minLength:"3" example:"my-gpt-4" doc:"Project handle"`
  ID         string `json:"id" path:"id" maxLength:"200" minLength:"3" example:"https%3A%2F%2Fid.salamanca.school%2Ftexts%2FW0017%3Afrontmatter.1.1%0A" doc:"Document identifier"`
}

type DeleteDocEmbeddingsResponse struct {
  Header []http.Header `json:"header,omitempty" doc:"Response headers"`
  Body string `json:"body" doc:"Success message"`
}

