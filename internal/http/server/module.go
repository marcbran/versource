package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/marcbran/versource/pkg/versource"
)

func (s *Server) handleGetModule(w http.ResponseWriter, r *http.Request) {
	moduleIDStr := chi.URLParam(r, "moduleID")
	moduleID, err := strconv.ParseUint(moduleIDStr, 10, 32)
	if err != nil {
		returnBadRequest(w, fmt.Errorf("invalid module ID"))
		return
	}

	req := versource.GetModuleRequest{
		ModuleID: uint(moduleID),
	}

	resp, err := s.facade.GetModule(r.Context(), req)
	if err != nil {
		returnError(w, err)
		return
	}

	returnSuccess(w, resp)
}

func (s *Server) handleListModules(w http.ResponseWriter, r *http.Request) {
	resp, err := s.facade.ListModules(r.Context(), versource.ListModulesRequest{})
	if err != nil {
		returnError(w, err)
		return
	}

	returnSuccess(w, resp)
}

func (s *Server) handleCreateModule(w http.ResponseWriter, r *http.Request) {
	var req versource.CreateModuleRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		returnBadRequest(w, fmt.Errorf("invalid request body"))
		return
	}

	resp, err := s.facade.CreateModule(r.Context(), req)
	if err != nil {
		returnError(w, err)
		return
	}

	returnCreated(w, resp)
}

func (s *Server) handleUpdateModule(w http.ResponseWriter, r *http.Request) {
	moduleIDStr := chi.URLParam(r, "moduleID")
	moduleID, err := strconv.ParseUint(moduleIDStr, 10, 32)
	if err != nil {
		returnBadRequest(w, fmt.Errorf("invalid module ID"))
		return
	}

	var req versource.UpdateModuleRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		returnBadRequest(w, fmt.Errorf("invalid request body"))
		return
	}

	req.ModuleID = uint(moduleID)

	resp, err := s.facade.UpdateModule(r.Context(), req)
	if err != nil {
		returnError(w, err)
		return
	}

	returnSuccess(w, resp)
}

func (s *Server) handleDeleteModule(w http.ResponseWriter, r *http.Request) {
	moduleIDStr := chi.URLParam(r, "moduleID")
	moduleID, err := strconv.ParseUint(moduleIDStr, 10, 32)
	if err != nil {
		returnBadRequest(w, fmt.Errorf("invalid module ID"))
		return
	}

	req := versource.DeleteModuleRequest{
		ModuleID: uint(moduleID),
	}

	resp, err := s.facade.DeleteModule(r.Context(), req)
	if err != nil {
		returnError(w, err)
		return
	}

	returnSuccess(w, resp)
}

func (s *Server) handleGetModuleVersion(w http.ResponseWriter, r *http.Request) {
	moduleVersionIDStr := chi.URLParam(r, "moduleVersionID")
	moduleVersionID, err := strconv.ParseUint(moduleVersionIDStr, 10, 32)
	if err != nil {
		returnBadRequest(w, fmt.Errorf("invalid module version ID"))
		return
	}

	req := versource.GetModuleVersionRequest{
		ModuleVersionID: uint(moduleVersionID),
	}

	resp, err := s.facade.GetModuleVersion(r.Context(), req)
	if err != nil {
		returnError(w, err)
		return
	}

	returnSuccess(w, resp)
}

func (s *Server) handleListModuleVersions(w http.ResponseWriter, r *http.Request) {
	resp, err := s.facade.ListModuleVersions(r.Context(), versource.ListModuleVersionsRequest{})
	if err != nil {
		returnError(w, err)
		return
	}

	returnSuccess(w, resp)
}

func (s *Server) handleListModuleVersionsForModule(w http.ResponseWriter, r *http.Request) {
	moduleIDStr := chi.URLParam(r, "moduleID")
	moduleID, err := strconv.ParseUint(moduleIDStr, 10, 32)
	if err != nil {
		returnBadRequest(w, fmt.Errorf("invalid module ID"))
		return
	}

	moduleIDUint := uint(moduleID)
	req := versource.ListModuleVersionsRequest{ModuleID: &moduleIDUint}

	resp, err := s.facade.ListModuleVersions(r.Context(), req)
	if err != nil {
		returnError(w, err)
		return
	}

	returnSuccess(w, resp)
}
