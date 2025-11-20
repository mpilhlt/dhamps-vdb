package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mpilhlt/dhamps-vdb/internal/database"
	"github.com/mpilhlt/dhamps-vdb/internal/models"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func putLLMFunc(ctx context.Context, input *models.PutLLMRequest) (*models.UploadLLMResponse, error) {
	if input.LLMServiceHandle != input.Body.LLMServiceHandle {
		return nil, huma.Error400BadRequest(fmt.Sprintf("llm-service handle in URL (\"%s\") does not match llm-service handle in body (\"%s\")", input.LLMServiceHandle, input.Body.LLMServiceHandle))
	}

	// Check if user exists
	u, err := getUserFunc(ctx, &models.GetUserRequest{UserHandle: input.UserHandle})
	if err != nil {
		return nil, err
	}
	if u.Body.UserHandle != input.UserHandle {
		return nil, huma.Error404NotFound(fmt.Sprintf("user %s not found", input.UserHandle))
	}

	// Get the database connection pool from the context
	pool, err := GetDBPool(ctx)
	if err != nil {
		return nil, err
	}

	// Execute all database operations within a transaction
	var llmServiceID int32
	var llmServiceHandle string
	var owner string

	err = database.WithTransaction(ctx, pool, func(tx pgx.Tx) error {
		queries := database.New(tx)

		// 1. Upsert LLM service
		llm, err := queries.UpsertLLM(ctx, database.UpsertLLMParams{
			Owner:            input.UserHandle,
			LLMServiceHandle: input.LLMServiceHandle,
			Endpoint:         input.Body.Endpoint,
			Description:      pgtype.Text{String: input.Body.Description, Valid: true},
			APIKey:           pgtype.Text{String: input.Body.APIKey, Valid: true},
			APIStandard:      input.Body.APIStandard,
			Model:            input.Body.Model,
			Dimensions:       int32(input.Body.Dimensions),
		})
		if err != nil {
			return fmt.Errorf("unable to upload llm service. %v", err)
		}

		llmServiceID = llm.LLMServiceID
		llmServiceHandle = llm.LLMServiceHandle
		owner = llm.Owner

		// 2. Link llm service to user
		err = queries.LinkUserToLLM(ctx, database.LinkUserToLLMParams{UserHandle: input.UserHandle, LLMServiceID: llm.LLMServiceID, Role: "owner"})
		if err != nil {
			return fmt.Errorf("unable to link llm service to user. %v", err)
		}

		return nil
	})

	if err != nil {
		return nil, huma.Error500InternalServerError(err.Error())
	}

	// Build response
	response := &models.UploadLLMResponse{}
	response.Body.Owner = owner
	response.Body.LLMServiceHandle = llmServiceHandle
	response.Body.LLMServiceID = int(llmServiceID)

	return response, nil
}

// Create a llm service (without a handle being present in the URL)
func postLLMFunc(ctx context.Context, input *models.PostLLMRequest) (*models.UploadLLMResponse, error) {
	return putLLMFunc(ctx, &models.PutLLMRequest{UserHandle: input.UserHandle, LLMServiceHandle: input.Body.LLMServiceHandle, Body: input.Body})
}

func getLLMFunc(ctx context.Context, input *models.GetLLMRequest) (*models.GetLLMResponse, error) {
	// Check if user exists
	u, err := getUserFunc(ctx, &models.GetUserRequest{UserHandle: input.UserHandle})
	if err != nil {
		return nil, err
	}
	if u.Body.UserHandle != input.UserHandle {
		return nil, huma.Error404NotFound(fmt.Sprintf("user %s not found", input.UserHandle))
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
		if err.Error() == "no rows in result set" {
			return nil, huma.Error404NotFound(fmt.Sprintf("llm service %s for user %s not found", input.LLMServiceHandle, input.UserHandle))
		}
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to retrieve llm service %s for user %s. %v", input.LLMServiceHandle, input.UserHandle, err))
	}
	if llm.LLMServiceHandle != input.LLMServiceHandle {
		return nil, huma.Error404NotFound(fmt.Sprintf("llm service %s for user %s not found", input.LLMServiceHandle, input.UserHandle))
	}

	// Build response
	ls := models.LLMService{
		Owner:            llm.Owner,
		LLMServiceHandle: llm.LLMServiceHandle,
		LLMServiceID:     int(llm.LLMServiceID),
		Endpoint:         llm.Endpoint,
		Description:      llm.Description.String,
		APIKey:           llm.APIKey.String,
		APIStandard:      llm.APIStandard,
		Model:            llm.Model,
		Dimensions:       int32(llm.Dimensions),
	}
	response := &models.GetLLMResponse{}
	response.Body = ls

	return response, nil
}

func getUserLLMsFunc(ctx context.Context, input *models.GetUserLLMsRequest) (*models.GetUserLLMsResponse, error) {
	// Check if user exists
	u, err := getUserFunc(ctx, &models.GetUserRequest{UserHandle: input.UserHandle})
	if err != nil {
		return nil, err
	}
	if u.Body.UserHandle != input.UserHandle {
		return nil, huma.Error404NotFound(fmt.Sprintf("user %s not found", input.UserHandle))
	}

	// Get the database connection pool from the context
	pool, err := GetDBPool(ctx)
	if err != nil {
		return nil, err
	}

	// Run the query
	queries := database.New(pool)
	llms, err := queries.GetLLMsByUser(ctx, database.GetLLMsByUserParams{UserHandle: input.UserHandle, Limit: int32(input.Limit), Offset: int32(input.Offset)})
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, huma.Error404NotFound(fmt.Sprintf("no llm services for %s found", input.UserHandle))
		}
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to retrieve llm services. %v", err))
	}
	if len(llms) == 0 {
		return nil, huma.Error404NotFound(fmt.Sprintf("no llm services for %s found", input.UserHandle))
	}

	// Build response
	ls := []models.LLMService{}
	for _, llm := range llms {
		ls = append(ls, models.LLMService{
			Owner:            llm.Owner,
			LLMServiceHandle: llm.LLMServiceHandle,
			LLMServiceID:     int(llm.LLMServiceID),
			Endpoint:         llm.Endpoint,
			Description:      llm.Description.String,
			APIKey:           llm.APIKey.String,
			APIStandard:      llm.APIStandard,
			Model:            llm.Model,
			Dimensions:       int32(llm.Dimensions),
		})
	}
	response := &models.GetUserLLMsResponse{}
	response.Body.LLMServices = ls

	return response, nil
}

func deleteLLMFunc(ctx context.Context, input *models.DeleteLLMRequest) (*models.DeleteLLMResponse, error) {
	// Check if user exists
	u, err := getUserFunc(ctx, &models.GetUserRequest{UserHandle: input.UserHandle})
	if err != nil {
		return nil, err
	}
	if u.Body.UserHandle != input.UserHandle {
		return nil, huma.Error404NotFound(fmt.Sprintf("user %s not found", input.UserHandle))
	}

	// Check if llm service exists
	_, err = getLLMFunc(ctx, &models.GetLLMRequest{UserHandle: input.UserHandle, LLMServiceHandle: input.LLMServiceHandle})
	if err != nil {
		return nil, err
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
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to delete llm service %s for user %s. %v", input.LLMServiceHandle, input.UserHandle, err))
	}

	// Build response
	response := &models.DeleteLLMResponse{}

	return response, nil
}

// RegisterLLMServiceRoutes registers the routes for the management of LLM services
func RegisterLLMServicesRoutes(pool *pgxpool.Pool, api huma.API) error {
	// Define huma.Operations for each route
	postLLMServiceOp := huma.Operation{
		OperationID:   "postLLMService",
		Method:        http.MethodPost,
		Path:          "/v1/llm-services/{user_handle}",
		DefaultStatus: http.StatusCreated,
		Summary:       "Create llm service",
		Security: []map[string][]string{
			{"adminAuth": []string{"admin"}},
			{"ownerAuth": []string{"owner"}},
		},
		Tags: []string{"llm-services"},
	}
	putLLMServiceOp := huma.Operation{
		OperationID:   "putLLMService",
		Method:        http.MethodPut,
		Path:          "/v1/llm-services/{user_handle}/{llm_service_handle}",
		DefaultStatus: http.StatusCreated,
		Summary:       "Create or update llm service",
		Security: []map[string][]string{
			{"adminAuth": []string{"admin"}},
			{"ownerAuth": []string{"owner"}},
		},
		Tags: []string{"llm-services"},
	}
	getUserLLMServicesOp := huma.Operation{
		OperationID: "getUserLLMServices",
		Method:      http.MethodGet,
		Path:        "/v1/llm-services/{user_handle}",
		Summary:     "Get all llm services for a user",
		Security: []map[string][]string{
			{"adminAuth": []string{"admin"}},
			{"ownerAuth": []string{"owner"}},
			{"readerAuth": []string{"reader"}},
		},
		Tags: []string{"llm-services"},
	}
	getLLMServiceOp := huma.Operation{
		OperationID: "getLLMService",
		Method:      http.MethodGet,
		Path:        "/v1/llm-services/{user_handle}/{llm_service_handle}",
		Summary:     "Get a specific llm service for a user",
		Security: []map[string][]string{
			{"adminAuth": []string{"admin"}},
			{"ownerAuth": []string{"owner"}},
			{"readerAuth": []string{"reader"}},
		},
		Tags: []string{"llm-services"},
	}
	deleteLLMServiceOp := huma.Operation{
		OperationID:   "deleteLLMService",
		Method:        http.MethodDelete,
		Path:          "/v1/llm-services/{user_handle}/{llm_service_handle}",
		DefaultStatus: http.StatusNoContent,
		Summary:       "Delete a user's llm_service and all embeddings associated to it",
		Security: []map[string][]string{
			{"adminAuth": []string{"admin"}},
			{"ownerAuth": []string{"owner"}},
		},
		Tags: []string{"llm-services"},
	}

	huma.Register(api, postLLMServiceOp, addPoolToContext(pool, postLLMFunc))
	huma.Register(api, putLLMServiceOp, addPoolToContext(pool, putLLMFunc))
	huma.Register(api, getUserLLMServicesOp, addPoolToContext(pool, getUserLLMsFunc))
	huma.Register(api, getLLMServiceOp, addPoolToContext(pool, getLLMFunc))
	huma.Register(api, deleteLLMServiceOp, addPoolToContext(pool, deleteLLMFunc))
	return nil
}
