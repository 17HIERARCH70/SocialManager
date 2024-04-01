package models

import "time"

type Email struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	Sender     string    `json:"sender"`
	Subject    string    `json:"subject"`
	Body       string    `json:"body"`
	ReceivedAt time.Time `json:"received_at"`
}

type EmailFilter struct {
	UserID int64  `json:"user_id"`
	Query  string `json:"query"` // Пример: "from:someone@example.com"
}
