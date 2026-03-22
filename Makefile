docs:
	swag init -g cmd/api/main.go -o docs

build:
	go build -o bin/api ./cmd/api
	go build -o bin/update ./cmd/update
	go build -o bin/worker ./cmd/worker

run:
	go run ./cmd/api

update:
	go run ./cmd/update

worker:
	go run ./cmd/worker