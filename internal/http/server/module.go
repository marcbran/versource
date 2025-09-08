package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/marcbran/versource/internal"
)

func (s *Server) handleListModules(w http.ResponseWriter, r *http.Request) {
	resp, err := s.listModules.Exec(r.Context(), internal.ListModulesRequest{})
	if err != nil {
		returnError(w, err)
		return
	}

	returnSuccess(w, resp)
}

func (s *Server) handleGetModule(w http.ResponseWriter, r *http.Request) {
	moduleIDStr := chi.URLParam(r, "moduleID")
	moduleID, err := strconv.ParseUint(moduleIDStr, 10, 32)
	if err != nil {
		returnBadRequest(w, fmt.Errorf("invalid module ID"))
		return
	}

	req := internal.GetModuleRequest{
		ModuleID: uint(moduleID),
	}

	resp, err := s.getModule.Exec(r.Context(), req)
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

	req := internal.GetModuleVersionRequest{
		ModuleVersionID: uint(moduleVersionID),
	}

	resp, err := s.getModuleVersion.Exec(r.Context(), req)
	if err != nil {
		returnError(w, err)
		return
	}

	returnSuccess(w, resp)
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

func (s *Server) handleUpdateModule(w http.ResponseWriter, r *http.Request) {
	moduleIDStr := chi.URLParam(r, "moduleID")
	moduleID, err := strconv.ParseUint(moduleIDStr, 10, 32)
	if err != nil {
		returnBadRequest(w, fmt.Errorf("invalid module ID"))
		return
	}

	var req internal.UpdateModuleRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		returnBadRequest(w, fmt.Errorf("invalid request body"))
		return
	}

	req.ModuleID = uint(moduleID)

	resp, err := s.updateModule.Exec(r.Context(), req)
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

	req := internal.DeleteModuleRequest{
		ModuleID: uint(moduleID),
	}

	resp, err := s.deleteModule.Exec(r.Context(), req)
	if err != nil {
		returnError(w, err)
		return
	}

	returnSuccess(w, resp)
}

func (s *Server) handleListModuleVersions(w http.ResponseWriter, r *http.Request) {
	resp, err := s.listModuleVersions.Exec(r.Context(), internal.ListModuleVersionsRequest{})
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

	resp, err := s.listModuleVersionsForModule.Exec(r.Context(), internal.ListModuleVersionsForModuleRequest{
		ModuleID: uint(moduleID),
	})
	if err != nil {
		returnError(w, err)
		return
	}

	returnSuccess(w, resp)
}
