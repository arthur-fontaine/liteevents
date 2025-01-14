package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "./events.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	auth := Auth{
		Passphrase: os.Getenv("LITEEVENTS_PASSPHRASE"),
	}

	// Initialize schema
	if err := initDB(db); err != nil {
		log.Fatal(err)
	}

	hub := NewHub()
	go hub.Run()

	http.Handle("/", handleIndex(&auth))
	http.Handle("/ws", handleWS(&auth, hub))
	http.Handle("/api/events", handleEvents(&auth, db, hub))
	http.Handle("/login", handleAuth())

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func initDB(db *sql.DB) error {
	schemaSQL, err := os.ReadFile("schema.sql")
	if err != nil {
		return err
	}

	_, err = db.Exec(string(schemaSQL))
	return err
}
