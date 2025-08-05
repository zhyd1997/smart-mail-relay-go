# Smart Mail Relay Service - Setup Guide

This guide will help you set up and run the Smart Mail Relay Service.

## Prerequisites

- Go 1.21 or later
- Docker and Docker Compose
- Gmail account with OAuth2 credentials or IMAP access

## Quick Start with Docker

### 1. Clone and Setup

```bash
git clone <repository-url>
cd smart-mail-relay-go
```

### 2. Configure Gmail OAuth2 (Recommended Method)

#### Step 1: Create Google Cloud Project
1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing one
3. Enable Gmail API:
   - Go to "APIs & Services" > "Library"
   - Search for "Gmail API" and enable it

#### Step 2: Create OAuth2 Credentials
1. Go to "APIs & Services" > "Credentials"
2. Click "Create Credentials" > "OAuth 2.0 Client IDs"
3. Application type: Desktop application
4. Name: "Smart Mail Relay"
5. Download the JSON file and note the Client ID and Client Secret

#### Step 3: Get Refresh Token
```bash
# Set your credentials
export GMAIL_CLIENT_ID="your-client-id"
export GMAIL_CLIENT_SECRET="your-client-secret"

# Run the token helper
go run tools/get_token.go
```

Follow the instructions to get your refresh token.

### 3. Configure Environment Variables

Create a `.env` file in the project root:

```bash
# Gmail OAuth2 Configuration
GMAIL_CLIENT_ID=your-client-id
GMAIL_CLIENT_SECRET=your-client-secret
GMAIL_REFRESH_TOKEN=your-refresh-token
GMAIL_USER_EMAIL=your-email@gmail.com

# Alternative: IMAP Configuration (if not using OAuth2)
# GMAIL_USE_IMAP=true
# GMAIL_IMAP_USER=your-email@gmail.com
# GMAIL_IMAP_PASSWORD=your-app-password

# Scheduler Configuration
SCHEDULER_INTERVAL_MINUTES=5
SCHEDULER_MAX_RETRIES=3
```

### 4. Configure Application (Optional)

Copy the sample configuration file and modify it if needed:

```bash
cp config/config.yaml.example config/config.yaml
```

### 5. Start the Service

```bash
# Start all services (MySQL, App, Prometheus, Grafana)
docker-compose up -d

# Check logs
docker-compose logs -f app
```

### 6. Verify Installation

```bash
# Health check
curl http://localhost:8080/healthz

# List forwarding rules (should be empty initially)
curl http://localhost:8080/api/v1/rules

# Start the scheduler
curl -X POST http://localhost:8080/api/v1/scheduler/start

# Check scheduler status
curl http://localhost:8080/api/v1/scheduler/status
```

## Manual Setup (Without Docker)

### 1. Install Dependencies

```bash
go mod download
```

### 2. Setup MySQL Database

```bash
# Install MySQL (Ubuntu/Debian)
sudo apt-get install mysql-server

# Or use Docker for MySQL only
docker run --name mysql -e MYSQL_ROOT_PASSWORD=rootpassword \
  -e MYSQL_DATABASE=smart_mail_relay \
  -e MYSQL_USER=smart_mail_relay \
  -e MYSQL_PASSWORD=password \
  -p 3306:3306 -d mysql:8.0
```

### 3. Configure Database Connection

Update the database configuration in `config/config.yaml` or set environment variables:

```bash
export DB_HOST=localhost
export DB_PORT=3306
export DB_USER=smart_mail_relay
export DB_PASSWORD=password
export DB_NAME=smart_mail_relay
```

### 4. Run the Application

```bash
go run ./cmd/api
```

## API Usage Examples

### Create a Forwarding Rule

```bash
curl -X POST http://localhost:8080/api/v1/rules \
  -H 'Content-Type: application/json' \
  -d '{
    "keyword": "urgent",
    "target_email": "admin@company.com",
    "enabled": true
  }'
```

### List All Rules

```bash
curl http://localhost:8080/api/v1/rules
```

### Enable/Disable a Rule

```bash
# Enable rule with ID 1
curl -X PATCH http://localhost:8080/api/v1/rules/1/enable

# Disable rule with ID 1
curl -X PATCH http://localhost:8080/api/v1/rules/1/disable
```

### View Forward Logs

```bash
curl http://localhost:8080/api/v1/logs
```

### Control Scheduler

```bash
# Start scheduler
curl -X POST http://localhost:8080/api/v1/scheduler/start

# Stop scheduler
curl -X POST http://localhost:8080/api/v1/scheduler/stop

# Run once
curl -X POST http://localhost:8080/api/v1/scheduler/run-once

# Check status
curl http://localhost:8080/api/v1/scheduler/status
```

## Monitoring

### Prometheus Metrics

```bash
curl http://localhost:8080/metrics
```

### Grafana Dashboard

Access Grafana at `http://localhost:3000`:
- Username: `admin`
- Password: `admin`

### Health Check

```bash
curl http://localhost:8080/healthz
```

## Email Processing

The service processes emails with subjects in the format:
`<keyword> - <recipient_name>`

For example:
- `urgent - John Doe` → matches rule with keyword "urgent"
- `support - Customer Service` → matches rule with keyword "support"

## Troubleshooting

### Common Issues

1. **OAuth2 Token Expired**
   ```bash
   # Generate new refresh token
   go run tools/get_token.go
   ```

2. **Database Connection Failed**
   ```bash
   # Check MySQL is running
   docker-compose ps mysql
   
   # Check logs
   docker-compose logs mysql
   ```

3. **Gmail API Quota Exceeded**
   - The service includes exponential backoff
   - Check Gmail API quotas in Google Cloud Console

4. **IMAP Authentication Failed**
   - Use App Passwords instead of regular passwords
   - Enable 2-factor authentication on Gmail

### Debug Mode

```bash
# Set debug logging
export LOG_LEVEL=debug
docker-compose up
```

### View Logs

```bash
# Application logs
docker-compose logs app

# All services
docker-compose logs -f
```

## Development

### Run Tests

```bash
go test -v
```

### Build Binary

```bash
go build -o smart-mail-relay .
```

### Code Formatting

```bash
go fmt ./...
```

### Linting

```bash
golangci-lint run
```

## Production Deployment

1. **Use Environment Variables for Secrets**
   ```bash
   export GMAIL_CLIENT_ID="your-client-id"
   export GMAIL_CLIENT_SECRET="your-client-secret"
   export GMAIL_REFRESH_TOKEN="your-refresh-token"
   ```

2. **Configure HTTPS**
   - Use a reverse proxy (nginx, traefik)
   - Set up SSL certificates

3. **Database Security**
   - Use strong passwords
   - Restrict network access
   - Enable SSL connections

4. **Monitoring**
   - Set up Prometheus alerting
   - Configure Grafana dashboards
   - Use log aggregation (ELK stack)

5. **Backup Strategy**
   - Regular database backups
   - Configuration backups
   - Log rotation

## Security Considerations

- Never commit OAuth2 credentials to version control
- Use App Passwords for IMAP access
- Restrict database access in production
- Use HTTPS in production environments
- Regularly rotate refresh tokens
- Monitor for unusual activity

## Support

For issues and questions:
1. Check the logs: `docker-compose logs app`
2. Verify configuration: `curl http://localhost:8080/healthz`
3. Check database connectivity
4. Verify Gmail API credentials 