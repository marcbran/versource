package platform

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

func TestMatchPathPrefix(t *testing.T) {
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
			name:       "prefix match",
			routePath:  "modules",
			actualPath: "modules/123",
			want:       map[string]string{},
		},
		{
			name:       "longer prefix match",
			routePath:  "modules",
			actualPath: "modules/123/edit",
			want:       map[string]string{},
		},
		{
			name:       "no match different parts",
			routePath:  "modules",
			actualPath: "components",
			want:       nil,
		},
		{
			name:       "parameter prefix match",
			routePath:  "modules/{moduleID}",
			actualPath: "modules/123/edit",
			want:       map[string]string{"moduleID": "123"},
		},
		{
			name:       "multiple parameters prefix match",
			routePath:  "modules/{moduleID}/edit",
			actualPath: "modules/123/edit/v1",
			want:       map[string]string{"moduleID": "123"},
		},
		{
			name:       "parameter with text prefix match",
			routePath:  "modules/{moduleID}/edit",
			actualPath: "modules/123/edit",
			want:       map[string]string{"moduleID": "123"},
		},
		{
			name:       "no match with parameter",
			routePath:  "modules/{moduleID}",
			actualPath: "components/123",
			want:       nil,
		},
		{
			name:       "empty paths",
			routePath:  "",
			actualPath: "",
			want:       map[string]string{},
		},
		{
			name:       "root path prefix",
			routePath:  "",
			actualPath: "modules",
			want:       map[string]string{},
		},
		{
			name:       "registered path longer than current",
			routePath:  "modules/123/edit",
			actualPath: "modules/123",
			want:       nil,
		},
		{
			name:       "partial match not prefix",
			routePath:  "modules/abc",
			actualPath: "modules/123",
			want:       nil,
		},
		{
			name:       "query parameters with prefix match",
			routePath:  "modules/{moduleID}",
			actualPath: "modules/123/edit?version=1.0",
			want:       map[string]string{"moduleID": "123", "version": "1.0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchPathPrefix(tt.routePath, tt.actualPath)
			if !mapsEqual(got, tt.want) {
				t.Errorf("matchPathPrefix() = %v (len=%d), want %v (len=%d)", got, len(got), tt.want, len(tt.want))
			}
		})
	}
}

func TestFindAllMatchingKeyBindings(t *testing.T) {
	tests := []struct {
		name        string
		keyBindings map[string]KeyBindingsFunc
		currentPath string
		expected    KeyBindings
	}{
		{
			name: "exact match",
			keyBindings: map[string]KeyBindingsFunc{
				"modules": func(params map[string]string) KeyBindings {
					return KeyBindings{}.With("m", "Go to modules", "/modules")
				},
			},
			currentPath: "modules",
			expected:    KeyBindings{}.With("m", "Go to modules", "/modules"),
		},
		{
			name: "prefix match for modules/123",
			keyBindings: map[string]KeyBindingsFunc{
				"modules": func(params map[string]string) KeyBindings {
					return KeyBindings{}.With("m", "Go to modules", "/modules")
				},
				"modules/{moduleID}": func(params map[string]string) KeyBindings {
					return KeyBindings{}.With("e", "Edit module", "/modules/{moduleID}/edit")
				},
				"modules/{moduleID}/edit": func(params map[string]string) KeyBindings {
					return KeyBindings{}.With("s", "Save module", "/modules/{moduleID}")
				},
			},
			currentPath: "modules/123",
			expected:    KeyBindings{}.With("e", "Edit module", "/modules/{moduleID}/edit").With("m", "Go to modules", "/modules"),
		},
		{
			name: "longest prefix match for modules/123/edit",
			keyBindings: map[string]KeyBindingsFunc{
				"modules": func(params map[string]string) KeyBindings {
					return KeyBindings{}.With("m", "Go to modules", "/modules")
				},
				"modules/{moduleID}": func(params map[string]string) KeyBindings {
					return KeyBindings{}.With("e", "Edit module", "/modules/{moduleID}/edit")
				},
				"modules/{moduleID}/edit": func(params map[string]string) KeyBindings {
					return KeyBindings{}.With("s", "Save module", "/modules/{moduleID}")
				},
			},
			currentPath: "modules/123/edit",
			expected:    KeyBindings{}.With("s", "Save module", "/modules/{moduleID}").With("e", "Edit module", "/modules/{moduleID}/edit").With("m", "Go to modules", "/modules"),
		},
		{
			name: "no match",
			keyBindings: map[string]KeyBindingsFunc{
				"modules": func(params map[string]string) KeyBindings {
					return KeyBindings{}.With("m", "Go to modules", "/modules")
				},
			},
			currentPath: "other",
			expected:    KeyBindings{},
		},
		{
			name:        "empty key bindings",
			keyBindings: map[string]KeyBindingsFunc{},
			currentPath: "modules",
			expected:    KeyBindings{},
		},
		{
			name: "multiple prefix matches with different lengths",
			keyBindings: map[string]KeyBindingsFunc{
				"a":       func(params map[string]string) KeyBindings { return KeyBindings{}.With("1", "Action 1", "/a") },
				"a/b":     func(params map[string]string) KeyBindings { return KeyBindings{}.With("2", "Action 2", "/a/b") },
				"a/b/c":   func(params map[string]string) KeyBindings { return KeyBindings{}.With("3", "Action 3", "/a/b/c") },
				"a/b/c/d": func(params map[string]string) KeyBindings { return KeyBindings{}.With("4", "Action 4", "/a/b/c/d") },
			},
			currentPath: "a/b/c/d/e",
			expected:    KeyBindings{}.With("4", "Action 4", "/a/b/c/d").With("3", "Action 3", "/a/b/c").With("2", "Action 2", "/a/b").With("1", "Action 1", "/a"),
		},
		{
			name: "root path should be selected when no other matches",
			keyBindings: map[string]KeyBindingsFunc{
				"": func(params map[string]string) KeyBindings { return KeyBindings{}.With("r", "Root action", "/") },
				"modules": func(params map[string]string) KeyBindings {
					return KeyBindings{}.With("m", "Go to modules", "/modules")
				},
			},
			currentPath: "other",
			expected:    KeyBindings{}.With("r", "Root action", "/"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findAllMatchingKeyBindings(tt.keyBindings, tt.currentPath)
			if !keyBindingsEqual(got, tt.expected) {
				t.Errorf("findAllMatchingKeyBindings() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func mapsEqual(a, b map[string]string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
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

func keyBindingsEqual(a, b KeyBindings) bool {
	if len(a) != len(b) {
		return false
	}
	for i, bindingA := range a {
		bindingB := b[i]
		if bindingA.Key != bindingB.Key || bindingA.Help != bindingB.Help || bindingA.Command != bindingB.Command {
			return false
		}
	}
	return true
}
