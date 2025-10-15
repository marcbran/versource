package internal

import (
	"context"

	"github.com/marcbran/versource/pkg/versource"
)

type ViewQueryParser interface {
	Parse(query string) (*versource.ViewResource, error)
}

type ViewResourceRepo interface {
	GetViewResource(ctx context.Context, viewResourceID uint) (*versource.ViewResource, error)
	GetViewResourceByName(ctx context.Context, name string) (*versource.ViewResource, error)
	ListViewResources(ctx context.Context) ([]versource.ViewResource, error)
	CreateViewResource(ctx context.Context, viewResource *versource.ViewResource) error
	UpdateViewResource(ctx context.Context, viewResource *versource.ViewResource) error
	DeleteViewResource(ctx context.Context, viewResourceID uint) error
	SaveDatabaseView(ctx context.Context, name, query string) error
	DropDatabaseView(ctx context.Context, name string) error
}

type GetViewResource struct {
	viewResourceRepo ViewResourceRepo
	tx               TransactionManager
}

func NewGetViewResource(viewResourceRepo ViewResourceRepo, tx TransactionManager) *GetViewResource {
	return &GetViewResource{
		viewResourceRepo: viewResourceRepo,
		tx:               tx,
	}
}

func (g *GetViewResource) Exec(ctx context.Context, req versource.GetViewResourceRequest) (*versource.GetViewResourceResponse, error) {
	var viewResource *versource.ViewResource
	err := g.tx.Checkout(ctx, MainBranch, func(ctx context.Context) error {
		var err error
		viewResource, err = g.viewResourceRepo.GetViewResource(ctx, req.ViewResourceID)
		return err
	})
	if err != nil {
		return nil, versource.InternalErrE("failed to get view resource", err)
	}

	if viewResource == nil {
		return nil, versource.UserErr("view resource not found")
	}

	return &versource.GetViewResourceResponse{
		ViewResource: *viewResource,
	}, nil
}

type ListViewResources struct {
	viewResourceRepo ViewResourceRepo
	tx               TransactionManager
}

func NewListViewResources(viewResourceRepo ViewResourceRepo, tx TransactionManager) *ListViewResources {
	return &ListViewResources{
		viewResourceRepo: viewResourceRepo,
		tx:               tx,
	}
}

func (l *ListViewResources) Exec(ctx context.Context, req versource.ListViewResourcesRequest) (*versource.ListViewResourcesResponse, error) {
	var viewResources []versource.ViewResource
	err := l.tx.Checkout(ctx, MainBranch, func(ctx context.Context) error {
		var err error
		viewResources, err = l.viewResourceRepo.ListViewResources(ctx)
		return err
	})
	if err != nil {
		return nil, versource.InternalErrE("failed to list view resources", err)
	}

	return &versource.ListViewResourcesResponse{
		ViewResources: viewResources,
	}, nil
}

type SaveViewResource struct {
	viewResourceRepo ViewResourceRepo
	queryParser      ViewQueryParser
	tx               TransactionManager
}

func NewSaveViewResource(viewResourceRepo ViewResourceRepo, queryParser ViewQueryParser, tx TransactionManager) *SaveViewResource {
	return &SaveViewResource{
		viewResourceRepo: viewResourceRepo,
		queryParser:      queryParser,
		tx:               tx,
	}
}

func (s *SaveViewResource) Exec(ctx context.Context, req versource.SaveViewResourceRequest) (*versource.SaveViewResourceResponse, error) {
	if req.Query == "" {
		return nil, versource.UserErr("query is required")
	}

	viewResource, err := s.queryParser.Parse(req.Query)
	if err != nil {
		return nil, versource.UserErrE("invalid query", err)
	}

	var response *versource.SaveViewResourceResponse
	err = s.tx.Do(ctx, MainBranch, "save view resource", func(ctx context.Context) error {
		existing, err := s.viewResourceRepo.GetViewResourceByName(ctx, viewResource.Name)
		if err != nil {
			return versource.InternalErrE("failed to check existing view resource", err)
		}

		if existing != nil {
			existing.Query = viewResource.Query
			err = s.viewResourceRepo.UpdateViewResource(ctx, existing)
			if err != nil {
				return versource.InternalErrE("failed to update view resource", err)
			}
			viewResource = existing
		} else {
			err = s.viewResourceRepo.CreateViewResource(ctx, viewResource)
			if err != nil {
				return versource.InternalErrE("failed to create view resource", err)
			}
		}

		err = s.viewResourceRepo.SaveDatabaseView(ctx, viewResource.Name, viewResource.Query)
		if err != nil {
			return versource.InternalErrE("failed to save database view", err)
		}

		response = &versource.SaveViewResourceResponse{
			ID:    viewResource.ID,
			Name:  viewResource.Name,
			Query: viewResource.Query,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return response, nil
}

type DeleteViewResource struct {
	viewResourceRepo ViewResourceRepo
	tx               TransactionManager
}

func NewDeleteViewResource(viewResourceRepo ViewResourceRepo, tx TransactionManager) *DeleteViewResource {
	return &DeleteViewResource{
		viewResourceRepo: viewResourceRepo,
		tx:               tx,
	}
}

func (d *DeleteViewResource) Exec(ctx context.Context, req versource.DeleteViewResourceRequest) (*versource.DeleteViewResourceResponse, error) {
	if req.ViewResourceID == 0 {
		return nil, versource.UserErr("viewResourceId is required")
	}

	var response *versource.DeleteViewResourceResponse
	err := d.tx.Do(ctx, MainBranch, "delete view resource", func(ctx context.Context) error {
		viewResource, err := d.viewResourceRepo.GetViewResource(ctx, req.ViewResourceID)
		if err != nil {
			return versource.InternalErrE("failed to get view resource", err)
		}
		if viewResource == nil {
			return versource.UserErr("view resource not found")
		}

		err = d.viewResourceRepo.DeleteViewResource(ctx, req.ViewResourceID)
		if err != nil {
			return versource.InternalErrE("failed to delete view resource", err)
		}

		err = d.viewResourceRepo.DropDatabaseView(ctx, viewResource.Name)
		if err != nil {
			return versource.InternalErrE("failed to drop database view", err)
		}

		response = &versource.DeleteViewResourceResponse{
			ViewResourceID: req.ViewResourceID,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return response, nil
}
