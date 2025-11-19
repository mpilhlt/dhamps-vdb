package models

import (
	"testing"
)

// TestGetIDsEmbeddingssInput tests the GetIDs method for EmbeddingssInput
func TestGetIDsEmbeddingssInput(t *testing.T) {
	tests := []struct {
		name     string
		list     EmbeddingssInput
		expected []string
	}{
		{
			name:     "Empty list",
			list:     EmbeddingssInput{},
			expected: []string{},
		},
		{
			name: "Single item",
			list: EmbeddingssInput{
				{TextID: "id1"},
			},
			expected: []string{"id1"},
		},
		{
			name: "Multiple items",
			list: EmbeddingssInput{
				{TextID: "id1"},
				{TextID: "id2"},
				{TextID: "id3"},
			},
			expected: []string{"id1", "id2", "id3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.list.GetIDs()
			if len(result) != len(tt.expected) {
				t.Errorf("GetIDs() returned %d items, want %d", len(result), len(tt.expected))
			}
			for i, id := range result {
				if id != tt.expected[i] {
					t.Errorf("GetIDs()[%d] = %v, want %v", i, id, tt.expected[i])
				}
			}
		})
	}
}

// TestGetIDsEmbeddingss tests the GetIDs method for Embeddingss
func TestGetIDsEmbeddingss(t *testing.T) {
	tests := []struct {
		name     string
		list     Embeddingss
		expected []string
	}{
		{
			name:     "Empty list",
			list:     Embeddingss{},
			expected: []string{},
		},
		{
			name: "Single item",
			list: Embeddingss{
				{TextID: "id1"},
			},
			expected: []string{"id1"},
		},
		{
			name: "Multiple items",
			list: Embeddingss{
				{TextID: "id1"},
				{TextID: "id2"},
				{TextID: "id3"},
			},
			expected: []string{"id1", "id2", "id3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.list.GetIDs()
			if len(result) != len(tt.expected) {
				t.Errorf("GetIDs() returned %d items, want %d", len(result), len(tt.expected))
			}
			for i, id := range result {
				if id != tt.expected[i] {
					t.Errorf("GetIDs()[%d] = %v, want %v", i, id, tt.expected[i])
				}
			}
		})
	}
}

// Note: Most of the models package consists of struct definitions used for API requests
// and responses. These are validated through the handler integration tests which use
// these models for all API operations. The validation includes:
// - JSON marshalling/unmarshalling
// - Schema validation through Huma framework
// - Data integrity checks
// - Required fields and constraints
