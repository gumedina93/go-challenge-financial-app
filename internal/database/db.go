package database

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"go-challenge-financial-chat/internal/models"
)

type Database interface {
	CreateUser(username, passwordHash string) error
	GetUser(username string) (*models.User, error)
	SaveMessage(userID int, username, content string) error
	GetRecentMessages(limit int) ([]models.Message, error)
	Close() error
}

type DB struct {
	conn *sql.DB
}

func New(connString string) (*DB, error) {
	db, err := sql.Open("mysql", connString)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &DB{conn: db}, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) CreateUser(username, passwordHash string) error {
	query := "INSERT INTO users (username, password_hash) VALUES (?, ?)"
	_, err := db.conn.Exec(query, username, passwordHash)
	return err
}

func (db *DB) GetUser(username string) (*models.User, error) {
	query := "SELECT id, username, password_hash, created_at FROM users WHERE username = ?"
	row := db.conn.QueryRow(query, username)

	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (db *DB) SaveMessage(userID int, username, content string) error {
	query := "INSERT INTO messages (user_id, username, content) VALUES (?, ?, ?)"
	_, err := db.conn.Exec(query, userID, username, content)
	return err
}

func (db *DB) GetRecentMessages(limit int) ([]models.Message, error) {
	query := `SELECT id, user_id, username, content, created_at 
              FROM messages 
              ORDER BY created_at ASC 
              LIMIT ?`

	rows, err := db.conn.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		err := rows.Scan(&msg.ID, &msg.UserID, &msg.Username, &msg.Content, &msg.CreatedAt)
		if err != nil {
			log.Printf("Error scanning message: %v", err)
			continue
		}
		messages = append(messages, msg)
	}

	return messages, nil
}
