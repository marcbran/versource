package internal

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/google/uuid"
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
	UpsertStateResources(ctx context.Context, resources []StateResource) error
}

type Resource struct {
	ID            string         `gorm:"primarykey;type:varchar(36)" json:"id" yaml:"id"`
	Provider      string         `json:"provider" yaml:"provider"`
	ProviderAlias *string        `json:"providerAlias" yaml:"providerAlias"`
	ResourceType  string         `json:"resourceType" yaml:"resourceType"`
	Namespace     *string        `json:"namespace" yaml:"namespace"`
	Name          string         `json:"name" yaml:"name"`
	Attributes    datatypes.JSON `gorm:"type:jsonb" json:"attributes" yaml:"attributes"`
}

func (r *Resource) GenerateID() {
	providerAlias := ""
	if r.ProviderAlias != nil {
		providerAlias = *r.ProviderAlias
	}
	namespace := ""
	if r.Namespace != nil {
		namespace = *r.Namespace
	}

	input := fmt.Sprintf("%s%s%s%s%s", r.Provider, providerAlias, r.ResourceType, namespace, r.Name)
	hash := sha256.Sum256([]byte(input))

	namespaceUUID := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	resourceUUID := uuid.NewSHA1(namespaceUUID, hash[:])

	r.ID = resourceUUID.String()
}

type ResourceRepo interface {
	UpsertResources(ctx context.Context, resources []Resource) error
}
