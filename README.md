# Smart Mail Relay Service

A Go-based email forwarding service that automatically forwards emails based on keyword matching rules. The service supports both Gmail API (OAuth2) and IMAP for email fetching, with a robust scheduling system and comprehensive monitoring.

## Features

- **Email Fetching**: Supports both Gmail API (OAuth2) and IMAP
- **Keyword Matching**: Parses email subjects and matches against forwarding rules
- **Idempotent Processing**: Prevents duplicate email processing
- **Scheduled Processing**: Configurable interval-based email processing
- **REST API**: Full CRUD operations for forwarding rules
- **Health Monitoring**: Health checks and Prometheus metrics
- **Graceful Shutdown**: Proper signal handling and cleanup
- **Docker Support**: Complete containerization with docker-compose

## Project Structure

```
smart-mail-relay-go/
├── cmd/
│   └── api/
│       └── main.go                 # Entry point
├── config/
│   ├── config.go                   # Viper configuration
│   └── config.yaml.example         # Sample configuration
├── internal/
│   ├── database/                   # Database connection setup
│   ├── handler/                    # HTTP handlers
│   │   └── scheduler/              # Scheduler control endpoints
│   ├── metrics/                    # Prometheus metrics
│   ├── model/                      # GORM models
│   ├── repository/                 # Data access layer
│   ├── router/                     # Gin router
│   └── service/                    # Application services
│       ├── scheduler/              # Scheduler core and processing
│       └── mail_service.go         # Mail service
└── tools/
    └── get_token.go                # OAuth2 helper
```

## Architecture

The service consists of several key components:

- **Mail Service** (`internal/service/mail_service.go`): Fetches, parses, and forwards emails
- **Scheduler Service** (`internal/service/scheduler`): Manages periodic processing cycles and email processing
- **REST API** (`internal/handler`): Gin router with rule, log, and scheduler endpoints
- **Database Layer**: MySQL with GORM for persistence
- **Metrics**: Prometheus metrics for monitoring

## Database Schema

### Tables

1. **forward_rules**: Stores email forwarding rules
   - `id` (Primary Key)
   - `keyword` (Unique, indexed)
   - `target_email`
   - `enabled` (Boolean)
   - `created_at`, `updated_at`

2. **processed_emails**: Ensures idempotency
   - `id` (Primary Key)
   - `message_id` (Unique, indexed)
   - `processed_at`

3. **forward_logs**: Tracks all forwarding attempts
   - `id` (Primary Key)
   - `message_id` (Indexed)
   - `rule_id` (Foreign Key, indexed)
   - `status` (success/failure/skipped/error)
   - `error_msg`
   - `created_at`

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Gmail account with OAuth2 credentials or IMAP access

### 1. Clone and Setup

```bash
git clone <repository-url>
cd smart-mail-relay-go
```

### 2. Configure Gmail OAuth2 (Recommended)

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing one
3. Enable Gmail API
4. Create OAuth2 credentials:
   - Application type: Desktop application
   - Scopes: `https://www.googleapis.com/auth/gmail.readonly` and `https://www.googleapis.com/auth/gmail.send`
5. Download the credentials and note the Client ID and Client Secret
6. Add yourself as a test user:
   - Go to Audience 
   - Under 'Test users' click 'Add Users' 
   - Add your email and click 'Save'

### 3. Get Refresh Token

Use the provided script or manually obtain a refresh token:

```bash
# Create a simple script to get refresh token
cat > get_token.go << 'EOF'
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"

    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
    "google.golang.org/api/gmail/v1"
    "google.golang.org/api/option"
)

func main() {
    clientID := os.Getenv("GMAIL_CLIENT_ID")
    clientSecret := os.Getenv("GMAIL_CLIENT_SECRET")
    
    if clientID == "" || clientSecret == "" {
        log.Fatal("Please set GMAIL_CLIENT_ID and GMAIL_CLIENT_SECRET environment variables")
    }

    config := &oauth2.Config{
        ClientID:     clientID,
        ClientSecret: clientSecret,
        Scopes:       []string{gmail.GmailReadonlyScope, gmail.GmailSendScope},
        Endpoint:     google.Endpoint,
        RedirectURL:  "http://localhost:8080/callback",
    }

    authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
    fmt.Printf("Go to the following link in your browser: %v\n", authURL)

    var authCode string
    fmt.Print("Enter the authorization code: ")
    fmt.Scan(&authCode)

    tok, err := config.Exchange(context.Background(), authCode)
    if err != nil {
        log.Fatalf("Unable to retrieve token from web: %v", err)
    }

    fmt.Printf("Refresh Token: %s\n", tok.RefreshToken)
}
EOF

# Run the script
export GMAIL_CLIENT_ID="your-client-id"
export GMAIL_CLIENT_SECRET="your-client-secret"
go run get_token.go
```

### 4. Configure Environment Variables

Create a `.env` file:

```bash
# Gmail OAuth2 Configuration
GMAIL_CLIENT_ID=your-client-id
GMAIL_CLIENT_SECRET=your-client-secret
GMAIL_REFRESH_TOKEN=your-refresh-token
GMAIL_USER_EMAIL=your-email@gmail.com

# Alternative: IMAP Configuration
# GMAIL_USE_IMAP=true
# GMAIL_IMAP_USER=your-email@gmail.com
# GMAIL_IMAP_PASSWORD=your-app-password

# Scheduler Configuration
SCHEDULER_INTERVAL_MINUTES=5
SCHEDULER_MAX_RETRIES=3
```

### 5. Configure Application (Optional)

Copy the sample configuration file and modify it if needed:

```bash
cp config/config.yaml.example config/config.yaml
```

### 6. Start the Service

```bash
# Start all services
docker-compose up -d

# Check logs
docker-compose logs -f app
```

### 7. Verify Installation

```bash
# Health check
curl http://localhost:8080/healthz

# List forwarding rules
curl http://localhost:8080/api/v1/rules
```

## API Documentation

### Health Check

```http
GET /healthz
```

Returns service health status including database and Gmail connectivity.

### Forwarding Rules

#### List Rules
```http
GET /api/v1/rules
```

#### Create Rule
```http
POST /api/v1/rules
Content-Type: application/json

{
  "keyword": "urgent",
  "target_email": "admin@company.com",
  "enabled": true
}
```

#### Get Rule
```http
GET /api/v1/rules/{id}
```

#### Update Rule
```http
PUT /api/v1/rules/{id}
Content-Type: application/json

{
  "keyword": "urgent",
  "target_email": "admin@company.com",
  "enabled": true
}
```

#### Delete Rule
```http
DELETE /api/v1/rules/{id}
```

#### Enable/Disable Rule
```http
PATCH /api/v1/rules/{id}/enable
PATCH /api/v1/rules/{id}/disable
```

### Forward Logs

#### List Logs
```http
GET /api/v1/logs?page=1&limit=50
```

#### Get Log
```http
GET /api/v1/logs/{id}
```

### Scheduler Control

#### Start Scheduler
```http
POST /api/v1/scheduler/start
```

#### Stop Scheduler
```http
POST /api/v1/scheduler/stop
```

#### Run Once
```http
POST /api/v1/scheduler/run-once
```

#### Get Status
```http
GET /api/v1/scheduler/status
```

### Metrics

```http
GET /metrics
```

Returns Prometheus metrics including:
- `smart_mail_relay_pull_count`: Number of email fetch operations
- `smart_mail_relay_match_count`: Number of emails that matched rules
- `smart_mail_relay_forward_successes`: Successful forwards
- `smart_mail_relay_forward_failures`: Failed forwards
- `smart_mail_relay_processing_duration_seconds`: Processing time histogram
- `smart_mail_relay_active_rules`: Number of active rules
- `smart_mail_relay_total_rules`: Total number of rules

## Email Processing Logic

1. **Fetch**: Retrieve new emails from Gmail/IMAP
2. **Parse**: Extract keyword from subject (format: `<keyword> - <recipient_name>`)
3. **Match**: Find matching forwarding rule
4. **Check**: Verify email hasn't been processed before
5. **Forward**: Send email to target address
6. **Log**: Record the attempt in forward_logs
7. **Mark**: Mark email as processed

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | Database host | `localhost` |
| `DB_PORT` | Database port | `3306` |
| `DB_USER` | Database user | `smart_mail_relay` |
| `DB_PASSWORD` | Database password | `password` |
| `DB_NAME` | Database name | `smart_mail_relay` |
| `GMAIL_CLIENT_ID` | OAuth2 client ID | - |
| `GMAIL_CLIENT_SECRET` | OAuth2 client secret | - |
| `GMAIL_REFRESH_TOKEN` | OAuth2 refresh token | - |
| `GMAIL_USER_EMAIL` | Gmail user email | - |
| `GMAIL_USE_IMAP` | Use IMAP instead of API | `false` |
| `GMAIL_IMAP_HOST` | IMAP host | `imap.gmail.com` |
| `GMAIL_IMAP_PORT` | IMAP port | `993` |
| `GMAIL_IMAP_USER` | IMAP username | - |
| `GMAIL_IMAP_PASSWORD` | IMAP password | - |
| `SCHEDULER_INTERVAL_MINUTES` | Processing interval | `5` |
| `SCHEDULER_MAX_RETRIES` | Max retry attempts | `3` |
| `SERVER_PORT` | HTTP server port | `8080` |

### Configuration File

The service also supports a `config/config.yaml` file for configuration. Copy `config/config.yaml.example` to `config/config.yaml`. Environment variables take precedence over the config file.

## Monitoring

### Prometheus

The service exposes Prometheus metrics at `/metrics`. Use the provided Prometheus configuration to scrape metrics.

### Grafana

A Grafana instance is included in docker-compose for metrics visualization. Access it at `http://localhost:3000` (admin/admin).

### Health Checks

- **Application**: `http://localhost:8080/healthz`
- **Database**: MySQL health check in docker-compose
- **Container**: Docker health checks configured

## Development

### Local Development

```bash
# Install dependencies
go mod download

# Run locally (requires MySQL)
go run ./cmd/api

# Run tests
go test ./...
```

### Building

```bash
# Build binary
go build -o smart-mail-relay ./cmd/api

# Build Docker image
docker build -t smart-mail-relay .
```

## Troubleshooting

### Common Issues

1. **OAuth2 Token Expired**: Refresh tokens can expire. Generate a new one using the token script.
2. **Gmail API Quota**: Gmail API has rate limits. The service includes exponential backoff.
3. **Database Connection**: Ensure MySQL is running and accessible.
4. **IMAP Authentication**: For IMAP, use App Passwords instead of regular passwords.

### Logs

```bash
# Application logs
docker-compose logs app

# Database logs
docker-compose logs mysql

# All logs
docker-compose logs -f
```

### Debug Mode

Set log level to debug:

```bash
export LOG_LEVEL=debug
docker-compose up
```

## Security Considerations

1. **OAuth2 Credentials**: Store securely, never commit to version control
2. **Database Passwords**: Use strong passwords in production
3. **Network Access**: Restrict database access in production
4. **App Passwords**: Use Gmail App Passwords for IMAP access
5. **HTTPS**: Use HTTPS in production environments

## Production Deployment

1. Use proper secrets management
2. Configure HTTPS/TLS
3. Set up proper monitoring and alerting
4. Use production-grade MySQL
5. Configure backup strategies
6. Set appropriate resource limits
7. Use proper logging aggregation

## License

[Add your license here] 