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

func TestCreateTask(t *testing.T) {
	h, database := newTestHandler(t)
	userID, _ := database.CreateUser("alice", "pw")
	token, _ := database.CreateSession(userID)
	projectID, _ := database.CreateProject("My Project", "")

	form := url.Values{
		"title":      {"Fix login"},
		"project_id": {fmt.Sprint(projectID)},
		"status":     {"todo"},
		"priority":   {"medium"},
	}
	req := httptest.NewRequest("POST", "/tasks", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("want 303, got %d", w.Code)
	}
	want := fmt.Sprintf("/projects/%d", projectID)
	if got := w.Header().Get("Location"); got != want {
		t.Errorf("want redirect to %q, got %q", want, got)
	}

	tasks, err := database.ListTasks(models.TaskFilters{ProjectID: projectID})
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("want 1 task, got %d", len(tasks))
	}
	if tasks[0].Title != "Fix login" {
		t.Errorf("want title %q, got %q", "Fix login", tasks[0].Title)
	}
}

func TestToggleTask(t *testing.T) {
	h, database := newTestHandler(t)
	userID, _ := database.CreateUser("alice", "pw")
	token, _ := database.CreateSession(userID)
	projectID, _ := database.CreateProject("My Project", "")
	taskID, _ := database.CreateTask(db.CreateTaskParams{
		Title:     "Some task",
		Status:    models.StatusTodo,
		Priority:  models.PriorityMedium,
		ProjectID: projectID,
		CreatedBy: userID,
	})

	// First toggle: todo -> done
	req := httptest.NewRequest("POST", fmt.Sprintf("/tasks/%d/toggle", taskID), nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("want 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, fmt.Sprintf("task-%d", taskID)) {
		t.Errorf("want task row HTML with id task-%d in body, got: %s", taskID, body)
	}
	// "checked" attribute on the checkbox indicates the task is done
	if !strings.Contains(body, "checked") {
		t.Errorf("want task row to show checked (done) state, got: %s", body)
	}

	task, err := database.GetTask(taskID)
	if err != nil {
		t.Fatalf("GetTask: %v", err)
	}
	if task.Status != models.StatusDone {
		t.Errorf("want status %q after first toggle, got %q", models.StatusDone, task.Status)
	}

	// Second toggle: done -> todo
	req2 := httptest.NewRequest("POST", fmt.Sprintf("/tasks/%d/toggle", taskID), nil)
	req2.AddCookie(&http.Cookie{Name: "session", Value: token})
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, req2)

	task2, err := database.GetTask(taskID)
	if err != nil {
		t.Fatalf("GetTask (second): %v", err)
	}
	if task2.Status != models.StatusTodo {
		t.Errorf("want status %q after second toggle, got %q", models.StatusTodo, task2.Status)
	}
}

func TestUpdateTask(t *testing.T) {
	h, database := newTestHandler(t)
	userID, _ := database.CreateUser("alice", "pw")
	token, _ := database.CreateSession(userID)
	projectID, _ := database.CreateProject("My Project", "")
	taskID, _ := database.CreateTask(db.CreateTaskParams{
		Title:     "Original title",
		Status:    models.StatusTodo,
		Priority:  models.PriorityLow,
		ProjectID: projectID,
		CreatedBy: userID,
	})

	form := url.Values{
		"title":    {"Updated title"},
		"status":   {"in progress"},
		"priority": {"high"},
	}
	req := httptest.NewRequest("POST", fmt.Sprintf("/tasks/%d", taskID), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("want 303, got %d", w.Code)
	}
	wantLoc := fmt.Sprintf("/projects/%d", projectID)
	if got := w.Header().Get("Location"); got != wantLoc {
		t.Errorf("want redirect to %q, got %q", wantLoc, got)
	}

	task, err := database.GetTask(taskID)
	if err != nil {
		t.Fatalf("GetTask: %v", err)
	}
	if task.Title != "Updated title" {
		t.Errorf("want title %q, got %q", "Updated title", task.Title)
	}
	if task.Status != models.StatusInProgress {
		t.Errorf("want status %q, got %q", models.StatusInProgress, task.Status)
	}
	if task.Priority != models.PriorityHigh {
		t.Errorf("want priority %q, got %q", models.PriorityHigh, task.Priority)
	}
}

func TestDeleteTask(t *testing.T) {
	h, database := newTestHandler(t)
	userID, _ := database.CreateUser("alice", "pw")
	token, _ := database.CreateSession(userID)
	projectID, _ := database.CreateProject("My Project", "")
	taskID, _ := database.CreateTask(db.CreateTaskParams{
		Title:     "Doomed task",
		Status:    models.StatusTodo,
		Priority:  models.PriorityLow,
		ProjectID: projectID,
		CreatedBy: userID,
	})

	req := httptest.NewRequest("POST", fmt.Sprintf("/tasks/%d/delete", taskID), nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("want 303, got %d", w.Code)
	}
	wantLoc := fmt.Sprintf("/projects/%d", projectID)
	if got := w.Header().Get("Location"); got != wantLoc {
		t.Errorf("want redirect to %q, got %q", wantLoc, got)
	}

	_, err := database.GetTask(taskID)
	if err == nil {
		t.Error("want task to be deleted, but GetTask returned no error")
	}
}

func TestGetTaskDeleteConfirm(t *testing.T) {
	h, database := newTestHandler(t)
	userID, _ := database.CreateUser("alice", "pw")
	token, _ := database.CreateSession(userID)
	projectID, _ := database.CreateProject("My Project", "")
	taskID, _ := database.CreateTask(db.CreateTaskParams{
		Title:     "Doomed task",
		Status:    models.StatusTodo,
		Priority:  models.PriorityLow,
		ProjectID: projectID,
		CreatedBy: userID,
	})

	req := httptest.NewRequest("GET", fmt.Sprintf("/partials/tasks/%d/delete-confirm", taskID), nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("want 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Delete this task?") {
		t.Errorf("response missing expected heading; body: %s", w.Body.String())
	}
}

func TestToggleTaskOnProjectPage(t *testing.T) {
	h, database := newTestHandler(t)
	userID, _ := database.CreateUser("alice", "pw")
	token, _ := database.CreateSession(userID)
	projectID, _ := database.CreateProject("p", "")
	taskID, _ := database.CreateTask(db.CreateTaskParams{
		Title: "t", Status: models.StatusTodo,
		Priority: models.PriorityLow, ProjectID: projectID, CreatedBy: userID,
	})

	req := httptest.NewRequest("POST", fmt.Sprintf("/tasks/%d/toggle", taskID), nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	req.Header.Set("HX-Current-URL", fmt.Sprintf("http://localhost:8080/projects/%d", projectID))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("want 200, got %d", w.Code)
	}
	if got := w.Header().Get("HX-Reswap"); got != "none" {
		t.Errorf("want HX-Reswap: none, got %q", got)
	}
	body := w.Body.String()
	if !strings.Contains(body, "project-stats") {
		t.Errorf("want project-stats OOB in body, got %q", body)
	}
	if !strings.Contains(body, "project-task-groups") {
		t.Errorf("want project-task-groups OOB in body, got %q", body)
	}
}
