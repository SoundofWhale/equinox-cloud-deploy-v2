package arbiter

import (
	"database/sql"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// EmergencySession represents an active emergency mode session.
type EmergencySession struct {
	ID          string    `json:"id"`
	ActivatedAt time.Time `json:"activated_at"`
	EndsAt      time.Time `json:"ends_at"`
	Active      bool      `json:"active"`
}

const emergencyDuration = 4 * time.Hour

// EmergencyManager handles the 4-hour emergency mode timer per user.
type EmergencyManager struct {
	mu       sync.Mutex
	sessions map[string]*EmergencySession // map of userID -> EmergencySession
	db       *sql.DB
	onExpire func() // callback when emergency expires
}

// NewEmergencyManager creates a manager that watches emergency sessions.
func NewEmergencyManager(db *sql.DB, onExpire func()) *EmergencyManager {
	em := &EmergencyManager{
		db:       db,
		onExpire: func() {}, // default empty callback
		sessions: make(map[string]*EmergencySession),
	}
	if onExpire != nil {
		em.onExpire = onExpire
	}

	// Create table if not exists with user_id
	db.Exec(`CREATE TABLE IF NOT EXISTS emergency_sessions (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		activated_at DATETIME NOT NULL,
		ends_at DATETIME NOT NULL,
		active INTEGER NOT NULL DEFAULT 1
	)`)

	// Check for active sessions on startup
	em.loadAllActive()
	return em
}

// Activate starts a new 4-hour emergency session for a specific user.
func (em *EmergencyManager) Activate(userID string) *EmergencySession {
	em.mu.Lock()
	defer em.mu.Unlock()

	// Deactivate any current session for this user
	if current, exists := em.sessions[userID]; exists && current.Active {
		em.deactivateUnsafe(current.ID, userID)
	}

	now := time.Now()
	session := &EmergencySession{
		ID:          uuid.New().String(),
		ActivatedAt: now,
		EndsAt:      now.Add(emergencyDuration),
		Active:      true,
	}

	em.db.Exec(
		`INSERT INTO emergency_sessions (id, user_id, activated_at, ends_at, active) VALUES (?, ?, ?, ?, 1)`,
		session.ID, userID, session.ActivatedAt, session.EndsAt,
	)

	em.sessions[userID] = session

	// Start timer goroutine
	go em.watchTimer(session, userID)

	log.Printf("🚨 Emergency mode activated for user %s: %s (expires %s)", userID, session.ID, session.EndsAt.Format("15:04:05"))
	return session
}

// Status returns the current emergency session for a specific user.
func (em *EmergencyManager) Status(userID string) *EmergencySession {
	em.mu.Lock()
	defer em.mu.Unlock()
	return em.sessions[userID]
}

// watchTimer waits until the session expires, then triggers the callback.
func (em *EmergencyManager) watchTimer(session *EmergencySession, userID string) {
	remaining := time.Until(session.EndsAt)
	if remaining <= 0 {
		em.expire(session.ID, userID)
		return
	}

	timer := time.NewTimer(remaining)
	<-timer.C

	em.expire(session.ID, userID)
}

func (em *EmergencyManager) expire(id string, userID string) {
	em.mu.Lock()
	defer em.mu.Unlock()

	em.deactivateUnsafe(id, userID)
	log.Printf("⏰ Emergency mode expired for user %s: %s → Equi cough state", userID, id)

	if em.onExpire != nil {
		em.onExpire()
	}
}

func (em *EmergencyManager) deactivateUnsafe(id string, userID string) {
	em.db.Exec(`UPDATE emergency_sessions SET active = 0 WHERE id = ? AND user_id = ?`, id, userID)
	if current, exists := em.sessions[userID]; exists && current.ID == id {
		current.Active = false
		delete(em.sessions, userID)
	}
}

func (em *EmergencyManager) loadAllActive() {
	rows, err := em.db.Query(`SELECT id, user_id, activated_at, ends_at, active FROM emergency_sessions WHERE active = 1`)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var s EmergencySession
		var userID string
		var active int
		err := rows.Scan(&s.ID, &userID, &s.ActivatedAt, &s.EndsAt, &active)
		if err != nil {
			continue
		}
		s.Active = active == 1

		if s.Active && time.Now().Before(s.EndsAt) {
			em.sessions[userID] = &s
			go em.watchTimer(&s, userID) // resume timer
			log.Printf("🚨 Resumed emergency session for user %s: %s (expires %s)", userID, s.ID, s.EndsAt.Format("15:04:05"))
		} else if s.Active {
			// Already expired — mark as inactive
			em.deactivateUnsafe(s.ID, userID)
		}
	}
}
