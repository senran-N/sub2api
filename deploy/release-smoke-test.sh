#!/bin/sh
set -eu

if [ "${1:-}" = "" ]; then
    echo "usage: $0 <image-ref>" >&2
    exit 2
fi

IMAGE="$1"
NETWORK="sub2api-smoke"
POSTGRES_CONTAINER="sub2api-smoke-postgres"
REDIS_CONTAINER="sub2api-smoke-redis"
APP_CONTAINER="sub2api-smoke-app"
APP_PORT="18080"

cleanup() {
    docker rm -f "${APP_CONTAINER}" "${POSTGRES_CONTAINER}" "${REDIS_CONTAINER}" 2>/dev/null || true
    docker network rm "${NETWORK}" 2>/dev/null || true
}

trap cleanup EXIT

docker network create "${NETWORK}" >/dev/null

docker run -d \
    --name "${POSTGRES_CONTAINER}" \
    --network "${NETWORK}" \
    -e POSTGRES_DB=sub2api \
    -e POSTGRES_USER=sub2api \
    -e POSTGRES_PASSWORD=sub2api_smoke_pw \
    postgres:18-alpine >/dev/null

docker run -d \
    --name "${REDIS_CONTAINER}" \
    --network "${NETWORK}" \
    redis:7-alpine >/dev/null

until docker exec "${POSTGRES_CONTAINER}" pg_isready -U sub2api -d sub2api >/dev/null 2>&1; do
    sleep 2
done

for attempt in $(seq 1 12); do
    if docker pull "${IMAGE}"; then
        break
    fi
    if [ "${attempt}" -eq 12 ]; then
        echo "failed to pull released image: ${IMAGE}" >&2
        exit 1
    fi
    sleep 10
done

docker run -d \
    --name "${APP_CONTAINER}" \
    --network "${NETWORK}" \
    -p "${APP_PORT}:8080" \
    -e AUTO_SETUP=true \
    -e SERVER_PORT=8080 \
    -e SERVER_MODE=release \
    -e SERVER_SHUTDOWN_TIMEOUT_SECONDS=45 \
    -e DATABASE_HOST="${POSTGRES_CONTAINER}" \
    -e DATABASE_PORT=5432 \
    -e DATABASE_USER=sub2api \
    -e DATABASE_PASSWORD=sub2api_smoke_pw \
    -e DATABASE_DBNAME=sub2api \
    -e REDIS_HOST="${REDIS_CONTAINER}" \
    -e REDIS_PORT=6379 \
    -e JWT_SECRET=0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef \
    -e TOTP_ENCRYPTION_KEY=abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789 \
    -e SECURITY_URL_ALLOWLIST_ENABLED=true \
    -e SECURITY_URL_ALLOWLIST_ALLOW_INSECURE_HTTP=false \
    -e SECURITY_URL_ALLOWLIST_ALLOW_PRIVATE_HOSTS=false \
    "${IMAGE}" >/dev/null

for attempt in $(seq 1 30); do
    if curl -fsS "http://127.0.0.1:${APP_PORT}/health" >/tmp/sub2api-smoke-health.txt; then
        break
    fi
    if [ "${attempt}" -eq 30 ]; then
        echo "smoke test health check failed" >&2
        docker logs "${APP_CONTAINER}" || true
        exit 1
    fi
    sleep 2
done

curl -fsS -o /tmp/sub2api-smoke-index.html "http://127.0.0.1:${APP_PORT}/"
test -s /tmp/sub2api-smoke-index.html

echo "release smoke test passed for ${IMAGE}"
