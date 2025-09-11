package tfmodule

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/infra/tfexec"
)

type Executor struct {
	component *internal.Component
	delegate  internal.Executor
	workDir   string
	tempDir   string
}

func NewExecutor(component *internal.Component, workdir string, logs io.Writer) (internal.Executor, error) {
	tempDir, err := os.MkdirTemp("", "versource-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	tf, err := tfexec.NewExecutor(component, tempDir, logs)
	if err != nil {
		return nil, fmt.Errorf("failed to create terraform instance: %w", err)
	}
	return &Executor{
		component: component,
		delegate:  tf,
		workDir:   workdir,
		tempDir:   tempDir,
	}, nil
}

func (e *Executor) Init(ctx context.Context) error {
	terraformStack, err := newTerraformStackFromComponent(e.component, e.workDir)
	if err != nil {
		return fmt.Errorf("failed to convert component to terraform stack: %w", err)
	}
	mainJSONPath := filepath.Join(e.tempDir, "main.tf.json")
	jsonData, err := json.MarshalIndent(terraformStack, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal stack config: %w", err)
	}
	err = os.WriteFile(mainJSONPath, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write stack config: %w", err)
	}

	modulesDir := filepath.Join(e.workDir, "modules")
	if _, err := os.Stat(modulesDir); err == nil {
		absModulesDir, err := filepath.Abs(modulesDir)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for modules directory: %w", err)
		}
		modulesLink := filepath.Join(e.tempDir, "modules")
		err = os.Symlink(absModulesDir, modulesLink)
		if err != nil {
			return fmt.Errorf("failed to create modules symlink: %w", err)
		}
	}

	return e.delegate.Init(ctx)
}

func (e *Executor) Plan(ctx context.Context) (internal.PlanPath, internal.PlanResourceCounts, error) {
	return e.delegate.Plan(ctx)
}

func (e *Executor) Apply(ctx context.Context, planPath internal.PlanPath) (internal.State, []internal.StateResource, error) {
	return e.delegate.Apply(ctx, planPath)
}

func (e *Executor) Close() error {
	return os.RemoveAll(e.tempDir)
}
