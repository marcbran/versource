package internal

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type State struct {
	ID          uint           `gorm:"primarykey"`
	Component   Component      `gorm:"foreignKey:ComponentID"`
	ComponentID uint           `gorm:"uniqueIndex"`
	Output      datatypes.JSON `gorm:"type:jsonb"`
}

type StateRepo interface {
	UpsertState(ctx context.Context, state *State) error
}

type StateResource struct {
	ID           uint  `gorm:"primarykey"`
	State        State `gorm:"foreignKey:StateID"`
	StateID      uint
	Resource     Resource `gorm:"foreignKey:ResourceID"`
	ResourceID   string
	Mode         ResourceMode
	ProviderName string
	Type         string
	Address      string
	Count        *int
	ForEach      *string
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
	ID            string `gorm:"primarykey;type:varchar(36)"`
	Provider      string
	ProviderAlias *string
	ResourceType  string
	Namespace     *string
	Name          string
	Attributes    datatypes.JSON `gorm:"type:jsonb"`
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
