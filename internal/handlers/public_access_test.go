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

// TestPublicAccess tests the public access functionality when "*" is in authorizedReaders
func TestPublicAccess(t *testing.T) {
	// Get the database connection pool from package variable
	pool := connPool

	// Create a mock key generator
	mockKeyGen := new(MockKeyGen)
	// Set up expectations for the mock key generator
	mockKeyGen.On("RandomKey", 32).Return("12345678901234567890123456789012", nil).Maybe()

	// Start the server
	err, shutDownServer := startTestServer(t, pool, mockKeyGen)
	assert.NoError(t, err)

	// Create user bob to be used in tests
	bobJSON := `{"user_handle": "bob", "name": "Bob Smith", "email": "bob@foo.bar"}`
	bobAPIKey, err := createUser(t, bobJSON)
	if err != nil {
		t.Fatalf("Error creating user bob for testing: %v\n", err)
	}

	// Create a public project with "*" in authorizedReaders
	publicProjectJSON := `{"project_handle": "public-test", "description": "A public test project", "authorizedReaders": ["*"]}`
	_, err = createProject(t, publicProjectJSON, "bob", bobAPIKey)
	if err != nil {
		t.Fatalf("Error creating project bob/public-test for testing: %v\n", err)
	}

	// Create API standard to be used in embeddings tests
	apiStandardJSON := `{"api_standard_handle": "openai", "description": "OpenAI Embeddings API", "key_method": "auth_bearer", "key_field": "Authorization" }`
	_, err = createAPIStandard(t, apiStandardJSON, options.AdminKey)
	if err != nil {
		// Ignore error if API standard already exists from previous test
		if err.Error() != "status code 409" {
			t.Logf("Warning: Error creating API standard (may already exist): %v\n", err)
		}
	}

	// Create LLM Service to be used in embeddings tests
	llmServiceJSON := `{ "llm_service_handle": "test1", "endpoint": "https://api.foo.bar/v1/embed", "description": "An LLM Service just for testing if the dhamps-vdb code is working", "api_key": "0123456789", "api_standard": "openai", "model": "embed-test1", "dimensions": 5}`
	_, err = createLLMService(t, llmServiceJSON, "bob", bobAPIKey)
	if err != nil {
		t.Fatalf("Error creating LLM service openai-large for testing: %v\n", err)
	}

	// Post some embeddings to the public project
	_, err = postEmbeddings(t, "../../testdata/valid_embeddings.json", "bob", "public-test", bobAPIKey)
	if err != nil {
		t.Fatalf("Error posting embeddings: %v\n", err)
	}

	fmt.Printf("\nRunning public access tests ...\n\n")

	// Define test cases
	tt := []struct {
		name         string
		method       string
		requestPath  string
		apiKeyHeader string
		expectStatus int
		checkSuccess bool // If true, check for 200/2xx status instead of specific body
	}{
		{
			name:         "Get project embeddings without authentication (public project)",
			method:       http.MethodGet,
			requestPath:  "/v1/embeddings/bob/public-test",
			apiKeyHeader: "",
			expectStatus: http.StatusOK,
			checkSuccess: true,
		},
		{
			name:         "Get document embeddings without authentication (public project)",
			method:       http.MethodGet,
			requestPath:  "/v1/embeddings/bob/public-test/https%3A%2F%2Fid.salamanca.school%2Ftexts%2FW0001%3Avol1.1.1.1.1",
			apiKeyHeader: "",
			expectStatus: http.StatusOK,
			checkSuccess: true,
		},
		{
			name:         "Get similars without authentication (public project)",
			method:       http.MethodGet,
			requestPath:  "/v1/similars/bob/public-test/https%3A%2F%2Fid.salamanca.school%2Ftexts%2FW0001%3Avol1.1.1.1.1",
			apiKeyHeader: "",
			expectStatus: http.StatusOK,
			checkSuccess: true,
		},
		{
			name:         "Get project metadata without authentication (public project)",
			method:       http.MethodGet,
			requestPath:  "/v1/projects/bob/public-test",
			apiKeyHeader: "",
			expectStatus: http.StatusOK,
			checkSuccess: true,
		},
		{
			name:         "Post embeddings without authentication should still be unauthorized (public project)",
			method:       http.MethodPost,
			requestPath:  "/v1/embeddings/bob/public-test",
			apiKeyHeader: "",
			expectStatus: http.StatusUnauthorized,
			checkSuccess: false,
		},
	}

	for _, v := range tt {
		t.Run(v.name, func(t *testing.T) {
			requestURL := fmt.Sprintf("http://%v:%d%v", options.Host, options.Port, v.requestPath)
			req, err := http.NewRequest(v.method, requestURL, nil)
			assert.NoError(t, err)

			if v.apiKeyHeader != "" {
				req.Header.Add("Authorization", "Bearer "+v.apiKeyHeader)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Errorf("Error sending request: %v\n", err)
			}
			assert.NoError(t, err)

			// Check status code
			assert.Equal(t, v.expectStatus, resp.StatusCode, "Status code mismatch for %s", v.name)

			if v.checkSuccess && resp.StatusCode >= 200 && resp.StatusCode < 300 {
				t.Logf("✓ %s: Got successful response with status %d", v.name, resp.StatusCode)
			}

			resp.Body.Close()
		})
	}

	// Cleanup
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
}

// Helper function to post embeddings
func postEmbeddings(t *testing.T, bodyPath, user, project, apiKey string) (string, error) {
	f, err := os.Open(bodyPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	b := new(bytes.Buffer)
	_, err = io.Copy(b, f)
	if err != nil {
		return "", err
	}

	requestURL := fmt.Sprintf("http://%v:%d/v1/embeddings/%s/%s", options.Host, options.Port, user, project)
	req, err := http.NewRequest(http.MethodPost, requestURL, bytes.NewReader(b.Bytes()))
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(bodyBytes), nil
}
