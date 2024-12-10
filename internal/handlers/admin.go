package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mpilhlt/dhamps-vdb/internal/database"
	"github.com/mpilhlt/dhamps-vdb/internal/models"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func resetDbFunc(ctx context.Context, input *models.ResetDbRequest) (*models.ResetDbResponse, error) {
	// Get the database connection pool from the context
	pool, err := GetDBPool(ctx)
	if err != nil {
		fmt.Printf("    Resetting Database: error getting database connection pool: %v\n", err)
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to get database connection pool. %v", err))
	} else if pool == nil {
		fmt.Print("    Resetting Database: database connection pool is nil\n")
		return nil, huma.Error500InternalServerError("database connection pool is nil")
	}

	queries := database.New(pool)

	// delete all records
	fmt.Print("    Resetting Database: deleting all records...\n")
	err = queries.DeleteAllRecords(ctx)
	if err != nil {
		fmt.Printf("    Resetting Database: error deleting all records: %v\n", err)
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to delete all records. %v", err))
	}

	fmt.Print("    Resetting Database: resetting serials...\n")
	err = queries.ResetAllSerials(ctx)
	if err != nil {
		fmt.Printf("    Resetting Database: error resetting serials: %v\n", err)
		return nil, huma.Error500InternalServerError(fmt.Sprintf("unable to reset serials. %v", err))
	}

	// Build response
	response := &models.ResetDbResponse{}
	return response, nil
}

// RegisterUsersRoutes registers all the admin routes with the API
func RegisterAdminRoutes(pool *pgxpool.Pool, api huma.API) error {
	// Define huma.Operations for each route
	footgunOp := huma.Operation{
		OperationID:   "footgun",
		Method:        http.MethodGet,
		Path:          "/v1/admin/footgun",
		DefaultStatus: http.StatusNoContent,
		Summary:       "Remove all records from database and reset serials/counters",
		Security: []map[string][]string{
			{"adminAuth": []string{"admin"}},
		},
		// MaxBodyBytes int64 `yaml:"-"` // Max size of the request body in bytes (-1 for unlimited)
		// BodyReadTimeout time.Duration `yaml:"-" // Time to wait for the request body to be read (-1 for unlimited)
		// Middlewares Middlewares `yaml:"-"` // Middleware to run before the operation, useful for logging, etc.
		Tags: []string{"admin"},
	}

	// Register the routes with middleware
	huma.Register(api, footgunOp, addPoolToContext(pool, resetDbFunc))
	return nil
}
