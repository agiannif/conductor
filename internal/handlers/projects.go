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
	layoutData := h.layoutData(r)
	if err := templates.Projects(layoutData, projects).Render(r.Context(), w); err != nil {
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
	tasks, err := h.db.ListTasks(models.TaskFilters{ProjectID: id})
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	users, _ := h.db.ListUsers()
	layoutData := h.layoutData(r)
	if err := templates.Project(layoutData, project, tasks, users).Render(r.Context(), w); err != nil {
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
	count, _ := h.db.ProjectTaskCount(id)
	if count > 0 {
		// Return blocked response — HTMX will swap modal content with this message
		w.WriteHeader(http.StatusConflict)
		fmt.Fprintf(w, `<p class="text-red-600">This project has %d task(s). Delete or reassign them first.</p>`, count)
		return
	}
	if err := h.db.DeleteProject(id); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
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

// layoutData builds the sidebar counts for the layout template.
func (h *Handler) layoutData(r *http.Request) templates.LayoutData {
	user := currentUser(r)
	users, _ := h.db.ListUsers()
	projects, _ := h.db.ListProjects()
	allTasks, _ := h.db.ListTasks(models.TaskFilters{})
	myTasks, _ := h.db.ListTasks(models.TaskFilters{AssigneeID: user.ID})
	return templates.LayoutData{
		User:         user,
		Users:        users,
		ProjectCount: len(projects),
		AllTaskCount: len(allTasks),
		MyTaskCount:  len(myTasks),
	}
}
