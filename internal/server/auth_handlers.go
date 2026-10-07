package server

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"net/mail"
	"strings"

	"cockadeck/internal/auth"
	"cockadeck/internal/store/sqlcgen"
	"cockadeck/views"
)

func (s *Server) loginPage(w http.ResponseWriter, r *http.Request) {
	views.LoginPage("", "").Render(r.Context(), w)
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	u, err := s.q.GetUserByEmail(r.Context(), email)
	if err != nil || !auth.VerifyPassword(u.PasswordHash, password) {
		views.LoginPage(email, "Invalid email or password.").Render(r.Context(), w)
		return
	}
	s.issueSession(w, r, u.ID)
}

func (s *Server) registerPage(w http.ResponseWriter, r *http.Request) {
	views.RegisterPage("", "", "").Render(r.Context(), w)
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	email := strings.TrimSpace(r.FormValue("email"))
	displayName := strings.TrimSpace(r.FormValue("display_name"))
	password := r.FormValue("password")

	fail := func(code int, msg string) {
		w.WriteHeader(code)
		views.RegisterPage(email, displayName, msg).Render(r.Context(), w)
	}

	if addr, err := mail.ParseAddress(email); err != nil || addr.Address != email {
		fail(http.StatusUnprocessableEntity, "Enter a valid email address.")
		return
	}
	if displayName == "" {
		fail(http.StatusUnprocessableEntity, "Display name is required.")
		return
	}
	if len(password) < 10 {
		fail(http.StatusUnprocessableEntity, "Password must be at least 10 characters.")
		return
	}

	if _, err := s.q.GetUserByEmail(r.Context(), email); err == nil {
		fail(http.StatusConflict, "An account with this email already exists.")
		return
	} else if !errors.Is(err, sql.ErrNoRows) {
		slog.Error("register: lookup user", "err", err)
		fail(http.StatusInternalServerError, "Registration failed, try again.")
		return
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		slog.Error("register: hash password", "err", err)
		fail(http.StatusInternalServerError, "Registration failed, try again.")
		return
	}
	u, err := s.q.CreateUser(r.Context(), sqlcgen.CreateUserParams{
		Email:        email,
		DisplayName:  displayName,
		PasswordHash: hash,
	})
	if err != nil {
		slog.Error("register: create user", "err", err)
		fail(http.StatusInternalServerError, "Registration failed, try again.")
		return
	}
	s.issueSession(w, r, u.ID)
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	auth.ClearTokenCookie(w, s.cfg.Prod())
	redirect(w, r, "/login")
}

func (s *Server) issueSession(w http.ResponseWriter, r *http.Request, userID int64) {
	token, err := auth.IssueToken(s.cfg.JWTSecret, userID)
	if err != nil {
		slog.Error("issue token", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	auth.SetTokenCookie(w, token, s.cfg.Prod())
	redirect(w, r, "/")
}
