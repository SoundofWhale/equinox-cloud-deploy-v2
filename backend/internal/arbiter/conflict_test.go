package arbiter

import (
	"testing"
	"time"

	"equinox/internal/models"
)

func TestConflictHardSlotBlocks(t *testing.T) {
	slots := []models.Slot{
		{ID: "1", Name: "Sleep", StartTime: "23:00", EndTime: "07:00", RecurRule: "daily", Type: models.SlotHard},
	}

	// Task at midnight — should be blocked by Sleep
	taskStart := time.Date(2025, 6, 15, 0, 30, 0, 0, time.Local)
	taskEnd := time.Date(2025, 6, 15, 1, 30, 0, 0, time.Local)

	result := CheckConflict(taskStart, taskEnd, slots)

	if !result.HasConflict {
		t.Fatal("expected conflict with Sleep slot")
	}
	if !result.Blocking {
		t.Fatal("expected hard slot to be blocking")
	}
	if result.SlotName != "Sleep" {
		t.Fatalf("expected slot name 'Sleep', got %q", result.SlotName)
	}
	t.Logf("✅ Hard conflict detected: %s", result.Message)
}

func TestConflictSoftSlotWarns(t *testing.T) {
	slots := []models.Slot{
		{ID: "2", Name: "Lunch", StartTime: "12:00", EndTime: "13:00", RecurRule: "daily", Type: models.SlotSoft},
	}

	taskStart := time.Date(2025, 6, 15, 12, 30, 0, 0, time.Local)
	taskEnd := time.Date(2025, 6, 15, 13, 30, 0, 0, time.Local)

	result := CheckConflict(taskStart, taskEnd, slots)

	if !result.HasConflict {
		t.Fatal("expected conflict with Lunch slot")
	}
	if result.Blocking {
		t.Fatal("soft slot should not be blocking")
	}
	t.Logf("✅ Soft conflict detected: %s", result.Message)
}

func TestNoConflict(t *testing.T) {
	slots := []models.Slot{
		{ID: "1", Name: "Sleep", StartTime: "23:00", EndTime: "07:00", RecurRule: "daily", Type: models.SlotHard},
		{ID: "2", Name: "Family", StartTime: "18:00", EndTime: "20:00", RecurRule: "daily", Type: models.SlotHard},
	}

	// Task at 10am — no conflict
	taskStart := time.Date(2025, 6, 15, 10, 0, 0, 0, time.Local)
	taskEnd := time.Date(2025, 6, 15, 11, 0, 0, 0, time.Local)

	result := CheckConflict(taskStart, taskEnd, slots)

	if result.HasConflict {
		t.Fatalf("expected no conflict, got: %s", result.Message)
	}
	t.Log("✅ No conflict — task allowed")
}

func TestOvernightSlot(t *testing.T) {
	slots := []models.Slot{
		{ID: "1", Name: "Sleep", StartTime: "23:00", EndTime: "07:00", RecurRule: "daily", Type: models.SlotHard},
	}

	// Task at 6am — should conflict (inside 23:00–07:00 overnight range)
	taskStart := time.Date(2025, 6, 15, 6, 0, 0, 0, time.Local)
	taskEnd := time.Date(2025, 6, 15, 6, 30, 0, 0, time.Local)

	result := CheckConflict(taskStart, taskEnd, slots)

	if !result.HasConflict {
		t.Fatal("expected conflict with overnight Sleep slot at 6am")
	}
	t.Logf("✅ Overnight slot conflict detected: %s", result.Message)
}

func TestFamilyTimeConflict(t *testing.T) {
	slots := models.DefaultHardSlots()

	// Task at 19:00 — should conflict with Family Time (18:00–20:00)
	taskStart := time.Date(2025, 6, 15, 19, 0, 0, 0, time.Local)
	taskEnd := time.Date(2025, 6, 15, 19, 30, 0, 0, time.Local)

	result := CheckConflict(taskStart, taskEnd, slots)

	if !result.HasConflict {
		t.Fatal("expected conflict with Family Time")
	}
	if result.SlotName != "Family Time" {
		t.Fatalf("expected 'Family Time', got %q", result.SlotName)
	}
	t.Logf("✅ Family Time conflict: %s", result.Message)
}

func TestWeeklyRecurrence(t *testing.T) {
	slots := []models.Slot{
		{
			ID:        "3",
			Name:      "Rest",
			StartTime: "13:00",
			EndTime:   "18:00",
			RecurRule: "weekly:Sat",
			Type:      models.SlotHard,
		},
	}

	// June 14, 2025 is a Saturday
	taskStartSat := time.Date(2025, 6, 14, 14, 0, 0, 0, time.Local)
	taskEndSat := time.Date(2025, 6, 14, 15, 0, 0, 0, time.Local)

	resultSat := CheckConflict(taskStartSat, taskEndSat, slots)
	if !resultSat.HasConflict {
		t.Error("expected conflict on Saturday for 'weekly:Sat' slot")
	}

	// June 16, 2025 is a Monday
	taskStartMon := time.Date(2025, 6, 16, 14, 0, 0, 0, time.Local)
	taskEndMon := time.Date(2025, 6, 16, 15, 0, 0, 0, time.Local)

	resultMon := CheckConflict(taskStartMon, taskEndMon, slots)
	if resultMon.HasConflict {
		t.Error("expected NO conflict on Monday for 'weekly:Sat' slot")
	}
}
