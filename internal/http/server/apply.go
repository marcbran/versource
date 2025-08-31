package server

import (
	"net/http"

	"github.com/marcbran/versource/internal"
)

func (s *Server) handleListApplies(w http.ResponseWriter, r *http.Request) {
	resp, err := s.listApplies.Exec(r.Context(), internal.ListAppliesRequest{})
	if err != nil {
		returnError(w, err)
		return
	}

	returnSuccess(w, resp)
}
