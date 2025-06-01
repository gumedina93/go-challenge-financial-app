package models

import (
	"time"
)

type User struct {
	ID           int       `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	PasswordHash string    `json:"-" db:"password_hash"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type Message struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	Username  string    `json:"username" db:"username"`
	Content   string    `json:"content" db:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type StockQuote struct {
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price"`
	Date   string  `json:"date"`
	Time   string  `json:"time"`
}

type WSMessage struct {
	Type     string    `json:"type"`
	Username string    `json:"username"`
	Content  string    `json:"content"`
	Time     time.Time `json:"time"`
}
