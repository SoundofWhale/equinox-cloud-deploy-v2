package arbiter

import (
	"fmt"
	"time"

	"equinox/internal/models"
)

// ConflictResult describes whether a new task conflicts with existing slots.
type ConflictResult struct {
	HasConflict bool   `json:"has_conflict"`
	Blocking    bool   `json:"blocking"`     // true = hard slot, cannot proceed
	SlotName    string `json:"slot_name"`
	Message     string `json:"message"`
}

// CheckConflict tests whether the task's TimeBlock overlaps with any slot.
// If the slot is Hard → task creation is blocked.
// If the slot is Soft → warning is returned, user can override.
func CheckConflict(taskStart, taskEnd time.Time, slots []models.Slot) ConflictResult {
	for _, slot := range slots {
		if checkSlotOverlap(taskStart, taskEnd, slot) {
			if slot.Type == models.SlotHard {
				return ConflictResult{
					HasConflict: true,
					Blocking:    true,
					SlotName:    slot.Name,
					Message:     fmt.Sprintf("⛔ Blocked by protected slot: %s (%s–%s)", slot.Name, slot.StartTime, slot.EndTime),
				}
			}
			return ConflictResult{
				HasConflict: true,
				Blocking:    false,
				SlotName:    slot.Name,
				Message:     fmt.Sprintf("⚠️ Overlap with: %s (%s–%s). Continue anyway?", slot.Name, slot.StartTime, slot.EndTime),
			}
		}
	}
	return ConflictResult{HasConflict: false}
}

// checkSlotOverlap handles both normal and overnight slots with recurrence rules.
func checkSlotOverlap(taskStart, taskEnd time.Time, slot models.Slot) bool {
	loc := taskStart.Location()
	year, month, day := taskStart.Date()

	startH, startM := parseHHMM(slot.StartTime)
	endH, endM := parseHHMM(slot.EndTime)

	isOvernight := endH < startH || (endH == startH && endM <= startM)

	if isOvernight {
		// Window 1: starts YESTERDAY, ends TODAY
		if matchesRecurrence(taskStart.AddDate(0, 0, -1), slot.RecurRule) {
			s1 := time.Date(year, month, day-1, startH, startM, 0, 0, loc)
			e1 := time.Date(year, month, day, endH, endM, 0, 0, loc)
			if overlaps(taskStart, taskEnd, s1, e1) {
				return true
			}
		}
		// Window 2: starts TODAY, ends TOMORROW
		if matchesRecurrence(taskStart, slot.RecurRule) {
			s2 := time.Date(year, month, day, startH, startM, 0, 0, loc)
			e2 := time.Date(year, month, day+1, endH, endM, 0, 0, loc)
			if overlaps(taskStart, taskEnd, s2, e2) {
				return true
			}
		}
		return false
	}

	// Normal same-day slot
	if !matchesRecurrence(taskStart, slot.RecurRule) {
		return false
	}
	s := time.Date(year, month, day, startH, startM, 0, 0, loc)
	e := time.Date(year, month, day, endH, endM, 0, 0, loc)
	return overlaps(taskStart, taskEnd, s, e)
}

// matchesRecurrence checks if a date satisfies the recurrence rule.
// Supported: "daily", "weekly:Mon", "weekly:Tue", etc.
func matchesRecurrence(t time.Time, rule string) bool {
	if rule == "daily" || rule == "" {
		return true
	}
	if len(rule) > 7 && rule[:7] == "weekly:" {
		dayStr := rule[7:]
		switch dayStr {
		case "Sun": return t.Weekday() == time.Sunday
		case "Mon": return t.Weekday() == time.Monday
		case "Tue": return t.Weekday() == time.Tuesday
		case "Wed": return t.Weekday() == time.Wednesday
		case "Thu": return t.Weekday() == time.Thursday
		case "Fri": return t.Weekday() == time.Friday
		case "Sat": return t.Weekday() == time.Saturday
		}
	}
	return false
}

// overlaps returns true if two time ranges overlap.
func overlaps(s1, e1, s2, e2 time.Time) bool {
	return s1.Before(e2) && e1.After(s2)
}

// parseHHMM extracts hour and minute from "HH:MM" string.
func parseHHMM(s string) (int, int) {
	var h, m int
	fmt.Sscanf(s, "%d:%d", &h, &m)
	return h, m
}
