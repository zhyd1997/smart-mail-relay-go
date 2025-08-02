# Smart Mail Relay Service - Project Overview

## 🚀 Complete Go Service for Intelligent Email Forwarding

This is a production-ready Go service that automatically forwards emails based on keyword matching rules. The service supports both Gmail API (OAuth2) and IMAP for email fetching, with comprehensive monitoring, logging, and a REST API for management.

## 📁 Project Structure

```
smart-mail-relay-go/
├── main.go                 # Application entry point with graceful shutdown
├── config.go              # Configuration management with Viper
├── models.go              # GORM database models
├── fetcher.go             # Email fetching (Gmail API + IMAP)
├── parser.go              # Email parsing and keyword extraction
├── forwarder.go           # Email forwarding via Gmail API
├── scheduler.go           # Cron-based job scheduling
├── handlers.go            # HTTP REST API handlers
├── metrics.go             # Prometheus metrics collection
├── main_test.go           # Unit tests
├── config.yaml            # Configuration file
├── docker-compose.yml     # Multi-service Docker setup
├── Dockerfile             # Multi-stage Docker build
├── init.sql               # Database initialization
├── prometheus.yml         # Prometheus configuration
├── Makefile               # Development and deployment tasks
├── README.md              # Comprehensive documentation
├── SETUP.md               # Step-by-step setup guide
├── .gitignore             # Git ignore rules
└── tools/
    └── get_token.go       # OAuth2 token helper utility
```

## 🏗️ Architecture Components

### 1. **Email Fetcher** (`fetcher.go`)
- **Gmail API Fetcher**: Uses OAuth2 for secure access
- **IMAP Fetcher**: Alternative method using IMAP protocol
- Supports both methods with configurable switching
- Handles rate limiting and exponential backoff

### 2. **Email Parser** (`parser.go`)
- Extracts keywords from email subjects
- Pattern: `<keyword> - <recipient_name>`
- Multiple matching strategies (exact, case-insensitive, partial)
- Ensures idempotency with processed email tracking

### 3. **Email Forwarder** (`forwarder.go`)
- Forwards emails via Gmail API
- Preserves original email structure and headers
- HTML to plain text conversion
- Retry logic with exponential backoff

### 4. **Scheduler** (`scheduler.go`)
- Cron-based periodic processing
- Configurable intervals (default: 5 minutes)
- Graceful shutdown handling
- Manual trigger support

### 5. **HTTP Server** (`handlers.go`)
- RESTful API for rule management
- Health check endpoints
- Prometheus metrics endpoint
- Comprehensive error handling

### 6. **Database Layer** (`models.go`)
- **forward_rules**: Email forwarding rules
- **processed_emails**: Idempotency tracking
- **forward_logs**: Audit trail and monitoring
- GORM with MySQL support

## 🗄️ Database Schema

### Tables

1. **forward_rules**
   ```sql
   - id (Primary Key, Auto Increment)
   - keyword (Unique, Indexed)
   - target_email
   - enabled (Boolean)
   - created_at, updated_at
   - deleted_at (Soft Delete)
   ```

2. **processed_emails**
   ```sql
   - id (Primary Key, Auto Increment)
   - message_id (Unique, Indexed)
   - processed_at
   - deleted_at (Soft Delete)
   ```

3. **forward_logs**
   ```sql
   - id (Primary Key, Auto Increment)
   - message_id (Indexed)
   - rule_id (Foreign Key, Indexed)
   - status (success/failure/skipped/error)
   - error_msg
   - created_at
   - deleted_at (Soft Delete)
   ```

## 🔧 Configuration Management

### Environment Variables
- Database connection settings
- Gmail OAuth2 credentials
- IMAP settings (alternative)
- Scheduler configuration
- Server settings

### Configuration File
- YAML-based configuration
- Environment variable override support
- Default values for all settings

## 📊 Monitoring & Observability

### Prometheus Metrics
- `smart_mail_relay_pull_count`: Email fetch operations
- `smart_mail_relay_match_count`: Successful rule matches
- `smart_mail_relay_forward_successes`: Successful forwards
- `smart_mail_relay_forward_failures`: Failed forwards
- `smart_mail_relay_processing_duration_seconds`: Processing time
- `smart_mail_relay_active_rules`: Active rule count
- `smart_mail_relay_total_rules`: Total rule count

### Health Checks
- Database connectivity
- Gmail API connectivity
- Scheduler status
- Service health endpoint

### Logging
- Structured JSON logging with Logrus
- Configurable log levels
- Request/response logging
- Error tracking and debugging

## 🔌 REST API Endpoints

### Health & Monitoring
- `GET /healthz` - Service health check
- `GET /metrics` - Prometheus metrics

### Forwarding Rules
- `GET /api/v1/rules` - List all rules
- `POST /api/v1/rules` - Create new rule
- `GET /api/v1/rules/{id}` - Get specific rule
- `PUT /api/v1/rules/{id}` - Update rule
- `DELETE /api/v1/rules/{id}` - Delete rule
- `PATCH /api/v1/rules/{id}/enable` - Enable rule
- `PATCH /api/v1/rules/{id}/disable` - Disable rule

### Forward Logs
- `GET /api/v1/logs` - List logs with pagination
- `GET /api/v1/logs/{id}` - Get specific log

### Scheduler Control
- `POST /api/v1/scheduler/start` - Start scheduler
- `POST /api/v1/scheduler/stop` - Stop scheduler
- `POST /api/v1/scheduler/run-once` - Manual run
- `GET /api/v1/scheduler/status` - Scheduler status

## 🐳 Docker Support

### Multi-Service Setup
- **MySQL**: Database with initialization
- **Application**: Main service with health checks
- **Prometheus**: Metrics collection
- **Grafana**: Metrics visualization

### Features
- Health checks for all services
- Volume persistence for data
- Environment variable configuration
- Non-root user execution
- Multi-stage builds for optimization

## 🧪 Testing

### Unit Tests
- Configuration validation
- Database DSN generation
- Email parsing logic
- Model validation
- All tests passing ✅

### Test Coverage
- Core business logic
- Configuration handling
- Data model validation
- Utility functions

## 🚀 Deployment Options

### 1. Docker Compose (Recommended)
```bash
docker-compose up -d
```

### 2. Manual Deployment
```bash
go build -o smart-mail-relay .
./smart-mail-relay
```

### 3. Kubernetes Ready
- ConfigMap for configuration
- Secret for credentials
- Deployment with health checks
- Service and ingress setup

## 🔒 Security Features

### Authentication & Authorization
- OAuth2 for Gmail API access
- Secure credential management
- Environment variable secrets

### Data Protection
- Database connection encryption
- Secure email forwarding
- Audit logging for all operations

### Network Security
- HTTPS support (production)
- Database access restrictions
- Rate limiting and backoff

## 📈 Scalability Features

### Performance
- Connection pooling
- Efficient database queries
- Asynchronous processing
- Configurable batch sizes

### Reliability
- Idempotent operations
- Retry mechanisms
- Graceful error handling
- Circuit breaker patterns

### Monitoring
- Real-time metrics
- Performance tracking
- Error rate monitoring
- Resource utilization

## 🛠️ Development Tools

### Makefile Commands
- `make build` - Build application
- `make test` - Run tests
- `make docker-build` - Build Docker image
- `make docker-run` - Start with docker-compose
- `make get-token` - Generate OAuth2 token

### Code Quality
- Go modules for dependency management
- Linting with golangci-lint
- Code formatting with gofmt
- Comprehensive error handling

## 📚 Documentation

### Guides
- **README.md**: Comprehensive project overview
- **SETUP.md**: Step-by-step setup instructions
- **PROJECT_OVERVIEW.md**: This architecture overview

### Examples
- API usage examples
- Configuration samples
- Docker deployment
- Production setup

## 🎯 Key Features Summary

✅ **Email Fetching**: Gmail API + IMAP support  
✅ **Keyword Matching**: Intelligent subject parsing  
✅ **Idempotent Processing**: No duplicate forwards  
✅ **Scheduled Processing**: Configurable intervals  
✅ **REST API**: Full CRUD operations  
✅ **Health Monitoring**: Comprehensive checks  
✅ **Prometheus Metrics**: Production monitoring  
✅ **Graceful Shutdown**: Signal handling  
✅ **Docker Support**: Complete containerization  
✅ **Database Persistence**: MySQL with GORM  
✅ **Error Handling**: Robust error management  
✅ **Logging**: Structured JSON logs  
✅ **Testing**: Unit test coverage  
✅ **Documentation**: Comprehensive guides  

## 🚀 Ready for Production

This service is production-ready with:
- Comprehensive error handling
- Monitoring and observability
- Security best practices
- Scalable architecture
- Complete documentation
- Docker containerization
- Health checks and logging

The service can be deployed immediately and scaled as needed for production workloads. 