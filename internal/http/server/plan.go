package server

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/marcbran/versource/internal"
	log "github.com/sirupsen/logrus"
)

func (s *Server) handleListPlans(w http.ResponseWriter, r *http.Request) {
	changesetName := chi.URLParam(r, "changesetName")

	var req internal.ListPlansRequest
	if changesetName != "" {
		req.Changeset = &changesetName
	}

	resp, err := s.listPlans.Exec(r.Context(), req)
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

func (s *Server) handleGetPlanLog(w http.ResponseWriter, r *http.Request) {
	planIDStr := chi.URLParam(r, "planID")

	planID, err := strconv.ParseUint(planIDStr, 10, 32)
	if err != nil {
		returnBadRequest(w, fmt.Errorf("invalid plan ID: %s", planIDStr))
		return
	}

	req := internal.GetPlanLogRequest{
		PlanID: uint(planID),
	}

	response, err := s.getPlanLog.Exec(r.Context(), req)
	if err != nil {
		returnError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)

	_, err = io.Copy(w, response.Content)
	if err != nil {
		log.WithError(err).Error("Failed to copy log content to response")
		http.Error(w, "Failed to stream log content", http.StatusInternalServerError)
		return
	}

	response.Content.Close()
}
