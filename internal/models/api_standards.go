package models

import (
	"net/http"
)

// LLMService is a service for managing LLM data.
type APIStandard struct {
	APIStandardHandle string `json:"api_standard_handle" minLength:"3" maxLength:"20" example:"openai-v1" doc:"Handle for the API standard"`
	Description       string `json:"description" doc:"Description of the API standard"`
	KeyMethod         string `json:"key_method" doc:"Method for providing the API key" example:"bearer"`
	KeyField          string `json:"key_field" doc:"HTTP Request field name to store the API key in (in bearer method)" example:"Authorization"`
}

// Request and Response structs for the API standard administration API
// The request structs must be structs with fields for the request path/query/header/cookie parameters and/or body.
// The response structs must be structs with fields for the output headers and body of the operation, if any.

// Put/post api_standard
// PUT Path: "/v1/api-standards/{api_standard_handle}"

type PutAPIStandardRequest struct {
	APIStandardHandle string      `json:"api_standard_handle" path:"api_standard_handle" maxLength:"20" minLength:"3" example:"openai-v1" doc:"Handle for the API standard"`
	Body              APIStandard `json:"api_standard" doc:"API standard to create or update"`
}

// POST Path: "/v1/api-standards"

type PostAPIStandardRequest struct {
	Body APIStandard `json:"api_standard" doc:"API standard to create or update"`
}

type UploadAPIStandardResponse struct {
	Header []http.Header `json:"header,omitempty" doc:"Response headers"`
	Body   struct {
		APIStandardHandle string `json:"api_standard_handle" doc:"Handle of created or updated API standard"`
	}
}

// Get all registered API standards
// Path: "/v1/api-standards"

type GetAPIStandardsRequest struct {
	Limit  int `json:"limit,omitempty" query:"limit" minimum:"1" maximum:"200" example:"10" default:"20" doc:"Maximum number of embeddings to return"`
	Offset int `json:"offset,omitempty" query:"offset" minimum:"0" example:"0" default:"0" doc:"Offset into the list of embeddings"`
}

type GetAPIStandardsResponse struct {
	Header []http.Header `json:"header,omitempty" doc:"Response headers"`
	Body   struct {
		APIStandards []APIStandard `json:"api_standards" doc:"List of API standards"`
	}
}

// Get single API standard
// Path: "/v1/api-standards/{api_standard_handle}"

type GetAPIStandardRequest struct {
	APIStandardHandle string `json:"api_standard_handle" path:"api_standard_handle" maxLength:"20" minLength:"3" example:"openai-v1" doc:"API standard handle"`
}

type GetAPIStandardResponse struct {
	Header []http.Header `json:"header,omitempty" doc:"Response headers"`
	Body   APIStandard   `json:"api standard" doc:"API standard"`
}

// Delete API standard
// Path: "/v1/api-standards/{api_standard_handle}"

type DeleteAPIStandardRequest struct {
	APIStandardHandle string `json:"api_standard_handle" path:"api_standard_handle" maxLength:"20" minLength:"3" example:"openai-v1" doc:"API standard handle"`
}

type DeleteAPIStandardResponse struct {
	Header []http.Header `json:"header,omitempty" doc:"Response headers"`
}
