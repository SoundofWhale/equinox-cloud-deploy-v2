package snapshot

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "github.com/mutecomm/go-sqlcipher/v4"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test_snap.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	// Create snapshots table
	db.Exec(`CREATE TABLE IF NOT EXISTS snapshots (
		id TEXT PRIMARY KEY,
		node_id TEXT NOT NULL,
		timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		compressed BLOB NOT NULL
	)`)

	return db
}

func TestTakeAndRestoreSnapshot(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	svc := NewSnapshotService(db)

	// Take a snapshot with sample data
	payload := map[string]interface{}{
		"id":    "task-1",
		"title": "Test Task",
		"subtasks": []map[string]string{
			{"id": "sub-1", "title": "Subtask A"},
		},
	}

	err := svc.TakeSnapshot("task-1", payload)
	if err != nil {
		t.Fatalf("TakeSnapshot failed: %v", err)
	}

	// List snapshots
	snaps, err := svc.ListSnapshots("task-1")
	if err != nil {
		t.Fatalf("ListSnapshots failed: %v", err)
	}
	if len(snaps) != 1 {
		t.Fatalf("expected 1 snapshot, got %d", len(snaps))
	}
	t.Logf("✅ Snapshot created: %s at %s", snaps[0].ID, snaps[0].CreatedAt)

	// Restore
	restored, err := svc.RestoreSnapshot(snaps[0].ID)
	if err != nil {
		t.Fatalf("RestoreSnapshot failed: %v", err)
	}
	if len(restored) == 0 {
		t.Fatal("restored payload is empty")
	}
	t.Logf("✅ Restored %d bytes of JSON", len(restored))
}

func TestPruneExcessSnapshots(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	svc := NewSnapshotService(db)
	svc.MaxPerNode = 3 // low limit for testing

	// Take 5 snapshots
	for i := 0; i < 5; i++ {
		err := svc.TakeSnapshot("node-prune", map[string]int{"i": i})
		if err != nil {
			t.Fatalf("snapshot %d failed: %v", i, err)
		}
	}

	// Should only have 3 remaining (pruned 2)
	snaps, _ := svc.ListSnapshots("node-prune")
	if len(snaps) != 3 {
		t.Fatalf("expected 3 snapshots after pruning, got %d", len(snaps))
	}
	t.Logf("✅ Pruning works: 5 taken, 3 retained (max=%d)", svc.MaxPerNode)
}

func TestGzipCompression(t *testing.T) {
	original := []byte(`{"title":"Hello World","subtasks":[{"id":"1"},{"id":"2"},{"id":"3"}]}`)

	compressed, err := gzipData(original)
	if err != nil {
		t.Fatalf("gzip failed: %v", err)
	}

	decompressed, err := gunzipData(compressed)
	if err != nil {
		t.Fatalf("gunzip failed: %v", err)
	}

	if string(decompressed) != string(original) {
		t.Fatalf("round-trip mismatch")
	}
	t.Logf("✅ Gzip round-trip: %d → %d bytes (%.0f%% compression)",
		len(original), len(compressed), float64(len(compressed))/float64(len(original))*100)
}
