FROM ghcr.io/realkarych/seqwall:latest AS seqwall
FROM migrate/migrate:v4.17.1 AS migrate

FROM alpine:3.20

RUN apk add --no-cache bash ca-certificates postgresql-client

COPY --from=seqwall /usr/bin/seqwall /usr/local/bin/seqwall
COPY --from=migrate /usr/local/bin/migrate /usr/local/bin/migrate

WORKDIR /app

COPY migrations /app/migrations
COPY scripts /app/scripts

RUN chmod +x /app/scripts/*.sh

ENTRYPOINT ["/app/scripts/migrator-entrypoint.sh"]
