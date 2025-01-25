# Wallet API

A RESTful API for a digital wallet system built with Go, featuring user management, transactions, and background processing.

## Features

- User registration and authentication (JWT)
- Top-up balance
- Make payments
- Transfer money between users (async processing)
- Transaction history
- Profile management
- Background task monitoring dashboard

## Tech Stack

- Go 1.21
- Gin Web Framework
- GORM (PostgreSQL)
- JWT Authentication
- Redis + Asynq (Background Processing)
- Asynqmon (Task Monitoring Dashboard)

## Project Structure

```
.
├── cmd
│   └── api
│       └── main.go
├── config
│   └── config.yaml
├── internal
│   ├── delivery
│   │   └── http
│   ├── domain
│   ├── middleware
│   ├── repository
│   └── usecase
├── pkg
│   ├── auth
│   ├── database
│   └── queue
└── README.md
```

## Getting Started

### Running the Application

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/wallet-api.git
   cd wallet-api
   ```

2. Configure the application:
   - Copy `config/config.yaml.example` to `config/config.yaml`
   - Update the configuration values if needed

3. Start the application:
    ```bash
    go run cmd/api/main.go
    ```

The API will be available at `http://localhost:8080`
The monitoring dashboard will be available at `http://localhost:8081/monitoring`

## Monitoring Dashboard

The application includes a web-based monitoring dashboard for background tasks. You can access it at `http://localhost:8081/monitoring`. The dashboard provides:

- Real-time monitoring of queue tasks
- View task details including payload and status
- View worker processes and their status
- View retry attempts and failures
- Queue statistics and metrics
- Ability to retry failed tasks
- Queue management capabilities

## API Endpoints

### Public Endpoints

- `POST /register` - Register a new user
- `POST /login` - Login and get JWT tokens

### Protected Endpoints (Requires JWT)

- `POST /topup` - Add balance to wallet
- `POST /pay` - Make a payment
- `POST /transfer` - Transfer money to another user
- `GET /transactions` - Get transaction history
- `PUT /profile` - Update user profile

## Example Requests

### Register

```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "John",
    "last_name": "Doe",
    "phone_number": "1234567890",
    "address": "123 Main St",
    "pin": "123456"
  }'
```

### Login

```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "1234567890",
    "pin": "123456"
  }'
```

### Top Up

```bash
curl -X POST http://localhost:8080/topup \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 100000
  }'
```

## Development

### Running Tests

```bash
go test ./...
```

### Database Migrations

The application uses GORM auto-migration to manage the database schema. Migrations are automatically applied when the application starts.

## License

MIT License 