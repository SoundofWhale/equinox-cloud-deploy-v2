package ai

import (
	"database/sql"
	"fmt"
	"time"
)

// NudgeService checks for upcoming personal tasks to remind user during work.
type NudgeService struct {
	DB *sql.DB
}

// NewNudgeService creates a nudge service.
func NewNudgeService(db *sql.DB) *NudgeService {
	return &NudgeService{DB: db}
}

// CheckPersonalNudge looks 4 hours ahead for personal tasks with time blocks for a specific user.
// Returns a nudge if any upcoming personal task is found.
func (ns *NudgeService) CheckPersonalNudge(userID string) *Nudge {
	now := time.Now()
	horizon := now.Add(4 * time.Hour)

	row := ns.DB.QueryRow(
		`SELECT id, title, time_block_start FROM tasks 
		 WHERE dimension = 'personal' 
		   AND user_id = ?
		   AND archived = 0 
		   AND completed_at IS NULL
		   AND time_block_start IS NOT NULL
		   AND time_block_start > ?
		   AND time_block_start < ?
		 ORDER BY time_block_start ASC 
		 LIMIT 1`,
		userID, now, horizon,
	)

	var id, title string
	var tbStart time.Time
	if err := row.Scan(&id, &title, &tbStart); err != nil {
		return nil // no upcoming personal tasks
	}

	remaining := time.Until(tbStart)
	hours := int(remaining.Hours())
	minutes := int(remaining.Minutes()) % 60

	var timeStr string
	if hours > 0 {
		timeStr = fmt.Sprintf("%dч %dмин", hours, minutes)
	} else {
		timeStr = fmt.Sprintf("%dмин", minutes)
	}

	return &Nudge{
		Message: fmt.Sprintf("🌿 Напоминание: «%s» через %s в личном пространстве.", title, timeStr),
		Type:    "inter_context",
		TaskID:  id,
	}
}
