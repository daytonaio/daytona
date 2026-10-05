package workspace

import (
	"reflect"
	"testing"
)

func TestFilterList(t *testing.T) {
	workspaces := []string{"dev-env-1", "prod-cluster", "test-sandbox", "dev-backend"}

	tests := []struct {
		name     string
		query    string
		expected []string
	}{
		{
			name:     "empty query returns all",
			query:    "",
			expected: workspaces,
		},
		{
			name:     "filter matching dev",
			query:    "dev",
			expected: []string{"dev-env-1", "dev-backend"},
		},
		{
			name:     "case insensitive match",
			query:    "PROD",
			expected: []string{"prod-cluster"},
		},
		{
			name:     "no match",
			query:    "nonexistent",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterList(workspaces, tt.query)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("FilterList(%v, %q) = %v; want %v", workspaces, tt.query, got, tt.expected)
			}
		})
	}
}

func TestSelectWorkspace(t *testing.T) {
	workspaces := []string{"alpha", "beta", "gamma"}
	selected, err := SelectWorkspace(workspaces, "bet")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(selected) != 1 || selected[0] != "beta" {
		t.Errorf("expected ['beta'], got %v", selected)
	}

	_, err = SelectWorkspace(workspaces, "omega")
	if err == nil {
		t.Errorf("expected error for non-matching query")
	}
}
