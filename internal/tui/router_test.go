package tui

import (
	"testing"
)

func TestMatchPath(t *testing.T) {
	tests := []struct {
		name       string
		routePath  string
		actualPath string
		want       map[string]string
	}{
		{
			name:       "exact match",
			routePath:  "modules",
			actualPath: "modules",
			want:       map[string]string{},
		},
		{
			name:       "no match different length",
			routePath:  "modules",
			actualPath: "modules/versions",
			want:       nil,
		},
		{
			name:       "no match different parts",
			routePath:  "modules",
			actualPath: "changesets",
			want:       nil,
		},
		{
			name:       "single parameter",
			routePath:  "modules/{moduleID}",
			actualPath: "modules/123",
			want:       map[string]string{"moduleID": "123"},
		},
		{
			name:       "multiple parameters",
			routePath:  "modules/{moduleID}/moduleversions",
			actualPath: "modules/456/moduleversions",
			want:       map[string]string{"moduleID": "456"},
		},
		{
			name:       "parameter with text",
			routePath:  "modules/{moduleID}/versions",
			actualPath: "modules/789/versions",
			want:       map[string]string{"moduleID": "789"},
		},
		{
			name:       "no match with parameter",
			routePath:  "modules/{moduleID}/versions",
			actualPath: "modules/789/other",
			want:       nil,
		},
		{
			name:       "empty paths",
			routePath:  "",
			actualPath: "",
			want:       map[string]string{},
		},
		{
			name:       "root path",
			routePath:  "/",
			actualPath: "/",
			want:       map[string]string{},
		},
		{
			name:       "query parameters",
			routePath:  "components/{componentID}",
			actualPath: "components/123?module-id=456",
			want:       map[string]string{"componentID": "123", "module-id": "456"},
		},
		{
			name:       "multiple query parameters",
			routePath:  "components/{componentID}",
			actualPath: "components/123?module-id=456&status=active",
			want:       map[string]string{"componentID": "123", "module-id": "456", "status": "active"},
		},
		{
			name:       "path with query parameter only",
			routePath:  "components",
			actualPath: "components?module-id=456",
			want:       map[string]string{"module-id": "456"},
		},
		{
			name:       "empty query string",
			routePath:  "components/{componentID}",
			actualPath: "components/123?",
			want:       map[string]string{"componentID": "123"},
		},
		{
			name:       "query with empty value",
			routePath:  "components/{componentID}",
			actualPath: "components/123?module-id=456&empty=",
			want:       map[string]string{"componentID": "123", "module-id": "456", "empty": ""},
		},
		{
			name:       "complex query parameters",
			routePath:  "modules/{moduleID}/moduleversions",
			actualPath: "modules/789/moduleversions?status=active&limit=10&offset=0",
			want:       map[string]string{"moduleID": "789", "status": "active", "limit": "10", "offset": "0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchPath(tt.routePath, tt.actualPath)
			if !mapsEqual(got, tt.want) {
				t.Errorf("matchPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func mapsEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}
