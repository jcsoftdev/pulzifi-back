# Multi-Tenant Development Setup

## Arquitectura

```
Browser (demo.localhost:3000)
    ↓
Next.js Frontend (localhost:3000)
    ↓ HTTP requests to localhost:80
Nginx (Docker - puerto 80)
    ↓ Extrae tenant del subdomain → X-Tenant header
Monolith Backend (Docker - puerto 9090)
    ↓
PostgreSQL + Redis (Docker)
```

## Setup

### 1. Configurar hosts

Edita `/etc/hosts`:
```bash
sudo nano /etc/hosts

# Añade:
127.0.0.1 demo.localhost
127.0.0.1 tenant1.localhost
127.0.0.1 tenant2.localhost
```

### 2. Iniciar backend con Nginx

```bash
docker-compose -f docker-compose.monolith.yml up -d
```

Esto inicia:
- PostgreSQL (5434)
- Redis (6379)
- Monolith (9090)
- **Nginx (80/443)** - Extrae tenant y hace proxy al monolith

### 3. Iniciar frontend

```bash
cd frontend
bun dev
```

Frontend corre en `http://localhost:3000`

## Uso

### Acceder con diferentes tenants

- **http://demo.localhost:3000** → Tenant: "demo"
- **http://tenant1.localhost:3000** → Tenant: "tenant1"
- **http://tenant2.localhost:3000** → Tenant: "tenant2"

### Cómo funciona

1. Navegas a `demo.localhost:3000`
2. Frontend detecta tenant "demo" del subdomain
3. Frontend hace request a `http://localhost/api/...`
4. **Nginx recibe request**:
   - Extrae tenant del header `Host: demo.localhost`
   - Añade header `X-Tenant: demo`
   - Hace proxy a `monolith:9090/api/...`
5. Backend recibe `X-Tenant: demo` y responde con datos del tenant

## Ventajas

✅ **Desarrollo**: Nginx + monolith en Docker, frontend fuera  
✅ **Producción**: Todo en Docker con Nginx como API Gateway  
✅ **Multi-tenant**: Tenant automático desde subdomain  
✅ **CORS**: Configurado para localhost:3000  
✅ **Hot reload**: Frontend se recarga automáticamente  

## Variables de entorno

**Frontend** (`.env.local`):
```bash
NEXT_PUBLIC_APP_DOMAIN=localhost
SERVER_API_URL=http://localhost:9090
```

**Backend** (`.env`):
```bash
HTTP_PORT=9090
COOKIE_DOMAIN=localhost
FRONTEND_URL=http://localhost:3000
DB_HOST=postgres
REDIS_HOST=redis
```

## Logs

Ver logs del monolith:
```bash
docker logs -f pulzifi-monolith
```

## Troubleshooting

**Error de CORS?**
- Verifica que el frontend corra en `localhost:3000`
- Verifica `CORS_ALLOWED_ORIGINS` en `.env`

**Backend no recibe X-Tenant?**
- El frontend proxy extrae el tenant del subdomain y lo envía como header `X-Tenant`
- Usa subdomain: `demo.localhost:3000` (no solo `localhost:3000`)
