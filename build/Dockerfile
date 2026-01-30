# Stage 1: Build
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy dependency files first for better caching
COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify

# Copy source code
COPY cmd ./cmd
COPY internal ./internal
COPY pkg ./pkg

# Build all binaries
ENV CGO_ENABLED=0
RUN go build -ldflags="-w -s" -o bin/api ./cmd/api
RUN go build -ldflags="-w -s" -o bin/migrate ./cmd/migrate
RUN go build -ldflags="-w -s" -o bin/seed ./cmd/seed

# Stage 2: Runtime
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata curl

# Copy binaries
COPY --from=builder /app/bin/api /usr/local/bin/api
COPY --from=builder /app/bin/migrate /usr/local/bin/migrate
COPY --from=builder /app/bin/seed /usr/local/bin/seed

# Copy migrations (if you have them in a migrations folder)
COPY migrations ./migrations

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:${SERVER_PORT:-8080}/health || exit 1

# Default entrypoint
ENTRYPOINT ["api"]