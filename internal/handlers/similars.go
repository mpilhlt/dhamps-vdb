package handlers

import (
	"context"
	"net/http"

	"github.com/mpilhlt/dhamps-vdb/internal/database"
	"github.com/mpilhlt/dhamps-vdb/internal/models"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Define handler functions for each route
func getSimilarFunc(ctx context.Context, input *models.GetSimilarRequest) (*models.SimilarResponse, error) {
	// Get the database connection pool from the context
	pool, err := GetDBPool(ctx)
	if err != nil {
		return nil, err
	}

	// Build query parameters

	// Run the query
	queries := database.New(pool)
	sim, err := queries.GetSimilarsByID(ctx, database.GetSimilarsByIDParams{
		// TODO: Add User and Project fields
		// User: input.User,
		// Project: pgtype.Text{ String: input.Project, Valid: true },
		TextID: pgtype.Text{String: input.TextID, Valid: true},
		Limit:  min(int32(input.Limit), int32(input.Count)),
		Offset: int32(input.Offset),
	})
	if err != nil {
		return nil, huma.Error500InternalServerError("unable to get similar items")
	}
	if len(sim) == 0 {
		return nil, huma.Error404NotFound("no similar items found")
	}

	// Build response
	s := []string{}
	for _, r := range sim {
		s = append(s, r.TextID.String)
	}
	response := &models.SimilarResponse{}
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
