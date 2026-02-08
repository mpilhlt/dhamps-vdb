package handlers_test

import (
	"bytes"
	"encoding/json"
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

	// Create API standard to be used in embeddings tests
	apiStandardJSON := `{"api_standard_handle": "openai", "description": "OpenAI Embeddings API", "key_method": "auth_bearer", "key_field": "Authorization" }`
	_, err = createAPIStandard(t, apiStandardJSON, options.AdminKey)
	if err != nil {
		t.Fatalf("Error creating API standard openai for testing: %v\n", err)
	}

	// Create LLM Service Instance with 5 dimensions for testing
	InstanceInstanceJSON := `{ "instance_handle": "embedding1", "endpoint": "https://api.openai.com/v1/embeddings", "description": "My OpenAI test service", "api_standard": "openai", "model": "text-embedding-3-large", "dimensions": 5}`
	_, err = createInstance(t, InstanceInstanceJSON, "alice", aliceAPIKey)
	if err != nil {
		t.Fatalf("Error creating LLM Service Instance embedding1 for testing: %v\n", err)
	}

	// Create project without schema
	projectJSON := `{"project_handle": "test1", "description": "A test project", "instance_owner": "alice", "instance_handle": "embedding1"}`
	_, err = createProject(t, projectJSON, "alice", aliceAPIKey)
	if err != nil {
		t.Fatalf("Error creating project alice/test1 for testing: %v\n", err)
	}

	// Create project with metadata schema
	projectWithSchemaJSON := `{"project_handle": "test-schema", "description": "Test project with schema", "metadataScheme": "{\"type\":\"object\",\"properties\":{\"author\":{\"type\":\"string\"},\"year\":{\"type\":\"integer\"}},\"required\":[\"author\"]}", "instance_owner": "alice", "instance_handle": "embedding1"}`
	_, err = createProject(t, projectWithSchemaJSON, "alice", aliceAPIKey)
	if err != nil {
		t.Fatalf("Error creating project alice/test-schema for testing: %v\n", err)
	}

	fmt.Printf("\nRunning validation tests ...\n\n")

	// Define test cases
	tt := []struct {
		name         string
		method       string
		requestPath  string
		bodyPath     string
		apiKey       string
		expectBody   string
		expectStatus int16
	}{
		{
			name:         "Post embeddings with wrong vector length",
			method:       http.MethodPost,
			requestPath:  "/v1/embeddings/alice/test1",
			bodyPath:     "../../testdata/invalid_embeddings_wrong_dims.json",
			apiKey:       aliceAPIKey,
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/ErrorModel.json\",\n  \"title\": \"Bad Request\",\n  \"status\": 400,\n  \"detail\": \"Dimension validation failed for input test-wrong-dims: vector length mismatch for text_id 'test-wrong-dims': actual vector has 3 elements but vector_dim declares 5\"\n}\n",
			expectStatus: http.StatusBadRequest,
		},
		{
			name:         "Post embeddings with dimension mismatch to LLM service",
			method:       http.MethodPost,
			requestPath:  "/v1/embeddings/alice/test1",
			bodyPath:     "../../testdata/invalid_embeddings_dimension_mismatch.json",
			apiKey:       aliceAPIKey,
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/ErrorModel.json\",\n  \"title\": \"Bad Request\",\n  \"status\": 400,\n  \"detail\": \"Dimension validation failed for input test-mismatch-dims: vector dimension mismatch: embedding declares 3072 dimensions but LLM service instance 'embedding1' expects 5 dimensions\"\n}\n",
			expectStatus: http.StatusBadRequest,
		},
		{
			name:         "Post embeddings with valid metadata against schema",
			method:       http.MethodPost,
			requestPath:  "/v1/embeddings/alice/test-schema",
			bodyPath:     "../../testdata/valid_embeddings_with_schema.json",
			apiKey:       aliceAPIKey,
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/UploadProjEmbeddingsResponseBody.json\",\n  \"ids\": [\n    \"test-valid-metadata\"\n  ]\n}\n",
			expectStatus: http.StatusCreated,
		},
		{
			name:         "Post embeddings with invalid metadata against schema",
			method:       http.MethodPost,
			requestPath:  "/v1/embeddings/alice/test-schema",
			bodyPath:     "../../testdata/invalid_embeddings_schema_violation.json",
			apiKey:       aliceAPIKey,
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/ErrorModel.json\",\n  \"title\": \"Bad Request\",\n  \"status\": 400,\n  \"detail\": \"metadata validation failed for text_id 'test-invalid-metadata': metadata validation failed:\\n  - (root): author is required\"\n}\n",
			expectStatus: http.StatusBadRequest,
		},
	}

	for _, v := range tt {
		t.Run(v.name, func(t *testing.T) {

			// We need to handle the body only for POST requests
			reqBody := io.Reader(nil)
			if v.method == http.MethodPost {
				f, err := os.Open(v.bodyPath)
				assert.NoError(t, err)
				defer func() {
					if err := f.Close(); err != nil {
						t.Fatal(err)
					}
				}()
				b := new(bytes.Buffer)
				_, err = io.Copy(b, f)
				assert.NoError(t, err)
				reqBody = bytes.NewReader(b.Bytes())
			}
			requestURL := fmt.Sprintf("http://%v:%d%v", options.Host, options.Port, v.requestPath)
			req, err := http.NewRequest(v.method, requestURL, reqBody)
			assert.NoError(t, err)
			req.Header.Set("Authorization", "Bearer "+v.apiKey)
			req.Header.Set("Content-Type", "application/json")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Errorf("Error sending request: %v\n", err)
			}
			defer resp.Body.Close()
			assert.NoError(t, err)

			if resp.StatusCode != int(v.expectStatus) {
				t.Errorf("Expected status code %d, got %s\n", v.expectStatus, resp.Status)
			} else {
				t.Logf("Expected status code %d, got %s\n", v.expectStatus, resp.Status)
			}

			respBody, err := io.ReadAll(resp.Body) // response body is []byte
			assert.NoError(t, err)
			formattedResp := ""
			if v.expectBody != "" {
				fr := new(bytes.Buffer)
				err = json.Indent(fr, respBody, "", "  ")
				assert.NoError(t, err)
				formattedResp = fr.String()
			}
			assert.Equal(t, v.expectBody, formattedResp, "they should be equal")
		})
	}

	// Verify that the expectations regarding the mock key generation were met
	mockKeyGen.AssertExpectations(t)

	// Cleanup removes items created by the put function test
	// (deleting '/users/alice' should delete all the
	//  projects, instances and embeddings connected to alice as well)
	t.Cleanup(func() {
		fmt.Print("\n\nRunning cleanup ...\n\n")

		requestURL := fmt.Sprintf("http://%s:%d/v1/admin/footgun", options.Host, options.Port)
		req, err := http.NewRequest(http.MethodGet, requestURL, nil)
		assert.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+options.AdminKey)
		_, err = http.DefaultClient.Do(req)
		if err != nil && err.Error() != "no rows in result set" {
			t.Fatalf("Error sending request: %v\n", err)
		}
		assert.NoError(t, err)

		fmt.Print("Shutting down server\n\n")
		shutDownServer()
	})

	fmt.Printf("\n")
}
