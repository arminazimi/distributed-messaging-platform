BENCH_BASE_URL ?= http://localhost:8080
BENCH_RPS ?= 1000
BENCH_DURATION ?= 30s
BENCH_CONCURRENCY ?= 200
BENCH_USERS ?= 5000
BENCH_RECIPIENTS ?= 1
BENCH_EXPRESS_RATIO ?= 0.2
BENCH_SEED_BALANCE ?= 100000
BENCH_SEED_TIMEOUT ?= 5m

.PHONY: build run loadtest benchmark seed lint test tidy swag docker

build:
	go build ./...

run:
	go run ./cmd/api

loadtest:
	go run ./cmd/loadtest -base-url $(BENCH_BASE_URL) -rps $(BENCH_RPS) -duration $(BENCH_DURATION) -concurrency $(BENCH_CONCURRENCY) -users $(BENCH_USERS) -recipients $(BENCH_RECIPIENTS) -express-ratio $(BENCH_EXPRESS_RATIO)

benchmark: loadtest

seed:
	DB_HOST=localhost DB_PORT=3306 DB_USER_NAME=sms_user DB_PASSWORD=sms_pass DB_NAME=sms_gateway \
	go run ./cmd/loadtest -seed-only -seed-method db -seed-balance $(BENCH_SEED_BALANCE) -seed-timeout $(BENCH_SEED_TIMEOUT) -users $(BENCH_USERS)

lint:
	golangci-lint run

test:
	go test ./...

tidy:
	go mod tidy

swag:
	swag fmt && swag init -g cmd/api/main.go -o docs

docker:
	docker compose up --build
