package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/marcbran/versource/pkg/versource"
)

func (s *Server) handleListChangesets(w http.ResponseWriter, r *http.Request) {
	resp, err := s.facade.ListChangesets(r.Context(), versource.ListChangesetsRequest{})
	if err != nil {
		returnError(w, err)
		return
	}

	returnSuccess(w, resp)
}

func (s *Server) handleCreateChangeset(w http.ResponseWriter, r *http.Request) {
	var req versource.CreateChangesetRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		returnBadRequest(w, fmt.Errorf("invalid request body"))
		return
	}

	resp, err := s.facade.CreateChangeset(r.Context(), req)
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

	req := versource.CreateMergeRequest{
		ChangesetName: changesetName,
	}

	resp, err := s.facade.CreateMerge(r.Context(), req)
	if err != nil {
		returnError(w, err)
		return
	}

	returnSuccess(w, resp)
}

func (s *Server) handleDeleteChangeset(w http.ResponseWriter, r *http.Request) {
	changesetName := chi.URLParam(r, "changesetName")
	if changesetName == "" {
		returnBadRequest(w, fmt.Errorf("changeset name is required"))
		return
	}

	req := versource.DeleteChangesetRequest{
		ChangesetName: changesetName,
	}

	resp, err := s.facade.DeleteChangeset(r.Context(), req)
	if err != nil {
		returnError(w, err)
		return
	}

	returnSuccess(w, resp)
}
