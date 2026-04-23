package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// DB wraps sql.DB with application-specific methods.
type DB struct {
	sql *sql.DB
}

// Open opens the SQLite database at path, sets required PRAGMAs, and runs migrations.
func Open(path string) (*DB, error) {
	sqlDB, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// Single writer connection; WAL allows concurrent reads.
	sqlDB.SetMaxOpenConns(1)

	d := &DB{sql: sqlDB}
	if err := d.configure(); err != nil {
		sqlDB.Close()
		return nil, err
	}
	if err := d.migrate(); err != nil {
		sqlDB.Close()
		return nil, err
	}
	return d, nil
}

func (d *DB) Close() error {
	return d.sql.Close()
}

func (d *DB) configure() error {
	pragmas := []string{
		`PRAGMA journal_mode=WAL`,
		`PRAGMA foreign_keys=ON`,
	}
	for _, p := range pragmas {
		if _, err := d.sql.Exec(p); err != nil {
			return fmt.Errorf("pragma %q: %w", p, err)
		}
	}
	return nil
}

const schema = `
CREATE TABLE IF NOT EXISTS users (
	id            INTEGER PRIMARY KEY AUTOINCREMENT,
	username      TEXT UNIQUE NOT NULL CHECK(length(username) <= 64),
	password_hash TEXT NOT NULL,
	created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sessions (
	id         TEXT PRIMARY KEY,
	user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	expires_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS projects (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	name        TEXT NOT NULL CHECK(length(name) <= 255),
	description TEXT,
	created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tasks (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	title       TEXT NOT NULL CHECK(length(title) <= 255),
	description TEXT,
	link        TEXT CHECK(length(link) <= 2048),
	status      TEXT NOT NULL CHECK(status IN ('todo', 'in progress', 'blocked', 'done')),
	category    TEXT NOT NULL DEFAULT '' CHECK(length(category) <= 64),
	priority    TEXT NOT NULL CHECK(priority IN ('critical', 'high', 'medium', 'low')),
	project_id  INTEGER NOT NULL REFERENCES projects(id) ON DELETE RESTRICT,
	assignee_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
	due_date    DATE,
	created_by  INTEGER REFERENCES users(id) ON DELETE SET NULL,
	created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER IF NOT EXISTS tasks_updated_at
AFTER UPDATE ON tasks FOR EACH ROW BEGIN
	UPDATE tasks SET updated_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;
`

func (d *DB) migrate() error {
	if _, err := d.sql.Exec(schema); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return nil
}
