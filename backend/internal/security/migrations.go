package security

import (
	"log"
	"strings"
)

// migrate creates all required tables if they do not already exist.
// Every statement is idempotent — safe to run on every startup.
func (db *DB) migrate() error {
	migrations := []string{
		`PRAGMA journal_mode=WAL`,
		`PRAGMA busy_timeout=5000`,
		`CREATE TABLE IF NOT EXISTS tasks (
			id TEXT PRIMARY KEY, title TEXT NOT NULL, template TEXT DEFAULT 'task',
			dimension TEXT DEFAULT 'work', created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME, archived INTEGER DEFAULT 0,
			time_block_start DATETIME, time_block_end DATETIME,
			text TEXT DEFAULT '', ritual_cron TEXT DEFAULT '',
			x REAL DEFAULT 0.0, y REAL DEFAULT 0.0,
			files TEXT DEFAULT '[]', modules TEXT DEFAULT '[]',
			parent_id TEXT, meetings TEXT DEFAULT '[]',
			user_id TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_user_id ON tasks(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_parent_id ON tasks(parent_id)`,
		`CREATE TABLE IF NOT EXISTS subtasks (
			id TEXT PRIMARY KEY, parent_id TEXT NOT NULL, title TEXT NOT NULL,
			done INTEGER DEFAULT 0, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, completed_at DATETIME,
			user_id TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE INDEX IF NOT EXISTS idx_subtasks_parent_id ON subtasks(parent_id)`,
		`CREATE INDEX IF NOT EXISTS idx_subtasks_user_id ON subtasks(user_id)`,
		`CREATE TABLE IF NOT EXISTS checklist_items (
			id TEXT PRIMARY KEY, task_id TEXT NOT NULL, label TEXT NOT NULL, done INTEGER DEFAULT 0,
			user_id TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE INDEX IF NOT EXISTS idx_checklist_user_id ON checklist_items(user_id)`,
		`CREATE TABLE IF NOT EXISTS slots (
			id TEXT PRIMARY KEY, name TEXT NOT NULL, start_time TEXT NOT NULL,
			end_time TEXT NOT NULL, recur_rule TEXT DEFAULT 'daily', slot_type TEXT DEFAULT 'hard',
			user_id TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS snapshots (
			id TEXT PRIMARY KEY, node_id TEXT NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP, compressed BLOB NOT NULL,
			user_id TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS media_files (
			id TEXT PRIMARY KEY, original_name TEXT NOT NULL,
			mime_type TEXT DEFAULT 'application/octet-stream', path TEXT NOT NULL, task_id TEXT NOT NULL,
			user_id TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS emergency_sessions (
			id TEXT PRIMARY KEY, user_id TEXT NOT NULL,
			activated_at DATETIME NOT NULL, ends_at DATETIME NOT NULL,
			active INTEGER NOT NULL DEFAULT 1
		)`,
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY, email TEXT NOT NULL, password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_lower ON users(LOWER(email))`,
		// Safely attempt to add new columns to existing tables for backwards compatibility
		`ALTER TABLE tasks ADD COLUMN user_id TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE subtasks ADD COLUMN user_id TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE checklist_items ADD COLUMN user_id TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE slots ADD COLUMN user_id TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE snapshots ADD COLUMN user_id TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE emergency_sessions ADD COLUMN user_id TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE users ADD COLUMN email TEXT NOT NULL DEFAULT ''`,
	}

	for i, stmt := range migrations {
		if _, err := db.Conn.Exec(stmt); err != nil {
			// Ignore errors for ALTER TABLE if column already exists
			log.Printf("migration %d/%d: %v (ignoring if expected)", i+1, len(migrations), err)
		} else {
			log.Printf("migration %d/%d applied", i+1, len(migrations))
		}
	}

	// Specialized check for legacy 'username' constraint issue
	// If users table has a NOT NULL username but we want email-only
	_, err := db.Conn.Exec(`INSERT INTO users (id, email, password_hash) VALUES ('test_probe', 'test@test.com', 'test')`)
	if err != nil && strings.Contains(err.Error(), "NOT NULL constraint failed: users.username") {
		log.Println("⚠️ Legacy username constraint detected. Realigning users table...")
		cleanup := []string{
			`CREATE TABLE users_new (id TEXT PRIMARY KEY, email TEXT NOT NULL, password_hash TEXT NOT NULL, created_at DATETIME DEFAULT CURRENT_TIMESTAMP)`,
			`INSERT INTO users_new (id, email, password_hash, created_at) SELECT id, COALESCE(NULLIF(email, ''), username), password_hash, created_at FROM users`,
			`DROP TABLE users`,
			`ALTER TABLE users_new RENAME TO users`,
			`CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_lower ON users(LOWER(email))`,
		}
		for _, s := range cleanup {
			if _, err := db.Conn.Exec(s); err != nil {
				log.Printf("⚠️ Cleanup error: %v", err)
			}
		}
	} else if err == nil {
		// Clean up the probe
		db.Conn.Exec(`DELETE FROM users WHERE id = 'test_probe'`)
	}

	return nil
}
