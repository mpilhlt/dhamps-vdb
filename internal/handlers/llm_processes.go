package handlers

import (
	"context"
	"net/http"

	"github.com/mpilhlt/dhamps-vdb/internal/models"

	"github.com/danielgtaylor/huma/v2"
)

// Define handler functions for each route
func postLLMProcessFunc(ctx context.Context, input *models.LLMProcessRequest) (*models.LLMProcessResponse, error) {
	// Implement your logic here
	return nil, nil
}

// RegisterLLMProcessRoutes registers the routes for the LLMProcess service
func RegisterLLMProcessRoutes(api huma.API) {
	// Define huma.Operations for each route
	postLLMProcessOp := huma.Operation{
		OperationID: "postLLMProcess",
		Method:      http.MethodPost,
		Path:        "/v1/llm-process/{user_handle}",
		Summary:     "Process text with LLM service",
		Security: []map[string][]string{
			{"adminAuth": []string{"admin"}},
			{"ownerAuth": []string{"owner"}},
		},
		Tags: []string{"llm-process"},
	}

	huma.Register(api, postLLMProcessOp, postLLMProcessFunc)
}
