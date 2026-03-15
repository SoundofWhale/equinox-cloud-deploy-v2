package snapshot

import (
	"bytes"
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

// Snapshot is a compressed point-in-time backup of a task subtree.
type Snapshot struct {
	ID        string    `json:"id"`
	NodeID    string    `json:"node_id"`
	CreatedAt time.Time `json:"created_at"`
	Payload   []byte    `json:"-"` // gzip-compressed JSON
}

// SnapshotService manages daily snapshots with retention policy.
type SnapshotService struct {
	DB        *sql.DB
	MaxPerNode int // max snapshots retained per node (default 30)
}

// NewSnapshotService creates a service with 30-snapshot retention.
func NewSnapshotService(db *sql.DB) *SnapshotService {
	return &SnapshotService{DB: db, MaxPerNode: 30}
}

// TakeSnapshot captures the current state of a node as gzip JSON.
func (s *SnapshotService) TakeSnapshot(userID, nodeID string, payload interface{}) error {
	// Serialize to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	// Compress with gzip
	compressed, err := gzipData(jsonData)
	if err != nil {
		return fmt.Errorf("gzip: %w", err)
	}

	snap := Snapshot{
		ID:        uuid.New().String(),
		NodeID:    nodeID,
		CreatedAt: time.Now(),
		Payload:   compressed,
	}

	_, err = s.DB.Exec(
		`INSERT INTO snapshots (id, node_id, timestamp, compressed, user_id) VALUES (?, ?, ?, ?, ?)`,
		snap.ID, snap.NodeID, snap.CreatedAt, snap.Payload, userID,
	)
	if err != nil {
		return fmt.Errorf("insert snapshot: %w", err)
	}

	// Prune old snapshots beyond retention limit
	if err := s.pruneNode(userID, nodeID); err != nil {
		log.Printf("⚠️ prune snapshots for %s: %v", nodeID, err)
	}

	log.Printf("📸 Snapshot taken for node %s (%d bytes compressed)", nodeID, len(compressed))
	return nil
}

// ListSnapshots returns all snapshots for a node, newest first.
func (s *SnapshotService) ListSnapshots(userID, nodeID string) ([]Snapshot, error) {
	rows, err := s.DB.Query(
		`SELECT id, node_id, timestamp FROM snapshots WHERE node_id = ? AND user_id = ? ORDER BY timestamp DESC`,
		nodeID, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	snaps := []Snapshot{}
	for rows.Next() {
		var snap Snapshot
		if err := rows.Scan(&snap.ID, &snap.NodeID, &snap.CreatedAt); err != nil {
			return nil, err
		}
		snaps = append(snaps, snap)
	}
	return snaps, nil
}

// RestoreSnapshot decompresses and returns the payload of a snapshot.
func (s *SnapshotService) RestoreSnapshot(userID, snapshotID string) ([]byte, error) {
	var compressed []byte
	err := s.DB.QueryRow(`SELECT compressed FROM snapshots WHERE id = ? AND user_id = ?`, snapshotID, userID).Scan(&compressed)
	if err != nil {
		return nil, fmt.Errorf("snapshot not found: %w", err)
	}
	return gunzipData(compressed)
}

// pruneNode removes excess snapshots beyond MaxPerNode for a given node.
func (s *SnapshotService) pruneNode(userID, nodeID string) error {
	_, err := s.DB.Exec(
		`DELETE FROM snapshots WHERE node_id = ? AND user_id = ? AND id NOT IN (
			SELECT id FROM snapshots WHERE node_id = ? AND user_id = ? ORDER BY timestamp DESC LIMIT ?
		)`,
		nodeID, userID, nodeID, userID, s.MaxPerNode,
	)
	return err
}

// StartScheduler starts a cron job that takes snapshots daily at midnight.
// The taskFetcher function should return all active node IDs and their payloads.
func (s *SnapshotService) StartScheduler(taskFetcher func() (map[string]interface{}, error)) {
	c := cron.New()
	c.AddFunc("0 0 * * *", func() { // midnight every day
		data, err := taskFetcher()
		if err != nil {
			log.Printf("❌ Snapshot scheduler: fetch failed: %v", err)
			return
		}
		for nodeID, payload := range data {
			if err := s.TakeSnapshot("", nodeID, payload); err != nil {
				log.Printf("❌ Snapshot failed for %s: %v", nodeID, err)
			}
		}
		log.Printf("📸 Daily snapshot complete: %d nodes", len(data))
	})
	c.Start()
	log.Println("📸 Snapshot scheduler started (daily midnight)")
}

// ── Compression helpers ────────────────────────────────────────────

func gzipData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write(data); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func gunzipData(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}
