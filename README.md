# Coupon System MVP

A robust coupon management system for a medicine ordering platform, built with Go and designed for production use.

## Features

- Admin coupon creation and management
- Coupon validation with multiple constraints
- Support for different usage types (one-time, multi-use, time-based)
- Concurrent validation handling
- Redis caching for performance
- PostgreSQL for persistent storage
- Dockerized deployment
- OpenAPI documentation

## Architecture

### Components

1. **Domain Layer**
   - Core business logic and entities
   - Validation rules and business constraints
   - Located in `internal/domain/`

2. **Service Layer**
   - Business logic implementation
   - Transaction management
   - Located in `internal/service/`

3. **Infrastructure Layer**
   - Database connections and migrations
   - Redis caching
   - Located in `internal/infrastructure/`

4. **API Layer**
   - HTTP handlers
   - Request/response handling
   - Located in `internal/api/`

### Concurrency & Caching

- **Concurrency Handling**
  - Distributed locking using Redis
  - Mutex for concurrent validations
  - Database transactions for data consistency

- **Caching Strategy**
  - Redis-based caching with TTL
  - Cache invalidation on coupon updates
  - LRU eviction policy

## Installation

### Prerequisites
- Go 1.21 or later
- Docker and Docker Compose
- PostgreSQL 15 or later
- Redis 7 or later

### Local Development Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/yourusername/coupon-system.git
   cd coupon-system
   ```

2. **Install Go dependencies**
   ```bash
   go mod download
   ```

3. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Run the application**
   ```bash
   go run cmd/main.go
   ```

### Docker Setup

1. **Build and run with Docker Compose**
   ```bash
   docker compose up --build
   ```

2. **Access the application**
   - API: http://localhost:8080
   - Swagger UI: http://localhost:8080/swagger/index.html

## API Examples

### Get Applicable Coupons

```bash
curl -X GET http://localhost:8080/api/v1/coupons/applicable \
  -H "Content-Type: application/json" \
  -d '{
    "medicine_ids": ["med1", "med2"],
    "categories": ["pain-relief", "vitamins"],
    "order_value": 150.00,
    "user_id": "user123"
  }'
```

Response:
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "code": "SUMMER20",
    "expiry_date": "2024-12-31T23:59:59Z",
    "usage_type": "multi_use",
    "applicable_medicine_ids": ["med1", "med2"],
    "applicable_categories": ["pain-relief"],
    "min_order_value": 100.00,
    "discount_type": "percentage",
    "discount_value": 20.00,
    "max_usage_per_user": 3
  }
]
```

### Validate Coupon

```bash
curl -X POST http://localhost:8080/api/v1/coupons/validate \
  -H "Content-Type: application/json" \
  -d '{
    "code": "SUMMER20",
    "medicine_ids": ["med1", "med2"],
    "categories": ["pain-relief"],
    "order_value": 150.00,
    "user_id": "user123"
  }'
```

Response:
```json
{
  "is_valid": true,
  "message": "Coupon is valid",
  "discount": 30.00,
  "final_amount": 120.00
}
```

## Rate Limiting

The API implements rate limiting using Redis:
- 100 requests per minute per IP address
- Rate limit headers included in responses:
  - `X-RateLimit-Limit`: Maximum requests allowed
  - `X-RateLimit-Remaining`: Remaining requests
  - `X-RateLimit-Reset`: Time until limit resets

Example response headers:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 99
X-RateLimit-Reset: 1625097600
```

## Health Monitoring

The system includes a health check endpoint that monitors:
- Database connectivity
- Redis connectivity
- System status

Example health check response:
```json
{
  "status": "healthy",
  "timestamp": "2024-03-15T10:30:00Z",
  "services": {
    "database": "healthy",
    "redis": "healthy"
  }
}
```

## Development

### Running Tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test ./... -cover

# Run specific test package
go test ./internal/handler/...
```

### Database Migrations
```bash
# Run migrations
go run cmd/migrate/main.go
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Code Style
- Follow Go standard formatting (`go fmt`)
- Run linter (`golangci-lint run`)
- Write tests for new features
- Update documentation

## Production Considerations

1. **Security**
   - Environment variables for sensitive data
   - Input validation
   - Rate limiting
   - HTTPS/TLS configuration
   - API key authentication

2. **Monitoring**
   - Health check endpoints
   - Metrics collection
   - Logging
   - Error tracking
   - Performance monitoring

3. **Scalability**
   - Horizontal scaling support
   - Connection pooling
   - Caching strategy
   - Load balancing
   - Database sharding

## License

MIT License - see the [LICENSE](LICENSE) file for details 