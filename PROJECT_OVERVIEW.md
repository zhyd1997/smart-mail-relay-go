# Smart Mail Relay Service - Project Overview

## 🚀 Complete Go Service for Intelligent Email Forwarding

This is a production-ready Go service that automatically forwards emails based on keyword matching rules. The service supports both Gmail API (OAuth2) and IMAP for email fetching, with comprehensive monitoring, logging, and a REST API for management.

## 📁 Project Structure

```
smart-mail-relay-go/
├── cmd/
│   └── api/
│       └── main.go                 # Application entry point
├── config/
│   ├── config.go                   # Viper configuration
│   └── config.yaml.example         # Sample configuration
├── internal/
│   ├── database/                   # Database connection setup
│   ├── handler/                    # HTTP handlers
│   ├── metrics/                    # Prometheus metrics
│   ├── model/                      # GORM models
│   ├── repository/                 # Data access layer
│   ├── router/                     # Gin router
│   └── service/                    # Mail and scheduler services
├── tools/
│   └── get_token.go                # OAuth2 token helper utility
├── docker-compose.yml              # Multi-service Docker setup
├── Dockerfile                      # Multi-stage Docker build
├── Makefile                        # Development and deployment tasks
├── README.md                       # Comprehensive documentation
├── SETUP.md                        # Step-by-step setup guide
└── main_test.go                    # Unit tests
```

## 🏗️ Architecture Components

### 1. **Mail Service** (`internal/service/mail_service.go`)
- Combines email fetching, parsing, and forwarding
- Supports Gmail API and IMAP
- Includes idempotent processing and logging

### 2. **Scheduler Service** (`internal/service/scheduler_service.go`)
- Cron-based periodic processing
- Configurable intervals and graceful shutdown
- Manual trigger support

### 3. **REST API Layer** (`internal/router`, `internal/handler`)
- Gin router mapping to rule, log, and scheduler handlers
- Health check and metrics endpoints
- Comprehensive error handling

### 4. **Database Layer** (`internal/database`, `internal/repository`, `internal/model`)
- Database connection management
- GORM models and repository pattern
- MySQL persistence

### 5. **Metrics** (`internal/metrics`)
- Prometheus metrics collection
- Service monitoring and observability

### 6. **Configuration** (`config`)
- Viper-based configuration loading
- Environment variable overrides

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
go build -o smart-mail-relay ./cmd/api
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