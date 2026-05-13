package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"conductor/internal/models"
	"conductor/web/templates"
)

func (h *Handler) getProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := h.db.ListProjects()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	users, _ := h.db.ListUsers()
	filters := parseTaskFilters(r)
	layoutData := h.layoutData(r)
	if err := templates.Projects(layoutData, projects, users, filters).Render(r.Context(), w); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func (h *Handler) getProject(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	project, err := h.db.GetProject(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	filters := parseTaskFilters(r)
	filters.ProjectID = id
	tasks, err := h.db.ListTasks(filters)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	users, _ := h.db.ListUsers()
	layoutData := h.layoutData(r)
	if err := templates.Project(layoutData, project, tasks, users, filters).Render(r.Context(), w); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func (h *Handler) postCreateProject(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.FormValue("name"))
	description := strings.TrimSpace(r.FormValue("description"))
	if name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}
	id, err := h.db.CreateProject(name, description)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/projects/%d", id), http.StatusSeeOther)
}

func (h *Handler) postUpdateProject(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	description := strings.TrimSpace(r.FormValue("description"))
	if name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}
	if err := h.db.UpdateProject(id, name, description); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/projects/%d", id), http.StatusSeeOther)
}

func (h *Handler) postDeleteProject(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	count, err := h.db.ProjectNonDoneTaskCount(id)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if count > 0 {
		if r.Header.Get("HX-Request") == "true" {
			_ = templates.ProjectDeleteError(count).Render(r.Context(), w)
			return
		}
		w.WriteHeader(http.StatusConflict)
		_ = templates.ProjectDeleteError(count).Render(r.Context(), w)
		return
	}
	if err := h.db.DeleteProjectTasks(id); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if err := h.db.DeleteProject(id); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) getProjectForm(w http.ResponseWriter, r *http.Request) {
	if err := templates.ProjectForm(templates.ProjectFormData{}).Render(r.Context(), w); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func (h *Handler) getProjectEditForm(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	project, err := h.db.GetProject(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := templates.ProjectForm(templates.ProjectFormData{Project: &project}).Render(r.Context(), w); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func (h *Handler) getProjectDeleteConfirm(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	project, err := h.db.GetProject(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := templates.ProjectDeleteConfirm(project).Render(r.Context(), w); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

// layoutData builds the sidebar counts for the layout template.
func (h *Handler) layoutData(r *http.Request) templates.LayoutData {
	user := currentUser(r)
	users, _ := h.db.ListUsers()
	projectCount, _ := h.db.CountProjects()
	allTaskCount, _ := h.db.CountTasks(models.TaskFilters{})
	myTaskCount, _ := h.db.CountTasks(models.TaskFilters{AssigneeID: user.ID, ExcludeDone: true})
	return templates.LayoutData{
		User:         user,
		Users:        users,
		ProjectCount: projectCount,
		AllTaskCount: allTaskCount,
		MyTaskCount:  myTaskCount,
		CurrentPath:  r.URL.Path,
	}
}
