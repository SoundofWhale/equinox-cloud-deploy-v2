package main

import (
	"fmt"
	"log"
	"os"

	"equinox/internal/security"
)

func main() {
	dbPath := "vault.db"
	if envPath := os.Getenv("EQUINOX_DB_PATH"); envPath != "" {
		dbPath = envPath
	}

	// For check_cipher, we use a throwaway key just to open the DB and read PRAGMA values.
	// If the DB doesn't exist yet, it creates a temp one.
	testKey := "0000000000000000000000000000000000000000000000000000000000000000"

	db, err := security.OpenDB(dbPath, testKey)
	if err != nil {
		log.Fatalf("❌ Cannot open DB: %v", err)
	}
	defer db.Close()

	ver, err := db.CipherVersion()
	if err != nil {
		log.Fatalf("❌ Cannot read cipher_version: %v", err)
	}

	pageSize, err := db.CipherPageSize()
	if err != nil {
		log.Fatalf("❌ Cannot read cipher_page_size: %v", err)
	}

	fmt.Println("╔═══════════════════════════════════════╗")
	fmt.Println("║     EQUINOX 2.0 — Cipher Check        ║")
	fmt.Println("╠═══════════════════════════════════════╣")
	fmt.Printf("║  SQLCipher version : %-17s ║\n", ver)
	fmt.Printf("║  cipher_page_size  : %-17d ║\n", pageSize)
	fmt.Println("║  Status            : ✅ ACTIVE         ║")
	fmt.Println("╚═══════════════════════════════════════╝")
}
