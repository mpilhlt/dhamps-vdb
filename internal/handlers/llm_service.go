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
	if u, err := getUserFunc(ctx, &models.GetUserRequest{UserHandle: input.UserHandle}); u.Body.UserHandle != input.UserHandle {
		return nil, huma.Error404NotFound(fmt.Sprintf("user %s not found", input.UserHandle))
	} else if err != nil {
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to get user %s", input.UserHandle))
	}

	// Get the database connection pool from the context
	pool, err := GetDBPool(ctx)
	if err != nil {
		return nil, err
	}

	// Run the query
	queries := database.New(pool)
	llm, err := queries.UpsertLLM(ctx, database.UpsertLLMParams{
		Owner:            input.UserHandle,
		LLMServiceHandle: input.Body.LLMService.LLMServiceHandle,
		Endpoint:         input.Body.LLMService.Endpoint,
		ApiKey:           pgtype.Text{String: input.Body.LLMService.APIKey, Valid: true},
		ApiStandard:      input.Body.LLMService.ApiStandard,
	})
	if err != nil {
		return nil, huma.Error500InternalServerError("unable to upload llm service")
	}
	// Add llm service to user
	err = queries.LinkUserToLLM(ctx, database.LinkUserToLLMParams{UserHandle: input.UserHandle, LLMServiceID: llm.LLMServiceID})
	if err != nil {
		return nil, huma.Error500InternalServerError("unable to link llm service to user")
	}

	// Build response
	response := &models.UploadLLMResponse{}
	response.Body.LLMServiceHandle = llm.LLMServiceHandle

	return response, nil
}

func putLLMFunc(ctx context.Context, input *models.PutLLMRequest) (*models.UploadLLMResponse, error) {
	// Check if user exists
	if u, err := getUserFunc(ctx, &models.GetUserRequest{UserHandle: input.UserHandle}); u.Body.UserHandle != input.UserHandle {
		return nil, huma.Error404NotFound(fmt.Sprintf("user %s not found", input.UserHandle))
	} else if err != nil {
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to get user %s", input.UserHandle))
	}

	// Get the database connection pool from the context
	pool, err := GetDBPool(ctx)
	if err != nil {
		return nil, err
	}

	// Run the query
	queries := database.New(pool)
	llm, err := queries.UpsertLLM(ctx, database.UpsertLLMParams{
		Owner:            input.UserHandle,
		LLMServiceHandle: input.LLMServiceHandle,
		Endpoint:         input.Body.LLMService.Endpoint,
		ApiKey:           pgtype.Text{String: input.Body.LLMService.APIKey, Valid: true},
		ApiStandard:      input.Body.LLMService.ApiStandard,
	})
	if err != nil {
		return nil, huma.Error500InternalServerError("unable to upload llm service")
	}
	// Add llm service to user
	err = queries.LinkUserToLLM(ctx, database.LinkUserToLLMParams{UserHandle: input.UserHandle, LLMServiceID: llm.LLMServiceID})
	if err != nil {
		return nil, huma.Error500InternalServerError("unable to link llm service to user")
	}

	// Build response
	response := &models.UploadLLMResponse{}
	response.Body.LLMServiceHandle = llm.LLMServiceHandle

	return response, nil
}

func getLLMFunc(ctx context.Context, input *models.GetLLMRequest) (*models.GetLLMResponse, error) {
	// Check if user exists
	if u, err := getUserFunc(ctx, &models.GetUserRequest{UserHandle: input.UserHandle}); u.Body.UserHandle != input.UserHandle {
		return nil, huma.Error404NotFound(fmt.Sprintf("user %s not found", input.UserHandle))
	} else if err != nil {
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to get user %s", input.UserHandle))
	}

	// Get the database connection pool from the context
	pool, err := GetDBPool(ctx)
	if err != nil {
		return nil, err
	}

	// Run the query
	queries := database.New(pool)
	llm, err := queries.RetrieveLLM(ctx, database.RetrieveLLMParams{Owner: input.UserHandle, LLMServiceHandle: input.LLMServiceHandle})
	if err != nil {
		return nil, huma.Error500InternalServerError("unable to retrieve embeddings")
	}
	if llm.LLMServiceHandle != input.LLMServiceHandle {
		return nil, huma.Error404NotFound(fmt.Sprintf("embeddings for %s not found", input.LLMServiceHandle))
	}

	// Build response
	ls := models.LLMService{
		LLMServiceHandle: llm.LLMServiceHandle,
		Endpoint:         llm.Endpoint,
		APIKey:           llm.ApiKey.String,
		ApiStandard:      llm.ApiStandard,
	}
	response := &models.GetLLMResponse{}
	response.Body.LLMService = ls

	return response, nil
}

func getUserLLMsFunc(ctx context.Context, input *models.GetUserLLMsRequest) (*models.GetUserLLMsResponse, error) {
	// Check if user exists
	if u, err := getUserFunc(ctx, &models.GetUserRequest{UserHandle: input.UserHandle}); u.Body.UserHandle != input.UserHandle {
		return nil, huma.Error404NotFound(fmt.Sprintf("user %s not found", input.UserHandle))
	} else if err != nil {
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to get user %s", input.UserHandle))
	}

	// Get the database connection pool from the context
	pool, err := GetDBPool(ctx)
	if err != nil {
		return nil, err
	}

	// Run the query
	queries := database.New(pool)
	llm, err := queries.GetLLMsByUser(ctx, database.GetLLMsByUserParams{UserHandle: input.UserHandle, Limit: int32(input.Limit), Offset: int32(input.Offset)})
	if err != nil {
		return nil, huma.Error500InternalServerError("unable to retrieve embeddings")
	}
	if len(llm) == 0 {
		return nil, huma.Error404NotFound(fmt.Sprintf("no llm services for %s found", input.UserHandle))
	}

	// Build response
	ls := []models.LLMService{}
	for _, l := range llm {
		ls = append(ls, models.LLMService{
			LLMServiceHandle: l.LLMServiceHandle,
			Endpoint:         l.Endpoint,
			APIKey:           l.ApiKey.String,
			ApiStandard:      l.ApiStandard,
		})
	}
	response := &models.GetUserLLMsResponse{}
	response.Body.LLMServices = ls

	return response, nil
}

func deleteLLMFunc(ctx context.Context, input *models.DeleteLLMRequest) (*models.DeleteLLMResponse, error) {
	// Check if user exists
	if u, err := getUserFunc(ctx, &models.GetUserRequest{UserHandle: input.UserHandle}); u.Body.UserHandle != input.UserHandle {
		return nil, huma.Error404NotFound(fmt.Sprintf("user %s not found", input.UserHandle))
	} else if err != nil {
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to get user %s", input.UserHandle))
	}
	// Check if llm service exists
	if llm, err := getLLMFunc(ctx, &models.GetLLMRequest{UserHandle: input.UserHandle, LLMServiceHandle: input.LLMServiceHandle}); llm.Body.LLMService.LLMServiceHandle != input.LLMServiceHandle {
		return nil, huma.Error404NotFound(fmt.Sprintf("llm service %s not found for user %s", input.LLMServiceHandle, input.UserHandle))
	} else if err != nil {
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to get llm service %s for user %s", input.LLMServiceHandle, input.UserHandle))
	}

	// Get the database connection pool from the context
	pool, err := GetDBPool(ctx)
	if err != nil {
		return nil, err
	}

	// Run the query
	queries := database.New(pool)
	err = queries.DeleteLLM(ctx, database.DeleteLLMParams{Owner: input.UserHandle, LLMServiceHandle: input.LLMServiceHandle})
	if err != nil {
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to delete llm service %s for user %s", input.LLMServiceHandle, input.UserHandle))
	}

	// Build response
	response := &models.DeleteLLMResponse{}
	response.Body.Message = fmt.Sprintf("llm service %s deleted for user %s", input.LLMServiceHandle, input.UserHandle)

	return response, nil
}

// RegisterLLMServiceRoutes registers the routes for the management of LLM services
func RegisterLLMServiceRoutes(pool *pgxpool.Pool, api huma.API) error {
	// Define huma.Operations for each route
	postLLMServiceOp := huma.Operation{
		OperationID: "postLLMService",
		Method:      http.MethodPost,
		Path:        "/llmservices/{user_handle}",
		Summary:     "Create llm service",
		Security: []map[string][]string{
			{"adminAuth": []string{"admin"}},
			{"ownerAuth": []string{"owner"}},
		},
		Tags: []string{"llmservices"},
	}
	putLLMServiceOp := huma.Operation{
		OperationID: "putLLMService",
		Method:      http.MethodPut,
		Path:        "/llmservices/{user_handle}/{llmservice_handle}",
		Summary:     "Create or update llm service",
		Security: []map[string][]string{
			{"adminAuth": []string{"admin"}},
			{"ownerAuth": []string{"owner"}},
		},
		Tags: []string{"llmservices"},
	}
	getUserLLMServicesOp := huma.Operation{
		OperationID: "getUserLLMServices",
		Method:      http.MethodGet,
		Path:        "/llmservices/{user_handle}",
		Summary:     "Get all llm services for a user",
		Security: []map[string][]string{
			{"adminAuth": []string{"admin"}},
			{"ownerAuth": []string{"owner"}},
			{"readerAuth": []string{"reader"}},
		},
		Tags: []string{"llmservices"},
	}
	getLLMServiceOp := huma.Operation{
		OperationID: "getLLMService",
		Method:      http.MethodGet,
		Path:        "/llmservices/{user_handle}/{llmservice_handle}",
		Summary:     "Get a specific llm service for a user",
		Security: []map[string][]string{
			{"adminAuth": []string{"admin"}},
			{"ownerAuth": []string{"owner"}},
			{"readerAuth": []string{"reader"}},
		},
		Tags: []string{"llmservices"},
	}
	deleteLLMServiceOp := huma.Operation{
		OperationID: "deleteLLMService",
		Method:      http.MethodDelete,
		Path:        "/llmservices/{user_handle}/{llmservice_handle}",
		Summary:     "Delete all embeddings for a user",
		Security: []map[string][]string{
			{"adminAuth": []string{"admin"}},
			{"ownerAuth": []string{"owner"}},
			{"readerAuth": []string{"reader"}},
		},
		Tags: []string{"llmservices"},
	}

	huma.Register(api, postLLMServiceOp, addPoolToContext(pool, postLLMFunc))
	huma.Register(api, putLLMServiceOp, addPoolToContext(pool, putLLMFunc))
	huma.Register(api, getUserLLMServicesOp, addPoolToContext(pool, getUserLLMsFunc))
	huma.Register(api, getLLMServiceOp, addPoolToContext(pool, getLLMFunc))
	huma.Register(api, deleteLLMServiceOp, addPoolToContext(pool, deleteLLMFunc))
	return nil
}
