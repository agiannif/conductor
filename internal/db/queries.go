package db

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"conductor/internal/models"
)

// ErrNotFound is returned by Get* methods when a row does not exist.
var ErrNotFound = errors.New("not found")

// SessionTTL is the duration for which a session is valid (sliding window).
const SessionTTL = 30 * 24 * time.Hour

// priorityOrder is the SQL CASE expression for ordering tasks by priority.
const priorityOrder = `CASE priority
	WHEN 'critical' THEN 1
	WHEN 'high'     THEN 2
	WHEN 'medium'   THEN 3
	WHEN 'low'      THEN 4
	ELSE 5
END`

// ---- Users ----

func (d *DB) CreateUser(username, passwordHash string) (int64, error) {
	res, err := d.sql.Exec(
		`INSERT INTO users (username, password_hash) VALUES (?, ?)`,
		username, passwordHash,
	)
	if err != nil {
		return 0, fmt.Errorf("CreateUser: %w", err)
	}
	return res.LastInsertId()
}

func (d *DB) GetUserByUsername(username string) (models.User, string, error) {
	var u models.User
	var hash string
	err := d.sql.QueryRow(
		`SELECT id, username, password_hash, created_at FROM users WHERE username = ?`,
		username,
	).Scan(&u.ID, &u.Username, &hash, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return u, "", fmt.Errorf("user not found: %w", ErrNotFound)
	}
	return u, hash, err
}

func (d *DB) GetUserByID(id int64) (models.User, string, error) {
	var u models.User
	var hash string
	err := d.sql.QueryRow(
		`SELECT id, username, password_hash, created_at FROM users WHERE id = ?`, id,
	).Scan(&u.ID, &u.Username, &hash, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return u, "", fmt.Errorf("user not found: %w", ErrNotFound)
	}
	return u, hash, err
}

func (d *DB) ListUsers() ([]models.User, error) {
	rows, err := d.sql.Query(`SELECT id, username, created_at FROM users ORDER BY username`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Username, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (d *DB) DeleteUser(userID int64) error {
	_, err := d.sql.Exec(`DELETE FROM users WHERE id = ?`, userID)
	return err
}

func (d *DB) UpdateUserPassword(userID int64, passwordHash string) error {
	res, err := d.sql.Exec(
		`UPDATE users SET password_hash = ? WHERE id = ?`,
		passwordHash, userID,
	)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("user not found: %w", ErrNotFound)
	}
	return nil
}

// ---- Sessions ----

func (d *DB) CreateSession(userID int64) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("rand: %w", err)
	}
	token := hex.EncodeToString(b)
	expiresAt := time.Now().Add(SessionTTL)
	_, err := d.sql.Exec(
		`INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)`,
		token, userID, expiresAt,
	)
	return token, err
}

func (d *DB) GetSession(token string) (models.Session, error) {
	var s models.Session
	err := d.sql.QueryRow(
		`SELECT id, user_id, expires_at FROM sessions WHERE id = ? AND expires_at > datetime('now')`,
		token,
	).Scan(&s.ID, &s.UserID, &s.ExpiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return s, fmt.Errorf("session not found or expired: %w", ErrNotFound)
	}
	return s, err
}

func (d *DB) ExtendSession(token string) error {
	expiresAt := time.Now().Add(SessionTTL)
	_, err := d.sql.Exec(
		`UPDATE sessions SET expires_at = ? WHERE id = ?`,
		expiresAt, token,
	)
	return err
}

func (d *DB) DeleteSession(token string) error {
	_, err := d.sql.Exec(`DELETE FROM sessions WHERE id = ?`, token)
	return err
}

func (d *DB) DeleteExpiredSessions() error {
	_, err := d.sql.Exec(`DELETE FROM sessions WHERE expires_at <= datetime('now')`)
	return err
}

// ---- Projects ----

func (d *DB) CreateProject(name, description string) (int64, error) {
	res, err := d.sql.Exec(
		`INSERT INTO projects (name, description) VALUES (?, ?)`,
		name, description,
	)
	if err != nil {
		return 0, fmt.Errorf("CreateProject: %w", err)
	}
	return res.LastInsertId()
}

func (d *DB) GetProject(id int64) (models.Project, error) {
	var p models.Project
	err := d.sql.QueryRow(`
		SELECT p.id, p.name, COALESCE(p.description, ''), p.created_at,
		       COUNT(t.id) AS total_tasks,
		       SUM(CASE WHEN t.status = 'done' THEN 1 ELSE 0 END) AS done_tasks
		FROM projects p
		LEFT JOIN tasks t ON t.project_id = p.id
		WHERE p.id = ?
		GROUP BY p.id
	`, id).Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.TotalTasks, &p.DoneTasks)
	if errors.Is(err, sql.ErrNoRows) {
		return p, fmt.Errorf("project not found: %w", ErrNotFound)
	}
	return p, err
}

func (d *DB) ListProjects() ([]models.Project, error) {
	rows, err := d.sql.Query(`
		SELECT p.id, p.name, COALESCE(p.description, ''), p.created_at,
		       COUNT(t.id) AS total_tasks,
		       SUM(CASE WHEN t.status = 'done' THEN 1 ELSE 0 END) AS done_tasks
		FROM projects p
		LEFT JOIN tasks t ON t.project_id = p.id
		GROUP BY p.id
		ORDER BY p.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var projects []models.Project
	for rows.Next() {
		var p models.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.TotalTasks, &p.DoneTasks); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

func (d *DB) UpdateProject(id int64, name, description string) error {
	res, err := d.sql.Exec(
		`UPDATE projects SET name = ?, description = ? WHERE id = ?`,
		name, description, id,
	)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("project not found: %w", ErrNotFound)
	}
	return nil
}

func (d *DB) DeleteProject(id int64) error {
	_, err := d.sql.Exec(`DELETE FROM projects WHERE id = ?`, id)
	return err
}

func (d *DB) ProjectTaskCount(id int64) (int, error) {
	var count int
	err := d.sql.QueryRow(`SELECT COUNT(*) FROM tasks WHERE project_id = ?`, id).Scan(&count)
	return count, err
}

func (d *DB) ProjectNonDoneTaskCount(id int64) (int, error) {
	var count int
	err := d.sql.QueryRow(`SELECT COUNT(*) FROM tasks WHERE project_id = ? AND status != 'done'`, id).Scan(&count)
	return count, err
}

func (d *DB) DeleteProjectTasks(id int64) error {
	_, err := d.sql.Exec(`DELETE FROM tasks WHERE project_id = ?`, id)
	return err
}

func (d *DB) CountProjects() (int, error) {
	var count int
	err := d.sql.QueryRow(`SELECT COUNT(*) FROM projects`).Scan(&count)
	return count, err
}

// CountTasks returns the number of tasks matching the given filters.
func (d *DB) CountTasks(f models.TaskFilters) (int, error) {
	where := []string{"1=1"}
	args := []any{}

	if f.ProjectID != 0 {
		where = append(where, "project_id = ?")
		args = append(args, f.ProjectID)
	}
	if f.Status != "" {
		where = append(where, "status = ?")
		args = append(args, f.Status)
	}
	if f.AssigneeID != 0 {
		where = append(where, "assignee_id = ?")
		args = append(args, f.AssigneeID)
	}
	if f.ExcludeDone {
		where = append(where, "status != 'done'")
	}

	var count int
	query := fmt.Sprintf(`SELECT COUNT(*) FROM tasks WHERE %s`, strings.Join(where, " AND "))
	err := d.sql.QueryRow(query, args...).Scan(&count)
	return count, err
}

// ---- Tasks ----

type CreateTaskParams struct {
	Title       string
	Description string
	Link        string
	Status      models.Status
	Category    string
	Priority    models.Priority
	ProjectID   int64
	AssigneeID  *int64
	DueDate     *time.Time
	CreatedBy   int64
}

type UpdateTaskParams struct {
	ID          int64
	Title       string
	Description string
	Link        string
	Status      models.Status
	Category    string
	Priority    models.Priority
	AssigneeID  *int64
	DueDate     *time.Time
}

func (d *DB) CreateTask(p CreateTaskParams) (int64, error) {
	res, err := d.sql.Exec(`
		INSERT INTO tasks
			(title, description, link, status, category, priority, project_id, assignee_id, due_date, created_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.Title, p.Description, p.Link, string(p.Status), p.Category, string(p.Priority),
		p.ProjectID, p.AssigneeID, p.DueDate, p.CreatedBy,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// scanTask scans a task row (with joined project/assignee/creator columns)
// from any scanner. Both GetTask and ListTasks share this logic.
func scanTask(scan func(...any) error) (models.Task, error) {
	var t models.Task
	var assigneeID sql.NullInt64
	var assigneeName sql.NullString
	var createdBy sql.NullInt64
	var createdByName sql.NullString
	var dueDate sql.NullTime

	if err := scan(
		&t.ID, &t.Title, &t.Description, &t.Link,
		&t.Status, &t.Category, &t.Priority,
		&t.ProjectID, &t.ProjectName,
		&assigneeID, &assigneeName,
		&dueDate,
		&createdBy, &createdByName,
		&t.CreatedAt, &t.UpdatedAt,
	); err != nil {
		return t, err
	}
	if assigneeID.Valid {
		t.AssigneeID = &assigneeID.Int64
		t.AssigneeName = assigneeName.String
	}
	if dueDate.Valid {
		t.DueDate = &dueDate.Time
	}
	if createdBy.Valid {
		t.CreatedBy = &createdBy.Int64
		t.CreatedByName = createdByName.String
	}
	return t, nil
}

func (d *DB) GetTask(id int64) (models.Task, error) {
	row := d.sql.QueryRow(`
		SELECT t.id, t.title, COALESCE(t.description,''), COALESCE(t.link,''),
		       t.status, t.category, t.priority,
		       t.project_id, p.name,
		       t.assignee_id, a.username,
		       t.due_date,
		       t.created_by, u.username,
		       t.created_at, t.updated_at
		FROM tasks t
		JOIN projects p ON p.id = t.project_id
		LEFT JOIN users a ON a.id = t.assignee_id
		LEFT JOIN users u ON u.id = t.created_by
		WHERE t.id = ?
	`, id)
	t, err := scanTask(row.Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return t, fmt.Errorf("task not found: %w", ErrNotFound)
	}
	return t, err
}

// ListTasks returns tasks matching the given filters, sorted by due_date asc (nulls last), then priority.
func (d *DB) ListTasks(f models.TaskFilters) ([]models.Task, error) {
	where := []string{"1=1"}
	args := []any{}

	if f.ProjectID != 0 {
		where = append(where, "t.project_id = ?")
		args = append(args, f.ProjectID)
	}
	if f.Status != "" {
		where = append(where, "t.status = ?")
		args = append(args, f.Status)
	}
	if f.Category != "" {
		where = append(where, "t.category = ?")
		args = append(args, f.Category)
	}
	if f.Priority != "" {
		where = append(where, "t.priority = ?")
		args = append(args, f.Priority)
	}
	if f.AssigneeID != 0 {
		where = append(where, "t.assignee_id = ?")
		args = append(args, f.AssigneeID)
	}
	switch f.Due {
	case "overdue":
		where = append(where, "t.due_date < date('now') AND t.status != 'done'")
	case "this_week":
		where = append(where, "t.due_date >= date('now') AND t.due_date <= date('now', '+6 days')")
	case "this_month":
		where = append(where, "t.due_date >= date('now') AND t.due_date <= date('now', 'start of month', '+1 month', '-1 day')")
	}

	query := fmt.Sprintf(`
		SELECT t.id, t.title, COALESCE(t.description,''), COALESCE(t.link,''),
		       t.status, t.category, t.priority,
		       t.project_id, p.name,
		       t.assignee_id, a.username,
		       t.due_date,
		       t.created_by, u.username,
		       t.created_at, t.updated_at
		FROM tasks t
		JOIN projects p ON p.id = t.project_id
		LEFT JOIN users a ON a.id = t.assignee_id
		LEFT JOIN users u ON u.id = t.created_by
		WHERE %s
		ORDER BY
			CASE WHEN t.due_date IS NULL THEN 1 ELSE 0 END,
			t.due_date ASC,
			%s
	`, strings.Join(where, " AND "), priorityOrder)

	rows, err := d.sql.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		t, err := scanTask(rows.Scan)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func (d *DB) UpdateTask(p UpdateTaskParams) error {
	res, err := d.sql.Exec(`
		UPDATE tasks SET
			title = ?, description = ?, link = ?, status = ?,
			category = ?, priority = ?, assignee_id = ?, due_date = ?
		WHERE id = ?`,
		p.Title, p.Description, p.Link, string(p.Status),
		p.Category, string(p.Priority), p.AssigneeID, p.DueDate, p.ID,
	)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("task not found: %w", ErrNotFound)
	}
	return nil
}

func (d *DB) DeleteTask(id int64) error {
	_, err := d.sql.Exec(`DELETE FROM tasks WHERE id = ?`, id)
	return err
}

// ToggleTask flips a task between done and todo. Any non-done status (todo,
// in progress, blocked) becomes done; done becomes todo. This is intentionally
// lossy — a task that was "in progress" will be "todo" after un-toggling.
func (d *DB) ToggleTask(id int64) error {
	res, err := d.sql.Exec(`
		UPDATE tasks SET status = CASE
			WHEN status = 'done' THEN 'todo'
			ELSE 'done'
		END
		WHERE id = ?`, id,
	)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("task not found: %w", ErrNotFound)
	}
	return nil
}

func (d *DB) ListCategories() ([]string, error) {
	rows, err := d.sql.Query(`SELECT DISTINCT category FROM tasks WHERE category != '' ORDER BY category`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cats []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, rows.Err()
}
