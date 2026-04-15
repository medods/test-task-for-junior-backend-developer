FROM golang:1.23.0-alpine AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/taskservice-api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/task-materializer ./cmd/materializer

FROM alpine:3.21

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /out/taskservice-api /app/taskservice-api
COPY --from=builder /out/task-materializer /app/task-materializer
COPY docker/start-scheduler.sh /app/start-scheduler.sh

RUN chmod +x /app/taskservice-api /app/task-materializer /app/start-scheduler.sh

EXPOSE 8080

CMD ["/app/taskservice-api"]
