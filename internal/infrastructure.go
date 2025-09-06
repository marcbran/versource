package internal

import (
	"context"
	"io"
)

type NewExecutor func(component *Component, workdir string, logs io.Writer) (Executor, error)

type Executor interface {
	io.Closer
	Init(ctx context.Context) error
	Plan(ctx context.Context) (PlanPath, error)
	Apply(ctx context.Context, planPath PlanPath) (State, []Resource, error)
}

type PlanPath string

type ModuleIngester interface {
	IngestModuleWithVersion(request CreateModuleRequest) (*Module, *ModuleVersion, error)
}

type LogStore interface {
	NewLogWriter(operationType string, operationID uint) (io.WriteCloser, error)
	StoreLog(ctx context.Context, operationType string, operationID uint, r io.Reader) error
	LoadLog(ctx context.Context, operationType string, operationID uint) (io.ReadCloser, error)
}
