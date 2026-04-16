#!/usr/bin/env bash
set -euo pipefail

exec /app/scripts/ci-migrate-one.sh down "$@"
