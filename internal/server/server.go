package server

import (
	"net/http"
	"time"

	"cockadeck/internal/auth"
	"cockadeck/internal/config"
	"cockadeck/internal/store/sqlcgen"
	"cockadeck/static"
	"cockadeck/views"

	"golang.org/x/time/rate"
)

type Server struct {
	cfg config.Config
	q   sqlcgen.Querier
}

func New(cfg config.Config, q sqlcgen.Querier) http.Handler {
	s := &Server{cfg: cfg, q: q}

	public := http.NewServeMux()
	authed := http.NewServeMux()

	authLimiter := newIPRateLimiter(rate.Every(6*time.Second), 5)
	public.Handle("GET /login", authLimiter.wrap(http.HandlerFunc(s.loginPage)))
	public.Handle("POST /login", authLimiter.wrap(http.HandlerFunc(s.login)))
	public.Handle("GET /register", authLimiter.wrap(http.HandlerFunc(s.registerPage)))
	public.Handle("POST /register", authLimiter.wrap(http.HandlerFunc(s.register)))
	public.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(static.FS)))

	authed.HandleFunc("GET /{$}", s.home)
	authed.HandleFunc("GET /decks", s.decksPage)
	authed.HandleFunc("POST /logout", s.logout)

	public.Handle("/", auth.RequireAuth(cfg.JWTSecret, authed))

	return recoverPanic(requestLog(secureHeaders(csrfCheck(public))))
}

func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/decks", http.StatusSeeOther)
}

func (s *Server) decksPage(w http.ResponseWriter, r *http.Request) {
	uid, _ := auth.UserIDFromContext(r.Context())
	u, err := s.q.GetUserByID(r.Context(), uid)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	views.DecksPage(u.DisplayName).Render(r.Context(), w)
}
