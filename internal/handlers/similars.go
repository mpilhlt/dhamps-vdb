package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/mpilhlt/dhamps-vdb/internal/database"
	"github.com/mpilhlt/dhamps-vdb/internal/models"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TODO: Allow to get similars to a submission that includes ready-made embeddings

// Define handler functions for each route
func getSimilarFunc(ctx context.Context, input *models.GetSimilarRequest) (*models.SimilarResponse, error) {
	// Check if only one of input.MetadataField and input.MetadataValue are given
	if input.MetadataPath != "" && input.MetadataValue == "" {
		return nil, huma.Error400BadRequest("metadata_path is set but metadata_value is not")
	}
	if input.MetadataPath == "" && input.MetadataValue != "" {
		return nil, huma.Error400BadRequest("metadata_value is set but metadata_path is not")
	}

	// Check if user exists
	_, err := getUserFunc(ctx, &models.GetUserRequest{UserHandle: input.UserHandle})
	if err != nil {
		return nil, err
	}

	// Check if project exists
	_, err = getProjectFunc(ctx, &models.GetProjectRequest{UserHandle: input.UserHandle, ProjectHandle: input.ProjectHandle})
	if err != nil {
		return nil, err
	}

	// Check if text exists
	_, err = getDocEmbeddingsFunc(ctx, &models.GetDocEmbeddingsRequest{UserHandle: input.UserHandle, ProjectHandle: input.ProjectHandle, TextID: input.TextID})
	// fmt.Printf("getting doc embeddings for %s\n", input.TextID)
	if err != nil {
		return nil, err
	}

	// Get the database connection pool from the context
	pool, err := GetDBPool(ctx)
	if err != nil {
		return nil, err
	}

	// Run the query, either with or without metadata filter
	queries := database.New(pool)
	var sim []pgtype.Text

	if input.MetadataPath == "" {
		params := database.GetSimilarsByIDParams{
			TextID:        pgtype.Text{String: url.QueryEscape(input.TextID), Valid: true},
			Owner:         input.UserHandle,
			ProjectHandle: input.ProjectHandle,
			Column4:       input.Threshold,
			Limit:         min(int32(input.Limit), int32(input.Count)),
			Offset:        int32(input.Offset),
		}
		fmt.Printf("getting similar items for %v\n", params)
		sim, err = queries.GetSimilarsByID(ctx, params)
	} else {
		params := database.GetSimilarsByIDWithFilterParams{
			TextID:        pgtype.Text{String: url.QueryEscape(input.TextID), Valid: true},
			Owner:         input.UserHandle,
			ProjectHandle: input.ProjectHandle,
			Column4:       input.Threshold,
			Column5:       input.MetadataPath,
			Column6:       input.MetadataValue,
			Limit:         min(int32(input.Limit), int32(input.Count)),
			Offset:        int32(input.Offset),
		}
		fmt.Printf("getting similar items for %v\n", params)
		sim, err = queries.GetSimilarsByIDWithFilter(ctx, params)
	}
	fmt.Printf("got this response from the database: %v\n", sim)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, huma.Error404NotFound("no similar items found")
		}
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to get similar items. %v", err))
	}
	if len(sim) == 0 {
		return nil, huma.Error404NotFound("no similar items found")
	}

	// Build response
	s := []string{}
	for _, r := range sim {
		s = append(s, r.String)
	}
	response := &models.SimilarResponse{}
	response.Body.UserHandle = input.UserHandle
	response.Body.ProjectHandle = input.ProjectHandle
	response.Body.IDs = s
	return response, nil
}

func postSimilarFunc(ctx context.Context, input *models.PostSimilarRequest) (*models.SimilarResponse, error) {
	// Implement your logic here
	return nil, nil
}

// RegisterSimilarRoutes registers the routes for the Similar service
func RegisterSimilarRoutes(pool *pgxpool.Pool, api huma.API) error {
	// Define huma.Operations for each route
	getSimilarOp := huma.Operation{
		OperationID: "getSimilar",
		Method:      http.MethodGet,
		Path:        "/v1/similars/{user_handle}/{project_handle}/{text_id}",
		Summary:     "Retrieve similar items for a particular document",
		Security: []map[string][]string{
			{"adminAuth": []string{"admin"}},
			{"ownerAuth": []string{"owner"}},
			{"readerAuth": []string{"reader"}},
		},
		Tags: []string{"similars"},
	}
	postSimilarOp := huma.Operation{
		OperationID: "postSimilar",
		Method:      http.MethodPost,
		Path:        "/v1/similars/{user_handle}/{project_handle}",
		Summary:     "Retrieve similar items for a query document",
		Security: []map[string][]string{
			{"adminAuth": []string{"admin"}},
			{"ownerAuth": []string{"owner"}},
			{"readerAuth": []string{"reader"}},
		},
		Tags: []string{"similars"},
	}

	huma.Register(api, getSimilarOp, addPoolToContext(pool, getSimilarFunc))
	huma.Register(api, postSimilarOp, addPoolToContext(pool, postSimilarFunc))
	return nil
}
