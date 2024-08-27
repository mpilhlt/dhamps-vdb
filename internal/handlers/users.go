package handlers

import (
  "context"
  "fmt"
  "net/http"

  "github.com/mpilhlt/dhamps-vdb/internal/models"

  "github.com/danielgtaylor/huma/v2"
)

// Define handler functions for each route
func putUserFunc(ctx context.Context, input *models.PutUserRequest) (*models.PutUserResponse, error) {
  // implement user creation or update logic here
  if input.Body.User.Handle == "bob" {
    return nil, huma.Error404NotFound("no action for bob")
  }
  response := &models.PutUserResponse{}
  response.Body.Handle = input.Body.User.Handle
  return response, nil
}

func getUsersFunc(ctx context.Context, input *models.GetUsersRequest) (*models.GetUsersResponse, error) {
  // implement user information logic here
  response := &models.GetUsersResponse{}
  response.Body.Handles = []string{"alice", "bob"}
  return response, nil
}

func patchUserFunc(ctx context.Context, input *models.PatchUserRequest) (*models.PatchUserResponse, error) {
  // implement user creation or update logic here
  if input.User == "bob" || input.Body.User.Handle == "bob" {
    return nil, huma.Error404NotFound("no action for bob")
  }
  response := &models.PatchUserResponse{}
  response.Body.User = input.Body.User
  return response, nil
}

func getUserFunc(ctx context.Context, input *models.GetUserRequest) (*models.GetUserResponse, error) {
  // implement user information logic here
  if input.User == "bob" {
    return nil, huma.Error404NotFound("no action for bob")
  }
  response := &models.GetUserResponse{}
  response.Body.User = models.User{
    Handle: input.User,
    Name:   "Alice",
    Email:  "al@ic.e",
    APIKey: "123",
    Projects: make([]models.Project, 0), 
  }
  return response, nil
}

func deleteUserFunc(ctx context.Context, input *models.DeleteUserRequest) (*models.DeleteUserResponse, error) {
  // implement user deletion logic here
  if input.User == "bob" {
    return nil, huma.Error404NotFound("no action for bob")
  }
  response := &models.DeleteUserResponse{}
  response.Body = fmt.Sprintf("Successfully deleted user %s", input.User)
  return response, nil
}

// RegisterUsersRoutes registers all the admin routes with the API
func RegisterUsersRoutes(api huma.API) {
  // Define huma.Operations for each route
  putUserOp := huma.Operation{
    OperationID: "putUser",
    Method:      http.MethodPut,
    Path:        "/users",
    Summary:     "Create or update a user",
    Tags:        []string{"admin", "users"},
  }
  postUserOp := huma.Operation{
    OperationID: "postUser",
    Method:      http.MethodPost,
    Path:        "/users",
    Summary:     "Create or update a user",
    Tags:        []string{"admin", "users"},
  }
  getUsersOp := huma.Operation{
    OperationID: "getUsers",
    Method:      http.MethodGet,
    Path:        "/users",
    Summary:     "Get information about all users",
    Tags:        []string{"admin", "users"},
  }
  patchUserOp := huma.Operation{
    OperationID: "patchUser",
    Method:      http.MethodPatch,
    Path:        "/users/{user}",
    Summary:     "Update a specific user",
    Tags:        []string{"admin", "users"},
  }
  getUserOp := huma.Operation{
    OperationID: "getUser",
    Method:      http.MethodGet,
    Path:        "/users/{user}",
    Summary:     "Get information about a specific user",
    Tags:        []string{"admin", "users"},
  }
  deleteUserOp := huma.Operation{
    OperationID: "deleteUser",
    Method:      http.MethodDelete,
    Path:        "/users/{user}",
    Summary:     "Delete a specific user",
    Tags:        []string{"admin", "users"},
  }

  huma.Register(api, putUserOp, putUserFunc)
  huma.Register(api, postUserOp, putUserFunc)
  huma.Register(api, getUsersOp, getUsersFunc)
  huma.Register(api, patchUserOp, patchUserFunc)
  huma.Register(api, getUserOp, getUserFunc)
  huma.Register(api, deleteUserOp, deleteUserFunc)
}
