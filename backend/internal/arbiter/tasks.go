package arbiter

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"equinox/internal/models"

	"github.com/google/uuid"
)

// TaskService provides CRUD operations with conflict checking.
type TaskService struct {
	DB *sql.DB
}

// NewTaskService creates a task service connected to the given DB.
func NewTaskService(db *sql.DB) *TaskService {
	return &TaskService{DB: db}
}

// GetTaskDepth calculates the depth of a task by traversing parent_id chain.
// Root tasks (no parent) have depth 0.
func (s *TaskService) GetTaskDepth(userID, taskID string) (int, error) {
	depth := 0
	currentID := taskID
	for depth <= models.MaxDepth {
		var parentID sql.NullString
		err := s.DB.QueryRow(`SELECT parent_id FROM tasks WHERE id = ? AND user_id = ?`, currentID, userID).Scan(&parentID)
		if err != nil {
			return 0, nil // task not found = root
		}
		if !parentID.Valid || parentID.String == "" {
			return depth, nil
		}
		depth++
		currentID = parentID.String
	}
	return depth, fmt.Errorf("depth exceeds maximum of %d", models.MaxDepth)
}

// GetChildren retrieves all direct child tasks for a given parent ID.
func (s *TaskService) GetChildren(userID, parentID string) ([]models.Task, error) {
	rows, err := s.DB.Query(
		`SELECT id, title, template, dimension, created_at, completed_at, archived,
		        time_block_start, time_block_end, text, ritual_cron, x, y, files, modules, parent_id, meetings
		 FROM tasks WHERE parent_id = ? AND user_id = ? AND archived = 0 ORDER BY created_at ASC`, parentID, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	children := []models.Task{}
	for rows.Next() {
		var t models.Task
		var completedAt sql.NullTime
		var tbStart, tbEnd sql.NullTime
		var filesJSON sql.NullString
		var modulesJSON sql.NullString
		var meetingsJSON sql.NullString

		err := rows.Scan(&t.ID, &t.Title, &t.Template, &t.Dimension, &t.CreatedAt,
			&completedAt, &t.Archived, &tbStart, &tbEnd, &t.Text, &t.RitualCron, &t.X, &t.Y, &filesJSON, &modulesJSON, &t.ParentID, &meetingsJSON)
		if err != nil {
			return nil, err
		}

		if completedAt.Valid {
			t.CompletedAt = &completedAt.Time
		}
		if tbStart.Valid && tbEnd.Valid {
			t.TimeBlock = &models.TimeBlock{Start: tbStart.Time, End: tbEnd.Time}
		}
		if filesJSON.Valid && filesJSON.String != "" {
			json.Unmarshal([]byte(filesJSON.String), &t.Files)
		}
		if t.Files == nil {
			t.Files = []string{}
		}
		if modulesJSON.Valid && modulesJSON.String != "" {
			json.Unmarshal([]byte(modulesJSON.String), &t.Modules)
		}
		if t.Modules == nil {
			t.Modules = []string{}
		}
		if meetingsJSON.Valid && meetingsJSON.String != "" {
			json.Unmarshal([]byte(meetingsJSON.String), &t.Meetings)
		}
		if t.Meetings == nil {
			t.Meetings = []models.Meeting{}
		}

		// Count grandchildren
		var count int
		s.DB.QueryRow(`SELECT COUNT(*) FROM tasks WHERE parent_id = ? AND user_id = ? AND archived = 0`, t.ID, userID).Scan(&count)
		t.ChildrenCount = count

		// Fetch subtasks
		subRows, err := s.DB.Query(`SELECT id, parent_id, title, done, created_at FROM subtasks WHERE parent_id = ? AND user_id = ?`, t.ID, userID)
		if err == nil {
			for subRows.Next() {
				var sub models.Subtask
				if err := subRows.Scan(&sub.ID, &sub.ParentID, &sub.Title, &sub.Done, &sub.CreatedAt); err == nil {
					t.Subtasks = append(t.Subtasks, sub)
				}
			}
			subRows.Close()
		}

		children = append(children, t)
	}

	return children, nil
}

// CreateTask validates against conflict slots, then inserts the task.
func (s *TaskService) CreateTask(userID string, task models.Task, slots []models.Slot) (*ConflictResult, error) {
	// Title is always required
	if task.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	// Depth check for child tasks
	if task.ParentID != nil && *task.ParentID != "" {
		parentDepth, err := s.GetTaskDepth(userID, *task.ParentID)
		if err != nil {
			return nil, fmt.Errorf("depth check failed: %w", err)
		}
		if parentDepth+1 > models.MaxDepth {
			return nil, fmt.Errorf("maximum nesting depth of %d exceeded", models.MaxDepth)
		}
	} else {
		// Ensure it's nil if it's empty string
		task.ParentID = nil
	}

	// If task has a TimeBlock, run conflict check — filter slots by user too
	userSlots := []models.Slot{}
	for _, sl := range slots {
		if sl.UserID == userID {
			userSlots = append(userSlots, sl)
		}
	}

	if task.TimeBlock != nil && !task.TimeBlock.Start.IsZero() {
		result := CheckConflict(task.TimeBlock.Start, task.TimeBlock.End, userSlots)
		if result.HasConflict && result.Blocking {
			return &result, nil // blocked by hard slot
		}
		if result.HasConflict {
			return &result, nil // soft warning — let frontend decide
		}
	}

	// Generate ID if empty
	if task.ID == "" {
		task.ID = uuid.New().String()
	}
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}

	// Insert into DB
	var tbStart, tbEnd *time.Time
	if task.TimeBlock != nil {
		tbStart = &task.TimeBlock.Start
		tbEnd = &task.TimeBlock.End
	}

	// Default modules if none provided
	if len(task.Modules) == 0 {
		task.Modules = []string{
			models.ModuleDescription,
			models.ModuleChecklist,
			models.ModuleAttachments,
			models.ModuleMeetings,
			models.ModuleAIAdvice,
		}
	}

	var filesJSON []byte
	filesJSON, _ = json.Marshal(task.Files)
	if string(filesJSON) == "null" {
		filesJSON = []byte("[]")
	}

	var modulesJSON []byte
	modulesJSON, _ = json.Marshal(task.Modules)
	if string(modulesJSON) == "null" {
		modulesJSON = []byte("[]")
	}

	var meetingsJSON []byte
	meetingsJSON, _ = json.Marshal(task.Meetings)
	if string(meetingsJSON) == "null" {
		meetingsJSON = []byte("[]")
	}

	_, err := s.DB.Exec(
		`INSERT INTO tasks (id, title, template, dimension, created_at, time_block_start, time_block_end, text, ritual_cron, x, y, files, modules, parent_id, meetings, user_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		task.ID, task.Title, task.Template, task.Dimension, task.CreatedAt,
		tbStart, tbEnd, task.Text, task.RitualCron, task.X, task.Y, string(filesJSON), string(modulesJSON), task.ParentID, string(meetingsJSON), userID,
	)
	if err != nil {
		return nil, fmt.Errorf("insert task: %w", err)
	}

	// Insert subtasks
	for _, sub := range task.Subtasks {
		if sub.ID == "" {
			sub.ID = uuid.New().String()
		}
		s.DB.Exec(
			`INSERT INTO subtasks (id, parent_id, title, done, created_at, user_id) VALUES (?, ?, ?, ?, ?, ?)`,
			sub.ID, task.ID, sub.Title, false, time.Now(), userID,
		)
	}

	// Insert checklist items
	for _, item := range task.Checklist {
		if item.ID == "" {
			item.ID = uuid.New().String()
		}
		s.DB.Exec(
			`INSERT INTO checklist_items (id, task_id, label, done, user_id) VALUES (?, ?, ?, ?, ?)`,
			item.ID, task.ID, item.Label, false, userID,
		)
	}

	return nil, nil
}

// GetTaskByID retrieves a single task by its ID with all related data.
func (s *TaskService) GetTaskByID(userID, id string) (*models.Task, error) {
	var t models.Task
	var completedAt sql.NullTime
	var tbStart, tbEnd sql.NullTime
	var filesJSON sql.NullString
	var modulesJSON sql.NullString
	var meetingsJSON sql.NullString
	var parentID sql.NullString

	err := s.DB.QueryRow(
		`SELECT id, title, template, dimension, created_at, completed_at, archived,
		        time_block_start, time_block_end, text, ritual_cron, x, y, files, modules, parent_id, meetings
		 FROM tasks WHERE id = ? AND user_id = ? AND archived = 0`, id, userID,
	).Scan(&t.ID, &t.Title, &t.Template, &t.Dimension, &t.CreatedAt,
		&completedAt, &t.Archived, &tbStart, &tbEnd, &t.Text, &t.RitualCron, &t.X, &t.Y, &filesJSON, &modulesJSON, &parentID, &meetingsJSON)
	if err != nil {
		return nil, err
	}

	if completedAt.Valid {
		t.CompletedAt = &completedAt.Time
	}
	if tbStart.Valid && tbEnd.Valid {
		t.TimeBlock = &models.TimeBlock{Start: tbStart.Time, End: tbEnd.Time}
	}
	if filesJSON.Valid && filesJSON.String != "" {
		json.Unmarshal([]byte(filesJSON.String), &t.Files)
	}
	if t.Files == nil {
		t.Files = []string{}
	}
	if modulesJSON.Valid && modulesJSON.String != "" {
		json.Unmarshal([]byte(modulesJSON.String), &t.Modules)
	}
	if t.Modules == nil {
		t.Modules = []string{}
	}
	if meetingsJSON.Valid && meetingsJSON.String != "" {
		json.Unmarshal([]byte(meetingsJSON.String), &t.Meetings)
	}
	if t.Meetings == nil {
		t.Meetings = []models.Meeting{}
	}
	if parentID.Valid {
		t.ParentID = &parentID.String
	}

	// Fetch subtasks
	subRows, err := s.DB.Query(`SELECT id, parent_id, title, done, created_at FROM subtasks WHERE parent_id = ? AND user_id = ?`, id, userID)
	if err == nil {
		defer subRows.Close()
		for subRows.Next() {
			var sub models.Subtask
			if err := subRows.Scan(&sub.ID, &sub.ParentID, &sub.Title, &sub.Done, &sub.CreatedAt); err == nil {
				t.Subtasks = append(t.Subtasks, sub)
			}
		}
	}

	// Count children
	var count int
	s.DB.QueryRow(`SELECT COUNT(*) FROM tasks WHERE parent_id = ? AND user_id = ? AND archived = 0`, id, userID).Scan(&count)
	t.ChildrenCount = count

	return &t, nil
}

// GetAllTasks retrieves all non-archived tasks.
func (s *TaskService) GetAllTasks(userID string) ([]models.Task, error) {
	rows, err := s.DB.Query(
		`SELECT id, title, template, dimension, created_at, completed_at, archived, 
		        time_block_start, time_block_end, text, ritual_cron, x, y, files, modules, parent_id, meetings
		 FROM tasks WHERE user_id = ? AND archived = 0 ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []models.Task{}
	taskIDs := []string{}
	taskMap := make(map[string]*models.Task)

	for rows.Next() {
		var t models.Task
		var completedAt sql.NullTime
		var tbStart, tbEnd sql.NullTime
		var filesJSON sql.NullString
		var modulesJSON sql.NullString
		var meetingsJSON sql.NullString

		err := rows.Scan(&t.ID, &t.Title, &t.Template, &t.Dimension, &t.CreatedAt,
			&completedAt, &t.Archived, &tbStart, &tbEnd, &t.Text, &t.RitualCron, &t.X, &t.Y, &filesJSON, &modulesJSON, &t.ParentID, &meetingsJSON)
		if err != nil {
			log.Printf("❌ GetAllTasks Scan error: %v", err)
			return nil, err
		}

		if completedAt.Valid {
			t.CompletedAt = &completedAt.Time
		}
		if tbStart.Valid && tbEnd.Valid {
			t.TimeBlock = &models.TimeBlock{Start: tbStart.Time, End: tbEnd.Time}
		}
		if filesJSON.Valid && filesJSON.String != "" {
			json.Unmarshal([]byte(filesJSON.String), &t.Files)
		}
		if t.Files == nil {
			t.Files = []string{}
		}
		if modulesJSON.Valid && modulesJSON.String != "" {
			json.Unmarshal([]byte(modulesJSON.String), &t.Modules)
		}
		if t.Modules == nil {
			t.Modules = []string{}
		}
		if meetingsJSON.Valid && meetingsJSON.String != "" {
			json.Unmarshal([]byte(meetingsJSON.String), &t.Meetings)
		}
		if t.Meetings == nil {
			t.Meetings = []models.Meeting{}
		}

		tasks = append(tasks, t)
		taskIDs = append(taskIDs, t.ID)
		taskMap[t.ID] = &tasks[len(tasks)-1]
	}
	rows.Close()

	// Re-map tasks to pointer map for attaching subtasks
	for i := range tasks {
		taskMap[tasks[i].ID] = &tasks[i]
	}

	// Fetch all relevant subtasks in one go if any tasks exist
	if len(taskIDs) > 0 {
		subRows, err := s.DB.Query(
			`SELECT id, parent_id, title, done, created_at FROM subtasks WHERE user_id = ?`, userID,
		)
		if err == nil {
			defer subRows.Close()
			for subRows.Next() {
				var sub models.Subtask
				err := subRows.Scan(&sub.ID, &sub.ParentID, &sub.Title, &sub.Done, &sub.CreatedAt)
				if err == nil {
					if parentTask, exists := taskMap[sub.ParentID]; exists {
						parentTask.Subtasks = append(parentTask.Subtasks, sub)
					}
				}
			}
		}
	}

	// Compute children_count for each task (how many tasks point to it via parent_id)
	for i := range tasks {
		count := 0
		for _, t := range tasks {
			if t.ParentID != nil && *t.ParentID == tasks[i].ID {
				count++
			}
		}
		tasks[i].ChildrenCount = count
	}

	return tasks, nil
}

// GetAllSlots retrieves all conflict slots from the DB for a specific user.
func (s *TaskService) GetAllSlots(userID string) ([]models.Slot, error) {
	rows, err := s.DB.Query(`SELECT id, name, start_time, end_time, recur_rule, slot_type FROM slots WHERE user_id = ?`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	slots := []models.Slot{}
	for rows.Next() {
		var slot models.Slot
		var slotType string
		err := rows.Scan(&slot.ID, &slot.Name, &slot.StartTime, &slot.EndTime, &slot.RecurRule, &slotType)
		if err != nil {
			return nil, err
		}
		slot.Type = models.SlotType(slotType)
		slot.UserID = userID
		slots = append(slots, slot)
	}
	return slots, nil
}

// CompleteTask marks a task as completed.
func (s *TaskService) CompleteTask(userID, id string) error {
	now := time.Now()
	_, err := s.DB.Exec(`UPDATE tasks SET completed_at = ? WHERE id = ? AND user_id = ?`, now, id, userID)
	return err
}

// UpdateTask updates title, text and coordinates of a task
func (s *TaskService) UpdateTask(userID string, task models.Task) error {
	var filesJSON []byte
	filesJSON, _ = json.Marshal(task.Files)
	if string(filesJSON) == "null" {
		filesJSON = []byte("[]")
	}

	var modulesJSON []byte
	modulesJSON, _ = json.Marshal(task.Modules)
	if string(modulesJSON) == "null" {
		modulesJSON = []byte("[]")
	}

	var meetingsJSON []byte
	meetingsJSON, _ = json.Marshal(task.Meetings)
	if string(meetingsJSON) == "null" {
		meetingsJSON = []byte("[]")
	}

	if task.ParentID != nil && *task.ParentID == "" {
		task.ParentID = nil
	}

	_, err := s.DB.Exec(`UPDATE tasks SET title = ?, text = ?, x = ?, y = ?, files = ?, modules = ?, parent_id = ?, meetings = ? WHERE id = ? AND user_id = ?`,
		task.Title, task.Text, task.X, task.Y, string(filesJSON), string(modulesJSON), task.ParentID, string(meetingsJSON), task.ID, userID)
	return err
}

// AddSubtask adds a subtask to a task
func (s *TaskService) AddSubtask(userID, parentID, title string) (*models.Subtask, error) {
	sub := models.Subtask{
		ID:        uuid.New().String(),
		ParentID:  parentID,
		Title:     title,
		Done:      false,
		CreatedAt: time.Now(),
	}
	_, err := s.DB.Exec(
		`INSERT INTO subtasks (id, parent_id, title, done, created_at, user_id) VALUES (?, ?, ?, ?, ?, ?)`,
		sub.ID, sub.ParentID, sub.Title, sub.Done, sub.CreatedAt, userID,
	)
	return &sub, err
}

// ToggleSubtask toggles the done status
func (s *TaskService) ToggleSubtask(userID, id string) error {
	_, err := s.DB.Exec(`UPDATE subtasks SET done = NOT done WHERE id = ? AND user_id = ?`, id, userID)
	return err
}

// EditSubtask changes the title
func (s *TaskService) EditSubtask(userID, id, title string) error {
	_, err := s.DB.Exec(`UPDATE subtasks SET title = ? WHERE id = ? AND user_id = ?`, title, id, userID)
	return err
}

// DeleteSubtask removes it
func (s *TaskService) DeleteSubtask(userID, id string) error {
	_, err := s.DB.Exec(`DELETE FROM subtasks WHERE id = ? AND user_id = ?`, id, userID)
	return err
}

// SoftDeleteTask archives a task (no data loss).
func (s *TaskService) SoftDeleteTask(userID, id string) error {
	_, err := s.DB.Exec(`UPDATE tasks SET archived = 1 WHERE id = ? AND user_id = ?`, id, userID)
	return err
}

// BranchChecklistItem converts a checklist item into a new child task.
func (s *TaskService) BranchChecklistItem(userID, taskID, itemID string) (*models.Task, error) {
	// Get the checklist item
	var label string
	var dimension string
	err := s.DB.QueryRow(`SELECT ci.label, t.dimension FROM checklist_items ci JOIN tasks t ON ci.task_id = t.id WHERE ci.id = ? AND ci.user_id = ?`, itemID, userID).Scan(&label, &dimension)
	if err != nil {
		return nil, fmt.Errorf("checklist item not found: %w", err)
	}

	// Create new child task
	var pid *string = &taskID
	if taskID == "" {
		pid = nil
	}
	newTask := models.Task{
		ID:        uuid.New().String(),
		UserID:    userID,
		Title:     label,
		Template:  models.TemplateTask,
		Dimension: dimension,
		CreatedAt: time.Now(),
		ParentID:  pid,
	}

	_, err = s.DB.Exec(
		`INSERT INTO tasks (id, title, template, dimension, created_at, parent_id, user_id) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		newTask.ID, newTask.Title, newTask.Template, newTask.Dimension, newTask.CreatedAt, newTask.ParentID, userID,
	)
	if err != nil {
		return nil, err
	}

	// Insert subtask relationship
	s.DB.Exec(`INSERT INTO subtasks (id, parent_id, title, done, created_at, user_id) VALUES (?, ?, ?, 0, ?, ?)`,
		uuid.New().String(), taskID, label, time.Now(), userID)

	return &newTask, nil
}

// NewFromTemplate creates a task with pre-populated modules based on template.
func NewFromTemplate(tmpl models.TaskTemplate) models.Task {
	base := models.Task{
		ID:        uuid.New().String(),
		Template:  tmpl,
		CreatedAt: time.Now(),
	}
	switch tmpl {
	case models.TemplateTask:
		base.TimeBlock = &models.TimeBlock{}
	case models.TemplateNote:
		// text only — no TimeBlock
	case models.TemplateMeeting:
		base.TimeBlock = &models.TimeBlock{}
		base.Checklist = []models.CheckItem{}
	case models.TemplateRitual:
		base.RitualCron = "0 8 * * *" // default 8am daily
	}

	// Default modules for all tasks
	base.Modules = []string{
		models.ModuleDescription,
		models.ModuleChecklist,
		models.ModuleAttachments,
		models.ModuleAIAdvice,
		models.ModuleMeetings,
	}
	return base
}

// TaskToJSON serializes a task to JSON bytes (used for snapshots).
func TaskToJSON(task models.Task) ([]byte, error) {
	return json.Marshal(task)
}

// GetContextPacket fetches a specific task context for session isolation.
func (s *TaskService) GetContextPacket(userID string, targetTaskId string, dimension string) (map[string]interface{}, error) {
	var tasks []models.Task

	if targetTaskId != "" {
		// Fetch only the specific task and its immediate subtasks/context
		rows, err := s.DB.Query(
			`SELECT id, title, template, dimension, created_at, completed_at, archived, 
			        time_block_start, time_block_end, text, ritual_cron, x, y, files, modules, parent_id, meetings
			 FROM tasks WHERE id = ? AND user_id = ?`, targetTaskId, userID,
		)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		if rows.Next() {
			var t models.Task
			var completedAt sql.NullTime
			var tbStart, tbEnd sql.NullTime
			var filesJSON sql.NullString
			var modulesJSON sql.NullString
			var meetingsJSON sql.NullString
			rows.Scan(&t.ID, &t.Title, &t.Template, &t.Dimension, &t.CreatedAt,
				&completedAt, &t.Archived, &tbStart, &tbEnd, &t.Text, &t.RitualCron, &t.X, &t.Y, &filesJSON, &modulesJSON, &t.ParentID, &meetingsJSON)

			if completedAt.Valid {
				t.CompletedAt = &completedAt.Time
			}
			if tbStart.Valid && tbEnd.Valid {
				t.TimeBlock = &models.TimeBlock{Start: tbStart.Time, End: tbEnd.Time}
			}
			if filesJSON.Valid && filesJSON.String != "" {
				json.Unmarshal([]byte(filesJSON.String), &t.Files)
			}
			if modulesJSON.Valid && modulesJSON.String != "" {
				json.Unmarshal([]byte(modulesJSON.String), &t.Modules)
			}
			if meetingsJSON.Valid && meetingsJSON.String != "" {
				json.Unmarshal([]byte(meetingsJSON.String), &t.Meetings)
			}

			// Fetch subtasks for this specific task
			subRows, _ := s.DB.Query(`SELECT id, parent_id, title, done, created_at FROM subtasks WHERE parent_id = ? AND user_id = ?`, t.ID, userID)
			if subRows != nil {
				defer subRows.Close()
				for subRows.Next() {
					var sub models.Subtask
					subRows.Scan(&sub.ID, &sub.ParentID, &sub.Title, &sub.Done, &sub.CreatedAt)
					t.Subtasks = append(t.Subtasks, sub)
				}
			}
			tasks = append(tasks, t)
		}
	} else {
		// Fetch all tasks for the dimension
		allTasks, err := s.GetAllTasks(userID)
		if err != nil {
			return nil, err
		}
		for _, t := range allTasks {
			if t.Dimension == dimension {
				tasks = append(tasks, t)
			}
		}
	}

	return map[string]interface{}{
		"session_id": uuid.New().String(),
		"version":    "2.0.0",
		"data": map[string]interface{}{
			"tasks":     tasks,
			"dimension": dimension,
		},
	}, nil
}
