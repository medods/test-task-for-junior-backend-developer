#!/bin/sh
set -eu

: "${DATABASE_DSN:?DATABASE_DSN is required}"

CRON_SCHEDULE="${SCHEDULER_CRON:-5 0 * * *}"
APP_TIMEZONE="${APP_TIMEZONE:-UTC}"
MATERIALIZATION_WINDOW_DAYS="${MATERIALIZATION_WINDOW_DAYS:-30}"

echo "Running bootstrap materialization"
/app/task-materializer

cat > /etc/crontabs/root <<EOF
SHELL=/bin/sh
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
DATABASE_DSN=${DATABASE_DSN}
APP_TIMEZONE=${APP_TIMEZONE}
MATERIALIZATION_WINDOW_DAYS=${MATERIALIZATION_WINDOW_DAYS}
${CRON_SCHEDULE} /app/task-materializer >> /proc/1/fd/1 2>> /proc/1/fd/2
EOF

exec crond -f -l 2
