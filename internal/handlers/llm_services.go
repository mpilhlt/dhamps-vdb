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

func putInstanceFunc(ctx context.Context, input *models.PutInstanceRequest) (*models.UploadInstanceResponse, error) {
	if input.InstanceHandle != input.Body.InstanceHandle {
		return nil, huma.Error400BadRequest(fmt.Sprintf("instance handle in URL (\"%s\") does not match instance handle in body (\"%s\")", input.InstanceHandle, input.Body.InstanceHandle))
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
		var APIKeyEncrypted []byte
		if input.Body.APIKey != "" && encKey != nil {
			APIKeyEncrypted, err = encKey.Encrypt(input.Body.APIKey)
			if err != nil {
				return fmt.Errorf("unable to encrypt API key: %v", err)
			}
		}

		// 1. Upsert LLM service instance
		llm, err := queries.UpsertInstance(ctx, database.UpsertInstanceParams{
			Owner:           input.UserHandle,
			InstanceHandle:  input.InstanceHandle,
			DefinitionID:    pgtype.Int4{Valid: false}, // Standalone instance (no definition reference)
			Endpoint:        input.Body.Endpoint,
			Description:     pgtype.Text{String: input.Body.Description, Valid: true},
			APIKeyEncrypted: APIKeyEncrypted,
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

		// Ownership is tracked via the owner column in instances
		// No need for separate linking table

		return nil
	})

	if err != nil {
		return nil, huma.Error500InternalServerError(err.Error())
	}

	// Build response
	response := &models.UploadInstanceResponse{}
	response.Body.Owner = owner
	response.Body.InstanceHandle = instanceHandle
	response.Body.InstanceID = int(instanceID)

	return response, nil
}

// Create a llm service (without a handle being present in the URL)
func postInstanceFunc(ctx context.Context, input *models.PostInstanceRequest) (*models.UploadInstanceResponse, error) {
	return putInstanceFunc(ctx, &models.PutInstanceRequest{UserHandle: input.UserHandle, InstanceHandle: input.Body.InstanceHandle, Body: input.Body})
}

func getInstanceFunc(ctx context.Context, input *models.GetInstanceRequest) (*models.GetInstanceResponse, error) {
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
	llm, err := queries.RetrieveInstance(ctx, database.RetrieveInstanceParams{
		Owner:          input.UserHandle,
		InstanceHandle: input.InstanceHandle,
	})
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, huma.Error404NotFound(fmt.Sprintf("llm service %s for user %s not found", input.InstanceHandle, input.UserHandle))
		}
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to retrieve llm service %s for user %s: %v", input.InstanceHandle, input.UserHandle, err))
	}
	if llm.InstanceHandle != input.InstanceHandle {
		return nil, huma.Error404NotFound(fmt.Sprintf("llm service %s for user %s not found", input.InstanceHandle, input.UserHandle))
	}

	// Build response (never return API key in plaintext)
	ls := models.Instance{
		InstanceID:     int(llm.InstanceID),
		Owner:          llm.Owner,
		InstanceHandle: llm.InstanceHandle,
		Endpoint:       llm.Endpoint,
		Description:    llm.Description.String,
		// APIKey:         "", // Never return API key
		APIStandard: llm.APIStandard,
		Model:       llm.Model,
		Dimensions:  int32(llm.Dimensions),
	}
	response := &models.GetInstanceResponse{}
	response.Body = ls

	return response, nil
}

func getUserInstancesFunc(ctx context.Context, input *models.GetUserInstancesRequest) (*models.GetUserInstancesResponse, error) {
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
	llms, err := queries.GetAccessibleInstancesByUser(ctx, database.GetAccessibleInstancesByUserParams{
		Owner:  input.UserHandle,
		Limit:  int32(input.Limit),
		Offset: int32(input.Offset),
	})
	if err != nil {
		if err.Error() == "no rows in result set" {
			// Return empty list instead of error
			response := &models.GetUserInstancesResponse{}
			response.Body.Instances = []models.InstanceOutput{}
			return response, nil
		}
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to retrieve llm services: %v", err))
	}

	// Build response (hide API keys for shared instances)
	ls := []models.InstanceOutput{}
	for _, llm := range llms {
		ls = append(ls, models.InstanceOutput{
			Owner:          llm.Owner,
			InstanceHandle: llm.InstanceHandle,
			InstanceID:     int(llm.InstanceID),
		})
	}
	response := &models.GetUserInstancesResponse{}
	response.Body.Instances = ls

	return response, nil
}

func deleteInstanceFunc(ctx context.Context, input *models.DeleteInstanceRequest) (*models.DeleteInstanceResponse, error) {
	// Check if user exists
	u, err := getUserFunc(ctx, &models.GetUserRequest{UserHandle: input.UserHandle})
	if err != nil {
		return nil, err
	}
	if u.Body.UserHandle != input.UserHandle {
		return nil, huma.Error404NotFound(fmt.Sprintf("user %s not found", input.UserHandle))
	}

	// Check if llm service instance exists
	_, err = getInstanceFunc(ctx, &models.GetInstanceRequest{UserHandle: input.UserHandle, InstanceHandle: input.InstanceHandle})
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
	err = queries.DeleteInstance(ctx, database.DeleteInstanceParams{
		Owner:          input.UserHandle,
		InstanceHandle: input.InstanceHandle,
	})
	if err != nil {
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to delete llm service %s for user %s: %v", input.InstanceHandle, input.UserHandle, err))
	}

	// Build response
	response := &models.DeleteInstanceResponse{}

	return response, nil
}

// RegisterInstancesRoutes registers the routes for the management of LLM service instances
func RegisterInstancesRoutes(pool *pgxpool.Pool, api huma.API) error {
	// Define huma.Operations for each route
	postInstanceOp := huma.Operation{
		OperationID:   "postInstance",
		Method:        http.MethodPost,
		Path:          "/v1/llm-instances/{user_handle}",
		DefaultStatus: http.StatusCreated,
		Summary:       "Create llm service instance",
		Security: []map[string][]string{
			{"adminAuth": []string{"admin"}},
			{"ownerAuth": []string{"owner"}},
		},
		Tags: []string{"llm-instances"},
	}
	putInstanceOp := huma.Operation{
		OperationID:   "putInstance",
		Method:        http.MethodPut,
		Path:          "/v1/llm-instances/{user_handle}/{instance_handle}",
		DefaultStatus: http.StatusCreated,
		Summary:       "Create or update llm service instance",
		Security: []map[string][]string{
			{"adminAuth": []string{"admin"}},
			{"ownerAuth": []string{"owner"}},
		},
		Tags: []string{"llm-instances"},
	}
	getUserInstancesOp := huma.Operation{
		OperationID: "getUserInstances",
		Method:      http.MethodGet,
		Path:        "/v1/llm-instances/{user_handle}",
		Summary:     "Get all llm service instances for a user",
		Security: []map[string][]string{
			{"adminAuth": []string{"admin"}},
			{"ownerAuth": []string{"owner"}},
			{"readerAuth": []string{"reader"}},
		},
		Tags: []string{"llm-instances"},
	}
	getInstanceOp := huma.Operation{
		OperationID: "getInstance",
		Method:      http.MethodGet,
		Path:        "/v1/llm-instances/{user_handle}/{instance_handle}",
		Summary:     "Get a specific llm service instance for a user",
		Security: []map[string][]string{
			{"adminAuth": []string{"admin"}},
			{"ownerAuth": []string{"owner"}},
			{"readerAuth": []string{"reader"}},
		},
		Tags: []string{"llm-instances"},
	}
	deleteInstanceOp := huma.Operation{
		OperationID:   "deleteInstance",
		Method:        http.MethodDelete,
		Path:          "/v1/llm-instances/{user_handle}/{instance_handle}",
		DefaultStatus: http.StatusNoContent,
		Summary:       "Delete a user's llm service instance and all embeddings associated to it",
		Security: []map[string][]string{
			{"adminAuth": []string{"admin"}},
			{"ownerAuth": []string{"owner"}},
		},
		Tags: []string{"llm-instances"},
	}

	huma.Register(api, postInstanceOp, addPoolToContext(pool, postInstanceFunc))
	huma.Register(api, putInstanceOp, addPoolToContext(pool, putInstanceFunc))
	huma.Register(api, getUserInstancesOp, addPoolToContext(pool, getUserInstancesFunc))
	huma.Register(api, getInstanceOp, addPoolToContext(pool, getInstanceFunc))
	huma.Register(api, deleteInstanceOp, addPoolToContext(pool, deleteInstanceFunc))
	return nil
}
