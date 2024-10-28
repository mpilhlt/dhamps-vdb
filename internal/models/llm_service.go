package models

import (
	"net/http"
)

// LLMService is a service for managing LLM data.
type LLMService struct {
  Handle           string `json:"serviceName" minLength:"3" maxLength:"20" example:"GPT-4 API" doc:"Service name"`
  Endpoint         string `json:"endpoint" example:"https://api.openai.com/v1/embeddings" doc:"Service endpoint"`
  Description      string `json:"description,omitempty" doc:"Service description"`
  APIKey           string `json:"apiKey,omitempty" example:"12345678901234567890123456789012" doc:"Authentication token for the service"`
  ApiStandard      string `json:"apiStandard" enum:"openai,custom" default:"openai" example:"openai" doc:"Standard of the API"`
  // ContextData      string `json:"contextData,omitempty" doc:"Context data that can be fed to the LLM service. Available in the request template as contextData variable."`
  // SystemPrompt     string `json:"systemPrompt,omitempty" example:"Return the embeddings for the following text:" doc:"System prompt for requests to the service. Available in the request template as systemPrompt variable."`
  // RequestTemplate  string `json:"requestTemplate,omitempty" doc:"Request template for the service. Can use input, contextData, and systemPrompt variables." example:"{\"input\": \"{{ input }}\", \"model\": \"text-embedding-3-small\"}"`
  // RespFieldName    string `json:"respFieldName,omitempty" default:"embedding" example:"embedding" doc:"Field name of the service response containing the embeddings. Supported is a top-level key of a json object."`
}

// Request and Response structs for the project administration API
// The request structs must be structs with fields for the request path/query/header/cookie parameters and/or body.
// The response structs must be structs with fields for the output headers and body of the operation, if any.

// Put/post project
// PUT Path: "/llmservices/{user}/{handle}"

type PutLLMRequest struct {
  User       string `json:"user" path:"user" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
  Handle     string `json:"handle" path:"handle" maxLength:"20" minLength:"3" example:"my-gpt-4" doc:"LLM service handle"`
  Body       struct {
    LLMService LLMService `json:"llm_service" doc:"LLM service to create or update"`
  }
}

// POST Path: "/llmservices/{user}"

type PostLLMRequest struct {
  User       string `json:"user" path:"user" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
  Body       struct {
    LLMService LLMService `json:"llm_service" doc:"LLM service to create or update"`
  }
}

type UploadLLMResponse struct {
  Header []http.Header `json:"header,omitempty" doc:"Response headers"`
  Body struct {
    Handle string `json:"handle" doc:"Handle of created or updated LLM service"`
  }
}

// Get all LLM services by user
// Path: "/llmservices/{user}"

type GetUserLLMsRequest struct {
  User       string `json:"user" path:"user" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
  Limit      int `json:"limit,omitempty" query:"limit" minimum:"1" maximum:"200" example:"10" default:"20" doc:"Maximum number of embeddings to return"`
  Offset     int `json:"offset,omitempty" query:"offset" minimum:"0" example:"0" default:"0" doc:"Offset into the list of embeddings"`
}

type GetUserLLMsResponse struct {
  Header []http.Header `json:"header,omitempty" doc:"Response headers"`
  Body struct {
    LLMServices []LLMService `json:"llm_service" doc:"List of LLM Services"`
  }
}

// Get single LLM service
// Path: "/llmservices/{user}/{handle}"

type GetLLMRequest struct {
  User       string `json:"user" path:"user" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
  Handle     string `json:"handle" path:"handle" maxLength:"20" minLength:"3" example:"my-gpt-4" doc:"LLM service handle"`
  Limit      int `json:"limit,omitempty" query:"limit" minimum:"1" maximum:"200" example:"10" default:"20" doc:"Maximum number of embeddings to return"`
  Offset     int `json:"offset,omitempty" query:"offset" minimum:"0" example:"0" default:"0" doc:"Offset into the list of embeddings"`
}

type GetLLMResponse struct {
  Header []http.Header `json:"header,omitempty" doc:"Response headers"`
  Body struct {
    LLMService LLMService `json:"llm_service" doc:"LLM Service"`
  }
}

// Delete LLM service
// Path: "/llmservices/{user}/{handle}"

type DeleteLLMRequest struct {
  User       string `json:"user" path:"user" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
  Handle     string `json:"handle" path:"handle" maxLength:"20" minLength:"3" example:"my-gpt-4" doc:"LLM service handle"`
}

type DeleteLLMResponse struct {
  Header []http.Header `json:"header,omitempty" doc:"Response headers"`
  Body struct {
    Message string `json:"message" doc:"Message indicating the deletion was successful"`
  }
}
