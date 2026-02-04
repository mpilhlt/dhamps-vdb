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

// TODO: Test against actual JSON body

// TestPublicAccess tests the public access functionality when "*" is in authorizedReaders
func TestPublicAccess(t *testing.T) {

	fmt.Printf("\n\n\n\n")

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
	InstanceJSON := `{ "instance_handle": "embedding1", "endpoint": "https://api.foo.bar/v1/embed", "description": "An LLM Service just for testing if the dhamps-vdb code is working", "api_standard": "openai", "model": "embed-test1", "dimensions": 5}`
	_, err = createInstance(t, InstanceJSON, "bob", bobAPIKey)
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
		bodyPath     string
		VDBKey       string
		expectBody   string
		expectStatus int16
	}{
		{
			name:         "Get project embeddings without authentication (public project)",
			method:       http.MethodGet,
			requestPath:  "/v1/embeddings/bob/public-test",
			bodyPath:     "",
			VDBKey:       "",
			expectBody:   "",
			expectStatus: http.StatusOK,
		},
		{
			name:         "Get document embeddings without authentication (public project)",
			method:       http.MethodGet,
			requestPath:  "/v1/embeddings/bob/public-test/https%3A%2F%2Fid.salamanca.school%2Ftexts%2FW0001%3Avol1.1.1.1.1",
			bodyPath:     "",
			VDBKey:       "",
			expectBody:   "",
			expectStatus: http.StatusOK,
		},
		{
			name:         "Get similars without authentication (public project)",
			method:       http.MethodGet,
			requestPath:  "/v1/similars/bob/public-test/https%3A%2F%2Fid.salamanca.school%2Ftexts%2FW0001%3Avol1.1.1.1.1",
			bodyPath:     "",
			VDBKey:       "",
			expectBody:   "",
			expectStatus: http.StatusOK,
		},
		{
			name:         "Get project metadata without authentication (public project)",
			method:       http.MethodGet,
			requestPath:  "/v1/projects/bob/public-test",
			bodyPath:     "",
			VDBKey:       "",
			expectBody:   "",
			expectStatus: http.StatusOK,
		},
		{
			name:         "Post embeddings without authentication (public project)",
			method:       http.MethodPost,
			requestPath:  "/v1/embeddings/bob/public-test",
			bodyPath:     "../../testdata/invalid_embeddings.json",
			VDBKey:       "",
			expectBody:   "",
			expectStatus: http.StatusUnauthorized,
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
			req.Header.Set("Authorization", "Bearer "+v.VDBKey)
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
