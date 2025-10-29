# Fleet Monitoring Service

A REST API service for monitoring device heartbeats and calculating uptime statistics for a fleet of devices.

## Quick Start

### Prerequisites
- [Go 1.25+](https://golang.org/doc/install)
- Docker (optional)

### Running the Service
```bash
git clone https://github.com/MahumAzeem/fleet-monitoring-service.git
cd fleet-monitoring-service

# Install dependencies and run
go mod tidy
go run cmd/main.go

# Or build and run
go build -o fleet-monitoring cmd/main.go
./fleet-monitoring.exe

# Run all tests
go test ./internal/devices/
```

**Service runs on port 6733** as specified in the API requirements.

### Docker Setup (Alternative)
```bash
# Build and run with Docker
docker build -t fleet-monitoring .
docker run -p 6733:6733 fleet-monitoring

# Or use Docker Compose
docker-compose up --build
```

## API Usage

Submit device heartbeat:
```bash
curl -X POST http://localhost:6733/devices/{deevice_id}/heartbeat
```
Get device statistics:
```bash
curl http://localhost:6733/devices/{device_id}/stats
```

---

## Implementation Overview

## Time Spent & Challenges

**Time:** Approximately 4-5 hours total.

**Background:** I had minimal Go experience before this project - mostly just reading documentation and fixing minor bugs. This was essentially my first real Go application from scratch.

**Most Difficult Parts:**
1. **Learning Go idioms** - I was familiar with basic syntax and have experience with C so the pointers and memory management was no issue. The real learning curve was Go’s project layout, dependency management, and module system.
I intentionally avoided bringing over OOP or Pythonic patterns and focused on composition, interfaces, and small, focused packages. 
2. **Testing in Go** - Learning table-driven tests, and Go's testing conventions while fighting the urge to write pytest-style tests was challenging but rewarding.


## Runtime Complexity

Time complexity: O(1) for heartbeats and stat updates (map lookups).
Space complexity: Grows with the number of unique minutes a device reports a heartbeat.


## Assumptions and Design Choices

### Key Assumptions
1. No persistence required for this demo.
2. Single instance deployment. No distributed system concerns
3. Heartbeat frequency - Devices report roughly once per minute and never out of order

### Design Trade-offs
1. **Simplicity vs Scalability** - Chose simple in-memory maps over databases for faster development
2. **Memory vs Performance** - Store per-minute buckets for O(1) lookups at cost of memory

### Production Scaling

**Infrastructure:**
- **Horizontal scaling:** Load balancer + multiple instances with shared storage
- **Database layer:** Perhaps time-series based storage
- **Message queues:** AWS SQS for async heartbeat processing so we don't lose data if the server crashes
- **Hot data vs historical data** - Recent metrics in memory, older in DB, would cold storage be needed


**Observability:**
- **System metrics:** CPU, memory, request latency, error rates of this system, not the devices
- **Business metrics:** We'd probably want fleet-wide averages too, not just per-device stats
- **Alert thresholds:** Device offline >5min, service error rate >1%, etc

**Security & Operations:**
- **Authentication:** API keys or JWT tokens for device registration
- **Rate limiting:** Prevent devices from spamming heartbeats and bringing down the service

## Data Model Extensions

To support more kinds of metrics, the main challenges are **API design** and **memory management**, not just data structures:

**1. Configurable Metrics:**
Load from config files (as json likely, not structs. Struct shown for example)
```go
type MetricConfig struct {
    Name        string
    Aggregation string  // "avg", "sum", "latest", "count"
    RetentionDays int
}
```

**2. Query Flexibility:**
Instead of: GET /devices/{id}/stats (fixed response)
Add: GET /devices/{id}/metrics?types=cpu,memory&window=1h




