package stock

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go-challenge-financial-chat/internal/models"
)

type Service struct {
	kafkaReader *kafka.Reader
	kafkaWriter *kafka.Writer
}

func NewService(brokers string) *Service {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{brokers},
		Topic:   "stock-requests",
		GroupID: "stock-bot",
	})

	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers),
		Topic:    "stock-quotes",
		Balancer: &kafka.LeastBytes{},
	}

	return &Service{
		kafkaReader: reader,
		kafkaWriter: writer,
	}
}

func (s *Service) Start() {
	log.Println("Stock bot started, listening for requests...")

	for {
		msg, err := s.kafkaReader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error reading from Kafka: %v", err)
			time.Sleep(time.Second)
			continue
		}

		var request map[string]string
		if err := json.Unmarshal(msg.Value, &request); err != nil {
			log.Printf("Error unmarshaling request: %v", err)
			continue
		}

		stockCode := request["stock_code"]
		user := request["user"]

		log.Printf("Processing stock request for %s from user %s", stockCode, user)

		quote, err := s.fetchStockQuote(stockCode)
		if err != nil {
			log.Printf("Error fetching stock quote for %s: %v", stockCode, err)
			continue
		}

		quoteBytes, _ := json.Marshal(quote)
		err = s.kafkaWriter.WriteMessages(context.Background(),
			kafka.Message{
				Key:   []byte(stockCode),
				Value: quoteBytes,
			},
		)

		if err != nil {
			log.Printf("Error sending quote to Kafka: %v", err)
		}
	}
}

func (s *Service) fetchStockQuote(stockCode string) (*models.StockQuote, error) {
	url := fmt.Sprintf("https://stooq.com/q/l/?s=%s&f=sd2t2ohlcv&h&e=csv", stockCode)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	reader := csv.NewReader(resp.Body)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("insufficient data received")
	}

	data := records[1]
	if len(data) < 7 {
		return nil, fmt.Errorf("invalid CSV format")
	}

	closePrice, err := strconv.ParseFloat(data[6], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid close price: %v", err)
	}

	return &models.StockQuote{
		Symbol: strings.ToUpper(data[0]),
		Price:  closePrice,
		Date:   data[1],
		Time:   data[2],
	}, nil
}

func (s *Service) Close() {
	log.Println("Closing stock bot service...")
	if err := s.kafkaReader.Close(); err != nil {
		log.Printf("Error stopping Kafka reader: %v", err)
	}

	if err := s.kafkaWriter.Close(); err != nil {
		log.Printf("Error stopping Kafka writer: %v", err)
	}

	log.Println("Stock bot service closed.")
}
