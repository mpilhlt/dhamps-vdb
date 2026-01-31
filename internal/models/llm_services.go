package models

import (
	"net/http"
)

// LLMService is a service for managing LLM data.
type LLMServiceInput struct {
	LLMServiceID     int    `json:"llm_service_id,omitempty" doc:"Unique service identifier" example:"153"`
	LLMServiceHandle string `json:"llm_service_handle" minLength:"3" maxLength:"20" example:"GPT-4 API" doc:"Service name"`
	Endpoint         string `json:"endpoint" example:"https://api.openai.com/v1/embeddings" doc:"Service endpoint"`
	Description      string `json:"description,omitempty" doc:"Service description"`
	APIKey           string `json:"api_key,omitempty" example:"12345678901234567890123456789012" doc:"Authentication token for the service"`
	APIStandard      string `json:"api_standard" default:"openai" example:"openai" doc:"Standard of the API"`
	Model            string `json:"model" example:"text-embedding-3-large" doc:"Model name"`
	Dimensions       int32  `json:"dimensions" example:"3072" doc:"Number of dimensions in the embeddings"`
	// ContextData      string `json:"contextData,omitempty" doc:"Context data that can be fed to the LLM service. Available in the request template as contextData variable."`
	// SystemPrompt     string `json:"systemPrompt,omitempty" example:"Return the embeddings for the following text:" doc:"System prompt for requests to the service. Available in the request template as systemPrompt variable."`
	// RequestTemplate  string `json:"requestTemplate,omitempty" doc:"Request template for the service. Can use input, contextData, and systemPrompt variables." example:"{\"input\": \"{{ input }}\", \"model\": \"text-embedding-3-small\"}"`
	// RespFieldName    string `json:"respFieldName,omitempty" default:"embedding" example:"embedding" doc:"Field name of the service response containing the embeddings. Supported is a top-level key of a json object."`
}

type LLMService struct {
	LLMServiceID     int    `json:"llm_service_id,omitempty" readOnly:"true" doc:"Unique service identifier" example:"153"`
	LLMServiceHandle string `json:"llm_service_handle" minLength:"3" maxLength:"20" example:"GPT-4 API" doc:"Service name"`
	Owner            string `json:"owner" readOnly:"true" doc:"User handle of the service owner"`
	Endpoint         string `json:"endpoint" example:"https://api.openai.com/v1/embeddings" doc:"Service endpoint"`
	Description      string `json:"description,omitempty" doc:"Service description"`
	APIKey           string `json:"api_key,omitempty" example:"12345678901234567890123456789012" doc:"Authentication token for the service"`
	APIStandard      string `json:"api_standard" default:"openai" example:"openai" doc:"Standard of the API"`
	Model            string `json:"model" example:"text-embedding-3-large" doc:"Model name"`
	Dimensions       int32  `json:"dimensions" example:"3072" doc:"Number of dimensions in the embeddings"`
	// ContextData      string `json:"contextData,omitempty" doc:"Context data that can be fed to the LLM service. Available in the request template as contextData variable."`
	// SystemPrompt     string `json:"systemPrompt,omitempty" example:"Return the embeddings for the following text:" doc:"System prompt for requests to the service. Available in the request template as systemPrompt variable."`
	// RequestTemplate  string `json:"requestTemplate,omitempty" doc:"Request template for the service. Can use input, contextData, and systemPrompt variables." example:"{\"input\": \"{{ input }}\", \"model\": \"text-embedding-3-small\"}"`
	// RespFieldName    string `json:"respFieldName,omitempty" default:"embedding" example:"embedding" doc:"Field name of the service response containing the embeddings. Supported is a top-level key of a json object."`
}

// Request and Response structs for the project administration API
// The request structs must be structs with fields for the request path/query/header/cookie parameters and/or body.
// The response structs must be structs with fields for the output headers and body of the operation, if any.

// Put/post llm-service
// PUT Path: "/v1/llm-services/{user_handle}/{llm_service_handle}"

type PutLLMRequest struct {
	UserHandle       string     `json:"user_handle" path:"user_handle" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
	LLMServiceHandle string     `json:"llm_service_handle" path:"llm_service_handle" maxLength:"20" minLength:"3" example:"my-gpt-4" doc:"LLM service handle"`
	Body             LLMService `json:"llm_service" doc:"LLM service to create or update"`
}

// POST Path: "/v1/llm-services/{user_handle}"

type PostLLMRequest struct {
	UserHandle string     `json:"user_handle" path:"user_handle" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
	Body       LLMService `json:"llm_service" doc:"LLM service to create or update"`
}

type UploadLLMResponse struct {
	Header []http.Header `json:"header,omitempty" doc:"Response headers"`
	Body   struct {
		Owner            string `json:"owner" doc:"User handle of the service owner"`
		LLMServiceHandle string `json:"llm_service_handle" doc:"Handle of created or updated LLM service"`
		LLMServiceID     int    `json:"llm_service_id" doc:"System identifier of created or updated LLM service"`
	}
}

// Get all LLM services by user
// Path: "/v1/llm-services/{user_handle}"

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
// Path: "/v1/llm-services/{user_handle}/{llm_service_handle}"

type GetLLMRequest struct {
	UserHandle       string `json:"user_handle" path:"user_handle" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
	LLMServiceHandle string `json:"llm_service_handle" path:"llm_service_handle" maxLength:"20" minLength:"3" example:"my-gpt-4" doc:"LLM service handle"`
	Limit            int    `json:"limit,omitempty" query:"limit" minimum:"1" maximum:"200" example:"10" default:"20" doc:"Maximum number of embeddings to return"`
	Offset           int    `json:"offset,omitempty" query:"offset" minimum:"0" example:"0" default:"0" doc:"Offset into the list of embeddings"`
}

type GetLLMResponse struct {
	Header []http.Header `json:"header,omitempty" doc:"Response headers"`
	Body   LLMService    `json:"llm_service" doc:"LLM Service"`
}

// Delete LLM service
// Path: "/v1/llm-services/{user_handle}/{llm_service_handle}"

type DeleteLLMRequest struct {
	UserHandle       string `json:"user_handle" path:"user_handle" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
	LLMServiceHandle string `json:"llm_service_handle" path:"llm_service_handle" maxLength:"20" minLength:"3" example:"my-gpt-4" doc:"LLM service handle"`
}

type DeleteLLMResponse struct {
	Header []http.Header `json:"header,omitempty" doc:"Response headers"`
}

// New types for Definitions and Instances architecture

// LLMServiceDefinition represents a template for LLM service configurations
// Definitions can be owned by _system (global templates) or individual users
type LLMServiceDefinition struct {
	DefinitionID     int    `json:"definition_id,omitempty" readOnly:"true" doc:"Unique definition identifier" example:"42"`
	DefinitionHandle string `json:"definition_handle" minLength:"3" maxLength:"20" example:"openai-large" doc:"Definition handle"`
	Owner            string `json:"owner" readOnly:"true" doc:"User handle of the definition owner (_system for global)" example:"_system"`
	Endpoint         string `json:"endpoint" example:"https://api.openai.com/v1/embeddings" doc:"Service endpoint"`
	Description      string `json:"description,omitempty" doc:"Service description"`
	APIStandard      string `json:"api_standard" example:"openai" doc:"Standard of the API"`
	Model            string `json:"model" example:"text-embedding-3-large" doc:"Model name"`
	Dimensions       int32  `json:"dimensions" example:"3072" doc:"Number of dimensions in the embeddings"`
}

// LLMServiceInstance represents a user-specific instance of an LLM service
// Instances can be based on a definition or standalone
type LLMServiceInstance struct {
	InstanceID       int     `json:"instance_id,omitempty" readOnly:"true" doc:"Unique instance identifier" example:"153"`
	InstanceHandle   string  `json:"instance_handle" minLength:"3" maxLength:"20" example:"my-openai-large" doc:"Instance handle"`
	Owner            string  `json:"owner" readOnly:"true" doc:"User handle of the instance owner"`
	DefinitionID     *int    `json:"definition_id,omitempty" doc:"Reference to definition (if based on one)"`
	DefinitionHandle string  `json:"definition_handle,omitempty" readOnly:"true" doc:"Handle of the definition (if based on one)"`
	Endpoint         string  `json:"endpoint" example:"https://api.openai.com/v1/embeddings" doc:"Service endpoint"`
	Description      string  `json:"description,omitempty" doc:"Service description"`
	APIKey           string  `json:"api_key,omitempty" writeOnly:"true" doc:"Authentication token (write-only, never returned)"`
	HasAPIKey        bool    `json:"has_api_key,omitempty" readOnly:"true" doc:"Indicates if instance has an API key configured"`
	APIStandard      string  `json:"api_standard" example:"openai" doc:"Standard of the API"`
	Model            string  `json:"model" example:"text-embedding-3-large" doc:"Model name"`
	Dimensions       int32   `json:"dimensions" example:"3072" doc:"Number of dimensions in the embeddings"`
	SharedWith       []string `json:"shared_with,omitempty" readOnly:"true" doc:"Users this instance is shared with"`
	IsShared         bool    `json:"is_shared,omitempty" readOnly:"true" doc:"Indicates if this is a shared instance (not owned by requesting user)"`
}

// CreateInstanceFromDefinitionRequest is for creating an instance based on a definition
type CreateInstanceFromDefinitionRequest struct {
	UserHandle       string  `json:"user_handle" path:"user_handle" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
	InstanceHandle   string  `json:"instance_handle" path:"instance_handle" maxLength:"20" minLength:"3" example:"my-openai" doc:"Instance handle"`
	DefinitionOwner  string  `json:"definition_owner" example:"_system" doc:"Owner of the definition to base instance on"`
	DefinitionHandle string  `json:"definition_handle" example:"openai-large" doc:"Handle of the definition to base instance on"`
	APIKey           string  `json:"api_key,omitempty" doc:"Optional API key for this instance"`
	Endpoint         *string `json:"endpoint,omitempty" doc:"Optional endpoint override"`
	Description      *string `json:"description,omitempty" doc:"Optional description override"`
}
