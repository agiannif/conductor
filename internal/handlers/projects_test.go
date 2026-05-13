package handlers_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"conductor/internal/db"
	"conductor/internal/models"
)

func TestCreateProject(t *testing.T) {
	h, database := newTestHandler(t)
	userID, _ := database.CreateUser("alice", "pw")
	token, _ := database.CreateSession(userID)

	form := url.Values{"name": {"Kitchen reno"}, "description": {"Phase 1"}}
	req := httptest.NewRequest("POST", "/projects", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("want 303, got %d", w.Code)
	}
	projects, _ := database.ListProjects()
	if len(projects) != 1 {
		t.Fatalf("want 1 project, got %d", len(projects))
	}
	if projects[0].Name != "Kitchen reno" {
		t.Errorf("want name %q, got %q", "Kitchen reno", projects[0].Name)
	}
	want := fmt.Sprintf("/projects/%d", projects[0].ID)
	if got := w.Header().Get("Location"); got != want {
		t.Errorf("want redirect to %q, got %q", want, got)
	}
}

func TestDeleteProjectBlocked(t *testing.T) {
	h, database := newTestHandler(t)
	userID, _ := database.CreateUser("alice", "pw")
	token, _ := database.CreateSession(userID)
	projectID, _ := database.CreateProject("p", "")
	database.CreateTask(db.CreateTaskParams{
		Title: "t", Status: models.StatusTodo,
		Priority: models.PriorityLow, ProjectID: projectID, CreatedBy: userID,
	})

	req := httptest.NewRequest("POST", fmt.Sprintf("/projects/%d/delete", projectID), nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	// Should return 409 Conflict (blocked by tasks)
	if w.Code != http.StatusConflict {
		t.Errorf("want 409, got %d", w.Code)
	}
	// Project should still exist
	projects, _ := database.ListProjects()
	if len(projects) != 1 {
		t.Error("project should still exist when delete is blocked")
	}
}

func TestGetProjectDeleteConfirm(t *testing.T) {
	h, database := newTestHandler(t)
	userID, _ := database.CreateUser("alice", "pw")
	token, _ := database.CreateSession(userID)
	projectID, _ := database.CreateProject("Garden", "")

	req := httptest.NewRequest("GET", fmt.Sprintf("/partials/projects/%d/delete-confirm", projectID), nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("want 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Garden") {
		t.Errorf("want project name in body, got %q", w.Body.String())
	}
}

func TestDeleteProjectHTMX(t *testing.T) {
	h, database := newTestHandler(t)
	userID, _ := database.CreateUser("alice", "pw")
	token, _ := database.CreateSession(userID)
	projectID, _ := database.CreateProject("Garden", "")

	req := httptest.NewRequest("POST", fmt.Sprintf("/projects/%d/delete", projectID), nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("want 200, got %d", w.Code)
	}
	if got := w.Header().Get("HX-Redirect"); got != "/" {
		t.Errorf("want HX-Redirect: /, got %q", got)
	}
	projects, _ := database.ListProjects()
	if len(projects) != 0 {
		t.Error("want project deleted, but it still exists")
	}
}

func TestDeleteProjectBlockedHTMX(t *testing.T) {
	h, database := newTestHandler(t)
	userID, _ := database.CreateUser("alice", "pw")
	token, _ := database.CreateSession(userID)
	projectID, _ := database.CreateProject("Garden", "")
	database.CreateTask(db.CreateTaskParams{
		Title: "t", Status: models.StatusTodo,
		Priority: models.PriorityLow, ProjectID: projectID, CreatedBy: userID,
	})

	req := httptest.NewRequest("POST", fmt.Sprintf("/projects/%d/delete", projectID), nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("want 200 for HTMX blocked delete, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Blocked") {
		t.Errorf("want error content in body, got %q", w.Body.String())
	}
	projects, _ := database.ListProjects()
	if len(projects) != 1 {
		t.Error("project should still exist when delete is blocked")
	}
}
