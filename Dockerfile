# syntax=docker/dockerfile:1.7

# --- Build stage ---
FROM golang:1.26-bookworm AS builder
ARG GOPROXY=https://go.devneeds.ir,direct
ARG GOSUMDB=off
ENV GOPROXY=$GOPROXY
ENV GOSUMDB=$GOSUMDB
WORKDIR /src

# Enable Go modules and download deps first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source
COPY . .

# Build the API binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/gateway ./cmd/api

# --- Runtime stage ---
FROM scratch AS runner
WORKDIR /app

# Copy binary and required runtime assets (DB schema for migrations)
COPY --from=builder /out/gateway /usr/local/bin/gateway
COPY db/db.sql db/db.sql

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/gateway"]
