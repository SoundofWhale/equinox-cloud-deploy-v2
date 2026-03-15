package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mutecomm/go-sqlcipher/v4"
)

func main() {
	dbPath := "/vault/vault.db"
	key := os.Getenv("EQUINOX_HEX_KEY")
	if key == "" {
		key = "0000000000000000000000000000000000000000000000000000000000000000"
	}

	dsn := fmt.Sprintf("file:%s?_pragma_key=x'%s'&_pragma_cipher_page_size=4096&_pragma_kdf_iter=256000",
		dbPath, key)

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	rows, err := db.Query("SELECT id, username, password_hash FROM users")
	if err != nil {
		log.Fatalf("failed to query users: %v", err)
	}
	defer rows.Close()

	fmt.Println("--- REGISTERED USERS ---")
	for rows.Next() {
		var id, username, hash string
		if err := rows.Scan(&id, &username, &hash); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("ID: %s | Login: %s | Hash: %s\n", id, username, hash)
	}

	fmt.Println("\n--- RECENT TASKS ---")
	taskRows, err := db.Query("SELECT id, title, user_id FROM tasks LIMIT 10")
	if err != nil {
		log.Printf("failed to query tasks: %v", err)
	} else {
		defer taskRows.Close()
		for taskRows.Next() {
			var tid, title, uid string
			if err := taskRows.Scan(&tid, &title, &uid); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("ID: %s | Title: %s | UserID: %s\n", tid, title, uid)
		}
	}
	fmt.Println("------------------------")
}
