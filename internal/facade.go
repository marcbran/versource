package internal

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
	EnsureChangeset(ctx context.Context, req EnsureChangesetRequest) (*EnsureChangesetResponse, error)

	GetMerge(ctx context.Context, req GetMergeRequest) (*GetMergeResponse, error)
	ListMerges(ctx context.Context, req ListMergesRequest) (*ListMergesResponse, error)
	CreateMerge(ctx context.Context, req CreateMergeRequest) (*CreateMergeResponse, error)

	GetRebase(ctx context.Context, req GetRebaseRequest) (*GetRebaseResponse, error)
	ListRebases(ctx context.Context, req ListRebasesRequest) (*ListRebasesResponse, error)
	CreateRebase(ctx context.Context, req CreateRebaseRequest) (*CreateRebaseResponse, error)

	GetComponent(ctx context.Context, req GetComponentRequest) (*GetComponentResponse, error)
	ListComponents(ctx context.Context, req ListComponentsRequest) (*ListComponentsResponse, error)
	GetComponentDiff(ctx context.Context, req GetComponentDiffRequest) (*GetComponentDiffResponse, error)
	ListComponentDiffs(ctx context.Context, req ListComponentDiffsRequest) (*ListComponentDiffsResponse, error)
	CreateComponent(ctx context.Context, req CreateComponentRequest) (*CreateComponentResponse, error)
	UpdateComponent(ctx context.Context, req UpdateComponentRequest) (*UpdateComponentResponse, error)
	DeleteComponent(ctx context.Context, req DeleteComponentRequest) (*DeleteComponentResponse, error)
	RestoreComponent(ctx context.Context, req RestoreComponentRequest) (*RestoreComponentResponse, error)

	GetPlan(ctx context.Context, req GetPlanRequest) (*GetPlanResponse, error)
	GetPlanLog(ctx context.Context, req GetPlanLogRequest) (*GetPlanLogResponse, error)
	ListPlans(ctx context.Context, req ListPlansRequest) (*ListPlansResponse, error)
	CreatePlan(ctx context.Context, req CreatePlanRequest) (*CreatePlanResponse, error)
	RunPlan(ctx context.Context, req RunPlanRequest) error

	GetApplyLog(ctx context.Context, req GetApplyLogRequest) (*GetApplyLogResponse, error)
	ListApplies(ctx context.Context, req ListAppliesRequest) (*ListAppliesResponse, error)
	RunApply(ctx context.Context, applyID uint) error
}

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
	ensureChangeset *EnsureChangeset

	getMerge    *GetMerge
	listMerges  *ListMerges
	createMerge *CreateMerge

	getRebase    *GetRebase
	listRebases  *ListRebases
	createRebase *CreateRebase

	getComponent       *GetComponent
	listComponents     *ListComponents
	getComponentDiff   *GetComponentDiff
	listComponentDiffs *ListComponentDiffs
	createComponent    *CreateComponent
	updateComponent    *UpdateComponent
	deleteComponent    *DeleteComponent
	restoreComponent   *RestoreComponent

	getPlan    *GetPlan
	getPlanLog *GetPlanLog
	listPlans  *ListPlans
	createPlan *CreatePlan
	runPlan    *RunPlan

	getApplyLog *GetApplyLog
	listApplies *ListApplies
	runApply    *RunApply

	planWorker   *PlanWorker
	applyWorker  *ApplyWorker
	mergeWorker  *MergeWorker
	rebaseWorker *RebaseWorker
}

func NewFacade(
	config *Config,
	componentRepo ComponentRepo,
	componentDiffRepo ComponentDiffRepo,
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
	transactionManager TransactionManager,
	newExecutor NewExecutor,
) Facade {
	runApply := NewRunApply(config, applyRepo, stateRepo, stateResourceRepo, resourceRepo, planStore, logStore, transactionManager, newExecutor)
	runPlan := NewRunPlan(config, planRepo, planStore, logStore, applyRepo, transactionManager, newExecutor)
	listComponentDiffs := NewListComponentDiffs(componentDiffRepo, transactionManager)
	runMerge := NewRunMerge(config, mergeRepo, changesetRepo, planRepo, planStore, logStore, transactionManager, listComponentDiffs, componentDiffRepo)
	runRebase := NewRunRebase(config, rebaseRepo, changesetRepo, transactionManager)
	applyWorker := NewApplyWorker(runApply, applyRepo)
	planWorker := NewPlanWorker(runPlan, planRepo)
	mergeWorker := NewMergeWorker(runMerge, mergeRepo)
	rebaseWorker := NewRebaseWorker(runRebase, rebaseRepo)
	createPlan := NewCreatePlan(componentRepo, planRepo, changesetRepo, transactionManager, planWorker)
	getMerge := NewGetMerge(mergeRepo, transactionManager)
	listMerges := NewListMerges(mergeRepo, transactionManager)
	createMerge := NewCreateMerge(changesetRepo, mergeRepo, transactionManager, mergeWorker)
	getRebase := NewGetRebase(rebaseRepo, transactionManager)
	listRebases := NewListRebases(rebaseRepo, transactionManager)
	createRebase := NewCreateRebase(changesetRepo, rebaseRepo, transactionManager, rebaseWorker)
	getPlan := NewGetPlan(planRepo, transactionManager)
	getPlanLog := NewGetPlanLog(logStore, transactionManager)
	getApplyLog := NewGetApplyLog(logStore)
	ensureChangeset := NewEnsureChangeset(changesetRepo, transactionManager)

	return &facade{
		getModule:          NewGetModule(moduleRepo, moduleVersionRepo, transactionManager),
		listModules:        NewListModules(moduleRepo, transactionManager),
		createModule:       NewCreateModule(moduleRepo, moduleVersionRepo, transactionManager),
		updateModule:       NewUpdateModule(moduleRepo, moduleVersionRepo, transactionManager),
		deleteModule:       NewDeleteModule(moduleRepo, componentRepo, transactionManager),
		getModuleVersion:   NewGetModuleVersion(moduleVersionRepo, transactionManager),
		listModuleVersions: NewListModuleVersions(moduleVersionRepo, transactionManager),
		listChangesets:     NewListChangesets(changesetRepo, transactionManager),
		createChangeset:    NewCreateChangeset(changesetRepo, transactionManager),
		ensureChangeset:    ensureChangeset,
		getMerge:           getMerge,
		listMerges:         listMerges,
		createMerge:        createMerge,
		getRebase:          getRebase,
		listRebases:        listRebases,
		createRebase:       createRebase,
		getComponent:       NewGetComponent(componentRepo, transactionManager),
		listComponents:     NewListComponents(componentRepo, transactionManager),
		getComponentDiff:   NewGetComponentDiff(componentDiffRepo, transactionManager),
		listComponentDiffs: listComponentDiffs,
		createComponent:    NewCreateComponent(componentRepo, moduleRepo, moduleVersionRepo, ensureChangeset, createPlan, transactionManager),
		updateComponent:    NewUpdateComponent(componentRepo, moduleVersionRepo, changesetRepo, ensureChangeset, createPlan, transactionManager),
		deleteComponent:    NewDeleteComponent(componentRepo, componentDiffRepo, changesetRepo, ensureChangeset, createPlan, transactionManager),
		restoreComponent:   NewRestoreComponent(componentRepo, componentDiffRepo, changesetRepo, ensureChangeset, createPlan, transactionManager),
		getPlan:            getPlan,
		getPlanLog:         getPlanLog,
		listPlans:          NewListPlans(planRepo, transactionManager),
		createPlan:         createPlan,
		runPlan:            runPlan,
		getApplyLog:        getApplyLog,
		listApplies:        NewListApplies(applyRepo, transactionManager),
		runApply:           runApply,
		planWorker:         planWorker,
		applyWorker:        applyWorker,
		mergeWorker:        mergeWorker,
		rebaseWorker:       rebaseWorker,
	}
}

func (f *facade) GetModule(ctx context.Context, req GetModuleRequest) (*GetModuleResponse, error) {
	return f.getModule.Exec(ctx, req)
}

func (f *facade) ListModules(ctx context.Context, req ListModulesRequest) (*ListModulesResponse, error) {
	return f.listModules.Exec(ctx, req)
}

func (f *facade) CreateModule(ctx context.Context, req CreateModuleRequest) (*CreateModuleResponse, error) {
	return f.createModule.Exec(ctx, req)
}

func (f *facade) UpdateModule(ctx context.Context, req UpdateModuleRequest) (*UpdateModuleResponse, error) {
	return f.updateModule.Exec(ctx, req)
}

func (f *facade) DeleteModule(ctx context.Context, req DeleteModuleRequest) (*DeleteModuleResponse, error) {
	return f.deleteModule.Exec(ctx, req)
}

func (f *facade) GetModuleVersion(ctx context.Context, req GetModuleVersionRequest) (*GetModuleVersionResponse, error) {
	return f.getModuleVersion.Exec(ctx, req)
}

func (f *facade) ListModuleVersions(ctx context.Context, req ListModuleVersionsRequest) (*ListModuleVersionsResponse, error) {
	return f.listModuleVersions.Exec(ctx, req)
}

func (f *facade) ListChangesets(ctx context.Context, req ListChangesetsRequest) (*ListChangesetsResponse, error) {
	return f.listChangesets.Exec(ctx, req)
}

func (f *facade) CreateChangeset(ctx context.Context, req CreateChangesetRequest) (*CreateChangesetResponse, error) {
	return f.createChangeset.Exec(ctx, req)
}

func (f *facade) EnsureChangeset(ctx context.Context, req EnsureChangesetRequest) (*EnsureChangesetResponse, error) {
	return f.ensureChangeset.Exec(ctx, req)
}

func (f *facade) GetMerge(ctx context.Context, req GetMergeRequest) (*GetMergeResponse, error) {
	return f.getMerge.Exec(ctx, req)
}

func (f *facade) ListMerges(ctx context.Context, req ListMergesRequest) (*ListMergesResponse, error) {
	return f.listMerges.Exec(ctx, req)
}

func (f *facade) CreateMerge(ctx context.Context, req CreateMergeRequest) (*CreateMergeResponse, error) {
	return f.createMerge.Exec(ctx, req)
}

func (f *facade) GetRebase(ctx context.Context, req GetRebaseRequest) (*GetRebaseResponse, error) {
	return f.getRebase.Exec(ctx, req)
}

func (f *facade) ListRebases(ctx context.Context, req ListRebasesRequest) (*ListRebasesResponse, error) {
	return f.listRebases.Exec(ctx, req)
}

func (f *facade) CreateRebase(ctx context.Context, req CreateRebaseRequest) (*CreateRebaseResponse, error) {
	return f.createRebase.Exec(ctx, req)
}

func (f *facade) GetComponent(ctx context.Context, req GetComponentRequest) (*GetComponentResponse, error) {
	return f.getComponent.Exec(ctx, req)
}

func (f *facade) ListComponents(ctx context.Context, req ListComponentsRequest) (*ListComponentsResponse, error) {
	return f.listComponents.Exec(ctx, req)
}

func (f *facade) GetComponentDiff(ctx context.Context, req GetComponentDiffRequest) (*GetComponentDiffResponse, error) {
	return f.getComponentDiff.Exec(ctx, req)
}

func (f *facade) ListComponentDiffs(ctx context.Context, req ListComponentDiffsRequest) (*ListComponentDiffsResponse, error) {
	return f.listComponentDiffs.Exec(ctx, req)
}

func (f *facade) CreateComponent(ctx context.Context, req CreateComponentRequest) (*CreateComponentResponse, error) {
	return f.createComponent.Exec(ctx, req)
}

func (f *facade) UpdateComponent(ctx context.Context, req UpdateComponentRequest) (*UpdateComponentResponse, error) {
	return f.updateComponent.Exec(ctx, req)
}

func (f *facade) DeleteComponent(ctx context.Context, req DeleteComponentRequest) (*DeleteComponentResponse, error) {
	return f.deleteComponent.Exec(ctx, req)
}

func (f *facade) RestoreComponent(ctx context.Context, req RestoreComponentRequest) (*RestoreComponentResponse, error) {
	return f.restoreComponent.Exec(ctx, req)
}

func (f *facade) GetPlan(ctx context.Context, req GetPlanRequest) (*GetPlanResponse, error) {
	return f.getPlan.Exec(ctx, req)
}

func (f *facade) GetPlanLog(ctx context.Context, req GetPlanLogRequest) (*GetPlanLogResponse, error) {
	return f.getPlanLog.Exec(ctx, req)
}

func (f *facade) ListPlans(ctx context.Context, req ListPlansRequest) (*ListPlansResponse, error) {
	return f.listPlans.Exec(ctx, req)
}

func (f *facade) CreatePlan(ctx context.Context, req CreatePlanRequest) (*CreatePlanResponse, error) {
	return f.createPlan.Exec(ctx, req)
}

func (f *facade) RunPlan(ctx context.Context, req RunPlanRequest) error {
	return f.runPlan.Exec(ctx, req)
}

func (f *facade) GetApplyLog(ctx context.Context, req GetApplyLogRequest) (*GetApplyLogResponse, error) {
	return f.getApplyLog.Exec(ctx, req)
}

func (f *facade) ListApplies(ctx context.Context, req ListAppliesRequest) (*ListAppliesResponse, error) {
	return f.listApplies.Exec(ctx, req)
}

func (f *facade) RunApply(ctx context.Context, applyID uint) error {
	return f.runApply.Exec(ctx, applyID)
}

func (f *facade) Start(ctx context.Context) {
	f.planWorker.Start(ctx)
	f.applyWorker.Start(ctx)
	f.mergeWorker.Start(ctx)
	f.rebaseWorker.Start(ctx)
}
