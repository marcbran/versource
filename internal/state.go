package internal

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type State struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Component   Component      `gorm:"foreignKey:ComponentID" json:"component"`
	ComponentID uint           `gorm:"uniqueIndex" json:"componentId"`
	Output      datatypes.JSON `gorm:"type:jsonb" json:"output"`
}

type StateRepo interface {
	UpsertState(ctx context.Context, state *State) error
}

type StateResource struct {
	ID           uint         `gorm:"primarykey" json:"id"`
	State        State        `gorm:"foreignKey:StateID" json:"state"`
	StateID      uint         `json:"stateId"`
	Resource     Resource     `gorm:"foreignKey:ResourceID" json:"resource"`
	ResourceID   string       `json:"resourceId"`
	Mode         ResourceMode `json:"mode"`
	ProviderName string       `json:"providerName"`
	Type         string       `json:"type"`
	Address      string       `json:"address"`
	Count        *int         `json:"count"`
	ForEach      *string      `json:"forEach"`
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
	ID            string         `gorm:"primarykey;type:varchar(36)" json:"id"`
	Provider      string         `json:"provider"`
	ProviderAlias *string        `json:"providerAlias"`
	ResourceType  string         `json:"resourceType"`
	Namespace     *string        `json:"namespace"`
	Name          string         `json:"name"`
	Attributes    datatypes.JSON `gorm:"type:jsonb" json:"attributes"`
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
