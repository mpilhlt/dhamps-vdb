package models

import "net/http"

type GetSimilarRequest struct {
  User       string  `json:"user" path:"user" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
  Project    string  `json:"project" path:"project" maxLength:"20" minLength:"3" example:"my-gpt-4" doc:"Project handle"`
  ID         string  `json:"id" path:"id" maxLength:"200" minLength:"3" example:"https%3A%2F%2Fid.salamanca.school%2Ftexts%2FW0017%3Afrontmatter.1.1%0A" doc:"Document identifier"`
  Count      int     `json:"count"`
  Threshold  float64 `json:"threshold"`
  Limit      int     `json:"limit,omitempty" query:"limit" minimum:"1" maximum:"200" example:"10" default:"10" doc:"Maximum number of similar documents to return"`
  Offset     int     `json:"offset,omitempty" query:"offset" minimum:"0" example:"0" default:"0" doc:"Offset into the list of similar documents"`
}

type PostSimilarRequest struct {
  User       string  `json:"user" path:"user" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
  Project    string  `json:"project" path:"project" maxLength:"20" minLength:"3" example:"my-gpt-4" doc:"Project handle"`
  LLMService string  `json:"llm_service" path:"llm_service" maxLength:"20" minLength:"3" example:"my-gpt-4" doc:"LLM service handle"`
  Count      int     `json:"count"`
  Threshold  float64 `json:"threshold"`
  Limit      int     `json:"limit,omitempty" query:"limit" minimum:"1" maximum:"200" example:"10" default:"10" doc:"Maximum number of similar documents to return"`
  Offset     int     `json:"offset,omitempty" query:"offset" minimum:"0" example:"0" default:"0" doc:"Offset into the list of similar documents"`
}

type SimilarResponse struct {
  Header []http.Header `json:"header,omitempty" doc:"Response headers"`
  Body struct {
    User       string `json:"user" doc:"User handle"`
    Project    string `json:"project" doc:"Project handle"`
    IDs        []string `json:"ids" doc:"List of similar document identifiers"`
  }
}
