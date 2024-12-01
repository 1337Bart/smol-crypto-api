.PHONY: proto build run test docker-build docker-run

# Go related variables
BINARY_NAME=server
MAIN_PATH=cmd/server/main.go

# Proto related variables
PROTO_DIR=api/proto


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