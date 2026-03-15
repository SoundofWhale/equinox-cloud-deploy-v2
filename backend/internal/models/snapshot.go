package models

import "time"

// Snapshot represents a compressed point-in-time backup of a single node (branch/planet).
type Snapshot struct {
	ID         string    `json:"id"`
	NodeID     string    `json:"node_id"`   // task ID this snapshot belongs to
	Timestamp  time.Time `json:"timestamp"`
	Compressed []byte    `json:"-"` // gzip JSON payload — not exposed in JSON API
}
