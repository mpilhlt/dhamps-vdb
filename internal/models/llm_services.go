package models

import (
	"net/http"
)

// LLMService is a service for managing LLM data.
type LLMService struct {
	LLMServiceHandle string `json:"llm_service_handle" minLength:"3" maxLength:"20" example:"GPT-4 API" doc:"Service name"`
	Endpoint         string `json:"endpoint" example:"https://api.openai.com/v1/embeddings" doc:"Service endpoint"`
	Description      string `json:"description,omitempty" doc:"Service description"`
	APIKey           string `json:"api_key,omitempty" example:"12345678901234567890123456789012" doc:"Authentication token for the service"`
	ApiStandard      string `json:"api_standard" enum:"openai,ollama,custom" default:"openai" example:"openai" doc:"Standard of the API"`
	LLModel          string `json:"model" example:"text-embedding-3-large" doc:"Model name"`
	Dimensions       int    `json:"dimensions" example:"3072" doc:"Number of dimensions in the embeddings"`
	// ContextData      string `json:"contextData,omitempty" doc:"Context data that can be fed to the LLM service. Available in the request template as contextData variable."`
	// SystemPrompt     string `json:"systemPrompt,omitempty" example:"Return the embeddings for the following text:" doc:"System prompt for requests to the service. Available in the request template as systemPrompt variable."`
	// RequestTemplate  string `json:"requestTemplate,omitempty" doc:"Request template for the service. Can use input, contextData, and systemPrompt variables." example:"{\"input\": \"{{ input }}\", \"model\": \"text-embedding-3-small\"}"`
	// RespFieldName    string `json:"respFieldName,omitempty" default:"embedding" example:"embedding" doc:"Field name of the service response containing the embeddings. Supported is a top-level key of a json object."`
}

// Request and Response structs for the project administration API
// The request structs must be structs with fields for the request path/query/header/cookie parameters and/or body.
// The response structs must be structs with fields for the output headers and body of the operation, if any.

// Put/post llm-service
// PUT Path: "/llm-services/{user_handle}/{llm_service_handle}"

type PutLLMRequest struct {
	UserHandle       string     `json:"user_handle" path:"user_handle" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
	LLMServiceHandle string     `json:"llm_service_handle" path:"llm_service_handle" maxLength:"20" minLength:"3" example:"my-gpt-4" doc:"LLM service handle"`
	Body             LLMService `json:"llm_service" doc:"LLM service to create or update"`
}

// POST Path: "/llm-services/{user_handle}"

type PostLLMRequest struct {
	UserHandle string     `json:"user_handle" path:"user_handle" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
	Body       LLMService `json:"llm_service" doc:"LLM service to create or update"`
}

type UploadLLMResponse struct {
	Header []http.Header `json:"header,omitempty" doc:"Response headers"`
	Body   struct {
		LLMServiceHandle string `json:"llm_service_handle" doc:"Handle of created or updated LLM service"`
	}
}

// Get all LLM services by user
// Path: "/llm-services/{user_handle}"

type GetUserLLMsRequest struct {
	UserHandle string `json:"user_handle" path:"user_handle" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
	Limit      int    `json:"limit,omitempty" query:"limit" minimum:"1" maximum:"200" example:"10" default:"20" doc:"Maximum number of embeddings to return"`
	Offset     int    `json:"offset,omitempty" query:"offset" minimum:"0" example:"0" default:"0" doc:"Offset into the list of embeddings"`
}

type GetUserLLMsResponse struct {
	Header []http.Header `json:"header,omitempty" doc:"Response headers"`
	Body   struct {
		LLMServices []LLMService `json:"llm_service" doc:"List of LLM Services"`
	}
}

// Get single LLM service
// Path: "/llm-services/{user_handle}/{llm_service_handle}"

type GetLLMRequest struct {
	UserHandle       string `json:"user_handle" path:"user_handle" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
	LLMServiceHandle string `json:"llm_service_handle" path:"llm_service_handle" maxLength:"20" minLength:"3" example:"my-gpt-4" doc:"LLM service handle"`
	Limit            int    `json:"limit,omitempty" query:"limit" minimum:"1" maximum:"200" example:"10" default:"20" doc:"Maximum number of embeddings to return"`
	Offset           int    `json:"offset,omitempty" query:"offset" minimum:"0" example:"0" default:"0" doc:"Offset into the list of embeddings"`
}

type GetLLMResponse struct {
	Header []http.Header `json:"header,omitempty" doc:"Response headers"`
	Body   struct {
		LLMService LLMService `json:"llm_service" doc:"LLM Service"`
	}
}

// Delete LLM service
// Path: "/llm-services/{user_handle}/{llm_service_handle}"

type DeleteLLMRequest struct {
	UserHandle       string `json:"user_handle" path:"user_handle" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
	LLMServiceHandle string `json:"llm_service_handle" path:"llm_service_handle" maxLength:"20" minLength:"3" example:"my-gpt-4" doc:"LLM service handle"`
}

type DeleteLLMResponse struct {
	Header []http.Header `json:"header,omitempty" doc:"Response headers"`
}
