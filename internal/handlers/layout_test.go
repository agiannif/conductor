package handlers

import (
	"context"
	"embed"
	"net/http/httptest"
	"testing"

	"conductor/internal/db"
	"conductor/internal/models"
)

func TestLayoutDataCurrentPath(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	h := New(database, false, embed.FS{})

	req := httptest.NewRequest("GET", "/tasks", nil)
	req = req.WithContext(context.WithValue(req.Context(), contextKeyUser, models.User{ID: 1}))
	data := h.layoutData(req)
	if data.CurrentPath != "/tasks" {
		t.Errorf("want CurrentPath %q, got %q", "/tasks", data.CurrentPath)
	}
}
