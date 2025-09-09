package storage

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mutecomm/go-sqlcipher/v4"
)

// Database wraps the SQLite database connection with encryption support
type Database struct {
	db        *sql.DB
	path      string
	encrypted bool
}

// SecurityManager interface for database encryption
type SecurityManager interface {
	GetDatabaseKey() (string, error)
}

// NewDatabase creates a new database connection
func NewDatabase(dbPath string) (*Database, error) {
	return NewDatabaseWithEncryption(dbPath, nil)
}

// NewDatabaseWithEncryption creates a new encrypted database connection
func NewDatabaseWithEncryption(dbPath string, securityManager SecurityManager) (*Database, error) {
	// Ensure directory exists
	if err := ensureDir(filepath.Dir(dbPath)); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	var dsn string
	encrypted := securityManager != nil

	if encrypted {
		// Use SQLCipher for encrypted database
		dsn = fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)", dbPath)
	} else {
		// Use regular SQLite for unencrypted database (fallback)
		dsn = fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)", dbPath)
	}

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set encryption key if security manager is provided
	if encrypted {
		key, err := securityManager.GetDatabaseKey()
		if err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to get database key: %w", err)
		}

		// Set the encryption key using PRAGMA key
		if _, err := db.Exec(fmt.Sprintf("PRAGMA key = %s", key)); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to set database encryption key: %w", err)
		}

		// Verify encryption by trying to read from sqlite_master
		var count int
		err = db.QueryRow("SELECT count(*) FROM sqlite_master").Scan(&count)
		if err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to verify database encryption (wrong key?): %w", err)
		}
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	storage := &Database{
		db:        db,
		path:      dbPath,
		encrypted: encrypted,
	}

	// Initialize schema
	if err := storage.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	encryptionStatus := "unencrypted"
	if encrypted {
		encryptionStatus = "encrypted"
	}
	log.Printf("Database initialized at %s (%s)", dbPath, encryptionStatus)
	return storage, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// IsEncrypted returns whether the database is encrypted
func (d *Database) IsEncrypted() bool {
	return d.encrypted
}

// GetPath returns the database file path
func (d *Database) GetPath() string {
	return d.path
}

// Query executes a query that returns rows
func (d *Database) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return d.db.Query(query, args...)
}

// QueryRow executes a query that returns a single row
func (d *Database) QueryRow(query string, args ...interface{}) *sql.Row {
	return d.db.QueryRow(query, args...)
}

// Exec executes a query that doesn't return rows
func (d *Database) Exec(query string, args ...interface{}) (sql.Result, error) {
	return d.db.Exec(query, args...)
}

// Begin starts a transaction
func (d *Database) Begin() (*sql.Tx, error) {
	return d.db.Begin()
}

// initSchema initializes the database schema
func (d *Database) initSchema() error {
	schema := `
	-- Contacts table
	CREATE TABLE IF NOT EXISTS contacts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tox_id TEXT UNIQUE,
		public_key BLOB NOT NULL,
		friend_id INTEGER UNIQUE NOT NULL,
		name TEXT NOT NULL DEFAULT '',
		status_message TEXT NOT NULL DEFAULT '',
		avatar BLOB,
		status INTEGER NOT NULL DEFAULT 0,
		is_blocked BOOLEAN NOT NULL DEFAULT 0,
		is_favorite BOOLEAN NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		last_seen_at DATETIME NOT NULL,
		UNIQUE(public_key)
	);

	-- Messages table
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		friend_id INTEGER NOT NULL,
		content TEXT NOT NULL,
		message_type INTEGER NOT NULL DEFAULT 0,
		is_outgoing BOOLEAN NOT NULL,
		timestamp DATETIME NOT NULL,
		delivered_at DATETIME,
		read_at DATETIME,
		edited_at DATETIME,
		original_content TEXT,
		file_path TEXT,
		file_size INTEGER,
		file_type TEXT,
		is_deleted BOOLEAN NOT NULL DEFAULT 0,
		reply_to_id INTEGER,
		FOREIGN KEY (friend_id) REFERENCES contacts(friend_id),
		FOREIGN KEY (reply_to_id) REFERENCES messages(id)
	);

	-- Settings table
	CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		updated_at DATETIME NOT NULL
	);

	-- File transfers table
	CREATE TABLE IF NOT EXISTS file_transfers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		friend_id INTEGER NOT NULL,
		file_name TEXT NOT NULL,
		file_size INTEGER NOT NULL,
		file_path TEXT,
		is_outgoing BOOLEAN NOT NULL,
		status INTEGER NOT NULL DEFAULT 0,
		progress INTEGER NOT NULL DEFAULT 0,
		started_at DATETIME NOT NULL,
		completed_at DATETIME,
		message_id INTEGER,
		FOREIGN KEY (friend_id) REFERENCES contacts(friend_id),
		FOREIGN KEY (message_id) REFERENCES messages(id)
	);

	-- Create indexes
	CREATE INDEX IF NOT EXISTS idx_messages_friend_id ON messages(friend_id);
	CREATE INDEX IF NOT EXISTS idx_messages_timestamp ON messages(timestamp);
	CREATE INDEX IF NOT EXISTS idx_contacts_friend_id ON contacts(friend_id);
	CREATE INDEX IF NOT EXISTS idx_file_transfers_friend_id ON file_transfers(friend_id);
	`

	_, err := d.db.Exec(schema)
	return err
}

// ensureDir creates a directory if it doesn't exist
func ensureDir(dir string) error {
	return os.MkdirAll(dir, 0700)
}
