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

func TestSimilarsFunc(t *testing.T) {

	// Get the database connection pool from package variable
	pool := connPool

	// Create a mock key generator
	mockKeyGen := new(MockKeyGen)
	// Set up expectations for the mock key generator
	mockKeyGen.On("RandomKey", 32).Return("12345678901234567890123456789012", nil).Maybe()

	// Start the server
	err, shutDownServer := startTestServer(t, pool, mockKeyGen)
	assert.NoError(t, err)

	// Create user
	aliceJSON := `{"user_handle": "alice", "name": "Alice Doe", "email": "alice@foo.bar"}`
	aliceAPIKey, err := createUser(t, aliceJSON)
	if err != nil {
		t.Fatalf("Error creating user alice for testing: %v\n", err)
	}

	// Create API standard
	apiStandardJSON := `{"api_standard_handle": "openai", "description": "OpenAI Embeddings API", "key_method": "auth_bearer", "key_field": "Authorization" }`
	_, err = createAPIStandard(t, apiStandardJSON, options.AdminKey)
	if err != nil {
		t.Fatalf("Error creating API standard openai for testing: %v\n", err)
	}

	// Create LLM Service
	InstanceJSON := `{ "instance_handle": "embedding1", "endpoint": "https://api.foo.bar/v1/embed", "description": "An LLM Service just for testing if the dhamps-vdb code is working", "api_standard": "openai", "model": "embed-test1", "dimensions": 5}`
	_, err = createInstance(t, InstanceJSON, "alice", aliceAPIKey)
	if err != nil {
		t.Fatalf("Error creating LLM service openai-large for testing: %v\n", err)
	}

	// Create project
	projectJSON := `{"project_handle": "test1", "description": "A test project", "instance_owner": "alice", "instance_handle": "embedding1"}`
	_, err = createProject(t, projectJSON, "alice", aliceAPIKey)
	if err != nil {
		t.Fatalf("Error creating project alice/test1 for testing: %v\n", err)
	}

	// Upload embeddings
	embeddingsFilePath := "../../testdata/valid_embeddings.json"
	embeddingsFile, err := os.Open(embeddingsFilePath)
	if err != nil {
		t.Fatalf("Error opening embeddings file: %v\n", err)
	}
	// Defer closing embeddingsFile
	defer func() {
		if err := embeddingsFile.Close(); err != nil {
			t.Fatalf("Error closing embeddings file: %v\n", err)
		}
	}()
	embeddingsData, err := io.ReadAll(embeddingsFile)
	if err != nil {
		t.Fatalf("Error reading embeddings file: %v\n", err)
	}
	err = createEmbeddings(t, embeddingsData, "alice", "test1", aliceAPIKey)
	if err != nil {
		t.Fatalf("Error creating embeddings for testing: %v\n", err)
	}

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
			name:         "Get similar passages, no parameters",
			method:       http.MethodGet,
			requestPath:  "/v1/similars/alice/test1/https%3A%2F%2Fid.salamanca.school%2Ftexts%2FW0001%3Avol1.1.1.1.1",
			bodyPath:     "",
			apiKey:       aliceAPIKey,
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/SimilarResponseBody.json\",\n  \"user_handle\": \"alice\",\n  \"project_handle\": \"test1\",\n  \"ids\": [\n    \"https%3A%2F%2Fid.salamanca.school%2Ftexts%2FW0001%3Avol1.2\",\n    \"https%3A%2F%2Fid.salamanca.school%2Ftexts%2FW0001%3Avol2\"\n  ]\n}\n",
			expectStatus: http.StatusOK,
		},
		{
			name:         "Get similar passages, with filter",
			method:       http.MethodGet,
			requestPath:  "/v1/similars/alice/test1/https%3A%2F%2Fid.salamanca.school%2Ftexts%2FW0001%3Avol1.1.1.1.1?metadata_path=author&metadata_value=Immanuel%20Kant",
			bodyPath:     "",
			apiKey:       options.AdminKey,
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/SimilarResponseBody.json\",\n  \"user_handle\": \"alice\",\n  \"project_handle\": \"test1\",\n  \"ids\": [\n    \"https%3A%2F%2Fid.salamanca.school%2Ftexts%2FW0001%3Avol2\"\n  ]\n}\n",
			expectStatus: http.StatusOK,
		},
	}

	for _, v := range tt {
		t.Run(v.name, func(t *testing.T) {

			// We need to handle the body only for PUT and POST requests
			// For GET and DELETE requests, the body is nil
			reqBody := io.Reader(nil)
			if v.method == http.MethodGet || v.method == http.MethodDelete {
				reqBody = nil
			} else {
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
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Errorf("Error sending request: %v\n", err)
			}
			// assert.NoError(t, err)
			defer resp.Body.Close()

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
			// if (resp.StatusCode != http.StatusOK) || (resp.StatusCode != int(v.expectStatus)) {
			assert.Equal(t, v.expectBody, formattedResp, "they should be equal")
			// }
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

	fmt.Printf("\n\n\n\n")
}

// TestPostSimilarStub is a placeholder test for the POST similar functionality.
// The POST similar functionality is not yet implemented (returns nil, nil).
// This test documents that the handler exists but is not yet functional.
func TestPostSimilarStub(t *testing.T) {
	t.Skip("POST similar functionality is not yet implemented")

	// TODO: When postSimilarFunc is implemented, add tests for:
	// - Valid POST request with embedding vector
	// - Invalid POST request with malformed data
	// - Finding similar items based on provided vector
	// - Authentication/authorization checks
	// - Edge cases and error handling
}
