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

	port := os.Getenv("LITEEVENTS_PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("Server starting on port", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func initDB(db *sql.DB) error {
	schemaSQL, err := os.ReadFile("schema.sql")
	if err != nil {
		return err
	}

	_, err = db.Exec(string(schemaSQL))
	return err
}
