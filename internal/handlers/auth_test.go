package handlers_test

import (
	"embed"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"conductor/internal/db"
	"conductor/internal/handlers"
)

func newTestHandler(t *testing.T) (*handlers.Handler, *db.DB) {
	t.Helper()
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	h := handlers.New(database, false, embed.FS{}) // secureCookie=false, empty FS for tests
	return h, database
}

func TestLoginSuccess(t *testing.T) {
	h, database := newTestHandler(t)

	// Create a test user directly in DB
	hash, _ := bcryptHash("password123")
	database.CreateUser("alice", hash)

	form := url.Values{"username": {"alice"}, "password": {"password123"}}
	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("want redirect 303, got %d", w.Code)
	}
	if w.Header().Get("Location") != "/" {
		t.Errorf("want redirect to /, got %q", w.Header().Get("Location"))
	}
	// Should set a session cookie
	cookies := w.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "session" {
			found = true
		}
	}
	if !found {
		t.Error("want session cookie to be set")
	}
}

func TestLoginBadPassword(t *testing.T) {
	h, database := newTestHandler(t)
	hash, _ := bcryptHash("correct")
	database.CreateUser("alice", hash)

	form := url.Values{"username": {"alice"}, "password": {"wrong"}}
	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("want 401, got %d", w.Code)
	}
}

func TestUnauthenticatedRedirect(t *testing.T) {
	h, _ := newTestHandler(t)
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("want redirect to login, got %d", w.Code)
	}
}
