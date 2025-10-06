package internal

import (
	"context"
)

type ViewResource struct {
	ID    uint   `gorm:"primarykey" json:"id" yaml:"id"`
	Name  string `gorm:"uniqueIndex;not null" json:"name" yaml:"name"`
	Query string `gorm:"not null" json:"query" yaml:"query"`
}

type ViewQueryParser interface {
	Parse(query string) (string, error)
}

type ViewResourceRepo interface {
	GetViewResource(ctx context.Context, viewResourceID uint) (*ViewResource, error)
	GetViewResourceByName(ctx context.Context, name string) (*ViewResource, error)
	ListViewResources(ctx context.Context) ([]ViewResource, error)
	CreateViewResource(ctx context.Context, viewResource *ViewResource) error
	UpdateViewResource(ctx context.Context, viewResource *ViewResource) error
	DeleteViewResource(ctx context.Context, viewResourceID uint) error
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

type GetViewResourceRequest struct {
	ViewResourceID uint `json:"viewResourceId" yaml:"viewResourceId"`
}

type GetViewResourceResponse struct {
	ViewResource ViewResource `json:"viewResource" yaml:"viewResource"`
}

func (g *GetViewResource) Exec(ctx context.Context, req GetViewResourceRequest) (*GetViewResourceResponse, error) {
	var viewResource *ViewResource
	err := g.tx.Checkout(ctx, MainBranch, func(ctx context.Context) error {
		var err error
		viewResource, err = g.viewResourceRepo.GetViewResource(ctx, req.ViewResourceID)
		return err
	})
	if err != nil {
		return nil, InternalErrE("failed to get view resource", err)
	}

	if viewResource == nil {
		return nil, UserErr("view resource not found")
	}

	return &GetViewResourceResponse{
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

type ListViewResourcesRequest struct{}

type ListViewResourcesResponse struct {
	ViewResources []ViewResource `json:"viewResources" yaml:"viewResources"`
}

func (l *ListViewResources) Exec(ctx context.Context, req ListViewResourcesRequest) (*ListViewResourcesResponse, error) {
	var viewResources []ViewResource
	err := l.tx.Checkout(ctx, MainBranch, func(ctx context.Context) error {
		var err error
		viewResources, err = l.viewResourceRepo.ListViewResources(ctx)
		return err
	})
	if err != nil {
		return nil, InternalErrE("failed to list view resources", err)
	}

	return &ListViewResourcesResponse{
		ViewResources: viewResources,
	}, nil
}

type CreateViewResource struct {
	viewResourceRepo ViewResourceRepo
	queryParser      ViewQueryParser
	tx               TransactionManager
}

func NewCreateViewResource(viewResourceRepo ViewResourceRepo, queryParser ViewQueryParser, tx TransactionManager) *CreateViewResource {
	return &CreateViewResource{
		viewResourceRepo: viewResourceRepo,
		queryParser:      queryParser,
		tx:               tx,
	}
}

type CreateViewResourceRequest struct {
	Name  string `json:"name" yaml:"name"`
	Query string `json:"query" yaml:"query"`
}

type CreateViewResourceResponse struct {
	ID    uint   `json:"id" yaml:"id"`
	Name  string `json:"name" yaml:"name"`
	Query string `json:"query" yaml:"query"`
}

func (c *CreateViewResource) Exec(ctx context.Context, req CreateViewResourceRequest) (*CreateViewResourceResponse, error) {
	if req.Name == "" {
		return nil, UserErr("name is required")
	}

	if req.Query == "" {
		return nil, UserErr("query is required")
	}

	validatedQuery, err := c.queryParser.Parse(req.Query)
	if err != nil {
		return nil, UserErrE("invalid query", err)
	}

	viewResource := &ViewResource{
		Name:  req.Name,
		Query: validatedQuery,
	}

	var response *CreateViewResourceResponse
	err = c.tx.Do(ctx, MainBranch, "create view resource", func(ctx context.Context) error {
		err := c.viewResourceRepo.CreateViewResource(ctx, viewResource)
		if err != nil {
			return InternalErrE("failed to create view resource", err)
		}

		response = &CreateViewResourceResponse{
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

type UpdateViewResource struct {
	viewResourceRepo ViewResourceRepo
	queryParser      ViewQueryParser
	tx               TransactionManager
}

func NewUpdateViewResource(viewResourceRepo ViewResourceRepo, queryParser ViewQueryParser, tx TransactionManager) *UpdateViewResource {
	return &UpdateViewResource{
		viewResourceRepo: viewResourceRepo,
		queryParser:      queryParser,
		tx:               tx,
	}
}

type UpdateViewResourceRequest struct {
	ViewResourceID uint    `json:"viewResourceId" yaml:"viewResourceId"`
	Name           *string `json:"name,omitempty" yaml:"name,omitempty"`
	Query          *string `json:"query,omitempty" yaml:"query,omitempty"`
}

type UpdateViewResourceResponse struct {
	ID    uint   `json:"id" yaml:"id"`
	Name  string `json:"name" yaml:"name"`
	Query string `json:"query" yaml:"query"`
}

func (u *UpdateViewResource) Exec(ctx context.Context, req UpdateViewResourceRequest) (*UpdateViewResourceResponse, error) {
	if req.ViewResourceID == 0 {
		return nil, UserErr("viewResourceId is required")
	}

	var response *UpdateViewResourceResponse
	err := u.tx.Do(ctx, MainBranch, "update view resource", func(ctx context.Context) error {
		viewResource, err := u.viewResourceRepo.GetViewResource(ctx, req.ViewResourceID)
		if err != nil {
			return InternalErrE("failed to get view resource", err)
		}
		if viewResource == nil {
			return UserErr("view resource not found")
		}

		if req.Name != nil {
			viewResource.Name = *req.Name
		}
		if req.Query != nil {
			validatedQuery, err := u.queryParser.Parse(*req.Query)
			if err != nil {
				return UserErrE("invalid query", err)
			}
			viewResource.Query = validatedQuery
		}

		err = u.viewResourceRepo.UpdateViewResource(ctx, viewResource)
		if err != nil {
			return InternalErrE("failed to update view resource", err)
		}

		response = &UpdateViewResourceResponse{
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

type DeleteViewResourceRequest struct {
	ViewResourceID uint `json:"viewResourceId" yaml:"viewResourceId"`
}

type DeleteViewResourceResponse struct {
	ViewResourceID uint `json:"viewResourceId" yaml:"viewResourceId"`
}

func (d *DeleteViewResource) Exec(ctx context.Context, req DeleteViewResourceRequest) (*DeleteViewResourceResponse, error) {
	if req.ViewResourceID == 0 {
		return nil, UserErr("viewResourceId is required")
	}

	var response *DeleteViewResourceResponse
	err := d.tx.Do(ctx, MainBranch, "delete view resource", func(ctx context.Context) error {
		viewResource, err := d.viewResourceRepo.GetViewResource(ctx, req.ViewResourceID)
		if err != nil {
			return InternalErrE("failed to get view resource", err)
		}
		if viewResource == nil {
			return UserErr("view resource not found")
		}

		err = d.viewResourceRepo.DeleteViewResource(ctx, req.ViewResourceID)
		if err != nil {
			return InternalErrE("failed to delete view resource", err)
		}

		response = &DeleteViewResourceResponse{
			ViewResourceID: req.ViewResourceID,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return response, nil
}
