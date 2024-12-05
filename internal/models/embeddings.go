package models

import (
	"net/http"
)

// Embeddings contains a single document's embeddings record with id, embeddings and possibly more information.
type Embeddings struct {
	TextID        string    `json:"text_id" doc:"Identifier for the document"`
	Vector        []float32 `json:"vector" doc:"Half-precision embeddings vector for the document"`
	VectorDim     int32     `json:"vector_dim" doc:"Dimensionality of the embeddings vector"`
	LLMServiceID  int32     `json:"llm_service_id" doc:"ID of the language model service used to generate the embeddings"`
	Text          string    `json:"text,omitempty" doc:"Text content of the document"`
	ProjectID     int       `json:"project_id,omitempty" doc:"Unique project identifier"`
	UserHandle    string    `json:"user_handle" path:"user_handle" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
	ProjectHandle string    `json:"project_handle" path:"project_handle" maxLength:"20" minLength:"3" example:"my-gpt-4" doc:"Project handle"`
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
// PUT Path: "/embeddings/{user_handle}/{project_handle}/{text_id}"

type PutProjEmbeddingsRequest struct {
	TextID        string `json:"text_id" path:"id" maxLength:"300" minLength:"3" example:"https%3A%2F%2Fid.salamanca.school%2Ftexts%2FW0017%3Afrontmatter.1.1%0A" doc:"Document identifier"`
	UserHandle    string `json:"user_handle" path:"user_handle" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
	ProjectHandle string `json:"project_handle" path:"project_handle" maxLength:"20" minLength:"3" example:"my-gpt-4" doc:"Project handle"`
	Body          Embeddings
}

// POST Path: "/embeddings/{user_handle}/{project_handle}"

type PostProjEmbeddingsRequest struct {
	UserHandle    string `json:"user_handle" path:"user_handle" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
	ProjectHandle string `json:"project_handle" path:"project_handle" maxLength:"20" minLength:"3" example:"my-gpt-4" doc:"Project handle"`
	Body          struct {
		Embeddings Embeddingss `json:"embeddings" doc:"List of document embeddings"`
	}
}

type UploadProjEmbeddingsResponse struct {
	Header []http.Header `json:"header,omitempty" doc:"Response headers"`
	Body   struct {
		IDs []string `json:"ids" doc:"List of document identifiers"`
	}
}

// Get project embeddings
// Path: "/embeddings/{user_handle}/{project_handle}"

type GetProjEmbeddingsRequest struct {
	UserHandle    string `json:"user_handle" path:"user_handle" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
	ProjectHandle string `json:"project_handle" path:"project_handle" maxLength:"20" minLength:"3" example:"my-gpt-4" doc:"Project handle"`
	Limit         int    `json:"limit,omitempty" query:"limit" minimum:"1" maximum:"200" example:"10" default:"20" doc:"Maximum number of embeddings to return"`
	Offset        int    `json:"offset,omitempty" query:"offset" minimum:"0" example:"0" default:"0" doc:"Offset into the list of embeddings"`
}

type GetProjEmbeddingsResponse struct {
	Header []http.Header `json:"header,omitempty" doc:"Response headers"`
	Body   struct {
		Embeddings Embeddingss `json:"embeddings" doc:"List of document embeddings"`
	}
}

// Delete project embeddings
// Path: "/embeddings/{user_handle}/{project_handle}"

type DeleteProjEmbeddingsRequest struct {
	UserHandle    string `json:"user_handle" path:"user_handle" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
	ProjectHandle string `json:"project_handle" path:"project_handle" maxLength:"20" minLength:"3" example:"my-gpt-4" doc:"Project handle"`
}

type DeleteProjEmbeddingsResponse struct {
	Header []http.Header `json:"header,omitempty" doc:"Response headers"`
}

// Get document embeddings
// Path: "/embeddings/{user_handle}/{project_handle}/{text_id}"

type GetDocEmbeddingsRequest struct {
	UserHandle    string `json:"user_handle" path:"user_handle" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
	ProjectHandle string `json:"project_handle" path:"project_handle" maxLength:"20" minLength:"3" example:"my-gpt-4" doc:"Project handle"`
	TextID        string `json:"text_id" path:"text_id" maxLength:"300" minLength:"3" example:"https%3A%2F%2Fid.salamanca.school%2Ftexts%2FW0017%3Afrontmatter.1.1%0A" doc:"Document identifier"`
}

type GetDocEmbeddingsResponse struct {
	Header []http.Header `json:"header,omitempty" doc:"Response headers"`
	Body   Embeddings
}

// Delete document embeddings
// Path: "/embeddings/{user_handle}/{project_handle}/{text_id}"

type DeleteDocEmbeddingsRequest struct {
	UserHandle    string `json:"user_handle" path:"user_handle" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
	ProjectHandle string `json:"project_handle" path:"project_handle" maxLength:"20" minLength:"3" example:"my-gpt-4" doc:"Project handle"`
	TextID        string `json:"text_id" path:"text_id" maxLength:"300" minLength:"3" example:"https%3A%2F%2Fid.salamanca.school%2Ftexts%2FW0017%3Afrontmatter.1.1%0A" doc:"Document identifier"`
}

type DeleteDocEmbeddingsResponse struct {
	Header []http.Header `json:"header,omitempty" doc:"Response headers"`
}
