package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
)

var db *sql.DB

func initDB() {
	godotenv.Load()
	dbPath := os.Getenv("DB_PATH")
	dbName := os.Getenv("DB_FILENAME")
	if dbPath == "" {
		dbPath = fmt.Sprintf("./%s", dbName)
	}

	dbDir := filepath.Dir(dbPath)
	if dbDir != "." && dbDir != "" {
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			log.Printf("[ERROR] failed to create directory \"%s\" where database is stored: %v", dbDir, err)
		}
	}

	var err error
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("[FATAL] failed to open database: %v", err)
	}
	log.Println("[SUCCESS] opened database successfully")

	schema := `
		CREATE TABLE IF NOT EXISTS chat_ids (
			chat_id INTEGER PRIMARY KEY NOT NULL
		);

		CREATE TABLE IF NOT EXISTS game_stats (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			app_id INTEGER NOT NULL,
			online_count INTEGER NOT NULL,
			checked_at DATETIME DEFAULT (datetime('now')) NOT NULL
		);

		CREATE TABLE IF NOT EXISTS user_settings (
			user_id INTEGER PRIMARY KEY,
			timezone TEXT DEFAULT +0
		);
		`

	_, err = db.Exec(schema)
	if err != nil {
		log.Fatalf("[FATAL] failed to initialize database schema: %v", err)
	}
	log.Println("[SUCCESS] Database schema initialized successfully")
	log.Println("[SUCCESS] created database successfully")
}
