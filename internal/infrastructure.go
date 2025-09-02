package internal

import (
	"context"
	"io"
)

type NewExecutor func(component *Component, workdir string) (Executor, error)

type Executor interface {
	io.Closer
	Init(ctx context.Context, logs io.Writer) error
	Plan(ctx context.Context, logs io.Writer) (PlanPath, error)
	Apply(ctx context.Context, planPath PlanPath, logs io.Writer) (State, []Resource, error)
}

type PlanPath string

type LogStore interface {
	NewLogWriter(operationType string, operationID uint) (io.WriteCloser, error)
	StoreLog(ctx context.Context, operationType string, operationID uint, r io.Reader) error
	LoadLog(ctx context.Context, operationType string, operationID uint) (io.ReadCloser, error)
}
