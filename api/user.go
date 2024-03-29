package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/alexedwards/argon2id"
	"github.com/artemsmotritel/oktion/templates"
	"github.com/artemsmotritel/oktion/types"
	"github.com/artemsmotritel/oktion/validation"
	"net/http"
	"strconv"
)

func (s *Server) handleSignUp(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.badRequestError(w, r, err.Error())
		return
	}

	signUpValidator := validation.NewSignUpValidator()
	ok, err := signUpValidator.Validate(r.Form, s.store)
	if err != nil {
		s.internalError(w, r)
		return
	}

	if !ok {
		w.Header().Set("HX-Retarget", "#sign-up-form")
		w.Header().Set("HX-Reswap", "outerHTML")
		handler := templates.NewSignUpErrorBadRequestHandler(signUpValidator.Values(), signUpValidator.Errors)
		handler.ServeHTTP(w, r)
		return
	}

	user := validation.MapUserCreateRequestToUser(signUpValidator)

	hash, err := argon2id.CreateHash(user.Password, argon2id.DefaultParams)
	if err != nil {
		s.internalError(w, r)
		return
	}

	user.Password = hash

	err = s.store.SaveUser(user)
	if err != nil {
		s.internalError(w, r)
		return
	}

	cookie := http.Cookie{
		Name:     "userId",
		Value:    strconv.FormatInt(user.ID, 10),
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, &cookie)
	w.Header().Set("HX-Push-Url", "/")
	c, _ := s.store.GetCategories()
	r = r.WithContext(context.WithValue(r.Context(), "isAuthorized", true))
	handler := templates.NewIndexBodyHandler(c)
	handler.ServeHTTP(w, r)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.badRequestError(w, r, err.Error())
		return
	}

	loginValidator := validation.NewLoginValidator()
	ok, err := loginValidator.Validate(r.Form, s.store)
	if err != nil {
		s.internalError(w, r)
		return
	}

	if !ok {
		w.Header().Set("HX-Retarget", "#login-form")
		w.Header().Set("HX-Reswap", "outerHTML")
		handler := templates.NewLoginErrorBadRequestHandler(loginValidator.Values(), loginValidator.Errors)
		handler.ServeHTTP(w, r)
		return
	}

	user, err := s.store.GetUserByEmail(loginValidator.Email)
	if err != nil {
		s.internalError(w, r)
		return
	}

	cookie := http.Cookie{
		Name:     "userId",
		Value:    strconv.FormatInt(user.ID, 10),
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, &cookie)
	w.Header().Set("HX-Push-Url", "/")
	c, _ := s.store.GetCategories()
	r = r.WithContext(context.WithValue(r.Context(), "isAuthorized", true))
	handler := templates.NewIndexBodyHandler(c)
	handler.ServeHTTP(w, r)
}

func (s *Server) handleGetUserByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)

	if err != nil {
		s.badRequestError(w, r, fmt.Sprintf("Bad user id in path: %s", r.PathValue("id")))
		return
	}

	user, err := s.store.GetUserByID(id)

	if err != nil {
		s.internalError(w, r)
		return
	}

	if user == nil {
		s.handleNotFound(w, r)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(user); err != nil {
		s.logger.Println("ERROR: ", err.Error())
	}
}

func (s *Server) handleGetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := s.store.GetUsers()

	if err != nil {
		s.internalError(w, r)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(users); err != nil {
		s.logger.Println("ERROR: ", err.Error())
	}
}

func (s *Server) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)

	if err != nil {
		s.badRequestError(w, r, fmt.Sprintf("Bad user id in path: %s", r.PathValue("id")))
		return
	}

	bodyReader := json.NewDecoder(r.Body)
	var userRequest types.UserUpdateRequest

	if err = bodyReader.Decode(&userRequest); err != nil {
		s.badRequestError(w, r, "Bad request body")
		return
	}

	if err = s.store.UpdateUser(id, &userRequest); err != nil {
		s.internalError(w, r)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)

	if err != nil {
		s.badRequestError(w, r, fmt.Sprintf("Bad user id in path: %s", r.PathValue("id")))
		return
	}

	if err = s.store.DeleteUser(id); err != nil {
		s.internalError(w, r)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleLogout(w http.ResponseWriter, _ *http.Request) {
	cookie := http.Cookie{
		Name:     "userId",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
	w.Header().Set("HX-Redirect", "/")
}
