package models

// LLMService is a service for managing LLM data.
type LLMService struct {
  ServiceName      string `json:"serviceName" minLength:"3" maxLength:"20" example:"GPT-4 API" doc:"Service name"`
  Endpoint         string `json:"endpoint" example:"https://api.openai.com/v1/embeddings" doc:"Service endpoint"`
  Token            string `json:"token,omitempty" example:"12345678901234567890123456789012" doc:"Authentication token for the service"`
  TokenMethod      string `json:"tokenMethod" enum:"header,query" default:"header" example:"header" doc:"Method for sending the token"`
  ContextData      string `json:"contextData,omitempty" doc:"Context data that can be fed to the LLM service. Available in the request template as contextData variable."`
  SystemPrompt     string `json:"systemPrompt,omitempty" example:"Return the embeddings for the following text:" doc:"System prompt for requests to the service. Available in the request template as systemPrompt variable."`
  RequestTemplate  string `json:"requestTemplate,omitempty" doc:"Request template for the service. Can use input, contextData, and systemPrompt variables." example:"{\"input\": \"{{ input }}\", \"model\": \"text-embedding-3-small\"}"`
  RespFieldName    string `json:"respFieldName,omitempty" default:"embedding" example:"embedding" doc:"Field name of the service response containing the embeddings. Supported is a top-level key of a json object."`
}
