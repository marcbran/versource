package tfmodule

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/marcbran/versource/internal"
)

func TestTerraformStack_AddModule(t *testing.T) {
	stack := NewTerraformStack()

	module := TerraformModule{
		Source:  "./modules/file",
		Version: "1.0.0",
		Variables: map[string]any{
			"filename": "test.txt",
			"content":  "test content",
		},
	}

	stack = stack.AddModule("file_module", module)

	if len(stack) != 1 {
		t.Errorf("Expected config to have 1 item, got %d", len(stack))
	}

	container, ok := stack[0].(TerraformModuleContainer)
	if !ok {
		t.Fatal("Expected first item to be TerraformModuleContainer")
	}

	if len(container.Module) != 1 {
		t.Errorf("Expected module container to have 1 module, got %d", len(container.Module))
	}

	addedModule, exists := container.Module["file_module"]
	if !exists {
		t.Fatal("Expected module 'file_module' to exist")
	}

	if addedModule.Source != "./modules/file" {
		t.Errorf("Expected source to be './modules/file', got %s", addedModule.Source)
	}

	if addedModule.Version != "1.0.0" {
		t.Errorf("Expected version to be '1.0.0', got %s", addedModule.Version)
	}
}

func TestTerraformStack_AddOutput(t *testing.T) {
	stack := NewTerraformStack()

	output := TerraformOutput{
		Value: "file_content",
	}

	stack = stack.AddOutput("file_output", output)

	if len(stack) != 1 {
		t.Errorf("Expected config to have 1 item, got %d", len(stack))
	}

	container, ok := stack[0].(TerraformOutputContainer)
	if !ok {
		t.Fatal("Expected first item to be TerraformOutputContainer")
	}

	if len(container.Output) != 1 {
		t.Errorf("Expected output container to have 1 output, got %d", len(container.Output))
	}

	addedOutput, exists := container.Output["file_output"]
	if !exists {
		t.Fatal("Expected output 'file_output' to exist")
	}

	if addedOutput.Value != "file_content" {
		t.Errorf("Expected value to be 'file_content', got %v", addedOutput.Value)
	}
}

func TestTerraformStack_JSONMarshaling(t *testing.T) {
	stack := NewTerraformStack()

	module := TerraformModule{
		Source:  "./modules/file",
		Version: "1.0.0",
		Variables: map[string]any{
			"filename": "test.txt",
			"content":  "test content",
		},
	}
	stack = stack.AddModule("file_module", module)

	output := TerraformOutput{
		Value: "file_content",
	}
	stack = stack.AddOutput("file_output", output)

	jsonData, err := json.MarshalIndent(stack, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	expected := `[
  {
    "module": {
      "file_module": {
        "content": "test content",
        "filename": "test.txt",
        "source": "./modules/file",
        "version": "1.0.0"
      }
    }
  },
  {
    "output": {
      "file_output": {
        "value": "file_content"
      }
    }
  }
]`

	if string(jsonData) != expected {
		t.Errorf("Expected JSON:\n%s\n\nGot JSON:\n%s", expected, string(jsonData))
	}
}

func TestTerraformStack_AddBackend(t *testing.T) {
	stack := NewTerraformStack()

	backend := TerraformBackend{
		Local: TerraformBackendLocal{
			Path: "/tmp/states/123.tfstate",
		},
	}

	stack = stack.AddBackend("backend", backend)

	if len(stack) != 1 {
		t.Errorf("Expected config to have 1 item, got %d", len(stack))
	}

	container, ok := stack[0].(TerraformTerraformContainer)
	if !ok {
		t.Fatal("Expected first item to be TerraformTerraformContainer")
	}

	backendConfig := container.Terraform
	addedBackend := backendConfig.Backend
	if addedBackend.Local.Path != "/tmp/states/123.tfstate" {
		t.Errorf("Expected backend path to be '/tmp/states/123.tfstate', got %s", addedBackend.Local.Path)
	}
}

func TestTerraformStack_JSONMarshalingWithBackend(t *testing.T) {
	stack := NewTerraformStack()

	module := TerraformModule{
		Source:  "./modules/file",
		Version: "1.0.0",
		Variables: map[string]any{
			"filename": "test.txt",
			"content":  "test content",
		},
	}
	stack = stack.AddModule("file_module", module)

	output := TerraformOutput{
		Value: "file_content",
	}
	stack = stack.AddOutput("file_output", output)

	backend := TerraformBackend{
		Local: TerraformBackendLocal{
			Path: "/tmp/states/123.tfstate",
		},
	}
	stack = stack.AddBackend("backend", backend)

	jsonData, err := json.MarshalIndent(stack, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	expected := `[
  {
    "module": {
      "file_module": {
        "content": "test content",
        "filename": "test.txt",
        "source": "./modules/file",
        "version": "1.0.0"
      }
    }
  },
  {
    "output": {
      "file_output": {
        "value": "file_content"
      }
    }
  },
  {
    "terraform": {
      "backend": {
        "local": {
          "path": "/tmp/states/123.tfstate"
        }
      }
    }
  }
]`

	if string(jsonData) != expected {
		t.Errorf("Expected JSON:\n%s\n\nGot JSON:\n%s", expected, string(jsonData))
	}
}

func TestBuildTerraformModule(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		version     string
		variables   []byte
		expected    TerraformModule
		expectError bool
		errorMsg    string
	}{
		{
			name:    "local path without version",
			source:  "./modules/file",
			version: "",
			expected: TerraformModule{
				Source:  "./modules/file",
				Version: "",
			},
		},
		{
			name:        "local path with version should error",
			source:      "./modules/file",
			version:     "1.0.0",
			expectError: true,
			errorMsg:    "local paths do not support version parameter",
		},
		{
			name:    "terraform registry with version",
			source:  "hashicorp/aws/aws",
			version: "5.0.0",
			expected: TerraformModule{
				Source:  "hashicorp/aws/aws",
				Version: "5.0.0",
			},
		},
		{
			name:        "terraform registry without version should error",
			source:      "hashicorp/aws/aws",
			version:     "",
			expectError: true,
			errorMsg:    "terraform registry sources require version parameter",
		},
		{
			name:    "github source with version",
			source:  "github.com/terraform-aws-modules/terraform-aws-vpc",
			version: "main",
			expected: TerraformModule{
				Source:  "github.com/terraform-aws-modules/terraform-aws-vpc?ref=main",
				Version: "",
			},
		},
		{
			name:    "github source with version and existing query params",
			source:  "github.com/terraform-aws-modules/terraform-aws-vpc?submodules=true",
			version: "main",
			expected: TerraformModule{
				Source:  "github.com/terraform-aws-modules/terraform-aws-vpc?submodules=true&ref=main",
				Version: "",
			},
		},
		{
			name:    "github source without version",
			source:  "github.com/terraform-aws-modules/terraform-aws-vpc",
			version: "",
			expected: TerraformModule{
				Source:  "github.com/terraform-aws-modules/terraform-aws-vpc",
				Version: "",
			},
		},
		{
			name:    "git source with version",
			source:  "git::https://github.com/terraform-aws-modules/terraform-aws-vpc",
			version: "v1.0.0",
			expected: TerraformModule{
				Source:  "git::https://github.com/terraform-aws-modules/terraform-aws-vpc?ref=v1.0.0",
				Version: "",
			},
		},
		{
			name:    "bitbucket source with version",
			source:  "bitbucket.org/company/terraform-module",
			version: "develop",
			expected: TerraformModule{
				Source:  "bitbucket.org/company/terraform-module?ref=develop",
				Version: "",
			},
		},
		{
			name:    "mercurial source with version",
			source:  "hg::https://bitbucket.org/company/terraform-module",
			version: "default",
			expected: TerraformModule{
				Source:  "hg::https://bitbucket.org/company/terraform-module?ref=default",
				Version: "",
			},
		},
		{
			name:    "S3 source with version",
			source:  "s3::https://my-bucket.s3.amazonaws.com/modules/vpc.zip",
			version: "abc123",
			expected: TerraformModule{
				Source:  "s3::https://my-bucket.s3.amazonaws.com/modules/vpc.zip?versionId=abc123",
				Version: "",
			},
		},
		{
			name:    "S3 source with version and existing query params",
			source:  "s3::https://my-bucket.s3.amazonaws.com/modules/vpc.zip?region=us-west-2",
			version: "abc123",
			expected: TerraformModule{
				Source:  "s3::https://my-bucket.s3.amazonaws.com/modules/vpc.zip?region=us-west-2&versionId=abc123",
				Version: "",
			},
		},
		{
			name:    "GCS source with version",
			source:  "gcs::https://storage.googleapis.com/my-bucket/modules/vpc.zip",
			version: "1234567890",
			expected: TerraformModule{
				Source:  "gcs::https://storage.googleapis.com/my-bucket/modules/vpc.zip?generation=1234567890",
				Version: "",
			},
		},
		{
			name:    "GCS source with version and existing query params",
			source:  "gcs::https://storage.googleapis.com/my-bucket/modules/vpc.zip?project=my-project",
			version: "1234567890",
			expected: TerraformModule{
				Source:  "gcs::https://storage.googleapis.com/my-bucket/modules/vpc.zip?project=my-project&generation=1234567890",
				Version: "",
			},
		},
		{
			name:      "with variables",
			source:    "hashicorp/aws/aws",
			version:   "5.0.0",
			variables: []byte(`{"region": "us-west-2", "instance_type": "t3.micro"}`),
			expected: TerraformModule{
				Source:  "hashicorp/aws/aws",
				Version: "5.0.0",
				Variables: map[string]any{
					"region":        "us-west-2",
					"instance_type": "t3.micro",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component := &versource.Component{
				ModuleVersion: versource.ModuleVersion{
					Module: versource.Module{
						Source: tt.source,
					},
					Version: tt.version,
				},
				Variables: tt.variables,
			}

			result, err := buildTerraformModule(component)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result.Source != tt.expected.Source {
				t.Errorf("Expected source '%s', got '%s'", tt.expected.Source, result.Source)
			}

			if result.Version != tt.expected.Version {
				t.Errorf("Expected version '%s', got '%s'", tt.expected.Version, result.Version)
			}

			if tt.variables != nil {
				if result.Variables == nil {
					t.Errorf("Expected variables but got nil")
					return
				}

				for key, expectedValue := range tt.expected.Variables {
					if actualValue, exists := result.Variables[key]; !exists {
						t.Errorf("Expected variable '%s' but it doesn't exist", key)
					} else if actualValue != expectedValue {
						t.Errorf("Expected variable '%s' to be '%v', got '%v'", key, expectedValue, actualValue)
					}
				}
			}
		})
	}
}
