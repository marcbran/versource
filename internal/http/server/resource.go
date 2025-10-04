package server

import (
	"net/http"

	"github.com/marcbran/versource/internal"
)

func (s *Server) handleListResources(w http.ResponseWriter, r *http.Request) {
	resp, err := s.facade.ListResources(r.Context(), internal.ListResourcesRequest{})
	if err != nil {
		returnError(w, err)
		return
	}

	returnSuccess(w, resp)
}
