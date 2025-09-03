package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/database"
	"github.com/marcbran/versource/internal/infra/tfexec"
	"github.com/marcbran/versource/internal/store/file"
	log "github.com/sirupsen/logrus"
)

func Serve(ctx context.Context, config *internal.Config) error {
	server, err := NewServer(config)
	if err != nil {
		return err
	}

	server.planWorker.Start(ctx)
	server.applyWorker.Start(ctx)

	addr := config.HTTP.Hostname + ":" + config.HTTP.Port
	httpServer := &http.Server{
		Addr:    addr,
		Handler: server.router,
	}

	go func() {
		log.WithField("addr", addr).Info("Starting HTTP server")
		err := httpServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.WithError(err).Error("Server error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		log.Info("Shutting down server...")
	case <-ctx.Done():
		log.Info("Context cancelled, shutting down server...")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return httpServer.Shutdown(shutdownCtx)
}

type Server struct {
	config                      *internal.Config
	router                      *chi.Mux
	listModules                 *internal.ListModules
	createModule                *internal.CreateModule
	updateModule                *internal.UpdateModule
	deleteModule                *internal.DeleteModule
	listModuleVersions          *internal.ListModuleVersions
	listModuleVersionsForModule *internal.ListModuleVersionsForModule
	listChangesets              *internal.ListChangesets
	createChangeset             *internal.CreateChangeset
	mergeChangeset              *internal.MergeChangeset
	listComponents              *internal.ListComponents
	createComponent             *internal.CreateComponent
	updateComponent             *internal.UpdateComponent
	listComponentDiffs          *internal.ListComponentDiffs
	listPlans                   *internal.ListPlans
	createPlan                  *internal.CreatePlan
	runPlan                     *internal.RunPlan
	runApply                    *internal.RunApply
	listApplies                 *internal.ListApplies
	getPlanLog                  *internal.GetPlanLog
	getApplyLog                 *internal.GetApplyLog
	planWorker                  *internal.PlanWorker
	applyWorker                 *internal.ApplyWorker
}

func NewServer(config *internal.Config) (*Server, error) {
	db, err := database.NewGormDb(config.Database)
	if err != nil {
		return nil, err
	}

	componentRepo := database.NewGormComponentRepo(db)
	componentDiffRepo := database.NewGormComponentDiffRepo(db)
	stateRepo := database.NewGormStateRepo(db)
	resourceRepo := database.NewGormResourceRepo(db)
	planRepo := database.NewGormPlanRepo(db)
	planStore := file.NewPlanStore(config.Terraform.WorkDir)
	logStore := file.NewLogStore(config.Terraform.WorkDir)
	applyRepo := database.NewGormApplyRepo(db)
	changesetRepo := database.NewGormChangesetRepo(db)
	moduleRepo := database.NewGormModuleRepo(db)
	moduleVersionRepo := database.NewGormModuleVersionRepo(db)
	transactionManager := database.NewGormTransactionManager(db)

	newExecutor := tfexec.NewExecutor

	runApply := internal.NewRunApply(config, applyRepo, stateRepo, resourceRepo, planStore, logStore, transactionManager, newExecutor)
	applyWorker := internal.NewApplyWorker(runApply, applyRepo)
	runPlan := internal.NewRunPlan(config, planRepo, planStore, logStore, applyRepo, transactionManager, newExecutor)
	planWorker := internal.NewPlanWorker(runPlan, planRepo)
	createPlan := internal.NewCreatePlan(componentRepo, planRepo, changesetRepo, transactionManager, planWorker)
	getPlanLog := internal.NewGetPlanLog(logStore)
	getApplyLog := internal.NewGetApplyLog(logStore)
	ensureChangeset := internal.NewEnsureChangeset(changesetRepo, transactionManager)

	s := &Server{
		config:                      config,
		router:                      chi.NewRouter(),
		listModules:                 internal.NewListModules(moduleRepo, transactionManager),
		createModule:                internal.NewCreateModule(moduleRepo, moduleVersionRepo, transactionManager),
		updateModule:                internal.NewUpdateModule(moduleRepo, moduleVersionRepo, transactionManager),
		deleteModule:                internal.NewDeleteModule(moduleRepo, componentRepo, transactionManager),
		listModuleVersions:          internal.NewListModuleVersions(moduleVersionRepo, transactionManager),
		listModuleVersionsForModule: internal.NewListModuleVersionsForModule(moduleVersionRepo, transactionManager),
		listChangesets:              internal.NewListChangesets(changesetRepo, transactionManager),
		createChangeset:             internal.NewCreateChangeset(changesetRepo, transactionManager),
		mergeChangeset:              internal.NewMergeChangeset(changesetRepo, applyRepo, applyWorker, transactionManager),
		listComponents:              internal.NewListComponents(componentRepo, transactionManager),
		createComponent:             internal.NewCreateComponent(componentRepo, moduleRepo, moduleVersionRepo, ensureChangeset, createPlan, transactionManager),
		updateComponent:             internal.NewUpdateComponent(componentRepo, moduleVersionRepo, ensureChangeset, transactionManager),
		listComponentDiffs:          internal.NewListComponentDiffs(componentDiffRepo, transactionManager),
		listPlans:                   internal.NewListPlans(planRepo, transactionManager),
		createPlan:                  createPlan,
		runPlan:                     runPlan,
		runApply:                    runApply,
		listApplies:                 internal.NewListApplies(applyRepo, transactionManager),
		getPlanLog:                  getPlanLog,
		getApplyLog:                 getApplyLog,
		planWorker:                  planWorker,
		applyWorker:                 applyWorker,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s, nil
}

func (s *Server) setupMiddleware() {
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.RequestID)
}

func (s *Server) setupRoutes() {
	s.router.Route("/api/v1", func(r chi.Router) {
		r.Get("/modules", s.handleListModules)
		r.Get("/module-versions", s.handleListModuleVersions)
		r.Get("/components", s.handleListComponents)
		r.Get("/plans", s.handleListPlans)
		r.Route("/plans/{planID}", func(r chi.Router) {
			r.Get("/logs", s.handleGetPlanLog)
		})
		r.Get("/applies", s.handleListApplies)
		r.Get("/changesets", s.handleListChangesets)
		r.Route("/applies/{applyID}", func(r chi.Router) {
			r.Get("/logs", s.handleGetApplyLog)
		})
		r.Post("/changesets", s.handleCreateChangeset)
		r.Post("/modules", s.handleCreateModule)
		r.Put("/modules/{moduleID}", s.handleUpdateModule)
		r.Delete("/modules/{moduleID}", s.handleDeleteModule)
		r.Get("/modules/{moduleID}/versions", s.handleListModuleVersionsForModule)
		r.Route("/changesets/{changesetName}", func(r chi.Router) {
			r.Get("/components", s.handleListComponents)
			r.Post("/components", s.handleCreateComponent)
			r.Get("/components/diffs", s.handleListComponentDiffs)
			r.Get("/plans", s.handleListPlans)
			r.Route("/components/{componentID}", func(r chi.Router) {
				r.Patch("/", s.handleUpdateComponent)
				r.Post("/plans", s.handleCreatePlan)
			})
			r.Post("/merge", s.handleMergeChangeset)
		})
	})
}

type ErrorResponse struct {
	Message string `json:"message"`
}

func returnSuccess(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	returnJSON(w, data)
}

func returnCreated(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	returnJSON(w, data)
}

func returnBadRequest(w http.ResponseWriter, err error) {
	log.WithError(err).Warn("Bad request error")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	returnJSON(w, ErrorResponse{
		Message: err.Error(),
	})
}

func returnInternalServerError(w http.ResponseWriter, err error) {
	log.WithError(err).Error("Internal server error")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	returnJSON(w, ErrorResponse{
		Message: err.Error(),
	})
}

func returnError(w http.ResponseWriter, err error) {
	if internal.IsUserError(err) {
		returnBadRequest(w, err)
		return
	}
	returnInternalServerError(w, err)
}

func returnJSON(w http.ResponseWriter, data any) {
	encodeErr := json.NewEncoder(w).Encode(data)
	if encodeErr != nil {
		log.WithError(encodeErr).Warn("Failed to encode JSON response")
	}
}
