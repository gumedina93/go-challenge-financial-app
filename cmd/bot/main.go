package main

import (
	"github.com/joho/godotenv"
	"go-challenge-financial-chat/internal/stock"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		panic("Error loading env file")
	}

	stockService := stock.NewService(os.Getenv("KAFKA_BROKERS"))
	defer stockService.Close()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Shutting down stock bot...")
		stockService.Close()
		os.Exit(0)
	}()

	stockService.Start()
}
