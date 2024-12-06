package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/mpilhlt/dhamps-vdb/internal/database"
	"github.com/mpilhlt/dhamps-vdb/internal/models"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

// Get user and project
func getUserProj(ctx context.Context, user, project string) (string, string, int32, error) {
	// Check if user and project exist
	u, err := getUserFunc(ctx, &models.GetUserRequest{UserHandle: user})
	if err != nil {
		if err.Error() == "no rows in result set" || err.Error() == fmt.Sprintf("user %s not found", user) {
			return "", "", 0, huma.Error404NotFound(fmt.Sprintf("user %s not found", user))
		}
		return "", "", 0, huma.Error500InternalServerError(fmt.Sprintf("unable to get user %s. %v", user, err))
	}
	if u.Body.UserHandle != user {
		return "", "", 0, huma.Error404NotFound(fmt.Sprintf("user %s not found", user))
	}
	p, err := getProjectFunc(ctx, &models.GetProjectRequest{UserHandle: user, ProjectHandle: project})
	if err != nil {
		if err.Error() == "no rows in result set" || err.Error() == fmt.Sprintf("user %s's project %s not found", user, project) {
			return "", "", 0, huma.Error404NotFound(fmt.Sprintf("%s's project %s not found", user, project))
		}
		return "", "", 0, huma.Error500InternalServerError(fmt.Sprintf("unable to get %s's project %s. %v", user, project, err))
	}
	if p.Body.ProjectHandle != project {
		return "", "", 0, huma.Error404NotFound(fmt.Sprintf("%s's project %s not found", user, project))
	}
	return u.Body.UserHandle, p.Body.ProjectHandle, int32(p.Body.ProjectID), nil
}

// Create a new embeddings
func postProjEmbeddingsFunc(ctx context.Context, input *models.PostProjEmbeddingsRequest) (*models.UploadProjEmbeddingsResponse, error) {
	// Check if user and project exist
	_, _, pid, err := getUserProj(ctx, input.UserHandle, input.ProjectHandle)
	if err != nil {
		return nil, err
	}

	// Check if llm service exists
	llm, err := getLLMFunc(ctx, &models.GetLLMRequest{UserHandle: input.UserHandle, LLMServiceHandle: input.Body.Embeddings[0].LLMServiceHandle})
	if err != nil {
		return nil, err
	}
	llmid := int32(llm.Body.LLMServiceID)

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
			TextID:       pgtype.Text{String: embedding.TextID, Valid: true},
			Owner:        input.UserHandle,
			ProjectID:    pid,
			LLMServiceID: llmid,
			Text:         pgtype.Text{String: embedding.Text, Valid: true},
			Vector:       pgvector.NewHalfVector(embedding.Vector),
			VectorDim:    embedding.VectorDim,
			Metadata:     embedding.Metadata,
		}
		// Run the queries (upload embeddings)
		result, err := queries.UpsertEmbeddings(ctx, params)
		if err != nil {
			fmt.Printf("Error: %v\n(Params were: %v)\n", err, params)
			return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to upload embeddings. %v", err))
		}
		ids = append(ids, result.TextID.String)
	}

	// Build response
	response := &models.UploadProjEmbeddingsResponse{}
	response.Body.IDs = ids
	return response, nil
}

func getProjEmbeddingsFunc(ctx context.Context, input *models.GetProjEmbeddingsRequest) (*models.GetProjEmbeddingsResponse, error) {
	// Check if user exists
	if _, err := getUserFunc(ctx, &models.GetUserRequest{UserHandle: input.UserHandle}); err != nil {
		return nil, err
	}

	// Check if project exists
	if _, err := getProjectFunc(ctx, &models.GetProjectRequest{UserHandle: input.UserHandle, ProjectHandle: input.ProjectHandle}); err != nil {
		return nil, err
	}

	// Get the database connection pool from the context
	pool, err := GetDBPool(ctx)
	if err != nil {
		return nil, err
	}

	// Build query parameters (embeddings)
	params := database.GetEmbeddingsByProjectParams{
		Owner:         input.UserHandle,
		ProjectHandle: input.ProjectHandle,
		Limit:         int32(input.Limit),
		Offset:        int32(input.Offset),
	}

	// Run the query
	queries := database.New(pool)
	embeddingss, err := queries.GetEmbeddingsByProject(ctx, params)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, huma.Error404NotFound(fmt.Sprintf("no embeddings found for user %s, project %s.", input.UserHandle, input.ProjectHandle))
		}
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to get embeddings for user %s, project %s. %v", input.UserHandle, input.ProjectHandle, err))
	}
	if len(embeddingss) == 0 {
		return nil, huma.Error404NotFound(fmt.Sprintf("no embeddings found for user %s, project %s.", input.UserHandle, input.ProjectHandle))
	}

	// Build the response
	e := []models.EmbeddingsDatabase{}
	for _, embeddings := range embeddingss {
		md := map[string]interface{}{}
		err = json.Unmarshal(embeddings.Metadata, &md)
		if err != nil {
			return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to unmarshal metadata for user %s, project %s, id %s. Metadata: %s. %v", input.UserHandle, input.ProjectHandle, embeddings.TextID.String, string(embeddings.Metadata), err))
		}
		e = append(e, models.EmbeddingsDatabase{
			TextID:           embeddings.TextID.String,
			UserHandle:       embeddings.Owner,
			ProjectHandle:    embeddings.ProjectHandle,
			ProjectID:        int(embeddings.ProjectID),
			LLMServiceHandle: embeddings.LLMServiceHandle,
			Vector:           embeddings.Vector.Slice(),
			VectorDim:        embeddings.VectorDim,
			Text:             embeddings.Text.String,
			Metadata:         md,
		})
	}
	response := &models.GetProjEmbeddingsResponse{}
	response.Body.Embeddings = e
	return response, nil
}

func deleteProjEmbeddingsFunc(ctx context.Context, input *models.DeleteProjEmbeddingsRequest) (*models.DeleteProjEmbeddingsResponse, error) {
	// Check if user and project exist
	_, _, _, err := getUserProj(ctx, input.UserHandle, input.ProjectHandle)
	if err != nil {
		if err.Error() == "no rows in result set" || err == huma.Error404NotFound(fmt.Sprintf("user %s's project %s not found", input.UserHandle, input.ProjectHandle)) {
			return nil, huma.Error404NotFound(fmt.Sprintf("project %s of user %s not found", input.ProjectHandle, input.UserHandle))
		}
		return nil, err
	}

	// Get the database connection pool from the context
	pool, err := GetDBPool(ctx)
	if err != nil {
		return nil, err
	}

	// Build query parameters (embeddings)
	params := database.DeleteEmbeddingsByProjectParams{
		Owner:         input.UserHandle,
		ProjectHandle: input.ProjectHandle,
	}

	// Run the query
	queries := database.New(pool)
	err = queries.DeleteEmbeddingsByProject(ctx, params)
	if err != nil {
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to delete embeddings for %s's project %s. %v", input.UserHandle, input.ProjectHandle, err))
	}

	// Build the response
	response := &models.DeleteProjEmbeddingsResponse{}
	return response, nil
}

func getDocEmbeddingsFunc(ctx context.Context, input *models.GetDocEmbeddingsRequest) (*models.GetDocEmbeddingsResponse, error) {
	// Check if user and project exist
	_, _, _, err := getUserProj(ctx, input.UserHandle, input.ProjectHandle)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, huma.Error404NotFound(fmt.Sprintf("project %s of user %s not found", input.ProjectHandle, input.UserHandle))
		}
		return nil, err
	}

	// Get the database connection pool from the context
	pool, err := GetDBPool(ctx)
	if err != nil {
		return nil, err
	}

	textid := url.QueryEscape(input.TextID)

	// Build query parameters (embeddings)
	params := database.RetrieveEmbeddingsParams{
		Owner:         input.UserHandle,
		ProjectHandle: input.ProjectHandle,
		TextID:        pgtype.Text{String: textid, Valid: true},
	}

	// fmt.Printf("getDocEmbeddings, textid: %v\n", textid)

	// Run the query
	queries := database.New(pool)
	embeddings, err := queries.RetrieveEmbeddings(ctx, params)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, huma.Error404NotFound(fmt.Sprintf("no embeddings found for user %s, project %s, id %s.", input.UserHandle, input.ProjectHandle, textid))
		}
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to get embeddings for user %s, project %s, id %s. %v", input.UserHandle, input.ProjectHandle, textid, err))
	}
	if embeddings.TextID.String == "" {
		return nil, huma.Error404NotFound(fmt.Sprintf("no embeddings found for user %s, project %s, id %s.", input.UserHandle, input.ProjectHandle, textid))
	}

	// Build the response
	md := map[string]interface{}{}
	err = json.Unmarshal(embeddings.Metadata, &md)
	if err != nil {
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to unmarshal metadata for user %s, project %s, id %s. Metadata: %s. %v", input.UserHandle, input.ProjectHandle, embeddings.TextID.String, string(embeddings.Metadata), err))
	}
	e := models.EmbeddingsDatabase{
		TextID:           embeddings.TextID.String,
		UserHandle:       embeddings.Owner,
		ProjectHandle:    embeddings.ProjectHandle,
		ProjectID:        int(embeddings.ProjectID),
		LLMServiceHandle: embeddings.LLMServiceHandle,
		Vector:           embeddings.Vector.Slice(),
		VectorDim:        embeddings.VectorDim,
		Text:             embeddings.Text.String,
		Metadata:         md,
	}
	response := &models.GetDocEmbeddingsResponse{}
	response.Body = e

	return response, nil
}

func deleteDocEmbeddingsFunc(ctx context.Context, input *models.DeleteDocEmbeddingsRequest) (*models.DeleteDocEmbeddingsResponse, error) {
	// Check if user and project exist
	_, _, _, err := getUserProj(ctx, input.UserHandle, input.ProjectHandle)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, huma.Error404NotFound(fmt.Sprintf("project %s of user %s not found", input.ProjectHandle, input.UserHandle))
		}
		return nil, err
	}

	textid := url.QueryEscape(input.TextID)

	// Check if embeddings with ID exist
	textidForChecking := input.TextID // the getDocEmbeddings expects a url-decoded path parameter
	_, err = getDocEmbeddingsFunc(ctx, &models.GetDocEmbeddingsRequest{UserHandle: input.UserHandle, ProjectHandle: input.ProjectHandle, TextID: textidForChecking})
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, huma.Error404NotFound(fmt.Sprintf("text id %s in %s's project %s not found", textid, input.UserHandle, input.ProjectHandle))
		}
		return nil, err
	}

	// Get the database connection pool from the context
	pool, err := GetDBPool(ctx)
	if err != nil {
		return nil, err
	}

	// Build query parameters for DeleteEmbeddings
	params := database.DeleteDocEmbeddingsParams{
		Owner:         input.UserHandle,
		ProjectHandle: input.ProjectHandle,
		TextID:        pgtype.Text{String: textid, Valid: true},
	}

	// fmt.Printf("deleteDocEmbeddings, textid: %v\n", textid)

	// Run the query
	queries := database.New(pool)
	err = queries.DeleteDocEmbeddings(ctx, params)
	if err != nil {
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to delete embeddings for text id %s in %s's project %s. %v", textid, input.UserHandle, input.ProjectHandle, err))
	}

	// Build the response
	response := &models.DeleteDocEmbeddingsResponse{}
	return response, nil
}

// RegisterEmbeddingsRoutes registers all the embeddings routes with the API
func RegisterEmbeddingsRoutes(pool *pgxpool.Pool, api huma.API) error {
	// Define huma.Operations for each route
	postProjEmbeddingsOp := huma.Operation{
		OperationID:   "postEmbeddings",
		Method:        http.MethodPost,
		Path:          "/v1/embeddings/{user_handle}/{project_handle}",
		DefaultStatus: http.StatusCreated,
		Summary:       "Create embeddings for a project",
		Security: []map[string][]string{
			{"adminAuth": []string{"admin"}},
			{"ownerAuth": []string{"owner"}},
		},
		Tags: []string{"embeddings"},
	}
	getProjEmbeddingsOp := huma.Operation{
		OperationID: "getEmbeddings",
		Method:      http.MethodGet,
		Path:        "/v1/embeddings/{user_handle}/{project_handle}",
		Summary:     "Get all embeddings for a project",
		Security: []map[string][]string{
			{"adminAuth": []string{"admin"}},
			{"ownerAuth": []string{"owner"}},
			{"readerAuth": []string{"reader"}},
		},
		Tags: []string{"embeddings"},
	}
	deleteProjEmbeddingsOp := huma.Operation{
		OperationID:   "deleteEmbeddings",
		Method:        http.MethodDelete,
		Path:          "/v1/embeddings/{user_handle}/{project_handle}",
		DefaultStatus: http.StatusNoContent,
		Summary:       "Delete all embeddings for a project",
		Security: []map[string][]string{
			{"adminAuth": []string{"admin"}},
			{"ownerAuth": []string{"owner"}},
		},
		Tags: []string{"embeddings"},
	}
	getDocEmbeddingsOp := huma.Operation{
		OperationID: "getDocEmbeddings",
		Method:      http.MethodGet,
		Path:        "/v1/embeddings/{user_handle}/{project_handle}/{text_id}",
		Summary:     "Get embeddings for a specific document",
		Security: []map[string][]string{
			{"adminAuth": []string{"admin"}},
			{"ownerAuth": []string{"owner"}},
			{"readerAuth": []string{"reader"}},
		},
		Tags: []string{"embeddings"},
	}
	deleteDocEmbeddingsOp := huma.Operation{
		OperationID:   "deleteDocEmbeddings",
		Method:        http.MethodDelete,
		Path:          "/v1/embeddings/{user_handle}/{project_handle}/{text_id}",
		DefaultStatus: http.StatusNoContent,
		Summary:       "Delete embeddings for a specific document",
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
