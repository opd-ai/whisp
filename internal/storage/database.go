package storage

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

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
	GetDatabaseKeyBytes() ([]byte, error)
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

	var db *sql.DB
	var err error

	if encrypted {
		// Use SQLCipher for encrypted database - use minimal DSN due to v4 compatibility issues
		dsn = fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=cipher_compatibility(3)", dbPath)
		db, err = sql.Open("sqlite3", dsn)
	} else {
		// Use regular SQLite for unencrypted database (fallback)
		dsn = fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)", dbPath)
		db, err = sql.Open("sqlite3", dsn)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set encryption key if security manager is provided
	if encrypted {
		// Try using raw key bytes with SQLCipher v3 compatibility mode
		keyBytes, err := securityManager.GetDatabaseKeyBytes()
		if err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to get database key bytes: %w", err)
		}
		defer func() {
			// Clear key bytes from memory
			for i := range keyBytes {
				keyBytes[i] = 0
			}
		}()

		// Convert to hex format for SQLCipher PRAGMA key
		hexKey := fmt.Sprintf("%x", keyBytes)

		// Set the encryption key using PRAGMA key
		// Note: cipher_compatibility(3) should already be set in DSN
		if _, err := db.Exec("PRAGMA key = '" + hexKey + "'"); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to set database encryption key: %w", err)
		}

		// Test database access by attempting a simple query
		// This will fail if encryption setup is incorrect
		var result int
		err = db.QueryRow("SELECT 1").Scan(&result)
		if err != nil {
			// Database might be new, try to create schema first
			_, createErr := db.Exec("CREATE TABLE IF NOT EXISTS _test_encryption (id INTEGER)")
			if createErr != nil {
				db.Close()
				return nil, fmt.Errorf("failed to verify database encryption: %w", err)
			}
			// Clean up test table
			db.Exec("DROP TABLE IF EXISTS _test_encryption")
		}
	} // Test connection
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

	// Run migrations
	if err := storage.runMigrations(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
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
		// For WAL mode, ensure all transactions are committed
		// Only do this for unencrypted databases that use WAL mode
		if !d.encrypted {
			if _, err := d.db.Exec("PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
				log.Printf("Warning: Failed to checkpoint WAL: %v", err)
			}
		}

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
		uuid TEXT UNIQUE NOT NULL,
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
	CREATE INDEX IF NOT EXISTS idx_messages_uuid ON messages(uuid);
	CREATE INDEX IF NOT EXISTS idx_contacts_friend_id ON contacts(friend_id);
	CREATE INDEX IF NOT EXISTS idx_file_transfers_friend_id ON file_transfers(friend_id);
	`

	_, err := d.db.Exec(schema)
	return err
}

// runMigrations runs database migrations to handle schema changes
func (d *Database) runMigrations() error {
	// Create migrations table if it doesn't exist
	migrationsSchema := `
	CREATE TABLE IF NOT EXISTS migrations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		version TEXT UNIQUE NOT NULL,
		applied_at DATETIME NOT NULL
	);`

	if _, err := d.db.Exec(migrationsSchema); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Define migrations
	migrations := []struct {
		version string
		sql     string
	}{
		{
			version: "add_uuid_to_messages",
			sql: `
			-- Check if uuid column exists, if not add it
			PRAGMA table_info(messages);
			-- This is a more complex migration that requires checking column existence
			-- For SQLite, we need to use a different approach
			`,
		},
		{
			version: "add_fts_message_search",
			sql: `
			-- Create FTS virtual table for message search optimization
			CREATE VIRTUAL TABLE IF NOT EXISTS messages_fts USING fts5(
				content,
				content_row='messages',
				content_rowid='id'
			);
			
			-- Populate FTS table with existing messages
			INSERT INTO messages_fts(rowid, content) 
			SELECT id, content FROM messages WHERE is_deleted = 0;
			
			-- Create triggers to keep FTS table synchronized with messages table
			CREATE TRIGGER IF NOT EXISTS messages_fts_insert AFTER INSERT ON messages BEGIN
				INSERT INTO messages_fts(rowid, content) VALUES (new.id, new.content);
			END;
			
			CREATE TRIGGER IF NOT EXISTS messages_fts_update AFTER UPDATE ON messages BEGIN
				DELETE FROM messages_fts WHERE rowid = old.id;
				INSERT INTO messages_fts(rowid, content) VALUES (new.id, new.content);
			END;
			
			CREATE TRIGGER IF NOT EXISTS messages_fts_delete AFTER DELETE ON messages BEGIN
				DELETE FROM messages_fts WHERE rowid = old.id;
			END;
			`,
		},
	}

	// Apply migrations
	for _, migration := range migrations {
		// Check if migration was already applied
		var count int
		err := d.db.QueryRow("SELECT COUNT(*) FROM migrations WHERE version = ?", migration.version).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check migration status: %w", err)
		}

		if count > 0 {
			continue // Migration already applied
		}

		// Special handling for UUID migration
		if migration.version == "add_uuid_to_messages" {
			if err := d.migrateAddUUIDToMessages(); err != nil {
				return fmt.Errorf("failed to apply UUID migration: %w", err)
			}
		} else if migration.version == "add_fts_message_search" {
			if err := d.migrateFTSMessageSearch(); err != nil {
				return fmt.Errorf("failed to apply FTS migration: %w", err)
			}
		} else {
			// Apply regular migration
			if _, err := d.db.Exec(migration.sql); err != nil {
				return fmt.Errorf("failed to apply migration %s: %w", migration.version, err)
			}
		}

		// Record migration as applied
		if _, err := d.db.Exec("INSERT INTO migrations (version, applied_at) VALUES (?, ?)",
			migration.version, "datetime('now')"); err != nil {
			return fmt.Errorf("failed to record migration: %w", err)
		}

		log.Printf("Applied migration: %s", migration.version)
	}

	return nil
}

// migrateAddUUIDToMessages adds UUID column to messages table if it doesn't exist
func (d *Database) migrateAddUUIDToMessages() error {
	// Check if uuid column already exists
	rows, err := d.db.Query("PRAGMA table_info(messages)")
	if err != nil {
		return fmt.Errorf("failed to get table info: %w", err)
	}
	defer rows.Close()

	hasUUID := false
	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, primaryKey int
		var defaultValue sql.NullString

		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &primaryKey); err != nil {
			return fmt.Errorf("failed to scan column info: %w", err)
		}

		if name == "uuid" {
			hasUUID = true
			break
		}
	}

	if hasUUID {
		return nil // UUID column already exists
	}

	// Add UUID column with a default value, then update existing rows
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Add the column with a temporary default
	if _, err := tx.Exec("ALTER TABLE messages ADD COLUMN uuid TEXT"); err != nil {
		return fmt.Errorf("failed to add uuid column: %w", err)
	}

	// Generate UUIDs for existing messages
	rows, err = tx.Query("SELECT id FROM messages WHERE uuid IS NULL OR uuid = ''")
	if err != nil {
		return fmt.Errorf("failed to query existing messages: %w", err)
	}

	var messageIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return fmt.Errorf("failed to scan message ID: %w", err)
		}
		messageIDs = append(messageIDs, id)
	}
	rows.Close()

	// Update each message with a UUID
	for _, id := range messageIDs {
		uuid := generateUUID()
		if _, err := tx.Exec("UPDATE messages SET uuid = ? WHERE id = ?", uuid, id); err != nil {
			return fmt.Errorf("failed to update message UUID: %w", err)
		}
	}

	// Create the unique constraint on uuid column (SQLite doesn't support adding constraints)
	// We'll create a unique index instead
	if _, err := tx.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_messages_uuid_unique ON messages(uuid)"); err != nil {
		return fmt.Errorf("failed to create unique index on uuid: %w", err)
	}

	return tx.Commit()
}

// migrateFTSMessageSearch creates the FTS virtual table and associated triggers for optimized message search
func (d *Database) migrateFTSMessageSearch() error {
	// First check if FTS5 is available
	if !d.isFTS5Available() {
		log.Printf("FTS5 not available, skipping FTS migration")
		return nil
	}

	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start FTS migration transaction: %w", err)
	}
	defer tx.Rollback()

	// Create FTS virtual table for message search optimization
	ftsSchema := `
	CREATE VIRTUAL TABLE IF NOT EXISTS messages_fts USING fts5(
		content,
		content_row='messages',
		content_rowid='id'
	);`

	if _, err := tx.Exec(ftsSchema); err != nil {
		return fmt.Errorf("failed to create FTS virtual table: %w", err)
	}

	// Populate FTS table with existing messages
	populateQuery := `
	INSERT INTO messages_fts(rowid, content) 
	SELECT id, content FROM messages WHERE is_deleted = 0 AND content IS NOT NULL AND content != '';`

	if _, err := tx.Exec(populateQuery); err != nil {
		return fmt.Errorf("failed to populate FTS table: %w", err)
	}

	// Create triggers to keep FTS table synchronized with messages table
	insertTrigger := `
	CREATE TRIGGER IF NOT EXISTS messages_fts_insert AFTER INSERT ON messages
	WHEN NEW.is_deleted = 0 AND NEW.content IS NOT NULL AND NEW.content != ''
	BEGIN
		INSERT INTO messages_fts(rowid, content) VALUES (NEW.id, NEW.content);
	END;`

	updateTrigger := `
	CREATE TRIGGER IF NOT EXISTS messages_fts_update AFTER UPDATE ON messages
	WHEN NEW.content IS NOT NULL AND NEW.content != ''
	BEGIN
		DELETE FROM messages_fts WHERE rowid = OLD.id;
		CASE WHEN NEW.is_deleted = 0 THEN
			INSERT INTO messages_fts(rowid, content) VALUES (NEW.id, NEW.content);
		END;
	END;`

	deleteTrigger := `
	CREATE TRIGGER IF NOT EXISTS messages_fts_delete AFTER DELETE ON messages BEGIN
		DELETE FROM messages_fts WHERE rowid = OLD.id;
	END;`

	for _, trigger := range []string{insertTrigger, updateTrigger, deleteTrigger} {
		if _, err := tx.Exec(trigger); err != nil {
			return fmt.Errorf("failed to create FTS trigger: %w", err)
		}
	}

	log.Printf("FTS5 message search index created successfully")
	return tx.Commit()
}

// isFTS5Available checks if SQLite has FTS5 module available
func (d *Database) isFTS5Available() bool {
	// Try to compile a simple FTS5 statement
	_, err := d.db.Exec(`SELECT 1 FROM pragma_compile_options WHERE compile_options LIKE '%FTS5%'`)
	if err != nil {
		return false
	}

	// Try creating a temporary FTS5 table
	_, err = d.db.Exec(`CREATE VIRTUAL TABLE IF NOT EXISTS _temp_fts_test USING fts5(content)`)
	if err != nil {
		return false
	}

	// Clean up test table
	d.db.Exec(`DROP TABLE IF EXISTS _temp_fts_test`)
	return true
}

// generateUUID generates a simple UUID string without external dependencies
func generateUUID() string {
	// Use a simple time-based approach for migration purposes
	// In real code, the message manager uses proper UUID library
	return fmt.Sprintf("msg_%d_%d", time.Now().UnixNano(), rand.Int63())
}

// ensureDir creates a directory if it doesn't exist
func ensureDir(dir string) error {
	return os.MkdirAll(dir, 0700)
}
