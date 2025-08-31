package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/marcbran/versource/internal"
)

func (s *Server) handleListComponents(w http.ResponseWriter, r *http.Request) {
	req := internal.ListComponentsRequest{}

	if moduleIDStr := r.URL.Query().Get("module-id"); moduleIDStr != "" {
		moduleID, err := strconv.ParseUint(moduleIDStr, 10, 32)
		if err != nil {
			returnBadRequest(w, fmt.Errorf("invalid module-id"))
			return
		}
		moduleIDUint := uint(moduleID)
		req.ModuleID = &moduleIDUint
	}

	if moduleVersionIDStr := r.URL.Query().Get("module-version-id"); moduleVersionIDStr != "" {
		moduleVersionID, err := strconv.ParseUint(moduleVersionIDStr, 10, 32)
		if err != nil {
			returnBadRequest(w, fmt.Errorf("invalid module-version-id"))
			return
		}
		moduleVersionIDUint := uint(moduleVersionID)
		req.ModuleVersionID = &moduleVersionIDUint
	}

	resp, err := s.listComponents.Exec(r.Context(), req)
	if err != nil {
		returnError(w, err)
		return
	}

	returnSuccess(w, resp)
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
