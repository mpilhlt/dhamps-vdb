package models

import (
  "net/http"

  "github.com/pgvector/pgvector-go"
)

type Vector pgvector.Vector

// Embeddings contains a single document's embeddings record with id, embeddings and possibly more information.
type Embeddings struct {
  ID                  string `json:"id" doc:"Identifier for the document"`
  Text                string `json:"text,omitempty" doc:"Text content of the document"`
  Vector              Vector `json:"vector" doc:"Embeddings for the document"`
  Metadata map[string]string `json:"metadata,omitempty" doc:"Metadata for the document. E.g. creation year, author name or text genre."`
}

type Embeddingss []Embeddings

func (es Embeddingss) GetIDs() []string {
  var ids []string
  for _, e := range es {
    ids = append(ids, e.ID)
  }
  return ids
}

// Request and Response structs for the user administration API
// The request structs must be structs with fields for the request path/query/header/cookie parameters and/or body.
// The response structs must be structs with fields for the output headers and body of the operation, if any.

// Put/post project embeddings request/response
// Path: "/embeddings/{user}/{project}"

type PutProjEmbeddingsRequest struct {
  User       string `json:"user" path:"user" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
  Project    string `json:"project" path:"project" maxLength:"20" minLength:"3" example:"my-gpt-4" doc:"Project handle"`
  Body       struct {
    Embeddings Embeddingss `json:"embeddings" doc:"List of document embeddings"`
  }
}

type PutProjEmbeddingsResponse struct {
  Header []http.Header `json:"header,omitempty" doc:"Response headers"`
  Body struct {
    IDs       []string `json:"ids" doc:"List of document identifiers"`
  }
}

// Get project embeddings request/response
// Path: "/embeddings/{user}/{project}"

type GetProjEmbeddingsRequest struct {
  User       string `json:"user" path:"user" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
  Project    string `json:"project" path:"project" maxLength:"20" minLength:"3" example:"my-gpt-4" doc:"Project handle"`
}

type GetProjEmbeddingsResponse struct {
  Header []http.Header `json:"header,omitempty" doc:"Response headers"`
  Body struct {
    Embeddings []Embeddings `json:"embeddings" doc:"List of document embeddings"`
  }
}

// Delete project embeddings request
// Path: "/embeddings/{user}/{project}"

type DeleteProjEmbeddingsRequest struct {
  User       string `json:"user" path:"user" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
  Project    string `json:"project" path:"project" maxLength:"20" minLength:"3" example:"my-gpt-4" doc:"Project handle"`
}

type DeleteProjEmbeddingsResponse struct {
  Header []http.Header `json:"header,omitempty" doc:"Response headers"`
  Body string `json:"body" doc:"Success message"`
}

// Get document embeddings request
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

// Put/post project embeddings request/response
// Path: "/embeddings/{user}/{project}/{id}"

type PatchDocEmbeddingsRequest struct {
  User       string `json:"user" path:"user" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
  Project    string `json:"project" path:"project" maxLength:"20" minLength:"3" example:"my-gpt-4" doc:"Project handle"`
  ID         string `json:"id" path:"id" maxLength:"200" minLength:"3" example:"https%3A%2F%2Fid.salamanca.school%2Ftexts%2FW0017%3Afrontmatter.1.1%0A" doc:"Document identifier"`
  Body       struct {
    Embeddings Embeddingss `json:"embeddings" doc:"List of document embeddings"`
  }
}

type PatchDocEmbeddingsResponse struct {
  Header []http.Header `json:"header,omitempty" doc:"Response headers"`
  Body struct {
    Embeddings Embeddings `json:"embeddings" doc:"List of document embeddings"`
  }
}

// Delete document embeddings request/response
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

