APP_NAME := wallet-api
COMPOSE := docker compose
K6_SCRIPT ?= load_test.js
K6_RATE ?= 1000
K6_DURATION ?= 30s
K6_VUS ?= 300
K6_MAX_VUS ?= 1000

.PHONY: help build test test-race test-cover docker-build up down restart logs ps load-test clean

help:
	@echo "Available commands:"
	@echo "  make build        Build Go application binary"
	@echo "  make test         Run Go tests"
	@echo "  make test-race    Run Go tests with race detector"
	@echo "  make test-cover   Run Go tests with coverage"
	@echo "  make docker-build Build Docker image via docker compose"
	@echo "  make up           Start docker compose stack"
	@echo "  make down         Stop docker compose stack"
	@echo "  make restart      Restart docker compose stack"
	@echo "  make load-test    Run k6 load test"

build:
	go build -o bin/$(APP_NAME) ./cmd

test:
	go test ./...

test-race:
	go test -race ./...

test-cover:
	go test -cover ./...

docker-build:
	$(COMPOSE) build

up:
	$(COMPOSE) up -d

down:
	$(COMPOSE) down

restart: down up

load-test:
	k6 run -e RATE=$(K6_RATE) -e DURATION=$(K6_DURATION) -e VUS=$(K6_VUS) -e MAX_VUS=$(K6_MAX_VUS) $(K6_SCRIPT)
