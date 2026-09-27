package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
)

var db *sql.DB

func initDB() {
	godotenv.Load()
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./chats.db"
	}

	var err error
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Opened database successfully")

	query := `
		CREATE TABLE IF NOT EXISTS chats (
			chat_id INTEGER PRIMARY KEY
		);`

	_, err = db.Exec(query)
	if err != nil {
		log.Fatal("error creating table", err)
	}
	log.Println("Created database successfully")
}
