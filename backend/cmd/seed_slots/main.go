package main

import (
	"fmt"
	"log"
	"os"

	"equinox/internal/models"
	"equinox/internal/security"

	"github.com/google/uuid"
)

func main() {
	dbPath := "vault.db"
	if envPath := os.Getenv("EQUINOX_DB_PATH"); envPath != "" {
		dbPath = envPath
	}

	// In a real flow, the key comes from the KeyManager.
	// For seeding, we use a default dev key.
	hexKey := os.Getenv("EQUINOX_HEX_KEY")
	if hexKey == "" {
		hexKey = "0000000000000000000000000000000000000000000000000000000000000000"
	}

	db, err := security.OpenDB(dbPath, hexKey)
	if err != nil {
		log.Fatalf("❌ Cannot open DB: %v", err)
	}
	defer db.Close()

	// Check if any slots already exist.
	var count int
	err = db.Conn.QueryRow("SELECT COUNT(*) FROM slots").Scan(&count)
	if err != nil {
		log.Fatalf("❌ Cannot count slots: %v", err)
	}
	if count > 0 {
		fmt.Printf("⏭️  Slots already seeded (%d found). Skipping.\n", count)
		return
	}

	// Insert default hard slots.
	defaults := models.DefaultHardSlots()
	for _, slot := range defaults {
		id := uuid.New().String()
		_, err := db.Conn.Exec(
			`INSERT INTO slots (id, name, start_time, end_time, recur_rule, slot_type) VALUES (?, ?, ?, ?, ?, ?)`,
			id, slot.Name, slot.StartTime, slot.EndTime, slot.RecurRule, string(slot.Type),
		)
		if err != nil {
			log.Fatalf("❌ Failed to insert slot %q: %v", slot.Name, err)
		}
		fmt.Printf("✅ Seeded: %s (%s–%s, %s)\n", slot.Name, slot.StartTime, slot.EndTime, slot.RecurRule)
	}

	fmt.Println("🛡️  All default hard slots seeded successfully.")
}
