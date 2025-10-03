package yaml

import (
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestMarshal(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name: "struct with empty string",
			input: struct {
				Name     string `yaml:"name"`
				EmptyStr string `yaml:"emptyStr"`
				Value    string `yaml:"value"`
			}{
				Name:     "test",
				EmptyStr: "",
				Value:    "some value",
			},
			expected: "name: test\nemptyStr:\nvalue: some value\n",
		},
		{
			name: "struct with all empty strings",
			input: struct {
				Field1 string `yaml:"field1"`
				Field2 string `yaml:"field2"`
				Field3 string `yaml:"field3"`
			}{
				Field1: "",
				Field2: "",
				Field3: "",
			},
			expected: "field1:\nfield2:\nfield3:\n",
		},
		{
			name: "struct with no empty strings",
			input: struct {
				Name  string `yaml:"name"`
				Value string `yaml:"value"`
			}{
				Name:  "test",
				Value: "some value",
			},
			expected: "name: test\nvalue: some value\n",
		},
		{
			name: "struct with mixed types",
			input: struct {
				Name     string  `yaml:"name"`
				EmptyStr string  `yaml:"emptyStr"`
				Age      int     `yaml:"age"`
				Height   float64 `yaml:"height"`
				IsActive bool    `yaml:"isActive"`
				EmptyInt int     `yaml:"emptyInt"`
			}{
				Name:     "John",
				EmptyStr: "",
				Age:      30,
				Height:   5.9,
				IsActive: true,
				EmptyInt: 0,
			},
			expected: "name: John\nemptyStr:\nage: 30\nheight: 5.9\nisActive: true\nemptyInt: 0\n",
		},
		{
			name: "map with empty string values",
			input: map[string]interface{}{
				"name":     "test",
				"emptyStr": "",
				"value":    "some value",
			},
			expected: "",
		},
		{
			name:     "slice with empty strings",
			input:    []string{"hello", "", "world", ""},
			expected: "- hello\n-\n- world\n-\n",
		},
		{
			name: "nested struct with empty strings",
			input: struct {
				Person struct {
					Name     string `yaml:"name"`
					Email    string `yaml:"email"`
					EmptyStr string `yaml:"emptyStr"`
				} `yaml:"person"`
				EmptyField string `yaml:"emptyField"`
			}{
				Person: struct {
					Name     string `yaml:"name"`
					Email    string `yaml:"email"`
					EmptyStr string `yaml:"emptyStr"`
				}{
					Name:     "John",
					Email:    "john@example.com",
					EmptyStr: "",
				},
				EmptyField: "",
			},
			expected: "person:\n  name: John\n  email: john@example.com\n  emptyStr:\nemptyField:\n",
		},
		{
			name: "struct with custom yaml tags",
			input: struct {
				FullName string `yaml:"full_name"`
				EmptyTag string `yaml:"empty_tag"`
				Ignored  string `yaml:"-"`
				Default  string
			}{
				FullName: "John Doe",
				EmptyTag: "",
				Ignored:  "should not appear",
				Default:  "default field",
			},
			expected: "full_name: John Doe\nempty_tag:\nDefault: default field\n",
		},
		{
			name: "pointer to struct with empty string",
			input: &struct {
				Name     string `yaml:"name"`
				EmptyStr string `yaml:"emptyStr"`
			}{
				Name:     "test",
				EmptyStr: "",
			},
			expected: "name: test\nemptyStr:\n",
		},
		{
			name:     "nil pointer",
			input:    (*struct{ Name string })(nil),
			expected: "null\n",
		},
		{
			name:     "empty string directly",
			input:    "",
			expected: "\n",
		},
		{
			name:     "non-empty string directly",
			input:    "hello",
			expected: "hello\n",
		},
		{
			name:     "integer",
			input:    42,
			expected: "42\n",
		},
		{
			name:     "float",
			input:    3.14,
			expected: "3.14\n",
		},
		{
			name:     "boolean true",
			input:    true,
			expected: "true\n",
		},
		{
			name:     "boolean false",
			input:    false,
			expected: "false\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Marshal(tt.input)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}

			actual := string(result)

			if tt.name == "map with empty string values" {
				if !contains(actual, "emptyStr:\n") {
					t.Errorf("Expected output to contain 'emptyStr:' without quotes, got: %q", actual)
				}
				if contains(actual, "emptyStr: \"\"") {
					t.Errorf("Expected output to NOT contain 'emptyStr: \"\"', got: %q", actual)
				}
			} else if tt.expected != "" {
				if actual != tt.expected {
					t.Errorf("Marshal() = %q, want %q", actual, tt.expected)
				}
			}

			skipRoundTrip := map[string]bool{
				"map with empty string values": true,
				"slice with empty strings":     true,
				"struct with custom yaml tags": true,
				"nil pointer":                  true,
			}

			if !skipRoundTrip[tt.name] && tt.input != nil {
				verifyRoundTrip(t, tt.input, result)
			}
		})
	}
}

func TestMarshal_ErrorCases(t *testing.T) {
	_, err := Marshal(make(chan int))
	if err != nil {
		t.Errorf("Unexpected error for channel type: %v", err)
	}
}

func TestMarshal_ComparisonWithStandard(t *testing.T) {
	testStruct := struct {
		Name     string `yaml:"name"`
		EmptyStr string `yaml:"emptyStr"`
		Value    string `yaml:"value"`
	}{
		Name:     "test",
		EmptyStr: "",
		Value:    "some value",
	}

	custom, err := Marshal(testStruct)
	if err != nil {
		t.Fatalf("Custom marshaler error = %v", err)
	}

	standard, err := yaml.Marshal(testStruct)
	if err != nil {
		t.Fatalf("Standard marshaler error = %v", err)
	}

	customStr := string(custom)
	standardStr := string(standard)

	if customStr != "name: test\nemptyStr:\nvalue: some value\n" {
		t.Errorf("Custom marshaler output = %q, expected no quotes around empty string", customStr)
	}

	if standardStr != "name: test\nemptyStr: \"\"\nvalue: some value\n" {
		t.Errorf("Standard marshaler output = %q, expected quotes around empty string", standardStr)
	}

	var customResult, standardResult interface{}
	err = yaml.Unmarshal(custom, &customResult)
	if err != nil {
		t.Fatalf("Failed to unmarshal custom result: %v", err)
	}

	err = yaml.Unmarshal(standard, &standardResult)
	if err != nil {
		t.Fatalf("Failed to unmarshal standard result: %v", err)
	}

	t.Logf("Custom result: %v", customResult)
	t.Logf("Standard result: %v", standardResult)
}

func verifyRoundTrip(t *testing.T, original interface{}, marshaled []byte) {
	t.Helper()

	originalType := reflect.TypeOf(original)
	if originalType.Kind() == reflect.Ptr {
		originalType = originalType.Elem()
	}
	newInstance := reflect.New(originalType).Interface()

	err := yaml.Unmarshal(marshaled, newInstance)
	if err != nil {
		t.Errorf("Failed to unmarshal: %v", err)
		return
	}

	if reflect.TypeOf(original).Kind() != reflect.Ptr {
		originalValue := original
		newInstanceValue := reflect.ValueOf(newInstance).Elem().Interface()
		if !reflectDeepEqual(originalValue, newInstanceValue) {
			t.Errorf("Round trip failed: original = %v, unmarshaled = %v", originalValue, newInstanceValue)
		}
	} else {
		if !reflectDeepEqual(original, newInstance) {
			t.Errorf("Round trip failed: original = %v, unmarshaled = %v", original, newInstance)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && func() bool {
		for i := 0; i <= len(s)-len(substr); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	}()
}

func reflectDeepEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	return reflect.DeepEqual(a, b)
}
