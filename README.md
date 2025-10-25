# Pulzifi Backend - Docker Setup

Complete Docker configuration for Pulzifi backend with microservices architecture.

## 📋 What's Included

✅ **PostgreSQL 17** - Multi-tenant database with automatic schema creation
✅ **Redis 7** - Caching and task queue (Asynq)
✅ **Kafka & Zookeeper** - Event streaming for inter-module communication
✅ **Nginx** - API Gateway with rate limiting and SSL/TLS
✅ **7 Microservices** - Auth, Organization, Workspace, Page, Alert, Monitoring, Insight
✅ **Kafka UI** - Visual Kafka cluster management
✅ **Docker Compose** - Complete orchestration
✅ **Makefile** - Convenient commands for development
✅ **Health Checks** - All services with health monitoring
✅ **Production Config** - Ready for production deployment

## 🚀 Quick Start

### 1-Command Setup (Fastest)

```bash
bash scripts/quick-start.sh
```

This script will:
- Check Docker installation
- Create `.env` file
- Generate SSL certificates
- Build Docker images
- Start all services
- Run health checks

### Manual Setup

```bash
# 1. Setup environment
cp .env.example .env

# 2. Generate SSL certificates
bash scripts/generate-ssl.sh

# 3. Build and start
make docker-build
make docker-up

# 4. Check health
make docker-health
```

## 📊 Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Frontend (React)                     │
└──────────────────────────┬──────────────────────────────┘
                           │
                    HTTP/HTTPS Request
                   (X-Tenant header)
                           │
        ┌──────────────────▼──────────────────┐
        │   Nginx API Gateway (Port 80/443)   │
        │    - Route by path (/api/*)         │
        │    - SSL/TLS termination            │
        │    - Rate limiting                  │
        │    - Header extraction              │
        └──────────────────┬──────────────────┘
                           │
        ┌──────────────────┴──────────────────┐
        │                                     │
        ├─→ /api/auth/*           ── Auth Service (8080/9080)
        ├─→ /api/organizations/*  ── Organization Service (8082/9082)
        ├─→ /api/workspaces/*     ── Workspace Service (8083/9083)
        ├─→ /api/pages/*          ── Page Service (8084/9084)
        ├─→ /api/alerts/*         ── Alert Service (8085/9085)
        ├─→ /api/monitoring/*     ── Monitoring Service (8086/9086)
        └─→ /api/insights/*       ── Insight Service (8087/9087)
                           │
        ┌──────────────────┴──────────────────┐
        │                                     │
        ├─→ PostgreSQL (5434)                │ Data Storage
        ├─→ Redis (6379)                     │ Cache/Queue
        ├─→ Kafka (9092)                     │ Events
        └─→ Zookeeper (2181)                 │ Coordination
```

## 🎯 Service Endpoints

### Through API Gateway (Recommended)

| Service | Endpoint | HTTP | gRPC |
|---------|----------|------|------|
| **Auth** | http://localhost/api/auth | 8080 | 9080 |
| **Organization** | http://localhost/api/organizations | 8082 | 9082 |
| **Workspace** | http://localhost/api/workspaces | 8083 | 9083 |
| **Page** | http://localhost/api/pages | 8084 | 9084 |
| **Alert** | http://localhost/api/alerts | 8085 | 9085 |
| **Monitoring** | http://localhost/api/monitoring | 8086 | 9086 |
| **Insight** | http://localhost/api/insights | 8087 | 9087 |

### Supporting Services

| Service | Port | URL |
|---------|------|-----|
| PostgreSQL | 5434 | `postgres://localhost:5434` |
| Redis | 6379 | `redis://localhost:6379` |
| Kafka | 9092 | `localhost:9092` |
| Zookeeper | 2181 | `localhost:2181` |
| **Kafka UI** | 8081 | http://localhost:8081 |

## 📝 Common Commands

```bash
# View all available commands
make help

# Start/Stop
make docker-up          # Start all services
make docker-down        # Stop all services
make docker-restart     # Restart all services

# Monitoring
make docker-logs        # View all logs
make docker-logs-f      # Follow logs (real-time)
make docker-ps          # List containers
make docker-health      # Check service health

# Database
make database-shell     # Connect to PostgreSQL
make database-init      # Initialize schema
make database-backup    # Backup database
make database-restore   # Restore from backup

# Redis
make redis-shell        # Connect to Redis CLI
make redis-flush        # Clear Redis cache

# Kafka
make kafka-topics       # List topics
make kafka-create-topic # Create new topic
make kafka-delete-topic # Delete topic

# Development
make test              # Run tests
make test-coverage     # With coverage report
make build             # Build project
make lint              # Run linter
make fmt               # Format code
```

## 🔧 Configuration

### Environment Variables

Copy `.env.example` to `.env` and update as needed:

```bash
# Database
DB_HOST=postgres
DB_PORT=5434
DB_NAME=pulzifi
DB_USER=pulzifi_user
DB_PASSWORD=pulzifi_password

# Redis
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=redis_password

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRATION=3600

# External Services
SENDGRID_API_KEY=...
AI_API_KEY=...
SLACK_BOT_TOKEN=...
```

See `.env.example` for all available options.

## 🔐 SSL Certificates

### Development (Self-Signed)

Already generated in `quick-start.sh`. To regenerate:

```bash
bash scripts/generate-ssl.sh
```

### Production (Let's Encrypt)

```bash
bash scripts/generate-ssl.sh production
```

## 🏥 Health Checks

All services include health check endpoints:

```bash
# Quick health status
make docker-health

# Detailed health check
bash scripts/health-check.sh

# Individual service health
curl http://localhost:8080/health      # Auth
curl http://localhost:8082/health      # Organization
curl http://localhost:8083/health      # Workspace
# ... etc
```

## 📊 Monitoring

### View Resource Usage

```bash
# Real-time resource usage
docker stats

# Disk usage summary
docker system df

# Network usage
docker stats --format "table {{.Container}}\t{{.MemUsage}}\t{{.CPUPerc}}"
```

### Kafka UI

Open http://localhost:8081 to:
- View topics and partitions
- Monitor consumer groups
- Inspect message content
- Check broker health

### Database Stats

```bash
make database-shell

# In psql:
\dt                    # List tables
SELECT COUNT(*) FROM pg_tables WHERE schemaname='public';  # Table count
SELECT schemaname FROM information_schema.schemata;        # List schemas
```

## 🐛 Troubleshooting

### Services Not Starting

```bash
# Check logs
make docker-logs

# Check specific service
make docker-logs-auth-service

# Verify Docker is running
docker ps

# Check available disk space
docker system df
```

### Database Connection Issues

```bash
# Test PostgreSQL
docker-compose exec postgres psql -U pulzifi_user -d pulzifi -c "SELECT 1"

# View PostgreSQL logs
make docker-logs-postgres

# Reset database (WARNING: deletes all data)
make database-reset
```

### Port Conflicts

If ports are already in use:

```bash
# Find what's using the port
lsof -i :5434   # PostgreSQL
lsof -i :6379   # Redis
lsof -i :9092   # Kafka
lsof -i :80     # HTTP
lsof -i :443    # HTTPS
```

### Clean Reset

```bash
# Stop and remove everything (WARNING: deletes all data)
make docker-clean

# Rebuild from scratch
make docker-build
make docker-up
```

## 📦 Production Deployment

### 1. Generate Production SSL

```bash
bash scripts/generate-ssl.sh production
```

### 2. Update Environment

Edit `.env` with production values:
- Strong passwords
- Real API keys
- Production URLs
- Email service credentials

### 3. Use Production Compose

```bash
# Build production images
docker-compose -f docker-compose.prod.yml build

# Start production services
docker-compose -f docker-compose.prod.yml up -d
```

### 4. Features in Production Config

- Service replication (2 replicas each)
- Resource limits and reservations
- Structured logging (JSON format)
- Health checks with custom timing
- Backup volumes
- Nginx caching
- Automatic restarts

## 📚 File Structure

```
pulzifi-back/
├── docker-compose.yml           # Development config
├── docker-compose.prod.yml      # Production config
├── Dockerfile                   # Multi-stage build
├── Makefile                     # Convenient commands
├── nginx.conf                   # API Gateway config
├── .env.example                 # Environment template
├── .gitignore                   # Git ignores
│
├── scripts/
│   ├── quick-start.sh          # One-command setup
│   ├── generate-ssl.sh         # SSL certificate generation
│   └── health-check.sh         # Service health verification
│
├── ssl/                         # SSL certificates (auto-generated)
│   ├── cert.pem
│   └── key.pem
│
├── database_setup.sql           # Database schema initialization
└── README.md                    # This file
```

## 🔍 Debugging

### View Raw Logs

```bash
# All services
docker-compose logs

# Specific service
docker-compose logs auth-service

# Follow in real-time
docker-compose logs -f workspace-service

# Last 100 lines
docker-compose logs --tail=100

# With timestamps
docker-compose logs --timestamps
```

### Execute Commands in Containers

```bash
# PostgreSQL
docker-compose exec postgres psql -U pulzifi_user -d pulzifi

# Redis
docker-compose exec redis redis-cli -a redis_password

# Kafka
docker-compose exec kafka kafka-topics --list --bootstrap-server localhost:9092

# Any service
docker-compose exec auth-service /bin/sh
```

### Inspect Container Details

```bash
# Container info
docker inspect pulzifi-auth

# Environment variables
docker inspect pulzifi-postgres --format='{{json .Config.Env}}'

# Network
docker network inspect pulzifi-network

# Volumes
docker volume inspect pulzifi-back_postgres_data
```

## 🎓 Learning Resources

- [Docker Documentation](https://docs.docker.com)
- [Docker Compose Reference](https://docs.docker.com/compose/compose-file/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Redis Documentation](https://redis.io/documentation)
- [Apache Kafka](https://kafka.apache.org/documentation/)
- [Nginx Configuration](https://nginx.org/en/docs/)

## 📄 License

See LICENSE file in repository.

## ✅ Checklist

After setup, verify:

- [ ] Docker is running
- [ ] All services are healthy: `make docker-health`
- [ ] PostgreSQL is initialized: `make database-shell`
- [ ] Can access Kafka UI: http://localhost:8081
- [ ] API Gateway responds: `curl http://localhost/health`
- [ ] Services are logging: `make docker-logs`
- [ ] Backups location exists: `./backups/`

## 🆘 Support

For issues:
1. Check logs: `make docker-logs`
2. Run health check: `make docker-health`

---

**Ready to start?** Run `bash scripts/quick-start.sh` 🚀
