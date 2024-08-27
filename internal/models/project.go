package models

import "net/http"

// Project is a project that a user is a member of.
type Project struct {
  Handle               string `json:"handle" minLength:"3" maxLength:"20" example:"my-gpt-4" doc:"Project handle"`
  Description          string `json:"description,omitempty" maxLength:"255" doc:"Description of the project."`
  AuthorizedReaders  []string `json:"authorizedReaders" default:"*" example:"[\"jdoe\",\"foobar\"]" doc:"Account names allowed to retrieve information from the project. Defaults to everyone ([\"*\"])"`
  LLMServices    []LLMService `json:"llmServices" default:nil doc:"LLM services used in the project"`
}

// Request and Response structs for the project administration API
// The request structs must be structs with fields for the request path/query/header/cookie parameters and/or body.
// The response structs must be structs with fields for the output headers and body of the operation, if any.

// Put/post project request/response
// Path: "/projects/{user}"

type PutProjectRequest struct {
  User       string `json:"user" path:"user" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
  Body       struct {
    Project Project `json:"project" doc:"Project information"`
  }
}

type PutProjectResponse struct {
  Header []http.Header `json:"header,omitempty" doc:"Response headers"`
  Body struct {
    Handle string `json:"id" doc:"Handle of created or updated project"`
  }
}

// Get all project request/response
// Path: "/projects/{user}"

type GetProjectsRequest struct {
  User       string `json:"user" path:"user" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
}

type GetProjectsResponse struct {
  Header []http.Header `json:"header,omitempty" doc:"Response headers"`
  Body struct {
    Handles []string `json:"handles" doc:"Handles of all registered projects for specified user"`
  }
}

// Patch project request/response
// Path: "/projects/{user}/{project}"

type PatchProjectRequest struct {
  User       string `json:"user" path:"user" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
  Project    string `json:"project" path:"project" maxLength:"20" minLength:"3" example:"my-gpt-4" doc:"Project handle"`
  Body       struct {
    Project Project `json:"project" doc:"Project information"`
  }
}

type PatchProjectResponse struct {
  Header []http.Header `json:"header,omitempty" doc:"Response headers"`
  Body struct {
    Project Project `json:"project" doc:"Updated project information"`
  }
}

// Get project request/response
// Path: "/projects/{user}/{project}"

type GetProjectRequest struct {
  User       string `json:"user" path:"user" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
  Project    string `json:"project" path:"project" maxLength:"20" minLength:"3" example:"my-gpt-4" doc:"Project handle"`
}

type GetProjectResponse struct {
  Header []http.Header `json:"header,omitempty" doc:"Response headers"`
  Body struct {
    Project Project `json:"project" doc:"Project information"`
  }
}

// Delete project request/response
// Path: "/projects/{user}/{project}"

type DeleteProjectRequest struct {
  User       string `json:"user" path:"user" maxLength:"20" minLength:"3" example:"jdoe" doc:"User handle"`
  Project    string `json:"project" path:"project" maxLength:"20" minLength:"3" example:"my-gpt-4" doc:"Project handle"`
}

type DeleteProjectResponse struct {
  Header []http.Header `json:"header,omitempty" doc:"Response headers"`
  Body string `json:"body" doc:"Status message"`
}
