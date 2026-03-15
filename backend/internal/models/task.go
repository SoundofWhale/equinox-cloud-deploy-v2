package models

import "time"

// TaskTemplate defines the type of task based on its module composition.
type TaskTemplate string

const (
	TemplateTask    TaskTemplate = "task"    // Title + TimeBlock
	TemplateNote    TaskTemplate = "note"    // Title + Text
	TemplateMeeting TaskTemplate = "meeting" // Title + TimeBlock + Checklist
	TemplateRitual  TaskTemplate = "ritual"  // Title + RitualCron
)

// MaxDepth is the maximum allowed nesting level for subtask hierarchies.
const MaxDepth = 42

// DefaultChildModules are the modules assigned to newly created child tasks.
var DefaultChildModules = []string{ModuleDescription, ModuleChecklist, ModuleAIAdvice}

// Task represents a top-level task (planet in Work dimension, branch in Personal).
type Task struct {
	ID          string       `json:"id"`
	UserID      string       `json:"user_id"`
	Title       string       `json:"title"`
	Template    TaskTemplate `json:"template"`
	Dimension   string       `json:"dimension"` // "work" or "personal"
	CreatedAt   time.Time    `json:"created_at"`
	CompletedAt *time.Time   `json:"completed_at,omitempty"`
	Archived    bool         `json:"archived"`
	ParentID    *string      `json:"parent_id,omitempty"`

	// Canvas Coordinates
	X float64 `json:"x"`
	Y float64 `json:"y"`

	// Optional modules
	Modules   []string    `json:"modules"`
	TimeBlock *TimeBlock  `json:"time_block,omitempty"`
	Checklist []CheckItem `json:"checklist,omitempty"`
	Text      string      `json:"text,omitempty"`
	RitualCron string     `json:"ritual_cron,omitempty"`

	// Relationships
	Subtasks []Subtask `json:"subtasks,omitempty"`
	Files    []string  `json:"files"`
	Meetings []Meeting `json:"meetings"`

	// Computed hierarchy fields
	Depth         int `json:"depth"`
	ChildrenCount int `json:"children_count"`
}

const (
	ModuleDescription = "description"
	ModuleChecklist   = "checklist"
	ModuleAttachments = "attachments"
	ModuleAIAdvice    = "ai_advice"
	ModuleMeetings    = "meetings"
)

// Subtask is a satellite (Work) or sub-branch (Personal) attached to a Task.
type Subtask struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	ParentID    string     `json:"parent_id"`
	Title       string     `json:"title"`
	Done        bool       `json:"done"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// Meeting represents a scheduled event within a task.
type Meeting struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

// TimeBlock represents a scheduled window or deadline.
type TimeBlock struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// CheckItem is a single swipeable checklist entry.
type CheckItem struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
	Label  string `json:"label"`
	Done   bool   `json:"done"`
}

// MediaFile represents an encrypted media attachment stored in /vault/media/.
type MediaFile struct {
	ID           string `json:"id"` // UUID — also the filename in vault
	UserID       string `json:"user_id"`
	OriginalName string `json:"original_name"`
	MimeType     string `json:"mime_type"`
	Path         string `json:"path"` // relative path inside vault
	TaskID       string `json:"task_id"`
}
