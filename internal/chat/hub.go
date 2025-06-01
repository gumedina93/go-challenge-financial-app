package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"go-challenge-financial-chat/internal/database"
	"go-challenge-financial-chat/internal/models"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan models.WSMessage
	username string
	userID   int
}

type Hub struct {
	clients     map[*Client]bool
	broadcast   chan models.WSMessage
	register    chan *Client
	unregister  chan *Client
	db          database.Database
	kafkaWriter *kafka.Writer
}

func NewHub(db database.Database, brokers string) *Hub {
	kafkaWriter := &kafka.Writer{
		Addr:     kafka.TCP(brokers),
		Topic:    "stock-requests",
		Balancer: &kafka.LeastBytes{},
	}

	return &Hub{
		clients:     make(map[*Client]bool),
		broadcast:   make(chan models.WSMessage),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		db:          db,
		kafkaWriter: kafkaWriter,
	}
}

func (h *Hub) Run(brokers string) {
	go h.listenForStockQuotes(brokers)

	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Printf("Client %s connected", client.username)

			messages, err := h.db.GetRecentMessages(50)
			if err != nil {
				log.Printf("Error getting recent messages: %v", err)
			} else {
				for _, msg := range messages {
					wsMsg := models.WSMessage{
						Type:     "message",
						Username: msg.Username,
						Content:  msg.Content,
						Time:     msg.CreatedAt,
					}
					select {
					case client.send <- wsMsg:
					default:
						close(client.send)
						delete(h.clients, client)
					}
				}
			}

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("Client %s disconnected", client.username)
			}

		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

func (h *Hub) listenForStockQuotes(brokers string) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{brokers},
		Topic:   "stock-quotes",
		GroupID: "chat-app",
	})
	defer reader.Close()

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error reading from Kafka: %v", err)
			time.Sleep(time.Second)
			continue
		}

		var stockQuote models.StockQuote
		if err := json.Unmarshal(msg.Value, &stockQuote); err != nil {
			log.Printf("Error unmarshaling stock quote: %v", err)
			continue
		}

		botMessage := models.WSMessage{
			Type:     "message",
			Username: "StockBot",
			Content:  fmt.Sprintf("%s quote is $%.2f per share", stockQuote.Symbol, stockQuote.Price),
			Time:     time.Now(),
		}

		if err := h.db.SaveMessage(1, "StockBot", botMessage.Content); err != nil {
			log.Printf("Error saving bot message: %v", err)
		}

		h.broadcast <- botMessage
	}
}

func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request, username string, userID int) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		hub:      h,
		conn:     conn,
		send:     make(chan models.WSMessage, 256),
		username: username,
		userID:   userID,
	}

	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}

/*
readPump reads incoming messages from the WebSocket connection and checks if should send message to Kafka
or broadcasting it via the hub
*/
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var wsMsg models.WSMessage
		err := c.conn.ReadJSON(&wsMsg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		wsMsg.Username = c.username
		wsMsg.Time = time.Now()

		if strings.HasPrefix(wsMsg.Content, "/stock=") {
			stockCode := strings.TrimPrefix(wsMsg.Content, "/stock=")
			stockRequest := map[string]string{
				"stock_code": stockCode,
				"user":       c.username,
			}

			reqBytes, _ := json.Marshal(stockRequest)
			err := c.hub.kafkaWriter.WriteMessages(context.Background(),
				kafka.Message{
					Key:   []byte(stockCode),
					Value: reqBytes,
				},
			)

			if err != nil {
				log.Printf("Error sending to Kafka: %v", err)
			}

			continue
		}

		if err := c.hub.db.SaveMessage(c.userID, c.username, wsMsg.Content); err != nil {
			log.Printf("Error saving message: %v", err)
		}

		c.hub.broadcast <- wsMsg
	}
}

/*
writePump takes messages from the client.send channel and write them out to the WebSocket connection
*/
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteJSON(message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
