package models

import "net/http"

type GetSimilarRequest struct {
	UserHandle    string  `json:"user_handle" path:"user_handle" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
	ProjectHandle string  `json:"project_handle" path:"project_handle" maxLength:"20" minLength:"3" example:"my-gpt-4" doc:"Project handle"`
	TextID        string  `json:"text_id" path:"text_id" maxLength:"300" minLength:"3" example:"https%3A%2F%2Fid.salamanca.school%2Ftexts%2FW0017%3Afrontmatter.1.1%0A" doc:"Document identifier"`
	Count         int     `json:"count" query:"count" minimum:"1" maximum:"200" example:"10" default:"10" doc:"Number of similar documents to return"`
	Threshold     float64 `json:"threshold" query:"threshold" minimum:"0" maximum:"1" example:"0.5" default:"0.5" doc:"Similarity threshold"`
	MetadataPath  string  `json:"metadata_path,omitempty" query:"metadata_path" example:"{'author'}" doc:"Path to a field in the json metadata"`
	MetadataValue string  `json:"metadata_value,omitempty" query:"metadata_value" example:"'Hans Mustermann'" doc:"Value to filter out in the json metadata"`
	Limit         int     `json:"limit,omitempty" query:"limit" minimum:"1" maximum:"200" example:"10" default:"10" doc:"Maximum number of similar documents to return"`
	Offset        int     `json:"offset,omitempty" query:"offset" minimum:"0" example:"0" default:"0" doc:"Offset into the list of similar documents"`
}

type PostSimilarRequest struct {
	UserHandle    string  `json:"user_handle" path:"user_handle" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
	ProjectHandle string  `json:"project_handle" path:"project_handle" maxLength:"20" minLength:"3" example:"my-gpt-4" doc:"Project handle"`
	Count         int     `json:"count" query:"count" minimum:"1" maximum:"200" example:"10" default:"10" doc:"Number of similar documents to return"`
	Threshold     float64 `json:"threshold" query:"threshold" minimum:"0" maximum:"1" example:"0.5" default:"0.5" doc:"Similarity threshold"`
	MetadataPath  string  `json:"metadata_path,omitempty" query:"metadata_path" example:"{'author'}" doc:"Path to a field in the json metadata"`
	MetadataValue string  `json:"metadata_value,omitempty" query:"metadata_value" example:"'Hans Mustermann'" doc:"Value to filter out in the json metadata"`
	Limit         int     `json:"limit,omitempty" query:"limit" minimum:"1" maximum:"200" example:"10" default:"10" doc:"Maximum number of similar documents to return"`
	Offset        int     `json:"offset,omitempty" query:"offset" minimum:"0" example:"0" default:"0" doc:"Offset into the list of similar documents"`
	Body          struct {
		Vector []float32 `json:"vector" doc:"Embeddings vector to find similar documents for"`
	}
}

type SimilarResponse struct {
	Header []http.Header `json:"header,omitempty" doc:"Response headers"`
	Body   struct {
		UserHandle    string   `json:"user_handle" doc:"User handle"`
		ProjectHandle string   `json:"project_handle" doc:"Project handle"`
		IDs           []string `json:"ids" doc:"List of similar document identifiers"`
	}
}
