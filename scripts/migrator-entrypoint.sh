#!/usr/bin/env bash
set -euo pipefail

main_db_url="${DATABASE_URL:-}"
test_db_url="${TEST_DATABASE_URL:-}"
migrations_path="/app/migrations"

postgres_host="${POSTGRES_HOST:-db}"
postgres_port="${POSTGRES_PORT:-5432}"

if [[ -z "$main_db_url" ]]; then
  echo "DATABASE_URL is required" >&2
  exit 1
fi

if [[ -z "$test_db_url" ]]; then
  echo "TEST_DATABASE_URL is required" >&2
  exit 1
fi

until pg_isready -h "$postgres_host" -p "$postgres_port" -U "${POSTGRES_USER:?POSTGRES_USER is required}" >/dev/null 2>&1; do
  sleep 1
done

export PGPASSWORD="${POSTGRES_PASSWORD:?POSTGRES_PASSWORD is required}"
psql -h "$postgres_host" -p "$postgres_port" -U "$POSTGRES_USER" -d postgres -v ON_ERROR_STOP=1 <<SQL
SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '${POSTGRES_TEST_DB:?POSTGRES_TEST_DB is required}';
DROP DATABASE IF EXISTS "${POSTGRES_TEST_DB}";
CREATE DATABASE "${POSTGRES_TEST_DB}";
SQL

export MIGRATIONS_PATH="$migrations_path"

export DATABASE_URL="$test_db_url"

seqwall staircase \
  --postgres-url "$test_db_url" \
  --migrations-path "$migrations_path" \
  --migrations-extension ".up.sql" \
  --upgrade "/app/scripts/ci-up-one.sh {current_migration}" \
  --downgrade "/app/scripts/ci-down-one.sh {current_migration}"

export DATABASE_URL="$main_db_url"

migration_version="${MIGRATION_VERSION:-}"
if [[ -z "$migration_version" ]]; then
  migrate -path "$migrations_path" -database "$main_db_url" up
  exit 0
fi

normalized_version="$(printf '%s' "$migration_version" | sed 's/^0*//')"
if [[ -z "$normalized_version" ]]; then
  normalized_version="0"
fi

migrate -path "$migrations_path" -database "$main_db_url" goto "$normalized_version"
