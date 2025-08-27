package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/database"
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
	config          *internal.Config
	router          *chi.Mux
	createComponent *internal.CreateComponent
	updateComponent *internal.UpdateComponent
	createPlan      *internal.CreatePlan
	createChangeset *internal.CreateChangeset
	mergeChangeset  *internal.MergeChangeset
	runPlan         *internal.RunPlan
	runApply        *internal.RunApply
	createModule    *internal.CreateModule
	planWorker      *internal.PlanWorker
	applyWorker     *internal.ApplyWorker
}

func NewServer(config *internal.Config) (*Server, error) {
	db, err := database.NewGormDb(config.Database)
	if err != nil {
		return nil, err
	}

	componentRepo := database.NewGormComponentRepo(db)
	stateRepo := database.NewGormStateRepo(db)
	planRepo := database.NewGormPlanRepo(db)
	planStore := file.NewPlanStore(config.Terraform.WorkDir)
	applyRepo := database.NewGormApplyRepo(db)
	changesetRepo := database.NewGormChangesetRepo(db)
	moduleRepo := database.NewGormModuleRepo(db)
	moduleVersionRepo := database.NewGormModuleVersionRepo(db)
	transactionManager := database.NewGormTransactionManager(db)

	runApply := internal.NewRunApply(config, applyRepo, stateRepo, planStore, transactionManager)
	applyWorker := internal.NewApplyWorker(runApply, applyRepo)

	runPlan := internal.NewRunPlan(config, planRepo, planStore, applyRepo, transactionManager)
	planWorker := internal.NewPlanWorker(runPlan, planRepo)
	createPlan := internal.NewCreatePlan(componentRepo, planRepo, changesetRepo, transactionManager, planWorker)
	ensureChangeset := internal.NewEnsureChangeset(changesetRepo, transactionManager)

	s := &Server{
		config:          config,
		router:          chi.NewRouter(),
		createComponent: internal.NewCreateComponent(componentRepo, ensureChangeset, createPlan, transactionManager),
		updateComponent: internal.NewUpdateComponent(componentRepo, ensureChangeset, transactionManager),
		createPlan:      createPlan,
		createChangeset: internal.NewCreateChangeset(changesetRepo, transactionManager),
		mergeChangeset:  internal.NewMergeChangeset(changesetRepo, applyRepo, applyWorker, transactionManager),
		runPlan:         runPlan,
		runApply:        runApply,
		createModule:    internal.NewCreateModule(moduleRepo, moduleVersionRepo, transactionManager),
		planWorker:      planWorker,
		applyWorker:     applyWorker,
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
		r.Post("/changesets", s.handleCreateChangeset)
		r.Post("/modules", s.handleCreateModule)
		r.Route("/changesets/{changesetName}", func(r chi.Router) {
			r.Post("/components", s.handleCreateComponent)
			r.Route("/components/{componentID}", func(r chi.Router) {
				r.Patch("/", s.handleUpdateComponent)
				r.Post("/plans", s.handleCreatePlan)
			})
			r.Post("/merge", s.handleMergeChangeset)
		})
	})
}

func (s *Server) handleCreateChangeset(w http.ResponseWriter, r *http.Request) {
	var req internal.CreateChangesetRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		returnBadRequest(w, fmt.Errorf("invalid request body"))
		return
	}

	resp, err := s.createChangeset.Exec(r.Context(), req)
	if err != nil {
		returnError(w, err)
		return
	}

	returnCreated(w, resp)
}

func (s *Server) handleCreateModule(w http.ResponseWriter, r *http.Request) {
	var req internal.CreateModuleRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		returnBadRequest(w, fmt.Errorf("invalid request body"))
		return
	}

	resp, err := s.createModule.Exec(r.Context(), req)
	if err != nil {
		returnError(w, err)
		return
	}

	returnCreated(w, resp)
}

func (s *Server) handleMergeChangeset(w http.ResponseWriter, r *http.Request) {
	changesetName := chi.URLParam(r, "changesetName")
	if changesetName == "" {
		returnBadRequest(w, fmt.Errorf("changeset name is required"))
		return
	}

	req := internal.MergeChangesetRequest{
		ChangesetName: changesetName,
	}

	resp, err := s.mergeChangeset.Exec(r.Context(), req)
	if err != nil {
		returnError(w, err)
		return
	}

	returnSuccess(w, resp)
}

func (s *Server) handleCreatePlan(w http.ResponseWriter, r *http.Request) {
	changesetName := chi.URLParam(r, "changesetName")
	if changesetName == "" {
		returnBadRequest(w, fmt.Errorf("changeset name is required"))
		return
	}

	componentIDStr := chi.URLParam(r, "componentID")
	componentID, err := strconv.ParseUint(componentIDStr, 10, 64)
	if err != nil {
		returnBadRequest(w, fmt.Errorf("invalid component ID"))
		return
	}

	req := internal.CreatePlanRequest{
		ComponentID: uint(componentID),
		Changeset:   changesetName,
	}

	resp, err := s.createPlan.Exec(r.Context(), req)
	if err != nil {
		returnError(w, err)
		return
	}

	returnCreated(w, resp)
}

func (s *Server) handleCreateComponent(w http.ResponseWriter, r *http.Request) {
	changesetName := chi.URLParam(r, "changesetName")
	if changesetName == "" {
		returnBadRequest(w, fmt.Errorf("changeset name is required"))
		return
	}

	var req internal.CreateComponentRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		returnBadRequest(w, fmt.Errorf("invalid request body"))
		return
	}

	req.Changeset = changesetName

	resp, err := s.createComponent.Exec(r.Context(), req)
	if err != nil {
		returnError(w, err)
		return
	}

	returnCreated(w, resp)
}

func (s *Server) handleUpdateComponent(w http.ResponseWriter, r *http.Request) {
	changesetName := chi.URLParam(r, "changesetName")
	if changesetName == "" {
		returnBadRequest(w, fmt.Errorf("changeset name is required"))
		return
	}

	componentIDStr := chi.URLParam(r, "componentID")
	componentID, err := strconv.ParseUint(componentIDStr, 10, 64)
	if err != nil {
		returnBadRequest(w, fmt.Errorf("invalid component ID"))
		return
	}

	var req internal.UpdateComponentRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		returnBadRequest(w, fmt.Errorf("invalid request body"))
		return
	}

	req.Changeset = changesetName
	req.ComponentID = uint(componentID)

	resp, err := s.updateComponent.Exec(r.Context(), req)
	if err != nil {
		returnError(w, err)
		return
	}

	returnSuccess(w, resp)
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
