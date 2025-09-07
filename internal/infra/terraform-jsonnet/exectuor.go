package tfjsonnet

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jsonnet-bundler/jsonnet-bundler/pkg"
	"github.com/jsonnet-bundler/jsonnet-bundler/spec/v1/deps"
	"github.com/marcbran/jpoet/pkg/jpoet"
	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/infra/terraform-jsonnet/lib/imports"
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
	terraformDir := filepath.Join(tempDir, ".terraform-jsonnet")
	err = os.MkdirAll(terraformDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create terraform jsonnet directory: %w", err)
	}
	tf, err := tfexec.NewExecutor(component, terraformDir, logs)
	if err != nil {
		return nil, fmt.Errorf("failed to create terraform instance: %w", err)
	}
	return Executor{
		component: component,
		delegate:  tf,
		workDir:   workdir,
		tempDir:   tempDir,
	}, nil
}

func (e Executor) Init(ctx context.Context) error {
	vendorDir := filepath.Join(e.tempDir, "vendor")
	err := os.MkdirAll(vendorDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create vendor directory: %w", err)
	}
	tmpDir := filepath.Join(vendorDir, ".tmp")
	err = os.MkdirAll(tmpDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create tmp directory: %w", err)
	}
	file := newJsonnetBundlerFromComponent(e.component)
	_, err = pkg.Ensure(file, vendorDir, file.Dependencies)
	if err != nil {
		return fmt.Errorf("failed to ensure dependencies: %w", err)
	}
	importSource := strings.TrimPrefix(strings.TrimPrefix(e.component.ModuleVersion.Module.Source, deps.GitSchemeHTTPS), deps.GitSchemeSSH)
	statePath, err := filepath.Abs(filepath.Join(e.workDir, "states", fmt.Sprintf("%d.tfstate", e.component.ID)))
	if err != nil {
		return fmt.Errorf("failed to get absolute path for state file: %w", err)
	}
	terraformDir := filepath.Join(e.tempDir, ".terraform-jsonnet")
	err = jpoet.NewEval().
		FileImport([]string{vendorDir}).
		FSImport(lib).
		FSImport(imports.Fs).
		Serialize(false).
		TLACode("module", fmt.Sprintf("import '%s/main.tf.jsonnet'", importSource)).
		TLACode("var", string(e.component.Variables)).
		TLAVar("statePath", statePath).
		FileInput("./lib/gen.libsonnet").
		DirectoryOutput(terraformDir).
		Eval()
	if err != nil {
		return err
	}
	return e.delegate.Init(ctx)
}

func (e Executor) Plan(ctx context.Context) (internal.PlanPath, error) {
	return e.delegate.Plan(ctx)
}

func (e Executor) Apply(ctx context.Context, planPath internal.PlanPath) (internal.State, []internal.Resource, error) {
	return e.delegate.Apply(ctx, planPath)
}

func (e Executor) Close() error {
	return os.RemoveAll(e.tempDir)
}
