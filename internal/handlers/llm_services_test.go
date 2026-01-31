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

func TestLLMServicesFunc(t *testing.T) {
	// Get the database connection pool from package variable
	pool := connPool

	// Create a mock key generator
	mockKeyGen := new(MockKeyGen)
	// Set up expectations for the mock key generator
	mockKeyGen.On("RandomKey", 32).Return("12345678901234567890123456789012", nil).Maybe()

	// Start the server
	err, shutDownServer := startTestServer(t, pool, mockKeyGen)
	assert.NoError(t, err)

	// Create user to be used in llm-service tests
	aliceJSON := `{"user_handle": "alice", "name": "Alice Doe", "email": "alice@foo.bar"}`
	aliceAPIKey, err := createUser(t, aliceJSON)
	if err != nil {
		t.Fatalf("Error creating user alice for testing: %v\n", err)
	}

	// Create API standard to be used in llm-service tests
	openaiJSON := `{"api_standard_handle": "openai", "description": "OpenAI Embeddings API", "key_method": "auth_bearer", "key_field": "Authorization" }`
	_, err = createAPIStandard(t, openaiJSON, options.AdminKey)
	if err != nil {
		t.Fatalf("Error creating API standard openai for testing: %v\n", err)
	}

	fmt.Printf("\nRunning llm-services tests ...\n\n")

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
			name:         "Put llm-service, invalid json",
			method:       http.MethodPut,
			requestPath:  "/v1/llm-services/alice/openai-large",
			bodyPath:     "../../testdata/invalid_llm_service.json",
			apiKey:       aliceAPIKey,
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/ErrorModel.json\",\n  \"title\": \"Unprocessable Entity\",\n  \"status\": 422,\n  \"detail\": \"validation failed\",\n  \"errors\": [\n    {\n      \"message\": \"expected required property model to be present\",\n      \"location\": \"body\",\n      \"value\": {\n        \"api_keX\": \"0123456789\",\n        \"api_standard\": \"openai\",\n        \"description\": \"My OpenAI reduced text-embedding-3-large service\",\n        \"dimensions\": 1024,\n        \"endpoint\": \"https://api.openai.com/v1/embeddings\",\n        \"llm_service_handle\": \"openai-error\"\n      }\n    },\n    {\n      \"message\": \"unexpected property\",\n      \"location\": \"body.api_keX\",\n      \"value\": {\n        \"api_keX\": \"0123456789\",\n        \"api_standard\": \"openai\",\n        \"description\": \"My OpenAI reduced text-embedding-3-large service\",\n        \"dimensions\": 1024,\n        \"endpoint\": \"https://api.openai.com/v1/embeddings\",\n        \"llm_service_handle\": \"openai-error\"\n      }\n    }\n  ]\n}\n",
			expectStatus: http.StatusUnprocessableEntity,
		},
		{
			name:         "Put llm-service, wrong path",
			method:       http.MethodPut,
			requestPath:  "/v1/llm-services/alice/nonexistent",
			bodyPath:     "../../testdata/valid_llm_service_test1.json",
			apiKey:       aliceAPIKey,
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/ErrorModel.json\",\n  \"title\": \"Bad Request\",\n  \"status\": 400,\n  \"detail\": \"llm-service handle in URL (\\\"nonexistent\\\") does not match llm-service handle in body (\\\"test1\\\")\"\n}\n",
			expectStatus: http.StatusBadRequest,
		},
		{
			name:         "Valid put llm-service",
			method:       http.MethodPut,
			requestPath:  "/v1/llm-services/alice/test1",
			bodyPath:     "../../testdata/valid_llm_service_test1.json",
			apiKey:       aliceAPIKey,
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/UploadLLMResponseBody.json\",\n  \"owner\": \"alice\",\n  \"llm_service_handle\": \"test1\",\n  \"llm_service_id\": 1\n}\n",
			expectStatus: http.StatusCreated,
		},
		{
			name:         "Post llm-service, invalid json",
			method:       http.MethodPost,
			requestPath:  "/v1/llm-services/alice",
			bodyPath:     "../../testdata/invalid_llm_service.json",
			apiKey:       aliceAPIKey,
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/ErrorModel.json\",\n  \"title\": \"Unprocessable Entity\",\n  \"status\": 422,\n  \"detail\": \"validation failed\",\n  \"errors\": [\n    {\n      \"message\": \"expected required property model to be present\",\n      \"location\": \"body\",\n      \"value\": {\n        \"api_keX\": \"0123456789\",\n        \"api_standard\": \"openai\",\n        \"description\": \"My OpenAI reduced text-embedding-3-large service\",\n        \"dimensions\": 1024,\n        \"endpoint\": \"https://api.openai.com/v1/embeddings\",\n        \"llm_service_handle\": \"openai-error\"\n      }\n    },\n    {\n      \"message\": \"unexpected property\",\n      \"location\": \"body.api_keX\",\n      \"value\": {\n        \"api_keX\": \"0123456789\",\n        \"api_standard\": \"openai\",\n        \"description\": \"My OpenAI reduced text-embedding-3-large service\",\n        \"dimensions\": 1024,\n        \"endpoint\": \"https://api.openai.com/v1/embeddings\",\n        \"llm_service_handle\": \"openai-error\"\n      }\n    }\n  ]\n}\n",
			expectStatus: http.StatusUnprocessableEntity,
		},
		{
			name:         "Valid post llm-service",
			method:       http.MethodPost,
			requestPath:  "/v1/llm-services/alice",
			bodyPath:     "../../testdata/valid_llm_service_test1.json",
			apiKey:       aliceAPIKey,
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/UploadLLMResponseBody.json\",\n  \"owner\": \"alice\",\n  \"llm_service_handle\": \"test1\",\n  \"llm_service_id\": 1\n}\n",
			expectStatus: http.StatusCreated,
		},
		{
			name:         "Get all llm-services, admin's api key",
			method:       http.MethodGet,
			requestPath:  "/v1/llm-services/alice",
			bodyPath:     "",
			apiKey:       options.AdminKey,
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/GetUserLLMsResponseBody.json\",\n  \"llm_service\": [\n    {\n      \"llm_service_id\": 1,\n      \"llm_service_handle\": \"test1\",\n      \"owner\": \"alice\",\n      \"endpoint\": \"https://api.foo.bar/v1/embed\",\n      \"description\": \"An LLM Service just for testing if the dhamps-vdb code is working\",\n      \"api_key\": \"0123456789\",\n      \"api_standard\": \"openai\",\n      \"model\": \"embed-test1\",\n      \"dimensions\": 5\n    }\n  ]\n}\n",
			expectStatus: http.StatusOK,
		},
		{
			name:         "Get all llm-services, alice's api key",
			method:       http.MethodGet,
			requestPath:  "/v1/llm-services/alice",
			bodyPath:     "",
			apiKey:       aliceAPIKey,
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/GetUserLLMsResponseBody.json\",\n  \"llm_service\": [\n    {\n      \"llm_service_id\": 1,\n      \"llm_service_handle\": \"test1\",\n      \"owner\": \"alice\",\n      \"endpoint\": \"https://api.foo.bar/v1/embed\",\n      \"description\": \"An LLM Service just for testing if the dhamps-vdb code is working\",\n      \"api_key\": \"0123456789\",\n      \"api_standard\": \"openai\",\n      \"model\": \"embed-test1\",\n      \"dimensions\": 5\n    }\n  ]\n}\n",
			expectStatus: http.StatusOK,
		},
		{
			name:         "Get all llm-services, unauthorized",
			method:       http.MethodGet,
			requestPath:  "/v1/llm-services/alice",
			bodyPath:     "",
			apiKey:       "",
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/ErrorModel.json\",\n  \"title\": \"Unauthorized\",\n  \"status\": 401,\n  \"detail\": \"Authentication failed. Perhaps a missing or incorrect API key?\"\n}\n",
			expectStatus: http.StatusUnauthorized,
		},
		{
			name:         "Get all llm-services, nonexistent user",
			method:       http.MethodGet,
			requestPath:  "/v1/llm-services/john",
			bodyPath:     "",
			apiKey:       options.AdminKey,
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/ErrorModel.json\",\n  \"title\": \"Not Found\",\n  \"status\": 404,\n  \"detail\": \"user john not found\"\n}\n",
			expectStatus: http.StatusNotFound,
		},
		{
			name:         "Get nonexistent llm-service",
			method:       http.MethodGet,
			requestPath:  "/v1/llm-services/alice/nonexistent",
			bodyPath:     "",
			apiKey:       aliceAPIKey,
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/ErrorModel.json\",\n  \"title\": \"Not Found\",\n  \"status\": 404,\n  \"detail\": \"llm service nonexistent for user alice not found\"\n}\n",
			expectStatus: http.StatusNotFound,
		},
		{
			name:         "Get single llm-service, nonexistent path",
			method:       http.MethodGet,
			requestPath:  "/v1/llm-services/alice/nonexistant",
			bodyPath:     "",
			apiKey:       aliceAPIKey,
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/ErrorModel.json\",\n  \"title\": \"Not Found\",\n  \"status\": 404,\n  \"detail\": \"llm service nonexistant for user alice not found\"\n}\n",
			expectStatus: http.StatusNotFound,
		},
		{
			name:         "Valid get single llm-service",
			method:       http.MethodGet,
			requestPath:  "/v1/llm-services/alice/test1",
			bodyPath:     "",
			apiKey:       aliceAPIKey,
			expectBody:   "{\n  \"llm_service_id\": 1,\n  \"llm_service_handle\": \"test1\",\n  \"owner\": \"alice\",\n  \"endpoint\": \"https://api.foo.bar/v1/embed\",\n  \"description\": \"An LLM Service just for testing if the dhamps-vdb code is working\",\n  \"api_key\": \"0123456789\",\n  \"api_standard\": \"openai\",\n  \"model\": \"embed-test1\",\n  \"dimensions\": 5\n}\n",
			expectStatus: http.StatusOK,
		},
		{
			name:         "Delete nonexistent llm-service",
			method:       http.MethodDelete,
			requestPath:  "/v1/llm-services/alice/nonexistent",
			bodyPath:     "",
			apiKey:       aliceAPIKey,
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/ErrorModel.json\",\n  \"title\": \"Not Found\",\n  \"status\": 404,\n  \"detail\": \"llm service nonexistent for user alice not found\"\n}\n",
			expectStatus: http.StatusNotFound,
		},
		{
			name:         "Delete llm-service, invalid user",
			method:       http.MethodDelete,
			requestPath:  "/v1/llm-services/john/test1",
			bodyPath:     "",
			apiKey:       options.AdminKey,
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/ErrorModel.json\",\n  \"title\": \"Not Found\",\n  \"status\": 404,\n  \"detail\": \"user john not found\"\n}\n",
			expectStatus: http.StatusNotFound,
		},
		{
			name:         "Delete llm-service, unauthorized",
			method:       http.MethodDelete,
			requestPath:  "/v1/llm-services/alice/test1",
			bodyPath:     "",
			apiKey:       "",
			expectBody:   "{\n  \"$schema\": \"http://localhost:8080/schemas/ErrorModel.json\",\n  \"title\": \"Unauthorized\",\n  \"status\": 401,\n  \"detail\": \"Authentication failed. Perhaps a missing or incorrect API key?\"\n}\n",
			expectStatus: http.StatusUnauthorized,
		},
		{
			name:         "Valid delete llm-service",
			method:       http.MethodDelete,
			requestPath:  "/v1/llm-services/alice/test1",
			bodyPath:     "",
			apiKey:       aliceAPIKey,
			expectBody:   "",
			expectStatus: http.StatusNoContent,
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
