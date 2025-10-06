package parser

import (
	"testing"
)

func TestSQLViewQueryParser_Parse(t *testing.T) {
	parser := NewSQLViewQueryParser()

	tests := []struct {
		name        string
		query       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid query with all required columns",
			query:       "SELECT uuid, provider, provider_alias, resource_type, namespace, name, attributes FROM resources",
			expectError: false,
		},
		{
			name:        "valid query with different column order",
			query:       "SELECT name, uuid, provider, resource_type, provider_alias, namespace, attributes FROM resources",
			expectError: false,
		},
		{
			name:        "invalid query with SELECT *",
			query:       "SELECT * FROM resources",
			expectError: true,
			errorMsg:    "SELECT * is not allowed",
		},
		{
			name:        "invalid query missing uuid column",
			query:       "SELECT provider, provider_alias, resource_type, namespace, name, attributes FROM resources",
			expectError: true,
			errorMsg:    "missing required column: uuid",
		},
		{
			name:        "invalid query with extra column",
			query:       "SELECT uuid, provider, provider_alias, resource_type, namespace, name, attributes, extra_column FROM resources",
			expectError: true,
			errorMsg:    "query must return exactly 7 columns",
		},
		{
			name:        "invalid SQL syntax",
			query:       "INVALID SQL QUERY",
			expectError: true,
			errorMsg:    "invalid SQL query",
		},
		{
			name:        "non-SELECT statement",
			query:       "INSERT INTO resources VALUES (1, 'test')",
			expectError: true,
			errorMsg:    "query must be a SELECT statement",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.query)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if result != tt.query {
					t.Errorf("expected result to be the same as input query")
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
