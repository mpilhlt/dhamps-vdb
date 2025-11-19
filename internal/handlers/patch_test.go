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

func TestPatchProjects(t *testing.T) {
	// Get the database connection pool from package variable
	pool := connPool

	// Create a mock key generator
	mockKeyGen := new(MockKeyGen)
	// Set up expectations for the mock key generator (alice and bob)
	mockKeyGen.On("RandomKey", 32).Return("12345678901234567890123456789012", nil).Once()
	mockKeyGen.On("RandomKey", 32).Return("23456789012345678901234567890123", nil).Once()

	// Start the server
	err, shutDownServer := startTestServer(t, pool, mockKeyGen)
	assert.NoError(t, err)

	// Create user to be used in project tests
	aliceJSON := `{"user_handle": "alice", "name": "Alice Doe", "email": "alice@foo.bar"}`
	aliceAPIKey, err := createUser(t, aliceJSON)
	if err != nil {
		t.Fatalf("Error creating user alice for testing: %v\n", err)
	}

	// Create bob user manually since createUser is hardcoded for alice
	bobJSON := `{"user_handle": "bob", "name": "Bob Smith", "email": "bob@foo.bar"}`
	requestURL := fmt.Sprintf("http://%s:%d/v1/users/bob", options.Host, options.Port)
	req, err := http.NewRequest(http.MethodPut, requestURL, bytes.NewBufferString(bobJSON))
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+options.AdminKey)
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Failed to create bob user, status: %d", resp.StatusCode)
	}
	resp.Body.Close()

	fmt.Printf("\nRunning PATCH tests for projects ...\n\n")

	// First, create a project to test PATCH on
	t.Run("Setup: Create project for PATCH testing", func(t *testing.T) {
		projectJSON := `{"project_handle": "patch_test", "description": "Initial description"}`
		requestURL := fmt.Sprintf("http://%v:%d/v1/projects/alice", options.Host, options.Port)
		req, err := http.NewRequest(http.MethodPost, requestURL, bytes.NewBufferString(projectJSON))
		assert.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+aliceAPIKey)
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	// Test PATCH to update description only
	t.Run("PATCH project description only", func(t *testing.T) {
		patchJSON := `{"description": "Updated description via PATCH"}`
		requestURL := fmt.Sprintf("http://%v:%d/v1/projects/alice/patch_test", options.Host, options.Port)
		req, err := http.NewRequest(http.MethodPatch, requestURL, bytes.NewBufferString(patchJSON))
		assert.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+aliceAPIKey)
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		// PATCH should succeed with 201 Created (since it calls PUT internally)
		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected status code %d, got %d. Body: %s", http.StatusCreated, resp.StatusCode, string(respBody))
		}
	})

	// Verify the PATCH update was applied
	t.Run("Verify PATCH updated description", func(t *testing.T) {
		requestURL := fmt.Sprintf("http://%v:%d/v1/projects/alice/patch_test", options.Host, options.Port)
		req, err := http.NewRequest(http.MethodGet, requestURL, nil)
		assert.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+aliceAPIKey)
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		var project map[string]interface{}
		err = json.Unmarshal(respBody, &project)
		assert.NoError(t, err)

		description, ok := project["description"].(string)
		assert.True(t, ok, "description field should be a string")
		assert.Equal(t, "Updated description via PATCH", description, "description should be updated")

		// Verify project_handle is still the same
		projectHandle, ok := project["project_handle"].(string)
		assert.True(t, ok, "project_handle field should be a string")
		assert.Equal(t, "patch_test", projectHandle, "project_handle should remain unchanged")
	})

	// Test PATCH to enable world-readable access
	t.Run("PATCH project to enable world-readable", func(t *testing.T) {
		patchJSON := `{"authorizedReaders": ["*"]}`
		requestURL := fmt.Sprintf("http://%v:%d/v1/projects/alice/patch_test", options.Host, options.Port)
		req, err := http.NewRequest(http.MethodPatch, requestURL, bytes.NewBufferString(patchJSON))
		assert.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+aliceAPIKey)
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected status code %d, got %d. Body: %s", http.StatusCreated, resp.StatusCode, string(respBody))
		}
	})

	// Verify world-readable access was enabled
	t.Run("Verify world-readable access enabled", func(t *testing.T) {
		requestURL := fmt.Sprintf("http://%v:%d/v1/projects/alice/patch_test", options.Host, options.Port)
		req, err := http.NewRequest(http.MethodGet, requestURL, nil)
		assert.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+aliceAPIKey)
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		var project map[string]interface{}
		err = json.Unmarshal(respBody, &project)
		assert.NoError(t, err)

		authorizedReaders, ok := project["authorizedReaders"].([]interface{})
		assert.True(t, ok, "authorizedReaders field should be an array")
		assert.Equal(t, 1, len(authorizedReaders), "should have one authorized reader")
		assert.Equal(t, "*", authorizedReaders[0], "authorized reader should be '*'")
	})

	// Test PATCH to add specific authorized readers
	t.Run("PATCH project to add specific authorized readers", func(t *testing.T) {
		patchJSON := `{"authorizedReaders": ["bob"]}`
		requestURL := fmt.Sprintf("http://%v:%d/v1/projects/alice/patch_test", options.Host, options.Port)
		req, err := http.NewRequest(http.MethodPatch, requestURL, bytes.NewBufferString(patchJSON))
		assert.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+aliceAPIKey)
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected status code %d, got %d. Body: %s", http.StatusCreated, resp.StatusCode, string(respBody))
		}
	})

	// Verify authorized readers updated
	t.Run("Verify authorized readers updated", func(t *testing.T) {
		requestURL := fmt.Sprintf("http://%v:%d/v1/projects/alice/patch_test", options.Host, options.Port)
		req, err := http.NewRequest(http.MethodGet, requestURL, nil)
		assert.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+aliceAPIKey)
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		var project map[string]interface{}
		err = json.Unmarshal(respBody, &project)
		assert.NoError(t, err)

		authorizedReaders, ok := project["authorizedReaders"].([]interface{})
		assert.True(t, ok, "authorizedReaders field should be an array")
		// Should contain alice (owner) and bob
		assert.GreaterOrEqual(t, len(authorizedReaders), 1, "should have at least one authorized reader")
		
		// Check if bob is in the list
		foundBob := false
		for _, reader := range authorizedReaders {
			if reader == "bob" {
				foundBob = true
				break
			}
		}
		assert.True(t, foundBob, "bob should be in authorized readers")
	})

	// Clean up
	fmt.Print("\n\nRunning cleanup ...\n\n")

	cleanupURL := fmt.Sprintf("http://%s:%d/v1/admin/footgun", options.Host, options.Port)
	cleanupReq, cleanupErr := http.NewRequest(http.MethodGet, cleanupURL, nil)
	assert.NoError(t, cleanupErr)
	cleanupReq.Header.Set("Authorization", "Bearer "+options.AdminKey)
	_, cleanupErr = http.DefaultClient.Do(cleanupReq)
	if cleanupErr != nil && cleanupErr.Error() != "no rows in result set" {
		t.Fatalf("Error sending request: %v\n", cleanupErr)
	}
	assert.NoError(t, cleanupErr)

	fmt.Print("Shutting down server\n\n")
	shutDownServer()
}
