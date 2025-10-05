package internal

import (
	"context"

	"gorm.io/datatypes"
)

type State struct {
	ID          uint           `gorm:"primarykey" json:"id" yaml:"id"`
	Component   Component      `gorm:"foreignKey:ComponentID" json:"component" yaml:"component"`
	ComponentID uint           `gorm:"uniqueIndex" json:"componentId" yaml:"componentId"`
	Output      datatypes.JSON `gorm:"type:jsonb" json:"output" yaml:"output"`
}

type StateRepo interface {
	UpsertState(ctx context.Context, state *State) error
}

type StateResource struct {
	ID           uint         `gorm:"primarykey" json:"id" yaml:"id"`
	State        State        `gorm:"foreignKey:StateID" json:"state" yaml:"state"`
	StateID      uint         `json:"stateId" yaml:"stateId"`
	Resource     Resource     `gorm:"foreignKey:ResourceID" json:"resource" yaml:"resource"`
	ResourceID   string       `json:"resourceId" yaml:"resourceId"`
	Mode         ResourceMode `json:"mode" yaml:"mode"`
	ProviderName string       `json:"providerName" yaml:"providerName"`
	Type         string       `json:"type" yaml:"type"`
	Address      string       `json:"address" yaml:"address"`
	Count        *int         `json:"count" yaml:"count"`
	ForEach      *string      `json:"forEach" yaml:"forEach"`
}

type ResourceMode string

const (
	DataResourceMode    ResourceMode = "data"
	ManagedResourceMode ResourceMode = "managed"
)

type StateResourceRepo interface {
	ListStateResourcesByStateID(ctx context.Context, stateID uint) ([]StateResource, error)
	InsertStateResources(ctx context.Context, resources []StateResource) error
	UpdateStateResources(ctx context.Context, resources []StateResource) error
	DeleteStateResources(ctx context.Context, stateResourceIDs []uint) error
}
