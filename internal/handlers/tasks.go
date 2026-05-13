package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"conductor/internal/db"
	"conductor/internal/models"
	"conductor/web/templates"
	"github.com/a-h/templ"
)

func (h *Handler) getAllTasks(w http.ResponseWriter, r *http.Request) {
	filters := parseTaskFilters(r)
	tasks, err := h.db.ListTasks(filters)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	projects, _ := h.db.ListProjects()
	users, _ := h.db.ListUsers()
	layout := h.layoutData(r)
	if err := templates.AllTasks(layout, tasks, projects, users, filters).Render(r.Context(), w); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func (h *Handler) getMyTasks(w http.ResponseWriter, r *http.Request) {
	filters := parseTaskFilters(r)
	filters.AssigneeID = currentUser(r).ID // locked to current user
	tasks, err := h.db.ListTasks(filters)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	layout := h.layoutData(r)
	if err := templates.MyTasks(layout, tasks).Render(r.Context(), w); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func (h *Handler) postCreateTask(w http.ResponseWriter, r *http.Request) {
	title := strings.TrimSpace(r.FormValue("title"))
	if title == "" {
		http.Error(w, "title required", http.StatusBadRequest)
		return
	}
	projectIDStr := r.FormValue("project_id")
	if projectIDStr == "" {
		http.Error(w, "project_id required", http.StatusBadRequest)
		return
	}
	projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid project_id", http.StatusBadRequest)
		return
	}

	status, ok := parseStatus(r.FormValue("status"))
	if !ok {
		http.Error(w, "invalid status", http.StatusBadRequest)
		return
	}
	priority, ok := parsePriority(r.FormValue("priority"))
	if !ok {
		http.Error(w, "invalid priority", http.StatusBadRequest)
		return
	}
	params := db.CreateTaskParams{
		Title:       title,
		Description: r.FormValue("description"),
		Link:        r.FormValue("link"),
		Status:      status,
		Category:    r.FormValue("category"),
		Priority:    priority,
		ProjectID:   projectID,
		AssigneeID:  parseOptionalInt64(r.FormValue("assignee_id")),
		DueDate:     parseOptionalDate(r.FormValue("due_date")),
		CreatedBy:   currentUser(r).ID,
	}
	if _, err := h.db.CreateTask(params); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/projects/%d", projectID), http.StatusSeeOther)
}

func (h *Handler) postUpdateTask(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	task, err := h.db.GetTask(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	if title == "" {
		http.Error(w, "title required", http.StatusBadRequest)
		return
	}

	status, ok := parseStatus(r.FormValue("status"))
	if !ok {
		http.Error(w, "invalid status", http.StatusBadRequest)
		return
	}
	priority, ok := parsePriority(r.FormValue("priority"))
	if !ok {
		http.Error(w, "invalid priority", http.StatusBadRequest)
		return
	}
	params := db.UpdateTaskParams{
		ID:          id,
		Title:       title,
		Description: r.FormValue("description"),
		Link:        r.FormValue("link"),
		Status:      status,
		Category:    r.FormValue("category"),
		Priority:    priority,
		AssigneeID:  parseOptionalInt64(r.FormValue("assignee_id")),
		DueDate:     parseOptionalDate(r.FormValue("due_date")),
	}
	if err := h.db.UpdateTask(params); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	dest := r.Header.Get("Referer")
	if dest == "" {
		dest = fmt.Sprintf("/projects/%d", task.ProjectID)
	}
	http.Redirect(w, r, dest, http.StatusSeeOther)
}

func (h *Handler) postDeleteTask(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	task, err := h.db.GetTask(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := h.db.DeleteTask(id); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	dest := fmt.Sprintf("/projects/%d", task.ProjectID)
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", dest)
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, dest, http.StatusSeeOther)
}

func (h *Handler) postToggleTask(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := h.db.ToggleTask(id); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	task, err := h.db.GetTask(id)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	// On the project detail page, replace the stats and task groups wholesale
	// so the progress bar and section positions update atomically.
	if projectID := projectPageID(r); projectID != 0 {
		project, _ := h.db.GetProject(projectID)
		tasks, _ := h.db.ListTasks(models.TaskFilters{ProjectID: projectID})
		w.Header().Set("HX-Reswap", "none")
		_ = templates.TaskDetailOOB(task).Render(r.Context(), w)
		_ = templates.ProjectStatsOOB(project).Render(r.Context(), w)
		_ = templates.ProjectTaskGroupsOOB(tasks).Render(r.Context(), w)
		user := currentUser(r)
		count, _ := h.db.CountTasks(models.TaskFilters{AssigneeID: user.ID, ExcludeDone: true})
		_ = templates.SidebarMyTaskCountOOB(count).Render(r.Context(), w)
		return
	}
	var row templ.Component
	switch r.URL.Query().Get("ctx") {
	case "all_tasks":
		row = templates.TaskRowWithProject(task)
	case "my_tasks":
		row = templates.MyTaskRow(task)
	default:
		row = templates.TaskRow(task)
	}
	if err := row.Render(r.Context(), w); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	// OOB: refresh sidebar panel if it's showing this task.
	_ = templates.TaskDetailOOB(task).Render(r.Context(), w)
	// OOB: refresh "My tasks" count, except on the My Tasks view where tasks stay
	// visible after being checked (count would be misleading until navigate-away).
	if r.URL.Query().Get("ctx") != "my_tasks" {
		user := currentUser(r)
		count, _ := h.db.CountTasks(models.TaskFilters{AssigneeID: user.ID, ExcludeDone: true})
		_ = templates.SidebarMyTaskCountOOB(count).Render(r.Context(), w)
	}
}

// projectPageID returns the project ID if the request originated from a project
// detail page (/projects/{id}), or 0 otherwise.
func projectPageID(r *http.Request) int64 {
	currentURL := r.Header.Get("HX-Current-URL")
	idx := strings.Index(currentURL, "/projects/")
	if idx == -1 {
		return 0
	}
	rest := currentURL[idx+len("/projects/"):]
	if q := strings.IndexAny(rest, "?#"); q != -1 {
		rest = rest[:q]
	}
	if strings.ContainsRune(rest, '/') {
		return 0
	}
	id, err := strconv.ParseInt(rest, 10, 64)
	if err != nil {
		return 0
	}
	return id
}

func (h *Handler) getTaskRows(w http.ResponseWriter, r *http.Request) {
	filters := parseTaskFilters(r)
	tasks, err := h.db.ListTasks(filters)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	// When called with a project_id the tasks are shown within a project card (no project name needed).
	// Without project_id it's the All Tasks flat list where project context must be visible.
	var rows templ.Component
	if filters.ProjectID != 0 {
		rows = templates.TaskRows(tasks)
	} else {
		rows = templates.TaskRowsWithProject(tasks)
	}
	if err := rows.Render(r.Context(), w); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func (h *Handler) getTaskForm(w http.ResponseWriter, r *http.Request) {
	projects, err := h.db.ListProjects()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	users, _ := h.db.ListUsers()
	categories, _ := h.db.ListCategories()
	data := templates.TaskFormData{
		Projects:   projects,
		Users:      users,
		Categories: categories,
	}
	if pidStr := r.URL.Query().Get("project_id"); pidStr != "" {
		if pid, err := strconv.ParseInt(pidStr, 10, 64); err == nil {
			data.PreselectedProjectID = pid
		}
	}
	if err := templates.TaskForm(data).Render(r.Context(), w); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func (h *Handler) getTaskDetail(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	task, err := h.db.GetTask(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := templates.TaskDetail(task).Render(r.Context(), w); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func (h *Handler) getTaskEditForm(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	task, err := h.db.GetTask(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	projects, _ := h.db.ListProjects()
	users, _ := h.db.ListUsers()
	categories, _ := h.db.ListCategories()
	if err := templates.TaskForm(templates.TaskFormData{Task: &task, Projects: projects, Users: users, Categories: categories}).Render(r.Context(), w); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func (h *Handler) getTaskDeleteConfirm(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	task, err := h.db.GetTask(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := templates.TaskDeleteConfirm(task).Render(r.Context(), w); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

// parseTaskFilters reads common task filter query/form parameters.
func parseTaskFilters(r *http.Request) models.TaskFilters {
	f := models.TaskFilters{}
	if s := r.URL.Query().Get("project_id"); s != "" {
		if id, err := strconv.ParseInt(s, 10, 64); err == nil {
			f.ProjectID = id
		}
	}
	if s := r.URL.Query().Get("assignee_id"); s != "" {
		if id, err := strconv.ParseInt(s, 10, 64); err == nil {
			f.AssigneeID = id
		}
	}
	f.Status = r.URL.Query().Get("status")
	f.Category = r.URL.Query().Get("category")
	f.Priority = r.URL.Query().Get("priority")
	f.Due = r.URL.Query().Get("due")
	return f
}

// parseOptionalInt64 parses a string to *int64, returning nil for empty or invalid input.
func parseOptionalInt64(s string) *int64 {
	if s == "" {
		return nil
	}
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil
	}
	return &id
}

// parseOptionalDate parses a YYYY-MM-DD string to *time.Time, returning nil for empty or invalid input.
func parseOptionalDate(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil
	}
	return &t
}

// parseStatus validates that s is a known Status value.
func parseStatus(s string) (models.Status, bool) {
	switch models.Status(s) {
	case models.StatusTodo, models.StatusInProgress, models.StatusBlocked, models.StatusDone:
		return models.Status(s), true
	}
	return "", false
}

// parsePriority validates that s is a known Priority value.
func parsePriority(s string) (models.Priority, bool) {
	switch models.Priority(s) {
	case models.PriorityCritical, models.PriorityHigh, models.PriorityMedium, models.PriorityLow:
		return models.Priority(s), true
	}
	return "", false
}
