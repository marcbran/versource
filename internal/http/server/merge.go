package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/marcbran/versource/internal"
)

func (s *Server) handleGetMerge(w http.ResponseWriter, r *http.Request) {
	mergeIDStr := chi.URLParam(r, "mergeID")
	changesetName := chi.URLParam(r, "changesetName")

	mergeID, err := strconv.ParseUint(mergeIDStr, 10, 32)
	if err != nil {
		returnBadRequest(w, fmt.Errorf("invalid merge ID: %s", mergeIDStr))
		return
	}

	req := internal.GetMergeRequest{
		MergeID:       uint(mergeID),
		ChangesetName: changesetName,
	}

	resp, err := s.facade.GetMerge(r.Context(), req)
	if err != nil {
		returnError(w, err)
		return
	}

	returnSuccess(w, resp)
}

func (s *Server) handleListMerges(w http.ResponseWriter, r *http.Request) {
	changesetName := chi.URLParam(r, "changesetName")

	req := internal.ListMergesRequest{
		ChangesetName: changesetName,
	}

	resp, err := s.facade.ListMerges(r.Context(), req)
	if err != nil {
		returnError(w, err)
		return
	}

	returnSuccess(w, resp)
}
