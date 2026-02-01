package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/mpilhlt/dhamps-vdb/internal/crypto"
	"github.com/mpilhlt/dhamps-vdb/internal/database"
	"github.com/mpilhlt/dhamps-vdb/internal/models"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// getEncryptionKey retrieves the encryption key, returns nil if not set (optional encryption)
func getEncryptionKey() *crypto.EncryptionKey {
	keyStr := os.Getenv("ENCRYPTION_KEY")
	if keyStr == "" {
		return nil
	}
	return crypto.NewEncryptionKey(keyStr)
}

func putLLMInstanceFunc(ctx context.Context, input *models.PutLLMRequest) (*models.UploadLLMResponse, error) {
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

	// Get encryption key if available
	encKey := getEncryptionKey()

	// Execute all database operations within a transaction
	var instanceID int32
	var instanceHandle string
	var owner string

	err = database.WithTransaction(ctx, pool, func(tx pgx.Tx) error {
		queries := database.New(tx)

		// Prepare API key encryption
		var apiKeyEncrypted []byte
		if input.Body.APIKey != "" && encKey != nil {
			apiKeyEncrypted, err = encKey.Encrypt(input.Body.APIKey)
			if err != nil {
				return fmt.Errorf("unable to encrypt API key: %v", err)
			}
		}

		// 1. Upsert LLM service instance
		llm, err := queries.UpsertLLMInstance(ctx, database.UpsertLLMInstanceParams{
			Owner:           input.UserHandle,
			InstanceHandle:  input.LLMServiceHandle,
			DefinitionID:    pgtype.Int4{Valid: false}, // Standalone instance (no definition reference)
			Endpoint:        input.Body.Endpoint,
			Description:     pgtype.Text{String: input.Body.Description, Valid: true},
			APIKey:          pgtype.Text{String: input.Body.APIKey, Valid: true},
			ApiKeyEncrypted: apiKeyEncrypted,
			APIStandard:     input.Body.APIStandard,
			Model:           input.Body.Model,
			Dimensions:      int32(input.Body.Dimensions),
		})
		if err != nil {
			return fmt.Errorf("unable to upload llm service instance: %v", err)
		}

		instanceID = llm.InstanceID
		instanceHandle = llm.InstanceHandle
		owner = llm.Owner

		// 2. Link llm service instance to user
		err = queries.LinkUserToLLMInstance(ctx, database.LinkUserToLLMInstanceParams{
			UserHandle: input.UserHandle,
			InstanceID: instanceID,
			Role:       "owner",
		})
		if err != nil {
			return fmt.Errorf("unable to link llm service instance to user: %v", err)
		}

		return nil
	})

	if err != nil {
		return nil, huma.Error500InternalServerError(err.Error())
	}

	// Build response
	response := &models.UploadLLMResponse{}
	response.Body.Owner = owner
	response.Body.LLMServiceHandle = instanceHandle
	response.Body.LLMServiceID = int(instanceID)

	return response, nil
}

// Create a llm service (without a handle being present in the URL)
func postLLMInstanceFunc(ctx context.Context, input *models.PostLLMRequest) (*models.UploadLLMResponse, error) {
	return putLLMInstanceFunc(ctx, &models.PutLLMRequest{UserHandle: input.UserHandle, LLMServiceHandle: input.Body.LLMServiceHandle, Body: input.Body})
}

func getLLMInstanceFunc(ctx context.Context, input *models.GetLLMRequest) (*models.GetLLMResponse, error) {
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
	llm, err := queries.RetrieveLLMInstance(ctx, database.RetrieveLLMInstanceParams{
		Owner:          input.UserHandle,
		InstanceHandle: input.LLMServiceHandle,
	})
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, huma.Error404NotFound(fmt.Sprintf("llm service %s for user %s not found", input.LLMServiceHandle, input.UserHandle))
		}
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to retrieve llm service %s for user %s: %v", input.LLMServiceHandle, input.UserHandle, err))
	}
	if llm.InstanceHandle != input.LLMServiceHandle {
		return nil, huma.Error404NotFound(fmt.Sprintf("llm service %s for user %s not found", input.LLMServiceHandle, input.UserHandle))
	}

	// Build response (never return API key in plaintext)
	ls := models.LLMService{
		Owner:            llm.Owner,
		LLMServiceHandle: llm.InstanceHandle,
		LLMServiceID:     int(llm.InstanceID),
		Endpoint:         llm.Endpoint,
		Description:      llm.Description.String,
		APIKey:           "", // Never return API key
		APIStandard:      llm.APIStandard,
		Model:            llm.Model,
		Dimensions:       int32(llm.Dimensions),
	}
	response := &models.GetLLMResponse{}
	response.Body = ls

	return response, nil
}

func getUserLLMInstancesFunc(ctx context.Context, input *models.GetUserLLMsRequest) (*models.GetUserLLMsResponse, error) {
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

	// Run the query - get all accessible instances (own + shared)
	queries := database.New(pool)
	llms, err := queries.GetAllAccessibleLLMInstances(ctx, database.GetAllAccessibleLLMInstancesParams{
		Owner:  input.UserHandle,
		Limit:  int32(input.Limit),
		Offset: int32(input.Offset),
	})
	if err != nil {
		if err.Error() == "no rows in result set" {
			// Return empty list instead of error
			response := &models.GetUserLLMsResponse{}
			response.Body.LLMServices = []models.LLMService{}
			return response, nil
		}
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to retrieve llm services: %v", err))
	}

	// Build response (hide API keys for shared instances)
	ls := []models.LLMService{}
	for _, llm := range llms {
		ls = append(ls, models.LLMService{
			Owner:            llm.Owner,
			LLMServiceHandle: llm.InstanceHandle,
			LLMServiceID:     int(llm.InstanceID),
			Endpoint:         llm.Endpoint,
			Description:      llm.Description.String,
			APIKey:           "", // Never return API key in list
			APIStandard:      llm.APIStandard,
			Model:            llm.Model,
			Dimensions:       int32(llm.Dimensions),
		})
	}
	response := &models.GetUserLLMsResponse{}
	response.Body.LLMServices = ls

	return response, nil
}

func deleteLLMInstanceFunc(ctx context.Context, input *models.DeleteLLMRequest) (*models.DeleteLLMResponse, error) {
	// Check if user exists
	u, err := getUserFunc(ctx, &models.GetUserRequest{UserHandle: input.UserHandle})
	if err != nil {
		return nil, err
	}
	if u.Body.UserHandle != input.UserHandle {
		return nil, huma.Error404NotFound(fmt.Sprintf("user %s not found", input.UserHandle))
	}

	// Check if llm service instance exists
	_, err = getLLMInstanceFunc(ctx, &models.GetLLMRequest{UserHandle: input.UserHandle, LLMServiceHandle: input.LLMServiceHandle})
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
	err = queries.DeleteLLMInstance(ctx, database.DeleteLLMInstanceParams{
		Owner:          input.UserHandle,
		InstanceHandle: input.LLMServiceHandle,
	})
	if err != nil {
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to delete llm service %s for user %s: %v", input.LLMServiceHandle, input.UserHandle, err))
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

	huma.Register(api, postLLMServiceOp, addPoolToContext(pool, postLLMInstanceFunc))
	huma.Register(api, putLLMServiceOp, addPoolToContext(pool, putLLMInstanceFunc))
	huma.Register(api, getUserLLMServicesOp, addPoolToContext(pool, getUserLLMInstancesFunc))
	huma.Register(api, getLLMServiceOp, addPoolToContext(pool, getLLMInstanceFunc))
	huma.Register(api, deleteLLMServiceOp, addPoolToContext(pool, deleteLLMInstanceFunc))
	return nil
}
