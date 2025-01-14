package main

import "time"

type Event struct {
    ID        int64     `json:"id"`
    Namespace string    `json:"namespace"`
    Type      string    `json:"type"`
    Data      string    `json:"data"`
    CreatedAt time.Time `json:"created_at"`
}
