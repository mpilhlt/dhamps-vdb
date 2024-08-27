package handlers

import (
  "context"
  "fmt"
  "net/http"

  "github.com/mpilhlt/dhamps-vdb/internal/models"

  "github.com/danielgtaylor/huma/v2"
)

// Define handler functions for each route
func putProjectFunc(ctx context.Context, input *models.PutProjectRequest) (*models.PutProjectResponse, error) {
  // implement project creation or update logic here
  if input.User == "bob" {
    return nil, huma.Error404NotFound("no action for bob")
  }
  response := &models.PutProjectResponse{}
  response.Body.Handle = input.Body.Project.Handle
  return response, nil
}

func getProjectsFunc(ctx context.Context, input *models.GetProjectsRequest) (*models.GetProjectsResponse, error) {
  // implement project information logic here
  if input.User == "bob" {
    return nil, huma.Error404NotFound("no action for bob")
  }
  response := &models.GetProjectsResponse{}
  response.Body.Handles = []string{"mock", "test"}
  return response, nil
}

func patchProjectFunc(ctx context.Context, input *models.PatchProjectRequest) (*models.PatchProjectResponse, error) {
  // implement project creation or update logic here
  if input.User == "bob" {
    return nil, huma.Error404NotFound("no action for bob")
  }
  response := &models.PatchProjectResponse{}
  response.Body.Project = input.Body.Project
  return response, nil
}

func getProjectFunc(ctx context.Context, input *models.GetProjectRequest) (*models.GetProjectResponse, error) {
  // implement project information logic here
  if input.User == "bob" {
    return nil, huma.Error404NotFound("no action for bob")
  }
  response := &models.GetProjectResponse{}
  response.Body.Project = models.Project{
    Handle: "mock",
    Description:  "Dummy project",
    AuthorizedReaders: []string{"alice"},
    LLMServices: nil,
  }
  return response, nil
}

func deleteProjectFunc(ctx context.Context, input *models.DeleteProjectRequest) (*models.DeleteProjectResponse, error) {
  // implement project deletion logic here
  if input.User == "bob" {
    return nil, huma.Error404NotFound("no action for bob")
  }
  response := &models.DeleteProjectResponse{}
  response.Body = fmt.Sprintf("Successfully deleted project %s", input.Project)
  return response, nil
}

// RegisterProjectRoutes registers all the project routes with the API
func RegisterProjectsRoutes(api huma.API) {
  // Define huma.Operations for each route
  putProjectOp := huma.Operation{
    OperationID: "putProject",
    Method:      http.MethodPut,
    Path:        "/projects/{user}",
    Summary:     "Create or update a project",
    Tags:        []string{"admin", "projects"},
  }
  postProjectOp := huma.Operation{
    OperationID: "postProject",
    Method:      http.MethodPost,
    Path:        "/projects/{user}",
    Summary:     "Create or update a project",
    Tags:        []string{"admin", "projects"},
  }
  getProjectsOp := huma.Operation{
    OperationID: "getProjects",
    Method:      http.MethodGet,
    Path:       "/projects/{user}",
    Summary:     "Get all projects for a specific user",
    Tags:        []string{"admin", "projects"},
  }
  patchProjectOp := huma.Operation{
    OperationID: "patchProject",
    Method:      http.MethodPatch,
    Path:        "/projects/{user}/{project}",
    Summary:     "Update a specific project",
    Tags:        []string{"admin", "projects"},
  }
  getProjectOp := huma.Operation{
    OperationID: "getProject",
    Method:      http.MethodGet,
    Path:        "/projects/{user}/{project}",
    Summary:     "Get a specific project",
    Tags:        []string{"admin", "projects"},
  }
  deleteProjectOp := huma.Operation{
    OperationID: "deleteProject",
    Method:      http.MethodDelete,
    Path:        "/projects/{user}/{project}",
    Summary:     "Delete a specific project",
    Tags:        []string{"admin", "projects"},
  }

  huma.Register(api, putProjectOp, putProjectFunc)
  huma.Register(api, postProjectOp, putProjectFunc)
  huma.Register(api, getProjectsOp, getProjectsFunc)
  huma.Register(api, patchProjectOp, patchProjectFunc)
  huma.Register(api, getProjectOp, getProjectFunc)
  huma.Register(api, deleteProjectOp, deleteProjectFunc)
}
