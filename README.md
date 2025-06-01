# Financial Chat Application with Stock Quotes

A real-time chat application built with Go, WebSockets, and Kafka that allows users to communicate and get stock quotes.

## Features

- User registration and authentication
- Real-time chat with WebSocket connections
- Stock quote commands using `/stock=SYMBOL` format
- Decoupled stock bot using Kafka message broker
- Message persistence with MySQL
- Last 50 messages display
- Responsive web interface

## Architecture

- **Main Server**: Handles HTTP requests, WebSocket connections, and user authentication
- **Stock Bot**: Separate service that processes stock requests via Kafka
- **MySQL**: Stores user accounts and chat messages
- **Kafka**: Message broker for decoupled stock quote processing

## Prerequisites

- Go 1.21+
- Docker and Docker Compose
- Git

## Quick Start

1. **Start dependencies:**
```bash
docker-compose up -d
```

2. **Wait for services to be ready:**
```bash
# Check if MySQL is ready
docker-compose logs mysql

# Check if Kafka is ready
docker-compose logs kafka
```

3. **Install Go dependencies:**
```bash
go mod tidy
```

4. **Run the main server:**
```bash
go run cmd/server/main.go
```

5. **Run the stock bot (in another terminal):**
```bash
go run cmd/bot/main.go
```

6. **Open your browser:**
    - Go to `http://localhost:8080`
    - Register two different users
    - Open two browser windows/tabs
    - Test the chat and stock commands

## Usage

### Chat Commands

- Regular messages: Just type and send
- Stock quotes: `/stock=SYMBOL` (e.g., `/stock=aapl.us`, `/stock=msft.us`)

### Testing Stock Quotes

Try these stock symbols:
- `/stock=aapl.us` - Apple Inc.
- `/stock=msft.us` - Microsoft Corp.
- `/stock=googl.us` - Alphabet Inc.
- `/stock=tsla.us` - Tesla Inc.

## API Endpoints

- `GET /` - Redirect to chat
- `GET /login` - Login page
- `POST /login` - Process login
- `GET /register` - Registration page
- `POST /register` - Process registration
- `GET /chat` - Chat room (requires authentication)
- `GET /ws` - WebSocket endpoint
- `POST /logout` - Logout

## Development

### Project Structure

```
go-challenge-financial-chat/
├── cmd/
│   ├── server/main.go          # Main HTTP server
│   └── bot/main.go             # Stock bot service
├── internal/
│   ├── auth/auth.go            # Authentication service
│   ├── chat/hub.go             # WebSocket hub
│   ├── database/db.go          # Database operations
│   ├── handlers/handlers.go    # HTTP handlers
│   ├── models/models.go        # Data models
│   └── stock/stock.go          # Stock service
├── web/
│   ├── static/                 # CSS and JS files
│   └── templates/              # HTML templates
├── docker-compose.yml          # Docker services
├── init.sql                    # Database schema
└── go.mod                      # Go dependencies
```

### Database Schema

The application uses two main tables:
- `users`: User accounts with hashed passwords
- `messages`: Chat messages with timestamps

### Message Flow

1. User sends message via WebSocket
2. Server processes message
3. If stock command, sends request to Kafka
4. Stock bot processes request and fetches data
5. Stock bot sends response back via Kafka
6. Server broadcasts response to all clients
7. Regular messages are saved to database and broadcasted

### Logs

- Server logs: Check console output from `go run cmd/server/main.go`
- Stock bot logs: Check console output from `go run cmd/bot/main.go`
- Database logs: `docker-compose logs mysql`
- Kafka logs: `docker-compose logs kafka`

### Reset Database

```bash
docker-compose down -v
docker-compose up -d
```

## Testing

Run unit tests:
```bash
go test ./...
```
