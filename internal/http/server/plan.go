package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/marcbran/versource/internal"
)

func (s *Server) handleListPlans(w http.ResponseWriter, r *http.Request) {
	resp, err := s.listPlans.Exec(r.Context(), internal.ListPlansRequest{})
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
