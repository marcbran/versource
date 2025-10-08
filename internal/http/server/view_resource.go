package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/marcbran/versource/internal"
)

func (s *Server) handleGetViewResource(w http.ResponseWriter, r *http.Request) {
	viewResourceIDStr := chi.URLParam(r, "viewResourceID")
	viewResourceID, err := strconv.ParseUint(viewResourceIDStr, 10, 32)
	if err != nil {
		returnBadRequest(w, fmt.Errorf("invalid view resource ID"))
		return
	}

	req := internal.GetViewResourceRequest{
		ViewResourceID: uint(viewResourceID),
	}

	resp, err := s.facade.GetViewResource(r.Context(), req)
	if err != nil {
		returnError(w, err)
		return
	}

	returnSuccess(w, resp)
}

func (s *Server) handleListViewResources(w http.ResponseWriter, r *http.Request) {
	resp, err := s.facade.ListViewResources(r.Context(), internal.ListViewResourcesRequest{})
	if err != nil {
		returnError(w, err)
		return
	}

	returnSuccess(w, resp)
}

func (s *Server) handleSaveViewResource(w http.ResponseWriter, r *http.Request) {
	var req internal.SaveViewResourceRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		returnBadRequest(w, fmt.Errorf("invalid request body"))
		return
	}

	resp, err := s.facade.SaveViewResource(r.Context(), req)
	if err != nil {
		returnError(w, err)
		return
	}

	returnSuccess(w, resp)
}

func (s *Server) handleDeleteViewResource(w http.ResponseWriter, r *http.Request) {
	viewResourceIDStr := chi.URLParam(r, "viewResourceID")
	viewResourceID, err := strconv.ParseUint(viewResourceIDStr, 10, 32)
	if err != nil {
		returnBadRequest(w, fmt.Errorf("invalid view resource ID"))
		return
	}

	req := internal.DeleteViewResourceRequest{
		ViewResourceID: uint(viewResourceID),
	}

	resp, err := s.facade.DeleteViewResource(r.Context(), req)
	if err != nil {
		returnError(w, err)
		return
	}

	returnSuccess(w, resp)
}
