package models

import "time"

// Status is the fixed set of valid task statuses.
type Status string

const (
	StatusTodo       Status = "todo"
	StatusInProgress Status = "in progress"
	StatusBlocked    Status = "blocked"
	StatusDone       Status = "done"
)

// Priority is the fixed set of valid task priorities.
type Priority string

const (
	PriorityCritical Priority = "critical"
	PriorityHigh     Priority = "high"
	PriorityMedium   Priority = "medium"
	PriorityLow      Priority = "low"
)

type User struct {
	ID        int64
	Username  string
	CreatedAt time.Time
}

type Session struct {
	ID        string
	UserID    int64
	ExpiresAt time.Time
}

type Project struct {
	ID          int64
	Name        string
	Description string
	CreatedAt   time.Time
	// Computed fields (not stored)
	TotalTasks int
	DoneTasks  int
}

func (p Project) ProgressPct() int {
	if p.TotalTasks == 0 {
		return 0
	}
	return int(float64(p.DoneTasks) / float64(p.TotalTasks) * 100)
}

func (p Project) OpenTasks() int {
	return p.TotalTasks - p.DoneTasks
}

type Task struct {
	ID            int64
	Title         string
	Description   string
	Link          string
	Status        Status
	Category      string
	Priority      Priority
	ProjectID     int64
	ProjectName   string // joined
	AssigneeID    *int64
	AssigneeName  string // joined, empty if unassigned
	DueDate       *time.Time
	CreatedBy     *int64
	CreatedByName string // joined, empty if user deleted
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (t Task) IsOverdue() bool {
	if t.DueDate == nil || t.Status == StatusDone {
		return false
	}
	return t.DueDate.Before(time.Now().Truncate(24 * time.Hour))
}

// TaskFilters holds the filter parameters for task list queries.
type TaskFilters struct {
	ProjectID  int64
	Status     string
	Category   string
	Priority   string
	AssigneeID int64
	Due        string // "overdue", "this_week", "this_month", ""
}
