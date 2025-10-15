package internal

import (
	"context"

	"github.com/marcbran/versource/pkg/versource"
)

type StateRepo interface {
	UpsertState(ctx context.Context, state *versource.State) error
}

type StateResourceRepo interface {
	ListStateResourcesByStateID(ctx context.Context, stateID uint) ([]versource.StateResource, error)
	InsertStateResources(ctx context.Context, resources []versource.StateResource) error
	UpdateStateResources(ctx context.Context, resources []versource.StateResource) error
	DeleteStateResources(ctx context.Context, stateResourceIDs []uint) error
}
