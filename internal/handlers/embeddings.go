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

// Get user and project
func getUserProj(ctx context.Context, user, project string) (string, int32, error) {
	// Check if user and project exist
	u, err := getUserFunc(ctx, &models.GetUserRequest{UserHandle: user})
	if err != nil {
		return "", 0, huma.Error500InternalServerError(fmt.Sprintf("unable to get user %s", user))
	} else if u.Body.UserHandle != user {
		return "", 0, huma.Error404NotFound(fmt.Sprintf("user %s not found", user))
	}
	p, err := getProjectFunc(ctx, &models.GetProjectRequest{UserHandle: user, ProjectHandle: project})
	if err != nil {
		return "", 0, huma.Error500InternalServerError(fmt.Sprintf("unable to get %s's project %s", user, project))
	} else if p.Body.Project.ProjectHandle != project {
		return "", 0, huma.Error404NotFound(fmt.Sprintf("%s's project %s not found", user, project))
	}
	return u.Body.UserHandle, int32(p.Body.Project.ProjectId), nil
}

// Create a new embeddings
func postProjEmbeddingsFunc(ctx context.Context, input *models.PostProjEmbeddingsRequest) (*models.UploadProjEmbeddingsResponse, error) {
	if len(input.Body.Embeddings) == 0 {
		return nil, huma.Error400BadRequest("nothing to do, because len(embeddings) == 0.")
	}

	// Check if user and project exist
	u, p, err := getUserProj(ctx, input.UserHandle, input.ProjectHandle)
	if err != nil {
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to get user %s", input.UserHandle))
	} else if u != input.UserHandle {
		return nil, huma.Error404NotFound(fmt.Sprintf("user %s not found", input.UserHandle))
	} else if p == 0 {
		return nil, huma.Error404NotFound(fmt.Sprintf("project %s not found", input.ProjectHandle))
	}

	// Get the database connection pool from the context
	pool, err := GetDBPool(ctx)
	if err != nil {
		return nil, err
	}

	// For each embedding, build query parameters and run the query
	ids := []string{}
	queries := database.New(pool)
	for _, embedding := range input.Body.Embeddings {
		// Build query parameters (embeddings)
		params := database.UpsertEmbeddingsParams{
			Owner:        u,
			ProjectID:    p,
			TextID:       pgtype.Text{String: embedding.TextID, Valid: true},
			Embedding:    embedding.Vector,
			EmbeddingDim: embedding.VectorDim,
			LLMServiceID: embedding.LLMServiceID,
			Text:         pgtype.Text{String: embedding.Text, Valid: true},
			// TODO: add metadata handling
			// Metadata: embedding.Metadata,
		}
		// Run the queries (upload embeddings)
		result, err := queries.UpsertEmbeddings(ctx, params)
		if err != nil {
			return nil, huma.Error500InternalServerError("unable to upload embeddings")
		}
		ids = append(ids, result.TextID.String)
	}

	// Build response
	response := &models.UploadProjEmbeddingsResponse{}
	response.Body.IDs = ids
	return response, nil
}

func getProjEmbeddingsFunc(ctx context.Context, input *models.GetProjEmbeddingsRequest) (*models.GetProjEmbeddingsResponse, error) {
	// Get the database connection pool from the context
	pool, err := GetDBPool(ctx)
	if err != nil {
		return nil, err
	}

	// Check if user exists
	queries := database.New(pool)
	_, err = queries.RetrieveUser(ctx, input.UserHandle)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, huma.Error404NotFound(fmt.Sprintf("user %s not found", input.UserHandle))
		}
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to check if user %s exists before deleting. %v", input.UserHandle, err))
	}
	user := input.UserHandle

	// Check if project exists
	p, err := queries.RetrieveProject(ctx, database.RetrieveProjectParams{Owner: input.UserHandle, ProjectHandle: input.ProjectHandle})
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, huma.Error404NotFound(fmt.Sprintf("project %s not found for user %s", input.ProjectHandle, input.UserHandle))
		}
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to check if project %s exists before deleting. %v", input.ProjectHandle, err))
	}
	project := p.ProjectHandle

	// Build query parameters (embeddings)
	params := database.GetEmbeddingsByProjectParams{
		Owner:         user,
		ProjectHandle: project,
		Limit:         int32(input.Limit),
		Offset:        int32(input.Offset),
	}

	// Run the query
	embeddings, err := queries.GetEmbeddingsByProject(ctx, params)
	if err != nil {
		return nil, huma.Error500InternalServerError("unable to get embeddings")
	}
	if len(embeddings) == 0 {
		return nil, huma.Error404NotFound("no embeddings found.")
	}

	// Build the response
	e := []models.Embeddings{}
	for _, embedding := range embeddings {
		e = append(e, models.Embeddings{
			TextID:       embedding.TextID.String,
			Vector:       embedding.Embedding,
			VectorDim:    embedding.EmbeddingDim,
			LLMServiceID: embedding.LLMServiceID,
			Text:         embedding.Text.String,
		})
	}
	response := &models.GetProjEmbeddingsResponse{}
	response.Body.Embeddings = e
	return response, nil
}

func deleteProjEmbeddingsFunc(ctx context.Context, input *models.DeleteProjEmbeddingsRequest) (*models.DeleteProjEmbeddingsResponse, error) {
	// Check if user and project exist
	u, p, err := getUserProj(ctx, input.UserHandle, input.ProjectHandle)
	if err != nil {
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to get user %s", input.UserHandle))
	} else if u != input.UserHandle {
		return nil, huma.Error404NotFound(fmt.Sprintf("user %s not found", input.UserHandle))
	} else if p == 0 {
		return nil, huma.Error404NotFound(fmt.Sprintf("project %s not found", input.ProjectHandle))
	}

	// Get the database connection pool from the context
	pool, err := GetDBPool(ctx)
	if err != nil {
		return nil, err
	}

	// Build query parameters (embeddings)
	params := database.DeleteEmbeddingsByProjectParams{
		Owner:         u,
		ProjectHandle: input.ProjectHandle,
	}

	// Run the query
	queries := database.New(pool)
	err = queries.DeleteEmbeddingsByProject(ctx, params)
	if err != nil {
		return nil, huma.Error500InternalServerError("unable to delete embeddings")
	}

	// Build the response
	response := &models.DeleteProjEmbeddingsResponse{}
	response.Body = fmt.Sprintf("Successfully deleted all embeddings for %s's project %s", input.UserHandle, input.ProjectHandle)

	return response, nil
}

func getDocEmbeddingsFunc(ctx context.Context, input *models.GetDocEmbeddingsRequest) (*models.GetDocEmbeddingsResponse, error) {
	// Check if user and project exist
	u, p, err := getUserProj(ctx, input.UserHandle, input.ProjectHandle)
	if err != nil {
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to get user %s", input.UserHandle))
	} else if u != input.UserHandle {
		return nil, huma.Error404NotFound(fmt.Sprintf("user %s not found", input.UserHandle))
	} else if p == 0 {
		return nil, huma.Error404NotFound(fmt.Sprintf("project %s not found", input.ProjectHandle))
	}

	// Get the database connection pool from the context
	pool, err := GetDBPool(ctx)
	if err != nil {
		return nil, err
	}

	// Build query parameters (embeddings)
	params := database.RetrieveEmbeddingsParams{
		Owner:         u,
		ProjectHandle: input.ProjectHandle,
		TextID:        pgtype.Text{String: input.TextID, Valid: true},
	}

	// Run the query
	queries := database.New(pool)
	embedding, err := queries.RetrieveEmbeddings(ctx, params)
	if err != nil {
		return nil, huma.Error500InternalServerError("unable to get embeddings")
	}
	if embedding.TextID.String == "" {
		return nil, huma.Error404NotFound("no embeddings found.")
	}

	// Build the response
	e := models.Embeddings{
		TextID:       embedding.TextID.String,
		Vector:       embedding.Embedding,
		VectorDim:    embedding.EmbeddingDim,
		LLMServiceID: embedding.LLMServiceID,
		Text:         embedding.Text.String,
	}
	response := &models.GetDocEmbeddingsResponse{}
	response.Body.Embeddings = e

	return response, nil
}

func deleteDocEmbeddingsFunc(ctx context.Context, input *models.DeleteDocEmbeddingsRequest) (*models.DeleteDocEmbeddingsResponse, error) {
	// Check if user and project exist
	u, p, err := getUserProj(ctx, input.UserHandle, input.ProjectHandle)
	if err != nil {
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to get user %s", input.UserHandle))
	} else if u != input.UserHandle {
		return nil, huma.Error404NotFound(fmt.Sprintf("user %s not found", input.UserHandle))
	} else if p == 0 {
		return nil, huma.Error404NotFound(fmt.Sprintf("project %s not found", input.ProjectHandle))
	}

	// Get the database connection pool from the context
	pool, err := GetDBPool(ctx)
	if err != nil {
		return nil, err
	}

	// Build query parameters (embeddings)
	params := database.DeleteDocEmbeddingsParams{
		Owner:         u,
		ProjectHandle: input.ProjectHandle,
		TextID:        pgtype.Text{String: input.TextID, Valid: true},
	}

	// Run the query
	queries := database.New(pool)
	err = queries.DeleteDocEmbeddings(ctx, params)
	if err != nil {
		return nil, huma.Error500InternalServerError("unable to delete embeddings")
	}

	// Build the response
	response := &models.DeleteDocEmbeddingsResponse{}
	response.Body = fmt.Sprintf("Successfully deleted embeddings for document %s (%s's project %s)", input.TextID, input.UserHandle, input.ProjectHandle)
	return response, nil
}

// RegisterEmbeddingsRoutes registers all the embeddings routes with the API
func RegisterEmbeddingsRoutes(pool *pgxpool.Pool, keyGen RandomKeyGenerator, api huma.API) error {
	// Define huma.Operations for each route
	postProjEmbeddingsOp := huma.Operation{
		OperationID: "postEmbeddings",
		Method:      http.MethodPost,
		Path:        "/embeddings/{user_handle}/{project_handle}",
		Summary:     "Create embeddings for a project",
		Security: []map[string][]string{
			{"adminAuth": []string{"admin"}},
			{"ownerAuth": []string{"owner"}},
		},
		Tags: []string{"embeddings"},
	}
	getProjEmbeddingsOp := huma.Operation{
		OperationID: "getEmbeddings",
		Method:      http.MethodGet,
		Path:        "/embeddings/{user_handle}/{project_handle}",
		Summary:     "Get all embeddings for a project",
		Security: []map[string][]string{
			{"adminAuth": []string{"admin"}},
			{"ownerAuth": []string{"owner"}},
			{"readerAuth": []string{"reader"}},
		},
		Tags: []string{"embeddings"},
	}
	deleteProjEmbeddingsOp := huma.Operation{
		OperationID: "deleteEmbeddings",
		Method:      http.MethodDelete,
		Path:        "/embeddings/{user_handle}/{project_handle}",
		Summary:     "Delete all embeddings for a project",
		Security: []map[string][]string{
			{"adminAuth": []string{"admin"}},
			{"ownerAuth": []string{"owner"}},
		},
		Tags: []string{"embeddings"},
	}
	getDocEmbeddingsOp := huma.Operation{
		OperationID: "getDocEmbeddings",
		Method:      http.MethodGet,
		Path:        "/embeddings/{user_handle}/{project_handle}/{text_id}",
		Summary:     "Get embeddings for a specific document",
		Security: []map[string][]string{
			{"adminAuth": []string{"admin"}},
			{"ownerAuth": []string{"owner"}},
			{"readerAuth": []string{"reader"}},
		},
		Tags: []string{"embeddings"},
	}
	deleteDocEmbeddingsOp := huma.Operation{
		OperationID: "deleteDocEmbeddings",
		Method:      http.MethodDelete,
		Path:        "/embeddings/{user_handle}/{project_handle}/{text_id}",
		Summary:     "Delete embeddings for a specific document",
		Security: []map[string][]string{
			{"adminAuth": []string{"admin"}},
			{"ownerAuth": []string{"owner"}},
		},
		Tags: []string{"embeddings"},
	}

	// huma.Register(api, putProjEmbeddingsOp, addPoolToContext(pool, putProjEmbeddingsFunc))
	huma.Register(api, postProjEmbeddingsOp, addPoolToContext(pool, postProjEmbeddingsFunc))
	huma.Register(api, getProjEmbeddingsOp, addPoolToContext(pool, getProjEmbeddingsFunc))
	huma.Register(api, deleteProjEmbeddingsOp, addPoolToContext(pool, deleteProjEmbeddingsFunc))
	huma.Register(api, getDocEmbeddingsOp, addPoolToContext(pool, getDocEmbeddingsFunc))
	huma.Register(api, deleteDocEmbeddingsOp, addPoolToContext(pool, deleteDocEmbeddingsFunc))
	return nil
}
