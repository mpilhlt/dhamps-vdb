package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmbeddingsFunc(t *testing.T) {
	// Get the database connection pool from package variable
	pool := connPool

	// Create a mock key generator
	mockKeyGen := new(MockKeyGen)
	// Set up expectations for the mock key generator
	mockKeyGen.On("RandomKey", 64).Return("12345678901234567890123456789012", nil)

	// Start the server
	err, shutDownServer := startTestServer(t, pool, mockKeyGen)
	assert.NoError(t, err)

	// Create user to be used in project tests
	aliceJSON := `{"handle": "alice", "name": "Alice Doe", "email": "alice@foo.bar"}`
	fmt.Print("    Creating user (alice) for testing ...\n")
	aliceAPIKey, err := createUser(t, aliceJSON)
	assert.NoError(t, err)

	// Create project to be used in embedding tests
	projectJSON := `{"handle": "test1", "description": "A test project"}`
	fmt.Print("    Creating project (test1) for testing ...\n")
	projectID, err := createProject(t, projectJSON, "alice", aliceAPIKey)
	assert.NoError(t, err)
	fmt.Printf("    Project ID: %s\n", strconv.Itoa(projectID))

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
			name:         "Valid get all embeddings",
			method:       http.MethodGet,
			requestPath:  "/embeddings/alice/test1",
			bodyPath:     "",
			apiKeyHeader: aliceAPIKey,
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/ErrorModel.json\",\n  \"title\": \"Not Found\",\n  \"status\": 404,\n  \"detail\": \"no embeddings found.\"\n}\n",
			expectStatus: 404,
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
			req.Header.Add("Authorization", "Bearer "+v.apiKeyHeader)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Errorf("Error sending request: %v\n", err)
			}
			assert.NoError(t, err)
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
				if respBody == nil {
					t.Errorf("Expected body %s, got nil\n", v.expectBody)
				} else {
					fr := new(bytes.Buffer)
					if isJSON(string(respBody)) && (strings.Contains(string(respBody), "{") || strings.Contains(string(respBody), "[")) {
						err = json.Indent(fr, respBody, "", "  ")
						// fmt.Printf("Error: %v\nresponse: %v\n", err, string(respBody))
						assert.NoError(t, err)
						formattedResp = fr.String()
					} else {
						formattedResp = string(respBody)
					}
				}
			}
			// if (resp.StatusCode != http.StatusOK) || (resp.StatusCode != int(v.expectStatus)) {
			assert.Equal(t, v.expectBody, formattedResp, "they should be equal")
			// }
		})
	}

	// Verify that the expectations regarding the mock key generation were met
	mockKeyGen.AssertExpectations(t)

	// Cleanup removes items created by the put function test
	// (deleting '/users/alice' should delete all the projects and embeddings connected
	//  to alice as well)
	t.Cleanup(func() {
		tt := []struct {
			name        string
			requestPath string
		}{
			{
				name:        "clean up alice",
				requestPath: "/users/alice",
			},
		}

		for _, v := range tt {
			fmt.Printf("Running cleanup: %s\n", v.name)
			requestURL := fmt.Sprintf("http://%s:%d%s", options.Host, options.Port, v.requestPath)
			req, err := http.NewRequest(http.MethodDelete, requestURL, nil)
			assert.NoError(t, err)
			_, err = http.DefaultClient.Do(req)
			if err != nil && err.Error() != "no rows in result set" {
				t.Fatalf("Error sending request: %v\n", err)
			}
			assert.NoError(t, err)
		}
		fmt.Print("Shutting down server\n")
		shutDownServer()
	})

}
