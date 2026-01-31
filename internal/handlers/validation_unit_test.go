package handlers

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/mpilhlt/dhamps-vdb/internal/models"
)

func TestValidateEmbeddingDimensions(t *testing.T) {
	tests := []struct {
		name          string
		embedding     models.EmbeddingsInput
		llmDimensions int32
		wantErr       bool
		errContains   string
	}{
		{
			name: "Valid embedding",
			embedding: models.EmbeddingsInput{
				TextID:           "test-id",
				LLMServiceHandle: "test-llm",
				Vector:           []float32{1.0, 2.0, 3.0},
				VectorDim:        3,
			},
			llmDimensions: 3,
			wantErr:       false,
		},
		{
			name: "Empty text_id",
			embedding: models.EmbeddingsInput{
				TextID:           "",
				LLMServiceHandle: "test-llm",
				Vector:           []float32{1.0, 2.0, 3.0},
				VectorDim:        3,
			},
			llmDimensions: 3,
			wantErr:       true,
			errContains:   "text_id cannot be empty",
		},
		{
			name: "Empty vector",
			embedding: models.EmbeddingsInput{
				TextID:           "test-id",
				LLMServiceHandle: "test-llm",
				Vector:           []float32{},
				VectorDim:        3,
			},
			llmDimensions: 3,
			wantErr:       true,
			errContains:   "vector cannot be empty",
		},
		{
			name: "Dimension mismatch with LLM service",
			embedding: models.EmbeddingsInput{
				TextID:           "test-id",
				LLMServiceHandle: "test-llm",
				Vector:           []float32{1.0, 2.0, 3.0},
				VectorDim:        5,
			},
			llmDimensions: 5,
			wantErr:       true,
			errContains:   "vector length mismatch",
		},
		{
			name: "Vector dim doesn't match LLM service",
			embedding: models.EmbeddingsInput{
				TextID:           "test-id",
				LLMServiceHandle: "test-llm",
				Vector:           []float32{1.0, 2.0, 3.0},
				VectorDim:        3,
			},
			llmDimensions: 5,
			wantErr:       true,
			errContains:   "vector dimension mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmbeddingDimensions(tt.embedding, tt.llmDimensions)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmbeddingDimensions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ValidateEmbeddingDimensions() error = %v, should contain %v", err.Error(), tt.errContains)
				}
			}
		})
	}
}

func TestValidateMetadataAgainstSchema(t *testing.T) {
	tests := []struct {
		name        string
		metadata    json.RawMessage
		schemaStr   string
		wantErr     bool
		errContains string
	}{
		{
			name:      "No schema provided",
			metadata:  json.RawMessage(`{"author": "John Doe"}`),
			schemaStr: "",
			wantErr:   false,
		},
		{
			name:      "No metadata provided",
			metadata:  json.RawMessage(``),
			schemaStr: `{"type":"object","properties":{"author":{"type":"string"}},"required":["author"]}`,
			wantErr:   false,
		},
		{
			name:      "Valid metadata",
			metadata:  json.RawMessage(`{"author": "John Doe", "year": 2021}`),
			schemaStr: `{"type":"object","properties":{"author":{"type":"string"},"year":{"type":"integer"}},"required":["author"]}`,
			wantErr:   false,
		},
		{
			name:        "Missing required field",
			metadata:    json.RawMessage(`{"year": 2021}`),
			schemaStr:   `{"type":"object","properties":{"author":{"type":"string"},"year":{"type":"integer"}},"required":["author"]}`,
			wantErr:     true,
			errContains: "author",
		},
		{
			name:        "Wrong type",
			metadata:    json.RawMessage(`{"author": "John Doe", "year": "2021"}`),
			schemaStr:   `{"type":"object","properties":{"author":{"type":"string"},"year":{"type":"integer"}},"required":["author"]}`,
			wantErr:     true,
			errContains: "year",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMetadataAgainstSchema(tt.metadata, tt.schemaStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMetadataAgainstSchema() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ValidateMetadataAgainstSchema() error = %v, should contain %v", err.Error(), tt.errContains)
				}
			}
		})
	}
}
