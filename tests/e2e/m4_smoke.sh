#!/usr/bin/env bash
# 烟雾测试：验证生产 compose 栈中 core、web 静态服务及 nginx 反代的连通性。
# Smoke test: verify connectivity of core, web static, and nginx proxy in the production compose stack.
#
# 用法：bash tests/e2e/m4_smoke.sh
# Usage: bash tests/e2e/m4_smoke.sh

set -euo pipefail
cd "$(dirname "$0")/../.."

cleanup() {
    echo "[m4-smoke] cleanup..."
    (cd deploy && docker compose -f docker-compose.prod.yml --env-file .env.smoke down 2>/dev/null) || true
    rm -f deploy/.env.smoke
    docker compose -f deploy/docker-compose.yml up -d 2>/dev/null || true
}
trap cleanup EXIT

cat > deploy/.env.smoke <<EOF
POSTGRES_DB=stellaris
POSTGRES_USER=stellaris
POSTGRES_PASSWORD=stellaris
USER_JWT_SECRET=smoke-test-user-secret-not-for-prod
PLANET_JWT_SECRET=smoke-test-planet-secret-not-for-prod
EOF

docker compose -f deploy/docker-compose.yml down 2>/dev/null || true
make docker-build
cd deploy && docker compose -f docker-compose.prod.yml --env-file .env.smoke up -d
cd ..

# 等 core 就绪（最多 30s）
# Wait for core to become ready (max 30s)
for _ in $(seq 1 30); do
    code=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:4228/api/auth/register || echo 000)
    [ "$code" = "405" ] && break
    sleep 1
done

# 1. core 直连
# 1. Direct connection to core
TOK=$(curl -s -X POST http://localhost:4228/api/auth/register \
    -H 'Content-Type: application/json' \
    -d "{\"email\":\"smoke-$$@m4.test\",\"password\":\"pwd\"}" \
    | grep -o '"token":"[^"]*' | cut -d'"' -f4)
[ -n "$TOK" ] || { echo "FAIL: register direct"; exit 1; }
echo "[m4-smoke] ✓ register via :4228 returned token"

# 2. web nginx 静态
# 2. Web nginx static
HTTP=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/)
[ "$HTTP" = "200" ] || { echo "FAIL: web HTTP $HTTP"; exit 1; }
echo "[m4-smoke] ✓ web index 200"

# 3. web → core API 反代
# 3. Web nginx proxy to core API
PROXY=$(curl -s -X POST http://localhost:8080/api/auth/register \
    -H 'Content-Type: application/json' \
    -d "{\"email\":\"smoke-proxy-$$@m4.test\",\"password\":\"pwd\"}" \
    | grep -o '"token":"[^"]*' | cut -d'"' -f4)
[ -n "$PROXY" ] || { echo "FAIL: web proxied API"; exit 1; }
echo "[m4-smoke] ✓ web → core /api proxy works"

echo "[m4-smoke] ALL PASS"
