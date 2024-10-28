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

// putUserFunc creates or updates a user
func putUserFunc(ctx context.Context, input *models.PutUserRequest) (*models.UploadUserResponse, error) {
	if input.Handle != input.Body.Handle {
		return nil, huma.Error400BadRequest(fmt.Sprintf("user handle in URL (%s) does not match user handle in body (%v).", input.Handle, input.Body.Handle))
	}

	// Get the database connection pool from the context
	pool, err := GetDBPool(ctx)
	if err != nil {
		return nil, err
	} else if pool == nil {
		return nil, huma.Error500InternalServerError("database connection pool is nil")
	}

	// Get the API key generator from the context
	keyGen, err := GetKeyGen(ctx)
	if err != nil {
		return nil, err
	}

	// Build query parameters (user - eventually with new API key)
	// Check if user already exists
	queries := database.New(pool)
	u, err := queries.RetrieveUser(ctx, input.Handle)
	if err != nil && err.Error() != "no rows in result set" {
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to check if user %s already exists. %v", input.Handle, err))
	}
	api_key := ""
	if u.Handle == input.Handle {
		// User exists, so don't create API key
		api_key = u.VdbApiKey
	} else {
		// User does not exist, so create a new API key
		k, err := keyGen.RandomKey(64)
		if err != nil {
			return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to create API key for user %s. %v", input.Handle, err))
		}
		api_key = k
	}
	user := database.UpsertUserParams{
		Handle:    input.Handle,
		Name:      pgtype.Text{String: input.Body.Name, Valid: true},
		Email:     input.Body.Email,
		VdbApiKey: api_key,
	}

	// Run the query
	u, err = queries.UpsertUser(ctx, user)
	if err != nil {
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to upload user. %v", err))
	}

	// Build the response
	response := &models.UploadUserResponse{}
	response.Body.Handle = u.Handle
	response.Body.APIKey = u.VdbApiKey
	return response, nil
}

// Create a user (without a handle being present in the URL)
func postUserFunc(ctx context.Context, input *models.PostUserRequest) (*models.UploadUserResponse, error) {
	u, err := putUserFunc(ctx, &models.PutUserRequest{Handle: input.Body.Handle, Body: input.Body})
	if err != nil {
		return nil, err
	}
	return u, nil
}

// Get all users
func getUsersFunc(ctx context.Context, input *models.GetUsersRequest) (*models.GetUsersResponse, error) {
	// Get the database connection pool from the context
	pool, err := GetDBPool(ctx)
	if err != nil {
		return nil, err
	} else if pool == nil {
		return nil, huma.Error500InternalServerError("database connection pool is nil")
	}

	// Run the query
	queries := database.New(pool)
	allUsers, err := queries.GetUsers(ctx, database.GetUsersParams{Limit: int32(input.Limit), Offset: int32(input.Offset)})
	if err != nil {
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to get list of users. %v", err))
	}
	if len(allUsers) == 0 {
		return nil, huma.Error404NotFound("no users found.")
	}

	// Build the response
	response := &models.GetUsersResponse{}
	response.Body = allUsers

	return response, nil
}

// Get a specific user
func getUserFunc(ctx context.Context, input *models.GetUserRequest) (*models.GetUserResponse, error) {
	// Get the database connection pool from the context
	pool, err := GetDBPool(ctx)
	if err != nil {
		return nil, err
	} else if pool == nil {
		return nil, huma.Error500InternalServerError("database connection pool is nil")
	}

	// Run the query
	queries := database.New(pool)
	u, err := queries.RetrieveUser(ctx, input.Handle)
	if err != nil {
		// return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to get user data for user %s. %v", input.User, err))
		return nil, huma.Error404NotFound(fmt.Sprintf("user %s not found. %v", input.Handle, err))
	}

	// Build the response
	returnUser := &models.User{
		Handle: u.Handle,
		Name:   u.Name.String,
		Email:  u.Email,
		APIKey: u.VdbApiKey,
	}
	response := &models.GetUserResponse{}
	response.Body = *returnUser

	return response, nil
}

// Delete a specific user
func deleteUserFunc(ctx context.Context, input *models.DeleteUserRequest) (*models.DeleteUserResponse, error) {
	// Get the database connection pool from the context
	pool, err := GetDBPool(ctx)
	if err != nil {
		return nil, err
	} else if pool == nil {
		return nil, huma.Error500InternalServerError("database connection pool is nil")
	}

	// Check if user exists
	queries := database.New(pool)
	_, err = queries.RetrieveUser(ctx, input.Handle)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, huma.Error404NotFound(fmt.Sprintf("user %s not found", input.Handle))
		}
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to check if user %s exists before deleting. %v", input.Handle, err))
	}

	// Run the query
	err = queries.DeleteUser(ctx, input.Handle)
	if err != nil {
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to delete user %s. %v", input.Handle, err))
	}

	// Build the response
	response := &models.DeleteUserResponse{}
	return response, nil
}

// RegisterUsersRoutes registers all the admin routes with the API
func RegisterUsersRoutes(pool *pgxpool.Pool, keyGen RandomKeyGenerator, api huma.API) error {
	// Define huma.Operations for each route
	putUserOp := huma.Operation{
		OperationID:   "putUser",
		Method:        http.MethodPut,
		Path:          "/users/{handle}",
		DefaultStatus: http.StatusCreated,
		Summary:       "Create or update a user",
		Tags:          []string{"admin", "users"},
	}
	postUserOp := huma.Operation{
		OperationID:   "postUser",
		Method:        http.MethodPost,
		Path:          "/users",
		DefaultStatus: http.StatusCreated,
		Summary:       "Create a user",
		Tags:          []string{"admin", "users"},
	}
	getUsersOp := huma.Operation{
		OperationID: "getUsers",
		Method:      http.MethodGet,
		Path:        "/users",
		Summary:     "Get information about all users",
		Tags:        []string{"admin", "users"},
	}
	getUserOp := huma.Operation{
		OperationID: "getUser",
		Method:      http.MethodGet,
		Path:        "/users/{handle}",
		Summary:     "Get information about a specific user",
		Tags:        []string{"admin", "users"},
	}
	deleteUserOp := huma.Operation{
		OperationID:   "deleteUser",
		Method:        http.MethodDelete,
		Path:          "/users/{handle}",
		DefaultStatus: http.StatusNoContent,
		Summary:       "Delete a specific user",
		Tags:          []string{"admin", "users"},
	}

	// Register the routes with middleware
	huma.Register(api, putUserOp, addPoolToContext(pool, addKeyGenToContext(keyGen, putUserFunc)))
	huma.Register(api, postUserOp, addPoolToContext(pool, addKeyGenToContext(keyGen, postUserFunc)))
	huma.Register(api, getUsersOp, addPoolToContext(pool, getUsersFunc))
	huma.Register(api, getUserOp, addPoolToContext(pool, getUserFunc))
	huma.Register(api, deleteUserOp, addPoolToContext(pool, deleteUserFunc))
	return nil
}
