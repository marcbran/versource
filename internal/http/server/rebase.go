package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/marcbran/versource/pkg/versource"
)

func (s *Server) handleGetRebase(w http.ResponseWriter, r *http.Request) {
	rebaseIDStr := chi.URLParam(r, "rebaseID")
	changesetName := chi.URLParam(r, "changesetName")

	rebaseID, err := strconv.ParseUint(rebaseIDStr, 10, 32)
	if err != nil {
		returnBadRequest(w, fmt.Errorf("invalid rebase ID: %s", rebaseIDStr))
		return
	}

	req := versource.GetRebaseRequest{
		RebaseID:      uint(rebaseID),
		ChangesetName: changesetName,
	}

	resp, err := s.facade.GetRebase(r.Context(), req)
	if err != nil {
		returnError(w, err)
		return
	}

	returnSuccess(w, resp)
}

func (s *Server) handleListRebases(w http.ResponseWriter, r *http.Request) {
	changesetName := chi.URLParam(r, "changesetName")

	req := versource.ListRebasesRequest{
		ChangesetName: changesetName,
	}

	resp, err := s.facade.ListRebases(r.Context(), req)
	if err != nil {
		returnError(w, err)
		return
	}

	returnSuccess(w, resp)
}

func (s *Server) handleCreateRebase(w http.ResponseWriter, r *http.Request) {
	changesetName := chi.URLParam(r, "changesetName")

	req := versource.CreateRebaseRequest{
		ChangesetName: changesetName,
	}

	resp, err := s.facade.CreateRebase(r.Context(), req)
	if err != nil {
		returnError(w, err)
		return
	}

	returnCreated(w, resp)
}
