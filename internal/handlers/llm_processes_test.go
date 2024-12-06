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

func TestLLMProcessesFunc(t *testing.T) {
	t.Skip("Skip LLM Process testing...")
	// Get the database connection pool from package variable
	pool := connPool

	// Create a mock key generator
	mockKeyGen := new(MockKeyGen)
	// Set up expectations for the mock key generator
	mockKeyGen.On("RandomKey", 32).Return("12345678901234567890123456789012", nil)

	// Start the server
	err, shutDownServer := startTestServer(t, pool, mockKeyGen)
	assert.NoError(t, err)

	// Create user to be used in project tests
	aliceJSON := `{"user_handle": "alice", "name": "Alice Doe", "email": "alice@foo.bar"}`
	fmt.Print("    Creating user (alice) for testing ...\n")
	aliceAPIKey, err := createUser(t, aliceJSON)
	if err != nil {
		t.Fatalf("Error creating user alice for testing: %v\n", err)
	}
	fmt.Printf("    User (alice) created with API key: %s\n", aliceAPIKey)
	fmt.Printf("    Running project tests ...\n")

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
			name:         "Valid get all projects, no projects present, admin's api key",
			method:       http.MethodGet,
			requestPath:  "/v1/projects/alice",
			bodyPath:     "",
			apiKey:       options.AdminKey,
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/GetProjectsResponseBody.json\",\n  \"projects\": []\n}\n",
			expectStatus: 200,
		},
		{
			name:         "Valid get all projects, no projects present, alice's api key",
			method:       http.MethodGet,
			requestPath:  "/v1/projects/alice",
			bodyPath:     "",
			apiKey:       aliceAPIKey,
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/GetProjectsResponseBody.json\",\n  \"projects\": []\n}\n",
			expectStatus: 200,
		},
		{
			name:         "Put project, valid json",
			method:       http.MethodPut,
			requestPath:  "/v1/projects/alice/test1",
			bodyPath:     "../../testdata/valid_project.json",
			apiKey:       aliceAPIKey,
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/UploadProjectResponseBody.json\",\n  \"project_handle\": \"test1\",\n  \"project_id\": 0\n}\n",
			expectStatus: 201,
		},
		{
			name:         "Put project, invalid json",
			method:       http.MethodPut,
			requestPath:  "/v1/projects/alice/test2",
			bodyPath:     "../../testdata/invalid_project.json",
			apiKey:       aliceAPIKey,
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/ErrorModel.json\",\n  \"title\": \"Unprocessable Entity\",\n  \"status\": 422,\n  \"detail\": \"validation failed\",\n  \"errors\": [\n    {\n      \"message\": \"expected required property project_handle to be present\",\n      \"location\": \"body\",\n      \"value\": {\n        \"description\": \"This is a test project\",\n        \"foo\": \"test1\"\n      }\n    },\n    {\n      \"message\": \"unexpected property\",\n      \"location\": \"body.foo\",\n      \"value\": {\n        \"description\": \"This is a test project\",\n        \"foo\": \"test1\"\n      }\n    }\n  ]\n}\n",
			expectStatus: 422,
		},
		{
			name:         "Put project, valid json but invalid project handle",
			method:       http.MethodPut,
			requestPath:  "/v1/projects/alice/test3",
			bodyPath:     "../../testdata/valid_project.json",
			apiKey:       aliceAPIKey,
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/ErrorModel.json\",\n  \"title\": \"Bad Request\",\n  \"status\": 400,\n  \"detail\": \"project handle in URL (test3) does not match project handle in body (test1)\"\n}\n",
			expectStatus: 400,
		},
		{
			name:         "Post project, valid json",
			method:       http.MethodPost,
			requestPath:  "/v1/projects/alice",
			bodyPath:     "../../testdata/valid_project.json",
			apiKey:       aliceAPIKey,
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/UploadProjectResponseBody.json\",\n  \"project_handle\": \"test1\",\n  \"project_id\": 0\n}\n",
			expectStatus: 201,
		},
		{
			name:         "Post project, invalid json",
			method:       http.MethodPost,
			requestPath:  "/v1/projects/alice",
			bodyPath:     "../../testdata/invalid_project.json",
			apiKey:       aliceAPIKey,
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/ErrorModel.json\",\n  \"title\": \"Unprocessable Entity\",\n  \"status\": 422,\n  \"detail\": \"validation failed\",\n  \"errors\": [\n    {\n      \"message\": \"expected required property project_handle to be present\",\n      \"location\": \"body\",\n      \"value\": {\n        \"description\": \"This is a test project\",\n        \"foo\": \"test1\"\n      }\n    },\n    {\n      \"message\": \"unexpected property\",\n      \"location\": \"body.foo\",\n      \"value\": {\n        \"description\": \"This is a test project\",\n        \"foo\": \"test1\"\n      }\n    }\n  ]\n}\n",
			expectStatus: 422,
		},
		{
			name:         "Valid get project",
			method:       http.MethodGet,
			requestPath:  "/v1/projects/alice/test1",
			bodyPath:     "",
			apiKey:       aliceAPIKey,
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/GetProjectResponseBody.json\",\n  \"project\": {\n    \"project_id\": 0,\n    \"project_handle\": \"test1\",\n    \"description\": \"This is a test project\",\n    \"authorizedReaders\": [\n      \"alice\"\n    ]\n  }\n}\n",
			expectStatus: 200,
		},
		{
			name:         "Valid get all projects",
			method:       http.MethodGet,
			requestPath:  "/v1/projects/alice",
			bodyPath:     "",
			apiKey:       aliceAPIKey,
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/GetProjectsResponseBody.json\",\n  \"projects\": [\n    {\n      \"project_id\": 1,\n      \"project_handle\": \"test1\",\n      \"description\": \"This is a test project\",\n      \"authorizedReaders\": [\n        \"alice\"\n      ]\n    }\n  ]\n}\n",
			expectStatus: 200,
		},
		{
			name:         "Valid get all projects, invalid user",
			method:       http.MethodGet,
			requestPath:  "/v1/projects/john",
			bodyPath:     "",
			apiKey:       aliceAPIKey,
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/ErrorModel.json\",\n  \"title\": \"Unauthorized\",\n  \"status\": 401,\n  \"detail\": \"Authentication failed. Perhaps a missing or incorrect API key?\"\n}\n",
			expectStatus: 401,
		},
		{
			name:         "Get nonexistent project",
			method:       http.MethodGet,
			requestPath:  "/v1/projects/alice/test2",
			bodyPath:     "",
			apiKey:       aliceAPIKey,
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/ErrorModel.json\",\n  \"title\": \"Not Found\",\n  \"status\": 404,\n  \"detail\": \"user alice's project test2 not found\"\n}\n",
			expectStatus: 404,
		},
		{
			name:         "Delete project",
			method:       http.MethodDelete,
			requestPath:  "/v1/projects/alice/test1",
			bodyPath:     "",
			apiKey:       aliceAPIKey,
			expectBody:   "",
			expectStatus: 204,
		},
		{
			name:         "Delete nonexistent project",
			method:       http.MethodDelete,
			requestPath:  "/v1/projects/alice/test2",
			bodyPath:     "",
			apiKey:       aliceAPIKey,
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/ErrorModel.json\",\n  \"title\": \"Not Found\",\n  \"status\": 404,\n  \"detail\": \"project test2 not found for user alice\"\n}\n",
			expectStatus: 404,
		},
		{
			name:         "Delete project, invalid user",
			method:       http.MethodDelete,
			requestPath:  "/v1/projects/john/test1",
			bodyPath:     "",
			apiKey:       aliceAPIKey,
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/ErrorModel.json\",\n  \"title\": \"Unauthorized\",\n  \"status\": 401,\n  \"detail\": \"Authentication failed. Perhaps a missing or incorrect API key?\"\n}\n",
			expectStatus: 401,
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
	// (deleting '/users/alice' should delete all the projects connected to alice as well)
	t.Cleanup(func() {
		tt := []struct {
			name        string
			requestPath string
		}{
			{
				name:        "clean up alice",
				requestPath: "/v1/users/alice",
			},
		}

		for _, v := range tt {
			fmt.Printf("Running cleanup: %s\n", v.name)
			requestURL := fmt.Sprintf("http://%s:%d%s", options.Host, options.Port, v.requestPath)
			req, err := http.NewRequest(http.MethodDelete, requestURL, nil)
			assert.NoError(t, err)
			req.Header.Set("Authorization", "Bearer "+options.AdminKey)
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
