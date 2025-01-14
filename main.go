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

	// Initialize schema
	if err := initDB(db); err != nil {
		log.Fatal(err)
	}

	hub := NewHub()
	go hub.Run()

	http.Handle("/", handleIndex(db))
	http.Handle("/ws", handleWS(hub))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/api/events", handleEvents(db, hub))

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
