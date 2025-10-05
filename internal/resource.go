package internal

import (
	"context"
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

type ResourceRepo interface {
	InsertResources(ctx context.Context, resources []Resource) error
	UpdateResources(ctx context.Context, resources []Resource) error
	DeleteResources(ctx context.Context, resourceUUIDs []string) error
	ListResources(ctx context.Context) ([]Resource, error)
}

type ListResources struct {
	resourceRepo ResourceRepo
	tx           TransactionManager
}

func NewListResources(resourceRepo ResourceRepo, tx TransactionManager) *ListResources {
	return &ListResources{
		resourceRepo: resourceRepo,
		tx:           tx,
	}
}

type ListResourcesRequest struct{}

type ListResourcesResponse struct {
	Resources []Resource `json:"resources" yaml:"resources"`
}

func (l *ListResources) Exec(ctx context.Context, req ListResourcesRequest) (*ListResourcesResponse, error) {
	var resources []Resource
	err := l.tx.Checkout(ctx, MainBranch, func(ctx context.Context) error {
		var err error
		resources, err = l.resourceRepo.ListResources(ctx)
		return err
	})
	if err != nil {
		return nil, InternalErrE("failed to list resources", err)
	}

	return &ListResourcesResponse{
		Resources: resources,
	}, nil
}

func applyResourceMapping(stateResources []StateResource, mapping ResourceMapping) []StateResource {
	keepMap := make(map[string]bool)
	if mapping.Keep != nil {
		for _, item := range *mapping.Keep {
			keepMap[string(item)] = true
		}
	}

	dropMap := make(map[string]bool)
	if mapping.Drop != nil {
		for _, item := range *mapping.Drop {
			dropMap[string(item)] = true
		}
	}

	var filtered []StateResource
	for _, sr := range stateResources {
		resourceKey := string(sr.Resource.Attributes)

		if mapping.Keep != nil && !keepMap[resourceKey] {
			continue
		}

		if dropMap[resourceKey] && !keepMap[resourceKey] {
			continue
		}

		filtered = append(filtered, sr)
	}

	if mapping.Add != nil {
		for _, resource := range *mapping.Add {
			stateResource := StateResource{
				Address:  resource.GenerateUUID(),
				Mode:     AddedResourceMode,
				Resource: resource,
			}
			filtered = append(filtered, stateResource)
		}
	}

	return filtered
}
