// cmd/seed/main_test.go
package main

import (
	"testing"

	"conductor/internal/db"
	"conductor/internal/models"
)

func openTestDB(t *testing.T) *db.DB {
	t.Helper()
	d, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	return d
}

func TestSeedDB(t *testing.T) {
	d := openTestDB(t)

	if err := seedDB(d); err != nil {
		t.Fatalf("seedDB: %v", err)
	}

	users, err := d.ListUsers()
	if err != nil {
		t.Fatalf("ListUsers: %v", err)
	}
	if len(users) != 3 {
		t.Errorf("want 3 users, got %d", len(users))
	}

	projects, err := d.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects: %v", err)
	}
	if len(projects) != 5 {
		t.Errorf("want 5 projects, got %d", len(projects))
	}

	tasks, err := d.ListTasks(models.TaskFilters{})
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}
	if len(tasks) != 18 {
		t.Errorf("want 18 tasks, got %d", len(tasks))
	}

	statuses := make(map[models.Status]bool)
	priorities := make(map[models.Priority]bool)
	hasOverdue := false
	hasNoDueDate := false
	hasUnassigned := false

	for _, task := range tasks {
		statuses[task.Status] = true
		priorities[task.Priority] = true
		if task.DueDate == nil {
			hasNoDueDate = true
		}
		if task.IsOverdue() {
			hasOverdue = true
		}
		if task.AssigneeID == nil {
			hasUnassigned = true
		}
	}

	for _, s := range []models.Status{
		models.StatusTodo, models.StatusInProgress,
		models.StatusBlocked, models.StatusDone,
	} {
		if !statuses[s] {
			t.Errorf("no tasks with status %q", s)
		}
	}
	for _, p := range []models.Priority{
		models.PriorityCritical, models.PriorityHigh,
		models.PriorityMedium, models.PriorityLow,
	} {
		if !priorities[p] {
			t.Errorf("no tasks with priority %q", p)
		}
	}
	if !hasOverdue {
		t.Error("no overdue non-done tasks")
	}
	if !hasNoDueDate {
		t.Error("no tasks without a due date")
	}
	if !hasUnassigned {
		t.Error("no unassigned tasks")
	}
}
