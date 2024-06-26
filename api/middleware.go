package api

import (
	"context"
	"fmt"
	"github.com/artemsmotritel/oktion/templates"
	"github.com/artemsmotritel/oktion/utils"
	"log"
	"net/http"
	"slices"
	"strconv"
)

func (s *Server) onlyAuthorizedMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isAuth, err := utils.ExtractValueFromContext[bool](r.Context(), "isAuthorized")
		if err != nil {
			isAuth = false
		}

		if !isAuth {
			w.WriteHeader(http.StatusUnauthorized)
			handler := templates.NewErrorPageHandler(templates.Unauthorized)
			handler.ServeHTTP(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) protectAuctionsMiddleware(next http.Handler, auctionIdWildcard string) http.Handler {
	return s.onlyAuthorizedMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userId, err := utils.ExtractValueFromContext[int64](r.Context(), "userId")
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				handler := templates.NewErrorPageHandler(templates.Unauthorized)
				handler.ServeHTTP(w, r)
				return
			}

			auctionId, err := strconv.ParseInt(r.PathValue(auctionIdWildcard), 10, 64)
			if err != nil {
				s.badRequestError(w, r, fmt.Sprintf("Bad auction id in path: %s", r.PathValue("id")))
				return
			}

			actualOwnerId, err := s.store.GetOwnerIDByAuctionID(auctionId)

			if userId != actualOwnerId {
				w.WriteHeader(http.StatusForbidden)
				handler := templates.NewErrorPageHandler(templates.Forbidden)
				handler.ServeHTTP(w, r)
				return
			}

			next.ServeHTTP(w, r)
		}))
}

func redirectUserMiddleware(next http.Handler) http.Handler {
	pathsToRedirectAuthorizedUser := []string{"/login", "/redirect-me"}
	pathsToRedirectUnauthorizedUser := []string{"/auctions/new", "/redirect-me"}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isAuth, err := utils.ExtractValueFromContext[bool](r.Context(), "isAuthorized")
		if err != nil {
			isAuth = false
		}

		if isAuth && slices.Contains[[]string](pathsToRedirectAuthorizedUser, r.URL.Path) {
			w.Header().Set("Location", "/profile")
			w.WriteHeader(http.StatusSeeOther)
			return
		}

		if !isAuth && slices.Contains[[]string](pathsToRedirectUnauthorizedUser, r.URL.Path) {
			w.Header().Add("HX-Location", "/login")

			w.WriteHeader(http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func setUserInfoToContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := extractUserIDFromCookie(r)

		if err != nil {
			r = r.WithContext(context.WithValue(r.Context(), "isAuthorized", false))
		} else {
			r = r.WithContext(context.WithValue(r.Context(), "userId", id))
			r = r.WithContext(context.WithValue(r.Context(), "isAuthorized", true))
		}

		// TODO choose an appropriate name for the context value
		// that should mark that the response should return only a part of the html and not the whole page
		r = r.WithContext(context.WithValue(r.Context(), "hxBoosted", r.Header.Get("HX-Request") == "true"))

		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hxBoosted, _ := utils.ExtractValueFromContext[bool](r.Context(), "hxBoosted")

		l := fmt.Sprintf("New Request: method - %s, url - %s, hx-boosted - %t", r.Method, r.URL.Path, hxBoosted)
		logger.Println(l)

		isAuth := r.Context().Value("isAuthorized").(bool)
		l = fmt.Sprintf("User: isAuthorized - %t", isAuth)
		if isAuth {
			id := r.Context().Value("userId").(int64)
			l += fmt.Sprintf(", id - %d", id)
		}
		logger.Println(l)

		next.ServeHTTP(w, r)
	})
}
