package templates

import (
	"conductor/internal/models"
	"testing"
	"time"
)

func TestTaskIsInBucket(t *testing.T) {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	tomorrow := now.AddDate(0, 0, 1)
	in8Days := now.AddDate(0, 0, 8)

	overdueTask := models.Task{Status: models.StatusTodo, DueDate: &yesterday}
	thisWeekTask := models.Task{Status: models.StatusTodo, DueDate: &tomorrow}
	laterTask := models.Task{Status: models.StatusTodo, DueDate: &in8Days}
	noDueTask := models.Task{Status: models.StatusTodo}
	doneTask := models.Task{Status: models.StatusDone, DueDate: &yesterday}

	cases := []struct {
		task   models.Task
		bucket string
		want   bool
	}{
		{overdueTask, "overdue", true},
		{overdueTask, "this_week", false},
		{overdueTask, "later", false},
		{thisWeekTask, "overdue", false},
		{thisWeekTask, "this_week", true},
		{thisWeekTask, "later", false},
		{laterTask, "overdue", false},
		{laterTask, "this_week", false},
		{laterTask, "later", true},
		{noDueTask, "overdue", false},
		{noDueTask, "this_week", false},
		{noDueTask, "later", true},
		{doneTask, "overdue", false},
		{doneTask, "later", false},
	}
	for _, c := range cases {
		got := taskIsInBucket(c.task, c.bucket)
		if got != c.want {
			t.Errorf("taskIsInBucket(%q, %q) = %v, want %v", c.task.Title, c.bucket, got, c.want)
		}
	}
}
