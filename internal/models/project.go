package models

import "net/http"

// Project is a project that a user is a member of.
type Project struct {
	Id                int          `json:"project_id" doc:"Unique project identifier"`
	Handle            string       `json:"handle" minLength:"3" maxLength:"20" example:"my-gpt-4" doc:"Project handle"`
	Description       string       `json:"description,omitempty" maxLength:"255" doc:"Description of the project."`
	MetadataScheme    string       `json:"metadataScheme,omitempty" doc:"Metadata json scheme used in the project."`
	AuthorizedReaders []string     `json:"authorizedReaders,omitempty" default:"" example:"[\"jdoe\",\"foobar\"]" doc:"Account names allowed to retrieve information from the project. Defaults to everyone ([\"*\"])"`
	LLMServices       []LLMService `json:"llmServices,omitempty" doc:"LLM services used in the project"`
}

type ProjectSubmission struct {
	Handle            string       `json:"handle" minLength:"3" maxLength:"20" example:"my-gpt-4" doc:"Project handle"`
	Description       string       `json:"description,omitempty" maxLength:"255" doc:"Description of the project."`
	MetadataScheme    string       `json:"metadataScheme,omitempty" doc:"Metadata json scheme used in the project."`
	AuthorizedReaders []string     `json:"authorizedReaders,omitempty" default:"" example:"[\"jdoe\",\"foobar\"]" doc:"Account names allowed to retrieve information from the project. Defaults to everyone ([\"*\"])"`
	LLMServices       []LLMService `json:"llmServices,omitempty" doc:"LLM services used in the project"`
}

// Request and Response structs for the project administration API
// The request structs must be structs with fields for the request path/query/header/cookie parameters and/or body.
// The response structs must be structs with fields for the output headers and body of the operation, if any.

// Put/post project
// PUT Path: "/projects/{user}/{project}"

type PutProjectRequest struct {
	User    string `json:"user" path:"user" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
	Project string `json:"project" path:"project" maxLength:"20" minLength:"3" example:"my-gpt-4" doc:"Project handle"`
	Body    ProjectSubmission
}

// POST Path: "/projects/{user}"

type PostProjectRequest struct {
	User string `json:"user" path:"user" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
	Body ProjectSubmission
}

type UploadProjectResponse struct {
	Header []http.Header `json:"header,omitempty" doc:"Response headers"`
	Body   struct {
		Handle string `json:"id" doc:"Handle of created or updated project"`
		Id     int    `json:"project_id" doc:"Unique project identifier"`
	}
}

// Get all projects by user
// Path: "/projects/{user}"

type GetProjectsRequest struct {
	User   string `json:"user" path:"user" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
	Limit  int    `json:"limit,omitempty" query:"limit" minimum:"1" maximum:"200" example:"10" default:"10" doc:"Maximum number of projects to return"`
	Offset int    `json:"offset,omitempty" query:"offset" minimum:"0" example:"0" default:"0" doc:"Offset into the list of projects"`
}

type GetProjectsResponse struct {
	Header []http.Header `json:"header,omitempty" doc:"Response headers"`
	Body   struct {
		// Handles []string `json:"handles" doc:"Handles of all registered projects for specified user"`
		Projects []Project `json:"projects" doc:"Projects that the user is a member of"`
	}
}

// Get single project
// Path: "/projects/{user}/{project}"

type GetProjectRequest struct {
	User    string `json:"user" path:"user" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
	Project string `json:"project" path:"project" maxLength:"20" minLength:"3" example:"my-gpt-4" doc:"Project handle"`
}

type GetProjectResponse struct {
	Header []http.Header `json:"header,omitempty" doc:"Response headers"`
	Body   struct {
		Project Project `json:"project" doc:"Project information"`
	}
}

// Delete project
// Path: "/projects/{user}/{project}"

type DeleteProjectRequest struct {
	User    string `json:"user" path:"user" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
	Project string `json:"project" path:"project" maxLength:"20" minLength:"3" example:"my-gpt-4" doc:"Project handle"`
}

type DeleteProjectResponse struct {
	Header []http.Header `json:"header,omitempty" doc:"Response headers"`
}
