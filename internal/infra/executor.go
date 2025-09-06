package infra

import (
	"fmt"
	"io"

	"github.com/marcbran/versource/internal"
	tfjsonnet "github.com/marcbran/versource/internal/infra/terraform-jsonnet"
	tfmodule "github.com/marcbran/versource/internal/infra/terraform-module"
)

const (
	ExecutorTypeTerraformModule  = "terraform-module"
	ExecutorTypeTerraformJsonnet = "terraform-jsonnet"
)

func NewExecutor(component *internal.Component, workdir string, logs io.Writer) (internal.Executor, error) {
	executorType := component.ModuleVersion.Module.ExecutorType

	switch executorType {
	case ExecutorTypeTerraformModule:
		return tfmodule.NewExecutor(component, workdir, logs)
	case ExecutorTypeTerraformJsonnet:
		return tfjsonnet.NewExecutor(component, workdir, logs)
	default:
		return nil, fmt.Errorf("unknown executor type: %s", executorType)
	}
}
