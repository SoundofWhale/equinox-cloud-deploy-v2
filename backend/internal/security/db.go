package security

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mutecomm/go-sqlcipher/v4"
)

// DB wraps a SQLCipher-encrypted database connection.
type DB struct {
	Conn *sql.DB
	Path string
}

// OpenDB opens (or creates) a SQLCipher‑encrypted database at the given path.
// The key must be a hex-encoded 32-byte AES-256 key.
func OpenDB(dbPath string, hexKey string) (*DB, error) {
	dsn := fmt.Sprintf("file:%s?_pragma_key=x'%s'&_pragma_cipher_page_size=4096&_pragma_kdf_iter=256000&_pragma_synchronous=NORMAL&_pragma_cache_size=-64000",
		dbPath, hexKey)

	conn, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	// Verify the connection actually works (catches wrong key immediately).
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ping db (wrong key?): %w", err)
	}

	db := &DB{Conn: conn, Path: dbPath}

	// Run migrations on every open — they are idempotent.
	if err := db.migrate(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	log.Printf("SQLCipher DB opened: %s (page_size=4096, kdf_iter=256000)", dbPath)
	return db, nil
}

// CipherVersion returns the linked SQLCipher library version string.
func (db *DB) CipherVersion() (string, error) {
	var version string
	err := db.Conn.QueryRow("PRAGMA cipher_version;").Scan(&version)
	return version, err
}

// CipherPageSize returns the current cipher_page_size pragma value.
func (db *DB) CipherPageSize() (int, error) {
	var size int
	err := db.Conn.QueryRow("PRAGMA cipher_page_size;").Scan(&size)
	return size, err
}

// Close closes the underlying database connection.
func (db *DB) Close() error {
	return db.Conn.Close()
}
