docs:
	swag init -g cmd/api/main.go -o docs

build:
	go build -o bin/api ./cmd/api
	go build -o bin/worker ./cmd/worker

start:
	go run ./cmd/api

worker:
	go run ./cmd/worker