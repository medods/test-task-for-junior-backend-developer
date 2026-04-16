#!/usr/bin/env bash
set -euo pipefail

action="${1:-}"
current_migration="${2:-}"

if [[ -z "$action" || -z "$current_migration" ]]; then
	echo "Usage: $0 <up|down> <path-to-current-migration.up.sql>" >&2
	exit 2
fi

if [[ "$action" != "up" && "$action" != "down" ]]; then
	echo "Invalid action: $action (expected up|down)" >&2
	exit 2
fi

file_name="$(basename "$current_migration")"
if [[ ! "$file_name" =~ ^([0-9]+)_.+\.up\.sql$ ]]; then
	echo "Unexpected migration filename: $file_name" >&2
	echo "Expected format like: 000006_some_name.up.sql" >&2
	exit 2
fi

raw_version="${BASH_REMATCH[1]}"
normalized_version="$((10#$raw_version))"

migrations_path="/app/migrations"
if [[ -n "${MIGRATIONS_PATH:-}" ]]; then
	if [[ -d "$MIGRATIONS_PATH" ]]; then
		migrations_path="$MIGRATIONS_PATH"
	else
		migrations_path="$(dirname "$MIGRATIONS_PATH")"
	fi
fi

db_url="${TEST_DATABASE_URL:-${DATABASE_URL:-}}"
if [[ -z "$db_url" ]]; then
	echo "DATABASE_URL or TEST_DATABASE_URL is required" >&2
	exit 1
fi

if [[ "$action" == "up" ]]; then
	exec migrate -path "$migrations_path" -database "$db_url" goto "$normalized_version"
fi

if [[ "$normalized_version" == "1" ]]; then
	exec migrate -path "$migrations_path" -database "$db_url" down 1
fi

target_version=$(( normalized_version > 0 ? normalized_version - 1 : 0 ))
exec migrate -path "$migrations_path" -database "$db_url" goto "$target_version"
