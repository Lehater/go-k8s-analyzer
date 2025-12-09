APP_NAME=go-k8s-analyzer

.PHONY: build test run docker-build docker-run

build:
	go build -o bin/server ./cmd/server

test:
	go test ./...

run:
	go run ./cmd/server

docker-build:
	docker build -t $(APP_NAME):latest .

docker-run:
	docker run --rm -p 8080:8080 --env-file .env $(APP_NAME):latest
