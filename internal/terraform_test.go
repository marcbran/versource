package internal

import (
	"encoding/json"
	"testing"
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
        "source": "./modules/file",
        "version": "1.0.0",
        "variables": {
          "content": "test content",
          "filename": "test.txt"
        }
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
        "source": "./modules/file",
        "version": "1.0.0",
        "variables": {
          "content": "test content",
          "filename": "test.txt"
        }
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
