package models

// SlotType determines whether a time slot is a hard block or a soft warning.
type SlotType string

const (
	SlotHard SlotType = "hard" // Blocks work task creation — never overridden
	SlotSoft SlotType = "soft" // Warning shown — user can confirm override
)

// Slot represents a protected time block in the conflict engine.
type Slot struct {
	ID        string   `json:"id"`
	UserID    string   `json:"user_id"`
	Name      string   `json:"name"`
	StartTime string   `json:"start_time"` // "HH:MM" format
	EndTime   string   `json:"end_time"`   // "HH:MM" format
	RecurRule string   `json:"recur_rule"` // "daily", "weekly:Sat", etc.
	Type      SlotType `json:"type"`
}

// DefaultHardSlots returns the 3 default hard slots seeded at first boot.
func DefaultHardSlots() []Slot {
	return []Slot{
		{
			Name:      "Sleep",
			StartTime: "23:00",
			EndTime:   "07:00",
			RecurRule: "daily",
			Type:      SlotHard,
		},
		{
			Name:      "Family Time",
			StartTime: "18:00",
			EndTime:   "20:00",
			RecurRule: "daily",
			Type:      SlotHard,
		},
		{
			Name:      "Rest",
			StartTime: "13:00",
			EndTime:   "18:00",
			RecurRule: "weekly:Sat",
			Type:      SlotHard,
		},
	}
}
