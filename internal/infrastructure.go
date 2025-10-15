package internal

import (
	"context"
	"io"

	"github.com/marcbran/versource/pkg/versource"
)

type NewExecutor func(component *versource.Component, workdir string, logs io.Writer) (Executor, error)

type Executor interface {
	io.Closer
	Init(ctx context.Context) error
	Plan(ctx context.Context) (PlanPath, PlanResourceCounts, error)
	Apply(ctx context.Context, planPath PlanPath) (versource.State, []versource.StateResource, error)
}

type PlanPath string

type PlanResourceCounts struct {
	AddCount     int
	ChangeCount  int
	DestroyCount int
}

type LogStore interface {
	NewLogWriter(operationType string, operationID uint) (io.WriteCloser, error)
	StoreLog(ctx context.Context, operationType string, operationID uint, r io.Reader) error
	LoadLog(ctx context.Context, operationType string, operationID uint) (io.ReadCloser, error)
	DeleteLog(ctx context.Context, operationType string, operationID uint) error
}
