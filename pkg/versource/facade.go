package versource

import (
	"context"
)

type Facade interface {
	Start(ctx context.Context)

	GetModule(ctx context.Context, req GetModuleRequest) (*GetModuleResponse, error)
	ListModules(ctx context.Context, req ListModulesRequest) (*ListModulesResponse, error)
	CreateModule(ctx context.Context, req CreateModuleRequest) (*CreateModuleResponse, error)
	UpdateModule(ctx context.Context, req UpdateModuleRequest) (*UpdateModuleResponse, error)
	DeleteModule(ctx context.Context, req DeleteModuleRequest) (*DeleteModuleResponse, error)
	GetModuleVersion(ctx context.Context, req GetModuleVersionRequest) (*GetModuleVersionResponse, error)
	ListModuleVersions(ctx context.Context, req ListModuleVersionsRequest) (*ListModuleVersionsResponse, error)

	ListChangesets(ctx context.Context, req ListChangesetsRequest) (*ListChangesetsResponse, error)
	CreateChangeset(ctx context.Context, req CreateChangesetRequest) (*CreateChangesetResponse, error)
	DeleteChangeset(ctx context.Context, req DeleteChangesetRequest) (*DeleteChangesetResponse, error)
	EnsureChangeset(ctx context.Context, req EnsureChangesetRequest) (*EnsureChangesetResponse, error)

	GetMerge(ctx context.Context, req GetMergeRequest) (*GetMergeResponse, error)
	ListMerges(ctx context.Context, req ListMergesRequest) (*ListMergesResponse, error)
	CreateMerge(ctx context.Context, req CreateMergeRequest) (*CreateMergeResponse, error)

	GetRebase(ctx context.Context, req GetRebaseRequest) (*GetRebaseResponse, error)
	ListRebases(ctx context.Context, req ListRebasesRequest) (*ListRebasesResponse, error)
	CreateRebase(ctx context.Context, req CreateRebaseRequest) (*CreateRebaseResponse, error)

	GetComponent(ctx context.Context, req GetComponentRequest) (*GetComponentResponse, error)
	ListComponents(ctx context.Context, req ListComponentsRequest) (*ListComponentsResponse, error)
	GetComponentChange(ctx context.Context, req GetComponentChangeRequest) (*GetComponentChangeResponse, error)
	ListComponentChanges(ctx context.Context, req ListComponentChangesRequest) (*ListComponentChangesResponse, error)
	CreateComponent(ctx context.Context, req CreateComponentRequest) (*CreateComponentResponse, error)
	UpdateComponent(ctx context.Context, req UpdateComponentRequest) (*UpdateComponentResponse, error)
	DeleteComponent(ctx context.Context, req DeleteComponentRequest) (*DeleteComponentResponse, error)
	RestoreComponent(ctx context.Context, req RestoreComponentRequest) (*RestoreComponentResponse, error)

	GetPlan(ctx context.Context, req GetPlanRequest) (*GetPlanResponse, error)
	GetPlanLog(ctx context.Context, req GetPlanLogRequest) (*GetPlanLogResponse, error)
	ListPlans(ctx context.Context, req ListPlansRequest) (*ListPlansResponse, error)
	CreatePlan(ctx context.Context, req CreatePlanRequest) (*CreatePlanResponse, error)
	RunPlan(ctx context.Context, planID uint) error

	GetApply(ctx context.Context, req GetApplyRequest) (*GetApplyResponse, error)
	GetApplyLog(ctx context.Context, req GetApplyLogRequest) (*GetApplyLogResponse, error)
	ListApplies(ctx context.Context, req ListAppliesRequest) (*ListAppliesResponse, error)
	RunApply(ctx context.Context, applyID uint) error

	ListResources(ctx context.Context, req ListResourcesRequest) (*ListResourcesResponse, error)

	GetViewResource(ctx context.Context, req GetViewResourceRequest) (*GetViewResourceResponse, error)
	ListViewResources(ctx context.Context, req ListViewResourcesRequest) (*ListViewResourcesResponse, error)
	SaveViewResource(ctx context.Context, req SaveViewResourceRequest) (*SaveViewResourceResponse, error)
	DeleteViewResource(ctx context.Context, req DeleteViewResourceRequest) (*DeleteViewResourceResponse, error)
}
