package handlers

import (
  "context"
  "fmt"
  "net/http"

  "github.com/mpilhlt/dhamps-vdb/internal/models"

  "github.com/danielgtaylor/huma/v2"
)

// Define handler functions for each route
func putProjEmbeddingsFunc(ctx context.Context, input *models.PutProjEmbeddingsRequest) (*models.PutProjEmbeddingsResponse, error) {
  // implement embeddings creation or update logic here
  if len(input.Body.Embeddings) == 0 {
    return nil, huma.Error404NotFound("nothing to do, because len(ebeddings) == 0.")
  }
  response := &models.PutProjEmbeddingsResponse{}
  response.Body.IDs = input.Body.Embeddings.GetIDs()
  return response, nil
}

func postProjEmbeddingsFunc(ctx context.Context, input *models.PutProjEmbeddingsRequest) (*models.PutProjEmbeddingsResponse, error) {
  // implement embeddings creation or update logic here
  if len(input.Body.Embeddings) == 0 {
    return nil, huma.Error404NotFound("nothing to do, because len(ebeddings) == 0.")
  }
  response := &models.PutProjEmbeddingsResponse{}
  response.Body.IDs = make([]string, 0)
  return response, nil
}

func getProjEmbeddingsFunc(ctx context.Context, input *models.GetProjEmbeddingsRequest) (*models.GetProjEmbeddingsResponse, error) {
  // implement embeddings information logic here
  if input.User == "bob" {
    return nil, huma.Error404NotFound("no action for bob")
  }
  response := &models.GetProjEmbeddingsResponse{}
  response.Body.Embeddings = make([]models.Embeddings, 0)
  return response, nil
}

func deleteProjEmbeddingsFunc(ctx context.Context, input *models.DeleteProjEmbeddingsRequest) (*models.DeleteProjEmbeddingsResponse, error) {
  // implement embeddings deletion logic here
  if input.User == "bob" {
    return nil, huma.Error404NotFound("no action for bob")
  }
  response := &models.DeleteProjEmbeddingsResponse{}
  response.Body = fmt.Sprintf("Successfully deleted all embeddings for project %s", input.Project)
  return response, nil
}

func getDocEmbeddingsFunc(ctx context.Context, input *models.GetDocEmbeddingsRequest) (*models.GetDocEmbeddingsResponse, error) {
  // implement embeddings information logic here
  if input.User == "bob" {
    return nil, huma.Error404NotFound("no action for bob")
  }
  response := &models.GetDocEmbeddingsResponse{}
  return response, nil
}

func patchDocEmbeddingsFunc(ctx context.Context, input *models.PatchDocEmbeddingsRequest) (*models.PatchDocEmbeddingsResponse, error) {
  // implement embeddings creation or update logic here
  if input.User == "bob" {
    return nil, huma.Error404NotFound("no action for bob")
  }
  response := &models.PatchDocEmbeddingsResponse{}
  response.Body.Embeddings.ID = input.ID
  return response, nil
}

func deleteDocEmbeddingsFunc(ctx context.Context, input *models.DeleteDocEmbeddingsRequest) (*models.DeleteDocEmbeddingsResponse, error) {
  // implement embeddings deletion logic here
  if input.User == "bob" {
    return nil, huma.Error404NotFound("no action for bob")
  }
  response := &models.DeleteDocEmbeddingsResponse{}
  response.Body = fmt.Sprintf("Successfully deleted embeddings for document %s (project %s)", input.ID, input.Project)
  return response, nil
}

// RegisterEmbeddingsRoutes registers all the embeddings routes with the API
func RegisterEmbeddingsRoutes(api huma.API) {
  // Define huma.Operations for each route
  putProjEmbeddingsOp := huma.Operation{
    OperationID: "putEmbeddings",
    Method:      http.MethodPut,
    Path:        "/embeddings/{user}/{project}",
    Summary:     "Create or update embeddings for a project",
    Tags:        []string{"embeddings"},
  }
  postProjEmbeddingsOp := huma.Operation{
    OperationID: "postEmbeddings",
    Method:      http.MethodPost,
    Path:        "/embeddings/{user}/{project}",
    Summary:     "Create or update embeddings for a project",
    Tags:        []string{"embeddings"},
  }
  getProjEmbeddingsOp := huma.Operation{
    OperationID: "getEmbeddings",
    Method:      http.MethodGet,
    Path:        "/embeddings/{user}/{project}",
    Summary:     "Get all embeddings for a project",
    Tags:        []string{"embeddings"},
  }
  deleteProjEmbeddingsOp := huma.Operation{
    OperationID: "deleteEmbeddings",
    Method:      http.MethodDelete,
    Path:        "/embeddings/{user}/{project}",
    Summary:     "Delete all embeddings for a project",
    Tags:        []string{"embeddings"},
  }
  getDocEmbeddingsOp := huma.Operation{
    OperationID: "getDocEmbeddings",
    Method:      http.MethodGet,
    Path:        "/embeddings/{user}/{project}/{id}",
    Summary:     "Get embeddings for a specific document",
    Tags:        []string{"embeddings"},
  }
  patchDocEmbeddingsOp := huma.Operation{
    OperationID: "patchDocEmbeddings",
    Method:      http.MethodPatch,
    Path:        "/embeddings/{user}/{project}/{id}",
    Summary:     "Patch embeddings for a specific document",
    Tags:        []string{"embeddings"},
  }
  deleteDocEmbeddingsOp := huma.Operation{
    OperationID: "deleteDocEmbeddings",
    Method:      http.MethodDelete,
    Path:        "/embeddings/{user}/{project}/{id}",
    Summary:     "Delete embeddings for a specific document",
    Tags:        []string{"embeddings"},
  }

  huma.Register(api, putProjEmbeddingsOp, putProjEmbeddingsFunc)
  huma.Register(api, postProjEmbeddingsOp, postProjEmbeddingsFunc)
  huma.Register(api, getProjEmbeddingsOp, getProjEmbeddingsFunc)
  huma.Register(api, deleteProjEmbeddingsOp, deleteProjEmbeddingsFunc)
  huma.Register(api, getDocEmbeddingsOp, getDocEmbeddingsFunc)
  huma.Register(api, patchDocEmbeddingsOp, patchDocEmbeddingsFunc)
  huma.Register(api, deleteDocEmbeddingsOp, deleteDocEmbeddingsFunc)
}
