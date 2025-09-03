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

func (s *Server) handleListApplies(w http.ResponseWriter, r *http.Request) {
	resp, err := s.listApplies.Exec(r.Context(), internal.ListAppliesRequest{})
	if err != nil {
		returnError(w, err)
		return
	}

	returnSuccess(w, resp)
}

func (s *Server) handleGetApplyLog(w http.ResponseWriter, r *http.Request) {
	applyIDStr := chi.URLParam(r, "applyID")

	applyID, err := strconv.ParseUint(applyIDStr, 10, 32)
	if err != nil {
		returnBadRequest(w, fmt.Errorf("invalid apply ID: %s", applyIDStr))
		return
	}

	req := internal.GetApplyLogRequest{
		ApplyID: uint(applyID),
	}

	response, err := s.getApplyLog.Exec(r.Context(), req)
	if err != nil {
		returnError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)

	_, err = io.Copy(w, response.Content)
	if err != nil {
		log.WithError(err).Error("Failed to copy apply log content to response")
		http.Error(w, "Failed to stream apply log content", http.StatusInternalServerError)
		return
	}

	response.Content.Close()
}
