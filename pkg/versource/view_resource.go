package versource

type ViewResource struct {
	ID    uint   `gorm:"primarykey" json:"id" yaml:"id"`
	Name  string `gorm:"uniqueIndex;not null" json:"name" yaml:"name"`
	Query string `gorm:"not null" json:"query" yaml:"query"`
}

type GetViewResourceRequest struct {
	ViewResourceID uint `json:"viewResourceId" yaml:"viewResourceId"`
}

type GetViewResourceResponse struct {
	ViewResource ViewResource `json:"viewResource" yaml:"viewResource"`
}

type ListViewResourcesRequest struct{}

type ListViewResourcesResponse struct {
	ViewResources []ViewResource `json:"viewResources" yaml:"viewResources"`
}

type SaveViewResourceRequest struct {
	Query string `json:"query" yaml:"query"`
}

type SaveViewResourceResponse struct {
	ID    uint   `json:"id" yaml:"id"`
	Name  string `json:"name" yaml:"name"`
	Query string `json:"query" yaml:"query"`
}

type DeleteViewResourceRequest struct {
	ViewResourceID uint `json:"viewResourceId" yaml:"viewResourceId"`
}

type DeleteViewResourceResponse struct {
	ViewResourceID uint `json:"viewResourceId" yaml:"viewResourceId"`
}
