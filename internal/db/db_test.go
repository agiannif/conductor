package db_test

import (
	"testing"
	"time"

	"conductor/internal/db"
	"conductor/internal/models"
)

func newTestDB(t *testing.T) *db.DB {
	t.Helper()
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	return database
}

func TestCreateAndGetProject(t *testing.T) {
	d := newTestDB(t)

	id, err := d.CreateProject("Kitchen reno", "Phase 1")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	projects, err := d.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("want 1 project, got %d", len(projects))
	}
	if projects[0].ID != id {
		t.Errorf("want ID %d, got %d", id, projects[0].ID)
	}
	if projects[0].Name != "Kitchen reno" {
		t.Errorf("want name %q, got %q", "Kitchen reno", projects[0].Name)
	}
}

func TestCreateAndGetTask(t *testing.T) {
	d := newTestDB(t)

	projectID, _ := d.CreateProject("Test project", "")
	userID, _ := d.CreateUser("alice", "hashedpw")

	due := time.Now().Add(24 * time.Hour)
	taskID, err := d.CreateTask(db.CreateTaskParams{
		Title:     "Buy milk",
		Status:    models.StatusTodo,
		Priority:  models.PriorityMedium,
		ProjectID: projectID,
		CreatedBy: userID,
		DueDate:   &due,
	})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	task, err := d.GetTask(taskID)
	if err != nil {
		t.Fatalf("GetTask: %v", err)
	}
	if task.Title != "Buy milk" {
		t.Errorf("want title %q, got %q", "Buy milk", task.Title)
	}
	if task.Status != models.StatusTodo {
		t.Errorf("want status todo, got %q", task.Status)
	}
}

func TestToggleTaskDone(t *testing.T) {
	d := newTestDB(t)
	projectID, _ := d.CreateProject("p", "")
	userID, _ := d.CreateUser("alice", "pw")
	taskID, _ := d.CreateTask(db.CreateTaskParams{
		Title: "t", Status: models.StatusTodo,
		Priority: models.PriorityLow, ProjectID: projectID, CreatedBy: userID,
	})

	if err := d.ToggleTask(taskID); err != nil {
		t.Fatalf("ToggleTask: %v", err)
	}
	task, _ := d.GetTask(taskID)
	if task.Status != models.StatusDone {
		t.Errorf("want done after first toggle, got %q", task.Status)
	}

	if err := d.ToggleTask(taskID); err != nil {
		t.Fatalf("ToggleTask: %v", err)
	}
	task, _ = d.GetTask(taskID)
	if task.Status != models.StatusTodo {
		t.Errorf("want todo after second toggle, got %q", task.Status)
	}
}

func TestSessionSlidingWindow(t *testing.T) {
	d := newTestDB(t)
	userID, _ := d.CreateUser("bob", "pw")

	token, err := d.CreateSession(userID)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	session, err := d.GetSession(token)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if session.UserID != userID {
		t.Errorf("want userID %d, got %d", userID, session.UserID)
	}

	// Extend session
	if err := d.ExtendSession(token); err != nil {
		t.Fatalf("ExtendSession: %v", err)
	}

	// Expired sessions should not be returned
	if err := d.DeleteExpiredSessions(); err != nil {
		t.Fatalf("DeleteExpiredSessions: %v", err)
	}
	_, err = d.GetSession(token)
	if err != nil {
		t.Errorf("valid session should still be gettable: %v", err)
	}
}

func TestDeleteProjectBlockedByTasks(t *testing.T) {
	d := newTestDB(t)
	projectID, _ := d.CreateProject("p", "")
	userID, _ := d.CreateUser("alice", "pw")
	d.CreateTask(db.CreateTaskParams{
		Title: "t", Status: models.StatusTodo,
		Priority: models.PriorityLow, ProjectID: projectID, CreatedBy: userID,
	})

	err := d.DeleteProject(projectID)
	if err == nil {
		t.Fatal("want error deleting project with tasks, got nil")
	}
}
