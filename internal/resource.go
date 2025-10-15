package internal

import (
	"context"

	"github.com/marcbran/versource/pkg/versource"
)

type ResourceRepo interface {
	InsertResources(ctx context.Context, resources []versource.Resource) error
	UpdateResources(ctx context.Context, resources []versource.Resource) error
	DeleteResources(ctx context.Context, resourceUUIDs []string) error
	ListResources(ctx context.Context) ([]versource.Resource, error)
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

func (l *ListResources) Exec(ctx context.Context, req versource.ListResourcesRequest) (*versource.ListResourcesResponse, error) {
	var resources []versource.Resource
	err := l.tx.Checkout(ctx, MainBranch, func(ctx context.Context) error {
		var err error
		resources, err = l.resourceRepo.ListResources(ctx)
		return err
	})
	if err != nil {
		return nil, versource.InternalErrE("failed to list resources", err)
	}

	return &versource.ListResourcesResponse{
		Resources: resources,
	}, nil
}

func applyResourceMapping(stateResources []versource.StateResource, mapping versource.ResourceMapping) []versource.StateResource {
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

	var filtered []versource.StateResource
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
			stateResource := versource.StateResource{
				Address:  resource.GenerateUUID(),
				Mode:     versource.AddedResourceMode,
				Resource: resource,
			}
			filtered = append(filtered, stateResource)
		}
	}

	return filtered
}
