package server_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"cockadeck/internal/auth"
	"cockadeck/internal/catalog"
	"cockadeck/internal/config"
	"cockadeck/internal/scryfall"
	"cockadeck/internal/server"
	"cockadeck/internal/store"
	"cockadeck/internal/store/sqlcgen"
)

func newTestServer(t *testing.T, scryfallURL string) (http.Handler, *sqlcgen.Queries) {
	t.Helper()
	db, err := store.Open(context.Background(), filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	q := sqlcgen.New(db)
	svc := catalog.New(q, db, scryfall.NewClient(scryfallURL))
	return server.New(config.Config{JWTSecret: "test-secret", Env: "dev"}, q, svc), q
}

// csrfClient posts a form with the Origin header the CSRF middleware requires.
func postForm(t *testing.T, srv *httptest.Server, client *http.Client, path string, form url.Values) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, srv.URL+path, strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Origin", srv.URL)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func noRedirectClient() *http.Client {
	return &http.Client{
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

// registerAndCookie creates an account and returns its session cookie.
func registerAndCookie(t *testing.T, srv *httptest.Server, client *http.Client) *http.Cookie {
	t.Helper()
	resp := postForm(t, srv, client, "/register", url.Values{
		"email":        {"user@example.com"},
		"display_name": {"Tester"},
		"password":     {"tencharacters!"},
	})
	defer resp.Body.Close()
	for _, c := range resp.Cookies() {
		if c.Name == auth.CookieName {
			return c
		}
	}
	t.Fatal("register: no session cookie")
	return nil
}

const fixtureCardJSON = `{
	"object": "card",
	"id": "aaaa-bbbb-cccc",
	"oracle_id": "oracle-1",
	"name": "Prospector's Pick",
	"set": "otj",
	"set_name": "Outlaws of Thunder Junction",
	"collector_number": "10",
	"rarity": "common",
	"mana_cost": "{1}{G}",
	"cmc": 2.0,
	"type_line": "Artifact — Equipment",
	"image_uris": {"normal": "https://cards.scryfall.io/normal/front/1/0/x.jpg"}
}`

// scryfallFixture serves canned responses and counts requests.
func scryfallFixture(t *testing.T) (*httptest.Server, *int) {
	t.Helper()
	var hits int
	mux := http.NewServeMux()
	mux.HandleFunc("GET /cards/otj/10", func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, fixtureCardJSON)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, `{"object":"error","code":"not_found","details":"not found"}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, &hits
}

func authedGet(t *testing.T, srv *httptest.Server, client *http.Client, cookie *http.Cookie, path string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, srv.URL+path, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.AddCookie(cookie)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func TestSearchFlow(t *testing.T) {
	fixture, hits := scryfallFixture(t)
	srv := httptest.NewServer(mustServer(newTestServer(t, fixture.URL)))
	defer srv.Close()
	client := noRedirectClient()
	cookie := registerAndCookie(t, srv, client)

	get := func(path string) (int, string, http.Header) {
		resp := authedGet(t, srv, client, cookie, path)
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, string(body), resp.Header
	}

	// Cache miss → hits Scryfall, renders card.
	code, body, hdr := get("/search/results?set=otj&num=10")
	if code != http.StatusOK || !strings.Contains(body, "Prospector") {
		t.Fatalf("search hit: status %d, body missing card name", code)
	}
	if got := hdr.Get("HX-Push-Url"); got != "/search?set=otj&num=10" {
		t.Fatalf("HX-Push-Url = %q", got)
	}

	// Cache hit → no second API call.
	if _, body, _ = get("/search/results?set=otj&num=10"); !strings.Contains(body, "Prospector") {
		t.Fatal("cached lookup missing card")
	}
	if *hits != 1 {
		t.Fatalf("cache miss: fixture hit %d times, want 1", *hits)
	}

	// Not found.
	if _, body, _ = get("/search/results?set=otj&num=999"); !strings.Contains(body, "No card found for OTJ #999") {
		t.Fatalf("not-found message missing: %s", body)
	}

	// Invalid input.
	if _, body, _ = get("/search/results?set=&num="); !strings.Contains(body, "Enter a set code") {
		t.Fatalf("invalid-input message missing: %s", body)
	}

	// Full page render with results (shareable URL).
	code, body, _ = get("/search?set=otj&num=10")
	if code != http.StatusOK || !strings.Contains(body, "Prospector") || !strings.Contains(body, "Card Search") {
		t.Fatal("full search page did not render results")
	}
}

func TestAuthFlow(t *testing.T) {
	srv := httptest.NewServer(mustServer(newTestServer(t, "http://127.0.0.1:1")))
	defer srv.Close()
	client := noRedirectClient()

	// Register.
	resp := postForm(t, srv, client, "/register", url.Values{
		"email":        {"user@example.com"},
		"display_name": {"Tester"},
		"password":     {"tencharacters!"},
	})
	if resp.StatusCode != http.StatusSeeOther {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		t.Fatalf("register: got %d, want 303; body: %s", resp.StatusCode, body)
	}
	cookies := resp.Cookies()
	resp.Body.Close()
	if len(cookies) == 0 || cookies[0].Name != auth.CookieName {
		t.Fatal("register: no session cookie set")
	}
	token := cookies[0].Value

	// Authed page.
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/decks", nil)
	req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: token})
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK || !strings.Contains(string(body), "Decks") {
		t.Fatalf("decks page: status %d, body missing content", resp.StatusCode)
	}

	// Duplicate email rejected with a clear message.
	resp = postForm(t, srv, client, "/register", url.Values{
		"email":        {"User@Example.com"},
		"display_name": {"Other"},
		"password":     {"tencharacters!"},
	})
	body, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusConflict || !strings.Contains(string(body), "already exists") {
		t.Fatalf("duplicate register: status %d", resp.StatusCode)
	}

	// Bad login.
	resp = postForm(t, srv, client, "/login", url.Values{
		"email":    {"user@example.com"},
		"password": {"wrong-password"},
	})
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("bad login: got %d, want re-rendered 200", resp.StatusCode)
	}

	// Good login.
	resp = postForm(t, srv, client, "/login", url.Values{
		"email":    {"user@example.com"},
		"password": {"tencharacters!"},
	})
	resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("login: got %d, want 303", resp.StatusCode)
	}
}

func TestAuthGating(t *testing.T) {
	srv := httptest.NewServer(mustServer(newTestServer(t, "http://127.0.0.1:1")))
	defer srv.Close()
	client := noRedirectClient()

	resp, err := client.Get(srv.URL + "/decks")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther || resp.Header.Get("Location") != "/login" {
		t.Fatalf("unauthed /decks: got %d %q", resp.StatusCode, resp.Header.Get("Location"))
	}

	resp, err = client.Get(srv.URL + "/login")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /login: got %d", resp.StatusCode)
	}
}

func TestCSRFCheck(t *testing.T) {
	srv := httptest.NewServer(mustServer(newTestServer(t, "http://127.0.0.1:1")))
	defer srv.Close()
	client := noRedirectClient()

	// Mutation without Origin/Referer is rejected.
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/login",
		strings.NewReader("email=a%40b.co&password=abcdefghij"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("csrf: got %d, want 403", resp.StatusCode)
	}

	// Cross-site origin is rejected too.
	req, _ = http.NewRequest(http.MethodPost, srv.URL+"/login",
		strings.NewReader("email=a%40b.co&password=abcdefghij"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Origin", "https://evil.example")
	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("csrf cross-origin: got %d, want 403", resp.StatusCode)
	}
}

func mustServer(h http.Handler, _ *sqlcgen.Queries) http.Handler { return h }
