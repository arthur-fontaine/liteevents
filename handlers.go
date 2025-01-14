package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"liteevents/views"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleIndex(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		component := views.Dashboard()
		component.Render(r.Context(), w)
	}
}

func handleWS(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
		client.hub.register <- client

		go client.writePump()
		go client.readPump()
	}
}

// New function to broadcast events
func broadcastEvent(hub *Hub, event Event) error {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return err
	}

	hub.broadcast <- eventJSON
	return nil
}

func handleEvents(db *sql.DB, hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			page, _ := strconv.Atoi(r.URL.Query().Get("page"))
			namespace := r.URL.Query().Get("namespace")
			if page < 1 {
				page = 1
			}

			limit := 20
			offset := (page - 1) * limit

			events, hasMore := fetchEvents(db, namespace, limit, offset)

			json.NewEncoder(w).Encode(map[string]interface{}{
				"events":   events,
				"has_more": hasMore,
			})

			return
		}

		if r.Method == http.MethodPost {
			var event Event
			if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			event.CreatedAt = time.Now()
			if err := insertEvent(db, event); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if err := broadcastEvent(hub, event); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusCreated)
			return
		}
	}
}

func fetchEvents(db *sql.DB, namespace string, limit, offset int) ([]Event, bool) {
	query := `
		SELECT id, namespace, type, data, created_at 
		FROM events 
		WHERE ($1 = '' OR namespace = $1)
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3`

	rows, err := db.Query(query, namespace, limit+1, offset)
	if err != nil {
		return nil, false
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.ID, &e.Namespace, &e.Type, &e.Data, &e.CreatedAt); err != nil {
			continue
		}
		events = append(events, e)
	}

	hasMore := len(events) > limit
	if hasMore {
		events = events[:limit]
	}

	return events, hasMore
}

// New function to insert events
func insertEvent(db *sql.DB, event Event) error {
	query := `
		INSERT INTO events (namespace, type, data, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	return db.QueryRow(query, event.Namespace, event.Type, event.Data, event.CreatedAt).Scan(&event.ID)
}
