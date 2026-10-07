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
	"cockadeck/internal/config"
	"cockadeck/internal/server"
	"cockadeck/internal/store"
	"cockadeck/internal/store/sqlcgen"
)

func newTestServer(t *testing.T) http.Handler {
	t.Helper()
	db, err := store.Open(context.Background(), filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return server.New(config.Config{JWTSecret: "test-secret", Env: "dev"}, sqlcgen.New(db))
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

func TestAuthFlow(t *testing.T) {
	srv := httptest.NewServer(newTestServer(t))
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
	if resp.StatusCode != http.StatusOK || !strings.Contains(string(body), "Tester") {
		t.Fatalf("decks page: status %d, body missing display name", resp.StatusCode)
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
	srv := httptest.NewServer(newTestServer(t))
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
	srv := httptest.NewServer(newTestServer(t))
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
