package server

import (
	"net/http"
	"time"

	"cockadeck/internal/auth"
	"cockadeck/internal/catalog"
	"cockadeck/internal/config"
	"cockadeck/internal/store/sqlcgen"
	"cockadeck/static"

	"golang.org/x/time/rate"
)

type Server struct {
	cfg   config.Config
	q     sqlcgen.Querier
	cards *catalog.Service
}

func New(cfg config.Config, q sqlcgen.Querier, cards *catalog.Service) http.Handler {
	s := &Server{cfg: cfg, q: q, cards: cards}

	public := http.NewServeMux()
	authed := http.NewServeMux()

	authLimiter := newIPRateLimiter(rate.Every(6*time.Second), 5)
	public.Handle("GET /login", authLimiter.wrap(http.HandlerFunc(s.loginPage)))
	public.Handle("POST /login", authLimiter.wrap(http.HandlerFunc(s.login)))
	public.Handle("GET /register", authLimiter.wrap(http.HandlerFunc(s.registerPage)))
	public.Handle("POST /register", authLimiter.wrap(http.HandlerFunc(s.register)))
	public.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(static.FS)))

	authed.HandleFunc("GET /{$}", s.home)
	authed.HandleFunc("GET /search", s.searchPage)
	authed.HandleFunc("GET /search/results", s.searchResults)
	authed.HandleFunc("POST /logout", s.logout)

	authed.HandleFunc("GET /decks", s.decksPage)
	authed.HandleFunc("POST /decks", s.createDeck)
	authed.HandleFunc("GET /decks/{id}", s.deckPage)
	authed.HandleFunc("POST /decks/{id}", s.updateDeck)
	authed.HandleFunc("DELETE /decks/{id}", s.deleteDeck)
	authed.HandleFunc("POST /decks/{id}/cards", s.addDeckCard)
	authed.HandleFunc("PATCH /decks/{id}/cards/{cardID}", s.updateDeckCard)
	authed.HandleFunc("DELETE /decks/{id}/cards/{cardID}", s.deleteDeckCard)

	public.Handle("/", auth.RequireAuth(cfg.JWTSecret, authed))

	return recoverPanic(requestLog(secureHeaders(csrfCheck(public))))
}

func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/decks", http.StatusSeeOther)
}
