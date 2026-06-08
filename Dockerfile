# syntax=docker/dockerfile:1.7

# --- Build stage ---
FROM golang:1.26-bookworm AS builder
ENV GOFLAGS=-mod=vendor
WORKDIR /src

# Copy source including the checked-in vendor directory.
COPY . .

# Build the API binary using vendored dependencies only.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/gateway ./cmd/api

# --- Runtime stage ---
FROM scratch AS runner
WORKDIR /app

# Copy binary and required runtime assets (DB schema for migrations)
COPY --from=builder /out/gateway /usr/local/bin/gateway
COPY db/db.sql db/db.sql

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/gateway"]
