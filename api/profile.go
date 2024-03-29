package api

import (
	"github.com/artemsmotritel/oktion/templates"
	"github.com/artemsmotritel/oktion/utils"
	"net/http"
)

func (s *Server) handleGetProfile(w http.ResponseWriter, r *http.Request) {
	userId, err := utils.ExtractValueFromContext[int64](r.Context(), "userId")
	if err != nil {
		s.handleUnauthorized(w, r)
		return
	}

	user, err := s.store.GetUserByID(userId)
	if err != nil {
		s.internalError(w, r)
		return
	}

	handler := templates.NewProfilePageHandler(user)
	handler.ServeHTTP(w, r)
}
