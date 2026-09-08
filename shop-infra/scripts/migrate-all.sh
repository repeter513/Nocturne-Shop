#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BASE="$(cd "$ROOT/.." && pwd)"

PG_USER="${POSTGRES_USER:-shop}"
PG_PASS="${POSTGRES_PASSWORD:-shop}"
PG_HOST="${POSTGRES_HOST:-localhost}"
PG_PORT="${POSTGRES_PORT:-5432}"
PG="postgres://${PG_USER}:${PG_PASS}@${PG_HOST}:${PG_PORT}"

wait_postgres() {
  until psql "${PG}/postgres?sslmode=disable" -c '\q' 2>/dev/null; do
    echo "waiting for postgres at ${PG_HOST}:${PG_PORT}..."
    sleep 1
  done
}

migrate() {
  local dir=$1 db=$2
  echo "==> ${dir} (${db})"
  (cd "${BASE}/${dir}" && DATABASE_URL="${PG}/${db}?sslmode=disable" make migrate-up)
}

wait_postgres

migrate shop-auth auth_db
migrate shop-catalog catalog_db
migrate shop-cart cart_db
migrate shop-order order_db
migrate shop-payment payment_db

echo "done"
