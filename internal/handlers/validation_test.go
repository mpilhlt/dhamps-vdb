package handlers_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidationFunc(t *testing.T) {
	// Get the database connection pool from package variable
	pool := connPool

	// Create a mock key generator
	mockKeyGen := new(MockKeyGen)
	// Set up expectations for the mock key generator
	mockKeyGen.On("RandomKey", 32).Return("12345678901234567890123456789012", nil).Maybe()

	// Start the server
	err, shutDownServer := startTestServer(t, pool, mockKeyGen)
	assert.NoError(t, err)

	// Create user to be used in tests
	aliceJSON := `{"user_handle": "alice", "name": "Alice Doe", "email": "alice@foo.bar"}`
	aliceAPIKey, err := createUser(t, aliceJSON)
	if err != nil {
		t.Fatalf("Error creating user alice for testing: %v\n", err)
	}

	// Create project without schema
	projectJSON := `{"project_handle": "test1", "description": "A test project"}`
	_, err = createProject(t, projectJSON, "alice", aliceAPIKey)
	if err != nil {
		t.Fatalf("Error creating project alice/test1 for testing: %v\n", err)
	}

	// Create project with metadata schema
	projectWithSchemaJSON := `{"project_handle": "test-schema", "description": "Test project with schema", "metadataScheme": "{\"type\":\"object\",\"properties\":{\"author\":{\"type\":\"string\"},\"year\":{\"type\":\"integer\"}},\"required\":[\"author\"]}"}`
	_, err = createProject(t, projectWithSchemaJSON, "alice", aliceAPIKey)
	if err != nil {
		t.Fatalf("Error creating project alice/test-schema for testing: %v\n", err)
	}

	// Create API standard to be used in embeddings tests
	apiStandardJSON := `{"api_standard_handle": "openai", "description": "OpenAI Embeddings API", "key_method": "auth_bearer", "key_field": "Authorization" }`
	_, err = createAPIStandard(t, apiStandardJSON, options.AdminKey)
	if err != nil {
		t.Fatalf("Error creating API standard openai for testing: %v\n", err)
	}

	// Create LLM Service with 5 dimensions for testing
	llmServiceJSON := `{ "llm_service_handle": "openai-large", "endpoint": "https://api.openai.com/v1/embeddings", "description": "My OpenAI test service", "api_key": "0123456789", "api_standard": "openai", "model": "text-embedding-3-large", "dimensions": 5}`
	_, err = createLLMService(t, llmServiceJSON, "alice", aliceAPIKey)
	if err != nil {
		t.Fatalf("Error creating LLM service openai-large for testing: %v\n", err)
	}

	fmt.Printf("\nRunning validation tests ...\n\n")

	// Define test cases
	tt := []struct {
		name         string
		method       string
		requestPath  string
		bodyPath     string
		apiKeyHeader string
		expectBody   string
		expectStatus int16
	}{
		{
			name:         "Post embeddings with wrong vector length",
			method:       http.MethodPost,
			requestPath:  "/v1/embeddings/alice/test1",
			bodyPath:     "../../testdata/invalid_embeddings_wrong_dims.json",
			apiKeyHeader: aliceAPIKey,
			expectBody:   "dimension validation failed: vector length mismatch",
			expectStatus: http.StatusBadRequest,
		},
		{
			name:         "Post embeddings with dimension mismatch to LLM service",
			method:       http.MethodPost,
			requestPath:  "/v1/embeddings/alice/test1",
			bodyPath:     "../../testdata/invalid_embeddings_dimension_mismatch.json",
			apiKeyHeader: aliceAPIKey,
			expectBody:   "dimension validation failed: vector dimension mismatch",
			expectStatus: http.StatusBadRequest,
		},
		{
			name:         "Post embeddings with valid metadata against schema",
			method:       http.MethodPost,
			requestPath:  "/v1/embeddings/alice/test-schema",
			bodyPath:     "../../testdata/valid_embeddings_with_schema.json",
			apiKeyHeader: aliceAPIKey,
			expectBody:   "test-valid-metadata",
			expectStatus: http.StatusCreated,
		},
		{
			name:         "Post embeddings with invalid metadata against schema",
			method:       http.MethodPost,
			requestPath:  "/v1/embeddings/alice/test-schema",
			bodyPath:     "../../testdata/invalid_embeddings_schema_violation.json",
			apiKeyHeader: aliceAPIKey,
			expectBody:   "metadata validation failed",
			expectStatus: http.StatusBadRequest,
		},
	}

	// Run the tests
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// Read the body from file if specified
			var bodyReader io.Reader
			if tc.bodyPath != "" {
				body, err := os.ReadFile(tc.bodyPath)
				if err != nil {
					t.Fatalf("Error reading body file: %v", err)
				}
				bodyReader = bytes.NewReader(body)
			}

			// Create the request
			req, err := http.NewRequest(tc.method, fmt.Sprintf("http://localhost:8080%s", tc.requestPath), bodyReader)
			assert.NoError(t, err)
			if tc.apiKeyHeader != "" {
				req.Header.Set("Authorization", "Bearer "+tc.apiKeyHeader)
			}
			req.Header.Set("Content-Type", "application/json")

			// Send the request
			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)
			defer resp.Body.Close()

			// Read the response body
			respBody, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)

			// Check the status code
			assert.Equal(t, int(tc.expectStatus), resp.StatusCode, "Status code mismatch for test: %s. Response body: %s", tc.name, string(respBody))

			// Check the response body contains expected text
			if tc.expectBody != "" {
				assert.Contains(t, string(respBody), tc.expectBody, "Response body mismatch for test: %s", tc.name)
			}
		})
	}

	// Shutdown the server
	shutDownServer()
}
