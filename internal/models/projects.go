package models

import "net/http"

// Project is a project that a user is a member of.
type Project struct {
	ProjectID          int          `json:"project_id" readOnly:"true" doc:"Unique project identifier"`
	ProjectHandle      string       `json:"project_handle" minLength:"3" maxLength:"20" example:"my-gpt-4" doc:"Project handle"`
	Owner              string       `json:"owner" readOnly:"true" doc:"User handle of the project owner"`
	Description        string       `json:"description,omitempty" maxLength:"255" doc:"Description of the project."`
	MetadataScheme     string       `json:"metadataScheme,omitempty" doc:"Metadata json scheme used in the project."`
	AuthorizedReaders  []string     `json:"authorizedReaders,omitempty" default:"" example:"[\"jdoe\",\"foobar\"]" doc:"Account names allowed to retrieve information from the project. Defaults to everyone ([\"*\"])"`
	LLMServices        []LLMService `json:"llmServices,omitempty" doc:"LLM services used in the project"`
	NumberOfEmbeddings int          `json:"number_of_embeddings" readOnly:"true" doc:"Number of embeddings in the project"`
}

type ProjectSubmission struct {
	ProjectHandle     string       `json:"project_handle" minLength:"3" maxLength:"20" example:"my-gpt-4" doc:"Project handle"`
	Description       string       `json:"description,omitempty" maxLength:"255" doc:"Description of the project."`
	MetadataScheme    string       `json:"metadataScheme,omitempty" doc:"Metadata json scheme used in the project."`
	AuthorizedReaders []string     `json:"authorizedReaders,omitempty" default:"" example:"[\"jdoe\",\"foobar\"]" doc:"Account names allowed to retrieve information from the project. Defaults to everyone ([\"*\"])"`
	LLMServices       []LLMService `json:"llmServices,omitempty" doc:"LLM services used in the project"`
}

// Request and Response structs for the project administration API
// The request structs must be structs with fields for the request path/query/header/cookie parameters and/or body.
// The response structs must be structs with fields for the output headers and body of the operation, if any.

// Put/post project
// PUT Path: "/v1/projects/{user_handle}/{project_handle}"

type PutProjectRequest struct {
	UserHandle    string `json:"user_handle" path:"user_handle" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
	ProjectHandle string `json:"project_handle" path:"project_handle" maxLength:"20" minLength:"3" example:"my-gpt-4" doc:"Project handle"`
	Body          Project
}

// POST Path: "/v1/projects/{user}"

type PostProjectRequest struct {
	UserHandle string `json:"user_handle" path:"user_handle" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
	Body       Project
}

type UploadProjectResponse struct {
	Header []http.Header `json:"header,omitempty" doc:"Response headers"`
	Body   struct {
		ProjectHandle string `json:"project_handle" doc:"Handle of created or updated project"`
		ProjectID     int    `json:"project_id" doc:"Unique project identifier"`
	}
}

// Get all projects by user
// Path: "/v1/projects/{user_handle}"

type GetProjectsRequest struct {
	UserHandle string `json:"user_handle" path:"user_handle" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
	Limit      int    `json:"limit,omitempty" query:"limit" minimum:"1" maximum:"200" example:"10" default:"10" doc:"Maximum number of projects to return"`
	Offset     int    `json:"offset,omitempty" query:"offset" minimum:"0" example:"0" default:"0" doc:"Offset into the list of projects"`
}

type GetProjectsResponse struct {
	Header []http.Header `json:"header,omitempty" doc:"Response headers"`
	Body   struct {
		// Handles []string `json:"handles" doc:"Handles of all registered projects for specified user"`
		Projects []Project `json:"projects" doc:"Projects that the user is a member of"`
	}
}

// Get single project
// Path: "/v1/projects/{user_handle}/{project_handle}"

type GetProjectRequest struct {
	UserHandle    string `json:"user_handle" path:"user_handle" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
	ProjectHandle string `json:"project_handle" path:"project_handle" maxLength:"20" minLength:"3" example:"my-gpt-4" doc:"Project handle"`
}

type GetProjectResponse struct {
	Header []http.Header `json:"header,omitempty" doc:"Response headers"`
	Body   Project       `json:"project" doc:"Project information"`
}

// Delete project
// Path: "/v1/projects/{user_handle}/{project_handle}"

type DeleteProjectRequest struct {
	UserHandle    string `json:"user_handle" path:"user_handle" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
	ProjectHandle string `json:"project_handle" path:"project_handle" maxLength:"20" minLength:"3" example:"my-gpt-4" doc:"Project handle"`
}

type DeleteProjectResponse struct {
	Header []http.Header `json:"header,omitempty" doc:"Response headers"`
}
