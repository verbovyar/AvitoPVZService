FROM golang:1.21-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git make

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN make build || go build -o avito-pvz-service ./cmd/main.go

FROM alpine:3.19

WORKDIR /app

RUN addgroup -S appgroup && \
    adduser -S appuser -G appgroup && \
    chown appuser:appgroup /app

COPY --from=builder --chown=appuser:appgroup /app/avito-pvz-service .
COPY --from=builder --chown=appuser:appgroup /app/migrations ./migrations

USER appuser

ENV APP_PORT=8080 \
    DB_HOST=postgres \
    DB_PORT=5432 \
    DB_USER=postgres \
    DB_PASSWORD=postgres \
    DB_NAME=avitopvz \
    MIGRATIONS_PATH=/app/migrations

EXPOSE $APP_PORT

HEALTHCHECK --interval=30s --timeout=3s \
    CMD wget --no-verbose --tries=1 --spider http://localhost:$APP_PORT/health || exit 1

CMD ["./avito-pvz-service"]