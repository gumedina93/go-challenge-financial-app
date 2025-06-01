package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"go-challenge-financial-chat/internal/auth"
	"go-challenge-financial-chat/internal/chat"
	"go-challenge-financial-chat/internal/database"
	"go-challenge-financial-chat/internal/handlers"
	"log"
	"net/http"
	"os"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		panic("Error loading env file")
	}

	db, err := database.New(getConnectionString())
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	authService := auth.NewService(db)
	brokers := os.Getenv("KAFKA_BROKERS")
	hub := chat.NewHub(db, brokers)

	go hub.Run(brokers)

	h := handlers.New(authService, hub, db)
	router := h.SetupRoutes()

	port := os.Getenv("SERVER_PORT")
	log.Printf("Server starting on %s", port)
	log.Fatal(http.ListenAndServe(port, router))
}

func getConnectionString() string {
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPass, dbHost, dbPort, dbName)
}
