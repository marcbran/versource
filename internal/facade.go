package internal

import (
	"context"

	"github.com/marcbran/versource/pkg/versource"
)

type facade struct {
	getModule          *GetModule
	listModules        *ListModules
	createModule       *CreateModule
	updateModule       *UpdateModule
	deleteModule       *DeleteModule
	getModuleVersion   *GetModuleVersion
	listModuleVersions *ListModuleVersions

	listChangesets  *ListChangesets
	createChangeset *CreateChangeset
	deleteChangeset *DeleteChangeset
	ensureChangeset *EnsureChangeset

	getMerge    *GetMerge
	listMerges  *ListMerges
	createMerge *CreateMerge

	getRebase    *GetRebase
	listRebases  *ListRebases
	createRebase *CreateRebase

	getComponent         *GetComponent
	listComponents       *ListComponents
	getComponentChange   *GetComponentChange
	listComponentChanges *ListComponentChanges
	createComponent      *CreateComponent
	updateComponent      *UpdateComponent
	deleteComponent      *DeleteComponent
	restoreComponent     *RestoreComponent

	getPlan    *GetPlan
	getPlanLog *GetPlanLog
	listPlans  *ListPlans
	createPlan *CreatePlan
	runPlan    *RunPlan

	getApply    *GetApply
	getApplyLog *GetApplyLog
	listApplies *ListApplies
	runApply    *RunApply

	listResources *ListResources

	getViewResource    *GetViewResource
	listViewResources  *ListViewResources
	saveViewResource   *SaveViewResource
	deleteViewResource *DeleteViewResource

	planWorker   *PlanWorker
	applyWorker  *ApplyWorker
	mergeWorker  *MergeWorker
	rebaseWorker *RebaseWorker
}

func NewFacade(
	config *versource.Config,
	componentRepo ComponentRepo,
	componentChangeRepo ComponentChangeRepo,
	stateRepo StateRepo,
	stateResourceRepo StateResourceRepo,
	resourceRepo ResourceRepo,
	planRepo PlanRepo,
	planStore PlanStore,
	logStore LogStore,
	applyRepo ApplyRepo,
	mergeRepo MergeRepo,
	rebaseRepo RebaseRepo,
	changesetRepo ChangesetRepo,
	moduleRepo ModuleRepo,
	moduleVersionRepo ModuleVersionRepo,
	viewResourceRepo ViewResourceRepo,
	queryParser ViewQueryParser,
	transactionManager TransactionManager,
	newExecutor NewExecutor,
) versource.Facade {
	runApply := NewRunApply(config, applyRepo, stateRepo, stateResourceRepo, resourceRepo, planStore, logStore, transactionManager, newExecutor, componentRepo)
	runPlan := NewRunPlan(config, planRepo, planStore, logStore, transactionManager, newExecutor, componentRepo)
	listComponentChanges := NewListComponentChanges(componentChangeRepo, transactionManager)
	applyWorker := NewApplyWorker(runApply, applyRepo, transactionManager)
	runMerge := NewRunMerge(config, mergeRepo, changesetRepo, planRepo, planStore, logStore, transactionManager, listComponentChanges, componentChangeRepo, applyRepo, applyWorker)
	planWorker := NewPlanWorker(runPlan, planRepo, transactionManager)
	createPlan := NewCreatePlan(componentRepo, componentChangeRepo, planRepo, changesetRepo, transactionManager, planWorker)
	runRebase := NewRunRebase(config, rebaseRepo, changesetRepo, transactionManager, listComponentChanges, createPlan)
	mergeWorker := NewMergeWorker(runMerge, mergeRepo, transactionManager)
	rebaseWorker := NewRebaseWorker(runRebase, rebaseRepo, transactionManager)
	getMerge := NewGetMerge(mergeRepo, transactionManager)
	listMerges := NewListMerges(mergeRepo, transactionManager)
	createMerge := NewCreateMerge(changesetRepo, mergeRepo, transactionManager, mergeWorker)
	getRebase := NewGetRebase(rebaseRepo, transactionManager)
	listRebases := NewListRebases(rebaseRepo, transactionManager)
	createRebase := NewCreateRebase(changesetRepo, rebaseRepo, transactionManager, rebaseWorker)
	getPlan := NewGetPlan(planRepo, componentRepo, transactionManager)
	getPlanLog := NewGetPlanLog(logStore, transactionManager)
	getApply := NewGetApply(applyRepo, componentRepo, transactionManager)
	getApplyLog := NewGetApplyLog(logStore)
	ensureChangeset := NewEnsureChangeset(changesetRepo, transactionManager)

	return &facade{
		getModule:            NewGetModule(moduleRepo, moduleVersionRepo, transactionManager),
		listModules:          NewListModules(moduleRepo, transactionManager),
		createModule:         NewCreateModule(moduleRepo, moduleVersionRepo, transactionManager),
		updateModule:         NewUpdateModule(moduleRepo, moduleVersionRepo, transactionManager),
		deleteModule:         NewDeleteModule(moduleRepo, componentRepo, transactionManager),
		getModuleVersion:     NewGetModuleVersion(moduleVersionRepo, transactionManager),
		listModuleVersions:   NewListModuleVersions(moduleVersionRepo, transactionManager),
		listChangesets:       NewListChangesets(changesetRepo, transactionManager),
		createChangeset:      NewCreateChangeset(changesetRepo, transactionManager),
		deleteChangeset:      NewDeleteChangeset(changesetRepo, planRepo, applyRepo, planStore, logStore, transactionManager),
		ensureChangeset:      ensureChangeset,
		getMerge:             getMerge,
		listMerges:           listMerges,
		createMerge:          createMerge,
		getRebase:            getRebase,
		listRebases:          listRebases,
		createRebase:         createRebase,
		getComponent:         NewGetComponent(componentRepo, transactionManager),
		listComponents:       NewListComponents(componentRepo, transactionManager),
		getComponentChange:   NewGetComponentChange(componentChangeRepo, transactionManager),
		listComponentChanges: listComponentChanges,
		createComponent:      NewCreateComponent(componentRepo, moduleRepo, moduleVersionRepo, ensureChangeset, createPlan, transactionManager),
		updateComponent:      NewUpdateComponent(componentRepo, moduleVersionRepo, changesetRepo, ensureChangeset, createPlan, transactionManager),
		deleteComponent:      NewDeleteComponent(componentRepo, componentChangeRepo, changesetRepo, ensureChangeset, createPlan, transactionManager),
		restoreComponent:     NewRestoreComponent(componentRepo, componentChangeRepo, changesetRepo, ensureChangeset, createPlan, transactionManager),
		getPlan:              getPlan,
		getPlanLog:           getPlanLog,
		listPlans:            NewListPlans(planRepo, transactionManager),
		createPlan:           createPlan,
		runPlan:              runPlan,
		getApply:             getApply,
		getApplyLog:          getApplyLog,
		listApplies:          NewListApplies(applyRepo, transactionManager),
		runApply:             runApply,
		listResources:        NewListResources(resourceRepo, transactionManager),
		getViewResource:      NewGetViewResource(viewResourceRepo, transactionManager),
		listViewResources:    NewListViewResources(viewResourceRepo, transactionManager),
		saveViewResource:     NewSaveViewResource(viewResourceRepo, queryParser, transactionManager),
		deleteViewResource:   NewDeleteViewResource(viewResourceRepo, transactionManager),
		planWorker:           planWorker,
		applyWorker:          applyWorker,
		mergeWorker:          mergeWorker,
		rebaseWorker:         rebaseWorker,
	}
}

func (f *facade) GetModule(ctx context.Context, req versource.GetModuleRequest) (*versource.GetModuleResponse, error) {
	return f.getModule.Exec(ctx, req)
}

func (f *facade) ListModules(ctx context.Context, req versource.ListModulesRequest) (*versource.ListModulesResponse, error) {
	return f.listModules.Exec(ctx, req)
}

func (f *facade) CreateModule(ctx context.Context, req versource.CreateModuleRequest) (*versource.CreateModuleResponse, error) {
	return f.createModule.Exec(ctx, req)
}

func (f *facade) UpdateModule(ctx context.Context, req versource.UpdateModuleRequest) (*versource.UpdateModuleResponse, error) {
	return f.updateModule.Exec(ctx, req)
}

func (f *facade) DeleteModule(ctx context.Context, req versource.DeleteModuleRequest) (*versource.DeleteModuleResponse, error) {
	return f.deleteModule.Exec(ctx, req)
}

func (f *facade) GetModuleVersion(ctx context.Context, req versource.GetModuleVersionRequest) (*versource.GetModuleVersionResponse, error) {
	return f.getModuleVersion.Exec(ctx, req)
}

func (f *facade) ListModuleVersions(ctx context.Context, req versource.ListModuleVersionsRequest) (*versource.ListModuleVersionsResponse, error) {
	return f.listModuleVersions.Exec(ctx, req)
}

func (f *facade) ListChangesets(ctx context.Context, req versource.ListChangesetsRequest) (*versource.ListChangesetsResponse, error) {
	return f.listChangesets.Exec(ctx, req)
}

func (f *facade) CreateChangeset(ctx context.Context, req versource.CreateChangesetRequest) (*versource.CreateChangesetResponse, error) {
	return f.createChangeset.Exec(ctx, req)
}

func (f *facade) DeleteChangeset(ctx context.Context, req versource.DeleteChangesetRequest) (*versource.DeleteChangesetResponse, error) {
	return f.deleteChangeset.Exec(ctx, req)
}

func (f *facade) EnsureChangeset(ctx context.Context, req versource.EnsureChangesetRequest) (*versource.EnsureChangesetResponse, error) {
	return f.ensureChangeset.Exec(ctx, req)
}

func (f *facade) GetMerge(ctx context.Context, req versource.GetMergeRequest) (*versource.GetMergeResponse, error) {
	return f.getMerge.Exec(ctx, req)
}

func (f *facade) ListMerges(ctx context.Context, req versource.ListMergesRequest) (*versource.ListMergesResponse, error) {
	return f.listMerges.Exec(ctx, req)
}

func (f *facade) CreateMerge(ctx context.Context, req versource.CreateMergeRequest) (*versource.CreateMergeResponse, error) {
	return f.createMerge.Exec(ctx, req)
}

func (f *facade) GetRebase(ctx context.Context, req versource.GetRebaseRequest) (*versource.GetRebaseResponse, error) {
	return f.getRebase.Exec(ctx, req)
}

func (f *facade) ListRebases(ctx context.Context, req versource.ListRebasesRequest) (*versource.ListRebasesResponse, error) {
	return f.listRebases.Exec(ctx, req)
}

func (f *facade) CreateRebase(ctx context.Context, req versource.CreateRebaseRequest) (*versource.CreateRebaseResponse, error) {
	return f.createRebase.Exec(ctx, req)
}

func (f *facade) GetComponent(ctx context.Context, req versource.GetComponentRequest) (*versource.GetComponentResponse, error) {
	return f.getComponent.Exec(ctx, req)
}

func (f *facade) ListComponents(ctx context.Context, req versource.ListComponentsRequest) (*versource.ListComponentsResponse, error) {
	return f.listComponents.Exec(ctx, req)
}

func (f *facade) GetComponentChange(ctx context.Context, req versource.GetComponentChangeRequest) (*versource.GetComponentChangeResponse, error) {
	return f.getComponentChange.Exec(ctx, req)
}

func (f *facade) ListComponentChanges(ctx context.Context, req versource.ListComponentChangesRequest) (*versource.ListComponentChangesResponse, error) {
	return f.listComponentChanges.Exec(ctx, req)
}

func (f *facade) CreateComponent(ctx context.Context, req versource.CreateComponentRequest) (*versource.CreateComponentResponse, error) {
	return f.createComponent.Exec(ctx, req)
}

func (f *facade) UpdateComponent(ctx context.Context, req versource.UpdateComponentRequest) (*versource.UpdateComponentResponse, error) {
	return f.updateComponent.Exec(ctx, req)
}

func (f *facade) DeleteComponent(ctx context.Context, req versource.DeleteComponentRequest) (*versource.DeleteComponentResponse, error) {
	return f.deleteComponent.Exec(ctx, req)
}

func (f *facade) RestoreComponent(ctx context.Context, req versource.RestoreComponentRequest) (*versource.RestoreComponentResponse, error) {
	return f.restoreComponent.Exec(ctx, req)
}

func (f *facade) GetPlan(ctx context.Context, req versource.GetPlanRequest) (*versource.GetPlanResponse, error) {
	return f.getPlan.Exec(ctx, req)
}

func (f *facade) GetPlanLog(ctx context.Context, req versource.GetPlanLogRequest) (*versource.GetPlanLogResponse, error) {
	return f.getPlanLog.Exec(ctx, req)
}

func (f *facade) ListPlans(ctx context.Context, req versource.ListPlansRequest) (*versource.ListPlansResponse, error) {
	return f.listPlans.Exec(ctx, req)
}

func (f *facade) CreatePlan(ctx context.Context, req versource.CreatePlanRequest) (*versource.CreatePlanResponse, error) {
	return f.createPlan.Exec(ctx, req)
}

func (f *facade) RunPlan(ctx context.Context, planID uint) error {
	return f.runPlan.Exec(ctx, planID)
}

func (f *facade) GetApply(ctx context.Context, req versource.GetApplyRequest) (*versource.GetApplyResponse, error) {
	return f.getApply.Exec(ctx, req)
}

func (f *facade) GetApplyLog(ctx context.Context, req versource.GetApplyLogRequest) (*versource.GetApplyLogResponse, error) {
	return f.getApplyLog.Exec(ctx, req)
}

func (f *facade) ListApplies(ctx context.Context, req versource.ListAppliesRequest) (*versource.ListAppliesResponse, error) {
	return f.listApplies.Exec(ctx, req)
}

func (f *facade) RunApply(ctx context.Context, applyID uint) error {
	return f.runApply.Exec(ctx, applyID)
}

func (f *facade) ListResources(ctx context.Context, req versource.ListResourcesRequest) (*versource.ListResourcesResponse, error) {
	return f.listResources.Exec(ctx, req)
}

func (f *facade) GetViewResource(ctx context.Context, req versource.GetViewResourceRequest) (*versource.GetViewResourceResponse, error) {
	return f.getViewResource.Exec(ctx, req)
}

func (f *facade) ListViewResources(ctx context.Context, req versource.ListViewResourcesRequest) (*versource.ListViewResourcesResponse, error) {
	return f.listViewResources.Exec(ctx, req)
}

func (f *facade) SaveViewResource(ctx context.Context, req versource.SaveViewResourceRequest) (*versource.SaveViewResourceResponse, error) {
	return f.saveViewResource.Exec(ctx, req)
}

func (f *facade) DeleteViewResource(ctx context.Context, req versource.DeleteViewResourceRequest) (*versource.DeleteViewResourceResponse, error) {
	return f.deleteViewResource.Exec(ctx, req)
}

func (f *facade) Start(ctx context.Context) {
	f.planWorker.Start(ctx)
	f.applyWorker.Start(ctx)
	f.mergeWorker.Start(ctx)
	f.rebaseWorker.Start(ctx)
}
