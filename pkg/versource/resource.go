package versource

import (
	"crypto/sha256"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type Resource struct {
	UUID          string         `gorm:"primarykey;type:varchar(36)" json:"uuid" yaml:"uuid"`
	Provider      string         `json:"provider" yaml:"provider"`
	ProviderAlias *string        `json:"providerAlias" yaml:"providerAlias"`
	ResourceType  string         `json:"resourceType" yaml:"resourceType"`
	Namespace     *string        `json:"namespace" yaml:"namespace"`
	Name          string         `json:"name" yaml:"name"`
	Attributes    datatypes.JSON `gorm:"type:jsonb" json:"attributes" yaml:"attributes"`
}

func (r Resource) GenerateUUID() string {
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

	return resourceUUID.String()
}

type ResourceMapping struct {
	Keep *[]datatypes.JSON `json:"keep,omitempty"`
	Drop *[]datatypes.JSON `json:"drop,omitempty"`
	Add  *[]Resource       `json:"add,omitempty"`
}

type ListResourcesRequest struct{}

type ListResourcesResponse struct {
	Resources []Resource `json:"resources" yaml:"resources"`
}
