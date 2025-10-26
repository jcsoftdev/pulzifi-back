# Testing Guide - Organization Module

This guide shows how to test the REST API and gRPC endpoints for the organization module.

## Prerequisites

1. Start the services:
```bash
make start
```

2. For gRPC testing, install `grpcurl`:
```bash
brew install grpcurl
```

3. For REST testing, you can use `curl` or Postman

---

## REST API Testing

### 1. Create Organization (POST)

```bash
curl -X POST http://localhost:8080/api/organizations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "name": "Acme Corp",
    "subdomain": "acme-corp"
  }'
```

**Success Response (201 Created):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Acme Corp",
  "subdomain": "acme-corp",
  "schema_name": "acme_corp",
  "created_at": "2025-10-25T18:30:00Z"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid input (name too short, invalid subdomain)
- `401 Unauthorized` - Missing or invalid JWT token
- `409 Conflict` - Subdomain already exists
- `500 Internal Server Error` - Database or other errors

### 2. Get Organization (GET)

```bash
curl -X GET http://localhost:8080/api/organizations/550e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Success Response (200 OK):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Acme Corp",
  "subdomain": "acme-corp",
  "schema_name": "acme_corp",
  "owner_user_id": "770e8400-e29b-41d4-a716-446655440000",
  "created_at": "2025-10-25T18:30:00Z",
  "updated_at": "2025-10-25T18:30:00Z"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid UUID format
- `401 Unauthorized` - Missing JWT token
- `404 Not Found` - Organization not found
- `500 Internal Server Error` - Database errors

### 3. Health Check (GET)

```bash
curl -X GET http://localhost:8080/health
```

**Response:**
```json
{
  "status": "healthy"
}
```

---

## gRPC Testing with grpcurl

The organization module exposes gRPC on port `9000`.

### 1. Create Organization (CreateOrganization RPC)

```bash
grpcurl -plaintext \
  -d '{
    "name": "Tech Startup",
    "subdomain": "tech-startup",
    "owner_user_id": "770e8400-e29b-41d4-a716-446655440000"
  }' \
  localhost:9000 organization.OrganizationService/CreateOrganization
```

**Success Response:**
```json
{
  "organization": {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "name": "Tech Startup",
    "subdomain": "tech-startup",
    "schema_name": "tech_startup",
    "owner_user_id": "770e8400-e29b-41d4-a716-446655440000",
    "created_at": "2025-10-25T18:35:00Z"
  }
}
```

### 2. Get Organization (GetOrganization RPC)

```bash
grpcurl -plaintext \
  -d '{
    "id": "550e8400-e29b-41d4-a716-446655440001"
  }' \
  localhost:9000 organization.OrganizationService/GetOrganization
```

**Success Response:**
```json
{
  "organization": {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "name": "Tech Startup",
    "subdomain": "tech-startup",
    "schema_name": "tech_startup",
    "owner_user_id": "770e8400-e29b-41d4-a716-446655440000",
    "created_at": "2025-10-25T18:35:00Z",
    "updated_at": "2025-10-25T18:35:00Z"
  }
}
```

### 3. Get Organization by Subdomain (GetOrganizationBySubdomain RPC)

```bash
grpcurl -plaintext \
  -d '{
    "subdomain": "tech-startup"
  }' \
  localhost:9000 organization.OrganizationService/GetOrganizationBySubdomain
```

### 4. List Available gRPC Services

```bash
grpcurl -plaintext localhost:9000 list
```

**Response:**
```
organization.OrganizationService
grpc.reflection.v1.ServerReflection
grpc.reflection.v1alpha.ServerReflection
```

### 5. Get Service Details

```bash
grpcurl -plaintext localhost:9000 describe organization.OrganizationService
```

**Response:**
```
organization.OrganizationService is a service:
  rpc CreateOrganization ( .organization.CreateOrganizationRequest ) returns ( .organization.CreateOrganizationReply );
  rpc GetOrganization ( .organization.GetOrganizationRequest ) returns ( .organization.GetOrganizationReply );
  rpc GetOrganizationBySubdomain ( .organization.GetOrganizationBySubdomainRequest ) returns ( .organization.GetOrganizationBySubdomainReply );
```

---

## Kafka Event Testing

### 1. Monitor Kafka Events

The organization module publishes events to these topics:
- `organization.created` - When a new organization is created
- `organization.deleted` - When an organization is deleted
- `organization.updated` - When an organization is updated

To monitor events in real-time:

```bash
# Install kcat (or kafka-console-consumer)
brew install kcat

# Monitor organization.created events
kcat -b localhost:9092 -t organization.created -C

# Monitor organization.deleted events
kcat -b localhost:9092 -t organization.deleted -C

# Monitor organization.updated events
kcat -b localhost:9092 -t organization.updated -C
```

### 2. Event Payload Example

When you create an organization via REST or gRPC, the following event is published:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440001",
  "name": "Tech Startup",
  "subdomain": "tech-startup",
  "schema_name": "tech_startup",
  "owner_user_id": "770e8400-e29b-41d4-a716-446655440000",
  "created_at": "2025-10-25T18:35:00Z"
}
```

---

## Using Postman for REST API Testing

### 1. Create a New Request

- **Method:** POST
- **URL:** `http://localhost:8080/api/organizations`
- **Headers:**
  - `Content-Type: application/json`
  - `Authorization: Bearer YOUR_JWT_TOKEN`
- **Body (raw JSON):**
```json
{
  "name": "Postman Test Org",
  "subdomain": "postman-test"
}
```

### 2. Save as Collection

Save these requests as a Postman collection for reuse:
1. Click "Save as Collection"
2. Name: "Organization Module Tests"
3. Import using collection import feature

---

## Using Go Client (gRPC)

If you want to test from Go code:

```go
package main

import (
	"context"
	pb "github.com/jcsoftdev/pulzifi-back/modules/organization/infrastructure/grpc/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Connect to gRPC server
	conn, err := grpc.Dial("localhost:9000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	client := pb.NewOrganizationServiceClient(conn)

	// Create organization
	resp, err := client.CreateOrganization(context.Background(), &pb.CreateOrganizationRequest{
		Name:        "Go Test Org",
		Subdomain:   "go-test",
		OwnerUserId: "770e8400-e29b-41d4-a716-446655440000",
	})
	if err != nil {
		panic(err)
	}

	println("Created organization:", resp.Organization.Id)

	// Get organization
	getResp, err := client.GetOrganization(context.Background(), &pb.GetOrganizationRequest{
		Id: resp.Organization.Id,
	})
	if err != nil {
		panic(err)
	}

	println("Retrieved organization:", getResp.Organization.Name)
}
```

---

## Testing Checklist

- [ ] REST: Create organization with valid data (201)
- [ ] REST: Create with duplicate subdomain (409)
- [ ] REST: Create with invalid name (400)
- [ ] REST: Get organization (200)
- [ ] REST: Get non-existent organization (404)
- [ ] REST: Missing auth header (401)
- [ ] gRPC: Create organization via gRPC
- [ ] gRPC: Get organization via gRPC
- [ ] gRPC: Verify event published to Kafka
- [ ] Kafka: Monitor organization.created topic receives event
- [ ] Kafka: Verify event payload format
- [ ] Health: Verify /health endpoint returns 200

---

## Troubleshooting

### REST Endpoints Not Responding

```bash
# Check if HTTP server is running
curl http://localhost:8080/health

# Check logs for errors
docker logs organization-service
```

### gRPC Connection Refused

```bash
# Verify gRPC server is listening
netstat -an | grep 9000

# Try connecting with grpcurl
grpcurl -plaintext localhost:9000 list
```

### Kafka Events Not Appearing

```bash
# Check if Kafka is running
docker-compose ps | grep kafka

# Verify event publishing in logs
docker logs organization-service | grep "Published"

# List all Kafka topics
kafka-topics --bootstrap-server localhost:9092 --list
```

### JWT Token Issues

For testing without proper JWT setup:
1. Use a mock token generator
2. Disable JWT validation in development
3. Extract user_id manually from the request context

---

## Performance Testing

### Load Testing REST API

```bash
# Install Apache Bench
brew install httpd

# Run 1000 requests with 10 concurrent
ab -n 1000 -c 10 -p payload.json -T application/json http://localhost:8080/api/organizations
```

### Load Testing gRPC

```bash
# Install ghz
brew install ghz

# Create ghz scenario
ghz --insecure \
  --proto ./infrastructure/grpc/proto/organization.proto \
  --call organization.OrganizationService/CreateOrganization \
  -d @ localhost:9000 << 'EOF'
{
  "name": "Load Test",
  "subdomain": "load-test-{random}",
  "owner_user_id": "770e8400-e29b-41d4-a716-446655440000"
}
EOF
```

---

## Notes

- All timestamps are in ISO 8601 format (UTC)
- UUIDs are in standard format: `550e8400-e29b-41d4-a716-446655440000`
- Subdomain must be 3-63 characters, alphanumeric + hyphens
- Organization name must be 2-255 characters
- All requests require authentication (JWT bearer token) for REST API
- gRPC server operates independently (inter-module communication)
