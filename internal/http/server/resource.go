package server

import (
	"net/http"

	"github.com/marcbran/versource/pkg/versource"
)

func (s *Server) handleListResources(w http.ResponseWriter, r *http.Request) {
	resp, err := s.facade.ListResources(r.Context(), versource.ListResourcesRequest{})
	if err != nil {
		returnError(w, err)
		return
	}

	returnSuccess(w, resp)
}
