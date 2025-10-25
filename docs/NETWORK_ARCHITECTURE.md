# Pulzifi Network Architecture

## Overview

Pulzifi usa una arquitectura de microservicios donde **cada módulo expone su propia REST API** y se comunica con otros módulos via **gRPC**. Un **Load Balancer** (Nginx/Traefik) extrae el tenant desde el subdomain y enruta requests a los módulos correspondientes.

---

## Arquitectura de Red

```
┌─────────────────────────────────────────────────────────────────┐
│                          INTERNET                                │
└──────────────────────────────┬──────────────────────────────────┘
                               │
                               │ HTTPS (SSL/TLS)
                               │
┌──────────────────────────────▼──────────────────────────────────┐
│              LOAD BALANCER / API GATEWAY                         │
│              (Nginx / Traefik / Kong)                            │
│                                                                  │
│  Responsabilidades:                                              │
│  ✅ Terminar SSL/TLS                                             │
│  ✅ Extraer subdomain (jcsoftdev-inc.pulzifi.com)              │
│  ✅ Normalizar a schema name (jcsoftdev_inc)                   │
│  ✅ Inyectar header: X-Tenant: jcsoftdev_inc                   │
│  ✅ Enrutar por path a módulos                                  │
│  ✅ Rate limiting por tenant                                     │
│                                                                  │
│  Rutas:                                                          │
│  /api/auth/*        → auth module:8080                          │
│  /api/organizations/* → organization module:8081                │
│  /api/workspaces/*  → workspace module:8082                     │
│  /api/pages/*       → page module:8083                          │
│  /api/alerts/*      → alert module:8084                         │
│  /api/insights/*    → insight module:8085                       │
│  /api/reports/*     → report module:8086                        │
│  /api/integrations/*→ integration module:8087                   │
│  /api/usage/*       → usage module:8088                         │
│                                                                  │
└───────┬──────────┬──────────┬──────────┬──────────┬─────────────┘
        │          │          │          │          │
        │ HTTP     │ HTTP     │ HTTP     │ HTTP     │ HTTP
        │ +Header  │ +Header  │ +Header  │ +Header  │ +Header
        │ X-Tenant │ X-Tenant │ X-Tenant │ X-Tenant │ X-Tenant
        │          │          │          │          │
┌───────▼──────┐ ┌─▼────────┐ ┌─▼──────┐ ┌─▼──────┐ ┌─▼──────────┐
│ Auth Module  │ │Workspace │ │  Page  │ │ Alert  │ │   More...  │
│   :8080      │ │Module    │ │ Module │ │ Module │ │            │
│              │ │  :8082   │ │ :8083  │ │ :8084  │ │            │
│  ┌────────┐  │ │          │ │        │ │        │ │            │
│  │HTTP    │  │ │┌────────┐│ │┌──────┐│ │┌──────┐│ │            │
│  │Server  │◄─┼─┼│HTTP    ││ ││HTTP  ││ ││HTTP  ││ │            │
│  └────────┘  │ ││Server  ││ ││Server││ ││Server││ │            │
│              │ │└────────┘│ │└──────┘│ │└──────┘│ │            │
│  ┌────────┐  │ │          │ │        │ │        │ │            │
│  │gRPC    │  │ │┌────────┐│ │┌──────┐│ │┌──────┐│ │            │
│  │Server  │  │ ││gRPC    ││ ││gRPC  ││ ││gRPC  ││ │            │
│  └────┬───┘  │ ││Server  ││ ││Server││ ││Server││ │            │
│       │      │ │└───┬────┘│ │└───┬──┘│ │└───┬──┘│ │            │
└───────┼──────┘ └────┼─────┘ └────┼───┘ └────┼───┘ └────────────┘
        │             │            │           │
        └─────────────┴────────────┴───────────┴────────────────────
                                │
                            gRPC Network
                    (Inter-Module Communication)
                                │
        ┌───────────────────────┴───────────────────────┐
        │                                                │
┌───────▼─────────┐                            ┌────────▼────────┐
│  Kafka Cluster  │                            │  Redis Cluster  │
│  (Events)       │                            │  (Asynq Jobs)   │
│                 │                            │                 │
│  Topics:        │                            │  Queues:        │
│  - check_completed                           │  - scheduled_checks│
│  - alert_created│                            │  - email_sending│
│  - etc.         │                            │  - ai_processing│
└─────────────────┘                            └─────────────────┘
```

---

## Flujo de Request Frontend → Backend

### Ejemplo: Crear un Page en un Workspace

```
1. Frontend hace request:
   POST https://jcsoftdev-inc.pulzifi.com/api/pages
   Headers:
     Authorization: Bearer <jwt_token>
   Body:
     {
       "workspace_id": "uuid-123",
       "url": "https://toyota.com",
       "name": "Toyota Homepage"
     }

2. Load Balancer (Nginx):
   - Extrae subdomain: "jcsoftdev-inc"
   - Normaliza a schema name: "jcsoftdev_inc"
   - Enruta por path /api/pages → page module (port 8083)
   - Inyecta headers:
       X-Tenant: jcsoftdev_inc
       Authorization: Bearer <jwt_token>

3. Page Module - HTTP Server (port 8083):
   - Middleware extrae tenant desde header X-Tenant
   - Middleware valida JWT y extrae user_id
   - Middleware inyecta tenant en context: ctx = context.WithValue(ctx, "tenant", "jcsoftdev_inc")
   - Handler invoca application layer: create_page.Handle(ctx, request)

4. Page Module - Application Layer:
   - create_page/handler.go orquesta el caso de uso
   - Valida que workspace existe (call workspace module via gRPC)
   - Crea page entity

5. Page Module - Repository Layer:
   - Extrae tenant desde context
   - Ejecuta: SET search_path TO jcsoftdev_inc
   - INSERT INTO pages (workspace_id, url, name, ...)

6. Page Module - Response:
   - HTTP 201 Created
   - Body: { "id": "page-uuid", "name": "Toyota Homepage", ... }

7. Load Balancer → Frontend:
   - Respuesta final al frontend
```

---

## Comunicación Entre Módulos

### HTTP (Frontend ↔ Module)
- **Propósito:** Frontend consume REST API de cada módulo
- **Puerto:** Cada módulo expone HTTP server en puerto diferente (8080, 8081, 8082, ...)
- **Tenant:** Extraído desde header `X-Tenant` (inyectado por Load Balancer)
- **Autenticación:** JWT token en header `Authorization`

### gRPC (Module ↔ Module)
- **Propósito:** Comunicación síncrona entre módulos
- **Puerto:** Cada módulo expone gRPC server en puerto diferente (9080, 9081, 9082, ...)
- **Tenant:** Incluido en gRPC metadata
- **Autenticación:** Service-to-service auth (API keys o mutual TLS)

**Ejemplo:**
```go
// Page module necesita validar que workspace existe
// Llama a workspace module via gRPC

// 1. Page module (client side)
import "modules/workspace/infrastructure/grpc/proto"

conn, _ := grpc.Dial("workspace-service:9082")
client := proto.NewWorkspaceServiceClient(conn)

// Incluir tenant en metadata
ctx := metadata.AppendToOutgoingContext(ctx, "x-tenant", tenant)

response, err := client.GetWorkspace(ctx, &proto.GetWorkspaceRequest{
    WorkspaceId: workspaceID,
})

// 2. Workspace module (server side)
// gRPC interceptor extrae tenant desde metadata
func TenantInterceptor(ctx context.Context, ...) {
    md, _ := metadata.FromIncomingContext(ctx)
    tenant := md.Get("x-tenant")[0]
    ctx = context.WithValue(ctx, "tenant", tenant)
    return handler(ctx, req)
}
```

### Kafka (Module → Module - Async Events)
- **Propósito:** Eventos de dominio asíncronos
- **Tenant:** Incluido en el mensaje JSON
- **No compartir structs:** Cada módulo deserializa a sus propios tipos

**Ejemplo:**
```json
// monitoring module publica evento
{
  "event_type": "check_completed",
  "tenant": "jcsoftdev_inc",
  "check_id": "check-uuid",
  "page_id": "page-uuid",
  "change_detected": true,
  "change_types": ["visual", "text"],
  "timestamp": "2025-10-25T10:30:00Z"
}

// alert module suscribe y procesa
```

---

## Configuración de Load Balancer

### Nginx Configuration Example

```nginx
# nginx.conf

# Upstream servers (módulos backend)
upstream auth_service {
    server auth:8080;
}

upstream workspace_service {
    server workspace:8082;
}

upstream page_service {
    server page:8083;
}

upstream alert_service {
    server alert:8084;
}

# Rate limiting por tenant
limit_req_zone $http_x_tenant zone=per_tenant:10m rate=100r/s;

server {
    listen 443 ssl http2;
    server_name *.pulzifi.com;

    # SSL Configuration
    ssl_certificate /etc/nginx/ssl/cert.pem;
    ssl_certificate_key /etc/nginx/ssl/key.pem;

    # Extraer subdomain y normalizar
    set $tenant "";
    if ($host ~* "^(.+)\.pulzifi\.com$") {
        set $tenant $1;
    }
    
    # Reemplazar guiones por underscores (jcsoftdev-inc → jcsoftdev_inc)
    set $tenant_normalized $tenant;
    if ($tenant ~* "(.+)-(.+)") {
        set $tenant_normalized "${1}_${2}";
    }

    # Rate limiting
    limit_req zone=per_tenant burst=20 nodelay;

    # Headers comunes para todos los upstreams
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header X-Tenant $tenant_normalized;

    # Rutas
    location /api/auth/ {
        proxy_pass http://auth_service;
    }

    location /api/organizations/ {
        proxy_pass http://organization_service;
    }

    location /api/workspaces/ {
        proxy_pass http://workspace_service;
    }

    location /api/pages/ {
        proxy_pass http://page_service;
    }

    location /api/alerts/ {
        proxy_pass http://alert_service;
    }

    # WebSocket support (para alerts en tiempo real)
    location /api/alerts/ws {
        proxy_pass http://alert_service;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }

    # Health checks
    location /health {
        access_log off;
        return 200 "healthy\n";
        add_header Content-Type text/plain;
    }
}

# Default server para subdomains inválidos
server {
    listen 443 ssl http2 default_server;
    server_name _;
    
    ssl_certificate /etc/nginx/ssl/cert.pem;
    ssl_certificate_key /etc/nginx/ssl/key.pem;
    
    return 404 '{"error": "Invalid subdomain"}';
}
```

### Traefik Configuration Example (docker-compose.yml)

```yaml
version: '3.8'

services:
  traefik:
    image: traefik:v2.10
    command:
      - "--api.insecure=true"
      - "--providers.docker=true"
      - "--entrypoints.web.address=:80"
      - "--entrypoints.websecure.address=:443"
    ports:
      - "80:80"
      - "443:443"
      - "8080:8080"  # Traefik dashboard
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro

  auth:
    build: ./modules/auth
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.auth.rule=PathPrefix(`/api/auth/`)"
      - "traefik.http.routers.auth.entrypoints=websecure"
      - "traefik.http.routers.auth.tls=true"
      - "traefik.http.middlewares.tenant-header.headers.customrequestheaders.X-Tenant=${TENANT}"
      - "traefik.http.routers.auth.middlewares=tenant-header"

  workspace:
    build: ./modules/workspace
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.workspace.rule=PathPrefix(`/api/workspaces/`)"
      - "traefik.http.routers.workspace.entrypoints=websecure"
      - "traefik.http.routers.workspace.tls=true"
      - "traefik.http.routers.workspace.middlewares=tenant-header"

  page:
    build: ./modules/page
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.page.rule=PathPrefix(`/api/pages/`)"
      - "traefik.http.routers.page.entrypoints=websecure"
      - "traefik.http.routers.page.tls=true"
      - "traefik.http.routers.page.middlewares=tenant-header"
```

---

## Estructura de Cada Módulo

```
modules/
  auth/
    domain/
      entities/
      repositories/
      events/
      services/
      errors/
    
    application/
      register/
        handler.go
        request.go
        response.go
      login/
        handler.go
        request.go
        response.go
    
    infrastructure/
      http/                    # ← REST API para frontend
        router.go              #    Rutas HTTP
        middleware.go          #    Extrae tenant, valida JWT
        handlers/              #    Adapta application handlers a HTTP
          register_handler.go
          login_handler.go
      
      grpc/                    # ← gRPC para inter-module communication
        proto/
          auth.proto
        server.go              #    gRPC server
        interceptors.go        #    Extrae tenant desde metadata
      
      persistence/
        user_postgres.go
        user_memory.go
        mapper.go
    
    main.go                    # Inicia HTTP server (8080) + gRPC server (9080)
```

---

## Puerto Assignments (Desarrollo)

| Módulo       | HTTP Port | gRPC Port | Base Path           |
|--------------|-----------|-----------|---------------------|
| auth         | 8080      | 9080      | /api/auth/*         |
| organization | 8081      | 9081      | /api/organizations/*|
| workspace    | 8082      | 9082      | /api/workspaces/*   |
| page         | 8083      | 9083      | /api/pages/*        |
| monitoring   | 8084      | 9084      | /api/monitoring/*   |
| alert        | 8085      | 9085      | /api/alerts/*       |
| insight      | 8086      | 9086      | /api/insights/*     |
| report       | 8087      | 9087      | /api/reports/*      |
| integration  | 8088      | 9088      | /api/integrations/* |
| usage        | 8089      | 9089      | /api/usage/*        |

---

## Ejemplo: main.go de un Módulo

```go
// modules/page/main.go

package main

import (
    "context"
    "log"
    "net"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/pulzifi/shared/config"
    "github.com/pulzifi/shared/database"
    "github.com/pulzifi/modules/page/infrastructure/http/router"
    "github.com/pulzifi/modules/page/infrastructure/grpc"
    "google.golang.org/grpc"
)

func main() {
    // 1. Cargar configuración
    cfg := config.Load()

    // 2. Conectar a base de datos
    db := database.Connect(cfg.DatabaseURL)
    defer db.Close()

    // 3. Instanciar repositories, handlers, etc.
    // ... (inyección de dependencias)

    // 4. Iniciar HTTP server (REST API para frontend)
    httpRouter := router.NewRouter(handlers, middlewares)
    httpServer := &http.Server{
        Addr:    ":8083",
        Handler: httpRouter,
    }

    go func() {
        log.Println("HTTP server listening on :8083")
        if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("HTTP server error: %v", err)
        }
    }()

    // 5. Iniciar gRPC server (inter-module communication)
    grpcServer := grpc.NewServer(
        grpc.UnaryInterceptor(grpc.TenantInterceptor),
    )
    pageService := grpc.NewPageService(handlers)
    proto.RegisterPageServiceServer(grpcServer, pageService)

    grpcListener, err := net.Listen("tcp", ":9083")
    if err != nil {
        log.Fatalf("Failed to listen on gRPC port: %v", err)
    }

    go func() {
        log.Println("gRPC server listening on :9083")
        if err := grpcServer.Serve(grpcListener); err != nil {
            log.Fatalf("gRPC server error: %v", err)
        }
    }()

    // 6. Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Shutting down servers...")

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := httpServer.Shutdown(ctx); err != nil {
        log.Printf("HTTP server shutdown error: %v", err)
    }

    grpcServer.GracefulStop()
    log.Println("Servers stopped")
}
```

---

## Resumen

✅ **NO hay módulo `gateway` en el código**  
✅ **Load Balancer (Nginx/Traefik) es infraestructura externa**  
✅ **Cada módulo expone HTTP (para frontend) + gRPC (para inter-module)**  
✅ **Tenant extraído por Load Balancer → pasado como header `X-Tenant`**  
✅ **Cada módulo valida JWT en su propio middleware**  
✅ **Comunicación síncrona entre módulos: gRPC**  
✅ **Comunicación asíncrona entre módulos: Kafka**

---

**Última actualización:** 2025-10-25
