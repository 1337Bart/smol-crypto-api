.PHONY: proto build run test docker-build docker-run

test:
	go test -v ./...

# Generate proto files
proto:
	buf generate

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

migrate-up:
	migrate -path internal/repository/postgres/migrations -database "postgresql://crypto_app:crypto_password@localhost:5432/crypto_db?sslmode=disable" up

migrate-down:
	migrate -path internal/repository/postgres/migrations -database "postgresql://crypto_app:crypto_password@localhost:5432/crypto_db?sslmode=disable" down

build-server:
	go build -o bin/server cmd/server/main.go
	go run cmd/server/main.go

fetch:
	go build -o bin/data_fetcher cmd/data_fetcher/main.go
	go run cmd/data_fetcher/main.go