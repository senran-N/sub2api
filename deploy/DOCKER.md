# Sub2API Docker Image

Sub2API is an AI API gateway platform for distributing and managing AI product subscription API quotas.

## Quick Start

The container expects PostgreSQL and Redis to be reachable. For a single-host deployment with bundled Postgres/Redis, use the Compose files in [`deploy/README.md`](./README.md) instead.

```bash
docker run -d \
  --name sub2api \
  -p 8080:8080 \
  -v sub2api_data:/app/data \
  -e AUTO_SETUP=true \
  -e SERVER_PORT=8080 \
  -e SERVER_MODE=release \
  -e SERVER_SHUTDOWN_TIMEOUT_SECONDS=45 \
  -e DATABASE_HOST=postgres.example.internal \
  -e DATABASE_PORT=5432 \
  -e DATABASE_USER=sub2api \
  -e DATABASE_PASSWORD=change_this_database_password \
  -e DATABASE_DBNAME=sub2api \
  -e REDIS_HOST=redis.example.internal \
  -e REDIS_PORT=6379 \
  -e JWT_SECRET="$(openssl rand -hex 32)" \
  -e TOTP_ENCRYPTION_KEY="$(openssl rand -hex 32)" \
  -e SECURITY_URL_ALLOWLIST_ENABLED=true \
  -e SECURITY_URL_ALLOWLIST_ALLOW_INSECURE_HTTP=false \
  -e SECURITY_URL_ALLOWLIST_ALLOW_PRIVATE_HOSTS=false \
  weishaw/sub2api:latest
```

## Recommended Compose Paths

Use the maintained Compose examples from this repository instead of writing a fresh file from scratch:

- `deploy/docker-compose.yml`: named volumes for app, Postgres, and Redis
- `deploy/docker-compose.local.yml`: local directories for easier backup and migration
- `deploy/docker-compose.standalone.yml`: external PostgreSQL and Redis
- `deploy/docker-compose.dev.yml`: build from the checked-out source tree

All of them share the same `.env.example` knobs for graceful shutdown, resource limits, and URL allowlist defaults.

## Important Environment Variables

| Variable | Required | Default | Notes |
|----------|----------|---------|-------|
| `DATABASE_HOST` | Yes | - | PostgreSQL host |
| `DATABASE_PORT` | No | `5432` | PostgreSQL port |
| `DATABASE_USER` | No | `sub2api` | PostgreSQL user |
| `DATABASE_PASSWORD` | Yes | - | PostgreSQL password |
| `DATABASE_DBNAME` | No | `sub2api` | PostgreSQL database name |
| `REDIS_HOST` | Yes | - | Redis host |
| `REDIS_PORT` | No | `6379` | Redis port |
| `REDIS_PASSWORD` | No | empty | Redis password |
| `SERVER_PORT` | No | `8080` | HTTP listen port |
| `SERVER_MODE` | No | `release` | `release` or `debug` |
| `SERVER_SHUTDOWN_TIMEOUT_SECONDS` | No | `45` | Graceful shutdown timeout |
| `JWT_SECRET` | Recommended | auto-generated if empty | Set a fixed value to keep sessions valid across restarts |
| `TOTP_ENCRYPTION_KEY` | Recommended | auto-generated if empty | Set a fixed value to keep existing 2FA secrets valid |
| `SECURITY_URL_ALLOWLIST_ENABLED` | No | `true` | Keeps upstream/pricing host validation enabled by default |
| `SECURITY_URL_ALLOWLIST_ALLOW_INSECURE_HTTP` | No | `false` | Only enable for trusted dev/test HTTP endpoints |
| `SECURITY_URL_ALLOWLIST_ALLOW_PRIVATE_HOSTS` | No | `false` | Only enable on trusted internal networks |
| `SECURITY_URL_ALLOWLIST_UPSTREAM_HOSTS` | No | empty | Extra upstream API hosts to append to the built-in allowlist |
| `SECURITY_URL_ALLOWLIST_PRICING_HOSTS` | No | empty | Extra pricing mirror hosts to append to the built-in allowlist |
| `SECURITY_URL_ALLOWLIST_CRS_HOSTS` | No | empty | CRS sync hosts to allow when enabling CRS synchronization |

If you use custom upstream, pricing, or CRS hosts while `SECURITY_URL_ALLOWLIST_ENABLED=true`, add those hosts to the matching allowlist in `config.yaml` or pass the matching `SECURITY_URL_ALLOWLIST_*_HOSTS` env var.

## Supported Architectures

- `linux/amd64`
- `linux/arm64`

## Tags

- `latest`: latest stable release
- `x.y.z`: specific version
- `x.y`: latest patch in a minor series
- `x`: latest minor in a major series

## Links

- [GitHub Repository](https://github.com/weishaw/sub2api)
- [Deployment Documentation](https://github.com/weishaw/sub2api/blob/main/deploy/README.md)
