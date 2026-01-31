package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/mpilhlt/dhamps-vdb/internal/models"
	"github.com/xeipuuv/gojsonschema"
)

// ValidateEmbeddingDimensions checks if the embeddings vector dimensions match the LLM service dimensions
func ValidateEmbeddingDimensions(embedding models.EmbeddingsInput, llmDimensions int32) error {
	// Check if text_id is not empty
	if embedding.TextID == "" {
		return fmt.Errorf("text_id cannot be empty")
	}

	// Check if vector is not empty
	if len(embedding.Vector) == 0 {
		return fmt.Errorf("vector cannot be empty for text_id '%s'", embedding.TextID)
	}

	// Check if declared vector_dim matches LLM service dimensions
	if embedding.VectorDim != llmDimensions {
		return fmt.Errorf("vector dimension mismatch: embedding declares %d dimensions but LLM service '%s' expects %d dimensions", 
			embedding.VectorDim, embedding.LLMServiceHandle, llmDimensions)
	}

	// Check if actual vector length matches declared vector_dim
	actualLength := int32(len(embedding.Vector))
	if actualLength != embedding.VectorDim {
		return fmt.Errorf("vector length mismatch for text_id '%s': actual vector has %d elements but vector_dim declares %d", 
			embedding.TextID, actualLength, embedding.VectorDim)
	}

	return nil
}

// ValidateMetadataAgainstSchema validates the metadata against a JSON schema if provided
func ValidateMetadataAgainstSchema(metadata json.RawMessage, schemaStr string) error {
	// If no schema is provided, skip validation
	if schemaStr == "" {
		return nil
	}

	// If no metadata is provided but schema exists, that's okay (metadata is optional)
	if len(metadata) == 0 || string(metadata) == "null" {
		return nil
	}

	// Parse the schema
	schemaLoader := gojsonschema.NewStringLoader(schemaStr)
	
	// Parse the metadata
	metadataLoader := gojsonschema.NewBytesLoader(metadata)

	// Validate
	result, err := gojsonschema.Validate(schemaLoader, metadataLoader)
	if err != nil {
		return fmt.Errorf("failed to validate metadata against schema: %v", err)
	}

	if !result.Valid() {
		// Build a helpful error message with all validation errors
		var errMsg string
		errMsg = "metadata validation failed:\n"
		for i, desc := range result.Errors() {
			if i > 0 {
				errMsg += "\n"
			}
			errMsg += fmt.Sprintf("  - %s", desc.String())
		}
		return fmt.Errorf("%s", errMsg)
	}

	return nil
}
