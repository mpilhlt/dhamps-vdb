package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInstanceSharingFunc(t *testing.T) {

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

	// Create users to be used in sharing tests
	aliceJSON := `{"user_handle": "alice", "name": "Alice Doe", "email": "alice@foo.bar"}`
	aliceAPIKey, err := createUser(t, aliceJSON)
	if err != nil {
		t.Fatalf("Error creating user alice for testing: %v\n", err)
	}

	bobJSON := `{"user_handle": "bob", "name": "Bob Smith", "email": "bob@foo.bar"}`
	bobAPIKey, err := createUser(t, bobJSON)
	if err != nil {
		t.Fatalf("Error creating user bob for testing: %v\n", err)
	}

	// Create API standard to be used in tests
	openaiJSON := `{"api_standard_handle": "openai", "description": "OpenAI Embeddings API", "key_method": "auth_bearer", "key_field": "Authorization" }`
	_, err = createAPIStandard(t, openaiJSON, options.AdminKey)
	if err != nil {
		t.Fatalf("Error creating API standard openai for testing: %v\n", err)
	}

	// Create an instance for alice
	instanceJSON := `{"instance_handle": "my-openai", "endpoint": "https://api.openai.com/v1/embeddings", "description": "Alice's OpenAI instance", "api_standard": "openai", "model": "text-embedding-3-large", "dimensions": 3072}`
	_, err = createInstance(t, "alice", instanceJSON, aliceAPIKey)
	if err != nil {
		t.Fatalf("Error creating instance for sharing tests: %v\n", err)
	}

	fmt.Printf("\nRunning llm-instances sharing tests ...\n\n")

	// Define test cases
	tt := []struct {
		name         string
		method       string
		requestPath  string
		bodyJSON     string
		VDBKey       string
		expectBody   string
		expectStatus int16
	}{
		{
			name:         "Share instance with bob - valid",
			method:       http.MethodPost,
			requestPath:  "/v1/llm-instances/alice/my-openai/share",
			bodyJSON:     `{"user_handle": "bob", "role": "reader"}`,
			VDBKey:       aliceAPIKey,
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/ShareInstanceResponseBody.json\",\n  \"owner\": \"alice\",\n  \"instance_handle\": \"my-openai\",\n  \"shared_with\": \"bob\",\n  \"role\": \"reader\"\n}\n",
			expectStatus: http.StatusCreated,
		},
		{
			name:         "Share instance with nonexistent user - should fail",
			method:       http.MethodPost,
			requestPath:  "/v1/llm-instances/alice/my-openai/share",
			bodyJSON:     `{"user_handle": "charlie", "role": "reader"}`,
			VDBKey:       aliceAPIKey,
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/ErrorModel.json\",\n  \"title\": \"Not Found\",\n  \"status\": 404,\n  \"detail\": \"user charlie not found\"\n}\n",
			expectStatus: http.StatusNotFound,
		},
		{
			name:         "Share nonexistent instance - should fail",
			method:       http.MethodPost,
			requestPath:  "/v1/llm-instances/alice/nonexistent/share",
			bodyJSON:     `{"user_handle": "bob", "role": "reader"}`,
			VDBKey:       aliceAPIKey,
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/ErrorModel.json\",\n  \"title\": \"Not Found\",\n  \"status\": 404,\n  \"detail\": \"instance alice/nonexistent not found\"\n}\n",
			expectStatus: http.StatusNotFound,
		},
		{
			name:         "Bob cannot share alice's instance - should fail",
			method:       http.MethodPost,
			requestPath:  "/v1/llm-instances/alice/my-openai/share",
			bodyJSON:     `{"user_handle": "alice", "role": "editor"}`,
			VDBKey:       bobAPIKey,
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/ErrorModel.json\",\n  \"title\": \"Forbidden\",\n  \"status\": 403,\n  \"detail\": \"you are not authorized to perform this action\"\n}\n",
			expectStatus: http.StatusForbidden,
		},
		{
			name:         "Get shared users for instance",
			method:       http.MethodGet,
			requestPath:  "/v1/llm-instances/alice/my-openai/shared-with",
			bodyJSON:     "",
			VDBKey:       aliceAPIKey,
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/GetInstanceSharedUsersResponseBody.json\",\n  \"shared_with\": [\n    {\n      \"user_handle\": \"bob\",\n      \"role\": \"reader\"\n    }\n  ]\n}\n",
			expectStatus: http.StatusOK,
		},
		{
			name:         "Unshare instance from bob",
			method:       http.MethodDelete,
			requestPath:  "/v1/llm-instances/alice/my-openai/share/bob",
			bodyJSON:     "",
			VDBKey:       aliceAPIKey,
			expectBody:   "",
			expectStatus: http.StatusNoContent,
		},
		{
			name:         "Get shared users after unsharing - should be empty",
			method:       http.MethodGet,
			requestPath:  "/v1/llm-instances/alice/my-openai/shared-with",
			bodyJSON:     "",
			VDBKey:       aliceAPIKey,
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/GetInstanceSharedUsersResponseBody.json\",\n  \"shared_with\": []\n}\n",
			expectStatus: http.StatusOK,
		},
	}

	for _, v := range tt {
		t.Run(v.name, func(t *testing.T) {
			requestURL := fmt.Sprintf("http://%s:%d%s", options.Host, options.Port, v.requestPath)

			var req *http.Request
			if v.bodyJSON != "" {
				req, err = http.NewRequest(v.method, requestURL, bytes.NewBuffer([]byte(v.bodyJSON)))
			} else {
				req, err = http.NewRequest(v.method, requestURL, nil)
			}
			assert.NoError(t, err)

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+v.VDBKey)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Errorf("Error sending request: %v\n", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != int(v.expectStatus) {
				t.Errorf("Expected status code %d, got %s\n", v.expectStatus, resp.Status)
			} else {
				t.Logf("Expected status code %d, got %s\n", v.expectStatus, resp.Status)
			}

			respBody, err := io.ReadAll(resp.Body)
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

	// Cleanup removes items created by the tests
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
}

// Note: createInstance helper function is defined in handlers_test.go

func TestDefinitionSharingFunc(t *testing.T) {

	fmt.Printf("\n\n\n\n")

	// Get the database connection pool from package variable
	pool := connPool

	// Create a mock key generator
	mockKeyGen := new(MockKeyGen)
	mockKeyGen.On("RandomKey", 32).Return("12345678901234567890123456789012", nil).Maybe()

	// Start the server
	err, shutDownServer := startTestServer(t, pool, mockKeyGen)
	assert.NoError(t, err)

	// Create users
	aliceJSON := `{"user_handle": "alice", "name": "Alice Doe", "email": "alice@foo.bar"}`
	aliceAPIKey, err := createUser(t, aliceJSON)
	if err != nil {
		t.Fatalf("Error creating user alice for testing: %v\n", err)
	}

	bobJSON := `{"user_handle": "bob", "name": "Bob Smith", "email": "bob@foo.bar"}`
	_, err = createUser(t, bobJSON)
	if err != nil {
		t.Fatalf("Error creating user bob for testing: %v\n", err)
	}

	// Create API standard
	openaiJSON := `{"api_standard_handle": "openai", "description": "OpenAI Embeddings API", "key_method": "auth_bearer", "key_field": "Authorization" }`
	_, err = createAPIStandard(t, openaiJSON, options.AdminKey)
	if err != nil {
		t.Fatalf("Error creating API standard openai for testing: %v\n", err)
	}

	// Note: _system user and definitions are created by migration 004
	// We can test sharing _system definitions and alice's own definitions

	fmt.Printf("\nRunning llm-definitions sharing tests ...\n\n")

	// Define test cases
	tt := []struct {
		name         string
		method       string
		requestPath  string
		bodyJSON     string
		VDBKey       string
		expectBody   string
		expectStatus int16
	}{
		{
			name:         "Admin shares _system definition with alice",
			method:       http.MethodPost,
			requestPath:  "/v1/llm-definitions/_system/openai-large/share",
			bodyJSON:     `{"user_handle": "alice", "role": "reader"}`,
			VDBKey:       options.AdminKey,
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/ShareDefinitionResponseBody.json\",\n  \"owner\": \"_system\",\n  \"definition_handle\": \"openai-large\",\n  \"shared_with\": \"alice\",\n  \"role\": \"reader\"\n}\n",
			expectStatus: http.StatusCreated,
		},
		{
			name:         "Alice cannot share _system definition - should fail",
			method:       http.MethodPost,
			requestPath:  "/v1/llm-definitions/_system/openai-large/share",
			bodyJSON:     `{"user_handle": "bob", "role": "reader"}`,
			VDBKey:       aliceAPIKey,
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/ErrorModel.json\",\n  \"title\": \"Forbidden\",\n  \"status\": 403,\n  \"detail\": \"you are not authorized to perform this action\"\n}\n",
			expectStatus: http.StatusForbidden,
		},
		{
			name:         "Get shared users for _system definition",
			method:       http.MethodGet,
			requestPath:  "/v1/llm-definitions/_system/openai-large/shared-with",
			bodyJSON:     "",
			VDBKey:       options.AdminKey,
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/GetDefinitionSharedUsersResponseBody.json\",\n  \"shared_with\": [\n    {\n      \"user_handle\": \"alice\",\n      \"role\": \"reader\"\n    }\n  ]\n}\n",
			expectStatus: http.StatusOK,
		},
		{
			name:         "Admin unshares _system definition from alice",
			method:       http.MethodDelete,
			requestPath:  "/v1/llm-definitions/_system/openai-large/share/alice",
			bodyJSON:     "",
			VDBKey:       options.AdminKey,
			expectBody:   "",
			expectStatus: http.StatusNoContent,
		},
		{
			name:         "Share with nonexistent user - should fail",
			method:       http.MethodPost,
			requestPath:  "/v1/llm-definitions/_system/openai-large/share",
			bodyJSON:     `{"user_handle": "charlie", "role": "reader"}`,
			VDBKey:       options.AdminKey,
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/ErrorModel.json\",\n  \"title\": \"Not Found\",\n  \"status\": 404,\n  \"detail\": \"user charlie not found\"\n}\n",
			expectStatus: http.StatusNotFound,
		},
	}

	for _, v := range tt {
		t.Run(v.name, func(t *testing.T) {
			requestURL := fmt.Sprintf("http://%s:%d%s", options.Host, options.Port, v.requestPath)

			var req *http.Request
			if v.bodyJSON != "" {
				req, err = http.NewRequest(v.method, requestURL, bytes.NewBuffer([]byte(v.bodyJSON)))
			} else {
				req, err = http.NewRequest(v.method, requestURL, nil)
			}
			assert.NoError(t, err)

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+v.VDBKey)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Errorf("Error sending request: %v\n", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != int(v.expectStatus) {
				t.Errorf("Expected status code %d, got %s\n", v.expectStatus, resp.Status)
			} else {
				t.Logf("Expected status code %d, got %s\n", v.expectStatus, resp.Status)
			}

			respBody, err := io.ReadAll(resp.Body)
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

	// Verify mock expectations
	mockKeyGen.AssertExpectations(t)

	// Cleanup
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
}
