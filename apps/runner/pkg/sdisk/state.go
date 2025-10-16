package sdisk

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DB handles local state persistence using SQLite
type DB struct {
	db *sql.DB
}

// DiskState represents a disk's state in the database
type DiskState struct {
	Name       string
	SizeGB     int64
	CreatedAt  time.Time
	ModifiedAt time.Time
	IsMounted  bool
	MountPath  string
	InS3       bool
	Checksum   string
}

// NewDB creates a new state database
func NewDB(dbPath string) (*DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create tables
	if err := createTables(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return &DB{db: db}, nil
}

// createTables creates the database schema
func createTables(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS disks (
		name TEXT PRIMARY KEY,
		size_gb INTEGER NOT NULL,
		created_at TIMESTAMP NOT NULL,
		modified_at TIMESTAMP NOT NULL,
		is_mounted BOOLEAN NOT NULL DEFAULT 0,
		mount_path TEXT,
		in_s3 BOOLEAN NOT NULL DEFAULT 0,
		checksum TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_disks_in_s3 ON disks(in_s3);
	CREATE INDEX IF NOT EXISTS idx_disks_is_mounted ON disks(is_mounted);
	`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	return nil
}

// SaveDisk saves or updates a disk's state
func (db *DB) SaveDisk(state *DiskState) error {
	query := `
	INSERT INTO disks (name, size_gb, created_at, modified_at, is_mounted, mount_path, in_s3, checksum)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(name) DO UPDATE SET
		size_gb = excluded.size_gb,
		modified_at = excluded.modified_at,
		is_mounted = excluded.is_mounted,
		mount_path = excluded.mount_path,
		in_s3 = excluded.in_s3,
		checksum = excluded.checksum
	`

	_, err := db.db.Exec(query,
		state.Name,
		state.SizeGB,
		state.CreatedAt,
		state.ModifiedAt,
		state.IsMounted,
		state.MountPath,
		state.InS3,
		state.Checksum,
	)

	if err != nil {
		return fmt.Errorf("failed to save disk: %w", err)
	}

	return nil
}

// GetDisk retrieves a disks's state
func (db *DB) GetDisk(name string) (*DiskState, error) {
	query := `
	SELECT name, size_gb, created_at, modified_at, is_mounted, mount_path, in_s3, checksum
	FROM disks
	WHERE name = ?
	`

	var state DiskState
	var mountPath sql.NullString
	var checksum sql.NullString

	err := db.db.QueryRow(query, name).Scan(
		&state.Name,
		&state.SizeGB,
		&state.CreatedAt,
		&state.ModifiedAt,
		&state.IsMounted,
		&mountPath,
		&state.InS3,
		&checksum,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get disk: %w", err)
	}

	if mountPath.Valid {
		state.MountPath = mountPath.String
	}
	if checksum.Valid {
		state.Checksum = checksum.String
	}

	return &state, nil
}

// ListDisks returns all disks
func (db *DB) ListDisks() ([]*DiskState, error) {
	query := `
	SELECT name, size_gb, created_at, modified_at, is_mounted, mount_path, in_s3, checksum
	FROM disks
	ORDER BY created_at DESC
	`

	rows, err := db.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list disks: %w", err)
	}
	defer rows.Close()

	var disks []*DiskState
	for rows.Next() {
		var state DiskState
		var mountPath sql.NullString
		var checksum sql.NullString

		if err := rows.Scan(
			&state.Name,
			&state.SizeGB,
			&state.CreatedAt,
			&state.ModifiedAt,
			&state.IsMounted,
			&mountPath,
			&state.InS3,
			&checksum,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if mountPath.Valid {
			state.MountPath = mountPath.String
		}
		if checksum.Valid {
			state.Checksum = checksum.String
		}

		disks = append(disks, &state)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return disks, nil
}

// DeleteDisk removes a disk from the database
func (db *DB) DeleteDisk(name string) error {
	query := `DELETE FROM disks WHERE name = ?`

	result, err := db.db.Exec(query, name)
	if err != nil {
		return fmt.Errorf("failed to delete disk: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("disk not found")
	}

	return nil
}

// UpdateMountState updates the mount state of a disk
func (db *DB) UpdateMountState(name string, isMounted bool, mountPath string) error {
	query := `
	UPDATE disks
	SET is_mounted = ?, mount_path = ?, modified_at = ?
	WHERE name = ?
	`

	result, err := db.db.Exec(query, isMounted, mountPath, time.Now(), name)
	if err != nil {
		return fmt.Errorf("failed to update mount state: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("disk not found")
	}

	return nil
}

// UpdateS3State updates the S3 sync state of a disk
func (db *DB) UpdateS3State(name string, inS3 bool, checksum string) error {
	query := `
	UPDATE disks
	SET in_s3 = ?, checksum = ?, modified_at = ?
	WHERE name = ?
	`

	result, err := db.db.Exec(query, inS3, checksum, time.Now(), name)
	if err != nil {
		return fmt.Errorf("failed to update S3 state: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("disk not found")
	}

	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.db.Close()
}
