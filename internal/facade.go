package internal

import (
	"context"
)

type Facade interface {
	GetModule(ctx context.Context, req GetModuleRequest) (*GetModuleResponse, error)
	ListModules(ctx context.Context, req ListModulesRequest) (*ListModulesResponse, error)
	CreateModule(ctx context.Context, req CreateModuleRequest) (*CreateModuleResponse, error)
	UpdateModule(ctx context.Context, req UpdateModuleRequest) (*UpdateModuleResponse, error)
	DeleteModule(ctx context.Context, req DeleteModuleRequest) (*DeleteModuleResponse, error)
	GetModuleVersion(ctx context.Context, req GetModuleVersionRequest) (*GetModuleVersionResponse, error)
	ListModuleVersions(ctx context.Context, req ListModuleVersionsRequest) (*ListModuleVersionsResponse, error)
	ListModuleVersionsForModule(ctx context.Context, req ListModuleVersionsForModuleRequest) (*ListModuleVersionsForModuleResponse, error)

	ListChangesets(ctx context.Context, req ListChangesetsRequest) (*ListChangesetsResponse, error)
	CreateChangeset(ctx context.Context, req CreateChangesetRequest) (*CreateChangesetResponse, error)
	EnsureChangeset(ctx context.Context, req EnsureChangesetRequest) (*EnsureChangesetResponse, error)
	MergeChangeset(ctx context.Context, req MergeChangesetRequest) (*MergeChangesetResponse, error)

	GetComponent(ctx context.Context, req GetComponentRequest) (*GetComponentResponse, error)
	ListComponents(ctx context.Context, req ListComponentsRequest) (*ListComponentsResponse, error)
	ListComponentDiffs(ctx context.Context, req ListComponentDiffsRequest) (*ListComponentDiffsResponse, error)
	CreateComponent(ctx context.Context, req CreateComponentRequest) (*CreateComponentResponse, error)
	UpdateComponent(ctx context.Context, req UpdateComponentRequest) (*UpdateComponentResponse, error)
	DeleteComponent(ctx context.Context, req DeleteComponentRequest) (*DeleteComponentResponse, error)
	RestoreComponent(ctx context.Context, req RestoreComponentRequest) (*RestoreComponentResponse, error)

	GetPlanLog(ctx context.Context, req GetPlanLogRequest) (*GetPlanLogResponse, error)
	ListPlans(ctx context.Context, req ListPlansRequest) (*ListPlansResponse, error)
	CreatePlan(ctx context.Context, req CreatePlanRequest) (*CreatePlanResponse, error)
	RunPlan(ctx context.Context, req RunPlanRequest) error

	GetApplyLog(ctx context.Context, req GetApplyLogRequest) (*GetApplyLogResponse, error)
	ListApplies(ctx context.Context, req ListAppliesRequest) (*ListAppliesResponse, error)
	RunApply(ctx context.Context, applyID uint) error
}

type facade struct {
	getModule                   *GetModule
	listModules                 *ListModules
	createModule                *CreateModule
	updateModule                *UpdateModule
	deleteModule                *DeleteModule
	getModuleVersion            *GetModuleVersion
	listModuleVersions          *ListModuleVersions
	listModuleVersionsForModule *ListModuleVersionsForModule

	listChangesets  *ListChangesets
	createChangeset *CreateChangeset
	ensureChangeset *EnsureChangeset
	mergeChangeset  *MergeChangeset

	getComponent       *GetComponent
	listComponents     *ListComponents
	listComponentDiffs *ListComponentDiffs
	createComponent    *CreateComponent
	updateComponent    *UpdateComponent
	deleteComponent    *DeleteComponent
	restoreComponent   *RestoreComponent

	getPlanLog *GetPlanLog
	listPlans  *ListPlans
	createPlan *CreatePlan
	runPlan    *RunPlan

	getApplyLog *GetApplyLog
	listApplies *ListApplies
	runApply    *RunApply
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
	changesetRepo ChangesetRepo,
	moduleRepo ModuleRepo,
	moduleVersionRepo ModuleVersionRepo,
	transactionManager TransactionManager,
	newExecutor NewExecutor,
) Facade {
	runApply := NewRunApply(config, applyRepo, stateRepo, stateResourceRepo, resourceRepo, planStore, logStore, transactionManager, newExecutor)
	applyWorker := NewApplyWorker(runApply, applyRepo)
	runPlan := NewRunPlan(config, planRepo, planStore, logStore, applyRepo, transactionManager, newExecutor)
	planWorker := NewPlanWorker(runPlan, planRepo)
	createPlan := NewCreatePlan(componentRepo, planRepo, changesetRepo, transactionManager, planWorker)
	getPlanLog := NewGetPlanLog(logStore)
	getApplyLog := NewGetApplyLog(logStore)
	ensureChangeset := NewEnsureChangeset(changesetRepo, transactionManager)

	return &facade{
		getModule:                   NewGetModule(moduleRepo, moduleVersionRepo, transactionManager),
		listModules:                 NewListModules(moduleRepo, transactionManager),
		createModule:                NewCreateModule(moduleRepo, moduleVersionRepo, transactionManager),
		updateModule:                NewUpdateModule(moduleRepo, moduleVersionRepo, transactionManager),
		deleteModule:                NewDeleteModule(moduleRepo, componentRepo, transactionManager),
		getModuleVersion:            NewGetModuleVersion(moduleVersionRepo, transactionManager),
		listModuleVersions:          NewListModuleVersions(moduleVersionRepo, transactionManager),
		listModuleVersionsForModule: NewListModuleVersionsForModule(moduleVersionRepo, transactionManager),
		listChangesets:              NewListChangesets(changesetRepo, transactionManager),
		createChangeset:             NewCreateChangeset(changesetRepo, transactionManager),
		ensureChangeset:             ensureChangeset,
		mergeChangeset:              NewMergeChangeset(changesetRepo, applyRepo, applyWorker, transactionManager),
		getComponent:                NewGetComponent(componentRepo, transactionManager),
		listComponents:              NewListComponents(componentRepo, transactionManager),
		listComponentDiffs:          NewListComponentDiffs(componentDiffRepo, transactionManager),
		createComponent:             NewCreateComponent(componentRepo, moduleRepo, moduleVersionRepo, ensureChangeset, createPlan, transactionManager),
		updateComponent:             NewUpdateComponent(componentRepo, moduleVersionRepo, ensureChangeset, createPlan, transactionManager),
		deleteComponent:             NewDeleteComponent(componentRepo, ensureChangeset, createPlan, transactionManager),
		restoreComponent:            NewRestoreComponent(componentRepo, ensureChangeset, createPlan, transactionManager),
		getPlanLog:                  getPlanLog,
		listPlans:                   NewListPlans(planRepo, transactionManager),
		createPlan:                  createPlan,
		runPlan:                     runPlan,
		getApplyLog:                 getApplyLog,
		listApplies:                 NewListApplies(applyRepo, transactionManager),
		runApply:                    runApply,
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

func (f *facade) ListModuleVersionsForModule(ctx context.Context, req ListModuleVersionsForModuleRequest) (*ListModuleVersionsForModuleResponse, error) {
	return f.listModuleVersionsForModule.Exec(ctx, req)
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

func (f *facade) MergeChangeset(ctx context.Context, req MergeChangesetRequest) (*MergeChangesetResponse, error) {
	return f.mergeChangeset.Exec(ctx, req)
}

func (f *facade) GetComponent(ctx context.Context, req GetComponentRequest) (*GetComponentResponse, error) {
	return f.getComponent.Exec(ctx, req)
}

func (f *facade) ListComponents(ctx context.Context, req ListComponentsRequest) (*ListComponentsResponse, error) {
	return f.listComponents.Exec(ctx, req)
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
