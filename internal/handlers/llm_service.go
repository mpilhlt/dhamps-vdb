package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mpilhlt/dhamps-vdb/internal/database"
	"github.com/mpilhlt/dhamps-vdb/internal/models"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Create a new llm service
func postLLMFunc(ctx context.Context, input *models.PostLLMRequest) (*models.UploadLLMResponse, error) {
  // Check if user exists
  if u, err := getUserFunc(ctx, &models.GetUserRequest{Handle: input.User}) ; u.Body.Handle != input.User {
    return nil, huma.Error404NotFound(fmt.Sprintf("user %s not found", input.User))
  } else if err != nil {
    return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to get user %s", input.User))
  }

  // Get the database connection pool from the context
  pool, err := GetDBPool(ctx)
  if err != nil {
      return nil, err
  }

  // Run the query
  queries := database.New(pool)
  llm, err := queries.UpsertLLM(ctx, database.UpsertLLMParams{
    Owner: input.User,
    Handle: input.Body.LLMService.Handle,
    Endpoint: input.Body.LLMService.Endpoint,
    ApiKey: pgtype.Text{ String: input.Body.LLMService.APIKey, Valid: true },
    ApiStandard: input.Body.LLMService.ApiStandard,
  })
  if err != nil {
    return nil, huma.Error500InternalServerError("unable to upload llm service")
  }
  // Add llm service to user
  err = queries.LinkUserToLLM(ctx, database.LinkUserToLLMParams{ User: input.User, Llmservice: llm.LlmserviceID })
  if err != nil {
    return nil, huma.Error500InternalServerError("unable to link llm service to user")
  }

  // Build response
  response := &models.UploadLLMResponse{}
  response.Body.Handle = llm.Handle

  return response, nil
}

func putLLMFunc(ctx context.Context, input *models.PutLLMRequest) (*models.UploadLLMResponse, error) {
  // Check if user exists
  if u, err := getUserFunc(ctx, &models.GetUserRequest{Handle: input.User}) ; u.Body.Handle != input.User {
    return nil, huma.Error404NotFound(fmt.Sprintf("user %s not found", input.User))
  } else if err != nil {
    return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to get user %s", input.User))
  }

  // Get the database connection pool from the context
  pool, err := GetDBPool(ctx)
  if err != nil {
      return nil, err
  }

  // Run the query
  queries := database.New(pool)
  llm, err := queries.UpsertLLM(ctx, database.UpsertLLMParams{
    Owner: input.User,
    Handle: input.Handle,
    Endpoint: input.Body.LLMService.Endpoint,
    ApiKey: pgtype.Text{ String: input.Body.LLMService.APIKey, Valid: true },
    ApiStandard: input.Body.LLMService.ApiStandard,
  })
  if err != nil {
    return nil, huma.Error500InternalServerError("unable to upload llm service")
  }
  // Add llm service to user
  err = queries.LinkUserToLLM(ctx, database.LinkUserToLLMParams{ User: input.User, Llmservice: llm.LlmserviceID })
  if err != nil {
    return nil, huma.Error500InternalServerError("unable to link llm service to user")
  }

  // Build response
  response := &models.UploadLLMResponse{}
  response.Body.Handle = llm.Handle

  return response, nil
}

func getLLMFunc(ctx context.Context, input *models.GetLLMRequest) (*models.GetLLMResponse, error) {
  // Check if user exists
  if u, err := getUserFunc(ctx, &models.GetUserRequest{Handle: input.User}) ; u.Body.Handle != input.User {
    return nil, huma.Error404NotFound(fmt.Sprintf("user %s not found", input.User))
  } else if err != nil {
    return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to get user %s", input.User))
  }

  // Get the database connection pool from the context
  pool, err := GetDBPool(ctx)
  if err != nil {
      return nil, err
  }

  // Run the query
  queries := database.New(pool)
  llm, err := queries.RetrieveLLM(ctx, database.RetrieveLLMParams{ Owner: input.User, Handle: input.Handle })
  if err != nil {
    return nil, huma.Error500InternalServerError("unable to retrieve embeddings")
  }
  if llm.Handle != input.Handle {
    return nil, huma.Error404NotFound(fmt.Sprintf("embeddings for %s not found", input.Handle))
  }

  // Build response
  ls := models.LLMService{
    Handle: llm.Handle,
    Endpoint: llm.Endpoint,
    APIKey: llm.ApiKey.String,
    ApiStandard: llm.ApiStandard,
  }
  response := &models.GetLLMResponse{}
  response.Body.LLMService = ls

  return response, nil
}

func getUserLLMsFunc(ctx context.Context, input *models.GetUserLLMsRequest) (*models.GetUserLLMsResponse, error) {
  // Check if user exists
  if u, err := getUserFunc(ctx, &models.GetUserRequest{Handle: input.User}) ; u.Body.Handle != input.User {
    return nil, huma.Error404NotFound(fmt.Sprintf("user %s not found", input.User))
  } else if err != nil {
    return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to get user %s", input.User))
  }

  // Get the database connection pool from the context
  pool, err := GetDBPool(ctx)
  if err != nil {
      return nil, err
  }

  // Run the query
  queries := database.New(pool)
  llm, err := queries.GetLLMsByUser(ctx, database.GetLLMsByUserParams{ UserHandle: input.User, Limit: int32(input.Limit), Offset: int32(input.Offset) })
  if err != nil {
    return nil, huma.Error500InternalServerError("unable to retrieve embeddings")
  }
  if len(llm) == 0 {
    return nil, huma.Error404NotFound(fmt.Sprintf("no llm services for %s found", input.User))
  }

  // Build response
  ls := []models.LLMService{}
  for _, l := range llm {
    ls = append(ls, models.LLMService{
      Handle: l.Handle,
      Endpoint: l.Endpoint,
      APIKey: l.ApiKey.String,
      ApiStandard: l.ApiStandard,
    })
  }
  response := &models.GetUserLLMsResponse{}
  response.Body.LLMServices = ls

  return response, nil
}

func deleteLLMFunc(ctx context.Context, input *models.DeleteLLMRequest) (*models.DeleteLLMResponse, error) {
  // Check if user exists
  if u, err := getUserFunc(ctx, &models.GetUserRequest{Handle: input.User}) ; u.Body.Handle != input.User {
    return nil, huma.Error404NotFound(fmt.Sprintf("user %s not found", input.User))
  } else if err != nil {
    return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to get user %s", input.User))
  }
  // Check if llm service exists
  if llm, err := getLLMFunc(ctx, &models.GetLLMRequest{User: input.User, Handle: input.Handle}) ; llm.Body.LLMService.Handle != input.Handle {
    return nil, huma.Error404NotFound(fmt.Sprintf("llm service %s not found for user %s", input.Handle, input.User))
  } else if err != nil {
    return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to get llm service %s for user %s", input.Handle, input.User))
  }

  // Get the database connection pool from the context
  pool, err := GetDBPool(ctx)
  if err != nil {
      return nil, err
  }

  // Run the query
  queries := database.New(pool)
  err = queries.DeleteLLM(ctx, database.DeleteLLMParams{ Owner: input.User, Handle: input.Handle })
  if err != nil {
    return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to delete llm service %s for user %s", input.Handle, input.User))
  }

  // Build response
  response := &models.DeleteLLMResponse{}
  response.Body.Message = fmt.Sprintf("llm service %s deleted for user %s", input.Handle, input.User)

  return response, nil
}

// RegisterLLMServiceRoutes registers the routes for the management of LLM services
func RegisterLLMServiceRoutes(pool *pgxpool.Pool, api huma.API) error {
  // Define huma.Operations for each route
  postLLMServiceOp := huma.Operation{
    OperationID: "postLLMService",
    Method:      http.MethodPost,
    Path:        "/llmservices/{user}",
    Summary:     "Create llm service",
    Tags:        []string{"llmservices"},
  }
  putLLMServiceOp := huma.Operation{
    OperationID: "putLLMService",
    Method:      http.MethodPut,
    Path:        "/llmservices/{user}/{handle}",
    Summary:     "Create or update llm service",
    Tags:        []string{"llmservices"},
  }
  getUserLLMServicesOp := huma.Operation{
    OperationID: "getUserLLMServices",
    Method:      http.MethodGet,
    Path:        "/llmservices/{user}",
    Summary:     "Get all llm services for a user",
    Tags:        []string{"llmservices"},
  }
  getLLMServiceOp := huma.Operation{
    OperationID: "getLLMService",
    Method:      http.MethodGet,
    Path:        "/llmservices/{user}/{handle}",
    Summary:     "Get a specific llm service for a user",
    Tags:        []string{"llmservices"},
  }
  deleteLLMServiceOp := huma.Operation{
    OperationID: "deleteLLMService",
    Method:      http.MethodDelete,
    Path:        "/llmservices/{user}/{handle}",
    Summary:     "Delete all embeddings for a user",
    Tags:        []string{"llmservices"},
  }

  huma.Register(api, postLLMServiceOp, addPoolToContext(pool, postLLMFunc))
  huma.Register(api, putLLMServiceOp, addPoolToContext(pool, putLLMFunc))
  huma.Register(api, getUserLLMServicesOp, addPoolToContext(pool, getUserLLMsFunc))
  huma.Register(api, getLLMServiceOp, addPoolToContext(pool, getLLMFunc))
  huma.Register(api, deleteLLMServiceOp, addPoolToContext(pool, deleteLLMFunc))
  return nil
}
