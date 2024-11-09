.PHONY: proto build run test docker-build docker-run

# Go related variables
BINARY_NAME=server
MAIN_PATH=cmd/server/main.go

# Docker related variables
DOCKER_IMAGE=crypto-training-app
DOCKER_TAG=latest

# Proto related variables
PROTO_DIR=api/proto

# Build the project
build:
	go build -o ${BINARY_NAME} ${MAIN_PATH}

	# Run the project
run:
	go run ${MAIN_PATH}

# Run tests
test:
	go test -v ./...

# Generate proto files
proto:
	buf generate

# Build docker image
docker-build:
	docker build -t ${DOCKER_IMAGE}:${DOCKER_TAG} .

# Run docker compose
docker-up:
	docker-compose up -d

# Stop docker compose
docker-down:
	docker-compose down

# Clean
clean:
	go clean
	rm -f ${BINARY_NAME}

# Install dependencies
deps:
	go mod tidy