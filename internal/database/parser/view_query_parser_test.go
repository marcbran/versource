package parser

import (
	"testing"
)

func TestSQLViewQueryParser_Parse(t *testing.T) {
	parser := NewSQLViewQueryParser()

	tests := []struct {
		name          string
		query         string
		expectedName  string
		expectedQuery string
		expectedError string
	}{
		{
			name:          "valid three-column query and computed fields present",
			query:         "SELECT 'aws' AS provider, 'ec2_instance' AS resource_type, 'web' AS name FROM resources WHERE provider = 'aws' AND resource_type = 'ec2_instance'",
			expectedName:  "aws_ec2_instance_aws_ec2_instance_web",
			expectedQuery: "SELECT 'aws' AS provider, 'ec2_instance' AS resource_type, 'web' AS name FROM resources WHERE provider = 'aws' AND resource_type = 'ec2_instance'",
			expectedError: "",
		},
		{
			name:          "invalid query with SELECT *",
			query:         "SELECT * FROM resources WHERE provider = 'aws' AND resource_type = 'ec2_instance'",
			expectedName:  "",
			expectedQuery: "",
			expectedError: "SELECT * is not allowed",
		},
		{
			name:          "invalid query missing provider column",
			query:         "SELECT 'ec2_instance' AS resource_type, 'web' AS name FROM resources WHERE provider = 'aws' AND resource_type = 'ec2_instance'",
			expectedName:  "",
			expectedQuery: "",
			expectedError: "missing required column: provider",
		},
		{
			name:          "invalid query with extra column",
			query:         "SELECT 'aws' AS provider, 'ec2_instance' AS resource_type, 'web' AS name, namespace FROM resources WHERE provider = 'aws' AND resource_type = 'ec2_instance'",
			expectedName:  "",
			expectedQuery: "",
			expectedError: "query must return exactly 3 columns",
		},
		{
			name:          "invalid provider not static string",
			query:         "SELECT provider, 'ec2_instance' AS resource_type, 'web' AS name FROM resources WHERE provider = 'aws' AND resource_type = 'ec2_instance'",
			expectedName:  "",
			expectedQuery: "",
			expectedError: "provider must be a static string literal",
		},
		{
			name:          "invalid resource_type not static string",
			query:         "SELECT 'aws' AS provider, resource_type, 'web' AS name FROM resources WHERE provider = 'aws' AND resource_type = 'ec2_instance'",
			expectedName:  "",
			expectedQuery: "",
			expectedError: "resource_type must be a static string literal",
		},
		{
			name:          "invalid name not static string",
			query:         "SELECT 'aws' AS provider, 'ec2_instance' AS resource_type, name FROM resources WHERE provider = 'aws' AND resource_type = 'ec2_instance'",
			expectedName:  "",
			expectedQuery: "",
			expectedError: "name must be a static string literal",
		},
		{
			name:          "invalid SQL syntax",
			query:         "INVALID SQL QUERY",
			expectedName:  "",
			expectedQuery: "",
			expectedError: "invalid SQL query",
		},
		{
			name:          "non-SELECT statement",
			query:         "INSERT INTO resources VALUES (1, 'test')",
			expectedName:  "",
			expectedQuery: "",
			expectedError: "query must be a SELECT statement",
		},
		{
			name:          "invalid query not using FROM resources",
			query:         "SELECT 'aws' AS provider, 'ec2_instance' AS resource_type, 'web' AS name FROM other_table",
			expectedName:  "",
			expectedQuery: "",
			expectedError: "query must use FROM resources",
		},
		{
			name:          "invalid query missing WHERE clause",
			query:         "SELECT 'aws' AS provider, 'ec2_instance' AS resource_type, 'web' AS name FROM resources",
			expectedName:  "",
			expectedQuery: "",
			expectedError: "query must have a WHERE clause",
		},
		{
			name:          "invalid query missing provider in WHERE clause",
			query:         "SELECT 'aws' AS provider, 'ec2_instance' AS resource_type, 'web' AS name FROM resources WHERE resource_type = 'ec2_instance'",
			expectedName:  "",
			expectedQuery: "",
			expectedError: "WHERE clause must include provider = 'string'",
		},
		{
			name:          "invalid query missing resource_type in WHERE clause",
			query:         "SELECT 'aws' AS provider, 'ec2_instance' AS resource_type, 'web' AS name FROM resources WHERE provider = 'aws'",
			expectedName:  "",
			expectedQuery: "",
			expectedError: "WHERE clause must include resource_type = 'string'",
		},
		{
			name:          "invalid query with non-string provider in WHERE clause",
			query:         "SELECT 'aws' AS provider, 'ec2_instance' AS resource_type, 'web' AS name FROM resources WHERE provider = 123 AND resource_type = 'ec2_instance'",
			expectedName:  "",
			expectedQuery: "",
			expectedError: "provider must be a static string literal in WHERE clause",
		},
		{
			name:          "invalid query with non-string resource_type in WHERE clause",
			query:         "SELECT 'aws' AS provider, 'ec2_instance' AS resource_type, 'web' AS name FROM resources WHERE provider = 'aws' AND resource_type = 456",
			expectedName:  "",
			expectedQuery: "",
			expectedError: "resource_type must be a static string literal in WHERE clause",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.query)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if !contains(err.Error(), tt.expectedError) {
					t.Errorf("expected error to contain '%s', got '%s'", tt.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result.Name != tt.expectedName {
				t.Errorf("expected name %q, got %q", tt.expectedName, result.Name)
			}
			if result.Query != tt.expectedQuery {
				t.Errorf("expected query %q, got %q", tt.expectedQuery, result.Query)
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
